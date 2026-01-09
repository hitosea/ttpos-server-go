# 总部-分店颗粒化同步（后端）Spec

> 为前端"总部-分店颗粒化同步"功能提供后端支持

---

## 📁 文档结构

```
shop-headquarters-branch-granular-sync-backend/
├── README.md                          # 📖 本文档：Spec概览和导航
├── requirements.md                    # 📋 需求文档：功能需求和验收标准
├── design.md                          # 🏗️ 设计文档：技术方案和实现细节
├── tasks.md                           # ✅ 任务分解：36个开发任务（24h）
├── PAYMENT_METHOD_SYNC_RULES.md       # ⚠️ 支付方式同步特殊规则（必读）
└── RELATED_DATA_GUIDE.md              # 🔗 关联数据获取指南（各数据类型的关联关系）
```

---

## 🎯 快速导航

### 我要...

| 场景 | 查阅文档 |
|------|---------|
| 了解需求和业务价值 | `requirements.md` |
| 理解技术方案和架构 | `design.md` |
| 开始编码实现 | `tasks.md` |
| 实现支付方式同步 | `PAYMENT_METHOD_SYNC_RULES.md` ⚠️ |
| 了解各数据类型的关联关系 | `RELATED_DATA_GUIDE.md` 🔗 |
| 查看API接口定义 | `design.md` → "API设计"章节 |
| 查看数据库设计 | `design.md` → "数据库设计"章节 |
| 查看DTO定义 | `design.md` → "数据模型"章节 |
| 查看测试用例 | `PAYMENT_METHOD_SYNC_RULES.md` → "测试用例"章节 |

---

## 📋 功能概览

### 核心功能

1. **新增5种同步数据类型**：
   - ✅ 优惠券（`marketing_coupon`）
   - ✅ 满额减（`full_reduction_activity`）- 有多语言
   - ✅ 菜品标签（`product_label`）
   - ✅ 营销活动（`marketing_activity`）- 有多语言
   - ⚠️ 支付方式（`payment_method`）- **复杂规则**

2. **2个新API接口**：
   - `POST /api/v1/shop/sync/headquarters_data_list` - 获取总部可同步数据列表
   - `POST /api/v1/shop/sync/granular_sync` - 颗粒化同步数据

### 特性亮点

- ✅ **关联数据识别**：返回关联数据的类型和uuid列表（如菜品标签关联的商品）
- ✅ **已同步状态**：返回分店已同步的uuid列表，前端根据此列表默认勾选
- ✅ **删除未勾选**：同步时自动删除分店中未勾选的总部数据（支付方式除外）
- ⚠️ **支付方式特殊规则**：过滤、同名判断、特殊code处理等

---

## ⚠️ 重要提醒

### 支付方式同步规则（必读）

支付方式同步与其他数据类型有显著不同，具有复杂的业务规则：

1. **获取列表时**：过滤 `code = 40` 和 `code = 10`
2. **删除策略**：**不删除**未勾选的总部数据
3. **同名判断**：基于 `payment_name` 字段
4. **特殊code**：90111, 90222, 90333 只更新 `headquarter_uuid`
5. **新增规则**：code自动生成，logo_file_uuid=0，其他字段用默认值

📖 **详细规则**：必须阅读 `PAYMENT_METHOD_SYNC_RULES.md`

---

## 📊 任务统计

- **总任务数**：36个
- **预估工时**：23.5小时
- **7个Phase**：数据库 → 常量 → Repository → Service → API → 测试 → 部署

### 关键任务

| 任务 | 难度 | 工时 | 重要性 |
|------|------|------|--------|
| Task 1.2: 确认支付方式表结构 | 中 | 1h | ⚠️ 关键 |
| Task 4.7: getPaymentMethodGroup | 中 | 0.5h | 重要 |
| Task 4.12: SyncPaymentMethodByUuids | **高** | 2.5h | ⚠️ 最复杂 |
| Task 6.5: 支付方式规则测试 | 高 | 1.5h | ⚠️ 必做 |

---

## 🚀 开始开发

### Step 1: 阅读文档

```bash
# 1. 了解需求
cat requirements.md

# 2. 理解支付方式规则（⚠️ 必读）
cat PAYMENT_METHOD_SYNC_RULES.md

# 3. 查看技术方案
cat design.md

# 4. 查看任务清单
cat tasks.md
```

### Step 2: 执行任务

按照 `tasks.md` 中的任务顺序依次执行：

1. **Phase 1**: 数据库迁移（添加 `headquarter_uuid` 字段）
2. **Phase 2**: 定义常量
3. **Phase 3**: 扩展 Repository
4. **Phase 4**: 实现 Service 层（⚠️ 支付方式最复杂）
5. **Phase 5**: 实现 API 层
6. **Phase 6**: 编写测试（⚠️ 支付方式规则测试必做）
7. **Phase 7**: 部署和联调

### Step 3: 测试验证

```bash
# 运行单元测试
cd main
go test ./app/service/... -v

# API测试
curl -X POST http://localhost:8080/api/v1/shop/sync/headquarters_data_list

# 集成测试（按 tasks.md 中的测试用例）
```

---

## 🔗 关联数据快速参考

| 数据类型 | 关联类型 | 获取方式 | 详细说明 |
|---------|---------|---------|---------|
| **material**（物品） | unit, material_category | 直接字段 + Preload("NotBaseUnitList") | RELATED_DATA_GUIDE.md |
| **product**（商品） | unit, category, flavor, sauce, attribute, bom_card | 直接字段 + Preload("Bom", "PackageAttributes") | RELATED_DATA_GUIDE.md |
| **bom_card**（成本卡） | material, unit | Preload("RelatedMaterialList") | RELATED_DATA_GUIDE.md |
| **product_label**（菜品标签） | product | Preload("ProductPackages") | design.md + RELATED_DATA_GUIDE.md |
| **coupon**（优惠券） | 无 | - | - |
| **full_reduction**（满额减） | 无 | - | - |
| **marketing_activity**（营销活动） | 无 | - | - |
| **payment_method**（支付方式） | 无 | - | PAYMENT_METHOD_SYNC_RULES.md |

---

## 📝 变更日志

### v1.2.0 (2025-12-05)

- ✅ 新增关联数据获取指南：`RELATED_DATA_GUIDE.md`
- ✅ 补充物品关联单位的获取方式
- ✅ 明确支付方式已同步状态通过名称匹配

### v1.1.0 (2025-12-05)

- ✅ 优化数据结构：去掉 `IsSynced`、`TotalCount`、`SyncedCount`
- ✅ 改进关联数据：使用 `RelatedData` 明确关联类型
- ✅ 新增支付方式规则文档：`PAYMENT_METHOD_SYNC_RULES.md`

### v1.0.0 (2025-12-05)

- ✅ 初始版本：需求、设计、任务分解

---

## 🔗 相关资源

- **DooTask 任务**: #37462
- **前端仓库 Spec**: `/home/coder/workspaces/ttpos-flutter/docs/shared/specs/active/shop-headquarters-branch-granular-sync/`
- **现有同步服务**: `main/app/service/sync.go`
- **商品关联关系**: `product_package商品关联表.txt`

---

**版本**: v1.2.0  
**创建日期**: 2025-12-05  
**更新日期**: 2025-12-05  
**维护者**: 曾振华  
**关联任务**: DooTask #37462
