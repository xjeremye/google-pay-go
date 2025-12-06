package router

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-pay-core/config"
	"github.com/golang-pay-core/internal/controller"
	"github.com/golang-pay-core/internal/middleware"
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

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": config.Cfg.App.Name,
			"version": config.Cfg.App.Version,
		})
	})

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

