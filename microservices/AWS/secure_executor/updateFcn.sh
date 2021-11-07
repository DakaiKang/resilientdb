rm -r build
mkdir build && cd build
cmake ..
make aws-lambda-package-secure_executor
cd ..
aws --region us-west-1 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region us-west-2 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
aws --region us-east-2 lambda update-function-code --function-name  secure_executor --zip-file fileb://build/secure_executor.zip &
wait
