FROM ubuntu:22.04

# 安装依赖
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    pkg-config \
    libgrpc++-dev \
    libprotobuf-dev \
    protobuf-compiler \
    protobuf-compiler-grpc \
    libcurl4-openssl-dev \
    && rm -rf /var/lib/apt/lists/*

# 设置工作目录
WORKDIR /app

# 复制源代码
COPY . .

# 编译
RUN mkdir -p build && cd build && \
    cmake .. && \
    make -j$(nproc)

# 暴露端口
EXPOSE 50051

# 启动 server
CMD ["./build/drop_server"]
