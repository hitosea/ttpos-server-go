# 获取单品备注预设选项列表 API 需求文档

> 本文档定义获取单品备注预设选项列表 API 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/item-remark-preset.md](../../../../team/proposals/2025-12/item-remark-preset.md) |
| **创建日期**      | 2025-12-05                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | 2025-12-05             |
| **审核意见** | 产品审核通过，可进入技术设计阶段         |

---

## 📋 概述

为收银机、点餐助手、H5、平板等终端提供获取单品备注预设选项列表的 API 接口，使各终端在点餐时能够获取预设的单品备注选项，供用户快速选择。

**功能范围**：
- ✅ 仅实现获取列表 API（GET 接口）
- ❌ 不包含预设选项的增删改功能（后续实现）
- ❌ 不包含选择预设选项后保存到订单的功能（后续实现）

**参考实现**：整单备注列表 API（`/h5/order/remark/list`、`/cashier/instant/order/remark/list` 等）

## 🎯 产品对齐

支持 Barrio 客户的需求：在点餐时能够快速选择预设的单品备注选项，提高点餐效率。本阶段先实现获取列表 API，为后续的预设选项选择功能打下基础。

## 📝 用户故事

**作为** 收银员/点餐助手/H5/平板用户  
**我想** 在点餐时获取可用的单品备注预设选项列表  
**以便于** 了解有哪些预设选项可供选择（为后续的选择功能做准备）

---

## 功能需求

### Requirement 1: 各终端获取单品备注预设选项列表 API

**用户故事**: 作为收银员/点餐助手/H5/平板用户，我想获取单品备注预设选项列表，以便于了解可用的预设选项

#### 验收标准

1. **WHEN** 收银机点餐端调用获取单品备注列表 API **THEN** 系统 **SHALL** 返回当前门店的单品备注预设选项列表（排除已删除的选项）

2. **WHEN** 点餐助手端调用获取单品备注列表 API **THEN** 系统 **SHALL** 返回当前门店的单品备注预设选项列表（排除已删除的选项）

3. **WHEN** H5 端调用获取单品备注列表 API **THEN** 系统 **SHALL** 返回当前门店的单品备注预设选项列表（排除已删除的选项）

4. **WHEN** 平板端调用获取单品备注列表 API **THEN** 系统 **SHALL** 返回当前门店的单品备注预设选项列表（排除已删除的选项）

5. **IF** 门店没有设置任何单品备注预设选项 **THEN** 系统 **SHALL** 返回空列表 `[]`

6. **IF** 单品备注预设选项包含多语言名称 **THEN** 系统 **SHALL** 返回完整的多语言名称信息

#### 具体要求

- [ ] 1.1 在收银机点餐端添加 API：`GET /cashier/instant/order/item/remark/list`
- [ ] 1.2 在点餐助手端添加 API：`GET /assistant/desk/order/item/remark/list`
- [ ] 1.3 在 H5 端添加 API：`GET /h5/order/item/remark/list`
- [ ] 1.4 在平板端添加 API：`GET /tablet/desk/order/item/remark/list`
- [ ] 1.5 所有 API 需要身份验证（JWT Token）
- [ ] 1.6 所有 API 返回格式统一：`{code, message, data: {list: []}}`
- [ ] 1.7 列表项包含：`uuid`、`locale_name`（多语言名称）
- [ ] 1.8 列表按创建时间倒序排列（最新的在前）
- [ ] 1.9 排除已逻辑删除的预设选项（`delete_time = 0`）
- [ ] 1.10 复用现有的 `otherSrv.GetOrderItemRemarkList()` Service 方法

#### API 设计

**请求**：
- Method: `GET`
- Path: 
  - `/cashier/instant/order/item/remark/list`
  - `/assistant/desk/order/item/remark/list`
  - `/h5/order/item/remark/list`
  - `/tablet/desk/order/item/remark/list`
- Headers: `Authorization: Bearer {JWT_TOKEN}`

**响应**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "uuid": 1234567890,
        "locale_name": {
          "zh": "不要香菜",
          "th": "ไม่ใส่ผักชี",
          "en": "No coriander",
          "zh_tw": "不要香菜",
          "ja": "コリアンダーなし",
          "ko": "고수 없음",
          "my": "ကြက်သွန်နီမပါ",
          "tr": "Kişniş yok",
          "sv": "Ingen koriander"
        }
      },
      {
        "uuid": 1234567891,
        "locale_name": {
          "zh": "少辣",
          "th": "เผ็ดน้อย",
          "en": "Less spicy",
          ...
        }
      }
    ]
  }
}
```

**错误响应**：
```json
{
  "code": 500,
  "message": "获取单品备注列表失败",
  "data": null
}
```

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/order/item/remark/list`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 考虑缓存策略（Redis，如果预设选项不频繁变更）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言（中文、英文、日语、韩语等）
- [x] 所有文案使用多语言实现
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证（JWT Token）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 错误日志记录（使用 Logger）
- [ ] 网络异常时优雅降级

---

## 验收标准

### 功能验收

1. **收银机点餐端 API**: 调用 `/cashier/instant/order/item/remark/list` 返回正确的单品备注列表
2. **点餐助手端 API**: 调用 `/assistant/desk/order/item/remark/list` 返回正确的单品备注列表
3. **H5 端 API**: 调用 `/h5/order/item/remark/list` 返回正确的单品备注列表
4. **平板端 API**: 调用 `/tablet/desk/order/item/remark/list` 返回正确的单品备注列表
5. **空列表处理**: 当没有预设选项时，返回空列表 `[]`
6. **多语言支持**: 返回的多语言名称包含所有支持的语言
7. **排序**: 列表按创建时间倒序排列
8. **软删除**: 已删除的预设选项不出现在列表中

### 测试验收

1. **单元测试**: Service 层测试覆盖率达标
2. **API 测试**: 所有 4 个接口测试通过
3. **集成测试**: 端到端流程测试通过（调用 API → 返回数据）

### 文档验收

1. **API 文档**: Swagger 文档完整（4 个接口）
2. **代码注释**: 所有新增代码有中文注释

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- 参考整单备注的实现方式

### 业务约束

- 仅实现获取列表功能，不涉及增删改
- 复用现有的 `GetOrderItemRemarkList` Service 方法
- 参考整单备注列表 API 的实现逻辑

### 资源约束

- 开发时间: 1-2 天
- Story Point: 2 SP

---

## 依赖关系

### 技术依赖

- `main/app/service/other.go` - `GetOrderItemRemarkList()` 方法（已存在）
- `main/app/repository/base/order_item_remark.go` - `GetOrderItemRemarkList()` 方法（已存在）
- `main/app/dto/resp/order_item_remark.go` - `OrderItemRemarkResp` 响应结构（已存在）

### 服务依赖

- 无外部服务依赖

### 业务依赖

- 依赖单品备注预设选项数据（由商家管理端设置，本 Spec 不涉及）

---

## 风险和缓解

### 风险 1: API 路径命名不一致

**影响**: 中  
**概率**: 低  
**缓解措施**:
- 参考整单备注的 API 路径命名规范
- 统一使用 `/order/item/remark/list` 格式

### 风险 2: 性能问题（如果预设选项很多）

**影响**: 低  
**概率**: 低  
**缓解措施**:
- 预设选项数量限制为 100 个（已在商家管理端限制）
- 如果后续需要，可以添加 Redis 缓存

---

## 时间表

- **Phase 1 - API 开发**: 0.5 天
  - 在 4 个终端添加获取列表 API
  - 复用现有的 Service 方法
- **Phase 2 - 测试**: 0.5 天
  - 单元测试
  - API 测试
- **总计**: 1 天（SP = 2）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 参考实现

- **整单备注列表 API**：
  - `main/app/api/v1/h5/h5_handler.go` - `OrderRemarkList()` (line 274-293)
  - `main/app/api/v1/cashier/cashier_instant.go` - `OrderRemarkList()` (line 470-489)
  - `main/app/api/v1/assistant/assistant_desk.go` - `OrderRemarkList()` (line 276-295)
  - `main/app/api/v1/tablet/tablet_desk.go` - `OrderRemarkList()` (待确认)
- **Service 方法**：
  - `main/app/service/other.go` - `GetOrderRemarkList()` (line 259-279)
  - `main/app/service/other.go` - `GetOrderItemRemarkList()` (line 372-392) - **已存在，可直接复用**
- **响应结构**：
  - `main/app/dto/resp/order_remark.go` - `OrderRemarkResp`
  - `main/app/dto/resp/order_item_remark.go` - `OrderItemRemarkResp` - **已存在**

### 外部参考

- 无

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-05  
**作者**: {团队/个人}  
**审核者**: {审核者}

