package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	rocketmq "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/golang-pay-core/config"
	"github.com/golang-pay-core/internal/logger"
	"go.uber.org/zap"
)

// RocketMQClient RocketMQ 客户端封装
type RocketMQClient struct {
	producer rocketmq.Producer
	enabled  bool
}

// NewRocketMQClient 创建 RocketMQ 客户端
func NewRocketMQClient() (*RocketMQClient, error) {
	cfg := config.GetConfig()

	// 检查是否启用 RocketMQ
	if !cfg.RocketMQ.Enabled {
		logger.Logger.Info("RocketMQ 未启用，将使用同步处理")
		return &RocketMQClient{
			enabled: false,
		}, nil
	}

	// 创建生产者配置
	endpoint := fmt.Sprintf("%s:%d", cfg.RocketMQ.Endpoint, cfg.RocketMQ.Port)

	// 构建凭证（如果启用 ACL）
	var creds *credentials.SessionCredentials
	if cfg.RocketMQ.AccessKey != "" && cfg.RocketMQ.AccessSecret != "" {
		creds = &credentials.SessionCredentials{
			AccessKey:    cfg.RocketMQ.AccessKey,
			AccessSecret: cfg.RocketMQ.AccessSecret,
		}
	}

	producerConfig := &rocketmq.Config{
		Endpoint:    endpoint,
		Credentials: creds,
	}

	// 创建生产者选项
	var opts []rocketmq.ProducerOption
	for _, topic := range cfg.RocketMQ.Topics {
		opts = append(opts, rocketmq.WithTopics(topic))
	}

	// 创建生产者
	producer, err := rocketmq.NewProducer(producerConfig, opts...)
	if err != nil {
		return nil, fmt.Errorf("创建 RocketMQ 生产者失败: %w", err)
	}

	// 启动生产者
	if err := producer.Start(); err != nil {
		return nil, fmt.Errorf("启动 RocketMQ 生产者失败: %w", err)
	}

	logger.Logger.Info("RocketMQ 生产者启动成功",
		zap.String("endpoint", endpoint),
		zap.String("producer_group", cfg.RocketMQ.ProducerGroup),
		zap.Strings("topics", cfg.RocketMQ.Topics))

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

// Close 关闭客户端
func (c *RocketMQClient) Close() error {
	if !c.enabled {
		return nil
	}

	if c.producer != nil {
		if err := c.producer.GracefulStop(); err != nil {
			return fmt.Errorf("关闭 RocketMQ 生产者失败: %w", err)
		}
	}

	logger.Logger.Info("RocketMQ 生产者已关闭")
	return nil
}

// IsEnabled 检查是否启用
func (c *RocketMQClient) IsEnabled() bool {
	return c.enabled
}
