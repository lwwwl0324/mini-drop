.PHONY: demo
demo:
	@echo "🚀 启动 Mini-Drop 所有服务..."
	docker compose up -d
	@echo "⏳ 等待服务就绪（15秒）..."
	sleep 15
	@echo "✅ 服务已启动，访问 http://localhost"
	@echo "📊 查看任务列表: curl http://localhost:8191/api/v1/tasks"
	@echo "🔥 在 Web 界面创建采样任务即可看到火焰图"
