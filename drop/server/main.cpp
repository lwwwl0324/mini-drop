#include <iostream>
#include <memory>
#include <string>
#include <map>
#include <mutex>
#include <ctime>
#include <chrono>
#include <thread>
#include <curl/curl.h>
#include <grpcpp/grpcpp.h>
#include "control.grpc.pb.h"
#include "healthcheck.grpc.pb.h"

using grpc::Server;
using grpc::ServerBuilder;
using grpc::ServerContext;
using grpc::Status;

// 任务队列
std::map<std::string, std::vector<hotmethod::TaskDesc>> task_queues_;
std::mutex task_mutex_;

// Agent 心跳记录
struct AgentInfo {
    std::string hostname;
    std::string ip;
    std::string uid;
    std::string version;
    std::chrono::steady_clock::time_point last_heartbeat;
    bool online;
};
std::map<std::string, AgentInfo> agents_;
std::mutex agent_mutex_;

// 回调 apiserver 发送心跳
void SendHeartbeatToApiserver(const std::string& agent_id, const std::string& hostname, 
                               const std::string& ip, const std::string& version) {
    CURL* curl = curl_easy_init();
    if (!curl) return;
    
    std::string url = "http://localhost:8191/api/v1/heartbeat";
    std::string json = "{\"agent_id\":\"" + agent_id + "\",\"hostname\":\"" + hostname + 
                        "\",\"ip_addr\":\"" + ip + "\",\"version\":\"" + version + "\"}";
    
    struct curl_slist* headers = NULL;
    headers = curl_slist_append(headers, "Content-Type: application/json");
    
    curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json.c_str());
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(curl, CURLOPT_TIMEOUT, 2L);
    
    CURLcode res = curl_easy_perform(curl);
    if (res != CURLE_OK) {
        // 静默失败，不影响主流程
    }
    
    curl_slist_free_all(headers);
    curl_easy_cleanup(curl);
}

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
                  << ", 主机名: " << request->host_name() 
                  << ", UID: " << request->uid() << std::endl;
        
        // 更新 Agent 心跳信息
        {
            std::lock_guard<std::mutex> lock(agent_mutex_);
            auto& agent = agents_[request->uid()];
            bool was_offline = !agent.online;
            agent.hostname = request->host_name();
            agent.ip = request->ip_addr();
            agent.uid = request->uid();
            agent.version = request->agent_version();
            agent.last_heartbeat = std::chrono::steady_clock::now();
            
            if (!agent.online) {
                agent.online = true;
                std::cout << "[Agent] " << request->uid() << " 上线" << std::endl;
            }
            
            // 回调 apiserver
            SendHeartbeatToApiserver(request->uid(), request->host_name(), 
                                      request->ip_addr(), request->agent_version());
        }
        
        // 检查是否有任务待执行
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

// 检查 Agent 离线状态
void CheckAgentOffline() {
    while (true) {
        std::this_thread::sleep_for(std::chrono::seconds(10));
        
        auto now = std::chrono::steady_clock::now();
        std::lock_guard<std::mutex> lock(agent_mutex_);
        
        for (auto& pair : agents_) {
            auto& agent = pair.second;
            auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(now - agent.last_heartbeat).count();
            
            if (agent.online && elapsed > 30) {
                agent.online = false;
                std::cout << "[Agent] " << agent.uid << " 离线 (上次心跳 " << elapsed << " 秒前)" << std::endl;
            }
        }
    }
}

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
    
    // 启动离线检查线程
    std::thread offline_thread(CheckAgentOffline);
    offline_thread.detach();
    
    server->Wait();
}

int main() {
    // 初始化 curl
    curl_global_init(CURL_GLOBAL_DEFAULT);
    RunServer();
    curl_global_cleanup();
    return 0;
}
