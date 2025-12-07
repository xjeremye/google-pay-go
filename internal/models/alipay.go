package models

import (
	"time"
)

// AlipayProduct 支付宝产品模型
type AlipayProduct struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string     `gorm:"type:varchar(255);not null;comment:产品名称" json:"name"`
	AppID           string     `gorm:"type:varchar(32);comment:应用ID" json:"app_id,omitempty"`
	UID             string     `gorm:"type:varchar(32);comment:商户号" json:"uid,omitempty"`
	PrivateKey      string     `gorm:"type:text;comment:应用私钥" json:"private_key,omitempty"`
	PublicKey       string     `gorm:"type:text;comment:应用公钥" json:"public_key,omitempty"`
	AppPublicCrt    string     `gorm:"type:text;comment:应用公钥证书" json:"app_public_crt,omitempty"`
	AlipayPublicCrt string     `gorm:"type:text;comment:支付宝公钥证书" json:"alipay_public_crt,omitempty"`
	AlipayRootCrt   string     `gorm:"type:text;comment:支付宝根证书" json:"alipay_root_crt,omitempty"`
	SignType        string     `gorm:"type:varchar(10);default:'0';comment:签名类型" json:"sign_type"`
	AccountType     int        `gorm:"not null;comment:账户类型" json:"account_type"`
	AppAuthToken    string     `gorm:"type:varchar(64);comment:应用授权令牌" json:"app_auth_token,omitempty"`
	Subject         string     `gorm:"type:varchar(255);comment:订单主题" json:"subject,omitempty"`
	ProxyIP         string     `gorm:"type:varchar(255);comment:代理IP" json:"proxy_ip,omitempty"`
	ProxyPort       int        `gorm:"comment:代理端口" json:"proxy_port,omitempty"`
	ProxyUser       string     `gorm:"type:varchar(255);comment:代理用户名" json:"proxy_user,omitempty"`
	ProxyPwd        string     `gorm:"type:varchar(255);comment:代理密码" json:"proxy_pwd,omitempty"`
	Status          bool       `gorm:"not null;default:1;comment:状态" json:"status"`
	IsDelete        bool       `gorm:"not null;default:0;comment:是否删除" json:"is_delete"`
	ParentID        *int64     `gorm:"index;comment:父产品ID" json:"parent_id,omitempty"`
	CreateDatetime  *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime  *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	
	// 产品限制相关字段
	LimitMoney      int    `gorm:"not null;default:0;comment:限额" json:"limit_money"`
	MaxMoney        int    `gorm:"not null;default:0;comment:最大金额" json:"max_money"`
	MinMoney        int    `gorm:"not null;default:0;comment:最小金额" json:"min_money"`
	FloatMaxMoney   int    `gorm:"not null;default:0;comment:浮动最大金额" json:"float_max_money"`
	FloatMinMoney   int    `gorm:"not null;default:0;comment:浮动最小金额" json:"float_min_money"`
	CanPay          bool   `gorm:"not null;default:1;comment:是否允许进单" json:"can_pay"`
	SettledMoneys   string `gorm:"type:json;default:'[]';comment:固定金额列表" json:"settled_moneys,omitempty"`
	DayCountLimit   int    `gorm:"not null;default:0;comment:日笔数限制" json:"day_count_limit"`
	WriteoffID      int64  `gorm:"index;not null;comment:关联核销" json:"writeoff_id"`

	// 关联关系
	Parent   *AlipayProduct `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Writeoff *Writeoff      `gorm:"foreignKey:WriteoffID" json:"writeoff,omitempty"`
}

// TableName 指定表名
func (AlipayProduct) TableName() string {
	return "dvadmin_alipay_product"
}

