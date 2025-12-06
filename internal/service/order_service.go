package service

import (
	"fmt"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/utils"
	"gorm.io/gorm"
)

type OrderService struct{}

// NewOrderService 创建订单服务
func NewOrderService() *OrderService {
	return &OrderService{}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(req *CreateOrderRequest) (*models.Order, error) {
	// 生成订单号
	orderNo := utils.GenerateOrderNo()
	
	now := time.Now()
	order := &models.Order{
		ID:            utils.GenerateID(),
		OrderNo:       orderNo,
		OutOrderNo:    req.OutOrderNo,
		OrderStatus:   models.OrderStatusPending,
		Money:         req.Money,
		Tax:           req.Tax,
		ProductName:   req.ProductName,
		ReqExtra:      req.ReqExtra,
		CreateDatetime: &now,
		Compatible:    0,
		Ver:           1,
		MerchantID:    req.MerchantID,
		PayChannelID:  req.PayChannelID,
	}

	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建订单
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	// 创建订单详情
	orderDetail := &models.OrderDetail{
		OrderID:     order.ID,
		NotifyURL:   req.NotifyURL,
		JumpURL:     req.JumpURL,
		ProductID:   req.ProductID,
		NotifyMoney: req.NotifyMoney,
		CreateDatetime: &now,
		MerchantTax: req.MerchantTax,
		Extra:       req.Extra,
	}

	if err := tx.Create(orderDetail).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建订单详情失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// 加载关联数据
	if err := database.DB.Preload("OrderDetail").First(order, order.ID).Error; err != nil {
		return nil, fmt.Errorf("加载订单数据失败: %w", err)
	}

	return order, nil
}

// GetOrderByOrderNo 根据订单号获取订单
func (s *OrderService) GetOrderByOrderNo(orderNo string) (*models.Order, error) {
	var order models.Order
	err := database.DB.Preload("OrderDetail").
		Preload("Merchant").
		Preload("PayChannel").
		Where("order_no = ?", orderNo).
		First(&order).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("订单不存在")
		}
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	return &order, nil
}

// GetOrderByOutOrderNo 根据商户订单号获取订单
func (s *OrderService) GetOrderByOutOrderNo(outOrderNo string, merchantID int64) (*models.Order, error) {
	var order models.Order
	err := database.DB.Preload("OrderDetail").
		Where("out_order_no = ? AND merchant_id = ?", outOrderNo, merchantID).
		First(&order).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("订单不存在")
		}
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	return &order, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(orderID string, status int, ticketNo string) error {
	now := time.Now()
	
	// 先更新版本号（使用原生 SQL）
	if err := database.DB.Exec("UPDATE dvadmin_order SET ver = ver + 1 WHERE id = ?", orderID).Error; err != nil {
		return fmt.Errorf("更新版本号失败: %w", err)
	}

	updates := map[string]interface{}{
		"order_status": status,
		"update_datetime": &now,
	}

	if status == models.OrderStatusPaid {
		updates["pay_datetime"] = &now
	}

	if ticketNo != "" {
		// 更新订单详情中的流水号
		if err := database.DB.Model(&models.OrderDetail{}).
			Where("order_id = ?", orderID).
			Update("ticket_no", ticketNo).Error; err != nil {
			return fmt.Errorf("更新订单详情失败: %w", err)
		}
	}

	if err := database.DB.Model(&models.Order{}).
		Where("id = ?", orderID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	return nil
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	OutOrderNo   string
	Money        int
	Tax          int
	ProductName  string
	ReqExtra     string
	NotifyURL    string
	JumpURL      string
	ProductID    string
	NotifyMoney  int
	MerchantTax  int
	Extra        string
	MerchantID   *int64
	PayChannelID *int64
}

