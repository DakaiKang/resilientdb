#include "serverless_request.h"

string ServerlessRequest::encode_string_as_list(string s) {
    string byte_list = "[";
  for (uint64_t i = 0; i < s.size(); i++) {
    byte c = s[i];
    byte_list += to_string(int(c));
    if (i != s.size() - 1) {
      byte_list += ",";
    }
  }
  byte_list += "]";
  return byte_list; 
}

string ServerlessRequest::create_encoded_sig_field(string sig, int index) {
  string sig_as_list = encode_string_as_list(sig);
  string indexed_sig = "signature" + to_string(index);
  string sig_field = "\"" + indexed_sig + "\":" + sig_as_list;
  return sig_field;  
}

string ServerlessRequest::create_encoded_msg_field(string msg, int index) {
  string msg_as_list = encode_string_as_list(msg);
  string indexed_msg = "message" + to_string(index);
  string msg_field = "\"" + indexed_msg + "\":" + msg_as_list;
  return msg_field;
}

string ServerlessRequest::create_encoded_key_field(string pkey, int index) {
  string key_as_list = encode_string_as_list(pkey);
  string indexed_key = "pkey" + to_string(index); 
  string key_field = "\"" + indexed_key + "\":" + key_as_list;
  return key_field;
}

string ServerlessRequest::create_cryptomethod_field(string data) {
  string cryptomethod_field = "\"cryptoMethod\":\"" + data + "\"";
  return cryptomethod_field;
}

string ServerlessRequest::create_seqnum_field(string seqn_num) {
  string seqnnum_field = "\"sequenceNumber\":\"" + seqn_num + "\"";
  return seqnnum_field;
}

string ServerlessRequest::create_client_id_field(string clnt_id) {
  string client_id_field = "\"clientId\":\"" + clnt_id + "\"";
  return client_id_field;
}

string ServerlessRequest::create_client_ts_field(string clnt_ts) {
  string client_ts_field = "\"clientTs\":\"" + clnt_ts + "\"";
  return client_ts_field;
}

string ServerlessRequest::create_contracts_field(string contract_set) {
  string contracts_field = "\"contracts\":{" + contract_set + "}";
  return contracts_field;
}

string ServerlessRequest::create_certificate_field(string certificate_set) {
  string certificate_field = "\"certificate\":{" + certificate_set + "}";
  return certificate_field;
}

void ServerlessRequest::set_seqnum(uint64_t seqn_num) {
  sequence_number = to_string(seqn_num);
}

void ServerlessRequest::set_client_id(uint64_t clnt_id) {
  client_id = to_string(clnt_id);
}

void ServerlessRequest::set_client_ts(uint64_t clnt_ts) {
  client_ts = to_string(clnt_ts);
}

void ServerlessRequest::set_contracts(string contract_set) {
  contracts = contract_set;
}

void ServerlessRequest::set_certificate(vector<PBFTCommitMessage*>* commit_msgs) {
  for (uint64_t i = 0; i < commit_msgs->size(); i++) {
    string commit_msg = (*commit_msgs)[i]->toString();
    string sig = (*commit_msgs)[i]->signature;
    string pkey = (*commit_msgs)[i]->pubKey;
    string msg_field = create_encoded_msg_field(commit_msg, i);
    string sig_field = create_encoded_sig_field(sig, i);
    string key_field = create_encoded_key_field(pkey, i);
    certificate += msg_field + "," + sig_field + "," + key_field;
    if (i != commit_msgs->size() - 1) {
      certificate += ",";
    }
  }
}

string ServerlessRequest::create_request() {
  string seqnnum_field = create_seqnum_field(sequence_number);
  string client_id_field = create_client_id_field(client_id);
  string client_ts_field = create_client_ts_field(client_ts);
  string contracts_field = create_contracts_field(contracts);
  string cryptomethod_field = create_cryptomethod_field("CMAC");
  string certificate_field = create_certificate_field(certificate);
#if BANKING_SMART_CONTRACT
  string meta = "\"meta\":\"sc\"";
#else
  string meta = "\"meta\":\"ycsb\"";
#endif
  string request = "{" + seqnnum_field + "," +
                    client_id_field + "," +
                    client_ts_field + "," +
                    contracts_field + "," +
                    cryptomethod_field + "," +
                    meta + "," +
                    certificate_field + "}";
  return request;
}
