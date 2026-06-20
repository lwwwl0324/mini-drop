package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"mini-drop/apiserver/internal/client"
	"mini-drop/apiserver/internal/db"
	"mini-drop/apiserver/internal/handler"
	"mini-drop/apiserver/internal/service"
)

func main() {
	// 1. 初始化数据库
	dsn := "host=localhost user=drop password=drop123 dbname=drop port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	if err := db.InitDB(dsn); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 2. 连接 drop_server
	dropClient, err := client.NewDropClient("localhost:50051")
	if err != nil {
		log.Fatalf("连接 drop_server 失败: %v", err)
	}
	defer dropClient.Close()

	// 3. 创建服务
	taskService := service.NewTaskService(dropClient, db.GetDB())
	taskHandler := handler.NewTaskHandler(taskService)

	// 4. 启动 HTTP 服务
	r := gin.Default()

	// 5. CORS 配置（允许前端跨域访问）
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
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
	}

	log.Println("apiserver 启动在 http://localhost:8191")
	r.Run(":8191")
}
