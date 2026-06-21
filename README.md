# 🔥 Mini-Drop 性能分析平台

一个面向 Linux 服务器的**按需性能采集 + 可视化分析**平台。用户通过 Web 界面下发采样任务，系统自动完成采集、存储、分析和可视化展示。

---

## ✨ 功能特性

- **Web UI**：React + TDesign，支持任务下发、状态查看、火焰图展示
- **Go apiserver**：HTTP API + gRPC 调度 + PostgreSQL 持久化
- **C++ 采集核心**：drop_server + drop_agent，基于 gRPC 通信
- **多采集器支持**：perf (CPU) + eBPF (IO追踪) + py-spy (Python)
- **火焰图生成**：基于 Brendan Gregg 的 FlameGraph 工具链
- **MinIO 存储**：采集数据和火焰图自动上传到对象存储
- **任务状态机**：PENDING → RUNNING → UPLOADING → DONE / FAILED
- **Agent 管理**：在线/离线检测 + 审计日志
- **Docker 化**：一键启动所有服务

---

## ⚙️ 环境要求

| 项             | 要求                                           |
| -------------- | ---------------------------------------------- |
| 操作系统       | Ubuntu 22.04+                                  |
| Linux 内核     | 6.8+（eBPF 需要）                              |
| Docker         | 20.10+                                         |
| Docker Compose | V2（`docker compose` 命令）                    |
| 权限           | 运行 `drop_agent` 需要 `sudo` 或 `cap_perfmon` |
| perf           | 需要设置 `kernel.perf_event_paranoid=1`        |

---

## 🏗️ 系统架构

```markdown
┌─────────────────────────────────────────────────────────────┐
│                    React Web 前端                          │
│              (localhost:5173)                              │
└──────────────────────┬──────────────────────────────────────┘
                       │ HTTP (CORS)
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go apiserver                            │
│              (localhost:8191)                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │   Handler  │  Service  │  Model  │  DB (GORM)      │   │
│  └─────────────────────────────────────────────────────┘   │
└────────┬──────────────────────────────────────┬────────────┘
         │ gRPC                                │ SQL
         ▼                                      ▼
┌─────────────────────────┐        ┌─────────────────────────┐
│      drop_server        │        │    PostgreSQL           │
│     (C++ gRPC Server)   │        │    (任务持久化)          │
│  - 任务队列管理          │        └─────────────────────────┘
│  - Agent 心跳调度        │
└───────────┬─────────────┘
            │ 心跳 + 任务下发
            ▼
┌─────────────────────────┐
│      drop_agent         │
│     (C++ 采集探针)       │
│  - 心跳上报              │
│  - perf / eBPF / py-spy  │
│  - 结果上传              │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│     MinIO 对象存储       │
│  - perf.data            │
│  - flamegraph.svg       │
│  - pyspy.svg            │
└─────────────────────────┘
```

---

## 🚀 一键启动（推荐）

```bash
# 克隆项目
git clone https://github.com/lwwwl0324/mini-drop.git
cd mini-drop

# 一键启动所有服务
make demo

# 查看服务状态
docker compose ps
```

**`make demo` 会自动完成**：

1. 启动 PostgreSQL + MinIO
2. 启动 drop_server + drop_agent
3. 启动 apiserver + web_frontend

---

## 🔧 手动启动

### 1. 启动基础设施（PostgreSQL + MinIO）

```bash
docker compose up -d postgres minio
```

### 2. 编译 C++ 采集核心

```bash
cd drop
mkdir -p build && cd build
cmake ..
make -j$(nproc)
```

### 3. 启动所有服务

| 服务         | 命令                                 | 端口         |
| ------------ | ------------------------------------ | ------------ |
| drop_server  | `./drop_server`                      | 50051 (gRPC) |
| drop_agent   | `sudo ./drop_agent`                  | -            |
| apiserver    | `cd apiserver && go run cmd/main.go` | 8191         |
| web_frontend | `cd web_frontend && npm run dev`     | 5173         |

---

## 🌐 访问地址

| 服务          | 地址                  |
| ------------- | --------------------- |
| Web 界面      | http://localhost:5173 |
| apiserver API | http://localhost:8191 |
| MinIO 控制台  | http://localhost:9001 |
| PostgreSQL    | localhost:5432        |

**MinIO 默认账号**：

- 用户名：`minioadmin`
- 密码：`minioadmin123`

**PostgreSQL 默认账号**：

- 用户名：`drop`
- 密码：`drop123`
- 数据库：`drop`

---

## 📁 目录结构

```
mini-drop/
├── drop/                    # C++ 采集核心
│   ├── agent/               # drop_agent 源码
│   ├── server/              # drop_server 源码
│   ├── common/              # 公共代码 + proto 定义
│   └── CMakeLists.txt
├── apiserver/               # Go 后端服务
│   ├── cmd/                 # 主程序入口
│   ├── internal/            # 内部包
│   │   ├── client/          # gRPC 客户端
│   │   ├── db/              # 数据库连接
│   │   ├── handler/         # HTTP 处理器
│   │   ├── model/           # 数据模型
│   │   └── service/         # 业务逻辑
│   ├── proto/               # gRPC 协议
│   └── go.mod
├── web_frontend/            # React 前端
│   ├── src/
│   │   ├── api/             # API 调用
│   │   ├── components/      # 公共组件
│   │   ├── pages/           # 页面组件
│   │   └── App.jsx
│   └── package.json
├── scripts/                 # Python 工具脚本
│   ├── generate_flamegraph.py
│   └── upload_perf.py
├── docker-compose.yml       # 服务编排
├── Makefile                 # 一键启动
└── README.md
```

---

## 🧪 端到端测试

1. 在 Web 界面点击 **"新建任务"**
2. 填写参数（PID 填 `0` 表示系统全局采样）
3. 选择采集器：`perf (CPU采样)` / `eBPF (IO追踪)` / `py-spy (Python)`
4. 点击 **"创建"**，等待 1-2 秒
5. 任务列表中显示新任务，状态从 `等待中` → `采集中`
6. 约 30 秒后，状态变为 `已完成`，出现 **"查看"** 按钮
7. 点击 **"查看"**，查看火焰图

---

## 📊 API 接口

| 接口                       | 方法 | 功能           |
| -------------------------- | ---- | -------------- |
| `/health`                  | GET  | 健康检查       |
| `/api/v1/tasks`            | POST | 创建任务       |
| `/api/v1/tasks`            | GET  | 任务列表       |
| `/api/v1/tasks/:id`        | GET  | 任务详情       |
| `/api/v1/agents`           | GET  | Agent 列表     |
| `/api/v1/agents/:id/audit` | GET  | Agent 审计日志 |

### 创建任务示例

```bash
curl -X POST http://localhost:8191/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "target_ip": "127.0.0.1",
    "pid": 0,
    "duration": 10,
    "frequency": 999,
    "profiler_type": "perf"
  }'
```

---

## ❓ 故障排查

### perf 权限不足

```bash
sudo sysctl -w kernel.perf_event_paranoid=1
```

### 端口被占用

```bash
# 检查端口占用
sudo lsof -i :9000   # MinIO
sudo lsof -i :5432   # PostgreSQL
sudo lsof -i :8191   # apiserver
```

### 查看服务日志

```bash
docker compose logs -f [service_name]
# 可选: postgres, minio, drop_server, drop_agent, apiserver, web_frontend
```

### 重新构建服务

```bash
docker compose build [service_name]
docker compose up -d [service_name]
```

---

## 📦 依赖清单

| 组件           | 版本   |
| -------------- | ------ |
| Ubuntu         | 22.04+ |
| Docker         | 20.10+ |
| Docker Compose | V2     |
| Go             | 1.18+  |
| Node.js        | 18+    |
| CMake          | 3.15+  |
| gRPC           | 1.48+  |
| Protobuf       | 3.20+  |

---

## 📝 License

MIT

---

## 🙏 致谢

- [FlameGraph](https://github.com/brendangregg/FlameGraph) - Brendan Gregg 的火焰图工具
- [MinIO](https://min.io/) - 高性能对象存储
- [TDesign](https://tdesign.tencent.com/) - 企业级设计系统
- [py-spy](https://github.com/benfred/py-spy) - Python 采样分析器


