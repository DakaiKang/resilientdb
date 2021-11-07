#!/bin/sh

cd ~/resilientdb/verifier/src && go build
cd ~/resilientdb/microservices/invoker && go build
cd ~/resilientdb/microservices/vmcs && go build