# 调整 GetMenuSnapshotResp.content 为 GetMenuSnapshotResp.menu_data 需求文档

> 本文档定义调整菜单快照响应字段命名的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/rename-menu-snapshot-content-field.md](../../../../team/proposals/2025-12/rename-menu-snapshot-content-field.md) |
| **创建日期**      | 2025-12-15                                                                                                 |
| **负责人**        | {姓名}                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | {YYYY-MM-DD}             |
| **审核意见** | {审核意见，如有}         |

---

## 📋 概述

统一菜单快照接口的字段命名，将 `GetMenuSnapshotResp.content` 重命名为 `GetMenuSnapshotResp.menu_data`，使其与 `SaveMenuSnapshotReq.menu_data` 保持一致，提升代码可读性和维护性。

## 🎯 产品对齐

该功能属于代码质量改进，通过统一 API 字段命名，降低开发人员理解成本，减少因命名不一致导致的潜在错误，提升整体代码质量和可维护性。

## 📝 用户故事

**作为** 开发人员  
**我想** 统一菜单快照接口的字段命名  
**以便于** 提高代码可读性和维护性，减少因命名不一致导致的错误

---

## 功能需求

### Requirement 1: 修改 Protobuf 字段定义

**用户故事**: 作为开发人员，我想将 `GetMenuSnapshotResp.content` 改为 `GetMenuSnapshotResp.menu_data`，以便于与 `SaveMenuSnapshotReq.menu_data` 保持一致

#### 验收标准

1. **WHEN** 查看 `menu.proto` 文件中的 `GetMenuSnapshotResp` 消息定义 **THEN** 字段名应为 `menu_data` **SHALL** 与 `SaveMenuSnapshotReq.menu_data` 保持一致
2. **WHEN** 字段编号为 2 **THEN** 字段类型为 `string` **SHALL** 保持不变
3. **WHEN** 查看字段注释 **THEN** 注释应清晰说明字段用途 **SHALL** 与原有语义保持一致

#### 具体要求

- [ ] 1.1 修改 `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto` 文件
- [ ] 1.2 将 `GetMenuSnapshotResp` 中的 `string content = 2;` 改为 `string menu_data = 2;`
- [ ] 1.3 更新字段注释，保持语义清晰

---

### Requirement 2: 重新生成 Protobuf Go 代码

**用户故事**: 作为开发人员，我想重新生成 protobuf Go 代码，以便于代码与新的字段定义保持一致

#### 验收标准

1. **WHEN** 执行 protobuf 代码生成命令 **THEN** 生成的 Go 代码 **SHALL** 包含 `MenuData` 字段和 `GetMenuData()` 方法
2. **WHEN** 查看生成的 `menu.pb.go` 文件 **THEN** 不应再包含 `Content` 字段和 `GetContent()` 方法
3. **WHEN** 编译项目 **THEN** 编译 **SHALL** 成功，无语法错误

#### 具体要求

- [ ] 2.1 执行 protobuf 代码生成命令（根据项目规范）
- [ ] 2.2 验证生成的 `ttpos-bmp/app/ttpos-takeout/api/menu/menu.pb.go` 文件
- [ ] 2.3 确认字段名和方法名正确更新

---

### Requirement 3: 更新业务代码中的字段引用

**用户故事**: 作为开发人员，我想更新业务代码中对字段的引用，以便于代码与新生成的 protobuf 定义保持一致

#### 验收标准

1. **WHEN** 查看 `channel_menu.go` 中的 `GetMenuSnapshot` 方法 **THEN** 字段赋值 **SHALL** 使用 `MenuData` 而非 `Content`
2. **WHEN** 编译项目 **THEN** 编译 **SHALL** 成功，无编译错误
3. **WHEN** 运行单元测试 **THEN** 相关测试 **SHALL** 通过

#### 具体要求

- [ ] 3.1 更新 `ttpos-bmp/app/ttpos-takeout/internal/logic/channel_menu/channel_menu.go` 中的字段引用
- [ ] 3.2 将 `resp.Content = content` 改为 `resp.MenuData = content`
- [ ] 3.3 检查是否有其他文件引用了该字段，如有则一并更新

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] Protobuf 字段命名使用 snake_case（如：`menu_data`）
- [x] 字段编号保持连续，不改变现有编号
- [x] 字段类型保持不变
- [x] 参考: `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范

### 性能要求

- [x] 不影响现有性能（纯字段重命名，无逻辑变更）
- [x] 响应时间保持不变

### 测试要求

- [ ] 单元测试覆盖相关业务逻辑
- [ ] 集成测试验证接口调用正常
- [ ] 手动测试验证字段重命名后功能正常
- [ ] 参考: `ttpos-bmp/.cursor/rules/go-rules.mdc` - 测试规范

### 可靠性要求

- [x] 向后兼容性检查（如有外部系统依赖，需协调更新）
- [x] 错误日志记录（使用 Logger）
- [x] 代码生成流程验证

---

## 验收标准

### 功能验收

1. **Proto 定义正确**: `GetMenuSnapshotResp` 中的字段名为 `menu_data`，与 `SaveMenuSnapshotReq.menu_data` 保持一致
2. **代码生成正确**: 生成的 Go 代码包含 `MenuData` 字段和 `GetMenuData()` 方法
3. **业务代码更新**: `channel_menu.go` 中的字段引用已更新为 `MenuData`
4. **编译通过**: 项目编译成功，无语法错误
5. **测试通过**: 相关单元测试和集成测试通过

### 测试验收

1. **单元测试**: 相关业务逻辑测试通过
2. **API 测试**: `GetMenuSnapshot` 接口测试通过，响应字段为 `menu_data`
3. **集成测试**: 端到端流程测试通过
4. **手动测试**: 验证接口调用和响应格式正确

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **代码注释**: 相关代码注释已更新
3. **变更记录**: 在相关文档中记录字段重命名变更

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`

### 业务约束

- 字段重命名不影响业务逻辑
- 保持字段编号不变（field number = 2）
- 保持字段类型不变（string）

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `protobuf` - Protobuf 代码生成工具
- `ttpos-bmp/app/ttpos-takeout/api/menu` - 生成的 Go 代码

### 服务依赖

- 无外部服务依赖

### 业务依赖

- 无业务依赖

---

## 风险和缓解

### 风险 1: 向后兼容性问题

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 检查是否有外部系统依赖该接口，如有需要协调更新
- 在开发环境充分测试
- 如有必要，提供迁移指南

### 风险 2: Protobuf 代码生成失败

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 在开发环境验证 protobuf 代码生成流程
- 确保代码生成工具版本正确
- 检查生成后的代码格式

### 风险 3: 遗漏字段引用更新

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 使用代码搜索工具全面检查字段引用
- 编译时检查所有引用
- 运行完整的测试套件

---

## 时间表

- **Phase 1 - Proto 定义修改**: 0.1 天
- **Phase 2 - 代码生成和验证**: 0.2 天
- **Phase 3 - 业务代码更新和测试**: 0.2 天
- **总计**: 0.5 天（SP = 1）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南

### 外部参考

- [Protocol Buffers Language Guide](https://protobuf.dev/programming-guides/proto3/)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-15  
**作者**: rikugun  
**审核者**: {审核者}
