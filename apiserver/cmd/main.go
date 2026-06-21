package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"mini-drop/apiserver/internal/client"
	"mini-drop/apiserver/internal/db"
	"mini-drop/apiserver/internal/handler"
	"mini-drop/apiserver/internal/model"
	"mini-drop/apiserver/internal/service"
	"mini-drop/apiserver/internal/storage"
)

func main() {
	dbHost := getEnv("DB_HOST", "postgres")
	dbUser := getEnv("DB_USER", "drop")
	dbPassword := getEnv("DB_PASSWORD", "drop123")
	dbName := getEnv("DB_NAME", "drop")
	dbPort := getEnv("DB_PORT", "5432")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	var err error
	for i := 0; i < 30; i++ {
		log.Printf("尝试连接数据库 (第 %d 次)...", i+1)
		err = db.InitDB(dsn)
		if err == nil {
			break
		}
		log.Printf("数据库连接失败: %v", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	if err := db.GetDB().AutoMigrate(&model.Agent{}, &model.AuditLog{}); err != nil {
		log.Fatalf("迁移表失败: %v", err)
	}
	log.Println("✅ 数据库表迁移完成")

	dropServerAddr := getEnv("DROP_SERVER_ADDR", "drop_server:50051")
	dropClient, err := client.NewDropClient(dropServerAddr)
	if err != nil {
		log.Fatalf("连接 drop_server 失败: %v", err)
	}
	defer dropClient.Close()

	// 初始化 MinIO 存储
	minioEndpoint := getEnv("MINIO_ENDPOINT", "minio:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "minioadmin123")
	minioBucket := "drop-data"

	storageClient, err := storage.NewMinioClient(minioEndpoint, minioAccessKey, minioSecretKey, minioBucket, false)
	if err != nil {
		log.Fatalf("连接 MinIO 失败: %v", err)
	}
	log.Println("✅ MinIO 连接成功")

	taskService := service.NewTaskService(dropClient, db.GetDB(), storageClient)
	taskHandler := handler.NewTaskHandler(taskService)

	agentService := service.NewAgentService(db.GetDB())
	agentHandler := handler.NewAgentHandler(agentService)
	heartbeatHandler := handler.NewHeartbeatHandler(agentService, db.GetDB())

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		api.POST("/tasks", taskHandler.CreateTask)
		api.GET("/tasks/:id", taskHandler.GetTask)
		api.GET("/tasks", taskHandler.ListTasks)

		api.GET("/agents", agentHandler.ListAgents)
		api.GET("/agents/:agent_id/audit", agentHandler.GetAuditLogs)
		api.POST("/heartbeat", heartbeatHandler.Receive)
	}

	port := getEnv("PORT", "8191")
	log.Printf("apiserver 启动在 http://0.0.0.0:%s", port)
	r.Run(":" + port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
