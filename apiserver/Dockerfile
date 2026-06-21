FROM golang:1.25-alpine

# 设置 Go 代理（必须在任何 go 命令之前）
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GOSUMDB=sum.golang.google.cn
ENV GO111MODULE=on

# 安装依赖
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN go build -o apiserver cmd/main.go

# 暴露端口
EXPOSE 8191

# 启动
CMD ["./apiserver"]
