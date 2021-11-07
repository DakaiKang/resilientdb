#ifndef _SERVERLESSREQUEST_H_
#define _SERVERLESSREQUEST_H_

#include "global.h"
#include "message.h"

class ServerlessRequest {
  public:
    //string write_request;
    //string read_request;
    string sequence_number;
    string contracts;
    string certificate;
    string client_id;
    string client_ts;
    void set_seqnum(uint64_t seqn_num);
    void set_client_id(uint64_t clnt_id);
    void set_client_ts(uint64_t clnt_ts);
    void set_contracts(string contract_set);
    void set_certificate(vector<PBFTCommitMessage*>* commit_msgs);
    string create_request();
  private:
    string encode_string_as_list(string s);
    string create_encoded_sig_field(string sig, int index);
    string create_encoded_msg_field(string msg, int index);
    string create_encoded_key_field(string pkey, int index);
    string create_cryptomethod_field(string data);
    string create_seqnum_field(string seqn_num);
    string create_client_id_field(string clnt_id);
    string create_client_ts_field(string clnt_ts);
    string create_contracts_field(string contract_set);
    string create_certificate_field(string certificate_set);
};

#endif
