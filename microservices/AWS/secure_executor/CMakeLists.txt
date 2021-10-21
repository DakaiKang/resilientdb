cmake_minimum_required(VERSION 3.5)

set(CMAKE_CXX_STANDARD 11)
set(REPO_ROOT "/home/sajjad/data/resdb")
set(CMAKE_PREFIX_PATH "${REPO_ROOT}/deps/aws-lambda-cpp/lambda-install")
list(APPEND CMAKE_PREFIX_PATH "${REPO_ROOT}/deps/AWS-SDK")


project(secure_executor LANGUAGES CXX)
find_package(aws-lambda-runtime REQUIRED)
find_package(AWSSDK COMPONENTS core)
add_executable(${PROJECT_NAME} "main.cpp")


include("${REPO_ROOT}/deps/cpr/build/conanbuildinfo.cmake") # Include Conan-generated file
conan_basic_setup(TARGETS)


include_directories("${REPO_ROOT}/deps/crypto")
target_link_libraries(${PROJECT_NAME} "${REPO_ROOT}/deps/crypto/libcryptopp.a")
target_link_libraries(${PROJECT_NAME} AWS::aws-lambda-runtime ${AWSSDK_LINK_LIBRARIES})
target_link_libraries(${PROJECT_NAME} CONAN_PKG::cpr)


aws_lambda_package_target(${PROJECT_NAME})
