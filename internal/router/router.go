package router

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-pay-core/config"
	"github.com/golang-pay-core/internal/controller"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 设置运行模式
	gin.SetMode(config.Cfg.App.Mode)

	r := gin.New()

	// 全局中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.Metrics()) // Prometheus 监控中间件

	// Swagger 文档（根据配置决定是否启用）
	if config.Cfg.Monitoring.SwaggerEnabled {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Prometheus 指标端点（需要认证）
	metricsGroup := r.Group("/metrics")
	metricsGroup.Use(middleware.MetricsAuth()) // 添加认证中间件
	metricsGroup.GET("", middleware.PrometheusHandler())

	// 健康检查（增强版）
	r.GET("/health", healthCheck)

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 订单相关路由
		orderController := controller.NewOrderController()
		orders := api.Group("/orders")
		{
			orders.POST("", orderController.CreateOrder)           // 创建订单
			orders.GET("/:order_no", orderController.GetOrder)     // 获取订单
			orders.GET("/query", orderController.QueryOrder)       // 查询订单
		}
	}

	return r
}

// healthCheck 健康检查端点
// @Summary 健康检查
// @Description 检查服务健康状态，包括数据库和 Redis 连接状态
// @Tags 系统
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func healthCheck(c *gin.Context) {
	health := gin.H{
		"status":  "ok",
		"service": config.Cfg.App.Name,
		"version": config.Cfg.App.Version,
		"mode":    config.Cfg.App.Mode,
	}

	// 检查数据库连接
	if database.DB != nil {
		sqlDB, err := database.DB.DB()
		if err == nil {
			if err := sqlDB.Ping(); err == nil {
				stats := sqlDB.Stats()
				health["database"] = gin.H{
					"status":           "ok",
					"open_connections": stats.OpenConnections,
					"in_use":           stats.InUse,
					"idle":             stats.Idle,
					"wait_count":       stats.WaitCount,
				}
			} else {
				health["database"] = gin.H{
					"status": "error",
					"error":  err.Error(),
				}
			}
		}
	}

	// 检查 Redis 连接
	if database.RDB != nil {
		ctx := database.GetContext()
		if err := database.RDB.Ping(ctx).Err(); err == nil {
			poolStats := database.RDB.PoolStats()
			health["redis"] = gin.H{
				"status":      "ok",
				"hits":        poolStats.Hits,
				"misses":      poolStats.Misses,
				"timeouts":    poolStats.Timeouts,
				"total_conns": poolStats.TotalConns,
				"idle_conns":  poolStats.IdleConns,
			}
		} else {
			health["redis"] = gin.H{
				"status": "error",
				"error":  err.Error(),
			}
		}
	}

	c.JSON(200, health)
}

