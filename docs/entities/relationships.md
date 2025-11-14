# 实体关联关系图

## 核心实体关系

### 1. 销售流程核心实体

```mermaid
erDiagram
    SaleBill ||--o{ SaleOrder : 包含
    SaleOrder ||--o{ SaleOrderProduct : 包含
    SaleOrder ||--o{ PaymentOrder : 支付
    SaleOrder ||--o{ Member : 会员
    SaleOrder ||--o{ Desk : 桌台
    SaleBill ||--o{ Staff : 收银员
    SaleBill ||--o{ Company : 公司
    
    SaleOrderProduct ||--o{ Product : 商品
    SaleOrderProduct ||--o{ ProductPackage : 套餐
    SaleOrderProduct ||--o{ SaleOrderProductBom : BOM
    SaleOrderProduct ||--o{ SaleOrderProductAttribute : 属性
    
    Product ||--o{ ProductCategory : 分类
    Product ||--o{ ProductBom : BOM
    Product ||--o{ ProductPackage : 套餐
```

### 2. 会员体系实体

```mermaid
erDiagram
    Member ||--o{ MemberLevel : 等级
    Member ||--o{ MemberCard : 会员卡
    Member ||--o{ MemberBalanceLog : 余额日志
    Member ||--o{ MemberPointLog : 积分日志
    Member ||--o{ MemberRechargeOrder : 充值订单
    Member ||--o{ MemberSaleOrder : 外卖订单
    
    MemberLevel ||--o{ Member : 拥有
    MemberCard ||--o{ MemberCardType : 卡类型
    MemberRechargeOrder ||--o{ PaymentOrder : 支付
    MemberRechargeOrder ||--o{ Staff : 操作员
```

### 3. 支付体系实体

```mermaid
erDiagram
    PaymentOrder ||--o{ PaymentMethod : 支付方式
    PaymentOrder ||--o{ RefundOrder : 退款
    PaymentOrder ||--o{ ReturnOrderAmount : 退款金额
    
    PaymentMethod ||--o{ File : 图片
    PaymentMethod ||--o{ PaymentOrder : 使用
```

### 4. 仓储物流实体

```mermaid
erDiagram
    Warehouse ||--o{ WarehouseItem : 商品
    Warehouse ||--o{ Company : 公司
    Warehouse ||--o{ MultiLanguageName : 多语言
    
    WarehouseItem ||--o{ Product : 商品
    WarehouseItem ||--o{ Warehouse : 仓库
```

### 5. 员工权限实体

```mermaid
erDiagram
    Staff ||--o{ Company : 公司
    Staff ||--o{ Device : 设备
    Staff ||--o{ StaffRole : 角色
    Staff ||--o{ StaffShiftLog : 交班
    Staff ||--o{ StaffLoginLog : 登录
    Staff ||--o{ StaffOperationLog : 操作
    
    StaffRole ||--o{ Role : 角色
    Role ||--o{ Staff : 拥有
```

### 6. 设备管理实体

```mermaid
erDiagram
    Device ||--o{ Staff : 绑定
    Device ||--o{ Printer : 打印机
    Device ||--o{ Company : 公司
```

### 7. 打印系统实体

```mermaid
erDiagram
    PrinterLog ||--o{ Printer : 打印机
    PrinterLog ||--o{ Company : 公司
    PrinterLog ||--o{ Staff : 操作员
    
    Printer ||--o{ Device : 关联
    Printer ||--o{ PrinterTemplate : 模板
    Printer ||--o{ PrinterCustomize : 自定义
```

### 8. H5扫码点餐实体

```mermaid
erDiagram
    H5Order ||--o{ Member : 会员
    H5Order ||--o{ Desk : 桌台
    H5Order ||--o{ Company : 公司
    H5Order ||--o{ H5OrderProduct : 商品
    H5Order ||--o{ PaymentOrder : 支付
    
    H5OrderProduct ||--o{ Product : 商品
```

### 9. 会员外卖实体

```mermaid
erDiagram
    MemberSaleOrder ||--o{ Member : 会员
    MemberSaleOrder ||--o{ Staff : 配送员
    MemberSaleOrder ||--o{ Desk : 桌台
    MemberSaleOrder ||--o{ Company : 公司
    MemberSaleOrder ||--o{ MemberSaleOrderProduct : 商品
    MemberSaleOrder ||--o{ PaymentOrder : 支付
    
    MemberSaleOrderProduct ||--o{ Product : 商品
```

## 主要关联关系说明

### 1. 销售流程关系
- **SaleBill** 是销售账单的核心，包含多个 **SaleOrder**
- **SaleOrder** 关联商品、会员、支付、收银员等信息
- **SaleOrderProduct** 记录订单中的具体商品，支持BOM和属性
- 支持套餐、规格、加料等复杂商品结构

### 2. 会员体系关系
- **Member** 是会员核心，关联等级、卡片、余额、积分等
- 支持多级会员等级和会员卡类型
- 完整的充值、消费、退款流程记录
- 积分和余额分离管理，支持冻结机制

### 3. 支付体系关系
- **PaymentOrder** 是支付核心，支持多种支付方式
- **PaymentMethod** 定义支付方式，支持手续费计算
- 完整的退款和反结账流程
- 支持第三方支付和内部支付

### 4. 多租户架构
- **Company** 是多租户核心，大部分实体都关联公司
- 支持总部-分店架构
- 统一的用户权限和设备管理

### 5. 实时通信
- 大部分实体继承 **BaseModel**，支持WebSocket推送
- 订单状态变更实时通知
- 设备状态实时同步

## 数据流向总结

### 销售数据流
1. 开台 → 创建SaleBill
2. 点餐 → 创建SaleOrder和SaleOrderProduct
3. 结账 → 创建PaymentOrder
4. 支付 → 更新订单状态
5. 打印 → 创建PrinterLog

### 会员数据流
1. 注册 → 创建Member
2. 充值 → 创建MemberRechargeOrder和PaymentOrder
3. 消费 → 扣减余额/积分，创建日志
4. 升级 → 更新MemberLevel

### 库存数据流
1. 入库 → 更新WarehouseItem
2. 调拨 → 创建TransferOrder
3. 盘点 → 创建StockReconciliation
4. 出库 → 更新库存数量

这个实体关系图展现了TTPOS系统的完整业务架构，涵盖了销售、会员、支付、仓储、员工、设备等各个模块的关联关系。