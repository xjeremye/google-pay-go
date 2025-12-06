package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-pay-core/config"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/router"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	if err := config.Load("config/config.yaml"); err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 初始化日志
	if err := logger.InitLogger(); err != nil {
		panic(fmt.Sprintf("初始化日志失败: %v", err))
	}
	defer logger.Sync()

	// 初始化数据库
	if err := database.InitMySQL(); err != nil {
		logger.Logger.Fatal("初始化数据库失败", zap.Error(err))
	}
	defer database.CloseMySQL()

	// 初始化 Redis
	if err := database.InitRedis(); err != nil {
		logger.Logger.Warn("初始化 Redis 失败", zap.Error(err))
		// Redis 不是必须的，可以继续运行
	}
	defer database.CloseRedis()

	// 设置路由
	r := router.SetupRouter()

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Cfg.App.Port),
		Handler:        r,
		ReadTimeout:    config.Cfg.App.ReadTimeout,
		WriteTimeout:   config.Cfg.App.WriteTimeout,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// 启动服务器（在 goroutine 中）
	go func() {
		logger.Logger.Info("服务器启动",
			zap.String("address", srv.Addr),
			zap.String("mode", config.Cfg.App.Mode),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info("正在关闭服务器...")

	// 设置 5 秒超时关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("服务器强制关闭", zap.Error(err))
	}

	logger.Logger.Info("服务器已关闭")
}
