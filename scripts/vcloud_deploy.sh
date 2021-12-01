#!/bin/bash
# RES_FILE --> Name of the result file.
#
USERNAME=ubuntu
HOSTS="$1"
NODE_CNT="$2"
RES_FILE="$3"
count=0
# verifier alias name (assumes alias is set in ssh config)
verifier_host="$4"

ssh -n ${verifier_host} "cd vdb; sudo ./verifier > ${RES_FILE}_v.out 2>&1" &

for HOSTNAME in ${HOSTS}; do
	i=1
	if [ $count -ge $NODE_CNT ]; then
		i=0
	    SCRIPT="ulimit -n 4096;./v_recv ${count} > v_recv.log 2>&1 & ./runcl -nid${count} > ${RES_FILE}${count}.out 2>&1"
	    echo "${HOSTNAME}: v_recv and runcl ${count}"
	else
	    SCRIPT="ulimit -n 4096; ./rundb -nid${count} > ${RES_FILE}${count}.out 2>&1"
	    echo "${HOSTNAME}: rundb ${count}"
	fi
	# if [ "$i" -eq 0 ];then
		ssh -n -o BatchMode=yes -o StrictHostKeyChecking=no -l ${USERNAME} ${HOSTNAME} "
		cd resilientdb
		${SCRIPT}" &
	# fi
	count=`expr $count + 1`
done
invoker=$(echo $HOSTS | cut -d' ' -f1)
echo $invoker
ssh -n -o BatchMode=yes -o StrictHostKeyChecking=no -l ${USERNAME} ${invoker} "cd resilientdb; ulimit -n 16000;./invoker > ${RES_FILE}_invoker.out 2>&1 " &

while [ $count -gt 0 ]; do
	wait $pids
	count=`expr $count - 1`
done
