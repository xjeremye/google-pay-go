package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/example/payment-core/internal/db"
	"github.com/example/payment-core/internal/plugin"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	kafka "github.com/segmentio/kafka-go"
)

// Service 封装订单相关的业务逻辑
// 使用 sqlc 生成的类型安全代码进行数据库操作
type Service struct {
	db        *sqlx.DB
	queries   *db.Queries
	redis     *redis.Client
	kafkaW    *kafka.Writer
	helper    *Helper
	pluginMgr *plugin.Manager
}

// NewService 创建订单服务
func NewService(dbConn *sqlx.DB, rdb *redis.Client, kw *kafka.Writer) *Service {
	s := &Service{
		db:        dbConn,
		queries:   db.New(dbConn),
		redis:     rdb,
		kafkaW:    kw,
		pluginMgr: plugin.NewManager(rdb),
	}
	s.helper = NewHelper(s)
	return s
}

// QueryStatus 查询订单状态
func (s *Service) QueryStatus(ctx context.Context, orderID string) (string, error) {
	// 使用 sqlc 生成的类型安全代码
	status, err := s.queries.GetOrderStatus(ctx, orderID)
	if err != nil {
		return "", fmt.Errorf("get order status failed: %w", err)
	}
	return status, nil
}

// UpdateStatus 根据渠道回调更新订单状态
func (s *Service) UpdateStatus(ctx context.Context, orderID, status, channelTradeNo string, payload map[string]interface{}) error {
	payloadBytes, _ := json.Marshal(payload)

	// 使用 sqlc 生成的类型安全代码
	var channelTradeNoNull sql.NullString
	if channelTradeNo != "" {
		channelTradeNoNull = sql.NullString{String: channelTradeNo, Valid: true}
	}

	err := s.queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		Status:         status,
		ChannelTradeNo: channelTradeNoNull,
		Payload:        payloadBytes,
		UpdatedAt:      time.Now().UTC(),
		OrderID:        orderID,
	})
	if err != nil {
		return fmt.Errorf("update order status failed: %w", err)
	}
	return nil
}

// OrderProcessingError 订单处理错误
type OrderProcessingError struct {
	Code    int
	Message string
	Data    interface{}
}

func (e *OrderProcessingError) Error() string {
	return e.Message
}

// RawCreateOrderRequest 原始创建订单请求参数
type RawCreateOrderRequest struct {
	OutOrderNo  string                 // mchOrderNo
	MerchantID  int                    // mchId
	ChannelID   int                    // channelId
	Money       int                    // amount
	NotifyURL   string                 // notifyUrl
	Extra       string                 // extra
	JumpURL     string                 // jumpUrl
	Compatible  int                    // compatible
	Test        bool                   // test
	Sign        string                 // sign
	RawSignData map[string]interface{} // 原始签名数据
}

// OrderCreateCtx 订单创建上下文
type OrderCreateCtx struct {
	OutOrderNo  string
	NotifyURL   string
	Money       int
	JumpURL     string
	NotifyMoney int
	Extra       string
	Compatible  int
	Test        bool

	// 以下字段在检查过程中填充
	MerchantID     int
	TenantID       int
	ChannelID      int
	PluginID       int
	PluginType     string
	PluginUpstream int
	DomainID       int
	DomainURL      string
	Domain         map[string]interface{} // Domain 对象
	Channel        map[string]interface{} // Channel 对象
	Plugin         map[string]interface{} // Plugin 对象
	PayType        map[string]interface{} // PayType 对象
	Responder      interface{}            // PluginResponder 对象
	ProductID      *int
	WriteoffID     *int
	CookieID       *int
	Tax            int
	SignRaw        string
	Sign           string
	SignKey        string

	// 订单对象
	Order    interface{} // Order 对象
	OrderNo  string
	Detail   interface{} // OrderDetail 对象
	DetailID int
}

// 实现 plugin.OrderContext 接口
func (o *OrderCreateCtx) GetOutOrderNo() string              { return o.OutOrderNo }
func (o *OrderCreateCtx) GetNotifyURL() string               { return o.NotifyURL }
func (o *OrderCreateCtx) GetMoney() int                      { return o.Money }
func (o *OrderCreateCtx) GetJumpURL() string                 { return o.JumpURL }
func (o *OrderCreateCtx) GetNotifyMoney() int                { return o.NotifyMoney }
func (o *OrderCreateCtx) GetExtra() string                   { return o.Extra }
func (o *OrderCreateCtx) GetCompatible() int                 { return o.Compatible }
func (o *OrderCreateCtx) GetTest() bool                      { return o.Test }
func (o *OrderCreateCtx) GetMerchantID() int                 { return o.MerchantID }
func (o *OrderCreateCtx) GetTenantID() int                   { return o.TenantID }
func (o *OrderCreateCtx) GetChannelID() int                  { return o.ChannelID }
func (o *OrderCreateCtx) GetPluginID() int                   { return o.PluginID }
func (o *OrderCreateCtx) GetPluginType() string              { return o.PluginType }
func (o *OrderCreateCtx) GetPluginUpstream() int             { return o.PluginUpstream }
func (o *OrderCreateCtx) GetDomainID() int                   { return o.DomainID }
func (o *OrderCreateCtx) GetDomainURL() string               { return o.DomainURL }
func (o *OrderCreateCtx) GetDomain() map[string]interface{}  { return o.Domain }
func (o *OrderCreateCtx) GetChannel() map[string]interface{} { return o.Channel }
func (o *OrderCreateCtx) GetPlugin() map[string]interface{}  { return o.Plugin }
func (o *OrderCreateCtx) GetPayType() map[string]interface{} { return o.PayType }
func (o *OrderCreateCtx) GetOrderNo() string                 { return o.OrderNo }
func (o *OrderCreateCtx) SetOrderNo(no string)               { o.OrderNo = no }
func (o *OrderCreateCtx) SetDomain(d map[string]interface{}) { o.Domain = d }
func (o *OrderCreateCtx) SetDomainURL(url string)            { o.DomainURL = url }
func (o *OrderCreateCtx) SetDomainID(id int)                 { o.DomainID = id }

// RawCreateOrderResponse 原始创建订单响应
type RawCreateOrderResponse struct {
	// Compatible == 1 时的响应格式
	TradeNo string `json:"trade_no,omitempty"`
	PayURL  string `json:"payurl,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Code    int    `json:"code,omitempty"`

	// Compatible == 0 时的响应格式
	MchOrderNo string `json:"mchOrderNo,omitempty"`
	PayOrderID string `json:"payOrderId,omitempty"`
	PayURL2    string `json:"payUrl,omitempty"`
	Sign       string `json:"sign,omitempty"`
}

// RawCreateOrder 原始创建订单核心逻辑
func (s *Service) RawCreateOrder(ctx context.Context, req RawCreateOrderRequest) (*RawCreateOrderResponse, *OrderProcessingError) {
	// 检查金额
	if req.Money <= 0 {
		return nil, &OrderProcessingError{
			Code:    0,
			Message: "金额必须大于0",
		}
	}

	// 创建上下文
	orderCtx := &OrderCreateCtx{
		OutOrderNo:  req.OutOrderNo,
		NotifyURL:   req.NotifyURL,
		Money:       req.Money,
		JumpURL:     req.JumpURL,
		NotifyMoney: req.Money,
		Extra:       req.Extra,
		Compatible:  req.Compatible,
		Test:        req.Test,
	}

	firstTime := time.Now().UnixMilli()

	// 执行一系列检查
	if err := s.helper.CheckMerchant(ctx, orderCtx, req.MerchantID); err != nil {
		return nil, err
	}
	if err := s.helper.CheckTenant(ctx, orderCtx); err != nil {
		return nil, err
	}
	if err := s.helper.CheckSign(ctx, orderCtx, req.RawSignData); err != nil {
		return nil, err
	}
	if err := s.helper.CheckOutOrderNo(ctx, orderCtx); err != nil {
		return nil, err
	}
	if err := s.helper.CheckChannel(ctx, orderCtx, req.ChannelID); err != nil {
		return nil, err
	}
	if err := s.helper.CheckPlugin(ctx, orderCtx); err != nil {
		return nil, err
	}
	if err := s.helper.CheckDomain(ctx, orderCtx); err != nil {
		return nil, err
	}
	if err := s.helper.CheckPluginResponder(ctx, orderCtx); err != nil {
		return nil, err
	}

	secondTime := time.Now().UnixMilli()
	if secondTime-firstTime > 1000 {
		// TODO: 记录日志
		fmt.Printf("%s拉单预操作耗时:%d ms\n", orderCtx.OutOrderNo, secondTime-firstTime)
	}

	// 处理 extra 相关逻辑
	// 从插件获取这些属性（使用类型断言访问能力接口）
	responderCouldExtra := false
	responderAutoExtra := false
	responderExtraNeedProduct := false
	responderExtraNeedCookie := false

	if orderCtx.Responder != nil {
		if pluginInstance, ok := orderCtx.Responder.(plugin.Plugin); ok {
			if caps, ok := pluginInstance.(plugin.PluginCapabilities); ok {
				responderCouldExtra = caps.CanHandleExtra()
				responderAutoExtra = caps.AutoExtra()
				responderExtraNeedProduct = caps.ExtraNeedProduct()
				responderExtraNeedCookie = caps.ExtraNeedCookie()
			}
		}
	}

	// TODO: 从 cache 获取 auto_extra 配置
	autoExtra := false

	if (orderCtx.Extra != "" && responderCouldExtra) || responderAutoExtra || autoExtra {
		// 预检查 extra
		flag, msg := s.helper.PreCheckExtra(ctx, orderCtx)
		if !flag {
			return nil, &OrderProcessingError{
				Code:    7319,
				Message: msg,
			}
		}

		// 如果需要产品
		if responderExtraNeedProduct {
			writeoffIDs := s.helper.GetWriteoffIDs(ctx, orderCtx.TenantID, orderCtx.Money, orderCtx.ChannelID)
			if len(writeoffIDs) == 0 {
				// TODO: 记录日志
				return nil, &OrderProcessingError{
					Code:    7318,
					Message: "无库存",
				}
			}

			productID, writeoffID, _ := s.helper.GetExtraWriteoffProduct(ctx, orderCtx, writeoffIDs)
			if productID == nil || writeoffID == nil {
				// TODO: 记录日志
				return nil, &OrderProcessingError{
					Code:    7318,
					Message: "无库存",
				}
			}
			orderCtx.ProductID = productID
			orderCtx.WriteoffID = writeoffID
		}

		thirdTime1 := time.Now().UnixMilli()
		if thirdTime1-secondTime > 1000 {
			// TODO: 记录日志
			fmt.Printf("%s拉单检测货物耗时:%d ms,目前总耗时:%d ms\n",
				orderCtx.OutOrderNo, thirdTime1-secondTime, thirdTime1-firstTime)
		}

		// 检测小号
		if responderExtraNeedCookie {
			cookieID := s.helper.GetExtraTenantCookie(ctx, orderCtx.PluginID, orderCtx.TenantID)
			if cookieID == nil {
				// TODO: 记录日志
				return nil, &OrderProcessingError{
					Code:    7318,
					Message: "无库存",
				}
			}
			orderCtx.CookieID = cookieID
		}

		thirdTime2 := time.Now().UnixMilli()
		if thirdTime2-secondTime > 1000 {
			// TODO: 记录日志
			fmt.Printf("%s拉单检测ck耗时:%d ms,目前总耗时:%d ms\n",
				orderCtx.OutOrderNo, thirdTime2-thirdTime1, thirdTime2-firstTime)
		}
	} else {
		if orderCtx.Test {
			// TODO: 从 kwargs 获取测试数据
		} else {
			// TODO: 等待产品
			s.helper.WaitProduct(ctx, orderCtx)
		}
	}

	// 获取插件超时时间
	outTimeSeconds := s.helper.GetPluginOutTime(ctx, orderCtx.PluginID)

	if !responderCouldExtra {
		orderCtx.Extra = ""
	}

	// 尝试创建订单和订单详情
	if err := s.helper.TryCreateOrder(ctx, orderCtx); err != nil {
		return nil, err
	}
	if err := s.helper.TryCreateOrderDetail(ctx, orderCtx); err != nil {
		return nil, err
	}

	// 检查余额
	if !s.helper.CheckTenantBalance(ctx, orderCtx) {
		// TODO: 删除订单和详情
		return nil, &OrderProcessingError{
			Code:    7315,
			Message: "余额不足",
		}
	}

	// 检查码商余额
	if orderCtx.WriteoffID != nil {
		if !s.helper.TakeUpWriteoffTax(ctx, *orderCtx.WriteoffID, orderCtx.Money) {
			// TODO: 删除订单和详情
			return nil, &OrderProcessingError{
				Code:    7315,
				Message: "码商余额不足",
			}
		}
	}

	// 更新或创建订单日志
	s.helper.UpdateOrCreateOrderLog(ctx, orderCtx.OutOrderNo, orderCtx.SignRaw, orderCtx.Sign)

	thirdTime := time.Now().UnixMilli()

	// 构建响应数据
	var resData RawCreateOrderResponse
	if orderCtx.Compatible == 1 {
		resData.TradeNo = orderCtx.OrderNo
	} else {
		resData.MchOrderNo = orderCtx.OutOrderNo
		resData.PayOrderID = orderCtx.OrderNo
	}

	// 处理 extra 支付 URL
	// 如果需要extra支付URL，可以通过CreateOrder方法获取
	responderExtraPayURL := false
	if (orderCtx.Extra != "" && responderCouldExtra) || autoExtra || responderAutoExtra {
		extraData := s.helper.GetExtraPayURL(ctx, orderCtx)
		if extraData == nil {
			s.helper.FailOrder(ctx, orderCtx.OrderNo)
			return nil, &OrderProcessingError{
				Code:    7320,
				Message: "创建失败",
				Data:    resData,
			}
		}

		// 从 extraData 中提取 cookie_id, product_id, writeoff_id
		if cookieID, ok := extraData["cookie_id"].(int); ok {
			orderCtx.CookieID = &cookieID
		}
		if productID, ok := extraData["product_id"].(int); ok {
			orderCtx.ProductID = &productID
		}
		if writeoffID, ok := extraData["writeoff_id"].(int); ok {
			orderCtx.WriteoffID = &writeoffID
		}

		// 更新响应数据
		// TODO: 将 extraData 中的其他字段更新到 resData

		// 添加定时查询任务
		s.helper.AddPluginQueryOrderJob(ctx, orderCtx.OrderNo, 5*time.Second)
	}

	// 生成支付 URL - 使用插件统一接口
	if !responderExtraPayURL && !autoExtra {
		// 获取插件实例
		pluginInstance, err := s.pluginMgr.GetPluginByCtx(ctx, orderCtx)
		if err != nil {
			return nil, &OrderProcessingError{
				Code:    7318,
				Message: "插件不可用",
			}
		}

		// 构建标准请求
		createReq := &plugin.CreateOrderRequest{
			OutOrderNo:     orderCtx.OutOrderNo,
			OrderNo:        orderCtx.OrderNo,
			Money:          orderCtx.Money,
			NotifyURL:      orderCtx.NotifyURL,
			JumpURL:        orderCtx.JumpURL,
			Extra:          orderCtx.Extra,
			MerchantID:     orderCtx.MerchantID,
			TenantID:       orderCtx.TenantID,
			ChannelID:      orderCtx.ChannelID,
			PluginID:       orderCtx.PluginID,
			PluginType:     orderCtx.PluginType,
			PluginUpstream: orderCtx.PluginUpstream,
			DomainID:       orderCtx.DomainID,
			DomainURL:      orderCtx.DomainURL,
			Domain:         orderCtx.Domain,
			Channel:        orderCtx.Channel,
			Plugin:         orderCtx.Plugin,
			PayType:        orderCtx.PayType,
			Compatible:     orderCtx.Compatible,
			Test:           orderCtx.Test,
		}

		// 调用插件创建订单（插件内部处理域名选择、配置获取等）
		createResp, err := pluginInstance.CreateOrder(ctx, createReq)
		if err != nil {
			return nil, &OrderProcessingError{
				Code:    7320,
				Message: fmt.Sprintf("创建订单失败: %v", err),
			}
		}

		// 检查响应
		if !createResp.IsSuccess() {
			return nil, &OrderProcessingError{
				Code:    createResp.ErrorCode,
				Message: createResp.ErrorMessage,
			}
		}

		// 设置支付URL
		resData.PayURL2 = createResp.PayURL

		// 设置缓存
		h := orderCtx.OrderNo
		s.redis.Set(ctx, h, orderCtx.OrderNo, time.Duration(outTimeSeconds+10)*time.Second)
		s.redis.Set(ctx, fmt.Sprintf("%s-h", orderCtx.OrderNo), h, time.Duration(outTimeSeconds+10)*time.Second)
	}

	forthTime := time.Now().UnixMilli()
	// TODO: 记录日志
	fmt.Printf("%s拉单0耗时:%d ms, 目前总耗时:%d ms\n",
		orderCtx.OutOrderNo, forthTime-thirdTime, forthTime-firstTime)

	// 设置超时任务
	s.helper.AddTimeoutCheckJob(ctx, orderCtx.OrderNo, orderCtx.TenantID, orderCtx.Tax, outTimeSeconds)

	// 设置提交通知任务
	ignoreNoURL := false // TODO: 从 cache 获取
	if !ignoreNoURL {
		s.helper.AddNotifyOrderSubmitJob(ctx, orderCtx, time.Microsecond*500)
	}

	// 设置缓存
	outTime := time.Now().Add(time.Duration(outTimeSeconds) * time.Second)
	s.redis.Set(ctx, fmt.Sprintf("%s-pay-time", orderCtx.OrderNo),
		outTime.Format("2006-01-02 15:04:05"), time.Duration(outTimeSeconds)*time.Second)
	s.redis.Set(ctx, fmt.Sprintf("%s-timeout", orderCtx.OrderNo),
		outTime.Format("2006-01-02 15:04:05"), time.Duration(outTimeSeconds)*time.Second)

	createArgs := map[string]interface{}{
		"order_no":        orderCtx.OrderNo,
		"out_order_no":    orderCtx.OutOrderNo,
		"plugin_id":       orderCtx.PluginID,
		"tax":             orderCtx.Tax,
		"plugin_type":     orderCtx.PluginType,
		"money":           orderCtx.Money,
		"domain_id":       orderCtx.DomainID,
		"notify_money":    orderCtx.NotifyMoney,
		"order_id":        0, // TODO: 从 orderCtx.Order 获取
		"product_id":      orderCtx.ProductID,
		"cookie_id":       orderCtx.CookieID,
		"channel_id":      orderCtx.ChannelID,
		"merchant_id":     orderCtx.MerchantID,
		"writeoff_id":     orderCtx.WriteoffID,
		"tenant_id":       orderCtx.TenantID,
		"create_datetime": time.Now().Format("2006-01-02 15:04:05"),
		"notify_url":      orderCtx.NotifyURL,
		"plugin_upstream": orderCtx.PluginUpstream,
	}
	createArgsBytes, _ := json.Marshal(createArgs)
	s.redis.Set(ctx, fmt.Sprintf("%s-create", orderCtx.OrderNo), string(createArgsBytes),
		time.Duration(outTimeSeconds+5)*time.Second)
	s.redis.Set(ctx, fmt.Sprintf("%s-create-time", orderCtx.OrderNo),
		time.Now().Format("2006-01-02 15:04:05"), time.Duration(outTimeSeconds)*time.Second)

	fifthTime := time.Now().UnixMilli()
	if fifthTime-forthTime > 500 {
		// TODO: 记录错误日志
		fmt.Printf("%s拉单最后阶段耗时:%d ms, 目前总耗时:%d ms\n",
			orderCtx.OutOrderNo, fifthTime-forthTime, fifthTime-firstTime)
	} else {
		// TODO: 记录信息日志
		fmt.Printf("%s拉单最后阶段耗时:%d ms, 目前总耗时:%d ms\n",
			orderCtx.OutOrderNo, fifthTime-forthTime, fifthTime-firstTime)
	}

	// 处理兼容模式响应
	if orderCtx.Compatible == 1 {
		resData.PayURL = resData.PayURL2
		resData.PayURL2 = ""
		resData.Msg = "订单创建成功"
		resData.Code = 1
	} else {
		// TODO: 生成签名
		// resData.Sign = s.helper.ToSign(resData, orderCtx)
	}

	return &resData, nil
}
