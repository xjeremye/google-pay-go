package service

import (
	"context"
	"fmt"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/plugin"
)

// waitProduct 等待产品（获取产品ID、核销ID、CookieID等）
// 参考 Python: BasePluginResponder.wait_product
func (s *OrderService) waitProduct(ctx context.Context, orderCtx *OrderCreateContext) *OrderError {
	// 获取插件实例
	pluginInstance, err := s.pluginManager.GetPluginByCtx(ctx, orderCtx)
	if err != nil {
		return NewOrderError(ErrCodePluginUnavailable, fmt.Sprintf("插件不可用: %v", err))
	}

	// 构建等待产品请求
	waitReq := &plugin.WaitProductRequest{
		OutOrderNo:     orderCtx.OutOrderNo,
		Money:          orderCtx.Money,
		NotifyMoney:    orderCtx.NotifyMoney,
		MerchantID:     orderCtx.MerchantID,
		TenantID:       orderCtx.TenantID,
		ChannelID:      orderCtx.ChannelID,
		PluginID:       orderCtx.PluginID,
		PluginType:     orderCtx.PluginType,
		PluginUpstream: orderCtx.PluginUpstream,
	}

	// 添加关联对象（转换为 map）
	if orderCtx.Channel != nil {
		channelMap := map[string]interface{}{
			"id":        orderCtx.Channel.ID,
			"name":      orderCtx.Channel.Name,
			"status":    orderCtx.Channel.Status,
			"plugin_id": orderCtx.Channel.PluginID,
		}
		waitReq.Channel = channelMap
	}

	// 调用插件等待产品
	waitResp, err := pluginInstance.WaitProduct(ctx, waitReq)
	if err != nil {
		return NewOrderError(ErrCodeCreateFailed, fmt.Sprintf("等待产品失败: %v", err))
	}

	// 检查响应
	if !waitResp.Success {
		return NewOrderError(waitResp.ErrorCode, waitResp.ErrorMessage)
	}

	// 更新上下文
	orderCtx.ProductID = waitResp.ProductID
	orderCtx.WriteoffID = waitResp.WriteoffID
	orderCtx.CookieID = waitResp.CookieID
	orderCtx.Money = waitResp.Money // 金额可能被调整

	return nil
}

// getWriteoffIDs 获取可用的核销ID列表
// 参考 Python: get_writeoff_ids(tenant_id, money, pay_channel_id=None)
func getWriteoffIDs(tenantID int64, money int, payChannelID *int64) ([]int64, error) {
	// 构建查询条件
	query := database.DB.Model(&models.Writeoff{}).
		Joins("JOIN dvadmin_system_users ON dvadmin_writeoff.system_user_id = dvadmin_system_users.id").
		Where("dvadmin_writeoff.parent_id = ?", tenantID).
		Where("dvadmin_system_users.status = ?", true).
		Where("dvadmin_system_users.is_active = ?", true).
		Where("dvadmin_writeoff.balance IS NULL OR dvadmin_writeoff.balance >= ?", money)

	var writeoffIDs []int64
	if err := query.Pluck("dvadmin_writeoff.id", &writeoffIDs).Error; err != nil {
		return nil, fmt.Errorf("查询核销ID失败: %w", err)
	}

	// 过滤掉被禁用的支付通道关联
	if payChannelID != nil {
		var disabledWriteoffIDs []int64
		if err := database.DB.Model(&models.WriteoffPayChannel{}).
			Where("pay_channel_id = ? AND status = ?", *payChannelID, false).
			Pluck("writeoff_id", &disabledWriteoffIDs).Error; err != nil {
			return nil, fmt.Errorf("查询禁用的核销通道关联失败: %w", err)
		}

		// 过滤掉被禁用的核销ID
		enabledWriteoffIDs := make([]int64, 0)
		disabledMap := make(map[int64]bool)
		for _, id := range disabledWriteoffIDs {
			disabledMap[id] = true
		}
		for _, id := range writeoffIDs {
			if !disabledMap[id] {
				enabledWriteoffIDs = append(enabledWriteoffIDs, id)
			}
		}
		writeoffIDs = enabledWriteoffIDs
	}

	return writeoffIDs, nil
}
