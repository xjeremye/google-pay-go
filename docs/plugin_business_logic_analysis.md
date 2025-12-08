# 支付插件业务逻辑分析文档

## 一、架构概述

### 1.1 插件系统架构

支付插件系统采用**策略模式**和**模板方法模式**，通过基类定义通用流程，子类实现具体业务逻辑。

```
BasePluginResponder (基类)
├── AlipayFacePluginResponder (支付宝当面付)
├── AlipayQrPluginResponder (支付宝扫码支付)
├── AlipayWapPluginResponder (支付宝手机网站支付)
├── AlipayAppPluginResponder (支付宝APP支付)
└── CommonRechargePluginResponder (通用充值插件)
```

### 1.2 核心组件

- **BasePluginResponder**: 插件基类，定义通用接口和默认实现
- **订单创建流程**: `wait_product` → `create_order` → `callback_submit`
- **回调处理流程**: `check_notify_success` → `callback_success` → 统计更新
- **产品选择**: `get_writeoff_product` 根据条件筛选可用产品

---

## 二、订单创建流程

### 2.1 完整流程

```
1. 验证商户、租户、签名、订单号、通道、插件
2. wait_product (等待产品)
   ├── 获取Cookie (如果需要)
   ├── 获取核销ID列表 (writeoff_ids)
   └── 选择产品 (get_writeoff_product)
3. 创建订单记录 (Order + OrderDetail)
4. create_order (调用插件创建支付订单)
   └── 返回支付URL
5. callback_submit (下单回调)
   └── 更新日统计 (submit_base_day_statistics)
```

### 2.2 wait_product 方法详解

**位置**: `base_plugin.py::BasePluginResponder.wait_product`

**流程**:
1. **获取Cookie** (如果 `_need_cookie=True`)
   - 从 `TenantCookie` 中随机选择一个可用的Cookie
   - 检查Cookie状态和关联文件状态

2. **获取核销ID列表** (如果 `_need_writeoff=True`)
   - 调用 `get_writeoff_ids(tenant_id, money, channel_id)`
   - 过滤余额不足的核销
   - 过滤被禁用的支付通道关联

3. **选择产品** (`get_writeoff_product`)
   - 子类实现具体的产品选择逻辑
   - 返回: `(product_id, writeoff_id, final_money)`

### 2.3 create_order 方法详解

**位置**: `alipay_plugin.py::AlipayFacePluginResponder.create_order`

**流程**:
1. **生成支付URL** (`get_pay_url`)
   - 根据插件类型生成不同的支付URL
   - 当面付: 生成支付宝授权URL
   - 扫码支付: 调用支付宝API生成二维码
   - 手机网站支付: 调用支付宝API生成支付页面URL

2. **更新订单状态** (`update_order_wait`)
   - 将订单状态从 `0` (待支付) 更新为 `2` (等待支付)

3. **返回响应**
   ```python
   {
       "code": 0,
       "msg": "成功",
       "data": {
           "pay_url": "支付URL"
       }
   }
   ```

---

## 三、产品选择逻辑 (get_writeoff_product)

### 3.1 支付宝产品选择流程

**位置**: `alipay_plugin.py::AlipayFacePluginResponder.get_writeoff_product`

#### 3.1.1 公池模式 (extra_arg == 1)

```python
# 从公池中随机选择产品
pool = AlipayPublicPool.objects.filter(
    status=True, 
    alipay__is_delete=False
).order_by("?")

# 检查核销余额 (take_up_tax)
# 返回产品ID和核销ID
```

#### 3.1.2 普通模式 (extra_arg == 0)

**筛选条件**:
1. **金额范围匹配**:
   - `max_money=0, min_money=0` (无限制)
   - `max_money>0, min_money=0, max_money>=money` (仅最大金额限制)
   - `max_money=0, min_money>0, min_money<=money` (仅最小金额限制)
   - `max_money>0, min_money>0, max_money>=money, min_money<=money` (范围限制)

2. **通道绑定**: `allow_pay_channels__id=channel_id`

3. **产品状态**: `can_pay=True, status=True, is_delete=False`

4. **父产品状态**: `parent__is_delete=False, parent__status=True` 或 `parent__isnull=True`

5. **核销匹配**: `writeoff_id__in=writeoff_ids` 或 神码模式

6. **固定金额**: `settled_moneys=[]` 或 `settled_moneys__contains=[money]`

**权重排序**:
```python
priority = -Ln(1.0 - Random()) / Coalesce(weight, 1)
# 权重越大，优先级越高
```

**限制检查** (遍历产品时):
1. **神码限额** (如果使用神码):
   - 检查当日成功金额 + 5分钟内待支付订单金额
   - 不能超过 `shenma.limit_money`

2. **产品日限额** (`limit_money != 0`):
   - 检查当日成功金额 + 5分钟内待支付订单金额
   - 不能超过 `product.limit_money`

3. **日笔数限制** (`day_count_limit > 0`):
   - 使用Redis原子计数
   - 不能超过 `day_count_limit`

4. **浮动金额** (`float_min_money != float_max_money != 0`):
   ```python
   money += random.randint(float_min_money, float_max_money)
   ```

5. **设置运行标记**:
   ```python
   cache.set(f'running_alipay_{product_id}', 1, timeout=300)
   ```

---

## 四、回调处理流程

### 4.1 回调类型

#### 4.1.1 callback_submit (下单回调)

**触发时机**: 订单创建成功后

**功能**:
- 更新日统计 (submit_base_day_statistics)
- 根据模式选择统计表:
  - 公池模式: `AlipayPublicPoolDayStatistics`
  - 神码模式: `AlipayShenmaDayStatistics`
  - 普通模式: `AlipayProductDayStatistics`

#### 4.1.2 callback_success (支付成功回调)

**触发时机**: 支付成功通知到达

**功能**:
1. **更新日统计** (success_base_day_statistics)
   - 增加成功金额和笔数

2. **分账处理** (根据 `collection_type`):
   - **分账模式** (`collection_type=0`):
     - 扣除手续费: `money - max(int(money * 0.006), 1)`
     - 调用 `start_group_split` 进行分账
   
   - **智能出款** (`collection_type=3`):
     - 根据缓存中的比例 (`{product_id}_alipay_max_ratio`) 计算分账金额
     - 剩余金额分配给随机转账用户 (`get_rand_transfer_user`)
     - 如果无可用用户，记录错误到 `SplitHistory`
   
   - **自动转账** (`collection_type=1`):
     - 扣除手续费后全部转账

3. **结算确认** (`account_type=6`):
   - 调用 `start_settle_confirm` 进行结算确认

4. **日笔数限制恢复** (如果订单从失败状态恢复):
   ```python
   atomic_incr_decr_redis_count(
       f'{date.today()}_{product.id}_day_count_limit', 
       1, 
       product.day_count_limit, 
       ex=3600*24
   )
   ```

#### 4.1.3 callback_timeout (超时回调)

**触发时机**: 订单超时未支付

**功能**:
1. **检查连续失败次数**:
   - 查询最近 `max_fail_count` 笔订单
   - 如果全部失败，自动关闭产品 (`close_alipay`)

2. **恢复日笔数限制**:
   ```python
   atomic_incr_decr_redis_count(
       f'{date.today()}_{product.id}_day_count_limit', 
       -1, 
       product.day_count_limit, 
       ex=3600*24
   )
   ```

3. **触发查询任务**:
   - 延迟30秒后查询订单状态 (`query_order`)

#### 4.1.4 check_notify_success (通知验证)

**位置**: `alipay_plugin.py::AlipayFacePluginResponder.check_notify_success`

**功能**:
- 检查 `trade_status` 是否为 `"trade_success"` 或 `"TRADE_SUCCESS"`
- 返回布尔值表示是否支付成功

---

## 五、订单查询逻辑

### 5.1 query_order 方法

**位置**: `alipay_plugin.py::AlipayFacePluginResponder.query_order`

**功能**:
- 主动查询订单支付状态
- 调用支付宝API (`api_alipay_face_pay_query`)
- 如果查询到支付成功 (`code=102`)，调用 `success_order_by_query` 更新订单状态

**触发场景**:
1. 订单超时后延迟查询
2. 手动触发查询 (`actively=True`)

---

## 六、支付宝插件类型

### 6.1 AlipayFacePluginResponder (当面付)

- **Key**: `alipay_face_to`
- **支付方式**: 支付宝授权跳转
- **URL格式**: `alipays://platformapi/startapp?appId=20000067&url=...`

### 6.2 AlipayQrPluginResponder (扫码支付)

- **Key**: `alipay_ddm`
- **支付方式**: 生成二维码
- **API**: `alipay.api_alipay_trade_precreate`

### 6.3 AlipayWapPluginResponder (手机网站支付)

- **Key**: `alipay_wap`
- **支付方式**: 手机网站支付页面
- **API**: `alipay.api_alipay_trade_wap_pay`

### 6.4 AlipayAppPluginResponder (APP支付)

- **Key**: `alipay_app`
- **支付方式**: APP内支付
- **API**: `alipay.api_alipay_trade_app_pay`
- **产品码**: `QUICK_MSECURITY_PAY`

### 6.5 AlipayPcPluginResponder (电脑网站支付)

- **Key**: `alipay_pc`
- **支付方式**: PC网页支付
- **API**: `alipay.api_alipay_trade_page_pay`

---

## 七、关键工具函数

### 7.1 统计相关

- `submit_base_day_statistics`: 提交日统计 (下单时)
- `success_base_day_statistics`: 成功日统计 (支付成功时)
- `update_product_tax`: 更新产品税额 (已注释)

### 7.2 分账相关

- `start_group_split`: 启动分账任务
- `get_rand_transfer_user`: 获取随机转账用户
- `add_split_order_job`: 添加分账订单任务

### 7.3 产品管理

- `close_alipay`: 关闭支付宝产品
- `take_up_tax`: 占用核销税额
- `atomic_incr_decr_redis_count`: Redis原子计数操作

### 7.4 订单处理

- `update_order_wait`: 更新订单为等待状态
- `success_order_by_query`: 通过查询更新订单为成功状态
- `get_writeoff_ids`: 获取可用的核销ID列表

---

## 八、数据模型关系

### 8.1 核心模型

```
Order (订单)
├── OrderDetail (订单详情)
│   ├── product_id → AlipayProduct (产品)
│   ├── writeoff_id → Writeoff (核销)
│   ├── cookie_id → TenantCookie (Cookie)
│   └── plugin_id → PayPlugin (插件)
├── PayChannel (支付通道)
└── Merchant (商户)
    └── Tenant (租户)
```

### 8.2 统计模型

```
AlipayProductDayStatistics (产品日统计)
├── product_id → AlipayProduct
├── pay_channel_id → PayChannel
├── date (日期)
├── success_money (成功金额)
└── success_count (成功笔数)

AlipayShenmaDayStatistics (神码日统计)
AlipayPublicPoolDayStatistics (公池日统计)
```

### 8.3 产品模型

```
AlipayProduct (支付宝产品)
├── writeoff_id → Writeoff (核销)
├── parent_id → AlipayProduct (父产品)
├── limit_money (日限额)
├── max_money (最大金额)
├── min_money (最小金额)
├── day_count_limit (日笔数限制)
├── collection_type (收款类型: 0=分账, 1=自动转账, 3=智能出款)
└── account_type (账户类型: 0/7=普通, 6=结算)
```

---

## 九、业务模式

### 9.1 普通模式

- 产品直接绑定核销
- 使用 `AlipayProductDayStatistics` 统计
- 标准分账流程

### 9.2 公池模式 (extra_arg=1)

- 产品从公池中随机选择
- 使用 `AlipayPublicPoolDayStatistics` 统计
- 公池产品共享使用

### 9.3 神码模式

- 产品通过 `AlipayShenma` 关联租户
- 使用 `AlipayShenmaDayStatistics` 统计
- 支持跨租户使用产品

### 9.4 B2B模式 (extra_arg=3)

- 企业支付场景
- 使用 `extend_params` 传递企业支付参数

---

## 十、TODO 列表

### 10.1 代码优化

- [ ] **重构产品选择逻辑**: `get_writeoff_product` 方法过长 (200+行)，建议拆分为多个小方法
- [ ] **统一错误处理**: 各插件错误处理方式不一致，需要统一异常处理机制
- [ ] **注释清理**: 存在大量注释掉的代码，需要清理或删除
- [ ] **类型注解**: 部分方法缺少类型注解，建议补充完整的类型提示
- [ ] **日志优化**: 关键业务节点需要添加更详细的日志记录

### 10.2 功能完善

- [ ] **产品选择性能优化**: 
  - 当前使用数据库查询+循环过滤，性能较差
  - 建议使用Redis缓存产品列表，减少数据库查询
  - 考虑使用消息队列异步更新产品状态

- [ ] **分账逻辑完善**:
  - `collection_type=3` (智能出款) 的用户选择逻辑需要优化
  - 当无可用转账用户时，应该有重试机制或降级方案

- [ ] **统计功能增强**:
  - 当前统计只记录成功金额和笔数
  - 建议增加失败统计、平均金额、成功率等指标

- [ ] **超时处理优化**:
  - `callback_timeout` 中的连续失败检查逻辑可以优化
  - 建议使用滑动窗口算法，而不是固定数量检查

### 10.3 业务逻辑

- [ ] **公池模式完善**:
  - 当前公池选择逻辑较简单，只有随机选择
  - 建议增加权重选择、负载均衡等策略

- [ ] **日限额检查优化**:
  - 当前检查5分钟内的待支付订单，时间窗口固定
  - 建议可配置时间窗口，或使用更精确的预占机制

- [ ] **浮动金额处理**:
  - 浮动金额在订单创建时应用，但回调时金额可能不匹配
  - 需要确保回调金额验证逻辑正确处理浮动金额

- [ ] **神码模式限制**:
  - 神码模式的限额检查逻辑需要更详细的文档说明
  - 建议增加神码模式的监控和告警

### 10.4 测试和文档

- [ ] **单元测试**: 
  - 缺少插件核心方法的单元测试
  - 建议为 `get_writeoff_product`、`callback_success` 等关键方法编写测试

- [ ] **集成测试**:
  - 缺少完整的订单创建流程测试
  - 建议编写端到端测试，覆盖各种业务场景

- [ ] **API文档**:
  - 插件接口缺少详细的API文档
  - 建议使用Sphinx或类似工具生成API文档

- [ ] **业务文档**:
  - 缺少业务流程图和架构图
  - 建议绘制订单创建、回调处理等关键流程的流程图

### 10.5 安全和稳定性

- [ ] **并发控制**:
  - 产品选择时可能存在并发问题
  - 建议使用分布式锁确保产品选择的原子性

- [ ] **数据一致性**:
  - 统计数据的更新可能存在并发问题
  - 建议使用数据库事务或乐观锁保证一致性

- [ ] **异常恢复**:
  - 当插件调用失败时，缺少重试机制
  - 建议增加重试逻辑和降级方案

- [ ] **监控告警**:
  - 缺少关键业务指标的监控
  - 建议增加产品选择失败率、回调处理延迟等监控指标

### 10.6 代码质量

- [ ] **代码规范**:
  - 部分代码不符合PEP8规范
  - 建议使用black、flake8等工具统一代码风格

- [ ] **依赖管理**:
  - 存在大量直接导入，缺少依赖注入
  - 建议使用依赖注入框架，提高代码可测试性

- [ ] **配置管理**:
  - 硬编码的配置值较多 (如手续费0.006)
  - 建议将配置项提取到配置文件或数据库

### 10.7 性能优化

- [ ] **数据库查询优化**:
  - `get_writeoff_product` 中存在N+1查询问题
  - 建议使用 `select_related` 和 `prefetch_related` 优化查询

- [ ] **缓存策略**:
  - 产品信息、核销信息等可以缓存
  - 建议增加Redis缓存层，减少数据库压力

- [ ] **异步处理**:
  - 统计更新、分账等操作可以异步处理
  - 建议使用Celery任务队列，提高响应速度

### 10.8 功能扩展

- [ ] **多币种支持**:
  - 当前只支持人民币
  - 建议扩展支持多币种支付

- [ ] **退款功能**:
  - `callback_refund` 方法存在但逻辑不完整
  - 建议完善退款流程和统计

- [ ] **订单替换**:
  - `replace_writeoff_product` 方法存在但使用场景不明确
  - 建议明确使用场景并完善逻辑

---

## 十一、关键代码位置索引

### 11.1 核心文件

- `base_plugin.py`: 插件基类定义
- `alipay_plugin.py`: 支付宝插件实现
- `create_order_hook.py`: 订单创建钩子
- `common_recharge/base.py`: 通用充值插件

### 11.2 工具函数

- `utils/util.py`: 通用工具函数
- `utils/alipay_split.py`: 分账相关函数
- `utils/alipay_tool.py`: 支付宝工具函数
- `utils/transfer.py`: 转账相关函数

### 11.3 模型定义

- `models/business/alipay.py`: 支付宝相关模型
- `models/order.py`: 订单模型
- `models/pay_channel.py`: 支付通道模型

---

## 十二、总结

支付插件系统是一个复杂的业务系统，涉及订单创建、产品选择、回调处理、分账等多个环节。当前实现已经覆盖了主要业务场景，但在代码质量、性能优化、测试覆盖等方面还有改进空间。

建议优先处理以下事项：
1. 重构产品选择逻辑，提高代码可读性和性能
2. 完善单元测试和集成测试
3. 优化数据库查询，减少N+1问题
4. 增加监控和告警机制
5. 完善文档和注释
