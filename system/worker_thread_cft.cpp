#include "global.h"
#include "message.h"
#include "thread.h"
#include "worker_thread.h"
#include "txn.h"
#include "wl.h"
#include "query.h"
#include "ycsb_query.h"
#include "math.h"
#include "msg_thread.h"
#include "msg_queue.h"
#include "work_queue.h"
#include "message.h"
#include "timer.h"
#include "chain.h"

#if CONSENSUS == PBFT && CFT
/**
 * Processes an incoming client batch and sends a Pre-prepare message to al replicas.
 *
 * This function assumes that a client sends a batch of transactions and 
 * for each transaction in the batch, a separate transaction manager is created. 
 * Next, this batch is forwarded to all the replicas as a BatchRequests Message, 
 * which corresponds to the Pre-Prepare stage in the PBFT protocol.
 *
 * @param msg Batch of Transactions of type CientQueryBatch from the client.
 * @return RC
 */
RC WorkerThread::process_client_batch(Message *msg)
{
    ClientQueryBatch *clbtch = (ClientQueryBatch *)msg;

    // printf("ClientQueryBatch: %ld, THD: %ld :: CL: %ld :: RQ: %ld\n", msg->txn_id, get_thd_id(), msg->return_node_id, clbtch->cqrySet[0]->requests[0]->key);
    // fflush(stdout);
    // Authenticate the client signature.
    validate_msg(clbtch);

    // Initialize all transaction mangers and Send BatchRequests message.
    create_and_send_batchreq(clbtch, clbtch->txn_id);

    return RCOK;
}

/**
 * Process incoming BatchRequests message from the Primary.
 *
 * This function is used by the non-primary or backup replicas to process an incoming
 * BatchRequests message sent by the primary replica. This processing would require 
 * sending messages of type PBFTPrepMessage, which correspond to the Prepare phase of 
 * the PBFT protocol. Due to network delays, it is possible that a repica may have 
 * received some messages of type PBFTPrepMessage and PBFTCommitMessage, prior to 
 * receiving this BatchRequests message.
 *
 * @param msg Batch of Transactions of type BatchRequests from the primary.
 * @return RC
 */
RC WorkerThread::process_batch(Message *msg)
{
    uint64_t cntime = get_sys_clock();

    BatchRequests *breq = (BatchRequests *)msg;

    // printf("BatchRequests: TID:%ld : VIEW: %ld : THD: %ld\n",breq->txn_id, breq->view, get_thd_id());
    // fflush(stdout);

    // Assert that only a non-primary replica has received this message.
    assert(g_node_id != get_current_view(get_thd_id()));

    // Check if the message is valid.
    validate_msg(breq);

    // Allocate transaction managers for all the transactions in the batch.
    set_txn_man_fields(breq, 0);

    txn_man->set_primarybatch(breq);

    // End the counter for pre-prepare phase as prepare phase starts next.
    double timepre = get_sys_clock() - cntime;
    INC_STATS(get_thd_id(), time_pre_prepare, timepre);

    // Sending Accept message
    Message *temp_msg = Message::create_message(txn_man, CFT_ACCEPT);
    CFTAcceptMessage *accept_msg = (CFTAcceptMessage *)temp_msg;
    vector<uint64_t> dest;
    dest.push_back(view_to_primary(get_current_view(get_thd_id())));
    msg_queue.enqueue(get_thd_id(), accept_msg, dest);
    dest.clear();

    ////////
    if (txn_man->is_committed())
    {
        // Proceed to executing this batch of transactions.
        send_execute_msg();

        // End the commit counter.
        // INC_STATS(get_thd_id(), time_commit, get_sys_clock() - txn_man->txn_stats.time_start_commit - timediff);
    }

    // Release this txn_man for other threads to use.
    bool ready = txn_man->set_ready();
    assert(ready);

    // UnSetting the ready for the txn id representing this batch.
    txn_man = get_transaction_manager(msg->txn_id, 0);
    unset_ready_txn(txn_man);

    return RCOK;
}

/**
 * Processes incoming Accept message.
 *
 * This functions precessing incoming messages of type CFTAcceptMessage. If the coordinater 
 * received f+1 identical accept messages from distinct replicas, then it creates 
 * and sends a CFTCommitMessage to all the replicas.
 *
 * @param msg Prepare message of type CFTAcceptMessage from a replica.
 * @return RC
 */
RC WorkerThread::process_cft_accept_msg(Message *msg)
{
    // cout << "CFTAcceptMessage: TID: " << msg->txn_id << " FROM: " << msg->return_node_id << endl;
    // fflush(stdout);

    // Check if the incoming message is valid.
    CFTAcceptMessage *accept_msg = (CFTAcceptMessage *)msg;
    validate_msg(accept_msg);

    uint64_t accept_cnt = 0;
    cft_accept_count.check_and_get(msg->txn_id, accept_cnt);
    accept_cnt++;
    cft_accept_count.add(msg->txn_id, accept_cnt);

    if (accept_cnt == g_min_invalid_nodes + 1)
    {
        txn_man->set_committed();

        Message *temp_msg = Message::create_message(txn_man, CFT_COMMIT);
        PBFTCommitMessage *commit_msg = (PBFTCommitMessage *)temp_msg;
        vector<uint64_t> dest;
        for (uint64_t i = 0; i < g_node_cnt; i++)
        {
            if (i == g_node_id)
            {
                continue;
            }
            dest.push_back(i);
        }

        msg_queue.enqueue(get_thd_id(), commit_msg, dest);
        dest.clear();

        send_execute_msg();
    }

    return RCOK;
}

/**
 * Processes incoming Commit message.
 *
 * This functions precessing incoming messages of type CFTCommitMessage. If the Coordinator 
 * received f+1 identical Commit messages from distinct replicas, then it asks the 
 * execute-thread to execute all the transactions in this batch.
 *
 * @param msg Commit message of type CFTCommitMessage from a replica.
 * @return RC
 */
RC WorkerThread::process_cft_commit_msg(Message *msg)
{
    // cout << "CFTCommitMessage: TID " << msg->txn_id << " FROM: " << msg->return_node_id << "\n";
    // fflush(stdout);

    // Check if message is valid.
    CFTCommitMessage *commit_msg = (CFTCommitMessage *)msg;
    validate_msg(commit_msg);

    txn_man->set_committed();

    // Add this message to execute thread's queue.
    send_execute_msg();

    INC_STATS(get_thd_id(), time_commit, get_sys_clock() - txn_man->txn_stats.time_start_commit);

    return RCOK;
}

RC WorkerThread::process_pbft_prep_msg(Message *msg)
{
    return RCOK;
}

bool WorkerThread::committed_local(PBFTCommitMessage *msg)
{
    return false;
}
RC WorkerThread::process_pbft_commit_msg(Message *msg)
{
    return RCOK;
}
#endif
