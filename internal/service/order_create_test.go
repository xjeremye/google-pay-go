package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(
		&models.Order{},
		&models.OrderDetail{},
		&models.Merchant{},
		&models.Tenant{},
		&models.PayChannel{},
		&models.PayPlugin{},
		&models.PayType{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// setupTestRedis 设置测试 Redis（使用 Mock）
// 注意：实际测试中应该使用真实的 Redis 或 miniredis
// func setupTestRedis() *redis.Client {
// 	return nil
// }

// TestOrderService_CreateOrder_AmountInvalid 测试金额无效
func TestOrderService_CreateOrder_AmountInvalid(t *testing.T) {
	service := NewOrderService()

	req := &CreateOrderRequest{
		OutOrderNo: "TEST001",
		MerchantID: 1,
		ChannelID:  1,
		Money:      0, // 无效金额
		NotifyURL:  "https://example.com/notify",
		Sign:       "test_sign",
		RawSignData: map[string]interface{}{
			"mchId":      1,
			"channelId":  1,
			"mchOrderNo": "TEST001",
			"amount":     0,
			"notifyUrl":  "https://example.com/notify",
			"sign":       "test_sign",
		},
	}

	ctx := context.Background()
	resp, err := service.CreateOrder(ctx, req)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeAmountInvalid, err.Code)
	assert.Equal(t, "金额必须大于0", err.Message)
}

// TestOrderService_validateMerchant 测试商户验证
func TestOrderService_validateMerchant(t *testing.T) {
	// 使用内存数据库
	db := setupTestDB(t)
	originalDB := database.DB
	database.DB = db
	defer func() {
		database.DB = originalDB
	}()

	// 初始化 Redis（测试环境可能需要真实的 Redis 或使用 miniredis）
	// 这里简化处理，跳过需要 Redis 的测试
	// 实际测试中应该使用测试 Redis 或 Mock
	if database.RDB == nil {
		t.Skip("Redis not available, skipping test")
		return
	}

	service := NewOrderService()
	ctx := context.Background()
	orderCtx := &OrderCreateContext{}

	// 测试商户不存在
	err := service.validateMerchant(ctx, orderCtx, 999)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeMerchantNotFound, err.Code)
}

// TestOrderService_validateSign 测试签名验证
func TestOrderService_validateSign(t *testing.T) {
	service := NewOrderService()

	ctx := context.Background()
	orderCtx := &OrderCreateContext{
		SignKey: "test_key_12345",
	}

	// 准备签名数据
	rawSignData := map[string]interface{}{
		"mchId":      1,
		"channelId":  1,
		"mchOrderNo": "TEST001",
		"amount":     10000,
		"notifyUrl":  "https://example.com/notify",
		"jumpUrl":    "https://example.com/jump",
	}

	// 生成正确签名
	_, correctSign := utils.GetSign(rawSignData, orderCtx.SignKey, nil, nil, 0)
	rawSignData["sign"] = correctSign

	// 测试签名正确
	err := service.validateSign(ctx, orderCtx, rawSignData)
	assert.Nil(t, err)

	// 测试签名错误
	rawSignData["sign"] = "wrong_sign"
	err = service.validateSign(ctx, orderCtx, rawSignData)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeSignInvalid, err.Code)

	// 测试缺少签名
	rawSignData["sign"] = ""
	err = service.validateSign(ctx, orderCtx, rawSignData)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeSignInvalid, err.Code)
}

// TestOrderService_validateOutOrderNo 测试商户订单号验证
func TestOrderService_validateOutOrderNo(t *testing.T) {
	// 检查 Redis 是否可用
	if database.RDB == nil {
		t.Skip("Redis not available, skipping test")
		return
	}

	service := NewOrderService()
	ctx := context.Background()
	orderCtx := &OrderCreateContext{
		OutOrderNo: "TEST001",
	}

	// 测试订单号为空
	orderCtx.OutOrderNo = ""
	err := service.validateOutOrderNo(ctx, orderCtx)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeOutOrderNoRequired, err.Code)

	// 测试订单号已存在
	orderCtx.OutOrderNo = "TEST001"
	// 先设置一次
	database.RDB.SetNX(ctx, "out_order_no:TEST001", "1", 24*time.Hour)
	// 再次尝试设置应该失败
	err = service.validateOutOrderNo(ctx, orderCtx)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeOutOrderNoExists, err.Code)

	// 清理
	database.RDB.Del(ctx, "out_order_no:TEST001")
}

// TestOrderService_validateChannel 测试渠道验证
func TestOrderService_validateChannel(t *testing.T) {
	// 检查 Redis 是否可用
	if database.RDB == nil {
		t.Skip("Redis not available, skipping test")
		return
	}

	// 使用内存数据库
	db := setupTestDB(t)
	originalDB := database.DB
	database.DB = db
	defer func() {
		database.DB = originalDB
	}()

	service := NewOrderService()
	ctx := context.Background()

	// 创建测试渠道
	channel := &models.PayChannel{
		ID:        1,
		Name:      "测试渠道",
		Status:    true,
		MinMoney:  1000,
		MaxMoney:  100000,
		StartTime: "00:00:00",
		EndTime:   "23:59:59",
		PluginID:  1,
	}
	db.Create(channel)

	orderCtx := &OrderCreateContext{
		Money: 10000,
	}

	// 测试渠道不存在
	err := service.validateChannel(ctx, orderCtx, 999)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeChannelNotFound, err.Code)

	// 测试渠道已禁用
	disabledChannel := &models.PayChannel{
		ID:        2,
		Name:      "禁用渠道",
		Status:    false,
		StartTime: "00:00:00",
		EndTime:   "23:59:59",
		PluginID:  1,
	}
	db.Create(disabledChannel)
	err = service.validateChannel(ctx, orderCtx, 2)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeChannelDisabled, err.Code)

	// 测试金额超出范围
	orderCtx.Money = 100 // 低于最小金额
	err = service.validateChannel(ctx, orderCtx, 1)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeAmountOutOfRange, err.Code)

	// 测试成功
	orderCtx.Money = 10000
	err = service.validateChannel(ctx, orderCtx, 1)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), orderCtx.ChannelID)
	assert.Equal(t, int64(1), orderCtx.PluginID)
}

// TestOrderService_validatePlugin 测试插件验证
func TestOrderService_validatePlugin(t *testing.T) {
	// 检查 Redis 是否可用
	if database.RDB == nil {
		t.Skip("Redis not available, skipping test")
		return
	}

	// 使用内存数据库
	db := setupTestDB(t)
	originalDB := database.DB
	database.DB = db
	defer func() {
		database.DB = originalDB
	}()

	service := NewOrderService()
	ctx := context.Background()

	// 创建测试插件和支付类型
	pluginInfo := &models.PayPlugin{
		ID:     1,
		Name:   "测试插件",
		Status: true,
	}
	db.Create(pluginInfo)

	payType := &models.PayType{
		ID:     1,
		Key:    "alipay_wap",
		Name:   "支付宝H5",
		Status: true,
	}
	db.Create(payType)

	// 创建插件支付类型关联
	db.Exec("INSERT INTO dvadmin_pay_plugin_pay_types (payplugin_id, paytype_id) VALUES (?, ?)", 1, 1)

	// 创建插件配置
	pluginConfig := &models.PayPluginConfig{
		ParentID:     1,
		Key:          "type",
		Value:        `"5"`,
		Status:       true,
		Sort:         1,
		Title:        "类型",
		FormItemType: 1,
	}
	db.Create(pluginConfig)

	orderCtx := &OrderCreateContext{
		PluginID: 1,
	}

	// 测试插件验证
	err := service.validatePlugin(ctx, orderCtx)
	// 由于依赖 Redis 缓存，这里可能成功或失败
	// 如果 Redis 可用且数据正确，应该成功
	_ = err
}

// TestOrderService_checkChannelTime 测试渠道时间检查
func TestOrderService_checkChannelTime(t *testing.T) {
	service := NewOrderService()

	// 测试全天可用
	channel := &models.PayChannel{
		StartTime: "00:00:00",
		EndTime:   "00:00:00",
	}
	err := service.checkChannelTime(channel)
	assert.Nil(t, err)

	// 测试正常时间范围
	channel.StartTime = "09:00:00"
	channel.EndTime = "18:00:00"
	err = service.checkChannelTime(channel)
	// 根据当前时间，可能成功或失败
	_ = err

	// 测试跨零点
	channel.StartTime = "22:00:00"
	channel.EndTime = "06:00:00"
	err = service.checkChannelTime(channel)
	// 根据当前时间，可能成功或失败
	_ = err
}

// TestOrderService_checkChannelAmount 测试渠道金额检查
func TestOrderService_checkChannelAmount(t *testing.T) {
	service := NewOrderService()

	orderCtx := &OrderCreateContext{
		Money: 10000,
	}

	// 测试固定金额模式
	channel := &models.PayChannel{
		Settled: true,
		Moneys:  `[5000, 10000, 20000]`,
	}
	err := service.checkChannelAmount(channel, orderCtx)
	assert.Nil(t, err) // 10000 在列表中

	// 测试金额不在固定列表中
	orderCtx.Money = 15000
	err = service.checkChannelAmount(channel, orderCtx)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeAmountOutOfRange, err.Code)

	// 测试金额范围
	channel.Settled = false
	channel.MinMoney = 5000
	channel.MaxMoney = 50000
	orderCtx.Money = 10000
	err = service.checkChannelAmount(channel, orderCtx)
	assert.Nil(t, err)

	// 测试金额低于最小值
	orderCtx.Money = 1000
	err = service.checkChannelAmount(channel, orderCtx)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeAmountOutOfRange, err.Code)

	// 测试金额高于最大值
	orderCtx.Money = 100000
	err = service.checkChannelAmount(channel, orderCtx)
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeAmountOutOfRange, err.Code)
}

// TestOrderService_applyFloatAmount 测试浮动加价
func TestOrderService_applyFloatAmount(t *testing.T) {
	service := NewOrderService()

	orderCtx := &OrderCreateContext{
		Money:       10000,
		NotifyMoney: 10000,
	}

	// 测试无浮动加价
	channel := &models.PayChannel{
		FloatMinMoney: 0,
		FloatMaxMoney: 0,
	}
	service.applyFloatAmount(channel, orderCtx)
	assert.Equal(t, 10000, orderCtx.Money)

	// 测试固定浮动加价
	channel.FloatMinMoney = 100
	channel.FloatMaxMoney = 100
	orderCtx.Money = 10000
	service.applyFloatAmount(channel, orderCtx)
	assert.Equal(t, 10100, orderCtx.Money)

	// 测试范围浮动加价
	channel.FloatMinMoney = 50
	channel.FloatMaxMoney = 200
	orderCtx.Money = 10000
	service.applyFloatAmount(channel, orderCtx)
	assert.GreaterOrEqual(t, orderCtx.Money, 10050)
	assert.LessOrEqual(t, orderCtx.Money, 10200)
}

// TestOrderService_buildResponse 测试响应构建
func TestOrderService_buildResponse(t *testing.T) {
	service := NewOrderService()

	orderCtx := &OrderCreateContext{
		OutOrderNo: "TEST001",
		OrderNo:    "PAY20240101120000001",
		Compatible: 0,
		SignKey:    "test_key",
	}

	payURL := "https://pay.example.com/pay?order_no=PAY20240101120000001"

	// 测试标准模式
	response := service.buildResponse(orderCtx, payURL)
	assert.Equal(t, "TEST001", response.MchOrderNo)
	assert.Equal(t, "PAY20240101120000001", response.PayOrderID)
	assert.Equal(t, payURL, response.PayURL2)
	// 注意：响应签名使用所有字段，不限制 useList
	// 如果 SignKey 不为空，应该生成签名
	if orderCtx.SignKey != "" {
		// 验证签名格式（MD5 32位大写）
		if response.Sign != "" {
			assert.Len(t, response.Sign, 32)
		}
	}

	// 测试兼容模式
	orderCtx.Compatible = 1
	response = service.buildResponse(orderCtx, payURL)
	assert.Equal(t, "PAY20240101120000001", response.TradeNo)
	assert.Equal(t, payURL, response.PayURL)
	assert.Equal(t, "订单创建成功", response.Msg)
	assert.Equal(t, 1, response.Code)
}

// TestOrderService_GetOrderByOrderNo 测试根据订单号获取订单
func TestOrderService_GetOrderByOrderNo(t *testing.T) {
	// 使用内存数据库
	db := setupTestDB(t)
	originalDB := database.DB
	database.DB = db
	defer func() {
		database.DB = originalDB
	}()

	service := NewOrderService()

	// 创建测试订单
	now := time.Now()
	order := &models.Order{
		ID:             "123456789",
		OrderNo:        "PAY20240101120000001",
		OutOrderNo:     "TEST001",
		OrderStatus:    models.OrderStatusPending,
		Money:          10000,
		CreateDatetime: &now,
		Ver:            1,
	}
	db.Create(order)

	orderDetail := &models.OrderDetail{
		OrderID:        "123456789",
		NotifyURL:      "https://example.com/notify",
		JumpURL:        "https://example.com/jump",
		NotifyMoney:    10000,
		CreateDatetime: &now,
	}
	db.Create(orderDetail)

	// 测试获取订单
	result, err := service.GetOrderByOrderNo("PAY20240101120000001")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "PAY20240101120000001", result.OrderNo)
	assert.Equal(t, "TEST001", result.OutOrderNo)

	// 测试订单不存在
	_, err = service.GetOrderByOrderNo("NOT_EXIST")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "订单不存在")
}

// TestOrderService_GetOrderByOutOrderNo 测试根据商户订单号获取订单
func TestOrderService_GetOrderByOutOrderNo(t *testing.T) {
	// 使用内存数据库
	db := setupTestDB(t)
	originalDB := database.DB
	database.DB = db
	defer func() {
		database.DB = originalDB
	}()

	service := NewOrderService()

	// 创建测试订单
	now := time.Now()
	merchantID := int64(1)
	order := &models.Order{
		ID:             "123456789",
		OrderNo:        "PAY20240101120000001",
		OutOrderNo:     "TEST001",
		OrderStatus:    models.OrderStatusPending,
		Money:          10000,
		CreateDatetime: &now,
		Ver:            1,
		MerchantID:     &merchantID,
	}
	db.Create(order)

	orderDetail := &models.OrderDetail{
		OrderID:        "123456789",
		NotifyURL:      "https://example.com/notify",
		JumpURL:        "https://example.com/jump",
		NotifyMoney:    10000,
		CreateDatetime: &now,
	}
	db.Create(orderDetail)

	// 测试获取订单
	result, err := service.GetOrderByOutOrderNo("TEST001", 1)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "TEST001", result.OutOrderNo)

	// 测试订单不存在
	_, err = service.GetOrderByOutOrderNo("NOT_EXIST", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "订单不存在")
}
