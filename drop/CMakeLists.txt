cmake_minimum_required(VERSION 3.15)
project(drop)

set(CMAKE_CXX_STANDARD 14)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 使用 pkg-config 查找 gRPC
find_package(PkgConfig REQUIRED)
pkg_check_modules(GRPC REQUIRED grpc++ grpc protobuf)

# Proto 文件目录
set(PROTO_DIR ${CMAKE_CURRENT_SOURCE_DIR}/common/proto)

# 收集所有 proto 文件
file(GLOB PROTO_FILES ${PROTO_DIR}/*.proto)

# 生成 proto 的 C++ 代码
foreach(PROTO_FILE ${PROTO_FILES})
    get_filename_component(PROTO_NAME ${PROTO_FILE} NAME_WE)
    set(PROTO_SRC ${PROTO_DIR}/${PROTO_NAME}.pb.cc)
    set(PROTO_GRPC_SRC ${PROTO_DIR}/${PROTO_NAME}.grpc.pb.cc)
    
    if(NOT EXISTS ${PROTO_SRC} OR NOT EXISTS ${PROTO_GRPC_SRC})
        message(STATUS "Generating protobuf files for ${PROTO_NAME}")
        execute_process(
            COMMAND protoc --proto_path=${PROTO_DIR} --cpp_out=${PROTO_DIR} ${PROTO_FILE}
            COMMAND protoc --proto_path=${PROTO_DIR} --grpc_out=${PROTO_DIR}
                    --plugin=protoc-gen-grpc=/usr/bin/grpc_cpp_plugin ${PROTO_FILE}
        )
    endif()
endforeach()

# 收集生成的源文件
file(GLOB PROTO_SRCS ${PROTO_DIR}/*.pb.cc ${PROTO_DIR}/*.grpc.pb.cc)

# 添加头文件搜索路径
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}
    ${PROTO_DIR}
    ${GRPC_INCLUDE_DIRS}
)

# ===== drop_server =====
add_executable(drop_server
    server/main.cpp
    ${PROTO_SRCS}
)

# 链接 gRPC 库（加上 grpc++_reflection）
target_link_libraries(drop_server 
    ${GRPC_LIBRARIES} 
    grpc++_reflection
    pthread
)

# ===== drop_agent =====
add_executable(drop_agent
    agent/main.cpp
    ${PROTO_SRCS}
)

target_link_libraries(drop_agent 
    ${GRPC_LIBRARIES} 
    pthread
)
