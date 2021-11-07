#ifndef _MICROSERVICEMANAGER_H_
#define _MICROSERVICEMANAGER_H_
#include "global.h"
#include "message.h"
#include <nng/nng.h>

#if SERVERLESS

enum ServType {
  INVKR = 0,
  VMCS
};

class MicroServiceManager
{
  public:
    void init(string servicename);
    void send_msg(string msg);
    string receive_msg();
    VerifierResponseMessage* request_vmsg();
  private:
    //MsgPattern msg_pattern;
    //int sock_pub;
    //int sock_sub;
    //int sock_pair;
    //string in_addr;
    //string out_addr;
    string addr;
    nng_socket sock;
    //void init_sender();
    //void init_receiver();
    void init_pair();
    void handle_error(const char* func, int rv);
};

#endif // serverless

#endif
