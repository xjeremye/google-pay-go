package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	rocketmq "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/golang-pay-core/config"
	"github.com/golang-pay-core/internal/alipay"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/plugin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RocketMQConsumer RocketMQ 消费者
type RocketMQConsumer struct {
	consumer rocketmq.PushConsumer
	enabled  bool
}

// NewRocketMQConsumer 创建 RocketMQ 消费者
func NewRocketMQConsumer() (*RocketMQConsumer, error) {
	cfg := config.GetConfig()

	// 检查是否启用 RocketMQ
	if !cfg.RocketMQ.Enabled {
		logger.Logger.Info("RocketMQ 未启用，消费者不会启动")
		return &RocketMQConsumer{
			enabled: false,
		}, nil
	}

	// 创建消费者配置
	endpoint := fmt.Sprintf("%s:%d", cfg.RocketMQ.Endpoint, cfg.RocketMQ.Port)

	// 构建凭证（如果启用 ACL）
	var creds *credentials.SessionCredentials
	if cfg.RocketMQ.AccessKey != "" && cfg.RocketMQ.AccessSecret != "" {
		creds = &credentials.SessionCredentials{
			AccessKey:    cfg.RocketMQ.AccessKey,
			AccessSecret: cfg.RocketMQ.AccessSecret,
		}
	}

	consumerConfig := &rocketmq.Config{
		Endpoint:      endpoint,
		ConsumerGroup: cfg.RocketMQ.ConsumerGroup,
		Credentials:   creds,
	}

	// 创建消息监听器（使用 FuncMessageListener）
	listener := &rocketmq.FuncMessageListener{
		Consume: func(message *rocketmq.MessageView) rocketmq.ConsumerResult {
			ctx := context.Background()
			topic := message.GetTopic()

			var err error
			switch topic {
			case "callback-submit":
				err = handleCallbackSubmitMessages(ctx, message)
			case "order-notify":
				err = handleOrderNotifyMessages(ctx, message)
			case "day-statistics":
				err = handleDayStatisticsMessages(ctx, message)
			case "alipay-notify":
				err = handleAlipayNotifyMessages(ctx, message)
			default:
				logger.Logger.Warn("未知的主题",
					zap.String("topic", topic),
					zap.String("message_id", message.GetMessageId()))
				return rocketmq.SUCCESS
			}

			if err != nil {
				logger.Logger.Error("处理消息失败",
					zap.String("topic", topic),
					zap.String("message_id", message.GetMessageId()),
					zap.Error(err))
				// 返回失败，消息会被重试
				return rocketmq.SUCCESS // 暂时返回成功，避免无限重试
			}

			return rocketmq.SUCCESS
		},
	}

	// 创建消费者（需要在创建时指定 MessageListener 和订阅表达式）
	consumer, err := rocketmq.NewPushConsumer(consumerConfig,
		rocketmq.WithPushSubscriptionExpressions(map[string]*rocketmq.FilterExpression{
			"callback-submit": rocketmq.SUB_ALL,
			"order-notify":    rocketmq.SUB_ALL,
			"day-statistics":  rocketmq.SUB_ALL,
			"alipay-notify":   rocketmq.SUB_ALL,
		}),
		rocketmq.WithPushMessageListener(listener),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 RocketMQ 消费者失败: %w", err)
	}

	// 启动消费者
	if err := consumer.Start(); err != nil {
		return nil, fmt.Errorf("启动 RocketMQ 消费者失败: %w", err)
	}

	logger.Logger.Info("RocketMQ 消费者启动成功",
		zap.String("endpoint", endpoint),
		zap.String("consumer_group", cfg.RocketMQ.ConsumerGroup))

	return &RocketMQConsumer{
		consumer: consumer,
		enabled:  true,
	}, nil
}

// handleCallbackSubmitMessages 处理 callback_submit 消息
func handleCallbackSubmitMessages(ctx context.Context, msg *rocketmq.MessageView) error {
	// 创建插件管理器（避免循环依赖，直接创建）
	pluginMgr := plugin.NewManager(database.RDB)

	var callbackMsg CallbackSubmitMessage
	if err := json.Unmarshal(msg.GetBody(), &callbackMsg); err != nil {
		logger.Logger.Error("解析 callback_submit 消息失败",
			zap.String("message_id", msg.GetMessageId()),
			zap.Error(err))
		return err
	}

	// 构建回调请求
	callbackReq := &plugin.CallbackSubmitRequest{
		OrderNo:        callbackMsg.OrderNo,
		OutOrderNo:     callbackMsg.OutOrderNo,
		PluginID:       callbackMsg.PluginID,
		Tax:            callbackMsg.Tax,
		PluginType:     callbackMsg.PluginType,
		Money:          callbackMsg.Money,
		DomainID:       callbackMsg.DomainID,
		NotifyMoney:    callbackMsg.NotifyMoney,
		OrderID:        callbackMsg.OrderID,
		ProductID:      callbackMsg.ProductID,
		CookieID:       callbackMsg.CookieID,
		ChannelID:      callbackMsg.ChannelID,
		MerchantID:     callbackMsg.MerchantID,
		WriteoffID:     callbackMsg.WriteoffID,
		TenantID:       callbackMsg.TenantID,
		CreateDatetime: callbackMsg.CreateDatetime,
		NotifyURL:      callbackMsg.NotifyURL,
		PluginUpstream: callbackMsg.PluginUpstream,
	}

	// 构建一个简单的 OrderContext 用于获取插件
	orderCtx := &simpleOrderContext{
		pluginID:   callbackMsg.PluginID,
		pluginType: callbackMsg.PluginType,
		channelID:  callbackMsg.ChannelID,
	}

	pluginInstance, err := pluginMgr.GetPluginByCtx(ctx, orderCtx)
	if err != nil {
		logger.Logger.Error("获取插件实例失败",
			zap.String("order_no", callbackMsg.OrderNo),
			zap.Int64("plugin_id", callbackMsg.PluginID),
			zap.String("plugin_type", callbackMsg.PluginType),
			zap.Error(err))
		return err
	}

	// 调用插件的 callback_submit 方法
	if err := pluginInstance.CallbackSubmit(ctx, callbackReq); err != nil {
		logger.Logger.Error("插件 callback_submit 执行失败",
			zap.String("order_no", callbackMsg.OrderNo),
			zap.String("plugin_type", callbackMsg.PluginType),
			zap.Error(err))
		// 注意：这里不返回错误，避免消息被重复消费
		// 如果处理失败，可以通过日志和监控系统告警
	} else {
		logger.Logger.Info("插件 callback_submit 执行成功",
			zap.String("order_no", callbackMsg.OrderNo),
			zap.String("plugin_type", callbackMsg.PluginType))
	}

	return nil
}

// handleAlipayNotifyMessages 处理支付宝回调消息
func handleAlipayNotifyMessages(ctx context.Context, msg *rocketmq.MessageView) error {
	var notifyMsg AlipayNotifyMessage
	if err := json.Unmarshal(msg.GetBody(), &notifyMsg); err != nil {
		logger.Logger.Error("解析 alipay_notify 消息失败",
			zap.String("message_id", msg.GetMessageId()),
			zap.Error(err))
		return err
	}

	// 获取产品信息（用于验证签名）
	product, err := alipay.GetAlipayProductByID(notifyMsg.ProductID)
	if err != nil {
		logger.Logger.Error("获取产品信息失败",
			zap.String("product_id", notifyMsg.ProductID),
			zap.Error(err))
		return err
	}

	// 创建支付宝客户端（用于验证签名）
	alipayClient, err := alipay.NewClient(product, "", false)
	if err != nil {
		logger.Logger.Error("创建支付宝客户端失败",
			zap.String("product_id", notifyMsg.ProductID),
			zap.Error(err))
		return err
	}

	// 验证签名（使用消息中的原始参数）
	if !alipay.VerifyNotify(notifyMsg.Params, alipayClient.AlipayPublicKey) {
		logger.Logger.Warn("支付宝回调签名验证失败",
			zap.String("product_id", notifyMsg.ProductID),
			zap.String("out_trade_no", notifyMsg.NotifyData.OutTradeNo))
		return fmt.Errorf("签名验证失败")
	}

	// 查询订单（通过 out_trade_no，即商户订单号）
	var order models.Order
	if err := database.DB.Where("out_order_no = ?", notifyMsg.NotifyData.OutTradeNo).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Logger.Warn("订单不存在",
				zap.String("out_order_no", notifyMsg.NotifyData.OutTradeNo))
			return nil // 订单不存在，不返回错误，避免重试
		}
		logger.Logger.Error("查询订单失败",
			zap.String("out_order_no", notifyMsg.NotifyData.OutTradeNo),
			zap.Error(err))
		return err
	}

	// 查询订单详情
	var orderDetail models.OrderDetail
	if err := database.DB.Where("order_id = ?", order.ID).First(&orderDetail).Error; err != nil {
		logger.Logger.Error("查询订单详情失败",
			zap.String("order_id", order.ID),
			zap.Error(err))
		return err
	}

	// 验证产品ID是否匹配
	if orderDetail.ProductID != notifyMsg.ProductID {
		logger.Logger.Warn("产品ID不匹配",
			zap.String("order_id", order.ID),
			zap.String("order_product_id", orderDetail.ProductID),
			zap.String("notify_product_id", notifyMsg.ProductID))
		return nil // 产品ID不匹配，不返回错误
	}

	// 验证金额是否匹配（如果回调中有金额）
	if notifyMsg.NotifyData.TotalAmount > 0 && notifyMsg.NotifyData.TotalAmount != order.Money {
		logger.Logger.Warn("金额不匹配",
			zap.String("order_id", order.ID),
			zap.Int("order_money", order.Money),
			zap.Int("notify_amount", notifyMsg.NotifyData.TotalAmount))
		// 金额不匹配，但不阻止处理（可能是部分退款等情况）
	}

	// 检查订单状态，避免重复处理
	if order.OrderStatus == models.OrderStatusPaid {
		if notifyMsg.NotifyData.TradeStatus == "TRADE_SUCCESS" || notifyMsg.NotifyData.TradeStatus == "TRADE_FINISHED" {
			logger.Logger.Info("订单已处理，跳过重复回调",
				zap.String("order_id", order.ID),
				zap.String("trade_status", notifyMsg.NotifyData.TradeStatus))
			// 仍然更新 ticket_no（如果还没有）
			if notifyMsg.NotifyData.TradeNo != "" && orderDetail.TicketNo == "" {
				database.DB.Model(&models.OrderDetail{}).
					Where("order_id = ?", order.ID).
					Update("ticket_no", notifyMsg.NotifyData.TradeNo)
			}
			return nil
		}
	}

	// 根据交易状态更新订单
	var newStatus int
	switch notifyMsg.NotifyData.TradeStatus {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		newStatus = models.OrderStatusPaid
	case "TRADE_CLOSED":
		newStatus = models.OrderStatusFailed
	default:
		logger.Logger.Info("未处理的交易状态",
			zap.String("order_id", order.ID),
			zap.String("trade_status", notifyMsg.NotifyData.TradeStatus))
		return nil
	}

	// 更新订单状态（使用独立的处理函数，避免循环依赖）
	if err := updateOrderStatusDirectly(ctx, order.ID, newStatus, notifyMsg.NotifyData.TradeNo); err != nil {
		logger.Logger.Error("更新订单状态失败",
			zap.String("order_id", order.ID),
			zap.Int("status", newStatus),
			zap.Error(err))
		return err
	}

	// 通知商户（如果订单成功）
	// 发送订单通知消息到队列（而不是直接调用 service）
	if newStatus == models.OrderStatusPaid {
		// 重新查询订单和详情（获取最新状态）
		var updatedOrder models.Order
		var updatedDetail models.OrderDetail
		if err := database.DB.Where("id = ?", order.ID).First(&updatedOrder).Error; err == nil {
			if err := database.DB.Where("order_id = ?", order.ID).First(&updatedDetail).Error; err == nil {
				// 发送订单通知消息到队列
				orderNotifyMsg := &OrderNotifyMessage{
					OrderID:    updatedOrder.ID,
					OrderNo:    updatedOrder.OrderNo,
					OutOrderNo: updatedOrder.OutOrderNo,
					Money:      updatedOrder.Money,
					Status:     updatedOrder.OrderStatus,
					TicketNo:   updatedDetail.TicketNo,
					NotifyURL:  updatedDetail.NotifyURL,
					Timestamp:  time.Now().Unix(),
					RetryCount: 0,
				}
				// 发送到订单通知队列（由另一个消费者处理）
				// 注意：这里需要获取 mqClient，但为了避免循环依赖，暂时先记录日志
				// 实际应该通过消息队列发送
				logger.Logger.Info("订单支付成功，需要通知商户",
					zap.String("order_id", updatedOrder.ID),
					zap.String("notify_url", updatedDetail.NotifyURL),
					zap.Any("notify_message", orderNotifyMsg))
				// TODO: 发送订单通知消息到 order-notify 主题
				// 可以通过创建一个全局的 mqClient 实例，或者通过依赖注入的方式
			}
		}
	}

	logger.Logger.Info("支付宝回调处理成功",
		zap.String("order_id", order.ID),
		zap.String("out_order_no", notifyMsg.NotifyData.OutTradeNo),
		zap.String("trade_no", notifyMsg.NotifyData.TradeNo),
		zap.String("trade_status", notifyMsg.NotifyData.TradeStatus),
		zap.Int("new_status", newStatus))

	return nil
}

// simpleOrderContext 简单的订单上下文（用于消费者获取插件）
type simpleOrderContext struct {
	pluginID   int64
	pluginType string
	channelID  int64
}

func (o *simpleOrderContext) GetOutOrderNo() string   { return "" }
func (o *simpleOrderContext) GetNotifyURL() string    { return "" }
func (o *simpleOrderContext) GetMoney() int           { return 0 }
func (o *simpleOrderContext) GetJumpURL() string      { return "" }
func (o *simpleOrderContext) GetNotifyMoney() int     { return 0 }
func (o *simpleOrderContext) GetExtra() string        { return "" }
func (o *simpleOrderContext) GetCompatible() int      { return 0 }
func (o *simpleOrderContext) GetTest() bool           { return false }
func (o *simpleOrderContext) GetMerchantID() int64    { return 0 }
func (o *simpleOrderContext) GetTenantID() int64      { return 0 }
func (o *simpleOrderContext) GetChannelID() int64     { return o.channelID }
func (o *simpleOrderContext) GetPluginID() int64      { return o.pluginID }
func (o *simpleOrderContext) GetPluginType() string   { return o.pluginType }
func (o *simpleOrderContext) GetPluginUpstream() int  { return 0 }
func (o *simpleOrderContext) GetDomainID() *int64     { return nil }
func (o *simpleOrderContext) GetDomainURL() string    { return "" }
func (o *simpleOrderContext) GetOrderNo() string      { return "" }
func (o *simpleOrderContext) SetOrderNo(no string)    {}
func (o *simpleOrderContext) SetDomainID(id int64)    {}
func (o *simpleOrderContext) SetDomainURL(url string) {}

// handleOrderNotifyMessages 处理订单通知消息
// 注意：这里只记录日志，实际通知逻辑应该在 service 层处理
// 避免循环依赖，这里只做消息解析和日志记录
func handleOrderNotifyMessages(ctx context.Context, msg *rocketmq.MessageView) error {
	var notifyMsg OrderNotifyMessage
	if err := json.Unmarshal(msg.GetBody(), &notifyMsg); err != nil {
		logger.Logger.Error("解析 order_notify 消息失败",
			zap.String("message_id", msg.GetMessageId()),
			zap.Error(err))
		return err
	}

	logger.Logger.Info("收到订单通知消息",
		zap.String("order_id", notifyMsg.OrderID),
		zap.String("order_no", notifyMsg.OrderNo),
		zap.Int("status", notifyMsg.Status))

	// TODO: 实际的通知处理逻辑应该在 service 层实现
	// 这里可以通过消息队列的延迟重试机制来处理失败的通知

	return nil
}

// handleDayStatisticsMessages 处理日统计数据更新消息
func handleDayStatisticsMessages(ctx context.Context, msg *rocketmq.MessageView) error {
	var statsMsg DayStatisticsMessage
	if err := json.Unmarshal(msg.GetBody(), &statsMsg); err != nil {
		logger.Logger.Error("解析 day_statistics 消息失败",
			zap.String("message_id", msg.GetMessageId()),
			zap.Error(err))
		return err
	}

	// 调用日统计服务更新数据
	// 这里需要根据实际的统计服务实现
	logger.Logger.Info("处理日统计数据更新",
		zap.String("product_id", statsMsg.ProductID),
		zap.String("statistics_type", statsMsg.StatisticsType),
		zap.String("date", statsMsg.Date))

	// TODO: 实现日统计数据更新逻辑
	// 可以调用 plugin/alipay/statistics.go 中的服务

	return nil
}

// Close 关闭消费者
func (c *RocketMQConsumer) Close() error {
	if !c.enabled {
		return nil
	}

	if c.consumer != nil {
		if err := c.consumer.GracefulStop(); err != nil {
			return fmt.Errorf("关闭 RocketMQ 消费者失败: %w", err)
		}
	}

	logger.Logger.Info("RocketMQ 消费者已关闭")
	return nil
}

// IsEnabled 检查是否启用
func (c *RocketMQConsumer) IsEnabled() bool {
	return c.enabled
}
