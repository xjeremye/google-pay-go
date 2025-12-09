package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	rocketmq "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/golang-pay-core/config"
	"github.com/golang-pay-core/internal/logger"
	"go.uber.org/zap"
)

// setupRocketMQLogger 配置 RocketMQ SDK 使用控制台输出日志
// 注意：此函数已不再使用，配置逻辑已移到 init() 和 redirectRocketMQLogs()
// 保留此函数仅用于向后兼容（如果其他地方有调用）
func setupRocketMQLogger() {
	os.Setenv("mq.consoleAppender.enabled", "false")
	if os.Getenv("rocketmq.client.logLevel") == "" {
		os.Setenv("rocketmq.client.logLevel", "WARN")
	}
	rocketmq.ResetLogger()
}

// redirectRocketMQLogs 确保 RocketMQ SDK 的日志配置已应用
// 由于 RocketMQ SDK 已配置为输出到控制台，而我们的 logger 也输出到控制台（当配置为 stdout 时）
// 日志会自然合并，不需要额外的重定向
func redirectRocketMQLogs() {
	cfg := config.GetConfig()

	// 确保配置已应用（如果之前没有设置）
	if os.Getenv("mq.consoleAppender.enabled") != "true" {
		setupRocketMQLogger()
	}

	// 根据配置更新日志级别（如果配置了）
	if cfg != nil && cfg.RocketMQ.LogLevel != "" {
		currentLevel := os.Getenv("rocketmq.client.logLevel")
		if currentLevel != cfg.RocketMQ.LogLevel {
			os.Setenv("rocketmq.client.logLevel", cfg.RocketMQ.LogLevel)
			rocketmq.ResetLogger()
		}
	}

	if logger.Logger != nil {
		logLevel := os.Getenv("rocketmq.client.logLevel")
		if logLevel == "" {
			logLevel = "WARN"
		}
		logger.Logger.Debug("RocketMQ SDK 日志配置已应用",
			zap.String("source", "rocketmq"),
			zap.String("log_level", logLevel))
	}
}

func init() {
	// init 函数中不能访问 config，因为 config 可能还未加载
	// 所以先设置默认值，在 redirectRocketMQLogs 中会根据配置更新
	os.Setenv("mq.consoleAppender.enabled", "true")
	if os.Getenv("rocketmq.client.logLevel") == "" {
		os.Setenv("rocketmq.client.logLevel", "WARN")
	}
	rocketmq.ResetLogger()
}

var (
	// globalMQClient 全局 RocketMQ 生产者客户端实例（单例模式）
	globalMQClient *RocketMQClient
	// globalMQClientInit 用于确保全局客户端只初始化一次
	globalMQClientInit sync.Once
)

// RocketMQClient RocketMQ 客户端封装
type RocketMQClient struct {
	producer rocketmq.Producer
	enabled  bool
}

// GetGlobalMQClient 获取全局 RocketMQ 客户端实例（单例模式）
func GetGlobalMQClient() *RocketMQClient {
	globalMQClientInit.Do(func() {
		client, err := NewRocketMQClient()
		if err != nil {
			if logger.Logger != nil {
				logger.Logger.Warn("初始化全局 RocketMQ 客户端失败", zap.Error(err))
			}
			globalMQClient = &RocketMQClient{enabled: false}
		} else {
			globalMQClient = client
		}
	})
	return globalMQClient
}

// NewRocketMQClient 创建 RocketMQ 客户端
func NewRocketMQClient() (*RocketMQClient, error) {
	cfg := config.GetConfig()

	// 确保 RocketMQ SDK 的日志已重定向到我们的 logger
	redirectRocketMQLogs()

	// 检查是否启用 RocketMQ
	if !cfg.RocketMQ.Enabled {
		if logger.Logger != nil {
			logger.Logger.Info("RocketMQ 未启用，将使用同步处理")
		}
		return &RocketMQClient{
			enabled: false,
		}, nil
	}

	// 创建生产者配置
	endpoint := fmt.Sprintf("%s:%d", cfg.RocketMQ.Endpoint, cfg.RocketMQ.Port)

	// 构建凭证（RocketMQ SDK 要求 Credentials 不能为 nil，即使不使用 ACL 也需要提供）
	// 如果没有配置 AccessKey/AccessSecret，使用空字符串作为默认值
	creds := &credentials.SessionCredentials{
		AccessKey:    cfg.RocketMQ.AccessKey,
		AccessSecret: cfg.RocketMQ.AccessSecret,
	}

	producerConfig := &rocketmq.Config{
		Endpoint:    endpoint,
		Credentials: creds, // 确保不为 nil
	}

	// 创建生产者选项
	var opts []rocketmq.ProducerOption
	for _, topic := range cfg.RocketMQ.Topics {
		opts = append(opts, rocketmq.WithTopics(topic))
	}

	// 创建生产者（使用 defer recover 捕获可能的 panic）
	var producer rocketmq.Producer
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("创建 RocketMQ 生产者时发生 panic: %v", r)
			}
		}()
		producer, err = rocketmq.NewProducer(producerConfig, opts...)
	}()

	if err != nil {
		if logger.Logger != nil {
			logger.Logger.Warn("创建 RocketMQ 生产者失败，将使用同步处理",
				zap.String("endpoint", endpoint),
				zap.String("producer_group", cfg.RocketMQ.ProducerGroup),
				zap.Error(err))
		}
		return &RocketMQClient{
			enabled: false,
		}, nil // 返回禁用状态的客户端，不返回错误
	}

	// 启动生产者（添加超时控制，避免长时间阻塞）
	startErr := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("启动 RocketMQ 生产者时发生 panic: %v", r)
			}
		}()

		// 使用 goroutine + context 实现超时控制
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- producer.Start()
		}()

		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return fmt.Errorf("启动 RocketMQ 生产者超时（10秒）: %w", ctx.Err())
		}
	}()

	if startErr != nil {
		if logger.Logger != nil {
			// 分析错误类型，提供更友好的诊断信息
			errMsg := startErr.Error()
			var suggestion string
			if strings.Contains(errMsg, "context deadline exceeded") || strings.Contains(errMsg, "create grpc conn failed") {
				suggestion = "请检查：1) RocketMQ 服务是否正在运行；2) endpoint 和 port 配置是否正确；3) 网络连接是否正常"
			} else if strings.Contains(errMsg, "topic route") {
				suggestion = "请检查：1) RocketMQ 服务是否正常运行；2) 配置的 topics 是否已在 RocketMQ 中创建"
			}

			logger.Logger.Warn("启动 RocketMQ 生产者失败，将使用同步处理",
				zap.String("endpoint", endpoint),
				zap.String("producer_group", cfg.RocketMQ.ProducerGroup),
				zap.Strings("topics", cfg.RocketMQ.Topics),
				zap.String("suggestion", suggestion),
				zap.Error(startErr))
		}
		// 尝试关闭生产者
		_ = producer.GracefulStop()
		return &RocketMQClient{
			enabled: false,
		}, nil // 返回禁用状态的客户端，不返回错误
	}

	if logger.Logger != nil {
		logger.Logger.Info("RocketMQ 生产者启动成功",
			zap.String("endpoint", endpoint),
			zap.String("producer_group", cfg.RocketMQ.ProducerGroup),
			zap.Strings("topics", cfg.RocketMQ.Topics))
	}

	return &RocketMQClient{
		producer: producer,
		enabled:  true,
	}, nil
}

// SendMessage 发送消息
func (c *RocketMQClient) SendMessage(ctx context.Context, topic, tag string, body interface{}) error {
	if !c.enabled {
		// 如果未启用 RocketMQ，直接返回（调用方应该使用同步处理）
		return fmt.Errorf("RocketMQ 未启用")
	}

	// 序列化消息体
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 构建消息
	message := &rocketmq.Message{
		Topic: topic,
		Body:  bodyBytes,
	}
	if tag != "" {
		message.SetTag(tag)
	}

	// 发送消息（异步发送，不等待结果）
	_, err = c.producer.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	return nil
}

// SendMessageSync 同步发送消息（等待结果）
func (c *RocketMQClient) SendMessageSync(ctx context.Context, topic, tag string, body interface{}) error {
	if !c.enabled {
		return fmt.Errorf("RocketMQ 未启用")
	}

	// 序列化消息体
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 构建消息
	message := &rocketmq.Message{
		Topic: topic,
		Body:  bodyBytes,
	}
	if tag != "" {
		message.SetTag(tag)
	}

	// 同步发送消息
	_, err = c.producer.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	return nil
}

// SendDelayMessage 发送延迟消息
func (c *RocketMQClient) SendDelayMessage(ctx context.Context, topic, tag string, body interface{}, delay time.Duration) error {
	if !c.enabled {
		return fmt.Errorf("RocketMQ 未启用")
	}

	// 序列化消息体
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 构建消息
	message := &rocketmq.Message{
		Topic: topic,
		Body:  bodyBytes,
	}
	if tag != "" {
		message.SetTag(tag)
	}
	// 设置延迟时间（RocketMQ 5.0+ 支持精确延迟时间）
	message.SetDelayTimestamp(time.Now().Add(delay))

	// 发送消息
	_, err = c.producer.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("发送延迟消息失败: %w", err)
	}

	return nil
}

// Close 关闭客户端（添加超时控制，避免长时间阻塞）
func (c *RocketMQClient) Close() error {
	if !c.enabled {
		return nil
	}

	if c.producer != nil {
		// 使用 goroutine + context 实现超时控制
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- c.producer.GracefulStop()
		}()

		select {
		case err := <-done:
			if err != nil {
				if logger.Logger != nil {
					logger.Logger.Error("关闭 RocketMQ 生产者失败", zap.Error(err))
				}
				return fmt.Errorf("关闭 RocketMQ 生产者失败: %w", err)
			}
		case <-ctx.Done():
			if logger.Logger != nil {
				logger.Logger.Warn("关闭 RocketMQ 生产者超时（5秒），强制退出", zap.Error(ctx.Err()))
			}
			// 超时后不返回错误，允许应用继续关闭
			return nil
		}
	}

	if logger.Logger != nil {
		logger.Logger.Info("RocketMQ 生产者已关闭")
	}
	return nil
}

// IsEnabled 检查是否启用
func (c *RocketMQClient) IsEnabled() bool {
	return c.enabled
}
