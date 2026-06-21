FROM golang:1.25-alpine

# 安装 Python3 和依赖（用于火焰图生成）
RUN apk add --no-cache python3 py3-pip bash

# 设置 Python 软链接
RUN ln -sf /usr/bin/python3 /usr/bin/python

# 安装 Python 依赖
RUN pip3 install minio

# 设置 Go 代理
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GOSUMDB=sum.golang.google.cn
ENV GO111MODULE=on

# 安装 git
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN go build -o apiserver cmd/main.go

# 暴露端口
EXPOSE 8191

# 启动
CMD ["./apiserver"]
