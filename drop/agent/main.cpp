#include <iostream>
#include <memory>
#include <string>
#include <thread>
#include <chrono>
#include <cstdlib>
#include <unistd.h>
#include <sys/wait.h>
#include <grpcpp/grpcpp.h>
#include "healthcheck.grpc.pb.h"

class Agent {
public:
    Agent(const std::string& server_addr, const std::string& hostname, const std::string& ip)
        : server_addr_(server_addr), hostname_(hostname), ip_(ip) {
        
        auto channel = grpc::CreateChannel(server_addr, grpc::InsecureChannelCredentials());
        stub_ = healthcheck::HealthCheck::NewStub(channel);
    }
    
    void Run() {
        std::cout << "drop_agent 启动，连接服务器: " << server_addr_ << std::endl;
        
        while (running_) {
            healthcheck::HealthCheckRequest request;
            request.set_host_name(hostname_);
            request.set_ip_addr(ip_);
            request.set_uid("agent_" + ip_);
            request.set_agent_version("0.2.0");
            
            healthcheck::HealthCheckResponse response;
            grpc::ClientContext context;
            
            auto status = stub_->Do(&context, request, &response);
            if (status.ok()) {
                std::cout << "[Heartbeat] 心跳成功" << std::endl;
                
                if (response.pending()) {
                    const auto& task = response.task_desc();
                    std::cout << "[Task] 收到任务: " << task.task_id() 
                              << ", PID: " << task.target_pid() 
                              << ", " << task.sample_hz() << "Hz, " 
                              << task.duration_sec() << "秒" << std::endl;
                    
                    ExecutePerf(task);
                }
            } else {
                std::cout << "[Heartbeat] 失败: " << status.error_message() << std::endl;
            }
            
            std::this_thread::sleep_for(std::chrono::seconds(5));
        }
    }
    
    void Stop() { running_ = false; }
    
private:
    void ExecutePerf(const healthcheck::TaskDesc& task) {
        std::string data_file = "/tmp/perf_" + task.task_id() + ".data";
        std::string cmd;
        
        // 根据 PID 构造命令（确保 -o 只出现一次）
        if (task.target_pid() == 0) {
            // 采样整个系统
            cmd = "perf record -F " + std::to_string(task.sample_hz()) +
                  " -a -g -o " + data_file +
                  " -- sleep " + std::to_string(task.duration_sec()) + " 2>&1";
        } else {
            // 采样指定进程
            cmd = "perf record -F " + std::to_string(task.sample_hz()) +
                  " -g -p " + std::to_string(task.target_pid()) +
                  " -o " + data_file +
                  " -- sleep " + std::to_string(task.duration_sec()) + " 2>&1";
        }
        
        std::cout << "[Task] 执行: " << cmd << std::endl;
        
        pid_t pid = fork();
        if (pid == 0) {
            execl("/bin/sh", "sh", "-c", cmd.c_str(), nullptr);
            exit(1);
        } else if (pid > 0) {
            int status;
            waitpid(pid, &status, 0);
            
            if (WIFEXITED(status) && WEXITSTATUS(status) == 0) {
                std::cout << "[Task] perf 成功: " << data_file << std::endl;
                
                // 上传到 MinIO
                std::string object_name = task.task_id() + "/perf.data";
                std::string upload_cmd = "python3 /home/lwl/upload_perf.py drop-data " + object_name + " " + data_file;
                
                std::cout << "[Task] 上传: " << upload_cmd << std::endl;
                
                int upload_status = system(upload_cmd.c_str());
                if (upload_status == 0) {
                    std::cout << "[Task] 上传成功: " << object_name << std::endl;
                } else {
                    std::cout << "[Task] 上传失败，退出码: " << upload_status << std::endl;
                }
            } else {
                std::cout << "[Task] perf 失败，退出码: " << WEXITSTATUS(status) << std::endl;
            }
        } else {
            std::cout << "[Task] fork 失败" << std::endl;
        }
    }
    
    std::string server_addr_;
    std::string hostname_;
    std::string ip_;
    std::unique_ptr<healthcheck::HealthCheck::Stub> stub_;
    bool running_ = true;
};

int main(int argc, char** argv) {
    std::string server_addr = "localhost:50051";
    std::string hostname = "ubuntu-vm";
    std::string ip = "127.0.0.1";
    
    if (argc > 1) server_addr = argv[1];
    if (argc > 2) hostname = argv[2];
    if (argc > 3) ip = argv[3];
    
    Agent agent(server_addr, hostname, ip);
    agent.Run();
    
    return 0;
}
