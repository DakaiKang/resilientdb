func_name="secure_executor"
# regions
  # [N. Virginia]	   us-east-1
  # [Ohio]	         us-east-2
  # [N. California]	 us-west-1
  # [Oregon]	       us-west-2
  # [Cape Town]	     af-south-1
  # [Hong Kong]	     ap-east-1
  # [Mumbai]	       ap-south-1
  # [Osaka]	         ap-northeast-3
  # [Seoul]	         ap-northeast-2
  # [Singapore]	     ap-southeast-1
  # [Sydney]	       ap-southeast-2
  # [Tokyo]	         ap-northeast-1
  # [Central]	       ca-central-1
  # [Frankfurt]	     eu-central-1
  # [Ireland]	       eu-west-1
  # [London]	       eu-west-2
  # [Milan]       	 eu-south-1
  # [Paris]	         eu-west-3
  # [Stockholm]	     eu-north-1
  # [Bahrain]	       me-south-1
  # [São Paulo]	     sa-east-1
region=""
mkdir build
cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make
make aws-lambda-package-${func_name}
cd ..
# aws iam create-role --role-name ${func_name}_e_role --assume-role-policy-document file://trust-policy.json > iamResponse.json
arn="arn:aws:iam::283888746952:role/secure_executor_e_role"
cd build
aws lambda --region ${region}create-function \
--function-name ${func_name} \
--role ${arn} \
--runtime provided \
--timeout 15 \
--memory-size 128 \
--handler ${func_name} \
--zip-file fileb://${func_name}.zip
