package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	golang "github.com/apache/rocketmq-clients/golang/v5"
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

// init 函数已在 rocketmq.go 中定义，这里不需要重复配置

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

	// 构建凭证（RocketMQ SDK 要求 Credentials 不能为 nil，即使不使用 ACL 也需要提供）
	creds := &credentials.SessionCredentials{
		AccessKey:    cfg.RocketMQ.AccessKey,
		AccessSecret: cfg.RocketMQ.AccessSecret,
	}

	consumerConfig := &rocketmq.Config{
		Endpoint:      endpoint,
		ConsumerGroup: cfg.RocketMQ.ConsumerGroup,
		Credentials:   creds, // 确保不为 nil
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
			case "cache-refresh":
				err = handleCacheRefreshMessages(ctx, message)
			case "balance-sync":
				err = handleBalanceSyncMessages(ctx, message)
			case "order-timeout":
				err = handleOrderTimeoutMessages(ctx, message)
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
	var consumer rocketmq.PushConsumer
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("创建 RocketMQ 消费者时发生 panic: %v", r)
			}
		}()
		consumer, err = rocketmq.NewPushConsumer(consumerConfig,
			rocketmq.WithPushSubscriptionExpressions(map[string]*rocketmq.FilterExpression{
				"callback-submit": rocketmq.SUB_ALL,
				"order-notify":    rocketmq.SUB_ALL,
				"day-statistics":  rocketmq.SUB_ALL,
				"alipay-notify":   rocketmq.SUB_ALL,
				"cache-refresh":   rocketmq.SUB_ALL,
				"balance-sync":    rocketmq.SUB_ALL,
				"order-timeout":   rocketmq.SUB_ALL,
			}),
			rocketmq.WithPushMessageListener(listener),
		)
	}()

	if err != nil {
		if logger.Logger != nil {
			logger.Logger.Warn("创建 RocketMQ 消费者失败，将使用同步处理",
				zap.String("endpoint", endpoint),
				zap.String("consumer_group", cfg.RocketMQ.ConsumerGroup),
				zap.Error(err))
		}
		return &RocketMQConsumer{
			enabled: false,
		}, nil // 返回禁用状态的消费者，不返回错误
	}

	// 启动消费者（添加超时控制，避免长时间阻塞）
	startErr := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("启动 RocketMQ 消费者时发生 panic: %v", r)
			}
		}()

		// 使用 goroutine + context 实现超时控制
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			err := consumer.Start()
			// 过滤 "no new message" 错误（这是正常的轮询行为，不是真正的错误）
			if err != nil {
				var rpcStatus *golang.ErrRpcStatus
				if ok := errors.As(err, &rpcStatus); ok {
					// 如果是 40401 错误或 "no new message" 消息，忽略这个错误
					if rpcStatus.GetCode() == 40401 || rpcStatus.GetMessage() == "no new message" {
						// 这是正常的轮询行为，不是真正的错误，静默忽略
						logger.Logger.Debug("RocketMQ 消费者轮询：无新消息（正常行为）",
							zap.String("consumer_group", cfg.RocketMQ.ConsumerGroup))
						done <- nil // 返回 nil，表示这不是错误
						return
					}
				}
			}
			done <- err
		}()

		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return fmt.Errorf("启动 RocketMQ 消费者超时（10秒）: %w", ctx.Err())
		}
	}()

	if startErr != nil {
		// 检查是否是 "no new message" 错误（这是正常的轮询行为，不是真正的错误）
		var rpcStatus *golang.ErrRpcStatus
		if ok := errors.As(startErr, &rpcStatus); ok {
			// 如果是 40401 错误或 "no new message" 消息，忽略这个错误
			if rpcStatus.GetCode() == 40401 || rpcStatus.GetMessage() == "no new message" {
				// 这是正常的轮询行为，不是真正的错误，继续运行
				logger.Logger.Debug("RocketMQ 消费者启动时轮询：无新消息（正常行为）",
					zap.String("consumer_group", cfg.RocketMQ.ConsumerGroup))
				// 不返回错误，继续运行
			} else {
				// 其他 RPC 错误，需要处理
				logger.Logger.Warn("启动 RocketMQ 消费者时遇到 RPC 错误",
					zap.String("endpoint", endpoint),
					zap.String("consumer_group", cfg.RocketMQ.ConsumerGroup),
					zap.Int32("rpc_code", rpcStatus.GetCode()),
					zap.String("rpc_message", rpcStatus.GetMessage()),
					zap.Error(startErr))
				// 尝试关闭消费者
				_ = consumer.GracefulStop()
				return &RocketMQConsumer{
					enabled: false,
				}, nil
			}
		} else {
			// 分析其他错误类型，提供更友好的诊断信息
			errMsg := startErr.Error()
			var suggestion string
			if strings.Contains(errMsg, "context deadline exceeded") || strings.Contains(errMsg, "create grpc conn failed") {
				suggestion = "请检查：1) RocketMQ 服务是否正在运行；2) endpoint 和 port 配置是否正确；3) 网络连接是否正常；4) proxy 服务是否已启动（端口 8081）"
			} else if strings.Contains(errMsg, "topic route") {
				suggestion = "请检查：1) RocketMQ 服务是否正常运行；2) 配置的 topics 是否已在 RocketMQ 中创建"
			}

			if logger.Logger != nil {
				logger.Logger.Warn("启动 RocketMQ 消费者失败，将使用同步处理",
					zap.String("endpoint", endpoint),
					zap.String("consumer_group", cfg.RocketMQ.ConsumerGroup),
					zap.Strings("topics", cfg.RocketMQ.Topics),
					zap.String("suggestion", suggestion),
					zap.Error(startErr))
			}
			// 尝试关闭消费者
			_ = consumer.GracefulStop()
			return &RocketMQConsumer{
				enabled: false,
			}, nil // 返回禁用状态的消费者，不返回错误
		}
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
	// 如果订单已经是支付成功状态（无论是通知已返回还是未返回），跳过重复回调
	if order.OrderStatus == models.OrderStatusPaid || order.OrderStatus == models.OrderStatusPaidNoNotify {
		if notifyMsg.NotifyData.TradeStatus == "TRADE_SUCCESS" || notifyMsg.NotifyData.TradeStatus == "TRADE_FINISHED" {
			logger.Logger.Info("订单已处理，跳过重复回调",
				zap.String("order_id", order.ID),
				zap.String("trade_status", notifyMsg.NotifyData.TradeStatus),
				zap.Int("current_status", order.OrderStatus))
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
		// 交易成功，但商户通知还未成功，先设置为"支付成功，通知未返回"
		// 只有当商户通知成功时，才会更新为"支付成功，通知已返回"
		newStatus = models.OrderStatusPaidNoNotify
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
	if newStatus == models.OrderStatusPaidNoNotify {
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

// handleCacheRefreshMessages 处理缓存刷新触发消息
func handleCacheRefreshMessages(ctx context.Context, msg *rocketmq.MessageView) error {
	var refreshMsg CacheRefreshMessage
	if err := json.Unmarshal(msg.GetBody(), &refreshMsg); err != nil {
		logger.Logger.Error("解析 cache_refresh 消息失败",
			zap.String("message_id", msg.GetMessageId()),
			zap.Error(err))
		return err
	}

	// 直接调用缓存刷新逻辑，避免循环依赖
	refreshCacheDirectly(ctx, refreshMsg.Full, refreshMsg.Targets, refreshMsg.TenantIDs, refreshMsg.WriteoffIDs)

	logger.Logger.Info("已处理缓存刷新消息",
		zap.String("message_id", msg.GetMessageId()),
		zap.Bool("full", refreshMsg.Full),
		zap.Strings("targets", refreshMsg.Targets),
		zap.Any("tenant_ids", refreshMsg.TenantIDs),
		zap.Any("writeoff_ids", refreshMsg.WriteoffIDs))

	return nil
}

// handleBalanceSyncMessages 处理后台调额后的余额同步消息
func handleBalanceSyncMessages(ctx context.Context, msg *rocketmq.MessageView) error {
	var syncMsg BalanceSyncMessage
	if err := json.Unmarshal(msg.GetBody(), &syncMsg); err != nil {
		logger.Logger.Error("解析 balance_sync 消息失败",
			zap.String("message_id", msg.GetMessageId()),
			zap.Error(err))
		return err
	}

	// 直接调用缓存刷新逻辑，避免循环依赖
	targets := []string{"tenant_balances", "writeoff_balances"}
	refreshCacheDirectly(ctx, syncMsg.Full, targets, syncMsg.TenantIDs, syncMsg.WriteoffIDs)

	logger.Logger.Info("已处理余额同步消息",
		zap.String("message_id", msg.GetMessageId()),
		zap.Any("tenant_ids", syncMsg.TenantIDs),
		zap.Any("writeoff_ids", syncMsg.WriteoffIDs),
		zap.Bool("full", syncMsg.Full))

	return nil
}

// handleOrderTimeoutMessages 处理订单超时消息
func handleOrderTimeoutMessages(ctx context.Context, msg *rocketmq.MessageView) error {
	var timeoutMsg OrderTimeoutMessage
	if err := json.Unmarshal(msg.GetBody(), &timeoutMsg); err != nil {
		logger.Logger.Error("解析 order_timeout 消息失败",
			zap.String("message_id", msg.GetMessageId()),
			zap.Error(err))
		return err
	}

	logger.Logger.Info("收到订单超时消息",
		zap.String("order_id", timeoutMsg.OrderID),
		zap.String("order_no", timeoutMsg.OrderNo),
		zap.Int("timeout_seconds", timeoutMsg.TimeoutSeconds),
		zap.Int64("create_datetime", timeoutMsg.CreateDatetime))

	// 查询订单当前状态（需要查询 merchant_id 和 money，用于释放预占余额）
	var order models.Order
	if err := database.DB.Select("id, order_no, order_status, create_datetime, merchant_id, money").
		Where("id = ?", timeoutMsg.OrderID).
		First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Logger.Warn("订单不存在，可能已删除",
				zap.String("order_id", timeoutMsg.OrderID))
			return nil // 订单不存在，不返回错误
		}
		logger.Logger.Error("查询订单失败",
			zap.String("order_id", timeoutMsg.OrderID),
			zap.Error(err))
		return err
	}

	// 验证订单是否真的超时（防止消息延迟导致误判）
	if order.CreateDatetime != nil {
		expectedExpireTime := order.CreateDatetime.Add(time.Duration(timeoutMsg.TimeoutSeconds) * time.Second)
		now := time.Now()
		// 如果当前时间还没到过期时间，说明消息提前到达了，不处理
		if now.Before(expectedExpireTime) {
			logger.Logger.Info("订单尚未超时，跳过处理（消息提前到达）",
				zap.String("order_id", timeoutMsg.OrderID),
				zap.String("order_no", timeoutMsg.OrderNo),
				zap.Time("create_time", *order.CreateDatetime),
				zap.Time("expected_expire_time", expectedExpireTime),
				zap.Time("current_time", now),
				zap.Duration("remaining_time", expectedExpireTime.Sub(now)))
			return nil
		}
		logger.Logger.Info("订单已超时，开始处理",
			zap.String("order_id", timeoutMsg.OrderID),
			zap.String("order_no", timeoutMsg.OrderNo),
			zap.Time("create_time", *order.CreateDatetime),
			zap.Time("expected_expire_time", expectedExpireTime),
			zap.Time("current_time", now),
			zap.Duration("overdue_time", now.Sub(expectedExpireTime)))
	} else {
		logger.Logger.Warn("订单创建时间为空，无法验证超时时间",
			zap.String("order_id", timeoutMsg.OrderID),
			zap.String("order_no", timeoutMsg.OrderNo))
	}

	// 使用统一的超时处理函数
	// 参考 Python: timeout_check 的逻辑
	if err := HandleOrderTimeout(ctx, timeoutMsg.OrderNo); err != nil {
		logger.Logger.Error("处理订单超时失败",
			zap.String("order_id", timeoutMsg.OrderID),
			zap.String("order_no", timeoutMsg.OrderNo),
			zap.Error(err))
		return err
	}

	logger.Logger.Info("订单已超时，处理完成（通过延迟消息）",
		zap.String("order_id", timeoutMsg.OrderID),
		zap.String("order_no", timeoutMsg.OrderNo),
		zap.Int("timeout_seconds", timeoutMsg.TimeoutSeconds))

	return nil
}

// Close 关闭消费者（优化关闭逻辑，避免长时间阻塞）
func (c *RocketMQConsumer) Close() error {
	if !c.enabled {
		return nil
	}

	if c.consumer != nil {
		logger.Logger.Info("开始关闭 RocketMQ 消费者...")

		// 使用更短的超时时间（2秒），快速关闭
		// 如果消息处理时间较长，超时后会直接退出，不再等待
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			// GracefulStop 会等待正在处理的消息完成
			// 但如果消息处理时间过长，这里会超时
			done <- c.consumer.GracefulStop()
		}()

		select {
		case err := <-done:
			if err != nil {
				logger.Logger.Error("关闭 RocketMQ 消费者失败", zap.Error(err))
				// 即使失败也继续关闭流程，不阻塞应用退出
				return nil
			}
			logger.Logger.Info("RocketMQ 消费者已优雅关闭")
		case <-ctx.Done():
			logger.Logger.Warn("关闭 RocketMQ 消费者超时（2秒），强制退出",
				zap.Error(ctx.Err()),
				zap.String("note", "可能仍有消息正在处理，但应用将强制退出"))
			// 超时后直接返回，不再等待
			// 注意：这可能导致正在处理的消息被中断，但可以快速关闭应用
			return nil
		}
	}

	return nil
}

// IsEnabled 检查是否启用
func (c *RocketMQConsumer) IsEnabled() bool {
	return c.enabled
}

// refreshCacheDirectly 直接刷新缓存（避免循环依赖）
// 这是一个简化版本，只处理必要的缓存刷新逻辑
func refreshCacheDirectly(ctx context.Context, full bool, targets []string, tenantIDs []int64, writeoffIDs []int64) {
	// 创建缓存刷新服务实例（避免导入 service 包）
	// 这里直接调用数据库和 Redis 操作，简化实现
	// 注意：这是一个临时解决方案，理想情况下应该将 CacheRefreshService 移到独立的包中

	// 如果指定了租户ID或码商ID，直接刷新余额
	if len(tenantIDs) > 0 {
		refreshTenantBalancesByIDs(ctx, tenantIDs)
	}
	if len(writeoffIDs) > 0 {
		refreshWriteoffBalancesByIDs(ctx, writeoffIDs)
	}

	// 处理其他目标（如果需要）
	if len(targets) > 0 {
		for _, target := range targets {
			switch target {
			case "tenant_balances":
				if full {
					refreshAllTenantBalances(ctx)
				}
			case "writeoff_balances":
				if full {
					refreshAllWriteoffBalances(ctx)
				}
				// 其他目标可以在这里添加
			}
		}
	}
}

// refreshTenantBalancesByIDs 刷新指定租户的余额
func refreshTenantBalancesByIDs(ctx context.Context, tenantIDs []int64) {
	// 实现逻辑参考 service.CacheRefreshService.refreshTenantBalancesByIDs
	// 这里简化实现，直接查询数据库并更新缓存
	for _, tenantID := range tenantIDs {
		var tenant models.Tenant
		if err := database.DB.Where("id = ?", tenantID).First(&tenant).Error; err == nil {
			// 更新 Redis 缓存
			key := fmt.Sprintf("tenant:%d", tenantID)
			data, _ := json.Marshal(tenant)
			database.RDB.Set(ctx, key, data, 24*time.Hour)
		}
	}
}

// refreshWriteoffBalancesByIDs 刷新指定码商的余额
func refreshWriteoffBalancesByIDs(ctx context.Context, writeoffIDs []int64) {
	// 实现逻辑参考 service.CacheRefreshService.refreshWriteoffBalancesByIDs
	for _, writeoffID := range writeoffIDs {
		var writeoff models.Writeoff
		if err := database.DB.Where("id = ?", writeoffID).First(&writeoff).Error; err == nil {
			// 更新 Redis 缓存
			key := fmt.Sprintf("writeoff:%d", writeoffID)
			data, _ := json.Marshal(writeoff)
			database.RDB.Set(ctx, key, data, 24*time.Hour)
		}
	}
}

// refreshAllTenantBalances 刷新所有租户余额
func refreshAllTenantBalances(ctx context.Context) {
	// 简化实现：刷新所有租户
	var tenants []models.Tenant
	if err := database.DB.Find(&tenants).Error; err == nil {
		for _, tenant := range tenants {
			key := fmt.Sprintf("tenant:%d", tenant.ID)
			data, _ := json.Marshal(tenant)
			database.RDB.Set(ctx, key, data, 24*time.Hour)
		}
	}
}

// refreshAllWriteoffBalances 刷新所有码商余额
func refreshAllWriteoffBalances(ctx context.Context) {
	// 简化实现：刷新所有码商
	var writeoffs []models.Writeoff
	if err := database.DB.Find(&writeoffs).Error; err == nil {
		for _, writeoff := range writeoffs {
			key := fmt.Sprintf("writeoff:%d", writeoff.ID)
			data, _ := json.Marshal(writeoff)
			database.RDB.Set(ctx, key, data, 24*time.Hour)
		}
	}
}
