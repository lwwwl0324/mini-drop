#include <iostream>
#include <memory>
#include <string>
#include <thread>
#include <chrono>
#include <cstdlib>
#include <unistd.h>
#include <sys/wait.h>
#include <signal.h>
#include <grpcpp/grpcpp.h>
#include "healthcheck.grpc.pb.h"
#include "hotmethod.pb.h"

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
            request.set_agent_version("0.3.0");
            
            healthcheck::HealthCheckResponse response;
            grpc::ClientContext context;
            
            auto status = stub_->Do(&context, request, &response);
            if (status.ok()) {
                std::cout << "[Heartbeat] 心跳成功" << std::endl;
                
                if (response.pending()) {
                    const auto& task = response.task_desc();
                    std::cout << "[Task] 收到任务: " << task.task_id() 
                              << ", Profiler: " << task.profiler_type()
                              << ", PID: " << task.sample_argv().pid() 
                              << ", " << task.sample_argv().hz() << "Hz, " 
                              << task.sample_argv().duration() << "秒" << std::endl;
                    
                    if (task.profiler_type() == 1) {
                        ExecuteBpftrace(task);
                    } else {
                        ExecutePerf(task);
                    }
                }
            } else {
                std::cout << "[Heartbeat] 失败: " << status.error_message() << std::endl;
            }
            
            std::this_thread::sleep_for(std::chrono::seconds(5));
        }
    }
    
    void Stop() { running_ = false; }
    
private:
    void ExecutePerf(const hotmethod::TaskDesc& task) {
        std::string data_file = "/tmp/perf_" + task.task_id() + ".data";
        std::string cmd;
        
        if (task.sample_argv().pid() == 0) {
            cmd = "perf record -F " + std::to_string(task.sample_argv().hz()) +
                  " -a -g -o " + data_file +
                  " -- sleep " + std::to_string(task.sample_argv().duration()) + " 2>&1";
        } else {
            cmd = "perf record -F " + std::to_string(task.sample_argv().hz()) +
                  " -g -p " + std::to_string(task.sample_argv().pid()) +
                  " -o " + data_file +
                  " -- sleep " + std::to_string(task.sample_argv().duration()) + " 2>&1";
        }
        
        std::cout << "[Perf] 执行: " << cmd << std::endl;
        
        pid_t pid = fork();
        if (pid == 0) {
            execl("/bin/sh", "sh", "-c", cmd.c_str(), nullptr);
            exit(1);
        } else if (pid > 0) {
            int status;
            waitpid(pid, &status, 0);
            
            if (WIFEXITED(status) && WEXITSTATUS(status) == 0) {
                std::cout << "[Perf] 成功: " << data_file << std::endl;
                
                std::string object_name = task.task_id() + "/perf.data";
                std::string upload_cmd = "python3 /home/lwl/mini-drop/scripts/upload_perf.py drop-data " + object_name + " " + data_file;
                system(upload_cmd.c_str());
            } else {
                std::cout << "[Perf] 失败，退出码: " << WEXITSTATUS(status) << std::endl;
            }
        }
    }
    
    void ExecuteBpftrace(const hotmethod::TaskDesc& task) {
        std::string script_path = "/home/lwl/mini-drop/scripts/io_trace.bt";
        std::string log_file = "/tmp/bpftrace_" + task.task_id() + ".log";
        std::string duration = std::to_string(task.sample_argv().duration());
        
        std::string cmd = "sudo timeout " + duration + "s bpftrace " + script_path + " > " + log_file + " 2>&1";
        
        std::cout << "[eBPF] 执行: " << cmd << std::endl;
        
        pid_t pid = fork();
        if (pid == 0) {
            execl("/bin/sh", "sh", "-c", cmd.c_str(), nullptr);
            exit(1);
        } else if (pid > 0) {
            int status;
            waitpid(pid, &status, 0);
            
            if (WIFEXITED(status) && WEXITSTATUS(status) == 0) {
                std::cout << "[eBPF] 采集完成: " << log_file << std::endl;
                
                std::string cat_cmd = "cat " + log_file;
                system(cat_cmd.c_str());
                
                std::string object_name = task.task_id() + "/bpftrace.log";
                std::string upload_cmd = "python3 /home/lwl/mini-drop/scripts/upload_perf.py drop-data " + object_name + " " + log_file;
                system(upload_cmd.c_str());
            } else {
                std::cout << "[eBPF] 失败，退出码: " << WEXITSTATUS(status) << std::endl;
            }
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
