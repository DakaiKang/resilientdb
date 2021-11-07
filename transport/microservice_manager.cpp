#include "global.h"
#include "message.h"
#include "microservice_manager.h"
//#include <memory>
//#include <nanomsg/nn.h>
//#include <nanomsg/pubsub.h>
//#include <nanomsg/pair.h>
#include <nng/nng.h>
#include <nng/protocol/pair0/pair.h>

#if SERVERLESS

void MicroServiceManager::init(string service_name) {
  string path_name = "/tmp/" + service_name;
  addr = "ipc://" + path_name;
  init_pair();
}

void MicroServiceManager::init_pair() {
  int rv;
  if ((rv = nng_pair0_open(&sock)) != 0) {
    handle_error("nng_pair0_open", rv);
  }
  if ((rv = nng_dial(sock, addr.c_str(), NULL, 0)) != 0) {
    handle_error("nng_dial", rv);
  }
  if ((rv = nng_setopt_ms(sock, NNG_OPT_RECVTIMEO, 100)) != 0) {
    handle_error("nng_setopt_ms", rv);
  }
  sleep(1);
}

void MicroServiceManager::handle_error(const char *func, int rv) {
  fprintf(stderr, "%s: %s\n", func, nng_strerror(rv));
}

void MicroServiceManager::send_msg(string msg) {
  int rv;
  char* data = new char[msg.size() + 1];
  std::copy(msg.begin(), msg.end(), data);
  data[msg.size()] = '\0';
  if ((rv = nng_send(sock, data, strlen(data), 0)) != 0) {
    handle_error("nng_send", rv);
  }
  delete data;
}

string MicroServiceManager::receive_msg() {
  char *buf = NULL;
  int rv;
  size_t sz;
  if ((rv = nng_recv(sock, &buf, &sz, NNG_FLAG_ALLOC)) == 0) {
    nng_free(buf, sz);
    return string(buf);
  }
  handle_error("nng_recv", rv);
  return "";
}

VerifierResponseMessage* MicroServiceManager::request_vmsg() {
  nng_msg* msg = NULL;
  int rv = nng_recvmsg(sock, &msg, NNG_FLAG_NONBLOCK);
  if (rv == 0) {
    VerifierResponseMessage* vmsg = (VerifierResponseMessage*)Message::create_message(V_RESP);
    vmsg->copy_from_buf((char *)(nng_msg_body(msg)));
    nng_msg_free(msg);
    return vmsg;
  }
  //handle_error("nng_recv :: request_vmsg", rv);
  return NULL;
}

#endif // serverless