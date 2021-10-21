# assumes that you have the appropriate entry in your locall SSH config:
LPURP='\033[1;35m'
NC='\033[0m'
cd src
printf "${LPURP}Building verifier${NC}\n"
go build verifier.go
printf "${LPURP}Erasing old verifier${NC}\n"
ssh -t sdb-ec2 'cd vdb;echo 'removing old binary';rm -f verifier;'
printf "${LPURP}Deploying verifier binary & config${NC}\n"
scp verifier sdb-ec2:~/vdb/
scp ../verifier_config.json sdb-ec2:~/vdb/
printf "${LPURP}Removing local binary${NC}\n"
rm -f verifier
printf "${LPURP}Done${NC}\n"
