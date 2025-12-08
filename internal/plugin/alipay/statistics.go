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
func (s *DayStatisticsService) submitProductDayStatistics(ctx context.Context, productID int64, date time.Time, channelID int64) error {
	// 使用 ON DUPLICATE KEY UPDATE 或先查询后更新
	var stats models.AlipayProductDay
	err := database.DB.Where("date = ? AND product_id = ? AND pay_channel_id = ?", date, productID, channelID).
		First(&stats).Error

	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		stats = models.AlipayProductDay{
			ProductID:    &productID,
			PayChannelID: &channelID,
			Date:         date,
			SubmitCount:  1,
			Ver:          1,
		}
		if err := database.DB.Create(&stats).Error; err != nil {
			return fmt.Errorf("创建产品日统计失败: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("查询产品日统计失败: %w", err)
	}

	// 存在，更新计数（使用原子操作）
	if err := database.DB.Model(&stats).
		Updates(map[string]interface{}{
			"submit_count": gorm.Expr("submit_count + 1"),
			"ver":          gorm.Expr("ver + 1"),
		}).Error; err != nil {
		return fmt.Errorf("更新产品日统计失败: %w", err)
	}

	return nil
}

// submitPublicPoolDayStatistics 更新公池日统计（公池模式）
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

	// 使用 ON DUPLICATE KEY UPDATE 或先查询后更新
	var stats models.AlipayPublicPoolDay
	err := database.DB.Where("date = ? AND pool_id = ? AND pay_channel_id = ?", date, pool.ID, channelID).
		First(&stats).Error

	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		stats = models.AlipayPublicPoolDay{
			PoolID:       &pool.ID,
			PayChannelID: &channelID,
			Date:         date,
			SubmitCount:  1,
			Ver:          1,
		}
		if err := database.DB.Create(&stats).Error; err != nil {
			return fmt.Errorf("创建公池日统计失败: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("查询公池日统计失败: %w", err)
	}

	// 存在，更新计数（使用原子操作）
	if err := database.DB.Model(&stats).
		Updates(map[string]interface{}{
			"submit_count": gorm.Expr("submit_count + 1"),
			"ver":          gorm.Expr("ver + 1"),
		}).Error; err != nil {
		return fmt.Errorf("更新公池日统计失败: %w", err)
	}

	return nil
}

// submitShenmaDayStatistics 更新神码日统计（神码模式）
func (s *DayStatisticsService) submitShenmaDayStatistics(ctx context.Context, shenmaID int64, date time.Time, channelID int64) error {
	// 使用 ON DUPLICATE KEY UPDATE 或先查询后更新
	var stats models.AlipayShenmaDay
	err := database.DB.Where("date = ? AND shenma_id = ? AND pay_channel_id = ?", date, shenmaID, channelID).
		First(&stats).Error

	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		stats = models.AlipayShenmaDay{
			ShenmaID:     &shenmaID,
			PayChannelID: &channelID,
			Date:         date,
			SubmitCount:  1,
			Ver:          1,
		}
		if err := database.DB.Create(&stats).Error; err != nil {
			return fmt.Errorf("创建神码日统计失败: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("查询神码日统计失败: %w", err)
	}

	// 存在，更新计数（使用原子操作）
	if err := database.DB.Model(&stats).
		Updates(map[string]interface{}{
			"submit_count": gorm.Expr("submit_count + 1"),
			"ver":          gorm.Expr("ver + 1"),
		}).Error; err != nil {
		return fmt.Errorf("更新神码日统计失败: %w", err)
	}

	return nil
}
