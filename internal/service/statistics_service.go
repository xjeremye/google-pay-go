package service

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
	case *models.WriteOffChannelDayStatistics:
		return s.successWriteoffChannelDayStatistics(ctx, v, successMoney, tax, dateOnly, updateFields)
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

	// 记录日志，帮助调试
	logger.Logger.Debug("更新租户日统计",
		zap.Time("date", date),
		zap.Int64("tenant_id", *stats.TenantID),
		zap.Int64("success_money", successMoney),
		zap.Int64("tax", tax))

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

	// 记录更新后的值（查询确认）
	var updatedStats models.TenantDayStatistics
	if err := database.DB.Where("date = ? AND tenant_id = ?", date, *stats.TenantID).First(&updatedStats).Error; err == nil {
		logger.Logger.Info("租户日统计更新成功",
			zap.Time("date", date),
			zap.Int64("tenant_id", *stats.TenantID),
			zap.Int64("total_tax", updatedStats.TotalTax),
			zap.Int64("success_money", updatedStats.SuccessMoney),
			zap.Int("success_count", updatedStats.SuccessCount))
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
	// 记录日志，帮助调试 tax 值
	logger.Logger.Info("更新核销日统计",
		zap.Int64("writeoff_id", *stats.WriteoffID),
		zap.Time("date", date),
		zap.Int64("success_money", successMoney),
		zap.Int64("tax", tax))

	// 设置默认值
	if stats.SuccessCount == 0 {
		stats.SuccessCount = 1
	}
	if stats.SuccessMoney == 0 {
		stats.SuccessMoney = successMoney
	}
	if stats.SubmitMoney == 0 {
		// 如果是新记录，submit_money 应该等于 success_money（成功的订单金额就是提交的金额）
		stats.SubmitMoney = successMoney
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

// successWriteoffChannelDayStatistics 更新核销通道日统计
// 参考 Python: success_base_day_statistics(WriteOffChannelDayStatistics, ...)
func (s *StatisticsService) successWriteoffChannelDayStatistics(
	ctx context.Context,
	stats *models.WriteOffChannelDayStatistics,
	successMoney int64,
	tax int64,
	date time.Time,
	updateFields map[string]interface{},
) error {
	// 记录日志，帮助调试
	logger.Logger.Info("更新核销通道日统计",
		zap.Int64("writeoff_id", *stats.WriteoffID),
		zap.Int64("pay_channel_id", *stats.PayChannelID),
		zap.Time("date", date),
		zap.Int64("success_money", successMoney),
		zap.Int64("tax", tax))

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
		return fmt.Errorf("核销通道统计缺少 writeoff_id")
	}
	if stats.PayChannelID == nil {
		return fmt.Errorf("核销通道统计缺少 pay_channel_id")
	}

	// 唯一约束: (date, writeoff_id, pay_channel_id)
	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
			{Name: "writeoff_id"},
			{Name: "pay_channel_id"},
		},
		DoUpdates: clause.Assignments(doUpdatesMap),
	}).Create(stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新核销通道日统计失败: %w", err)
	}

	return nil
}

// SubmitBaseDayStatistics 提交日统计（通用方法）
// 参考 Python: submit_base_day_statistics
func (s *StatisticsService) SubmitBaseDayStatistics(
	ctx context.Context,
	stats interface{},
	date time.Time,
) error {
	// 获取日期（只取日期部分，忽略时间）
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	// 根据类型选择不同的处理方式
	switch v := stats.(type) {
	case *models.WriteOffChannelDayStatistics:
		return s.submitWriteoffChannelDayStatistics(ctx, v, dateOnly)
	default:
		return fmt.Errorf("不支持的统计类型: %T", stats)
	}
}

// submitWriteoffChannelDayStatistics 更新核销通道日统计（订单提交时）
// 参考 Python: submit_base_day_statistics(WriteOffChannelDayStatistics, ...)
func (s *StatisticsService) submitWriteoffChannelDayStatistics(
	ctx context.Context,
	stats *models.WriteOffChannelDayStatistics,
	date time.Time,
) error {
	stats.Date = date
	if stats.SubmitCount == 0 {
		stats.SubmitCount = 1
	}
	if stats.Ver == 0 {
		stats.Ver = 1
	}

	if stats.WriteoffID == nil {
		return fmt.Errorf("核销通道统计缺少 writeoff_id")
	}
	if stats.PayChannelID == nil {
		return fmt.Errorf("核销通道统计缺少 pay_channel_id")
	}

	// 唯一约束: (date, writeoff_id, pay_channel_id)
	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
			{Name: "writeoff_id"},
			{Name: "pay_channel_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"submit_count": gorm.Expr("submit_count + 1"),
			"ver":          gorm.Expr("ver + 1"),
		}),
	}).Create(stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新核销通道日统计失败: %w", err)
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

	// 记录日志，帮助调试
	logger.Logger.Debug("更新全局日统计",
		zap.Time("date", date),
		zap.Int64("success_money", successMoney),
		zap.Int64("tax", tax),
		zap.Int("device_fields_count", len(updateFields)))

	err := database.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
		},
		DoUpdates: clause.Assignments(doUpdatesMap),
	}).Create(stats).Error

	if err != nil {
		return fmt.Errorf("创建/更新全局日统计失败: %w", err)
	}

	// 记录更新后的值（查询确认）
	var updatedStats models.DayStatistics
	if err := database.DB.Where("date = ?", date).First(&updatedStats).Error; err == nil {
		logger.Logger.Info("全局日统计更新成功",
			zap.Time("date", date),
			zap.Int64("total_tax", updatedStats.TotalTax),
			zap.Int64("success_money", updatedStats.SuccessMoney),
			zap.Int("success_count", updatedStats.SuccessCount))
	}

	return nil
}
