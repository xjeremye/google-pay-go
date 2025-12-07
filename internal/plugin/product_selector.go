package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
)

// getAlipayProduct 获取支付宝产品
// 参考 Python: AlipayFacePluginResponder.get_writeoff_product
// 在商户所属的码商(writeoff)下获取 alipay product 中所有可用的随机一个
func getAlipayProduct(ctx context.Context, req *WaitProductRequest, writeoffIDs []int64) (string, *int64, int, error) {
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
	for _, product := range products {
		// 检查固定金额
		// Python: Q(settled_moneys=[]) | Q(settled_moneys__contains=[money])
		if !checkSettledMoneys(product.SettledMoneys, money) {
			continue
		}

		// 检查限额（简化版，Python 有更复杂的逻辑）
		if product.LimitMoney > 0 {
			// TODO: 实现日限额检查（需要查询日统计表）
			// Python: 检查当日已收款金额+五分钟内等待支付的订单总和是否超过限额
			// 暂时跳过限额检查
		}

		// 检查日笔数限制（简化版）
		if product.DayCountLimit > 0 {
			// TODO: 实现日笔数限制检查（需要 Redis 计数）
			// Python: atomic_incr_decr_redis_count(success_key, 1, i['day_count_limit'], ex=3600 * 24)
			// 暂时跳过日笔数限制检查
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
		_ = todayTime // 暂时未使用，后续实现日限额检查时会用到
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
