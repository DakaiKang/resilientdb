fcnname=${PWD##*/}
mkdir build
cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make
make aws-lambda-package-${fcnname}
cd ..
aws iam create-role \
  --role-name ${fcnname}_e_role \
  --assume-role-policy-document file://trust-policy.json > iamResponse.json
arn=$(python3 getRoleArn.py)
cd build
aws lambda create-function \
--function-name ${fcnname} \
--role ${arn} \
--runtime provided \
--timeout 15 \
--memory-size 128 \
--handler ${fcnname} \
--zip-file fileb://${fcnname}.zip
