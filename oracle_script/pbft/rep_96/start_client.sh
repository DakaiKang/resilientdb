#!/bin/bash
CLIENT_NUM=1
tot=96
count=1

while(( $count <= $CLIENT_NUM ))
do
  idx=`expr $tot + $count`;
  echo $idx
  ../../../bazel-bin/example/kv_server_tools client${idx}.config get key&
  count=$(($count+1))
done

