package alipay

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/plugin"
)

// getAlipayProduct 获取支付宝产品
// 参考 Python: AlipayFacePluginResponder.get_writeoff_product
// 在商户所属的码商(writeoff)下获取 alipay product 中所有可用的随机一个
func getAlipayProduct(ctx context.Context, req *plugin.WaitProductRequest, writeoffIDs []int64) (string, *int64, int, error) {
	// 构建查询条件
	// Python: AlipayProduct.objects.filter(
	//     Q(max_money=0, min_money=0) |
	//     Q(max_money__gt=0, min_money=0, max_money__gte=money) |
	//     Q(max_money=0, min_money__gt=0, min_money__lte=money) |
	//     Q(max_money__gt=0, min_money__gt=0, max_money__gte=money, min_money__lte=money),
	//     Q(allow_pay_channels__id=channel_id),
	//     Q(parent__is_delete=False, parent__status=True) | Q(parent__isnull=True),
	//     Q(writeoff_id__in=writeoff_ids),
	//     Q(settled_moneys=[]) | Q(settled_moneys__contains=[money]),
	//     can_pay=True,
	//     status=True, is_delete=False,
	// )

	money := req.Money
	query := database.DB.Model(&models.AlipayProduct{}).
		Where("can_pay = ?", true).
		Where("status = ?", true).
		Where("is_delete = ?", false).
		Where("writeoff_id IN ?", writeoffIDs).
		Where("(max_money = 0 AND min_money = 0) OR "+
			"(max_money > 0 AND min_money = 0 AND max_money >= ?) OR "+
			"(max_money = 0 AND min_money > 0 AND min_money <= ?) OR "+
			"(max_money > 0 AND min_money > 0 AND max_money >= ? AND min_money <= ?)",
			money, money, money, money)

	// 检查父产品状态
	// Python: Q(parent__is_delete=False, parent__status=True) | Q(parent__isnull=True)
	query = query.Where("(parent_id IS NULL) OR " +
		"(parent_id IN (SELECT id FROM dvadmin_alipay_product WHERE is_delete = 0 AND status = 1))")

	// 检查允许的支付通道
	// Python: Q(allow_pay_channels__id=channel_id)
	query = query.Where("id IN (SELECT alipayproduct_id FROM dvadmin_alipay_product_allow_pay_channels WHERE paychannel_id = ?)",
		req.ChannelID)

	// 先查询所有产品，然后在代码中过滤固定金额
	// 因为 JSON 字段的查询比较复杂

	// 随机排序（简化版，Python 使用权重排序）
	query = query.Order("RAND()")

	var products []models.AlipayProduct
	if err := query.Find(&products).Error; err != nil {
		return "", nil, money, fmt.Errorf("查询产品失败: %w", err)
	}

	if len(products) == 0 {
		return "", nil, money, nil
	}

	// 遍历产品，检查各种限制
	todayTime := time.Now().Add(-5 * time.Minute)
	today := time.Now().Format("2006-01-02")

	for _, product := range products {
		// 检查固定金额
		// Python: Q(settled_moneys=[]) | Q(settled_moneys__contains=[money])
		if !checkSettledMoneys(product.SettledMoneys, money) {
			continue
		}

		// 检查日限额
		if product.LimitMoney > 0 {
			// 检查当日已收款金额+五分钟内等待支付的订单总和是否超过限额
			if !checkDailyLimit(ctx, product.ID, req.ChannelID, product.LimitMoney, money, todayTime) {
				continue
			}
		}

		// 检查日笔数限制
		if product.DayCountLimit > 0 {
			// 使用 Redis 原子计数检查日笔数限制
			if !checkDailyCountLimit(ctx, product.ID, product.DayCountLimit, today) {
				continue
			}
		}

		// 应用浮动金额
		finalMoney := money
		if product.FloatMinMoney != product.FloatMaxMoney && product.FloatMaxMoney != 0 {
			// Python: money += random.randint(i["float_min_money"], i["float_max_money"])
			if product.FloatMaxMoney > product.FloatMinMoney {
				floatAmount := rand.Intn(product.FloatMaxMoney-product.FloatMinMoney+1) + product.FloatMinMoney
				finalMoney += floatAmount
			}
		}

		// 返回第一个符合条件的产品
		productIDStr := fmt.Sprintf("%d", product.ID)
		return productIDStr, &product.WriteoffID, finalMoney, nil
	}

	return "", nil, money, nil
}

// parseSettledMoneys 解析固定金额列表
func parseSettledMoneys(settledMoneys string) ([]int, error) {
	if settledMoneys == "" || settledMoneys == "[]" {
		return []int{}, nil
	}
	var moneys []int
	if err := json.Unmarshal([]byte(settledMoneys), &moneys); err != nil {
		return nil, fmt.Errorf("解析固定金额列表失败: %w", err)
	}
	return moneys, nil
}

// checkSettledMoneys 检查金额是否在固定金额列表中
func checkSettledMoneys(settledMoneys string, money int) bool {
	moneys, err := parseSettledMoneys(settledMoneys)
	if err != nil {
		return false
	}
	if len(moneys) == 0 {
		return true // 空列表表示不限制
	}
	for _, m := range moneys {
		if m == money {
			return true
		}
	}
	return false
}

// checkDailyLimit 检查日限额
// 参考 Python: 检查当日已收款金额+五分钟内等待支付的订单总和是否超过限额
func checkDailyLimit(ctx context.Context, productID int64, channelID int64, limitMoney int, currentMoney int, sinceTime time.Time) bool {
	// 1. 查询产品日统计表中的已收款金额
	var productDay models.AlipayProductDay
	today := time.Now().Format("2006-01-02")

	var totalMoney int64 = 0

	// 查询日统计表
	if err := database.DB.Where("product_id = ? AND pay_channel_id = ? AND date = ?",
		productID, channelID, today).
		First(&productDay).Error; err == nil {
		totalMoney = productDay.SuccessMoney
	}

	// 2. 查询5分钟内待支付订单的金额总和
	var pendingMoney int64
	database.DB.Table("dvadmin_order").
		Joins("JOIN dvadmin_order_detail ON dvadmin_order.id = dvadmin_order_detail.order_id").
		Where("dvadmin_order_detail.product_id = ?", fmt.Sprintf("%d", productID)).
		Where("dvadmin_order.order_status = ?", models.OrderStatusPending).
		Where("dvadmin_order.create_datetime >= ?", sinceTime).
		Select("COALESCE(SUM(dvadmin_order.money), 0)").
		Scan(&pendingMoney)

	// 3. 检查是否超过限额
	totalMoney += pendingMoney + int64(currentMoney)
	return totalMoney <= int64(limitMoney)
}

// checkDailyCountLimit 检查日笔数限制
// 参考 Python: atomic_incr_decr_redis_count(success_key, 1, i['day_count_limit'], ex=3600 * 24)
func checkDailyCountLimit(ctx context.Context, productID int64, dayCountLimit int, date string) bool {
	// Redis key 格式：product:day_count:{product_id}:{date}
	redisKey := fmt.Sprintf("product:day_count:%d:%s", productID, date)

	// 使用 Redis 原子操作检查计数
	rdb := database.RDB

	// 先获取当前计数（如果 key 不存在，返回 0）
	currentCount, err := rdb.Get(ctx, redisKey).Int()
	if err != nil {
		// 如果 key 不存在（redis: nil），这是正常的，计数为 0
		// 检查错误信息是否包含 "nil"（表示 key 不存在）
		if err.Error() == "redis: nil" {
			currentCount = 0
		} else {
			// Redis 其他错误，允许通过（容错处理）
			return true
		}
	}

	// 如果当前计数已经达到或超过限制，不允许
	if currentCount >= dayCountLimit {
		return false
	}

	// 原子递增并检查是否超过限制
	// 使用 INCR 命令，如果结果超过限制则回退
	newCount, err := rdb.Incr(ctx, redisKey).Result()
	if err != nil {
		// Redis 错误，允许通过（容错处理）
		return true
	}

	// 设置过期时间（24小时）
	rdb.Expire(ctx, redisKey, 24*time.Hour)

	// 如果递增后超过限制，回退并拒绝
	if newCount > int64(dayCountLimit) {
		rdb.Decr(ctx, redisKey)
		return false
	}

	return true
}
