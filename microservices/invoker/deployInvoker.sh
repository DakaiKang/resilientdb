USERNAME="ubuntu"
DIR="resilientdb"

usage_msg() {
  printf "Usage: ./deployInvoker.sh <location_flag>\n"
}

from_cloud() {
SNODES=(
# Specify private IP addresses of each replica
"172.31.25.47"
"172.31.22.45"
"172.31.21.215"
"172.31.16.160"
#"172.31.22.136"
#"172.31.27.75"
#"172.31.23.68"
#"172.31.16.87"
#"172.31.17.174"
#"172.31.24.220"
#"172.31.22.220"
#"172.31.29.35"
#"172.31.30.161"
#"172.31.18.19"
#"172.31.21.182"
#"172.31.20.216"
)
count=${#SNODES[*]}
i=0
while (($i < $count))
do
    echo "------ ${SNODES[$i]} ------"
    ssh -t ${USERNAME}@${SNODES[$i]} 'cd resilientdb;echo 'removing old files';rm -f invoker invoker.log' 
    cat src/invoker_config.json
    scp src/invoker_config.json ${USERNAME}@${SNODES[$i]}:~/resilientdb/
    scp src/invoker ${USERNAME}@${SNODES[$i]}:~/resilientdb/
    i=$((i+1))
done
echo "------ DONE ------"
}

from_local() {
  SNODES=(
  # Specify aliases that have been set up in your ssh config:
  # "s-1"
  )
  count=${#SNODES[*]}
  i=0
  while (($i < $count))
  do
    echo "------ ${SNODES[$i]} ------"
    ssh -t ${SNODES[$i]} "cd ${DIR};echo 'removing old files';rm -f invoker invoker.log" 
    scp src/invoker_config.json ${SNODES[$i]}:~/${DIR}/
    scp src/invoker ${USERNAME}@${SNODES[$i]}:~/${DIR}/
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
