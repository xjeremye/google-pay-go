// @title           支付系统核心 API
// @version         1.0
// @description     基于 Golang 开发的高并发支付系统核心 API
// @termsOfService  https://example.com/terms/

// @contact.name   API Support
// @contact.url    https://example.com/support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @schemes   http https
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
	"github.com/golang-pay-core/internal/mq"
	_ "github.com/golang-pay-core/internal/plugin/alipay" // 导入以触发自动注册（包含 alipay_mock）
	"github.com/golang-pay-core/internal/router"
	"github.com/golang-pay-core/internal/service"
	"go.uber.org/zap"

	_ "github.com/golang-pay-core/docs" // Swagger 文档
)

func main() {
	// 加载配置
	// 支持通过环境变量 APP_ENV 或命令行参数指定环境
	// 环境变量优先级: 命令行参数 > 环境变量 APP_ENV > 默认 dev
	configPath := ""
	if len(os.Args) > 1 {
		// 支持命令行参数: ./app --config=config/config.prod.yaml
		// 或: ./app prod (自动选择 config.prod.yaml)
		arg := os.Args[1]
		if arg == "prod" || arg == "production" {
			configPath = "config/config.prod.yaml"
		} else if arg == "test" || arg == "testing" {
			configPath = "config/config.test.yaml"
		} else if arg == "dev" || arg == "development" {
			configPath = "config/config.yaml"
		} else if len(arg) > 0 && arg[0] != '-' {
			// 如果参数不是以 - 开头，可能是配置文件路径
			configPath = arg
		}
	}

	if err := config.Load(configPath); err != nil {
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

	// 插件已通过 init() 函数自动注册（导入 alipay 包时触发）
	// 所有插件的注册逻辑都在各自的包中管理，保持 main.go 的简洁性
	logger.Logger.Info("插件系统已初始化")

	refreshCtx := context.Background()

	// 启动通知重试服务（每30秒检查一次失败的通知并重试）
	notifyRetryService := service.NewNotifyRetryService()
	go notifyRetryService.Start(refreshCtx)
	logger.Logger.Info("通知重试服务已启动（每30秒检查一次失败的通知）")

	// 初始化全局 RocketMQ 生产者客户端（单例模式，避免重复创建）
	mqProducer := mq.GetGlobalMQClient()
	if mqProducer.IsEnabled() {
		logger.Logger.Info("RocketMQ 生产者已启动")
		defer func() {
			if err := mqProducer.Close(); err != nil {
				logger.Logger.Error("关闭 RocketMQ 生产者失败", zap.Error(err))
			}
		}()
	}

	// 初始化 RocketMQ 消费者（如果启用）
	mqConsumer, err := mq.NewRocketMQConsumer()
	if err != nil {
		logger.Logger.Warn("初始化 RocketMQ 消费者失败",
			zap.Error(err))
	} else if mqConsumer.IsEnabled() {
		logger.Logger.Info("RocketMQ 消费者已启动")
		defer func() {
			if err := mqConsumer.Close(); err != nil {
				logger.Logger.Error("关闭 RocketMQ 消费者失败", zap.Error(err))
			}
		}()
	}

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
