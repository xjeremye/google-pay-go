package models

import "time"

// QueryLog 查询日志模型（用于记录支付宝等第三方 API 调用）
type QueryLog struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OutOrderNo     string     `gorm:"type:varchar(32);index;comment:外部订单号" json:"out_order_no,omitempty"`
	OrderNo        string     `gorm:"type:varchar(32);index;comment:系统订单号" json:"order_no,omitempty"`
	URL            string     `gorm:"type:longtext;comment:地址" json:"url,omitempty"`
	RequestBody    string     `gorm:"type:longtext;comment:请求参数" json:"request_body,omitempty"`
	RequestMethod  string     `gorm:"type:varchar(8);comment:请求方式" json:"request_method,omitempty"`
	ResponseCode   string     `gorm:"type:varchar(32);comment:响应状态码" json:"response_code,omitempty"`
	JSONResult     string     `gorm:"type:longtext;comment:返回信息" json:"json_result,omitempty"`
	Remarks        string     `gorm:"type:varchar(255);index;comment:备注" json:"remarks,omitempty"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
}

// TableName 指定表名
func (QueryLog) TableName() string {
	return "dvadmin_query_log"
}
