fcnname=${PWD##*/}
LPURP='\033[1;35m'
NC='\033[0m'
cd build
printf "${LPURP}Building function${NC}\n"
make aws-lambda-package-${fcnname}
aws lambda update-function-code \
    --function-name  ${fcnname} \
    --zip-file fileb://${fcnname}.zip > updateResult.out &
pid=$! # Process Id of the previous running command
spin='-\|/'
i=0
while kill -0 $pid 2>/dev/null
do
  i=$(( (i+1) % ${#spin}))
  printf "\r${spin:$i:1} ${LPURP}Deploying Function${NC}"
  sleep .1
done
printf "\n${LPURP}Update Results:${NC}\n"
cat updateResult.out
printf "\n${LPURP}Cleaning Directory${NC}\n"
rm -f updateResult.out ${fcnname}*
printf "${LPURP}All Done. Function updated${NC}\n"
