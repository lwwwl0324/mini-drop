#include <iostream>
#include <memory>
#include <string>
#include <map>
#include <mutex>
#include <ctime>
#include <grpcpp/grpcpp.h>
#include <grpcpp/ext/proto_server_reflection_plugin.h>
#include "control.grpc.pb.h"
#include "healthcheck.grpc.pb.h"

using grpc::Server;
using grpc::ServerBuilder;
using grpc::ServerContext;
using grpc::Status;

// 任务队列（按 IP 存储）
std::map<std::string, std::vector<healthcheck::TaskDesc>> task_queues_;
std::mutex task_mutex_;

// Control 服务实现
class ControlServiceImpl final : public control::Control::Service {
    Status CreateTask(grpc::ServerContext* context,
                      const control::CreateTaskRequest* request,
                      control::CreateTaskResponse* response) override {
        std::cout << "[Control] 收到创建任务请求，目标 IP: " << request->target_ip() << std::endl;
        
        const auto& hotmethod_task = request->task_desc();
        
        healthcheck::TaskDesc task;
        task.set_task_id("task_" + std::to_string(std::time(nullptr)));
        task.set_task_type(hotmethod_task.task_type());
        task.set_duration_sec(hotmethod_task.sample_argv().duration());
        task.set_sample_hz(hotmethod_task.sample_argv().hz());
        task.set_target_pid(hotmethod_task.sample_argv().pid());
        
        std::lock_guard<std::mutex> lock(task_mutex_);
        task_queues_[request->target_ip()].push_back(task);
        
        std::cout << "[Control] 任务已入队，PID: " << task.target_pid() 
                  << ", 频率: " << task.sample_hz() << "Hz"
                  << ", 时长: " << task.duration_sec() << "秒"
                  << ", 队列长度: " << task_queues_[request->target_ip()].size() << std::endl;
        
        response->set_task_id(task.task_id());
        return Status::OK;
    }
};

// HealthCheck 服务实现
class HealthCheckServiceImpl final : public healthcheck::HealthCheck::Service {
    Status Do(grpc::ServerContext* context,
              const healthcheck::HealthCheckRequest* request,
              healthcheck::HealthCheckResponse* response) override {
        std::cout << "[HealthCheck] 收到心跳，IP: " << request->ip_addr() 
                  << ", 主机名: " << request->host_name() << std::endl;
        
        std::lock_guard<std::mutex> lock(task_mutex_);
        auto& queue = task_queues_[request->ip_addr()];
        
        if (!queue.empty()) {
            healthcheck::TaskDesc task = queue.front();
            queue.erase(queue.begin());
            
            *response->mutable_task_desc() = task;
            response->set_pending(true);
            std::cout << "[HealthCheck] 下发任务: " << task.task_id() 
                      << " (剩余任务: " << queue.size() << ")" << std::endl;
        } else {
            response->set_pending(false);
        }
        
        response->set_status(healthcheck::HealthCheckResponse::SERVING);
        return Status::OK;
    }
};

void RunServer() {
    std::string server_address("0.0.0.0:50051");
    
    // 启用 gRPC 反射（让 grpcurl 能发现服务）
    grpc::reflection::InitProtoReflectionServerBuilderPlugin();
    
    ControlServiceImpl control_service;
    HealthCheckServiceImpl health_service;
    
    ServerBuilder builder;
    builder.AddListeningPort(server_address, grpc::InsecureServerCredentials());
    builder.RegisterService(&control_service);
    builder.RegisterService(&health_service);
    
    std::unique_ptr<Server> server(builder.BuildAndStart());
    std::cout << "drop_server 监听在 " << server_address << std::endl;
    std::cout << "gRPC 反射已启用，可用 grpcurl 调用" << std::endl;
    
    server->Wait();
}

int main() {
    RunServer();
    return 0;
}
