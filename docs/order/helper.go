package order

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/example/payment-core/internal/infra"
	"github.com/example/payment-core/internal/plugin"
	"github.com/google/uuid"
)

// Helper 订单工具类，包含所有工具类相关的逻辑
type Helper struct {
	service *Service
}

// NewHelper 创建工具类
func NewHelper(service *Service) *Helper {
	return &Helper{
		service: service,
	}
}

// PreCheckExtra 预检查 extra
func (v *Helper) PreCheckExtra(ctx context.Context, orderCtx *OrderCreateCtx) (bool, string) {
	// TODO: 实现预检查 extra
	return false, ""
}

func (v *Helper) checkAndGetMerchant(ctx context.Context, orderCtx *OrderCreateCtx) (map[string]interface{}, *OrderProcessingError) {
	errData := &OrderProcessingError{
		Code:    7301,
		Message: "商户不存在",
	}

	merchant, err := infra.GetMerchantByID(ctx, v.service.redis, int64(orderCtx.MerchantID))
	if err != nil || merchant == nil {
		return nil, errData
	}

	// 从 map 中提取 system_user_id
	systemUserID, ok := merchant["system_user_id"]
	if !ok || systemUserID == nil {
		return nil, errData
	}

	// 处理不同类型的 system_user_id
	var userID int64
	switch v := systemUserID.(type) {
	case float64:
		userID = int64(v)
	case int64:
		userID = v
	case int:
		userID = int64(v)
	default:
		return nil, errData
	}

	merchantUser, err := infra.GetUserByID(ctx, v.service.redis, userID)
	if err != nil || merchantUser == nil {
		return nil, &OrderProcessingError{
			Code:    7302,
			Message: "商户已被禁用,请联系管理员",
		}
	}

	// 检查 status 字段（可能是 bool 或 float64）
	status := false
	if statusVal, ok := merchantUser["status"]; ok {
		if b, ok := statusVal.(bool); ok {
			status = b
		} else if f, ok := statusVal.(float64); ok {
			status = f != 0
		}
	}
	if !status {
		return nil, &OrderProcessingError{
			Code:    7302,
			Message: "商户已被禁用,请联系管理员",
		}
	}

	// 提取 key 字段
	if keyVal, ok := merchantUser["key"]; ok && keyVal != nil {
		if keyStr, ok := keyVal.(string); ok && keyStr != "" {
			orderCtx.SignKey = keyStr
		}
	}

	return merchant, nil
}

// CheckMerchant 检查商户
func (v *Helper) CheckMerchant(ctx context.Context, orderCtx *OrderCreateCtx, merchantID int) *OrderProcessingError {
	orderCtx.MerchantID = merchantID
	_, err := v.checkAndGetMerchant(ctx, orderCtx)
	if err != nil {
		return err
	}
	return nil
}

// CheckTenant 检查租户
func (v *Helper) CheckTenant(ctx context.Context, orderCtx *OrderCreateCtx) *OrderProcessingError {
	merchant, err := v.checkAndGetMerchant(ctx, orderCtx)
	if err != nil {
		return err
	}

	errData := &OrderProcessingError{
		Code:    7302,
		Message: "商户上级已被禁用,请联系管理员",
	}

	// 从 merchant map 中获取 parent_id
	parentID := int64(0)
	if merchant != nil {
		if pid, ok := merchant["parent_id"].(float64); ok {
			parentID = int64(pid)
		} else if pid, ok := merchant["parent_id"].(int64); ok {
			parentID = pid
		}
	}

	tenant, tenantErr := infra.GetTenantByID(ctx, v.service.redis, parentID)
	if tenantErr != nil || tenant == nil {
		return errData
	}

	// 从 map 中提取 system_user_id
	systemUserID, ok := tenant["system_user_id"]
	if !ok || systemUserID == nil {
		return errData
	}

	// 处理不同类型的 system_user_id
	var userID int64
	switch v := systemUserID.(type) {
	case float64:
		userID = int64(v)
	case int64:
		userID = v
	case int:
		userID = int64(v)
	default:
		return errData
	}

	tenantUser, tenantUserErr := infra.GetUserByID(ctx, v.service.redis, userID)
	if tenantUserErr != nil || tenantUser == nil {
		return errData
	}

	// 检查 status 字段
	status := false
	if statusVal, ok := tenantUser["status"]; ok {
		if b, ok := statusVal.(bool); ok {
			status = b
		} else if f, ok := statusVal.(float64); ok {
			status = f != 0
		}
	}
	if !status {
		return errData
	}

	// 提取 tenant id
	if idVal, ok := tenant["id"]; ok {
		switch v := idVal.(type) {
		case float64:
			orderCtx.TenantID = int(v)
		case int64:
			orderCtx.TenantID = int(v)
		case int:
			orderCtx.TenantID = v
		}
	}
	return nil
}

// CheckSign 检查签名
func (v *Helper) CheckSign(ctx context.Context, orderCtx *OrderCreateCtx, rawSignData map[string]interface{}) *OrderProcessingError {

	errData := &OrderProcessingError{
		Code:    7304,
		Message: "签名验证失败",
	}

	key := orderCtx.SignKey
	sign, _ := rawSignData["sign"].(string)

	if key == "" || sign == "" {
		return errData
	}

	// 生成签名
	_, actualSign := GetSign(rawSignData, key, nil, nil, orderCtx.Compatible)

	if sign != actualSign {
		fmt.Println("正确签名:", actualSign)
		return errData
	}

	return nil
}

// CheckOutOrderNo 检查外部订单号
func (v *Helper) CheckOutOrderNo(ctx context.Context, orderCtx *OrderCreateCtx) *OrderProcessingError {
	// 基础校验
	if orderCtx.OutOrderNo == "" {
		return &OrderProcessingError{
			Code:    7321,
			Message: "商户订单号不能为空",
		}
	}

	// 使用 Redis SetNX 做高并发下的商户订单号幂等控制
	// 维度：全局 out_order_no（全系统唯一）
	// 首次写入成功的请求才允许继续后续流程，其余并发请求直接视为重复单
	key := fmt.Sprintf("out_order_no:%s", orderCtx.OutOrderNo)

	ok, err := v.service.redis.SetNX(ctx, key, "1", 24*time.Hour).Result()
	if err != nil {
		// Redis 异常时，避免产生不确定的并发行为，直接中止，提示系统繁忙
		return &OrderProcessingError{
			Code:    7321,
			Message: "系统繁忙,请稍后重试",
		}
	}

	if !ok {
		// 已经存在相同的商户订单号，认为是重复请求
		return &OrderProcessingError{
			Code:    7321,
			Message: "商户订单号已存在",
		}
	}

	return nil
}

// CheckChannel 检查渠道
func (v *Helper) CheckChannel(ctx context.Context, orderCtx *OrderCreateCtx, channelID int) *OrderProcessingError {
	channel, err := infra.GetChannelByID(ctx, v.service.redis, int64(channelID))
	if err != nil || channel == nil {
		return &OrderProcessingError{
			Code:    7305,
			Message: "渠道不存在",
		}
	}
	orderCtx.ChannelID = channelID
	orderCtx.Channel = channel

	// 提取 plugin_id
	if pluginIDVal, ok := channel["plugin_id"]; ok {
		switch v := pluginIDVal.(type) {
		case float64:
			orderCtx.PluginID = int(v)
		case int64:
			orderCtx.PluginID = int(v)
		case int:
			orderCtx.PluginID = v
		}
	}

	plugin, err := infra.GetPluginByID(ctx, v.service.redis, int64(orderCtx.PluginID))
	if err != nil || plugin == nil {
		orderCtx.Plugin = nil
	} else {
		orderCtx.Plugin = plugin
	}

	if orderCtx.Test {
		return nil
	}

	// 检查 status 字段
	status := false
	if statusVal, ok := channel["status"]; ok {
		if b, ok := statusVal.(bool); ok {
			status = b
		} else if f, ok := statusVal.(float64); ok {
			status = f != 0
		}
	}
	if !status {
		return &OrderProcessingError{
			Code:    7306,
			Message: "渠道已被禁用,请联系管理员",
		}
	}

	// 检查通道可用时间
	startTimeStr, _ := channel["start_time"].(string)
	endTimeStr, _ := channel["end_time"].(string)
	// 格式如 "15:04:05"
	if startTimeStr != "" && endTimeStr != "" && startTimeStr != "00:00:00" && endTimeStr != "00:00:00" {
		layout := "15:04:05"
		startTime, err1 := time.Parse(layout, startTimeStr)
		endTime, err2 := time.Parse(layout, endTimeStr)
		nowTime := time.Now()
		currentTime, _ := time.Parse(layout, nowTime.Format(layout))

		if err1 == nil && err2 == nil {
			// 支持跨零点时段
			if startTime.Before(endTime) {
				if currentTime.Before(startTime) || currentTime.After(endTime) {
					return &OrderProcessingError{
						Code:    7309,
						Message: fmt.Sprintf("通道不在可使用时间[%s-%s]", startTimeStr, endTimeStr),
					}
				}
			} else if startTime.After(endTime) { // 跨0点情况
				if currentTime.Before(startTime) && currentTime.After(endTime) {
					return &OrderProcessingError{
						Code:    7309,
						Message: fmt.Sprintf("通道不在可使用时间[%s-%s]", startTimeStr, endTimeStr),
					}
				}
			}
		}
	}

	// 通道固定金额模式
	// 处理 moneys 字段：可能是 []interface{}（直接数组）或 string（JSON 字符串，需要解析）
	var moneys []interface{}
	channelMap := orderCtx.Channel
	if moneysRaw, exists := channelMap["moneys"]; exists && moneysRaw != nil {
		switch v := moneysRaw.(type) {
		case []interface{}:
			// 直接是数组类型
			moneys = v
		case string:
			// 是 JSON 字符串，需要解析
			if err := json.Unmarshal([]byte(v), &moneys); err != nil {
				// JSON 解析失败，忽略此字段（不进行金额校验）
				moneys = nil
			}
		}
	}

	if len(moneys) > 0 {
		valid := false
		for _, m := range moneys {
			if moneyVal, ok := m.(float64); ok && int(moneyVal) == orderCtx.Money {
				valid = true
				break
			}
		}
		if !valid {
			return &OrderProcessingError{
				Code:    7313,
				Message: fmt.Sprintf("金额%d不在范围内,可用:%v", orderCtx.Money, moneys),
			}
		}
	}

	// 通道浮动加价
	var floatMin, floatMax float64
	if floatMinVal, ok := channel["float_min_money"]; ok {
		switch v := floatMinVal.(type) {
		case float64:
			floatMin = v
		case int:
			floatMin = float64(v)
		case int64:
			floatMin = float64(v)
		}
	}
	if floatMaxVal, ok := channel["float_max_money"]; ok {
		switch v := floatMaxVal.(type) {
		case float64:
			floatMax = v
		case int:
			floatMax = float64(v)
		case int64:
			floatMax = float64(v)
		}
	}
	if !(floatMin == 0 && floatMax == 0) {
		// 随机浮动增加金额
		delta := int(floatMin)
		if floatMax > floatMin {
			delta = int(floatMin) + rand.Intn(int(floatMax-floatMin+1))
		}
		orderCtx.Money += delta
	}

	// 检查通道金额不能为0
	if orderCtx.Money == 0 {
		return &OrderProcessingError{
			Code:    7312,
			Message: "金额不能为0",
		}
	}

	// 检查单笔金额大小
	var minMoney, maxMoney float64
	if minMoneyVal, ok := channel["min_money"]; ok {
		switch v := minMoneyVal.(type) {
		case float64:
			minMoney = v
		case int:
			minMoney = float64(v)
		case int64:
			minMoney = float64(v)
		}
	}
	if maxMoneyVal, ok := channel["max_money"]; ok {
		switch v := maxMoneyVal.(type) {
		case float64:
			maxMoney = v
		case int:
			maxMoney = float64(v)
		case int64:
			maxMoney = float64(v)
		}
	}
	if !(minMoney == 0 && maxMoney == 0) {
		if float64(orderCtx.Money) < minMoney || float64(orderCtx.Money) > maxMoney {
			return &OrderProcessingError{
				Code:    7313,
				Message: fmt.Sprintf("金额%d不在范围[%d,%d]内", orderCtx.Money, int(minMoney), int(maxMoney)),
			}
		}
	}

	return nil
}

// CheckPlugin 检查插件
func (v *Helper) CheckPlugin(ctx context.Context, orderCtx *OrderCreateCtx) *OrderProcessingError {
	// 检查插件
	plugin, err := infra.GetPluginByID(ctx, v.service.redis, int64(orderCtx.PluginID))
	if err != nil || plugin == nil {
		return &OrderProcessingError{
			Code:    7316,
			Message: "该通道不可用1",
		}
	}

	// 检查 status 字段
	status := false
	if statusVal, ok := plugin["status"]; ok {
		if b, ok := statusVal.(bool); ok {
			status = b
		} else if f, ok := statusVal.(float64); ok {
			status = f != 0
		}
	}
	if !status {
		return &OrderProcessingError{
			Code:    7316,
			Message: "该通道不可用2",
		}
	}

	orderCtx.Plugin = plugin

	// 获取插件上游信息
	pluginConfigs, err := infra.GetPluginConfigByID(ctx, v.service.redis, int64(orderCtx.PluginID))
	if err != nil || len(pluginConfigs) == 0 {
		return &OrderProcessingError{
			Code:    7316,
			Message: "该通道不可用",
		}
	}

	// 寻找 key = type 的配置
	for _, config := range pluginConfigs {
		keyVal, _ := config["key"].(string)
		if keyVal == "type" {
			// 解析 value 字段（可能是 JSON 字符串或直接的值）
			valueVal, ok := config["value"]
			if !ok {
				continue
			}

			var typeValue interface{}
			if valueStr, ok := valueVal.(string); ok {
				// 尝试解析为 JSON
				if err := json.Unmarshal([]byte(valueStr), &typeValue); err != nil {
					// 如果不是 JSON，直接使用字符串值
					typeValue = valueStr
				}
			} else {
				typeValue = valueVal
			}

			switch v := typeValue.(type) {
			case float64:
				orderCtx.PluginUpstream = int(v)
			case string:
				if intValue, err := strconv.Atoi(v); err == nil {
					orderCtx.PluginUpstream = intValue
				}
			case int:
				orderCtx.PluginUpstream = v
			case int64:
				orderCtx.PluginUpstream = int(v)
			}
			break
		}
	}

	payTypes, err := infra.GetPluginPayTypeByID(ctx, v.service.redis, int64(orderCtx.PluginID))
	if err != nil || len(payTypes) == 0 {
		return &OrderProcessingError{
			Code:    7317,
			Message: "该通道不可用",
		}
	}

	// 从第一个 payType 中提取 paytype_id
	var payTypeID int64
	if len(payTypes) > 0 {
		if paytypeIDVal, ok := payTypes[0]["paytype_id"]; ok {
			switch v := paytypeIDVal.(type) {
			case float64:
				payTypeID = int64(v)
			case int64:
				payTypeID = v
			case int:
				payTypeID = int64(v)
			}
		}
	}

	payType, err := infra.GetPayTypeByID(ctx, v.service.redis, payTypeID)
	if err != nil || payType == nil {
		return &OrderProcessingError{
			Code:    7317,
			Message: "该通道不可用",
		}
	}

	// 检查 payType 的 status
	payTypeStatus := false
	if statusVal, ok := payType["status"]; ok {
		if b, ok := statusVal.(bool); ok {
			payTypeStatus = b
		} else if f, ok := statusVal.(float64); ok {
			payTypeStatus = f != 0
		}
	}
	if !payTypeStatus {
		return &OrderProcessingError{
			Code:    7317,
			Message: "该通道不可用",
		}
	}

	// 保持向后兼容，设置到 orderCtx.PayType
	orderCtx.PayType = payType
	if keyVal, ok := payType["key"]; ok {
		if keyStr, ok := keyVal.(string); ok {
			orderCtx.PluginType = keyStr
		}
	}
	return nil
}

// getAllDomains 获取所有域名（返回 map）
func (v *Helper) getAllDomains(ctx context.Context) ([]map[string]interface{}, error) {
	var allDomains []map[string]interface{}
	var cursor uint64
	var keys []string
	var err error

	// 使用 SCAN 迭代扫描所有 domain:* 的 key，避免 KEYS 命令阻塞 Redis
	for {
		keys, cursor, err = v.service.redis.Scan(ctx, cursor, "domain:*", 100).Result()
		if err != nil {
			return nil, err
		}

		// 批量获取这些 key 的值
		for _, key := range keys {
			val, err := v.service.redis.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			// 解析为数组
			var domains []map[string]interface{}
			if err := json.Unmarshal([]byte(val), &domains); err == nil {
				allDomains = append(allDomains, domains...)
			} else {
				// 尝试解析为单个对象
				var domainMap map[string]interface{}
				if err := json.Unmarshal([]byte(val), &domainMap); err == nil {
					allDomains = append(allDomains, domainMap)
				}
			}
		}

		// cursor 为 0 表示扫描完成
		if cursor == 0 {
			break
		}
	}

	return allDomains, nil
}

// getPluginPayDomain 获取插件单独的域名设置,如果不设置,则返回空字符串
func (v *Helper) getPluginPayDomain(ctx context.Context, pluginID int, channelID int) string {
	// 获取插件配置
	pluginConfigs, err := infra.GetConfigsByID(ctx, v.service.redis, "plugin_config", int64(pluginID))
	if err != nil || len(pluginConfigs) == 0 {
		return ""
	}

	// 查找 key = "pay_domain" 的配置
	for _, config := range pluginConfigs {
		key, _ := config["key"].(string)
		if key != "pay_domain" {
			continue
		}

		value, _ := config["value"].(string)
		if value == "" {
			return ""
		}

		// 尝试解析为 JSON
		var domains map[string]interface{}
		if err := json.Unmarshal([]byte(value), &domains); err == nil {
			// 成功解析为 JSON，尝试根据 channel_id 查找
			// 先尝试字符串 key (JSON 对象的 key 总是字符串)
			channelIDStr := strconv.Itoa(channelID)
			if domainURL, ok := domains[channelIDStr].(string); ok && domainURL != "" {
				return domainURL
			}
			// 最后尝试 "else" key
			if domainURL, ok := domains["else"].(string); ok && domainURL != "" {
				return domainURL
			}
		} else {
			// 解析失败，直接返回原始值
			return value
		}
	}

	return ""
}

// findDomainByURL 根据 URL 查找域名对象（返回 map）
func (v *Helper) findDomainByURL(ctx context.Context, url string) map[string]interface{} {
	domains, err := v.getAllDomains(ctx)
	if err != nil {
		return nil
	}

	for _, domain := range domains {
		if urlVal, ok := domain["url"].(string); ok && urlVal == url {
			return domain
		}
	}

	return nil
}

// CheckDomain 检查域名
func (v *Helper) CheckDomain(ctx context.Context, orderCtx *OrderCreateCtx) *OrderProcessingError {
	// 获取插件单独的域名设置
	domainURL := v.getPluginPayDomain(ctx, orderCtx.PluginID, orderCtx.ChannelID)

	// 插件是否自己设置了域名
	if domainURL != "" {
		// 尝试查找对应的 PayDomain 对象
		domain := v.findDomainByURL(ctx, domainURL)
		if domain != nil {
			orderCtx.Domain = domain
			// 提取 id
			if idVal, ok := domain["id"]; ok {
				switch v := idVal.(type) {
				case float64:
					orderCtx.DomainID = int(v)
				case int64:
					orderCtx.DomainID = int(v)
				case int:
					orderCtx.DomainID = v
				}
			}
			// 提取 url
			if urlVal, ok := domain["url"].(string); ok {
				orderCtx.DomainURL = urlVal
			}
			return nil
		}
		// 如果没找到 PayDomain 对象，直接设置 URL
		orderCtx.DomainURL = domainURL
		return nil
	}

	// 如果没有插件域名配置，从支付域名列表中随机抽取
	allDomains, err := v.getAllDomains(ctx)
	if err != nil {
		return &OrderProcessingError{
			Code:    7314,
			Message: "无可用收银台",
		}
	}

	// 根据 plugin_upstream 过滤域名
	var availableDomains []map[string]interface{}
	for _, domain := range allDomains {
		// 检查 status 字段
		status := false
		if statusVal, ok := domain["status"]; ok {
			if b, ok := statusVal.(bool); ok {
				status = b
			} else if f, ok := statusVal.(float64); ok {
				status = f != 0
			}
		}
		if !status {
			continue
		}

		if orderCtx.PluginUpstream == 6 {
			// 判断是否支持微信
			wechatStatus := false
			if wechatStatusVal, ok := domain["wechat_status"]; ok {
				if b, ok := wechatStatusVal.(bool); ok {
					wechatStatus = b
				} else if f, ok := wechatStatusVal.(float64); ok {
					wechatStatus = f != 0
				}
			}
			if !wechatStatus {
				continue
			}
		} else if orderCtx.PluginUpstream == 5 {
			// 判断是否支持支付宝
			payStatus := false
			if payStatusVal, ok := domain["pay_status"]; ok {
				if b, ok := payStatusVal.(bool); ok {
					payStatus = b
				} else if f, ok := payStatusVal.(float64); ok {
					payStatus = f != 0
				}
			}
			if !payStatus {
				continue
			}
		}

		availableDomains = append(availableDomains, domain)
	}

	if len(availableDomains) == 0 {
		return &OrderProcessingError{
			Code:    7314,
			Message: "无可用收银台",
		}
	}

	// 随机选择一个域名
	selectedDomain := availableDomains[rand.Intn(len(availableDomains))]
	orderCtx.Domain = selectedDomain
	// 提取 id
	if idVal, ok := selectedDomain["id"]; ok {
		switch v := idVal.(type) {
		case float64:
			orderCtx.DomainID = int(v)
		case int64:
			orderCtx.DomainID = int(v)
		case int:
			orderCtx.DomainID = v
		}
	}
	// 提取 url
	if urlVal, ok := selectedDomain["url"].(string); ok {
		orderCtx.DomainURL = urlVal
	}

	return nil
}

// CheckPluginResponder 检查插件响应器
func (v *Helper) CheckPluginResponder(ctx context.Context, orderCtx *OrderCreateCtx) *OrderProcessingError {
	// 通过插件管理器获取插件实例
	pluginInstance, err := v.service.pluginMgr.GetPluginByCtx(ctx, orderCtx)
	if err != nil {
		return &OrderProcessingError{
			Code:    7318,
			Message: "插件响应器不可用",
		}
	}

	// 将插件实例存储到上下文中
	orderCtx.Responder = pluginInstance
	return nil
}

func (s *Helper) GetWriteoffIDs(ctx context.Context, tenantID int, money int, channelID int) []int {
	// TODO: 实现获取核销ID列表
	return []int{}
}

func (s *Helper) GetExtraWriteoffProduct(ctx context.Context, orderCtx *OrderCreateCtx, writeoffIDs []int) (*int, *int, int) {
	// TODO: 实现获取核销产品
	return nil, nil, 0
}

func (s *Helper) GetExtraTenantCookie(ctx context.Context, pluginID int, tenantID int) *int {
	// TODO: 实现获取租户cookie
	return nil
}

func (s *Helper) WaitProduct(ctx context.Context, orderCtx *OrderCreateCtx) {
	// TODO: 实现等待产品
}

func (s *Helper) GetPluginOutTime(ctx context.Context, pluginID int) int {
	// 创建临时上下文获取插件
	orderCtx := &OrderCreateCtx{PluginID: pluginID}
	pluginInstance, err := s.service.pluginMgr.GetPluginByCtx(ctx, orderCtx)
	if err != nil {
		return 300 // 默认5分钟
	}

	// 使用类型断言访问能力接口
	if caps, ok := pluginInstance.(plugin.PluginCapabilities); ok {
		return caps.GetTimeout(ctx, pluginID)
	}
	return 300 // 默认5分钟
}

func (s *Helper) TryCreateOrder(ctx context.Context, orderCtx *OrderCreateCtx) *OrderProcessingError {
	// TODO: 实现创建订单
	orderCtx.OrderNo = uuid.NewString()
	return nil
}

func (s *Helper) TryCreateOrderDetail(ctx context.Context, orderCtx *OrderCreateCtx) *OrderProcessingError {
	// TODO: 实现创建订单详情
	return nil
}

func (s *Helper) CheckTenantBalance(ctx context.Context, orderCtx *OrderCreateCtx) bool {
	// TODO: 实现检查租户余额
	return true
}

func (s *Helper) TakeUpWriteoffTax(ctx context.Context, writeoffID int, money int) bool {
	// TODO: 实现占用码商税费
	return true
}

func (s *Helper) UpdateOrCreateOrderLog(ctx context.Context, outOrderNo string, signRaw string, sign string) {
	// TODO: 实现更新或创建订单日志
}

func (s *Helper) GetExtraPayURL(ctx context.Context, orderCtx *OrderCreateCtx) map[string]interface{} {
	// TODO: 如果需要extra支付URL，可以通过CreateOrder方法获取
	// 或者通过响应中的ExtraData获取
	return nil
}

func (s *Helper) FailOrder(ctx context.Context, orderNo string) {
	// TODO: 实现订单失败处理
}

func (s *Helper) AddPluginQueryOrderJob(ctx context.Context, orderNo string, delay time.Duration) {
	// TODO: 实现添加插件查询订单任务
}

func (s *Helper) GetAuthURL(url string, domain interface{}, orderCtx *OrderCreateCtx) string {
	// 授权URL的生成现在由插件在CreateOrder内部处理
	// 如果需要单独的授权URL，可以通过CreateOrder的响应获取
	return url
}

func (s *Helper) AddTimeoutCheckJob(ctx context.Context, orderNo string, tenantID int, tax int, seconds int) {
	// TODO: 实现添加超时检查任务
}

func (s *Helper) AddNotifyOrderSubmitJob(ctx context.Context, orderCtx *OrderCreateCtx, delay time.Duration) {
	// TODO: 实现添加通知订单提交任务
}
