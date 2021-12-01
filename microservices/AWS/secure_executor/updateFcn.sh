rm -r build
mkdir build && cd build
cmake ..
make aws-lambda-package-secure_executor
cd ..
aws --region us-west-1 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region us-west-2 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region us-east-2 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
# aws --region us-east-1 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region ca-central-1 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region eu-central-1 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region eu-west-1 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region eu-west-2 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region eu-west-3 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region eu-north-1 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
# aws --region ap-northeast-3 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region ap-northeast-2 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region ap-southeast-1 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region ap-southeast-2 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region ap-northeast-1 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
wait
