# Shop 端分批送厨功能增强 需求提案

> 本文档用于明确 Shop 端（商家后台）在分批送厨功能中的具体实现需求，基于 DooTask 任务 #36921。

---

## 📋 提案信息

| 项目       | 内容           |
| ---------- | -------------- |
| **提案人** | xiezhihuan     |
| **日期**   | 2025-11-20     |
| **目标版本** | v2.10.0       |
| **状态**   | 已创建 Spec   |
| **关联 Spec** | [story-shop-batch-tag-enhancement](../../shared/specs/story-shop-batch-tag-enhancement/) |
| **关联 Dootask** | #36921 - 新管理端/点餐助手/收银端-分批送厨功能 |

---

## 🎯 背景和动机

### 问题描述

当前分批送厨功能中，Shop 端（商家后台）的分批类型管理存在以下问题：

1. **缺少名称缩写字段**
   - 收银端和点餐助手需要在界面上显示分批类型的缩写
   - 当前没有专门的缩写字段，无法满足前端展示需求
   - 影响用户体验和界面美观度

2. **接口需要适配新字段**
   - 创建、编辑、详情接口需要支持缩写字段
   - 需要保证向后兼容性

**注意**：分批类型的多语言支持已在 v2.9.0 版本中实现，本次无需再次实现。

### 业务价值

解决这个问题能带来以下业务价值：

- **用户体验提升**：通过缩写字段优化收银端和点餐助手的界面展示
- **功能完善**：补齐分批类型管理功能的最后一块拼图
- **扩展性增强**：为未来功能扩展提供更好的基础

### 目标用户

- [x] **商户管理员** - 在 Shop 端管理分批类型
- [ ] 收银员
- [ ] 厨房人员
- [ ] 顾客

---

## 💡 解决方案概述

### 方案描述

在 Shop 端的分批类型管理功能中，增加以下能力：

1. **名称缩写字段**
   - 在分批类型中增加"名称缩写"字段（必填）
   - 用于收银端和点餐助手界面展示
   - 支持在创建和编辑时设置

2. **接口适配**
   - 修改分批类型的创建接口，支持缩写字段
   - 修改分批类型的编辑接口，支持缩写字段
   - 修改分批类型的详情接口，返回缩写字段

**注意**：多语言名称功能已在 v2.9.0 版本中实现，本次无需修改。

### 核心功能点

1. **数据库模型调整**
   - `batch_tag` 表增加 `abbreviation` 字段（名称缩写）

2. **API 接口调整**
   - 创建分批类型接口：支持缩写字段
   - 编辑分批类型接口：支持缩写字段
   - 获取分批类型详情接口：返回缩写字段

3. **数据验证**
   - 缩写字段必填验证
   - 缩写字段长度限制（1-10 个字符）

### 影响范围

**涉及终端**：
- [x] **Shop 商家管理端** - 分批类型管理功能
- [ ] POS 收银端 - 仅使用，不涉及管理
- [ ] Assistant 助手端 - 仅使用，不涉及管理
- [ ] 新管理端 - 仅使用，不涉及管理

**涉及模块**：
- [x] **UI 组件** - Shop 端分批类型管理页面
- [x] **API 接口** - 分批类型 CRUD 接口
- [x] **数据模型** - `batch_tag` 表结构
- [x] **业务逻辑** - 多语言名称处理、缩写验证
- [ ] 第三方集成
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **低**：基础的数据字段增加和接口调整，业务逻辑简单
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预计天数**: 1.5-2 天
- **预估 SP**: 2 SP（待技术评审确认）

**工作量拆分**：
- 数据库迁移脚本：0.5 天
- 后端 Model 和 DTO 调整：0.5 天
- Service 和 API 接口调整（创建/编辑/详情/列表）：0.5 天
- 测试和修复：0.5 天

### 风险识别

**潜在风险**：
1. **数据迁移风险**
   - 现有分批类型数据需要设置默认缩写
   - 需要为现有数据生成合理的默认值

2. **向后兼容性**
   - 需要保证现有接口的向后兼容
   - 前端可能需要适配新的数据结构

3. **缩写字段唯一性**
   - 是否需要保证缩写字段的唯一性
   - 如何处理重复的缩写

**缓解措施**：
1. **数据迁移策略**
   - 编写数据迁移脚本，为现有分批类型设置默认缩写
   - 从多语言名称中提取中文名称作为默认缩写，或使用名称的前几个字符

2. **接口兼容性**
   - 保持现有接口结构，新增字段作为必填字段
   - 提供数据转换层，确保旧版本前端仍可使用（但需要前端更新）

3. **缩写唯一性**
   - 建议不强制唯一性，允许商户自定义
   - 在 UI 层面提示用户建议使用唯一缩写

---

## 🔗 相关资源

### 参考需求

- DooTask 任务: #36921 - 新管理端/点餐助手/收银端-分批送厨功能
- 相关功能: 商品多语言名称、其他多语言功能

### 相关文档

- 数据库规范: `.cursor/rules/database.mdc`
- API 设计规范: `.cursor/rules/api.mdc`
- 多语言实现规范: 参考 `multi_language_name` 表的使用方式

### 技术实现参考

**现有代码结构**：
- Model: `main/app/model/product.go` - `BatchTag` 结构体
- Repository: `main/app/repository/batch_tag.go`
- DTO: `main/app/dto/resp/product_resp/product.go` - `BatchTag` 响应结构
- 数据库表: `ttpos_batch_tag`

**需要修改的文件**：
1. 数据库迁移脚本：`admin/database/migrations/` - 新增迁移文件（增加 abbreviation 字段）
2. Model: `main/app/model/product.go` - 在 BatchTag 结构体中增加 `Abbreviation` 字段
3. DTO: `main/app/dto/req/product.go` - 在 BatchTagAddReq 和 BatchTagEditReq 中增加 `Abbreviation` 字段
4. DTO: `main/app/dto/resp/product_resp/product.go` - 在 BatchTag 和 BatchTagDetail 中增加 `Abbreviation` 字段
5. Service: `main/app/service/product.go` - 在 AddBatchTag、EditBatchTag、GetBatchTag 方法中处理缩写字段
6. API Controller: `main/app/api/v1/shop/shop_batch_product.go` - 在 API 中增加缩写字段的参数验证

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名       | 签名/日期 |
| ------------ | ---------- | --------- |
| 产品经理     | 待定       |           |
| 技术负责人   | 待定       |           |
| 开发代表     | xiezhihuan |           |
| 测试代表     | 待定       |           |
| UI/UX 设计师 | 待定       |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[待评审会议后填写]
```

**下一步行动**：

- [x] 创建 Spec：`story-shop-batch-tag-enhancement` ✅
- [ ] 分配负责人：待定
- [ ] 目标 Sprint：待定

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 在 Shop 端管理分批类型时，能够设置名称缩写  
**以便于** 优化收银端和点餐助手的界面展示

### AC 验收标准（初稿）

1. **WHEN** 商户管理员创建分批类型 **THEN** 系统 **SHALL** 要求填写缩写字段
2. **WHEN** 商户管理员编辑分批类型 **THEN** 系统 **SHALL** 允许修改缩写字段
3. **WHEN** 商户管理员查看分批类型详情 **THEN** 系统 **SHALL** 显示缩写信息
4. **IF** 商户管理员未填写缩写字段 **THEN** 系统 **SHALL** 提示必填并阻止提交
5. **IF** 商户管理员填写的缩写超过长度限制（10个字符） **THEN** 系统 **SHALL** 提示错误并阻止提交

### 技术实现细节

#### 1. 数据库变更

**表结构调整**：
```sql
ALTER TABLE `ttpos_batch_tag` 
ADD COLUMN `abbreviation` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '名称缩写' AFTER `multi_language_name_uuid`;
```

**数据迁移**：
- 为现有数据设置默认缩写（从多语言名称中提取中文名称，或使用名称的前几个字符）

#### 2. Model 调整

```go
// BatchTag 分批类型表
type BatchTag struct {
    BaseModel
    Name                  string `gorm:"default:'';column:name;comment:'名称'"`
    MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称UUID'"`
    Abbreviation          string `gorm:"default:'';column:abbreviation;comment:'名称缩写'"`
    Color                 string `gorm:"default:'';column:color;comment:'颜色值，如#FF0000'"`
    Sort                  int    `gorm:"default:0;column:sort;comment:'排序(数字越小越靠前)';NOT NULL"`
    
    MultiLanguageName *MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}
```

#### 3. API 接口调整

**创建分批类型接口**：
- 请求参数增加 `abbreviation` 字段（必填）

**编辑分批类型接口**：
- 请求参数增加 `abbreviation` 字段（必填）

**获取分批类型详情接口**：
- 响应中返回 `abbreviation` 字段

#### 4. 验证规则

- `abbreviation` 字段：必填，长度 1-10 个字符

### 线框图/原型（可选）

[可附加 Shop 端分批类型管理页面的 UI 设计图]

---

## 📄 模板使用说明

### 与任务 #36921 的关系

本提案是任务 #36921（新管理端/点餐助手/收银端-分批送厨功能）中 Shop 端功能的具体实现方案。

**任务 #36921 的整体功能包括**：
- 新管理端：商品关联模式、类型管理
- 点餐助手：前置/后置关联模式
- 收银机：前置/后置关联模式
- **Shop 端：分批类型管理增强（本提案）**

**本提案聚焦于**：
- Shop 端的分批类型管理功能
- 名称缩写字段（多语言名称已在 v2.9.0 实现）

---

**版本**: v1.0.0  
**创建日期**: 2025-11-20  
**维护者**: xiezhihuan  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`, `.cursor/rules/database.mdc`

