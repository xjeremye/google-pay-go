# callback_plugin_submit 调用流程分析

## 一、函数定义和注册机制

### 1.1 函数定义

**文件位置**: `docs/backend/dvadmin/agent_manager/utils/create_order_hook.py`

```python
@order_submit_handle()  # 装饰器，将函数注册到钩子列表
def callback_plugin_submit(**kwargs):
    plugin_type = kwargs.get("plugin_type")
    if responder := get_plugin_by_key(plugin_type):
        try:
            responder.callback_submit(**kwargs)  # 调用插件的 callback_submit 方法
        except Exception as e:
            logger.error(f"插件{plugin_type} callback_plugin_submit触发错误:{e}")
```

### 1.2 注册机制

`@order_submit_handle()` 装饰器的工作原理：

```python
_create_order_submit = []  # 全局钩子函数列表

def order_submit_handle():
    """装饰器：将函数注册到 _create_order_submit 列表中"""
    def _decorator(func):
        _create_order_submit.append(func)  # 注册函数
        return func
    return _decorator

def notify_order_submit(**kwargs):
    """遍历所有注册的钩子函数并调用"""
    for func in _create_order_submit:
        try:
            func(**kwargs)  # 调用每个注册的函数
        except Exception as e:
            logger.error(f"触发notify_order_submit钩子函数({func})失败，错误信息：{e}")
```

**工作流程**:

1. 模块加载时，`callback_plugin_submit` 被 `@order_submit_handle()` 装饰
2. 装饰器将 `callback_plugin_submit` 添加到 `_create_order_submit` 列表
3. 当 `notify_order_submit(**kwargs)` 被调用时，会遍历列表并执行所有注册的函数

---

## 二、调用位置和时机

### 2.1 调用位置 1: 订单创建成功后

**文件位置**: `docs/backend/dvadmin/agent_manager/views/order.py`

**调用代码** (约第 900-903 行):

```python
# 在订单创建成功后，构建回调参数
create_args = {
    "order_no": ctx.order_no,
    "out_order_no": ctx.out_order_no,
    "plugin_id": ctx.plugin.id,
    "tax": ctx.tax,
    "plugin_type": ctx.plugin_type,
    "money": ctx.money,
    "domain_id": ctx.domain_id,
    "notify_money": ctx.notify_money,
    "order_id": ctx.order.id,
    "product_id": ctx.product_id,
    "cookie_id": ctx.cookie_id,
    "channel_id": ctx.channel_id,
    "merchant_id": ctx.merchant_id,
    "writeoff_id": ctx.writeoff_id,
    "tenant_id": ctx.tenant_id,
    "create_datetime": ctx.order.create_datetime,
    "notify_url": ctx.notify_url,
    "plugin_upstream": ctx.plugin_upstream,
}

# 通过调度器异步调用，延迟 500 微秒执行
if not cache.get(f"plugin_config.{ctx.plugin_type}.ignore_no_url"):
    scheduler.add_job(
        id=f"{ctx.order.order_no}-submit",
        func=notify_order_submit,  # 调用 notify_order_submit
        next_run_time=datetime.now() + timedelta(microseconds=500),
        kwargs=create_args
    )
```

**调用时机**:

- ✅ 订单创建成功
- ✅ 订单和订单详情已保存到数据库
- ✅ 支付URL已生成
- ✅ 延迟 500 微秒执行（确保订单数据已完全写入）

**触发条件**:

- 插件配置中 `ignore_no_url` 不为 `True`

### 2.2 调用位置 2: 订单失败时

**文件位置**: `docs/backend/dvadmin/agent_manager/views/order.py`

**调用代码** (约第 122-132 行):

```python
@retry_save()
def fail_order(order_no: str):
    """订单失败时的处理"""
    order_obj = Order.objects.filter(order_no=order_no).first()
    if order_obj:
        order_before = order_obj.order_status
        order_obj.order_status = 1  # 设置为失败状态
        order_obj.save()
        create_data = create_create_args(order_obj, order_before)  # 构建回调参数
        scheduler.add_job(
            id=f"order_no-submit",
            func=notify_order_submit,  # 调用 notify_order_submit
            kwargs=create_data
        )
```

**调用时机**:

- ✅ 订单创建失败
- ✅ 订单状态更新为失败 (status=1)
- ✅ 立即通过调度器异步执行

---

## 三、完整调用链

```
订单创建流程
    ↓
订单创建成功/失败
    ↓
构建回调参数 (create_args)
    ↓
scheduler.add_job(notify_order_submit, kwargs=create_args)
    ↓
notify_order_submit(**kwargs)
    ↓
遍历 _create_order_submit 列表
    ↓
callback_plugin_submit(**kwargs)
    ↓
get_plugin_by_key(plugin_type)
    ↓
responder.callback_submit(**kwargs)
    ↓
插件特定的 callback_submit 实现
    ├── AlipayFacePluginResponder.callback_submit
    ├── AlipayQrPluginResponder.callback_submit
    └── CommonRechargePluginResponder.callback_submit
    ↓
更新日统计 (submit_base_day_statistics)
```

---

## 四、参数说明

### 4.1 传递给 callback_plugin_submit 的参数

```python
{
    "order_no": str,           # 订单号
    "out_order_no": str,       # 商户订单号
    "plugin_id": int,           # 插件ID
    "tax": int,                # 税费
    "plugin_type": str,         # 插件类型 (如 "alipay_face_to")
    "money": int,              # 订单金额 (分)
    "domain_id": int,           # 域名ID
    "notify_money": int,        # 通知金额 (分)
    "order_id": int,           # 订单数据库ID
    "product_id": int,         # 产品ID
    "cookie_id": int,          # Cookie ID (可选)
    "channel_id": int,         # 支付通道ID
    "merchant_id": int,        # 商户ID
    "writeoff_id": int,        # 核销ID
    "tenant_id": int,          # 租户ID
    "create_datetime": datetime, # 订单创建时间
    "notify_url": str,         # 通知URL
    "plugin_upstream": int,    # 插件上游类型
}
```

### 4.2 插件 callback_submit 方法接收的参数

插件实现的 `callback_submit` 方法签名：

```python
def callback_submit(self, product_id: int, create_datetime: datetime, 
                   plugin_upstream, order_no, tenant_id, channel_id, 
                   money, **kwargs):
    # 实现统计更新等逻辑
    pass
```

---

## 五、实际业务逻辑

### 5.1 callback_submit 的作用

以 `AlipayFacePluginResponder.callback_submit` 为例：

```python
def callback_submit(self, product_id: int, create_datetime: datetime, 
                   plugin_upstream, order_no, tenant_id, channel_id, 
                   money, **kwargs):
    product = AlipayProduct.objects.get(id=product_id)
    self.set_model_value(Order, {"order_no": order_no}, {"remarks": product.name})
    channel = PayChannel.objects.filter(id=channel_id).first()
    
    if channel and channel.extra_arg == 1:  # 公池模式
        pool = AlipayPublicPool.objects.filter(alipay_id=product_id).first()
        if pool:
            submit_base_day_statistics(AlipayPublicPoolDayStatistics,
                                       pool_id=pool.id, 
                                       date=create_datetime.date(),
                                       pay_channel_id=channel_id)
    else:
        if tenant_id != product.writeoff.parent_id:  # 神码模式
            shenma = AlipayShenma.objects.filter(alipay_id=product_id, tenant=tenant_id).first()
            if shenma:
                submit_base_day_statistics(AlipayShenmaDayStatistics,
                                           shenma_id=shenma.id, 
                                           date=create_datetime.date(),
                                           pay_channel_id=channel_id)
        else:  # 普通模式
            submit_base_day_statistics(AlipayProductDayStatistics,
                                       product_id=product_id, 
                                       date=create_datetime.date(),
                                       pay_channel_id=channel_id)
```

**主要功能**:

1. ✅ 更新订单备注（产品名称）
2. ✅ 根据业务模式更新日统计：
   - 公池模式 → `AlipayPublicPoolDayStatistics`
   - 神码模式 → `AlipayShenmaDayStatistics`
   - 普通模式 → `AlipayProductDayStatistics`
3. ✅ 记录下单次数和金额（用于限额控制）

---

## 六、关键点总结

### 6.1 调用方式

- **异步调用**: 通过 `scheduler.add_job` 异步执行，不阻塞主流程
- **延迟执行**: 订单创建成功后延迟 500 微秒执行，确保数据一致性
- **钩子模式**: 使用装饰器模式注册钩子函数，便于扩展

### 6.2 调用时机

1. **订单创建成功**: 延迟 500 微秒后执行
2. **订单创建失败**: 立即执行（用于统计失败订单）

### 6.3 执行流程

1. `notify_order_submit` 被调度器调用
2. 遍历所有注册的钩子函数
3. `callback_plugin_submit` 被调用
4. 根据插件类型获取插件实例
5. 调用插件的 `callback_submit` 方法
6. 更新日统计等业务逻辑

### 6.4 注意事项

- ⚠️ 如果插件配置了 `ignore_no_url=True`，则不会触发回调
- ⚠️ 所有钩子函数的异常都会被捕获，不会影响其他钩子执行
- ⚠️ 回调是异步执行的，不能依赖其执行结果

---

## 七、相关文件

- `create_order_hook.py`: 钩子函数定义和注册机制
- `views/order.py`: 订单创建视图，调用 `notify_order_submit`
- `base_plugin.py`: 插件基类，定义 `callback_submit` 接口
- `alipay_plugin.py`: 支付宝插件实现，实现具体的 `callback_submit` 逻辑
