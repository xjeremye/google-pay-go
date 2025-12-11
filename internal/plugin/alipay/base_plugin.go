package alipay

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/plugin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BasePlugin 支付宝基础插件
// 继承自 plugin.BasePlugin，提供支付宝插件通用的功能
// 所有支付宝相关的插件（PhonePlugin、WapPlugin 等）应该继承或嵌入此插件
type BasePlugin struct {
	*plugin.BasePlugin // 嵌入 plugin.BasePlugin，继承通用功能
}

// NewBasePlugin 创建支付宝基础插件
func NewBasePlugin(pluginID int64) *BasePlugin {
	return &BasePlugin{
		BasePlugin: plugin.NewBasePlugin(pluginID),
	}
}

// WaitProduct 等待产品（支付宝通用实现）
// 参考 Python: BasePluginResponder.wait_product
// 通用实现：获取支付宝产品（适用于所有支付宝插件）
// 如果插件需要自定义逻辑，可以覆盖此方法
func (p *BasePlugin) WaitProduct(ctx context.Context, req *plugin.WaitProductRequest) (*plugin.WaitProductResponse, error) {
	// 获取可用的核销ID列表
	writeoffIDs, err := plugin.GetWriteoffIDsForPlugin(req.TenantID, req.Money, &req.ChannelID)
	if err != nil {
		if logger.Logger != nil {
			logger.Logger.Error("获取核销ID失败",
				zap.Int64("tenant_id", req.TenantID),
				zap.Int64("channel_id", req.ChannelID),
				zap.Int("money", req.Money),
				zap.Error(err))
		}
		return plugin.NewWaitProductErrorResponse(7318, fmt.Sprintf("获取核销ID失败: %v", err)), nil
	}
	if len(writeoffIDs) == 0 {
		if logger.Logger != nil {
			logger.Logger.Warn("没有可选核销",
				zap.Int64("tenant_id", req.TenantID),
				zap.Int64("channel_id", req.ChannelID),
				zap.Int("money", req.Money))
		}
		return plugin.NewWaitProductErrorResponse(7318, "没有可选核销"), nil
	}

	if logger.Logger != nil {
		logger.Logger.Debug("获取到可用核销ID",
			zap.Int64("tenant_id", req.TenantID),
			zap.Int64("channel_id", req.ChannelID),
			zap.Int("money", req.Money),
			zap.Int64s("writeoff_ids", writeoffIDs))
	}

	// 获取产品（通用实现：支付宝产品）
	productID, writeoffID, money, err := getAlipayProduct(ctx, req, writeoffIDs)
	if err != nil {
		if logger.Logger != nil {
			logger.Logger.Error("获取产品失败",
				zap.Int64("tenant_id", req.TenantID),
				zap.Int64("channel_id", req.ChannelID),
				zap.Int("money", req.Money),
				zap.Int64s("writeoff_ids", writeoffIDs),
				zap.Error(err))
		}
		return plugin.NewWaitProductErrorResponse(7318, fmt.Sprintf("获取产品失败: %v", err)), nil
	}
	if productID == "" {
		if logger.Logger != nil {
			logger.Logger.Warn("无货物库存",
				zap.Int64("tenant_id", req.TenantID),
				zap.Int64("channel_id", req.ChannelID),
				zap.Int("money", req.Money),
				zap.Int64s("writeoff_ids", writeoffIDs),
				zap.String("reason", "查询到0个产品或所有产品都不符合条件"))
		}
		return plugin.NewWaitProductErrorResponse(7318, "无货物库存"), nil
	}
	if writeoffID == nil {
		if logger.Logger != nil {
			logger.Logger.Warn("无核销库存",
				zap.Int64("tenant_id", req.TenantID),
				zap.Int64("channel_id", req.ChannelID),
				zap.String("product_id", productID))
		}
		return plugin.NewWaitProductErrorResponse(7318, "无核销库存"), nil
	}
	return plugin.NewWaitProductSuccessResponse(productID, writeoffID, "", money), nil
}

// CallbackSubmit 下单回调（订单创建成功后调用）
// 参考 Python: BasePluginResponder.callback_submit
// 支付宝通用实现：更新订单备注和日统计
func (p *BasePlugin) CallbackSubmit(ctx context.Context, req *plugin.CallbackSubmitRequest) error {
	// 1. 更新订单备注（产品名称）
	if req.ProductID != "" {
		productIDInt, err := parseProductIDInt(req.ProductID)
		if err == nil {
			var product models.AlipayProduct
			if err := database.DB.Select("name").Where("id = ?", productIDInt).First(&product).Error; err == nil {
				// 更新订单备注
				if err := database.DB.Model(&models.Order{}).
					Where("order_no = ?", req.OrderNo).
					Update("remarks", product.Name).Error; err != nil {
					logger.Logger.Warn("更新订单备注失败",
						zap.String("order_no", req.OrderNo),
						zap.Error(err))
				}
			}
		}
	}

	// 2. 更新日统计
	// 解析创建时间
	createDatetime, err := time.Parse("2006-01-02 15:04:05", req.CreateDatetime)
	if err != nil {
		// 如果解析失败，尝试其他格式
		createDatetime, err = time.Parse("2006-01-02T15:04:05Z07:00", req.CreateDatetime)
		if err != nil {
			// 如果还是失败，使用当前时间
			createDatetime = time.Now()
			logger.Logger.Warn("解析订单创建时间失败，使用当前时间",
				zap.String("order_no", req.OrderNo),
				zap.String("create_datetime", req.CreateDatetime),
				zap.Error(err))
		}
	}

	// 获取通道的 extra_arg
	var extraArg *int
	if req.ChannelID > 0 {
		var channel models.PayChannel
		if err := database.DB.Select("extra_arg").Where("id = ?", req.ChannelID).First(&channel).Error; err == nil {
			extraArg = channel.ExtraArg
		}
	}

	// 调用日统计服务（更新产品统计）
	dayStatsService := NewDayStatisticsService()
	if err := dayStatsService.SubmitBaseDayStatistics(ctx, req.ProductID, createDatetime, req.ChannelID, req.TenantID, extraArg); err != nil {
		logger.Logger.Error("更新产品日统计失败",
			zap.String("order_no", req.OrderNo),
			zap.String("product_id", req.ProductID),
			zap.Int64("channel_id", req.ChannelID),
			zap.Error(err))
		// 不返回错误，避免影响主流程
	}

	// 3. 更新全局日统计（submit_count 和 submit_money）
	// 参考 Python: 全局统计也需要更新提交统计
	date := time.Date(createDatetime.Year(), createDatetime.Month(), createDatetime.Day(), 0, 0, 0, 0, createDatetime.Location())
	globalStats := models.DayStatistics{
		Date:        date,
		SubmitCount: 1,
		SubmitMoney: int64(req.Money),
		Ver:         1,
	}

	err = database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"submit_count": gorm.Expr("submit_count + 1"),
			"submit_money": gorm.Expr("submit_money + ?", req.Money),
			"ver":          gorm.Expr("ver + 1"),
		}),
	}).Create(&globalStats).Error

	if err != nil {
		logger.Logger.Error("更新全局日统计失败",
			zap.String("order_no", req.OrderNo),
			zap.Error(err))
		// 不返回错误，避免影响主流程
	}

	return nil
}

// CallbackSuccess 支付成功回调
// 参考 Python: callback_success
// 主要功能：
// 1. 更新产品成功统计（success_count, success_money）
// 2. 处理分账（根据collection_type）
func (p *BasePlugin) CallbackSuccess(ctx context.Context, req *plugin.CallbackSuccessRequest) error {
	// 1. 更新产品成功统计
	// 解析创建时间
	createDatetime, err := time.Parse("2006-01-02 15:04:05", req.CreateDatetime)
	if err != nil {
		createDatetime = time.Now()
		logger.Logger.Warn("解析订单创建时间失败，使用当前时间",
			zap.String("order_no", req.OrderNo),
			zap.String("create_datetime", req.CreateDatetime),
			zap.Error(err))
	}

	// 获取通道的 extra_arg
	var extraArg *int
	if req.ChannelID > 0 {
		var channel models.PayChannel
		if err := database.DB.Select("extra_arg").Where("id = ?", req.ChannelID).First(&channel).Error; err == nil {
			extraArg = channel.ExtraArg
		}
	}

	// 调用日统计服务更新成功统计
	dayStatsService := NewDayStatisticsService()
	if err := dayStatsService.SuccessBaseDayStatistics(ctx, req.ProductID, createDatetime, req.ChannelID, req.TenantID, extraArg, int64(req.NotifyMoney)); err != nil {
		logger.Logger.Error("更新产品成功统计失败",
			zap.String("order_no", req.OrderNo),
			zap.String("product_id", req.ProductID),
			zap.Int64("channel_id", req.ChannelID),
			zap.Error(err))
		// 不返回错误，避免影响主流程
	}

	// 2. 处理分账（根据collection_type）
	// 参考 Python: callback_success 中的分账处理逻辑
	// 注意：分账处理逻辑比较复杂，这里先记录日志，后续可以单独实现
	logger.Logger.Info("订单支付成功，需要处理分账",
		zap.String("order_no", req.OrderNo),
		zap.String("product_id", req.ProductID),
		zap.Int64("notify_money", int64(req.NotifyMoney)))
	// TODO: 实现分账处理逻辑（根据collection_type判断分账模式）

	return nil
}

// parseProductIDInt 解析产品ID（字符串转 int64）
func parseProductIDInt(productID string) (int64, error) {
	if productID == "" {
		return 0, fmt.Errorf("产品ID不能为空")
	}
	// 尝试解析为 int64
	var id int64
	_, err := fmt.Sscanf(productID, "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("产品ID格式错误: %w", err)
	}
	return id, nil
}
