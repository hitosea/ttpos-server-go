# POS Invoice 外卖订单字段扩展 需求文档

> 本文档定义 POS Invoice 外卖订单字段扩展 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.12-pos-invoice-takeout-fields.md](../../../../team/proposals/2025-12/v2.12-pos-invoice-takeout-fields.md) |
| **创建日期**      | 2025-12-26                                                                                                   |
| **负责人**        | rikugun                                                                                                      |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |
| **关联任务**      | DooTask #38169                                                                                               |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核                   |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

当前 Grab 外卖平台接入 TTPOS 后，外卖订单需要同步到 ERPNext 的 POS Invoice 中。为了能够追溯和识别外卖订单的来源信息，需要在 POS Invoice 中记录第三方平台的订单号和平台标识。该功能实现外卖订单在 ERPNext 系统中的完整追溯，支持按平台进行数据分析和财务对账。

## 🎯 产品对齐

该功能支持产品在 ERP 集成的核心战略，实现外卖订单数据的完整记录和追溯，为商户提供更精细化的订单管理和数据分析能力，支持未来接入更多外卖平台（如 Foodpanda、Line Man 等）。

## 📝 用户故事

**作为** 商户管理员和财务人员  
**我想** 在 ERPNext 的 POS Invoice 中记录外卖订单的第三方平台订单号和平台标识  
**以便于** 能够追溯外卖订单来源，支持对账和数据分析

---

## 功能需求

### Requirement 1: ERPNext 自定义字段创建

**用户故事**: 作为系统管理员，我想在 POS Invoice 中增加外卖订单相关字段，以便于记录和追溯外卖订单信息

#### 验收标准

1. **WHEN** 执行 ERPNext 迁移脚本 **THEN** 系统 **SHALL** 在 POS Invoice DocType 中创建 `custom_takeout_order_no` 和 `custom_takeout_provider` 字段
2. **WHEN** 字段创建成功 **THEN** 系统 **SHALL** 支持在 POS Invoice 中存储和查询这些字段的值
3. **IF** 字段已存在 **THEN** 系统 **SHALL** 跳过创建，避免重复

#### 具体要求

- [ ] 1.1 创建 `01_pos_invoice_takeout_order_no.json` 迁移文件
  - 字段名：`custom_takeout_order_no`
  - 字段类型：Data（字符串）
  - 字段标签：Takeout Order No
  - DocType：POS Invoice
- [ ] 1.2 创建 `02_pos_invoice_takeout_provider.json` 迁移文件
  - 字段名：`custom_takeout_provider`
  - 字段类型：Data（字符串）
  - 字段标签：Takeout Provider
  - DocType：POS Invoice
- [ ] 1.3 字段位置：参考现有自定义字段的插入位置
- [ ] 1.4 迁移文件格式参考：`ttpos-bmp/app/ttpos-erp/manifest/erp-migrate/v2.12/01_custom_field/01_custom_payment_id.json`

---

### Requirement 2: Protobuf 接口扩展

**用户故事**: 作为开发者，我想在 SavePosInvoice 接口中支持外卖订单字段，以便于传递外卖订单信息

#### 验收标准

1. **WHEN** 调用 SavePosInvoice 时传入 `takeout_order_no` 和 `takeout_provider` **THEN** 系统 **SHALL** 接收并处理这些字段
2. **IF** 未传入这些字段 **THEN** 系统 **SHALL** 正常处理，不影响现有功能（向后兼容）
3. **WHEN** 字段传入后 **THEN** 系统 **SHALL** 将值保存到 ERPNext 的对应自定义字段中

#### 具体要求

- [ ] 2.1 在 `SavePosInvoiceReq` 中增加两个可选字段：
  - `optional string takeout_order_no = 17;` - 外卖订单号，可选
  - `optional string takeout_provider = 18;` - 外卖平台提供商，可选
- [ ] 2.2 字段编号从 17 开始（当前最大编号为 16）
- [ ] 2.3 字段必须为可选（optional），确保向后兼容
- [ ] 2.4 重新生成 protobuf Go 代码

---

### Requirement 3: DTO 结构更新

**用户故事**: 作为开发者，我想在 POS Invoice DTO 中包含外卖订单字段，以便于数据传递和序列化

#### 验收标准

1. **WHEN** 构建 POS Invoice 数据 **THEN** 系统 **SHALL** 支持设置 `custom_takeout_order_no` 和 `custom_takeout_provider` 字段
2. **WHEN** 从 ERPNext 查询 POS Invoice **THEN** 系统 **SHALL** 正确解析这两个字段的值
3. **IF** 字段值为空 **THEN** 系统 **SHALL** 正确处理，不报错

#### 具体要求

- [ ] 3.1 在 `POSInvoice` 结构体中增加字段：
  - `CustomTakeoutOrderNo string \`json:"custom_takeout_order_no,omitempty"\`` - 外卖订单号
  - `CustomTakeoutProvider string \`json:"custom_takeout_provider,omitempty"\`` - 外卖平台提供商
- [ ] 3.2 字段位置：放在其他自定义字段附近（如 `CustomPosOpeningEntry` 之后）
- [ ] 3.3 使用 `omitempty` 标签，确保空值不序列化

---

### Requirement 4: 业务逻辑实现

**用户故事**: 作为开发者，我想在 SavePosInvoice 实现中支持外卖订单字段的传递，以便于将数据保存到 ERPNext

#### 验收标准

1. **WHEN** SavePosInvoice 接收到 `takeout_order_no` 和 `takeout_provider` **THEN** 系统 **SHALL** 将这些值映射到 DTO 的对应字段
2. **WHEN** 构建 POS Invoice 数据时 **THEN** 系统 **SHALL** 将 DTO 字段值传递到 ERPNext
3. **IF** 字段值为空 **THEN** 系统 **SHALL** 不设置该字段，使用默认值（空字符串）

#### 具体要求

- [ ] 4.1 在 `buildPosInvoice` 方法中增加字段赋值逻辑：
  ```go
  if len(req.TakeoutOrderNo) > 0 {
      posInvoice.CustomTakeoutOrderNo = req.TakeoutOrderNo
  }
  if len(req.TakeoutProvider) > 0 {
      posInvoice.CustomTakeoutProvider = req.TakeoutProvider
  }
  ```
- [ ] 4.2 字段赋值位置：在设置其他自定义字段之后
- [ ] 4.3 确保空值检查，避免设置空字符串

---

### Requirement 5: 外卖订单数据同步

**用户故事**: 作为商户管理员，我想外卖订单支付完成时自动同步订单号和平台信息到 ERPNext，以便于完整记录订单信息

#### 验收标准

1. **WHEN** Grab 外卖订单支付完成同步到 ERPNext 时 **THEN** 系统 **SHALL** 自动填充 `custom_takeout_order_no` 为 Grab 订单号
2. **WHEN** Grab 外卖订单支付完成同步到 ERPNext 时 **THEN** 系统 **SHALL** 自动填充 `custom_takeout_provider` 为 "grab"
3. **IF** 订单不是外卖订单（`OrderSourceUuid = 0`） **THEN** 系统 **SHALL** 不设置这两个字段
4. **IF** 订单是外卖订单但缺少 `RelatedOrderNo` **THEN** 系统 **SHALL** 记录警告日志但不中断流程

#### 具体要求

- [ ] 5.1 在调用 SavePosInvoice 前，检查订单是否为外卖订单：
  - `OrderSourceUuid > 0` 表示外卖订单
- [ ] 5.2 从 `MemberSaleOrder.RelatedOrderNo` 获取第三方订单号
- [ ] 5.3 从 `MemberSaleOrder.RelatedOrderType` 获取平台标识（如 "grab"）
- [ ] 5.4 将这两个值传递给 SavePosInvoice 的对应字段
- [ ] 5.5 数据同步位置：在 `main/app/service/rpc/erp/selling.go` 的 SavePosInvoice 调用处

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Logic 分层（Go BMP 模块）
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Logic 层应独立且可复用
- **依赖管理**: 遵循 GoFrame 框架规范
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [ ] Protobuf 字段使用 snake_case 命名（如：`takeout_order_no`）
- [ ] 响应格式通过 `erp.ApiResponse` 包装
- [ ] 字段必须为可选（optional），确保向后兼容
- [ ] 参考: `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范

### 数据库设计要求

- [ ] ERPNext 自定义字段使用 `custom_` 前缀
- [ ] 字段类型：Data（字符串类型）
- [ ] 字段命名遵循 ERPNext 规范
- [ ] 参考: ERPNext 自定义字段文档

### 性能要求

- [ ] 字段查询不影响现有 POS Invoice 查询性能
- [ ] 字段保存操作时间 < 50ms
- [ ] 支持并发创建 POS Invoice

### 测试要求

- [ ] Logic 层测试覆盖率 ≥ 70%
- [ ] Protobuf 序列化/反序列化测试
- [ ] 集成测试覆盖外卖订单同步流程
- [ ] 向后兼容性测试（不传新字段时功能正常）

### 安全要求

- [ ] 字段值进行长度限制（防止过长字符串）
- [ ] 字段值进行格式验证（如平台标识只允许特定值）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 字段保存失败时不影响 POS Invoice 主流程
- [ ] 错误日志记录（使用 Logger）
- [ ] 支持字段值的查询和追溯

---

## 验收标准

### 功能验收

1. **ERPNext 字段创建**: 迁移脚本执行后，POS Invoice 中可看到两个新字段
2. **Protobuf 接口**: SavePosInvoice 接口支持接收和传递新字段
3. **DTO 结构**: POS Invoice DTO 包含新字段，可正确序列化/反序列化
4. **业务逻辑**: SavePosInvoice 实现中正确传递新字段到 ERPNext
5. **数据同步**: Grab 外卖订单支付完成时，自动填充新字段值

### 测试验收

1. **单元测试**: Logic 层字段赋值逻辑测试通过
2. **集成测试**: 端到端流程测试通过（从订单支付到 ERPNext 保存）
3. **向后兼容测试**: 不传新字段时，现有功能正常
4. **手动测试**: 在 ERPNext 中验证字段值和显示

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **API 文档**: Protobuf 接口文档更新
3. **迁移文档**: ERPNext 字段迁移脚本和说明完整
4. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- Protobuf 文件修改后需重新生成 Go 代码
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

#### ERPNext 集成

- 自定义字段必须使用 `custom_` 前缀
- 字段类型必须为 Data（字符串）
- 迁移文件格式必须符合 ERPNext 规范
- 参考现有自定义字段示例

### 业务约束

- 字段为可选，不影响现有店内订单流程
- 字段值长度限制：订单号最大 100 字符，平台标识最大 50 字符
- 平台标识只允许特定值：grab, foodpanda, lineman 等

### 资源约束

- 开发时间: 2-3 天
- Story Point: 3 SP (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `ttpos-bmp/app/ttpos-erp` - ERP 模块
- `github.com/gogf/gf/v2` - GoFrame 框架
- ERPNext API - 自定义字段管理

### 服务依赖

- **Main → BMP**: gRPC 调用 SavePosInvoice
- **BMP → ERPNext**: HTTP API 调用创建/更新 POS Invoice

### 业务依赖

- 外卖订单支付完成流程（InstantOrderPaymentFinish）
- MemberSaleOrder 数据模型（RelatedOrderNo, RelatedOrderType）
- ERPNext POS Invoice DocType

---

## 风险和缓解

### 风险 1: ERPNext 自定义字段迁移失败

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 使用标准的 ERPNext 自定义字段迁移方式
- 参考现有示例文件格式
- 在测试环境先验证迁移脚本
- 提供回滚方案

### 风险 2: 向后兼容性问题

**影响**: 高  
**概率**: 低  
**缓解措施**:

- Protobuf 字段设为可选（optional）
- 字段值检查，空值不设置
- 完整的向后兼容性测试
- 灰度发布，先小范围验证

### 风险 3: 字段命名不符合 ERPNext 规范

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 严格遵循 `custom_` 前缀规范
- 参考现有自定义字段命名
- 在 ERPNext 文档中验证命名规则

---

## 时间表

- **Phase 1 - ERPNext 字段创建**: 0.5 天
  - 创建迁移文件
  - 验证字段创建
- **Phase 2 - Protobuf 和 DTO 更新**: 0.5 天
  - 更新 Protobuf 定义
  - 更新 DTO 结构
  - 重新生成代码
- **Phase 3 - 业务逻辑实现**: 1 天
  - 实现字段传递逻辑
  - 实现数据同步逻辑
  - 单元测试
- **Phase 4 - 集成测试和文档**: 1 天
  - 端到端测试
  - 文档编写
  - 代码审查
- **总计**: 3 天（SP = 3）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范
- `.cursor/rules/api.mdc` - API 设计规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构
- ERPNext 自定义字段文档: https://docs.erpnext.com/docs/user/en/customize-erpnext/custom-fields

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- 现有自定义字段示例: `ttpos-bmp/app/ttpos-erp/manifest/erp-migrate/v2.12/01_custom_field/01_custom_payment_id.json`

### 外部参考

- ERPNext 自定义字段文档: https://docs.erpnext.com/docs/user/en/customize-erpnext/custom-fields
- ERPNext POS Invoice DocType: https://github.com/frappe/erpnext/blob/develop/erpnext/accounts/doctype/pos_invoice/pos_invoice.json

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-26  
**作者**: rikugun  
**审核者**: {审核者}

