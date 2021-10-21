fcnname=${PWD##*/}
LPURP='\033[1;35m'
NC='\033[0m'
rm -r build
mkdir build
cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make
printf "${LPURP}Building function ${fcnname}${NC}\n"
make aws-lambda-package-${fcnname}
printf "${LPURP}Done${NC}\n"
