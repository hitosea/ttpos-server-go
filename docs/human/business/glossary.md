# 业务术语表

> 👤 **受众**: 所有开发者  
> 📖 **用途**: 统一业务术语和技术术语的理解

---

## 餐饮业务术语

### 基础概念

| 术语 | 英文 | 说明 | 示例 |
|------|------|------|------|
| 商户 | Company | 使用系统的餐饮企业 | 某某餐饮公司 |
| 门店 | Shop | 商户下的具体店铺 | XX店、XX分店 |
| 桌台 | Desk/Table | 餐厅的就餐桌子 | 1号桌、A区2号桌 |
| 开台 | Open Desk | 顾客入座，开始点餐 | - |
| 清台 | Clean Desk | 顾客离开，清理桌台 | - |

---

### 订单相关

| 术语 | 英文 | 说明 | 示例 |
|------|------|------|------|
| 订单 | Order | 顾客的点餐记录 | 订单号：202311170001 |
| 销售单 | Sale Bill | 订单的财务单据 | - |
| 订单明细 | Order Item | 订单中的每个商品 | 宫保鸡丁 x2 |
| 加菜 | Add Dish | 订单创建后追加商品 | - |
| 退菜 | Return Dish | 取消订单中的某个商品 | - |
| 沽清 | Sold Out | 商品售罄，暂时下架 | 今日沽清：鱼香肉丝 |

---

### 支付相关

| 术语 | 英文 | 说明 | 示例 |
|------|------|------|------|
| 结账 | Checkout | 顾客付款完成订单 | - |
| 挂账 | Pending Payment | 记录消费，稍后付款 | 会员挂账 |
| 买单 | Pay Bill | 付款 | - |
| 退款 | Refund | 退还已支付金额 | - |
| 抹零 | Round Down | 抹去零头 | 128.5元 → 128元 |
| 折扣 | Discount | 价格折扣 | 8折、9折 |

---

### 会员相关

| 术语 | 英文 | 说明 | 示例 |
|------|------|------|------|
| 会员 | Member | 注册的顾客 | 会员卡号：10001 |
| 会员等级 | Member Level | 会员的级别 | 普通会员、VIP会员 |
| 储值卡 | Stored Value Card | 预充值卡 | 余额：500元 |
| 积分 | Points | 会员积分 | 消费累积积分 |
| 充值 | Recharge | 会员卡充值 | 充值300元 |

---

### 商品相关

| 术语 | 英文 | 说明 | 示例 |
|------|------|------|------|
| 商品 | Product | 菜品、饮料等 | 宫保鸡丁 |
| 菜品 | Dish | 餐饮商品 | 鱼香肉丝 |
| 套餐 | Combo/Set Meal | 组合套餐 | 双人套餐 |
| 商品分类 | Category | 商品的分类 | 川菜、粤菜 |
| 规格 | Specification | 商品的规格 | 大份、中份、小份 |
| 做法 | Cooking Method | 商品的做法 | 不辣、微辣、特辣 |

---

### 库存相关

| 术语 | 英文 | 说明 | 示例 |
|------|------|------|------|
| 原料 | Material | 制作菜品的原材料 | 鸡肉、辣椒 |
| 半成品 | Semi-finished | 预加工的原料 | 腌制好的肉 |
| 成品 | Finished Product | 可售卖的商品 | 成品菜 |
| 入库 | Stock In | 采购物品入库 | 入库50kg鸡肉 |
| 出库 | Stock Out | 物品出库使用 | 出库5kg鸡肉 |
| 盘点 | Inventory Count | 核对库存数量 | 月末盘点 |
| 报损 | Loss Report | 记录损耗 | 蔬菜变质报损 |

---

## 技术术语

### 系统架构

| 术语 | 说明 | 使用场景 |
|------|------|----------|
| 多租户 | 每个商户独立数据库 | 数据隔离 |
| 微服务 | ttpos-bmp 微服务群 | 服务拆分 |
| 单体应用 | main/admin 模块 | 核心业务 |
| DBM | Database Manager | 数据库管理器 |
| gRPC | 服务间通信协议 | 内部调用 |

---

### 数据库

| 术语 | 说明 | 示例 |
|------|------|------|
| UUID | 唯一标识符（bigint） | 用户UUID、订单UUID |
| 软删除 | 逻辑删除（delete_time） | delete_time = 1699999999 |
| 主从 | 数据库主从复制 | Master-Slave |
| 索引 | 数据库索引 | 主键索引、唯一索引 |
| 迁移 | 数据库版本管理 | migration |

---

### API 相关

| 术语 | 说明 | 示例 |
|------|------|------|
| RESTful | API 设计风格 | GET/POST/PUT/DELETE |
| snake_case | 蛇形命名 | order_list, user_info |
| camelCase | 驼峰命名 | orderList, userInfo |
| JWT | JSON Web Token | 用户认证 |
| Bearer Token | Token 前缀 | Authorization: Bearer xxx |

---

### 前端术语

| 术语 | 说明 | 使用场景 |
|------|------|----------|
| 收银端 | 收银员使用的终端 | Flutter 开发 |
| 管理后台 | 管理员使用的后台 | Vue3 + admin模块 |
| 店铺后台 | 店长使用的后台 | Vue3 + shop模块 |
| 组合式API | Vue3 Composition API | script setup |
| Pinia | Vue3 状态管理 | 替代 Vuex |

---

## 常用缩写

| 缩写 | 全称 | 说明 |
|------|------|------|
| TTPOS | TongTong POS | 系统名称 |
| POS | Point of Sale | 收银系统 |
| ERP | Enterprise Resource Planning | 企业资源计划 |
| BMP | Business Middle Platform | 业务中台 |
| CRM | Customer Relationship Management | 客户关系管理 |
| API | Application Programming Interface | 应用程序接口 |
| DTO | Data Transfer Object | 数据传输对象 |
| DAO | Data Access Object | 数据访问对象 |
| ORM | Object-Relational Mapping | 对象关系映射 |
| HTTP | Hypertext Transfer Protocol | 超文本传输协议 |
| CRUD | Create, Read, Update, Delete | 增删改查 |

---

## 状态码说明

### 订单状态

| 状态码 | 状态名称 | 说明 |
|--------|----------|------|
| 0 | 待支付 | Pending |
| 1 | 已支付 | Paid |
| 2 | 已完成 | Completed |
| 3 | 已取消 | Cancelled |
| 4 | 已退款 | Refunded |

### 会员状态

| 状态码 | 状态名称 | 说明 |
|--------|----------|------|
| 0 | 禁用 | Disabled |
| 1 | 正常 | Active |
| 2 | 冻结 | Frozen |

---

## 相关文档

- [业务流程 - 订单流程](./workflows/order-flow.md) - 订单完整流程
- [业务流程 - 支付流程](./workflows/payment-flow.md) - 支付完整流程
- [业务流程 - 库存流程](./workflows/inventory-flow.md) - 库存管理流程

---

**最后更新**: 2025-11-17  
**维护者**: TTPOS Team  
**版本**: v1.0

