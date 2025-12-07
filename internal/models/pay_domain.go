package models

import (
	"time"
)

// PayDomain 支付域名模型
type PayDomain struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Remarks         string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
	Modifier        string     `gorm:"type:varchar(255);comment:修改人" json:"modifier,omitempty"`
	UpdateDatetime  *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	CreateDatetime  *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	URL              string     `gorm:"uniqueIndex;type:varchar(255);not null;comment:域名" json:"url"`
	AppID            string     `gorm:"type:varchar(255);comment:app_id" json:"app_id,omitempty"`
	Status           bool       `gorm:"not null;default:1;comment:状态" json:"status"`
	PayStatus        bool       `gorm:"not null;default:0;comment:支付状态" json:"pay_status"`
	WechatStatus     bool       `gorm:"not null;default:0;comment:微信状态" json:"wechat_status"`
	SignType         int        `gorm:"not null;default:0;comment:签名类型" json:"sign_type"`
	PublicKey        string     `gorm:"type:longtext;comment:支付宝公钥" json:"public_key,omitempty"`
	PrivateKey       string     `gorm:"type:longtext;comment:应用私钥" json:"private_key,omitempty"`
	AppPublicCrt     string     `gorm:"type:longtext;comment:应用公钥证书" json:"app_public_crt,omitempty"`
	AlipayPublicCrt  string     `gorm:"type:longtext;comment:支付宝公钥证书" json:"alipay_public_crt,omitempty"`
	AlipayRootCrt    string     `gorm:"type:longtext;comment:支付宝根证书" json:"alipay_root_crt,omitempty"`
	AuthStatus       bool       `gorm:"not null;default:1;comment:鉴权状态" json:"auth_status"`
	AuthTimeout      int        `gorm:"not null;default:0;comment:鉴权时间" json:"auth_timeout"`
	AuthKey          string     `gorm:"type:varchar(255);comment:鉴权密钥" json:"auth_key,omitempty"`
	CreatorID        *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
}

// TableName 指定表名
func (PayDomain) TableName() string {
	return "dvadmin_pay_domain"
}

