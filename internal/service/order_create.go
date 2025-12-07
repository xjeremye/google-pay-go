package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-pay-core/config"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/plugin"
	"github.com/golang-pay-core/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// isValidJSON 检查字符串是否为有效的 JSON
func isValidJSON(s string) bool {
	if s == "" {
		return false
	}
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	OutOrderNo    string                 `json:"mchOrderNo" binding:"required"`   // 商户订单号
	MerchantID    int                    `json:"mchId" binding:"required"`        // 商户ID
	ChannelID     int                    `json:"channelId" binding:"required"`    // 渠道ID
	Money         int                    `json:"amount" binding:"required,min=1"` // 金额（分）
	NotifyURL     string                 `json:"notifyUrl" binding:"required"`    // 通知地址
	JumpURL       string                 `json:"jumpUrl"`                         // 跳转地址
	Extra         string                 `json:"extra"`                           // 额外参数
	Compatible    int                    `json:"compatible"`                      // 兼容模式 0/1
	Test          bool                   `json:"test"`                            // 测试模式
	Sign          string                 `json:"sign" binding:"required"`         // 签名
	RawSignData   map[string]interface{} `json:"-"`                               // 原始签名数据（内部使用）
	SignRaw       string                 `json:"-"`                               // 签名原始数据（JSON字符串，用于日志）
	RequestMethod string                 `json:"-"`                               // 请求方法（GET/POST）
	RequestBody   string                 `json:"-"`                               // 请求体（JSON字符串，用于日志）
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
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

// OrderCreateContext 订单创建上下文
type OrderCreateContext struct {
	// 请求参数
	OutOrderNo  string
	NotifyURL   string
	Money       int
	JumpURL     string
	NotifyMoney int
	Extra       string
	Compatible  int
	Test        bool

	// 验证后填充的字段
	MerchantID     int64
	TenantID       int64
	ChannelID      int64
	PluginID       int64
	PluginType     string
	PluginUpstream int
	DomainID       *int64
	DomainURL      string
	Tax            int
	SignKey        string
	WriteoffID     *int64 // 核销ID（可能从插件获取）
	ProductID      string // 产品ID（从插件获取）
	CookieID       string // Cookie ID（从插件获取）
	SignRaw        string // 签名原始数据
	Sign           string // 签名数据

	// 关联对象
	Merchant   *models.Merchant
	Tenant     *models.Tenant
	Writeoff   *models.Writeoff // 码商信息
	Channel    *models.PayChannel
	Plugin     *models.PayPlugin
	PayType    *models.PayType
	Domain     *models.PayDomain
	User       *SystemUser
	TenantUser *SystemUser

	// 订单信息
	OrderNo       string
	OrderID       string
	OrderDetailID int64 // 订单详情ID（创建后保存，避免重复查询）

	// 请求信息（用于日志记录）
	RequestMethod string // 请求方法（GET/POST）
	RequestBody   string // 请求体（JSON字符串）
}

// 实现 plugin.OrderContext 接口
func (o *OrderCreateContext) GetOutOrderNo() string   { return o.OutOrderNo }
func (o *OrderCreateContext) GetNotifyURL() string    { return o.NotifyURL }
func (o *OrderCreateContext) GetMoney() int           { return o.Money }
func (o *OrderCreateContext) GetJumpURL() string      { return o.JumpURL }
func (o *OrderCreateContext) GetNotifyMoney() int     { return o.NotifyMoney }
func (o *OrderCreateContext) GetExtra() string        { return o.Extra }
func (o *OrderCreateContext) GetCompatible() int      { return o.Compatible }
func (o *OrderCreateContext) GetTest() bool           { return o.Test }
func (o *OrderCreateContext) GetMerchantID() int64    { return o.MerchantID }
func (o *OrderCreateContext) GetTenantID() int64      { return o.TenantID }
func (o *OrderCreateContext) GetChannelID() int64     { return o.ChannelID }
func (o *OrderCreateContext) GetPluginID() int64      { return o.PluginID }
func (o *OrderCreateContext) GetPluginType() string   { return o.PluginType }
func (o *OrderCreateContext) GetPluginUpstream() int  { return o.PluginUpstream }
func (o *OrderCreateContext) GetDomainID() *int64     { return o.DomainID }
func (o *OrderCreateContext) GetDomainURL() string    { return o.DomainURL }
func (o *OrderCreateContext) GetOrderNo() string      { return o.OrderNo }
func (o *OrderCreateContext) SetOrderNo(no string)    { o.OrderNo = no }
func (o *OrderCreateContext) SetDomainID(id int64)    { o.DomainID = &id }
func (o *OrderCreateContext) SetDomainURL(url string) { o.DomainURL = url }

// OrderService 订单服务（重构版）
type OrderService struct {
	cacheService   *CacheService
	pluginService  *PluginService
	pluginManager  *plugin.Manager
	balanceService *BalanceService
	redis          *redis.Client
}

// NewOrderService 创建订单服务
func NewOrderService() *OrderService {
	pluginMgr := plugin.NewManager(database.RDB)
	pluginSvc := NewPluginService()

	// 设置插件信息提供者（实现 PluginInfoProvider 接口）
	pluginMgr.SetInfoProvider(&pluginInfoProviderAdapter{service: pluginSvc})

	return &OrderService{
		cacheService:   NewCacheService(),
		pluginService:  pluginSvc,
		pluginManager:  pluginMgr,
		balanceService: NewBalanceService(),
		redis:          database.RDB,
	}
}

// pluginInfoProviderAdapter 适配器，将 PluginService 适配为 PluginInfoProvider
type pluginInfoProviderAdapter struct {
	service *PluginService
}

func (a *pluginInfoProviderAdapter) GetPlugin(ctx context.Context, pluginID int64) (interface{}, error) {
	return a.service.GetPlugin(ctx, pluginID)
}

func (a *pluginInfoProviderAdapter) GetPluginUpstream(ctx context.Context, pluginID int64) (int, error) {
	return a.service.GetPluginUpstream(ctx, pluginID)
}

func (a *pluginInfoProviderAdapter) GetPluginPayTypes(ctx context.Context, pluginID int64) ([]interface{}, error) {
	payTypes, err := a.service.GetPluginPayTypes(ctx, pluginID)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(payTypes))
	for i, pt := range payTypes {
		result[i] = map[string]interface{}{
			"id":     pt.ID,
			"name":   pt.Name,
			"key":    pt.Key,
			"status": pt.Status,
		}
	}
	return result, nil
}

// CreateOrder 创建订单（主入口）
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, *OrderError) {
	startTime := time.Now()
	firstTime := startTime.UnixMilli()

	// 1. 基础验证
	if req.Money <= 0 {
		return nil, ErrAmountInvalid
	}

	// 2. 创建上下文
	orderCtx := &OrderCreateContext{
		OutOrderNo:    req.OutOrderNo,
		NotifyURL:     req.NotifyURL,
		Money:         req.Money,
		JumpURL:       req.JumpURL,
		NotifyMoney:   req.Money,
		Extra:         req.Extra,
		Compatible:    req.Compatible,
		Test:          req.Test,
		SignRaw:       req.SignRaw,
		RequestMethod: req.RequestMethod,
		RequestBody:   req.RequestBody,
	}

	// 3. 执行验证链
	if err := s.validateMerchant(ctx, orderCtx, int64(req.MerchantID)); err != nil {
		return nil, err
	}
	if err := s.validateTenant(ctx, orderCtx); err != nil {
		return nil, err
	}
	if err := s.validateSign(ctx, orderCtx, req.RawSignData); err != nil {
		return nil, err
	}
	if err := s.validateOutOrderNo(ctx, orderCtx); err != nil {
		return nil, err
	}
	if err := s.validateChannel(ctx, orderCtx, int64(req.ChannelID)); err != nil {
		return nil, err
	}
	if err := s.validatePlugin(ctx, orderCtx); err != nil {
		return nil, err
	}

	// 4. 验证域名（收银台）
	// 参考 Python: order_check_domain(ctx)
	if err := s.validateDomain(ctx, orderCtx); err != nil {
		return nil, err
	}

	// 记录预操作耗时
	secondTime := time.Now().UnixMilli()
	preOpElapsed := secondTime - firstTime
	if preOpElapsed > 1000 {
		logger.Logger.Error("拉单预操作耗时过长",
			zap.String("out_order_no", orderCtx.OutOrderNo),
			zap.Int64("elapsed_ms", preOpElapsed))
	}

	// 5. 等待产品（通过 product selector 选择产品）
	// 参考 Python: ctx.responder.wait_product(ctx)
	// 这一步必须在创建订单之前，因为需要先获取产品ID、核销ID等信息
	waitProductStartTime := time.Now()
	if err := s.waitProduct(ctx, orderCtx); err != nil {
		return nil, err
	}
	waitProductElapsed := time.Since(waitProductStartTime).Milliseconds()
	if waitProductElapsed > 1000 {
		logger.Logger.Error("拉单检测货物耗时过长",
			zap.String("out_order_no", orderCtx.OutOrderNo),
			zap.Int64("elapsed_ms", waitProductElapsed),
			zap.Int64("total_elapsed_ms", time.Since(startTime).Milliseconds()))
	}

	// 获取码商信息（如果存在码商ID）
	if orderCtx.WriteoffID != nil {
		writeoff, _, err := s.cacheService.GetWriteoffWithUser(ctx, *orderCtx.WriteoffID)
		if err != nil {
			// 如果获取失败，记录日志但不阻止订单创建（容错处理）
			logger.Logger.Warn("获取码商信息失败",
				zap.Int64("writeoff_id", *orderCtx.WriteoffID),
				zap.Error(err))
		} else {
			orderCtx.Writeoff = writeoff
		}
	}

	// 5. 预检查余额（使用缓存，快速检查）
	// 注意：这只是预检查，最终检查在创建订单的事务中进行
	if err := s.validateBalance(ctx, orderCtx); err != nil {
		return nil, err
	}

	// 6. 创建订单和详情（此时已经有产品ID、核销ID等信息）
	// 在事务中会再次检查余额，确保一致性
	orderDetailID, err := s.createOrderAndDetail(ctx, orderCtx)
	if err != nil {
		return nil, err
	}
	// 保存订单详情ID到上下文，避免后续重复查询
	orderCtx.OrderDetailID = orderDetailID

	// 9. 生成支付URL（使用插件系统，此时产品已选择）
	thirdTime := time.Now()
	payURL, err := s.generatePayURL(ctx, orderCtx)
	if err != nil {
		return nil, err
	}

	// 10. 生成鉴权链接并返回收银台地址（如果需要）
	// 参考 Python: get_auth_url 方法
	finalURL, err := s.getAuthURL(ctx, orderCtx, payURL)
	if err != nil {
		return nil, err
	}

	// 记录生成支付URL耗时
	forthTime := time.Now()
	payURLElapsed := forthTime.Sub(thirdTime).Milliseconds()
	logger.Logger.Info("拉单生成支付URL耗时",
		zap.String("out_order_no", orderCtx.OutOrderNo),
		zap.Int64("elapsed_ms", payURLElapsed),
		zap.Int64("total_elapsed_ms", forthTime.Sub(startTime).Milliseconds()))

	// 7. 设置缓存
	s.setCache(ctx, orderCtx)

	// 8. 构建响应
	response := s.buildResponse(orderCtx, finalURL)

	// 记录最后阶段耗时
	fifthTime := time.Now()
	lastStageElapsed := fifthTime.Sub(forthTime).Milliseconds()
	totalElapsed := fifthTime.Sub(startTime).Milliseconds()
	if lastStageElapsed > 500 {
		logger.Logger.Error("拉单最后阶段耗时过长",
			zap.String("out_order_no", orderCtx.OutOrderNo),
			zap.Int64("elapsed_ms", lastStageElapsed),
			zap.Int64("total_elapsed_ms", totalElapsed))
	} else {
		logger.Logger.Info("拉单最后阶段耗时",
			zap.String("out_order_no", orderCtx.OutOrderNo),
			zap.Int64("elapsed_ms", lastStageElapsed),
			zap.Int64("total_elapsed_ms", totalElapsed))
	}

	// 记录订单创建成功日志
	// 参考 Python: logger.info(f"订单创建成功, 订单号:{ctx.order.order_no}({ctx.out_order_no}), ...")
	logger.Logger.Info("订单创建成功",
		zap.String("order_no", orderCtx.OrderNo),
		zap.String("out_order_no", orderCtx.OutOrderNo),
		zap.Int("money", orderCtx.Money),
		zap.Int("tax", orderCtx.Tax),
		zap.Int64("merchant_id", orderCtx.MerchantID),
		zap.Int64("tenant_id", orderCtx.TenantID),
		zap.Int64("channel_id", orderCtx.ChannelID),
		zap.String("domain", func() string {
			if orderCtx.Domain != nil {
				return orderCtx.DomainURL
			}
			return ""
		}()),
		zap.String("product_id", orderCtx.ProductID),
		zap.String("plugin_type", orderCtx.PluginType),
		zap.Int("plugin_upstream", orderCtx.PluginUpstream),
		zap.Int64("total_elapsed_ms", totalElapsed))

	// 注意：响应信息的记录应该在 Controller 层完成，因为响应格式可能包含额外的包装
	// 这里保留作为备用，但主要逻辑应该在 Controller 层
	// s.updateOrderLogResponse(ctx, orderCtx, response)

	return response, nil
}

// validateMerchant 验证商户
func (s *OrderService) validateMerchant(ctx context.Context, orderCtx *OrderCreateContext, merchantID int64) *OrderError {
	merchant, user, err := s.cacheService.GetMerchantWithUser(ctx, merchantID)
	if err != nil {
		return ErrMerchantNotFound
	}

	if user == nil {
		return ErrMerchantDisabled
	}

	if !user.Status {
		return ErrMerchantDisabled
	}

	orderCtx.Merchant = merchant
	orderCtx.MerchantID = merchantID
	orderCtx.User = user
	orderCtx.SignKey = user.Key

	return nil
}

// validateTenant 验证租户
func (s *OrderService) validateTenant(ctx context.Context, orderCtx *OrderCreateContext) *OrderError {
	if orderCtx.Merchant == nil {
		return ErrMerchantNotFound
	}

	tenant, tenantUser, err := s.cacheService.GetTenantWithUser(ctx, orderCtx.Merchant.ParentID)
	if err != nil {
		return NewOrderError(ErrCodeMerchantDisabled, "商户上级已被禁用,请联系管理员")
	}

	if tenantUser == nil || !tenantUser.Status {
		return NewOrderError(ErrCodeMerchantDisabled, "商户上级已被禁用,请联系管理员")
	}

	orderCtx.Tenant = tenant
	orderCtx.TenantID = tenant.ID
	orderCtx.TenantUser = tenantUser

	return nil
}

// validateSign 验证签名
func (s *OrderService) validateSign(ctx context.Context, orderCtx *OrderCreateContext, rawSignData map[string]interface{}) *OrderError {
	if orderCtx.SignKey == "" {
		return ErrSignInvalid
	}

	sign, ok := rawSignData["sign"].(string)
	if !ok || sign == "" {
		return ErrSignInvalid
	}

	signRaw, actualSign := utils.GetSign(rawSignData, orderCtx.SignKey, nil, nil, orderCtx.Compatible)
	if sign != actualSign {
		// 如果是开发、测试模式则输出正确签名
		if config.Cfg.App.Mode == "debug" || config.Cfg.App.Mode == "test" {
			fmt.Println("sign", sign)
			fmt.Println("actualSign", actualSign)
		}
		return ErrSignInvalid
	}

	// 保存签名原始数据和签名
	// 注意：signRaw 是 GetSign 返回的用于签名的原始字符串（如 "key=value&key=value&key=xxx"）
	// 但根据 Python 代码和数据库表结构，sign_raw 字段应该存储的是原始请求数据的 JSON 字符串
	// 这里保存 signRaw（用于签名的原始字符串）到 orderCtx.SignRaw
	// 如果 Controller 层已经设置了 SignRaw（JSON 格式），这里会覆盖它
	// 为了保持与 Python 代码一致，这里使用 signRaw（用于签名的原始字符串）
	orderCtx.SignRaw = signRaw
	orderCtx.Sign = actualSign

	return nil
}

// validateOutOrderNo 验证商户订单号（使用 Redis SetNX 做幂等控制）
func (s *OrderService) validateOutOrderNo(ctx context.Context, orderCtx *OrderCreateContext) *OrderError {
	if orderCtx.OutOrderNo == "" {
		return ErrOutOrderNoRequired
	}

	key := fmt.Sprintf("out_order_no:%s", orderCtx.OutOrderNo)
	ok, err := s.redis.SetNX(ctx, key, "1", 24*time.Hour).Result()
	if err != nil {
		return ErrSystemBusy
	}

	if !ok {
		return ErrOutOrderNoExists
	}

	return nil
}

// validateChannel 验证渠道
// 注意：渠道信息从缓存获取，缓存刷新服务每秒更新一次，确保限额等限制的一致性
func (s *OrderService) validateChannel(ctx context.Context, orderCtx *OrderCreateContext, channelID int64) *OrderError {
	// 从缓存获取渠道信息（缓存刷新服务每秒更新，确保一致性）
	channel, err := s.cacheService.GetPayChannel(ctx, channelID)
	if err != nil {
		return ErrChannelNotFound
	}

	if !channel.Status {
		return ErrChannelDisabled
	}

	// 检查时间范围（基于缓存的渠道信息）
	if err := s.checkChannelTime(channel); err != nil {
		return err
	}

	// 检查金额范围（基于缓存的渠道信息，确保限额一致性）
	if err := s.checkChannelAmount(channel, orderCtx); err != nil {
		return err
	}

	// 应用浮动加价（基于缓存的渠道信息）
	s.applyFloatAmount(channel, orderCtx)

	orderCtx.Channel = channel
	orderCtx.ChannelID = channelID
	orderCtx.PluginID = channel.PluginID

	return nil
}

// validatePlugin 验证插件
func (s *OrderService) validatePlugin(ctx context.Context, orderCtx *OrderCreateContext) *OrderError {
	if orderCtx.PluginID == 0 {
		return ErrPluginUnavailable
	}

	// 获取插件信息
	pluginInfo, err := s.pluginService.GetPlugin(ctx, orderCtx.PluginID)
	if err != nil {
		return ErrPluginUnavailable
	}

	if !pluginInfo.Status {
		return ErrPluginUnavailable
	}

	orderCtx.Plugin = pluginInfo

	// 获取插件上游类型
	upstream, err := s.pluginService.GetPluginUpstream(ctx, orderCtx.PluginID)
	if err != nil {
		return ErrPluginUnavailable
	}
	orderCtx.PluginUpstream = upstream

	// 获取支付类型（关键：获取 PayType 的 key，如 alipay_wap）
	payTypes, err := s.pluginService.GetPluginPayTypes(ctx, orderCtx.PluginID)
	if err != nil || len(payTypes) == 0 {
		return NewOrderError(ErrCodePayTypeUnavailable, "该通道不可用")
	}

	// 使用第一个支付类型
	payType := payTypes[0]
	if !payType.Status {
		return NewOrderError(ErrCodePayTypeUnavailable, "该通道不可用")
	}

	orderCtx.PayType = &payType
	// 关键：使用 PayType 的 Key 作为 PluginType（如 alipay_wap）
	// 插件管理器将使用这个 key 来创建对应的插件实例
	orderCtx.PluginType = payType.Key

	if orderCtx.PluginType == "" {
		return NewOrderError(ErrCodePayTypeUnavailable, "支付类型标识不能为空")
	}

	return nil
}

// validateDomain 验证域名（收银台）
// 参考 Python: order_check_domain(ctx)
func (s *OrderService) validateDomain(ctx context.Context, orderCtx *OrderCreateContext) *OrderError {
	// 1. 首先检查插件是否自己设置了域名
	// Python: domain_url = get_plugin_pay_domain(ctx.plugin.id, ctx.channel_id)
	domainURL := s.getPluginPayDomain(ctx, orderCtx.PluginID, orderCtx.ChannelID)
	if domainURL != "" {
		// 如果插件设置了域名，尝试从缓存的域名列表中查找
		domains, err := s.cacheService.GetAvailableDomains(ctx, 0) // 获取所有域名
		if err == nil {
			for i := range domains {
				if domains[i].URL == domainURL {
					orderCtx.DomainID = &domains[i].ID
					orderCtx.DomainURL = domains[i].URL
					orderCtx.Domain = &domains[i]
					return nil
				}
			}
		}
		// 如果缓存中没有，尝试从数据库查找（降级方案）
		var domain models.PayDomain
		if err := database.DB.Where("url = ?", domainURL).First(&domain).Error; err == nil {
			orderCtx.DomainID = &domain.ID
			orderCtx.DomainURL = domain.URL
			orderCtx.Domain = &domain
			return nil
		}
		// 如果数据库中没有，直接使用插件设置的域名URL
		orderCtx.DomainURL = domainURL
		// DomainID 保持为 nil，表示使用自定义域名
		return nil
	}

	// 2. 如果没有插件自定义域名，从缓存的可用域名列表中随机选择一个
	// Python: 根据 plugin_upstream 判断是否支持微信/支付宝
	domains, err := s.cacheService.GetAvailableDomains(ctx, orderCtx.PluginUpstream)
	if err != nil || len(domains) == 0 {
		return NewOrderError(ErrCodeCreateFailed, "无可用收银台")
	}

	// 从列表中随机选择一个（使用内存随机，避免数据库 RAND()）
	selectedDomain := domains[rand.Intn(len(domains))]

	orderCtx.DomainID = &selectedDomain.ID
	orderCtx.DomainURL = selectedDomain.URL
	orderCtx.Domain = &selectedDomain

	return nil
}

// getPluginPayDomain 获取插件自定义域名（使用带缓存的插件配置服务）
// 参考 Python: get_plugin_pay_domain(plugin_id: int, channel_id: int)
func (s *OrderService) getPluginPayDomain(ctx context.Context, pluginID, channelID int64) string {
	// 使用带缓存的插件配置服务
	config, err := s.pluginService.GetPluginConfigByKey(ctx, pluginID, "pay_domain")
	if err != nil {
		// 如果查询失败，返回空字符串
		return ""
	}

	// 解析 JSON 值
	var valueMap map[string]interface{}
	if err := json.Unmarshal([]byte(config.Value), &valueMap); err != nil {
		return ""
	}

	// 尝试获取域名值（可能是字符串或对象）
	if domainURL, ok := valueMap["value"].(string); ok {
		return domainURL
	}
	if domainURL, ok := valueMap["url"].(string); ok {
		return domainURL
	}
	// 如果 value 本身就是字符串
	if config.Value != "" && config.Value[0] != '{' {
		return config.Value
	}

	return ""
}

// validateBalance 检查余额（预检查）
// 参考 Python: 检查租户余额是否足够，然后检查码商余额
// 注意：这是预检查，最终检查在创建订单时使用 Redis 原子操作确保一致性
func (s *OrderService) validateBalance(ctx context.Context, orderCtx *OrderCreateContext) *OrderError {
	// 1. 先判断租户余额
	if orderCtx.Tenant != nil {
		// 直接从租户信息中获取余额和信任标志
		balance := orderCtx.Tenant.Balance
		trust := orderCtx.Tenant.Trust

		// 从 Redis 获取预占余额（全部依赖 Redis，默认 0）
		preTax, err := s.balanceService.GetPreTax(ctx, orderCtx.TenantID)
		if err != nil {
			// 如果获取失败，记录日志但不阻止订单创建（容错处理）
			logger.Logger.Warn("获取预占余额失败",
				zap.Int64("tenant_id", orderCtx.TenantID),
				zap.Error(err))
			preTax = 0 // 默认使用 0
		}

		// 计算可用余额：余额 - 预占用金额
		availableBalance := balance - preTax

		// 检查余额是否足够
		if availableBalance < int64(orderCtx.Money) {
			// 如果 trust=false（不允许负数），则拒绝订单
			if !trust {
				return ErrBalanceInsufficient
			}
			// 如果 trust=true（允许负数），允许继续
			// 注意：最终检查在创建订单时使用 Redis 原子操作确保一致性
		}
	}

	// 2. 再判断码商余额（如果存在码商ID）
	if orderCtx.Writeoff != nil {
		// 码商余额判断逻辑与租户相同
		// 直接从码商信息中获取余额
		if orderCtx.Writeoff.Balance != nil {
			// 检查码商余额是否足够（码商没有预占余额的概念，直接检查余额）
			if *orderCtx.Writeoff.Balance < int64(orderCtx.Money) {
				return NewOrderError(ErrCodeBalanceInsufficient, "码商余额不足")
			}
		}
		// 如果余额为 nil，表示无限制，允许继续
	}

	return nil
}

// checkChannelTime 检查渠道可用时间
func (s *OrderService) checkChannelTime(channel *models.PayChannel) *OrderError {
	if channel.StartTime == "00:00:00" && channel.EndTime == "00:00:00" {
		return nil // 全天可用
	}

	layout := "15:04:05"
	startTime, err1 := time.Parse(layout, channel.StartTime)
	endTime, err2 := time.Parse(layout, channel.EndTime)
	if err1 != nil || err2 != nil {
		return nil // 时间格式错误，跳过检查
	}

	now := time.Now()
	currentTime, _ := time.Parse(layout, now.Format(layout))

	// 支持跨零点时段
	if startTime.Before(endTime) {
		if currentTime.Before(startTime) || currentTime.After(endTime) {
			return NewOrderError(ErrCodeChannelTimeInvalid,
				fmt.Sprintf("通道不在可使用时间[%s-%s]", channel.StartTime, channel.EndTime))
		}
	} else if startTime.After(endTime) { // 跨0点情况
		if currentTime.Before(startTime) && currentTime.After(endTime) {
			return NewOrderError(ErrCodeChannelTimeInvalid,
				fmt.Sprintf("通道不在可使用时间[%s-%s]", channel.StartTime, channel.EndTime))
		}
	}

	return nil
}

// checkChannelAmount 检查渠道金额限制
// 注意：渠道信息来自缓存，缓存刷新服务每秒更新，确保限额一致性
func (s *OrderService) checkChannelAmount(channel *models.PayChannel, orderCtx *OrderCreateContext) *OrderError {
	// 检查固定金额模式
	if channel.Settled && channel.Moneys != "" {
		var moneys []int
		if err := json.Unmarshal([]byte(channel.Moneys), &moneys); err == nil {
			valid := false
			for _, m := range moneys {
				if m == orderCtx.Money {
					valid = true
					break
				}
			}
			if !valid {
				return NewOrderError(ErrCodeAmountOutOfRange,
					fmt.Sprintf("金额%d不在范围内,可用:%v", orderCtx.Money, moneys))
			}
		}
	}

	// 检查金额范围
	if channel.MinMoney > 0 || channel.MaxMoney > 0 {
		if orderCtx.Money < channel.MinMoney || orderCtx.Money > channel.MaxMoney {
			return NewOrderError(ErrCodeAmountOutOfRange,
				fmt.Sprintf("金额%d不在范围[%d,%d]内", orderCtx.Money, channel.MinMoney, channel.MaxMoney))
		}
	}

	return nil
}

// applyFloatAmount 应用浮动加价
func (s *OrderService) applyFloatAmount(channel *models.PayChannel, orderCtx *OrderCreateContext) {
	if channel.FloatMinMoney > 0 || channel.FloatMaxMoney > 0 {
		delta := channel.FloatMinMoney
		if channel.FloatMaxMoney > channel.FloatMinMoney {
			delta = channel.FloatMinMoney + rand.Intn(channel.FloatMaxMoney-channel.FloatMinMoney+1)
		}
		orderCtx.Money += delta
	}

	if orderCtx.Money <= 0 {
		// 如果金额变为0或负数，恢复原金额
		orderCtx.Money = orderCtx.NotifyMoney
	}
}

// createOrderAndDetail 创建订单和订单详情，返回订单详情ID（避免后续重复查询）
func (s *OrderService) createOrderAndDetail(ctx context.Context, orderCtx *OrderCreateContext) (orderDetailID int64, orderError *OrderError) {
	// 生成订单号
	orderCtx.OrderNo = utils.GenerateOrderNo()
	orderCtx.OrderID = utils.GenerateID()

	now := time.Now()

	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 使用 Redis 原子操作预占余额（在事务外，避免数据库锁）
	// 这是最终检查，确保余额足够才创建订单
	// 1. 先判断租户余额
	if orderCtx.Tenant != nil {
		success, _, err := s.balanceService.ReserveBalance(ctx, orderCtx.TenantID, int64(orderCtx.Money))
		if err != nil {
			tx.Rollback()
			return 0, NewOrderError(ErrCodeCreateFailed, fmt.Sprintf("预占余额失败: %v", err))
		}
		if !success {
			tx.Rollback()
			return 0, ErrBalanceInsufficient
		}
	}

	// 2. 再判断码商余额（如果存在码商信息）
	if orderCtx.Writeoff != nil {
		// 码商余额判断逻辑与租户相同（检查余额是否足够）
		// 码商没有预占余额的概念，直接检查余额
		if orderCtx.Writeoff.Balance != nil {
			// 检查码商余额是否足够
			if *orderCtx.Writeoff.Balance < int64(orderCtx.Money) {
				tx.Rollback()
				// 如果租户余额已预占，需要释放
				if orderCtx.Tenant != nil {
					_ = s.balanceService.ReleasePreTax(ctx, orderCtx.TenantID, int64(orderCtx.Money))
				}
				return 0, NewOrderError(ErrCodeBalanceInsufficient, "码商余额不足")
			}
		}
		// 如果余额为 nil，表示无限制，允许继续
	}

	// 构建通道名称（格式：[通道ID]通道名称）
	productName := ""
	if orderCtx.Channel != nil {
		productName = fmt.Sprintf("[%d]%s", orderCtx.Channel.ID, orderCtx.Channel.Name)
	}

	// 创建订单
	order := &models.Order{
		ID:             orderCtx.OrderID,
		OrderNo:        orderCtx.OrderNo,
		OutOrderNo:     orderCtx.OutOrderNo,
		OrderStatus:    models.OrderStatusPending,
		Money:          orderCtx.Money,
		Tax:            orderCtx.Tax,
		ProductName:    productName,
		ReqExtra:       orderCtx.Extra,
		CreateDatetime: &now,
		Compatible:     orderCtx.Compatible,
		Ver:            1,
		MerchantID:     &orderCtx.MerchantID,
		PayChannelID:   &orderCtx.ChannelID,
		WriteoffID:     orderCtx.WriteoffID,
	}

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return 0, NewOrderError(ErrCodeCreateFailed, fmt.Sprintf("创建订单失败: %v", err))
	}

	// 创建订单详情
	// 处理 Extra 字段：如果为空字符串，设置为 NULL 或有效的 JSON
	extraValue := orderCtx.Extra
	if extraValue == "" {
		// 空字符串时，设置为有效的空 JSON 对象
		extraValue = "{}"
	} else {
		// 验证是否为有效的 JSON，如果不是则包装为 JSON 字符串
		if !isValidJSON(extraValue) {
			// 如果不是有效的 JSON，将其作为字符串值包装在 JSON 中
			extraValue = fmt.Sprintf(`{"value":%q}`, extraValue)
		}
	}

	orderDetail := &models.OrderDetail{
		OrderID:        orderCtx.OrderID,
		NotifyURL:      orderCtx.NotifyURL,
		JumpURL:        orderCtx.JumpURL,
		NotifyMoney:    orderCtx.NotifyMoney,
		PluginType:     orderCtx.PluginType,
		PluginUpstream: orderCtx.PluginUpstream,
		CreateDatetime: &now,
		Extra:          extraValue,
		PluginID:       &orderCtx.PluginID,
		DomainID:       orderCtx.DomainID,
		WriteoffID:     orderCtx.WriteoffID,
		ProductID:      orderCtx.ProductID,
		CookieID:       orderCtx.CookieID,
	}

	if err := tx.Create(orderDetail).Error; err != nil {
		tx.Rollback()
		return 0, NewOrderError(ErrCodeCreateFailed, fmt.Sprintf("创建订单详情失败: %v", err))
	}

	// 注意：预占余额已在事务外通过 Redis 原子操作完成，这里不再需要数据库操作

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		// 如果事务提交失败，需要回滚 Redis 中的预占余额
		if orderCtx.Tenant != nil {
			_ = s.balanceService.ReleasePreTax(ctx, orderCtx.TenantID, int64(orderCtx.Money))
		}
		return 0, NewOrderError(ErrCodeCreateFailed, fmt.Sprintf("提交事务失败: %v", err))
	}

	// 注意：余额和预占余额已完全由 Redis 管理，不需要使缓存失效

	// 注意：订单日志由 Controller 层创建（包含响应信息）
	// 这里不再创建，避免重复

	return orderDetail.ID, nil
}

// createOrderLog 创建订单日志
// 参考 Python: 只创建 order_log，不更新
// 重要：订单事务成功后异步创建 order_log（只包含请求信息）
// 注意：响应信息由 Controller 层在创建时一并写入，所以这里只创建包含请求信息的日志
// 如果订单事务失败，此方法不会被调用，因此不需要检查订单是否存在
func (s *OrderService) createOrderLog(ctx context.Context, orderCtx *OrderCreateContext) {
	if orderCtx.OutOrderNo == "" {
		return
	}

	// 创建 order_log（只创建，不更新，只包含请求信息）
	// 不需要检查订单是否存在，因为只有在事务成功后才调用此方法
	now := time.Now()
	orderLog := &models.OrderLog{
		OutOrderNo:     orderCtx.OutOrderNo,
		SignRaw:        orderCtx.SignRaw,
		Sign:           orderCtx.Sign,
		RequestBody:    orderCtx.RequestBody,
		RequestMethod:  orderCtx.RequestMethod,
		CreateDatetime: &now,
	}

	if err := database.DB.Create(orderLog).Error; err != nil {
		logger.Logger.Warn("创建订单日志失败",
			zap.String("out_order_no", orderCtx.OutOrderNo),
			zap.Error(err))
	}
}

// updateOrderLogResponse 更新订单日志的响应信息
// 注意：这个方法在 Service 层调用，但实际响应格式由 Controller 层决定
// 参考 Python: 响应信息应该在 Controller 层记录，因为响应格式可能包含额外的包装
// 这个方法保留作为备用，但主要逻辑应该在 Controller 层
func (s *OrderService) updateOrderLogResponse(ctx context.Context, orderCtx *OrderCreateContext, response *CreateOrderResponse) {
	if orderCtx.OutOrderNo == "" {
		return
	}

	// 将响应转换为JSON字符串
	responseJSON, err := json.Marshal(response)
	if err != nil {
		logger.Logger.Warn("序列化订单响应失败",
			zap.String("out_order_no", orderCtx.OutOrderNo),
			zap.Error(err))
		return
	}

	now := time.Now()
	// 更新订单日志的响应信息
	database.DB.Model(&models.OrderLog{}).
		Where("out_order_no = ?", orderCtx.OutOrderNo).
		Updates(map[string]interface{}{
			"json_result":     string(responseJSON),
			"response_code":   "200",
			"update_datetime": &now,
		})
}

// generatePayURL 生成支付URL（使用插件系统）
// 此时产品已经通过 waitProduct 选择好了，直接使用 orderCtx 中的产品信息
func (s *OrderService) generatePayURL(ctx context.Context, orderCtx *OrderCreateContext) (string, *OrderError) {
	// 获取插件实例
	pluginInstance, err := s.pluginManager.GetPluginByCtx(ctx, orderCtx)
	if err != nil {
		return "", NewOrderError(ErrCodePluginUnavailable, fmt.Sprintf("插件不可用: %v", err))
	}

	// 使用已保存的订单详情ID，避免重复查询数据库
	orderDetailID := orderCtx.OrderDetailID
	if orderDetailID == 0 {
		// 降级方案：如果订单详情ID未保存，则查询数据库
		var orderDetail models.OrderDetail
		if err := database.DB.Where("order_id = ?", orderCtx.OrderID).First(&orderDetail).Error; err != nil {
			return "", NewOrderError(ErrCodeCreateFailed, fmt.Sprintf("获取订单详情失败: %v", err))
		}
		orderDetailID = orderDetail.ID
	}

	// 构建插件请求
	// 产品ID已经从 waitProduct 中选择好了，存储在 orderCtx.ProductID 中
	createReq := &plugin.CreateOrderRequest{
		OutOrderNo:     orderCtx.OutOrderNo,
		OrderNo:        orderCtx.OrderNo,
		OrderID:        orderCtx.OrderID,
		DetailID:       orderDetailID,
		ProductID:      orderCtx.ProductID, // 使用已选择的产品ID
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
		Compatible:     orderCtx.Compatible,
		Test:           orderCtx.Test,
	}

	// 添加关联对象（转换为 map）
	if orderCtx.Channel != nil {
		channelMap := map[string]interface{}{
			"id":        orderCtx.Channel.ID,
			"name":      orderCtx.Channel.Name,
			"status":    orderCtx.Channel.Status,
			"plugin_id": orderCtx.Channel.PluginID,
		}
		createReq.Channel = channelMap
	}

	if orderCtx.Plugin != nil {
		pluginMap := map[string]interface{}{
			"id":          orderCtx.Plugin.ID,
			"name":        orderCtx.Plugin.Name,
			"status":      orderCtx.Plugin.Status,
			"description": orderCtx.Plugin.Description,
		}
		createReq.Plugin = pluginMap
	}

	if orderCtx.PayType != nil {
		payTypeMap := map[string]interface{}{
			"id":     orderCtx.PayType.ID,
			"name":   orderCtx.PayType.Name,
			"key":    orderCtx.PayType.Key,
			"status": orderCtx.PayType.Status,
		}
		createReq.PayType = payTypeMap
	}

	// 调用插件创建订单
	createResp, err := pluginInstance.CreateOrder(ctx, createReq)
	if err != nil {
		// 记录插件调用失败的错误
		logger.Logger.Error("插件创建订单失败",
			zap.String("out_order_no", orderCtx.OutOrderNo),
			zap.String("order_no", orderCtx.OrderNo),
			zap.Int64("plugin_id", orderCtx.PluginID),
			zap.String("plugin_type", orderCtx.PluginType),
			zap.Error(err))
		return "", NewOrderError(ErrCodeCreateFailed, fmt.Sprintf("创建订单失败: %v", err))
	}

	// 检查响应
	if !createResp.IsSuccess() {
		// 记录插件返回的错误
		logger.Logger.Error("插件返回错误",
			zap.String("out_order_no", orderCtx.OutOrderNo),
			zap.String("order_no", orderCtx.OrderNo),
			zap.Int("error_code", createResp.ErrorCode),
			zap.String("error_message", createResp.ErrorMessage))
		// 即使失败也记录响应（用于调试）
		if err := s.recordPluginResponseToOrderDetailByID(ctx, orderCtx.OrderDetailID, createResp); err != nil {
			logger.Logger.Warn("记录插件响应失败",
				zap.String("out_order_no", orderCtx.OutOrderNo),
				zap.Error(err))
		}
		return "", NewOrderError(createResp.ErrorCode, createResp.ErrorMessage)
	}

	// 记录插件响应到订单详情的 Extra 字段（成功情况，已包含 pay_url）
	// 参考 Python: 上游响应应该被记录，包括成功和失败的情况
	// 注意：成功时已经包含 pay_url，后续 getAuthURL 中如果不需要鉴权就不需要再次更新
	if err := s.recordPluginResponseToOrderDetailByID(ctx, orderCtx.OrderDetailID, createResp); err != nil {
		logger.Logger.Warn("记录插件响应失败",
			zap.String("out_order_no", orderCtx.OutOrderNo),
			zap.String("order_no", orderCtx.OrderNo),
			zap.Error(err))
		// 记录失败不影响主流程，继续执行
	}

	return createResp.PayURL, nil
}

// getAuthURL 生成鉴权链接并返回收银台地址
// 参考 Python: get_auth_url 方法
// 如果域名需要鉴权，生成鉴权链接并返回收银台地址；否则直接返回支付URL
func (s *OrderService) getAuthURL(ctx context.Context, orderCtx *OrderCreateContext, payURL string) (string, *OrderError) {
	// 检查域名是否需要鉴权
	if orderCtx.Domain == nil || !orderCtx.Domain.AuthStatus || orderCtx.Domain.AuthKey == "" {
		// 不需要鉴权，直接返回支付URL
		// 注意：pay_url 已经在 recordPluginResponseToOrderDetailByID 中存储了，不需要再次存储
		return payURL, nil
	}

	// 需要鉴权时，确保支付URL已存储到订单详情的 Extra 字段中（收银台需要获取）
	// 注意：如果 recordPluginResponseToOrderDetailByID 已经存储了 pay_url，这里可以跳过
	// 但为了确保数据一致性，仍然更新一次（使用合并更新，避免重复查询）
	if err := s.ensurePayURLInOrderDetail(ctx, orderCtx.OrderDetailID, payURL); err != nil {
		return "", NewOrderError(ErrCodeCreateFailed, fmt.Sprintf("存储支付URL失败: %v", err))
	}

	// 生成鉴权链接并返回收银台地址
	// 注意：这里传入的是域名的密钥（p_key），不是动态生成的 auth_key
	cashierURL, err := s.generateAuthURLAndCashier(orderCtx.DomainURL, orderCtx.OrderNo, orderCtx.Domain.AuthKey, orderCtx.Domain.AuthTimeout, payURL)
	if err != nil {
		return "", NewOrderError(ErrCodeCreateFailed, fmt.Sprintf("生成鉴权链接失败: %v", err))
	}

	return cashierURL, nil
}

// recordPluginResponseToOrderDetailByID 记录插件响应到订单详情的 Extra 字段（使用订单详情ID，避免查询）
// 参考 Python: 上游响应应该被记录，包括成功和失败的情况
func (s *OrderService) recordPluginResponseToOrderDetailByID(ctx context.Context, orderDetailID int64, pluginResp *plugin.CreateOrderResponse) error {
	// 获取订单详情的 Extra 字段（只查询需要的字段）
	var extra string
	if err := database.DB.Model(&models.OrderDetail{}).
		Where("id = ?", orderDetailID).
		Select("extra").
		Scan(&extra).Error; err != nil {
		return err
	}

	// 解析现有的 Extra 字段
	extraMap := make(map[string]interface{})
	if extra != "" && extra != "{}" {
		if err := json.Unmarshal([]byte(extra), &extraMap); err != nil {
			// 如果解析失败，创建新的 map
			extraMap = make(map[string]interface{})
		}
	}

	// 构建插件响应数据
	pluginResponseData := map[string]interface{}{
		"success": pluginResp.Success,
	}
	if pluginResp.Success {
		// 成功情况：记录支付URL
		pluginResponseData["pay_url"] = pluginResp.PayURL
		if len(pluginResp.ExtraData) > 0 {
			pluginResponseData["extra_data"] = pluginResp.ExtraData
		}
	} else {
		// 失败情况：记录错误信息
		pluginResponseData["error_code"] = pluginResp.ErrorCode
		pluginResponseData["error_message"] = pluginResp.ErrorMessage
		if len(pluginResp.ExtraData) > 0 {
			pluginResponseData["extra_data"] = pluginResp.ExtraData
		}
	}

	// 添加插件响应到 Extra（使用 plugin_response 作为 key）
	extraMap["plugin_response"] = pluginResponseData

	// 如果成功且有支付URL，也单独存储 pay_url（用于向后兼容）
	if pluginResp.Success && pluginResp.PayURL != "" {
		extraMap["pay_url"] = pluginResp.PayURL
	}

	// 序列化并更新
	extraJSON, err := json.Marshal(extraMap)
	if err != nil {
		return err
	}

	// 更新订单详情（直接使用ID更新，避免查询）
	return database.DB.Model(&models.OrderDetail{}).
		Where("id = ?", orderDetailID).
		Update("extra", string(extraJSON)).Error
}

// recordPluginResponseToOrderDetail 记录插件响应到订单详情的 Extra 字段（保留原方法作为降级方案）
func (s *OrderService) recordPluginResponseToOrderDetail(ctx context.Context, orderID string, pluginResp *plugin.CreateOrderResponse) error {
	// 获取订单详情
	var orderDetail models.OrderDetail
	if err := database.DB.Where("order_id = ?", orderID).First(&orderDetail).Error; err != nil {
		return err
	}
	return s.recordPluginResponseToOrderDetailByID(ctx, orderDetail.ID, pluginResp)
}

// ensurePayURLInOrderDetail 确保支付URL已存储在订单详情的 Extra 字段中
// 优化：如果 pay_url 已存在且相同，则跳过更新，减少数据库操作
func (s *OrderService) ensurePayURLInOrderDetail(ctx context.Context, orderDetailID int64, payURL string) error {
	// 获取订单详情的 Extra 字段（只查询需要的字段）
	var extra string
	if err := database.DB.Model(&models.OrderDetail{}).
		Where("id = ?", orderDetailID).
		Select("extra").
		Scan(&extra).Error; err != nil {
		return err
	}

	// 解析现有的 Extra 字段
	extraMap := make(map[string]interface{})
	if extra != "" && extra != "{}" {
		if err := json.Unmarshal([]byte(extra), &extraMap); err != nil {
			// 如果解析失败，创建新的 map
			extraMap = make(map[string]interface{})
		}
	}

	// 检查 pay_url 是否已存在且相同（避免不必要的更新）
	if existingPayURL, ok := extraMap["pay_url"].(string); ok && existingPayURL == payURL {
		// pay_url 已存在且相同，无需更新
		return nil
	}

	// 添加或更新支付URL到 Extra
	extraMap["pay_url"] = payURL

	// 序列化并更新
	extraJSON, err := json.Marshal(extraMap)
	if err != nil {
		return err
	}

	// 更新订单详情（直接使用ID更新，避免查询）
	return database.DB.Model(&models.OrderDetail{}).
		Where("id = ?", orderDetailID).
		Update("extra", string(extraJSON)).Error
}

// storePayURLToOrderDetailByID 将支付URL存储到订单详情的 Extra 字段中（保留作为降级方案）
// 参考 Python: 存储支付URL以便收银台获取
func (s *OrderService) storePayURLToOrderDetailByID(ctx context.Context, orderDetailID int64, payURL string) error {
	return s.ensurePayURLInOrderDetail(ctx, orderDetailID, payURL)
}

// storePayURLToOrderDetail 将支付URL存储到订单详情的 Extra 字段中（保留原方法作为降级方案）
func (s *OrderService) storePayURLToOrderDetail(ctx context.Context, orderID string, payURL string) error {
	// 获取订单详情
	var orderDetail models.OrderDetail
	if err := database.DB.Where("order_id = ?", orderID).First(&orderDetail).Error; err != nil {
		return err
	}
	return s.storePayURLToOrderDetailByID(ctx, orderDetail.ID, payURL)
}

// generateAuthURLAndCashier 生成鉴权链接并返回收银台地址
// 参考 Python: 生成鉴权链接，格式为 {domain_url}/api/pay/auth?order_no={order_no}&auth_key={auth_key}&timestamp={timestamp}&sign={sign}
// 返回收银台地址，格式为 {domain_url}/cashier?order_no={order_no}
func (s *OrderService) generateAuthURLAndCashier(domainURL, orderNo, pKey string, authTimeout int, payURL string) (string, error) {
	// 生成时间戳
	timestamp := time.Now().Unix()

	// 使用 get_auth_key 方法生成动态鉴权密钥
	// 参考 Python: get_auth_key(raw, p_key, offset=30)
	// raw 使用订单号，p_key 使用域名的密钥，offset 默认30秒
	authKey := utils.GetAuthKey(orderNo, pKey, 30)

	// 构建鉴权参数
	authParams := map[string]interface{}{
		"order_no":  orderNo,
		"auth_key":  authKey,
		"timestamp": timestamp,
		"pay_url":   payURL, // 将支付URL作为参数传递
	}

	// 生成签名（使用 MD5）
	// 参考 Python: 按 key 排序，拼接成 key=value&key=value 格式，最后加上 &key={auth_key}，然后 MD5
	signData := make(map[string]interface{})
	for k, v := range authParams {
		if k != "sign" {
			signData[k] = v
		}
	}

	// 生成签名（使用 utils 的签名方法）
	_, sign := utils.GetSign(signData, authKey, nil, nil, 0)

	// 构建鉴权URL（用于验证，但不需要返回）
	// 格式：{domain_url}/api/pay/auth?order_no={order_no}&auth_key={auth_key}&timestamp={timestamp}&pay_url={pay_url}&sign={sign}
	// 注意：鉴权URL由收银台调用，这里只生成但不返回
	// 使用 url.QueryEscape 对 pay_url 进行编码
	_ = fmt.Sprintf("%s/api/pay/auth?order_no=%s&auth_key=%s&timestamp=%d&pay_url=%s&sign=%s",
		domainURL, orderNo, authKey, timestamp, url.QueryEscape(payURL), sign)

	// 返回收银台地址（格式：{domain_url}/cashier?order_no={order_no}）
	// 参考 Python: 收银台地址格式
	cashierURL := fmt.Sprintf("%s/cashier?order_no=%s", domainURL, orderNo)

	// 如果 auth_timeout > 0，添加过期时间参数
	if authTimeout > 0 {
		expireTime := timestamp + int64(authTimeout)
		cashierURL = fmt.Sprintf("%s&expire_time=%d", cashierURL, expireTime)
	}

	return cashierURL, nil
}

// setCache 设置缓存
func (s *OrderService) setCache(ctx context.Context, orderCtx *OrderCreateContext) {
	// 缓存订单号映射
	s.redis.Set(ctx, fmt.Sprintf("order_no:%s", orderCtx.OrderNo), orderCtx.OrderID, 24*time.Hour)
	s.redis.Set(ctx, fmt.Sprintf("out_order_no_map:%s", orderCtx.OutOrderNo), orderCtx.OrderNo, 24*time.Hour)
}

// buildResponse 构建响应
func (s *OrderService) buildResponse(orderCtx *OrderCreateContext, payURL string) *CreateOrderResponse {
	response := &CreateOrderResponse{}

	if orderCtx.Compatible == 1 {
		// 兼容模式
		response.TradeNo = orderCtx.OrderNo
		response.PayURL = payURL
		response.Msg = "订单创建成功"
		response.Code = 1
	} else {
		// 标准模式
		response.MchOrderNo = orderCtx.OutOrderNo
		response.PayOrderID = orderCtx.OrderNo
		response.PayURL2 = payURL

		// 生成响应签名
		dataMap := map[string]interface{}{
			"mchOrderNo": response.MchOrderNo,
			"payOrderId": response.PayOrderID,
			"payUrl":     response.PayURL2,
		}
		// 响应签名使用所有字段（useList 为 nil 表示使用所有字段）
		response.Sign = utils.GenerateResponseSign(dataMap, orderCtx.SignKey, orderCtx.Compatible)
	}

	return response
}

// GetOrderByOrderNo 根据订单号获取订单
func (s *OrderService) GetOrderByOrderNo(orderNo string) (*models.Order, error) {
	var order models.Order
	err := database.DB.Preload("OrderDetail").
		Preload("Merchant").
		Preload("PayChannel").
		Where("order_no = ?", orderNo).
		First(&order).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("订单不存在")
		}
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	return &order, nil
}

// GetOrderByOutOrderNo 根据商户订单号获取订单
func (s *OrderService) GetOrderByOutOrderNo(outOrderNo string, merchantID int64) (*models.Order, error) {
	var order models.Order
	err := database.DB.Preload("OrderDetail").
		Where("out_order_no = ? AND merchant_id = ?", outOrderNo, merchantID).
		First(&order).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("订单不存在")
		}
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	return &order, nil
}

// UpdateOrderStatus 更新订单状态，并处理预占余额和余额扣减
// 使用事务确保一致性，并处理预占余额的释放和余额的扣减
// 优化：从缓存获取商户信息（包含租户ID），避免在事务中查询数据库
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status int, ticketNo string) error {
	// 先查询订单信息（在事务外，减少事务时间）
	var order models.Order
	if err := database.DB.Select("id, merchant_id, money, order_status").
		Where("id = ?", orderID).
		First(&order).Error; err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}

	// 检查订单状态是否已经变更（避免重复处理）
	if order.OrderStatus == status {
		return nil // 状态未变化，直接返回
	}

	// 从缓存获取商户信息（包含租户ID），避免在事务中查询数据库
	var tenantID *int64
	if order.MerchantID != nil {
		merchant, _, err := s.cacheService.GetMerchantWithUser(context.Background(), *order.MerchantID)
		if err == nil && merchant != nil && merchant.ParentID > 0 {
			// ParentID 是 int64 类型，不是指针，直接使用
			tenantID = &merchant.ParentID
		}
		// 如果缓存未命中，降级方案：在事务外查询（但这种情况应该很少）
		if tenantID == nil {
			// 降级方案：在事务外查询（这种情况应该很少，因为缓存刷新服务每秒更新）
			var merchant models.Merchant
			if err := database.DB.Select("parent_id").Where("id = ?", *order.MerchantID).First(&merchant).Error; err == nil && merchant.ParentID > 0 {
				tenantID = &merchant.ParentID
			}
		}
	}

	// 开启事务（现在事务中只需要更新操作，不需要查询）
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	// 处理租户余额（在事务中，确保一致性）
	if tenantID != nil {
		// 根据订单状态处理预占余额和余额
		switch status {
		case models.OrderStatusPaid:
			// 订单支付成功：从数据库扣减余额，从 Redis 释放预占
			// 先查询变更前的余额（使用 SELECT FOR UPDATE 确保一致性），用于记录流水
			var tenant models.Tenant
			if err := tx.Set("gorm:query_option", "FOR UPDATE").
				Where("id = ?", *tenantID).First(&tenant).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("查询租户余额失败: %w", err)
			}

			oldBalance := tenant.Balance
			newBalance := oldBalance - int64(order.Money)

			// 使用原子操作扣减余额
			if err := tx.Model(&models.Tenant{}).
				Where("id = ?", *tenantID).
				Update("balance", gorm.Expr("balance - ?", order.Money)).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("扣减租户余额失败: %w", err)
			}

			// 记录租户资金流水
			cashflow := &models.TenantCashflow{
				OldMoney:       oldBalance,
				NewMoney:       newBalance,
				ChangeMoney:    -int64(order.Money), // 负数表示扣减
				FlowType:       models.CashflowTypeOrderDeduct,
				OrderID:        &orderID,
				PayChannelID:   order.PayChannelID,
				TenantID:       *tenantID,
				CreateDatetime: &now,
			}
			if err := tx.Create(cashflow).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("记录租户资金流水失败: %w", err)
			}

			// 从 Redis 释放预占余额
			if err := s.balanceService.ReleasePreTax(ctx, *tenantID, int64(order.Money)); err != nil {
				// 如果 Redis 释放失败，记录日志但不回滚数据库事务（数据库已扣减）
				logger.Logger.Warn("释放预占余额失败",
					zap.Int64("tenant_id", *tenantID),
					zap.Int64("amount", int64(order.Money)),
					zap.Error(err))
			}
		case models.OrderStatusFailed, models.OrderStatusCancelled, models.OrderStatusExpired:
			// 订单失败/取消/过期：只从 Redis 释放预占，不扣减余额
			if err := s.balanceService.ReleasePreTax(ctx, *tenantID, int64(order.Money)); err != nil {
				tx.Rollback()
				return fmt.Errorf("释放预占余额失败: %w", err)
			}
			// OrderStatusPending 不需要处理（创建订单时已增加预占）
		}
	}

	// 处理码商余额（在事务中，确保一致性）
	// 码商余额扣减逻辑与租户相同：订单支付成功时从数据库扣减余额
	if order.WriteoffID != nil {
		switch status {
		case models.OrderStatusPaid:
			// 订单支付成功：从数据库扣减码商余额
			// 先查询变更前的余额（使用 SELECT FOR UPDATE 确保一致性），用于记录流水
			var writeoff models.Writeoff
			if err := tx.Set("gorm:query_option", "FOR UPDATE").
				Where("id = ?", *order.WriteoffID).First(&writeoff).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("查询码商余额失败: %w", err)
			}

			// 记录码商资金流水（无论余额是否为 NULL，都需要记录流水以便追踪）
			var oldBalance, newBalance int64
			if writeoff.Balance != nil {
				// 有余额限制：扣减余额并记录实际余额变化
				oldBalance = *writeoff.Balance
				newBalance = oldBalance - int64(order.Money)

				// 使用原子操作扣减余额
				if err := tx.Model(&models.Writeoff{}).
					Where("id = ? AND balance IS NOT NULL", *order.WriteoffID).
					Update("balance", gorm.Expr("balance - ?", order.Money)).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("扣减码商余额失败: %w", err)
				}
			} else {
				// 余额无限制：old_money 和 new_money 都设置为 0（表示无限制）
				// 这样仍然可以记录流水以便追踪订单使用了哪些码商
				oldBalance = 0
				newBalance = 0
			}

			// 记录码商资金流水
			cashflow := &models.WriteoffCashflow{
				OldMoney:       oldBalance,
				NewMoney:       newBalance,
				ChangeMoney:    -int64(order.Money), // 负数表示扣减
				FlowType:       models.CashflowTypeOrderDeduct,
				Tax:            0.00, // 费率，可以根据需要设置
				OrderID:        &orderID,
				PayChannelID:   order.PayChannelID,
				WriteoffID:     *order.WriteoffID,
				CreateDatetime: &now,
			}
			if err := tx.Create(cashflow).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("记录码商资金流水失败: %w", err)
			}
		case models.OrderStatusFailed, models.OrderStatusCancelled, models.OrderStatusExpired:
			// 订单失败/取消/过期：码商余额不需要处理（码商没有预占余额的概念）
			// 码商余额只在支付成功时扣减
		}
	}

	// 先更新版本号（使用原子操作）
	if err := tx.Exec("UPDATE dvadmin_order SET ver = ver + 1 WHERE id = ?", orderID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新版本号失败: %w", err)
	}

	// 更新订单状态（合并多个字段更新，减少数据库往返）
	updates := map[string]interface{}{
		"order_status":    status,
		"update_datetime": &now,
	}

	if status == models.OrderStatusPaid {
		updates["pay_datetime"] = &now
	}

	if err := tx.Model(&models.Order{}).
		Where("id = ?", orderID).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	// 更新订单详情的 ticket_no（如果提供）
	// 优化：与订单状态更新合并到同一个事务中，但分开执行以便于错误处理
	if ticketNo != "" {
		if err := tx.Model(&models.OrderDetail{}).
			Where("order_id = ?", orderID).
			Update("ticket_no", ticketNo).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("更新订单详情失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		// 如果事务提交失败，需要回滚 Redis 中的预占余额操作
		if tenantID != nil {
			switch status {
			case models.OrderStatusPaid:
				// 回滚：重新预占（因为已经释放了）
				// 注意：数据库余额扣减已回滚（事务回滚），只需要恢复 Redis 预占
				_, _, _ = s.balanceService.ReserveBalance(ctx, *tenantID, int64(order.Money))
			case models.OrderStatusFailed, models.OrderStatusCancelled, models.OrderStatusExpired:
				// 回滚：重新预占（因为已经释放了）
				_, _, _ = s.balanceService.ReserveBalance(ctx, *tenantID, int64(order.Money))
			}
		}
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 注意：余额从数据库扣减，预占余额由 Redis 管理，缓存刷新服务会定期从数据库同步余额到 Redis

	return nil
}
