package alipay

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DayStatisticsService 日统计服务（放在 plugin 包中避免循环依赖）
type DayStatisticsService struct{}

// NewDayStatisticsService 创建日统计服务
func NewDayStatisticsService() *DayStatisticsService {
	return &DayStatisticsService{}
}

// SubmitBaseDayStatistics 提交日统计（下单时）
// 参考 Python: submit_base_day_statistics
// 根据业务模式选择不同的统计表：
// - 公池模式 (extra_arg == 1): AlipayPublicPoolDayStatistics
// - 神码模式: AlipayShenmaDayStatistics
// - 普通模式: AlipayProductDayStatistics
func (s *DayStatisticsService) SubmitBaseDayStatistics(ctx context.Context, productID string, createDatetime time.Time, channelID int64, tenantID int64, extraArg *int) error {
	// 解析产品ID
	productIDInt, err := parseProductIDInt(productID)
	if err != nil {
		return fmt.Errorf("解析产品ID失败: %w", err)
	}

	// 获取日期（只取日期部分，忽略时间）
	date := time.Date(createDatetime.Year(), createDatetime.Month(), createDatetime.Day(), 0, 0, 0, 0, createDatetime.Location())

	// 判断业务模式
	if extraArg != nil && *extraArg == 1 {
		// 公池模式
		return s.submitPublicPoolDayStatistics(ctx, productIDInt, date, channelID)
	}

	// 检查是否是神码模式
	// 神码模式：tenant_id != product.writeoff.parent_id
	var product models.AlipayProduct
	if err := database.DB.Where("id = ?", productIDInt).First(&product).Error; err != nil {
		return fmt.Errorf("查询产品失败: %w", err)
	}

	// 获取核销信息
	var writeoff models.Writeoff
	if err := database.DB.Where("id = ?", product.WriteoffID).First(&writeoff).Error; err != nil {
		return fmt.Errorf("查询核销失败: %w", err)
	}

	// 判断是否是神码模式
	// 注意：writeoff.ParentID 是 int64 类型，不是指针
	if writeoff.ParentID > 0 && tenantID != writeoff.ParentID {
		// 神码模式：查找 AlipayShenma
		var shenma models.AlipayShenma
		if err := database.DB.Where("alipay_id = ? AND tenant_id = ?", productIDInt, tenantID).First(&shenma).Error; err == nil {
			return s.submitShenmaDayStatistics(ctx, shenma.ID, date, channelID)
		}
		// 如果找不到神码记录，降级为普通模式
	}

	// 普通模式
	return s.submitProductDayStatistics(ctx, productIDInt, date, channelID)
}

// submitProductDayStatistics 更新产品日统计（普通模式）
// 使用 INSERT ... ON DUPLICATE KEY UPDATE 避免并发竞态条件
func (s *DayStatisticsService) submitProductDayStatistics(ctx context.Context, productID int64, date time.Time, channelID int64) error {
	stats := models.AlipayProductDay{
		ProductID:    &productID,
		PayChannelID: &channelID,
		Date:         date,
		SubmitCount:  1,
		Ver:          1,
	}

	// 使用 ON DUPLICATE KEY UPDATE 实现原子 UPSERT
	// 如果记录已存在（唯一索引冲突），则更新计数；否则创建新记录
	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
			{Name: "product_id"},
			{Name: "pay_channel_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"submit_count": gorm.Expr("submit_count + 1"),
			"ver":          gorm.Expr("ver + 1"),
		}),
	}).Create(&stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新产品日统计失败: %w", err)
	}

	return nil
}

// submitPublicPoolDayStatistics 更新公池日统计（公池模式）
// 使用 INSERT ... ON DUPLICATE KEY UPDATE 避免并发竞态条件
func (s *DayStatisticsService) submitPublicPoolDayStatistics(ctx context.Context, productID int64, date time.Time, channelID int64) error {
	// 查找公池记录
	var pool models.AlipayPublicPool
	if err := database.DB.Where("alipay_id = ?", productID).First(&pool).Error; err != nil {
		// 如果找不到公池记录，记录警告但不返回错误（容错处理）
		logger.Logger.Warn("找不到公池记录，跳过公池日统计",
			zap.Int64("product_id", productID),
			zap.Int64("channel_id", channelID),
			zap.Error(err))
		return nil
	}

	stats := models.AlipayPublicPoolDay{
		PoolID:       &pool.ID,
		PayChannelID: &channelID,
		Date:         date,
		SubmitCount:  1,
		Ver:          1,
	}

	// 使用 ON DUPLICATE KEY UPDATE 实现原子 UPSERT
	// 如果记录已存在（唯一索引冲突），则更新计数；否则创建新记录
	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
			{Name: "pool_id"},
			{Name: "pay_channel_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"submit_count": gorm.Expr("submit_count + 1"),
			"ver":          gorm.Expr("ver + 1"),
		}),
	}).Create(&stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新公池日统计失败: %w", err)
	}

	return nil
}

// submitShenmaDayStatistics 更新神码日统计（神码模式）
// 使用 INSERT ... ON DUPLICATE KEY UPDATE 避免并发竞态条件
func (s *DayStatisticsService) submitShenmaDayStatistics(ctx context.Context, shenmaID int64, date time.Time, channelID int64) error {
	stats := models.AlipayShenmaDay{
		ShenmaID:     &shenmaID,
		PayChannelID: &channelID,
		Date:         date,
		SubmitCount:  1,
		Ver:          1,
	}

	// 使用 ON DUPLICATE KEY UPDATE 实现原子 UPSERT
	// 如果记录已存在（唯一索引冲突），则更新计数；否则创建新记录
	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
			{Name: "shenma_id"},
			{Name: "pay_channel_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"submit_count": gorm.Expr("submit_count + 1"),
			"ver":          gorm.Expr("ver + 1"),
		}),
	}).Create(&stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新神码日统计失败: %w", err)
	}

	return nil
}

// SuccessBaseDayStatistics 成功日统计（支付成功时）
// 参考 Python: success_base_day_statistics
// 根据业务模式选择不同的统计表并更新成功金额
func (s *DayStatisticsService) SuccessBaseDayStatistics(ctx context.Context, productID string, createDatetime time.Time, channelID int64, tenantID int64, extraArg *int, successMoney int64) error {
	// 解析产品ID
	productIDInt, err := parseProductIDInt(productID)
	if err != nil {
		return fmt.Errorf("解析产品ID失败: %w", err)
	}

	// 获取日期（只取日期部分，忽略时间）
	date := time.Date(createDatetime.Year(), createDatetime.Month(), createDatetime.Day(), 0, 0, 0, 0, createDatetime.Location())

	// 判断业务模式
	if extraArg != nil && *extraArg == 1 {
		// 公池模式
		return s.successPublicPoolDayStatistics(ctx, productIDInt, date, channelID, successMoney)
	}

	// 检查是否是神码模式
	// 神码模式：tenant_id != product.writeoff.parent_id
	var product models.AlipayProduct
	if err := database.DB.Where("id = ?", productIDInt).First(&product).Error; err != nil {
		return fmt.Errorf("查询产品失败: %w", err)
	}

	// 获取核销信息
	var writeoff models.Writeoff
	if err := database.DB.Where("id = ?", product.WriteoffID).First(&writeoff).Error; err != nil {
		return fmt.Errorf("查询核销失败: %w", err)
	}

	// 判断是否是神码模式
	// 注意：writeoff.ParentID 是 int64 类型，不是指针
	if writeoff.ParentID > 0 && tenantID != writeoff.ParentID {
		// 神码模式：查找 AlipayShenma
		var shenma models.AlipayShenma
		if err := database.DB.Where("alipay_id = ? AND tenant_id = ?", productIDInt, tenantID).First(&shenma).Error; err == nil {
			return s.successShenmaDayStatistics(ctx, shenma.ID, date, channelID, successMoney)
		}
		// 如果找不到神码记录，降级为普通模式
	}

	// 普通模式
	return s.successProductDayStatistics(ctx, productIDInt, date, channelID, successMoney)
}

// successProductDayStatistics 更新产品日统计（普通模式，成功）
func (s *DayStatisticsService) successProductDayStatistics(ctx context.Context, productID int64, date time.Time, channelID int64, successMoney int64) error {
	stats := models.AlipayProductDay{
		ProductID:    &productID,
		PayChannelID: &channelID,
		Date:         date,
		SuccessCount: 1,
		SuccessMoney: successMoney,
		Ver:          1,
	}

	// 使用 ON DUPLICATE KEY UPDATE 实现原子 UPSERT
	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
			{Name: "product_id"},
			{Name: "pay_channel_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"success_count": gorm.Expr("success_count + 1"),
			"success_money": gorm.Expr("success_money + ?", successMoney),
			"ver":           gorm.Expr("ver + 1"),
		}),
	}).Create(&stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新产品日统计失败: %w", err)
	}

	return nil
}

// successPublicPoolDayStatistics 更新公池日统计（公池模式，成功）
func (s *DayStatisticsService) successPublicPoolDayStatistics(ctx context.Context, productID int64, date time.Time, channelID int64, successMoney int64) error {
	// 查找公池记录
	var pool models.AlipayPublicPool
	if err := database.DB.Where("alipay_id = ?", productID).First(&pool).Error; err != nil {
		logger.Logger.Warn("找不到公池记录，跳过公池日统计",
			zap.Int64("product_id", productID),
			zap.Int64("channel_id", channelID),
			zap.Error(err))
		return nil
	}

	stats := models.AlipayPublicPoolDay{
		PoolID:       &pool.ID,
		PayChannelID: &channelID,
		Date:         date,
		SuccessCount: 1,
		SuccessMoney: successMoney,
		Ver:          1,
	}

	// 使用 ON DUPLICATE KEY UPDATE 实现原子 UPSERT
	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
			{Name: "pool_id"},
			{Name: "pay_channel_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"success_count": gorm.Expr("success_count + 1"),
			"success_money": gorm.Expr("success_money + ?", successMoney),
			"ver":           gorm.Expr("ver + 1"),
		}),
	}).Create(&stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新公池日统计失败: %w", err)
	}

	return nil
}

// successShenmaDayStatistics 更新神码日统计（神码模式，成功）
func (s *DayStatisticsService) successShenmaDayStatistics(ctx context.Context, shenmaID int64, date time.Time, channelID int64, successMoney int64) error {
	stats := models.AlipayShenmaDay{
		ShenmaID:     &shenmaID,
		PayChannelID: &channelID,
		Date:         date,
		SuccessCount: 1,
		SuccessMoney: successMoney,
		Ver:          1,
	}

	// 使用 ON DUPLICATE KEY UPDATE 实现原子 UPSERT
	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
			{Name: "shenma_id"},
			{Name: "pay_channel_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"success_count": gorm.Expr("success_count + 1"),
			"success_money": gorm.Expr("success_money + ?", successMoney),
			"ver":           gorm.Expr("ver + 1"),
		}),
	}).Create(&stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新神码日统计失败: %w", err)
	}

	return nil
}
