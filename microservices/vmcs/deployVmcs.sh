USERNAME="ubuntu"
DIR="resilientdb"
usage_msg() {
  printf "Usage: ./deployVmcs.sh <location flag>\n"
}

from_cloud() {
CNODES=(
# Specify private IP addresses of each client
#"172.31.29.79"
#"172.31.30.63"
#"172.31.23.30"
#"172.31.31.107"
#"172.31.24.79"
#"172.31.22.183"
#"172.31.28.122"
#"172.31.21.121"

#"172.31.21.216"
#"172.31.28.42"


#"172.31.20.197"
#"172.31.16.22"
#"172.31.21.30"
#"172.31.16.14"
#"172.31.30.162"
#"172.31.26.77"

"172.31.16.69"
"172.31.18.245"
"172.31.22.88"
"172.31.25.59"

)
count=${#CNODES[*]}
i=0
while (($i < $count))
do
  echo "------ ${CNODES[$i]} ------"
  ssh -t ${USERNAME}@${CNODES[$i]} 'cd resilientdb;echo 'removing old files';rm -f v_recv v_recv.log' 
  #scp src/vmcs_config.json ${USERNAME}@${CNODES[$i]}:~/resilientdb/
  scp src/v_recv ${USERNAME}@${CNODES[$i]}:~/resilientdb/
  i=$((i+1))
done
echo "------ DONE ------"
}

from_local() {
CNODES=(
# Specify aliases that have been set up in your ssh config:
# "scl-ec2"
)
count=${#CNODES[*]}
i=0
while (($i < $count))
do
  echo "------ ${CNODES[$i]} ------"
  ssh -t ${CNODES[$i]} "cd ${DIR};echo 'removing old files';rm -f v_recv v_recv.log" 
  #scp src/vmcs_config.json ${USERNAME}@${CNODES[$i]}:~/resilientdb/
  scp src/v_recv ${CNODES[$i]}:~/${DIR}/
  i=$((i+1))
done
echo "------ DONE ------"
}
if [[ $# -eq 1 ]]
then
  while getopts ':lc' flag; do
    case "${flag}" in
      l) local=true;;
      c) local=false;;
      *) usage_msg
      exit 1;;
    esac
  done
  case $local in
	  (true) from_local;;
	  (false) from_cloud
  esac
else
  usage_msg
fi
