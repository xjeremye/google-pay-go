package models

import (
	"time"
)

// DayStatistics 全局日统计模型
type DayStatistics struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SuccessCount int       `gorm:"not null;default:0;comment:成功订单数" json:"success_count"`
	SubmitCount  int       `gorm:"not null;default:0;comment:总提交订单数" json:"submit_count"`
	SuccessMoney int64     `gorm:"not null;default:0;comment:总收入" json:"success_money"`
	SubmitMoney  int64     `gorm:"not null;default:0;comment:总提交收入" json:"submit_money"`
	TotalTax     int64     `gorm:"not null;default:0;comment:总利润" json:"total_tax"`
	Date         time.Time `gorm:"type:date;uniqueIndex;not null;comment:日期" json:"date"`
	Ver          int64     `gorm:"not null;comment:版本号" json:"ver"`
	UnknownCount int       `gorm:"not null;default:0;comment:未知设备订单数" json:"unknown_count"`
	AndroidCount int       `gorm:"not null;default:0;comment:安卓订单数" json:"android_count"`
	IOSCount     int       `gorm:"not null;default:0;comment:苹果订单数" json:"ios_count"`
	PCCount      int       `gorm:"not null;default:0;comment:电脑(web)订单数" json:"pc_count"`
}

// TableName 指定表名
func (DayStatistics) TableName() string {
	return "dvadmin_day_statistics"
}

// PayChannelDayStatistics 支付通道日统计模型
type PayChannelDayStatistics struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SuccessCount int       `gorm:"not null;default:0;comment:成功订单数" json:"success_count"`
	SubmitCount  int       `gorm:"not null;default:0;comment:总提交订单数" json:"submit_count"`
	SuccessMoney int64     `gorm:"not null;default:0;comment:总收入" json:"success_money"`
	TotalTax     int64     `gorm:"not null;default:0;comment:总利润" json:"total_tax"`
	RealMoney    int64     `gorm:"not null;default:0;comment:实际收入" json:"real_money"`
	Date         time.Time `gorm:"type:date;not null;comment:日期" json:"date"`
	Ver          int64     `gorm:"not null;comment:版本号" json:"ver"`
	PayChannelID *int64    `gorm:"index;comment:关联支付通道" json:"pay_channel_id,omitempty"`
	TenantID     *int64    `gorm:"index;comment:租户" json:"tenant_id,omitempty"`
	MerchantID   *int64    `gorm:"index;comment:商户" json:"merchant_id,omitempty"`
	WriteoffID   *int64    `gorm:"index;comment:核销" json:"writeoff_id,omitempty"`
	UnknownCount int       `gorm:"not null;default:0;comment:未知设备订单数" json:"unknown_count"`
	AndroidCount int       `gorm:"not null;default:0;comment:安卓订单数" json:"android_count"`
	IOSCount     int       `gorm:"not null;default:0;comment:苹果订单数" json:"ios_count"`
	PCCount      int       `gorm:"not null;default:0;comment:电脑(web)订单数" json:"pc_count"`
}

// TableName 指定表名
func (PayChannelDayStatistics) TableName() string {
	return "dvadmin_day_statistics_pay_channel"
}

// MerchantDayStatistics 商户日统计模型
type MerchantDayStatistics struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SuccessCount int       `gorm:"not null;default:0;comment:成功订单数" json:"success_count"`
	SubmitCount  int       `gorm:"not null;default:0;comment:总提交订单数" json:"submit_count"`
	SuccessMoney int64     `gorm:"not null;default:0;comment:总收入" json:"success_money"`
	TotalTax     int64     `gorm:"not null;default:0;comment:总利润" json:"total_tax"`
	RealMoney    int64     `gorm:"not null;default:0;comment:实际收入" json:"real_money"`
	Date         time.Time `gorm:"type:date;not null;comment:日期" json:"date"`
	Ver          int64     `gorm:"not null;comment:版本号" json:"ver"`
	MerchantID   *int64    `gorm:"index;comment:商户" json:"merchant_id,omitempty"`
	UnknownCount int       `gorm:"not null;default:0;comment:未知设备订单数" json:"unknown_count"`
	AndroidCount int       `gorm:"not null;default:0;comment:安卓订单数" json:"android_count"`
	IOSCount     int       `gorm:"not null;default:0;comment:苹果订单数" json:"ios_count"`
	PCCount      int       `gorm:"not null;default:0;comment:电脑(web)订单数" json:"pc_count"`
}

// TableName 指定表名
func (MerchantDayStatistics) TableName() string {
	return "dvadmin_day_statistics_merchant"
}

// TenantDayStatistics 租户日统计模型
type TenantDayStatistics struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SuccessCount int       `gorm:"not null;default:0;comment:成功订单数" json:"success_count"`
	SubmitCount  int       `gorm:"not null;default:0;comment:总提交订单数" json:"submit_count"`
	SuccessMoney int64     `gorm:"not null;default:0;comment:总收入" json:"success_money"`
	TotalTax     int64     `gorm:"not null;default:0;comment:总利润" json:"total_tax"`
	Date         time.Time `gorm:"type:date;not null;comment:日期" json:"date"`
	Ver          int64     `gorm:"not null;comment:版本号" json:"ver"`
	TenantID     *int64    `gorm:"index;comment:租户" json:"tenant_id,omitempty"`
	UnknownCount int       `gorm:"not null;default:0;comment:未知设备订单数" json:"unknown_count"`
	AndroidCount int       `gorm:"not null;default:0;comment:安卓订单数" json:"android_count"`
	IOSCount     int       `gorm:"not null;default:0;comment:苹果订单数" json:"ios_count"`
	PCCount      int       `gorm:"not null;default:0;comment:电脑(web)订单数" json:"pc_count"`
}

// TableName 指定表名
func (TenantDayStatistics) TableName() string {
	return "dvadmin_day_statistics_tenant"
}

// WriteOffDayStatistics 核销日统计模型
type WriteOffDayStatistics struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SuccessCount int       `gorm:"not null;default:0;comment:成功订单数" json:"success_count"`
	SubmitCount  int       `gorm:"not null;default:0;comment:总提交订单数" json:"submit_count"`
	SuccessMoney int64     `gorm:"not null;default:0;comment:总收入" json:"success_money"`
	SubmitMoney  int64     `gorm:"not null;comment:总提交收入" json:"submit_money"`
	TotalTax     int64     `gorm:"not null;default:0;comment:总利润" json:"total_tax"`
	Date         time.Time `gorm:"type:date;not null;comment:日期" json:"date"`
	Ver          int64     `gorm:"not null;comment:版本号" json:"ver"`
	WriteoffID   *int64    `gorm:"index;comment:核销" json:"writeoff_id,omitempty"`
	UnknownCount int       `gorm:"not null;default:0;comment:未知设备订单数" json:"unknown_count"`
	AndroidCount int       `gorm:"not null;default:0;comment:安卓订单数" json:"android_count"`
	IOSCount     int       `gorm:"not null;default:0;comment:苹果订单数" json:"ios_count"`
	PCCount      int       `gorm:"not null;default:0;comment:电脑(web)订单数" json:"pc_count"`
}

// TableName 指定表名
func (WriteOffDayStatistics) TableName() string {
	return "dvadmin_day_statistics_writeoff"
}
