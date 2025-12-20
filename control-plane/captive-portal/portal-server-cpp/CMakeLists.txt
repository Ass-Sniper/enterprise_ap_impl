cmake_minimum_required(VERSION 3.10)
project(portal_server_cpp CXX)

set(CMAKE_CXX_STANDARD 11)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

add_executable(portal
    src/main.cpp
    src/http_server.cpp
    src/session_store.cpp
    src/hmac_sha256.cpp
    src/util.cpp
)

# 依赖：restbed + openssl
# Ubuntu: sudo apt-get install -y librestbed-dev libssl-dev
target_link_libraries(portal restbed ssl crypto pthread)
