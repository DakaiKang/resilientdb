#include "global.h"
#include "smart_contract.h"
#include "smart_contract_txn.h"

#if BANKING_SMART_CONTRACT
/*
returns:
     1 for commit 
     0 for abort
*/
uint64_t TransferMoneySmartContract::execute()
{
    string temp = db->Get(std::to_string(this->source_id));
    uint64_t source = temp.empty() ? 0 : stoi(temp);
    temp = db->Get(std::to_string(this->dest_id));
    uint64_t dest = temp.empty() ? 0 : stoi(temp);
    if (amount <= source)
    {
        db->Put(std::to_string(this->source_id), std::to_string(source - amount));
        db->Put(std::to_string(this->dest_id), std::to_string(dest + amount));
        return 1;
    }
    return 0;
}


string TransferMoneySmartContract::get_contract() { 
  //printf("getting TransferMoneySmartContract\n");
  //fflush(stdout);
  return "";
}

/*
returns:
     1 for commit 
*/
uint64_t DepositMoneySmartContract::execute()
{
    string temp = db->Get(std::to_string(this->dest_id));
    uint64_t dest = temp.empty() ? 0 : stoi(temp);
    db->Put(std::to_string(this->dest_id), std::to_string(dest + amount));
    return 1;
}

string DepositMoneySmartContract::get_contract() {
  // Only support for DepositMoney for testing purposes:
  return "{\"dest\":\"" + to_string(this->dest_id) + "\",\"amt\":\"" + to_string(amount) +"\"}";
}

/*
returns:
     1 for commit 
     0 for abort
*/
uint64_t WithdrawMoneySmartContract::execute()
{
    string temp = db->Get(std::to_string(this->source_id));
    uint64_t source = temp.empty() ? 0 : stoi(temp);
    if (amount <= source)
    {
        db->Put(std::to_string(this->source_id), std::to_string(source - amount));
        return 1;
    }
    return 0;
}

string WithdrawMoneySmartContract::get_contract() { 
  //printf("getting WithdrawMoneySmartContract\n");
  //fflush(stdout);
  return "";
}

/*
Smartt Contract Transaction Manager and Workload
*/

void SmartContractTxn::init(uint64_t thd_id, Workload *h_wl)
{
    TxnManager::init(thd_id, h_wl);
    _wl = (SCWorkload *)h_wl;
    reset();
}

void SmartContractTxn::reset()
{
    TxnManager::reset();
}

string SmartContract::get_contract() {
  string contract;
  switch (this->type) {
    case BSC_TRANSFER:
      {
        TransferMoneySmartContract* tm = (TransferMoneySmartContract*)this;
        contract = tm->get_contract();
        break;
      }
    case BSC_DEPOSIT:
      {
        DepositMoneySmartContract* dm = (DepositMoneySmartContract*)this;
        contract = dm->get_contract();
        break;
      }
    case BSC_WITHDRAW:
      {
        WithdrawMoneySmartContract* wm = (WithdrawMoneySmartContract*)this;
        contract = wm->get_contract();
        break;
      }
    default:
      assert(0);
      break;
  }
  return contract;
}

RC SmartContractTxn::run_txn()
{
    this->smart_contract->execute();
    return RCOK;
};

RC SCWorkload::init()
{
    Workload::init();
    return RCOK;
}

RC SCWorkload::get_txn_man(TxnManager *&txn_manager)
{
    DEBUG_M("YCSBWorkload::get_txn_man YCSBTxnManager alloc\n");
    txn_manager = (SmartContractTxn *)
                      mem_allocator.align_alloc(sizeof(SmartContractTxn));
    new (txn_manager) SmartContractTxn();
    return RCOK;
}

uint64_t SmartContract::execute()
{
    int result = 0;
    switch (this->type)
    {
    case BSC_TRANSFER:
    {
        TransferMoneySmartContract *tm = (TransferMoneySmartContract *)this;
        result = tm->execute();
        break;
    }
    case BSC_DEPOSIT:
    {
        DepositMoneySmartContract *dm = (DepositMoneySmartContract *)this;
        result = dm->execute();
        break;
    }
    case BSC_WITHDRAW:
    {
        WithdrawMoneySmartContract *wm = (WithdrawMoneySmartContract *)this;
        result = wm->execute();
        break;
    }
    default:
        assert(0);
        break;
    }

    if (result)
        return RCOK;
    else
        return NONE;
}

#endif
