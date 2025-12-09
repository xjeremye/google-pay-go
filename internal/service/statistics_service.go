package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StatisticsService 统计服务
type StatisticsService struct{}

// NewStatisticsService 创建统计服务
func NewStatisticsService() *StatisticsService {
	return &StatisticsService{}
}

// SuccessBaseDayStatistics 成功日统计（通用方法）
// 参考 Python: success_base_day_statistics
func (s *StatisticsService) SuccessBaseDayStatistics(
	ctx context.Context,
	stats interface{},
	successMoney int64,
	tax int64,
	date time.Time,
	updateFields map[string]interface{},
) error {
	// 获取日期（只取日期部分，忽略时间）
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	// 根据类型选择不同的处理方式
	switch v := stats.(type) {
	case *models.PayChannelDayStatistics:
		return s.successPayChannelDayStatistics(ctx, v, successMoney, tax, dateOnly, updateFields)
	case *models.MerchantDayStatistics:
		return s.successMerchantDayStatistics(ctx, v, successMoney, tax, dateOnly, updateFields)
	case *models.TenantDayStatistics:
		return s.successTenantDayStatistics(ctx, v, successMoney, tax, dateOnly, updateFields)
	case *models.WriteOffDayStatistics:
		return s.successWriteoffDayStatistics(ctx, v, successMoney, tax, dateOnly, updateFields)
	case *models.DayStatistics:
		return s.successDayStatistics(ctx, v, successMoney, tax, dateOnly, updateFields)
	default:
		return fmt.Errorf("不支持的统计类型: %T", stats)
	}
}

// successPayChannelDayStatistics 更新通道日统计
func (s *StatisticsService) successPayChannelDayStatistics(
	ctx context.Context,
	stats *models.PayChannelDayStatistics,
	successMoney int64,
	tax int64,
	date time.Time,
	updateFields map[string]interface{},
) error {
	// 设置默认值
	if stats.SuccessCount == 0 {
		stats.SuccessCount = 1
	}
	if stats.SuccessMoney == 0 {
		stats.SuccessMoney = successMoney
	}
	if stats.TotalTax == 0 {
		stats.TotalTax = tax
	}
	if stats.Ver == 0 {
		stats.Ver = 1
	}
	stats.Date = date

	// 合并更新字段
	doUpdatesMap := map[string]interface{}{
		"success_count": gorm.Expr("success_count + 1"),
		"success_money": gorm.Expr("success_money + ?", successMoney),
		"total_tax":     gorm.Expr("total_tax + ?", tax),
		"ver":           gorm.Expr("ver + 1"),
	}

	// 添加额外的更新字段
	for k, v := range updateFields {
		doUpdatesMap[k] = v
	}

	// 确定唯一索引列
	var conflictColumns []clause.Column
	if stats.PayChannelID != nil {
		conflictColumns = append(conflictColumns, clause.Column{Name: "date"}, clause.Column{Name: "pay_channel_id"})
		if stats.TenantID != nil {
			conflictColumns = append(conflictColumns, clause.Column{Name: "tenant_id"})
		}
		if stats.MerchantID != nil {
			conflictColumns = append(conflictColumns, clause.Column{Name: "merchant_id"})
		}
		if stats.WriteoffID != nil {
			conflictColumns = append(conflictColumns, clause.Column{Name: "writeoff_id"})
		}
	}

	if len(conflictColumns) == 0 {
		return fmt.Errorf("通道统计缺少必要的索引字段")
	}

	err := database.DB.Clauses(clause.OnConflict{
		Columns:   conflictColumns,
		DoUpdates: clause.Assignments(doUpdatesMap),
	}).Create(stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新通道日统计失败: %w", err)
	}

	return nil
}

// successMerchantDayStatistics 更新商户日统计
func (s *StatisticsService) successMerchantDayStatistics(
	ctx context.Context,
	stats *models.MerchantDayStatistics,
	successMoney int64,
	tax int64,
	date time.Time,
	updateFields map[string]interface{},
) error {
	// 设置默认值
	if stats.SuccessCount == 0 {
		stats.SuccessCount = 1
	}
	if stats.SuccessMoney == 0 {
		stats.SuccessMoney = successMoney
	}
	if stats.TotalTax == 0 {
		stats.TotalTax = tax
	}
	if stats.Ver == 0 {
		stats.Ver = 1
	}
	stats.Date = date

	// 合并更新字段
	doUpdatesMap := map[string]interface{}{
		"success_count": gorm.Expr("success_count + 1"),
		"success_money": gorm.Expr("success_money + ?", successMoney),
		"total_tax":     gorm.Expr("total_tax + ?", tax),
		"ver":           gorm.Expr("ver + 1"),
	}

	// 添加额外的更新字段
	for k, v := range updateFields {
		doUpdatesMap[k] = v
	}

	if stats.MerchantID == nil {
		return fmt.Errorf("商户统计缺少 merchant_id")
	}

	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
			{Name: "merchant_id"},
		},
		DoUpdates: clause.Assignments(doUpdatesMap),
	}).Create(stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新商户日统计失败: %w", err)
	}

	return nil
}

// successTenantDayStatistics 更新租户日统计
func (s *StatisticsService) successTenantDayStatistics(
	ctx context.Context,
	stats *models.TenantDayStatistics,
	successMoney int64,
	tax int64,
	date time.Time,
	updateFields map[string]interface{},
) error {
	// 设置默认值
	if stats.SuccessCount == 0 {
		stats.SuccessCount = 1
	}
	if stats.SuccessMoney == 0 {
		stats.SuccessMoney = successMoney
	}
	if stats.TotalTax == 0 {
		stats.TotalTax = tax
	}
	if stats.Ver == 0 {
		stats.Ver = 1
	}
	stats.Date = date

	// 合并更新字段
	doUpdatesMap := map[string]interface{}{
		"success_count": gorm.Expr("success_count + 1"),
		"success_money": gorm.Expr("success_money + ?", successMoney),
		"total_tax":     gorm.Expr("total_tax + ?", tax),
		"ver":           gorm.Expr("ver + 1"),
	}

	// 添加额外的更新字段
	for k, v := range updateFields {
		doUpdatesMap[k] = v
	}

	if stats.TenantID == nil {
		return fmt.Errorf("租户统计缺少 tenant_id")
	}

	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
			{Name: "tenant_id"},
		},
		DoUpdates: clause.Assignments(doUpdatesMap),
	}).Create(stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新租户日统计失败: %w", err)
	}

	return nil
}

// successWriteoffDayStatistics 更新核销日统计
func (s *StatisticsService) successWriteoffDayStatistics(
	ctx context.Context,
	stats *models.WriteOffDayStatistics,
	successMoney int64,
	tax int64,
	date time.Time,
	updateFields map[string]interface{},
) error {
	// 设置默认值
	if stats.SuccessCount == 0 {
		stats.SuccessCount = 1
	}
	if stats.SuccessMoney == 0 {
		stats.SuccessMoney = successMoney
	}
	if stats.TotalTax == 0 {
		stats.TotalTax = tax
	}
	if stats.Ver == 0 {
		stats.Ver = 1
	}
	stats.Date = date

	// 合并更新字段
	doUpdatesMap := map[string]interface{}{
		"success_count": gorm.Expr("success_count + 1"),
		"success_money": gorm.Expr("success_money + ?", successMoney),
		"total_tax":     gorm.Expr("total_tax + ?", tax),
		"ver":           gorm.Expr("ver + 1"),
	}

	// 添加额外的更新字段
	for k, v := range updateFields {
		doUpdatesMap[k] = v
	}

	if stats.WriteoffID == nil {
		return fmt.Errorf("核销统计缺少 writeoff_id")
	}

	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
			{Name: "writeoff_id"},
		},
		DoUpdates: clause.Assignments(doUpdatesMap),
	}).Create(stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新核销日统计失败: %w", err)
	}

	return nil
}

// successDayStatistics 更新全局日统计
func (s *StatisticsService) successDayStatistics(
	ctx context.Context,
	stats *models.DayStatistics,
	successMoney int64,
	tax int64,
	date time.Time,
	updateFields map[string]interface{},
) error {
	// 设置默认值
	if stats.SuccessCount == 0 {
		stats.SuccessCount = 1
	}
	if stats.SuccessMoney == 0 {
		stats.SuccessMoney = successMoney
	}
	if stats.TotalTax == 0 {
		stats.TotalTax = tax
	}
	if stats.Ver == 0 {
		stats.Ver = 1
	}
	stats.Date = date

	// 合并更新字段
	doUpdatesMap := map[string]interface{}{
		"success_count": gorm.Expr("success_count + 1"),
		"success_money": gorm.Expr("success_money + ?", successMoney),
		"total_tax":     gorm.Expr("total_tax + ?", tax),
		"ver":           gorm.Expr("ver + 1"),
	}

	// 添加额外的更新字段
	for k, v := range updateFields {
		doUpdatesMap[k] = v
	}

	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
		},
		DoUpdates: clause.Assignments(doUpdatesMap),
	}).Create(stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新全局日统计失败: %w", err)
	}

	return nil
}
