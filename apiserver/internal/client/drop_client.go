package client

import (
	"context"
	"fmt"
	"time"

	"mini-drop/apiserver/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DropClient struct {
	conn   *grpc.ClientConn
	client proto.ControlClient
}

func NewDropClient(addr string) (*DropClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("连接 drop_server 失败: %v", err)
	}

	return &DropClient{
		conn:   conn,
		client: proto.NewControlClient(conn),
	}, nil
}

func (c *DropClient) Close() error {
	return c.conn.Close()
}

func (c *DropClient) CreateTask(ctx context.Context, targetIP, taskID string, pid, duration, frequency int) (string, error) {
	// 构造 gRPC 请求
	req := &proto.CreateTaskRequest{
		TargetIp: targetIP,
		Service:  "hotmethod",
		TaskDesc: &proto.TaskDesc{
			TaskId:     taskID,
			TaskType:   0,
			ProfilerType: 0,
			SampleArgv: &proto.RecordArgv{
				Hz:       uint32(frequency),
				Duration: uint64(duration),
				Pid:      int32(pid),
				Callgraph: "fp",
			},
			TimeoutSec: 30,
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.CreateTask(ctx, req)
	if err != nil {
		return "", fmt.Errorf("gRPC 调用失败: %v", err)
	}

	return resp.TaskId, nil
}
