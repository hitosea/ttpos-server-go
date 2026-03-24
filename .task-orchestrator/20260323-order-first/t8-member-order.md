# T8: 会员端"先下单后付"模式下单流程

## 修改文件
- `main/app/service/order_member.go`

## 变更概要

### 1. CreateMemberDineInOrder - 新增"先下单后付"分支
- 在创建/更新订单成功后，获取 `store_scan_order` 配置
- 如果 `IsOrderFirstPayLater == 1`，调用 `createH5OrderForOrderFirst` 创建 H5 订单
- 获取配置失败时回退到默认模式（先付后下单），不影响正常流程

### 2. createH5OrderForOrderFirst - 新增方法
复用支付完成事件 `createH5OrderForMemberDineIn` 的核心逻辑，在一个事务中完成：
- 创建 H5 Order（status=1/待接单，order_type=1/会员端堂食）
- 创建 H5 Order Product 快照
- 更新 sale_order_product 的 h5_order_uuid 和 is_accept_order
- 设置 SaleBill.submit_pay_time 使订单在列表中可见

与支付完成事件处理器的区别：
- 不标记订单为已完成（SaleBill.Status 保持 Pending）
- 不执行自动接单

### 3. getMemberDineInOrderStatusInfo - 状态判断扩展
`SaleBillStatusPending` 分支新增 H5 订单判断：
- 有 H5 订单（先下单后付模式）：根据 H5 订单状态显示待接单/备餐中/已完成/已拒单
- 无 H5 订单（普通模式）：显示待支付（原逻辑不变）

### 4. GetMemberDineInOrderList - 列表过滤兼容
- "进行中"查询：SQL 条件增加 `SaleBillStatusPending`，内存过滤保留有 H5 订单的 Pending 订单
- "待支付"查询：内存过滤排除有 H5 订单的 Pending 订单（先下单后付的不是待支付）
- Pending + H5 订单的进行中过滤：H5 订单必须是待接单或已接单，且生产单未完成

### 5. GetMemberDineInOrderDetail - 详情页兼容
- "先下单后付"订单不显示支付倒计时（remainingPaymentTime = 0）
- "先下单后付"订单不返回支付方式列表（无需立即支付）

### 6. CancelMemberDineInOrder - 取消逻辑扩展
- 检测是否有关联的 H5 订单
- 有 H5 订单（先下单后付模式）：只有待接单状态可取消，在事务中同时取消订单和软删除 H5 订单
- 无 H5 订单（普通模式）：保持原逻辑不变

## 额外修复
- 修复了文件中存在的 Unicode 弯引号（curly quotes）替换为标准 ASCII 双引号

## 未修改的文件（按任务约束）
- `main/app/constant/order.go` - 已在 t7 完成
- `main/app/service/order_h5.go` - 由 t9 负责
