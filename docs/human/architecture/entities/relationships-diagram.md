# 实体关系图

> [待补充 by @开发者]

---

## 概述

本文档应包含TTPOS系统的完整实体关系图（ERD）。

---

## 待补充内容

### 核心实体关系

#### 订单系统
- SaleBill ← SaleOrder
- SaleOrder ← MemberSaleOrder
- SaleOrder ← H5Order
- PaymentOrder → SaleBill

#### 会员系统
- Member → MemberSaleOrder
- Company → Member

#### 产品系统
- Product → ProductCategory
- Product → ProductPackage
- Warehouse → Product

#### 门店系统
- Company → Desk
- Company → Staff
- Desk → SaleBill

### 建议工具

- draw.io
- Lucidchart
- PlantUML
- Mermaid

### 参考文档

- [base_model.md](./base_model.md) - 基础模型
- [relationships.md](./relationships.md) - 实体关系说明
- 各实体文档中的关系说明

---

**最后更新**: 2025-11-16

