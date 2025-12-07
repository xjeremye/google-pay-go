package models

import "time"

// MerchantNotification 商户通知模型
type MerchantNotification struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID        string     `gorm:"uniqueIndex;type:varchar(30);not null;comment:关联订单" json:"order_id"`
	Status         int        `gorm:"not null;comment:通知状态" json:"status"`
	Ver            int64      `gorm:"not null;comment:版本号" json:"ver"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
}

// TableName 指定表名
func (MerchantNotification) TableName() string {
	return "dvadmin_merchant_notification"
}

// NotificationStatus 通知状态常量
const (
	NotificationStatusPending  = 0 // 待通知
	NotificationStatusSuccess  = 1 // 通知成功
	NotificationStatusFailed   = 2 // 通知失败
	NotificationStatusRetrying = 3 // 重试中
	NotificationStatusMaxRetry = 4 // 达到最大重试次数
)

// MerchantNotificationHistory 商户通知记录模型
type MerchantNotificationHistory struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	NotificationID int64      `gorm:"index;not null;comment:关联通知" json:"notification_id"`
	URL            string     `gorm:"type:longtext;not null;comment:通知地址" json:"url"`
	RequestBody    string     `gorm:"type:longtext;comment:请求参数" json:"request_body,omitempty"`
	RequestMethod  string     `gorm:"type:varchar(8);comment:请求方式" json:"request_method,omitempty"`
	ResponseCode   int        `gorm:"not null;comment:响应状态码" json:"response_code"`
	JSONResult     string     `gorm:"type:longtext;comment:返回信息" json:"json_result,omitempty"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
}

// TableName 指定表名
func (MerchantNotificationHistory) TableName() string {
	return "dvadmin_merchant_notification_history"
}
