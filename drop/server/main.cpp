#include <iostream>
#include <memory>
#include <string>
#include <map>
#include <mutex>
#include <ctime>
#include <grpcpp/grpcpp.h>
#include "control.grpc.pb.h"
#include "healthcheck.grpc.pb.h"

using grpc::Server;
using grpc::ServerBuilder;
using grpc::ServerContext;
using grpc::Status;

std::map<std::string, std::vector<hotmethod::TaskDesc>> task_queues_;
std::mutex task_mutex_;

class ControlServiceImpl final : public control::Control::Service {
    Status CreateTask(grpc::ServerContext* context,
                      const control::CreateTaskRequest* request,
                      control::CreateTaskResponse* response) override {
        std::cout << "[Control] 收到创建任务请求，目标 IP: " << request->target_ip() << std::endl;
        
        hotmethod::TaskDesc task = request->task_desc();
        task.set_task_id("task_" + std::to_string(std::time(nullptr)));
        
        std::lock_guard<std::mutex> lock(task_mutex_);
        task_queues_[request->target_ip()].push_back(task);
        
        std::cout << "[Control] 任务已入队，Profiler: " << task.profiler_type()
                  << ", 队列长度: " << task_queues_[request->target_ip()].size() << std::endl;
        
        response->set_task_id(task.task_id());
        return Status::OK;
    }
};

class HealthCheckServiceImpl final : public healthcheck::HealthCheck::Service {
    Status Do(grpc::ServerContext* context,
              const healthcheck::HealthCheckRequest* request,
              healthcheck::HealthCheckResponse* response) override {
        std::cout << "[HealthCheck] 收到心跳，IP: " << request->ip_addr() 
                  << ", 主机名: " << request->host_name() << std::endl;
        
        std::lock_guard<std::mutex> lock(task_mutex_);
        auto& queue = task_queues_[request->ip_addr()];
        
        if (!queue.empty()) {
            hotmethod::TaskDesc task = queue.front();
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
    
    ControlServiceImpl control_service;
    HealthCheckServiceImpl health_service;
    
    ServerBuilder builder;
    builder.AddListeningPort(server_address, grpc::InsecureServerCredentials());
    builder.RegisterService(&control_service);
    builder.RegisterService(&health_service);
    
    std::unique_ptr<Server> server(builder.BuildAndStart());
    std::cout << "drop_server 监听在 " << server_address << std::endl;
    
    server->Wait();
}

int main() {
    RunServer();
    return 0;
}
