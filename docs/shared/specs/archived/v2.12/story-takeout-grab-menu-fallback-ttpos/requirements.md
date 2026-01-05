> ⚠️ **已归档** - 此 Spec 已随 v2.12 发布。
>
> - 归档时间: 2026-01-04
> - 归档人: weifashi

# Grab 菜单获取回退 TTPOS 导出接口 需求文档

> 本文档定义 Grab 菜单获取回退 TTPOS 导出接口功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/grab-menu-fallback-ttpos-export.md](../../../../team/proposals/2025-12/grab-menu-fallback-ttpos-export.md) |
| **创建日期**      | 2025-12-18                                                                                                 |
| **负责人**        | rikugun                                                                                                       |
| **目标 Sprint**   | Sprint TBD                                                                                                   |
| **涉及技术栈**    | [x] Go (ttpos-bmp/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | ✅ 已通过 |
| **审核人**   | rikugun              |
| **审核日期** | 2025-12-18              |
| **审核意见** | 需求清晰，可进入设计阶段          |

---

## 📋 概述

当商家绑定 Grab 外卖平台时选择"跳过导出菜单"，系统应支持在 Grab 调用 `HandleGetMenu` 获取菜单时，回退调用 TTPOS 主模块的 `/api/v1/takeout/menu/export` 接口实时获取菜单数据，确保商家可以正常使用 Grab 外卖服务。

## 🎯 产品对齐

- **支持灵活绑定**：商家可选择主动推送菜单或按需拉取菜单
- **提升用户体验**：减少菜单同步问题导致的服务中断
- **保持数据一致**：与 TTPOS 主系统菜单数据保持一致

## 📝 用户故事

**作为** Grab 外卖平台  
**我想** 在调用 GetMenu 接口时能够获取商家菜单  
**以便于** 在 Grab App 上正确展示商家的外卖菜单

---

## 功能需求

### Requirement 1: 优先读取本地菜单快照

**用户故事**: 作为系统，我想优先从本地缓存读取菜单，以便于减少外部调用提升性能

#### 验收标准

1. **WHEN** Grab 调用 GetMenu 接口 **AND** `channel_menu_snapshot.ttpos_menu_data` 存在有效数据 **THEN** 系统 **SHALL** 直接返回本地快照数据
2. **IF** 读取本地快照发生数据库错误 **THEN** 系统 **SHALL** 使用 `gerror.Wrap` 包装错误并返回

#### 具体要求

- [x] 1.1 使用 `service.ChannelMenu().GetTtposMenu()` 读取本地快照
- [x] 1.2 快照存在时直接解析并返回 `grabfood.GetMenuNewResponse`
- [x] 1.3 数据库错误使用 `gerror.Wrap(err, "描述")` 包装

---

### Requirement 2: 回退调用 TTPOS 导出接口

**用户故事**: 作为系统，我想在本地快照为空时调用 TTPOS 接口获取菜单，以便于保证菜单数据可用性

#### 验收标准

1. **WHEN** 本地菜单快照为空 **THEN** 系统 **SHALL** 调用 TTPOS `/api/v1/takeout/menu/export` 接口
2. **WHEN** 调用 TTPOS 接口 **THEN** 系统 **SHALL** 携带 `X-TTPOS-SECRET` 认证头
3. **IF** TTPOS 接口调用成功 **THEN** 系统 **SHALL** 将响应转换为 Grab SDK 格式返回
4. **IF** TTPOS 接口调用失败 **THEN** 系统 **SHALL** 记录错误日志并返回 `gerror.Wrap` 包装的错误

#### 具体要求

- [ ] 2.1 新增 `fetchMenuFromTTpos(ctx, shopUUID)` 方法
- [ ] 2.2 从配置读取 TTPOS endpoint (`app.ttposEndpoint`)
- [ ] 2.3 使用 MD5 加密生成认证头：`MD5(shopUUID + callbackSecret)`
- [ ] 2.4 请求体包含 `platform: "grab"` 和 `companyUuid`
- [ ] 2.5 解析响应中的 `data.menuData` 字段
- [ ] 2.6 配置缺失时返回 `gerror.NewCode(gcode.CodeMissingConfiguration, ...)`

---

### Requirement 3: 错误处理规范

**用户故事**: 作为运维人员，我想系统有规范的错误处理和日志记录，以便于问题排查

#### 验收标准

1. **WHEN** 参数无效 **THEN** 系统 **SHALL** 返回 `gerror.NewCode(gcode.CodeInvalidParameter, ...)`
2. **WHEN** 菜单未找到 **THEN** 系统 **SHALL** 返回 `gerror.NewCode(gcode.CodeNotFound, "menu not found")`
3. **WHEN** 回退调用 TTPOS **THEN** 系统 **SHALL** 记录 Info 级别日志
4. **IF** 回退调用失败 **THEN** 系统 **SHALL** 记录 Error 级别日志

#### 具体要求

- [ ] 3.1 所有错误使用 `gerror` 包装，包含错误码和描述
- [ ] 3.2 日志格式：`[Grab] {操作描述}: shopUUID=%d, {其他参数}`
- [ ] 3.3 敏感信息（如 secret）不记录到日志

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 遵循现有 `sGrabMenu` 服务结构
- **单一职责原则**: `fetchMenuFromTTpos` 方法独立负责 TTPOS 接口调用
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范

### API 设计要求

- [x] 使用现有 `HandleGetMenu` 接口，无需新增 API
- [x] 响应格式保持 Grab SDK `GetMenuNewResponse` 结构不变

### 性能要求

- [ ] TTPOS 接口调用超时设置为 10s
- [ ] 本地快照优先，减少外部调用

### 测试要求

- [ ] 单元测试覆盖 `fetchMenuFromTTpos` 方法
- [ ] Mock TTPOS 接口响应进行集成测试
- [ ] 测试场景：本地快照存在、本地快照为空、TTPOS 调用失败

### 安全要求

- [x] 认证头使用 MD5 加密生成
- [ ] 配置中的 `callbackSecret` 不记录到日志

---

## 验收标准

### 功能验收

1. **本地快照存在**: 调用 GetMenu 返回本地数据，不调用 TTPOS 接口
2. **本地快照为空**: 调用 GetMenu 回退调用 TTPOS 接口，返回正确菜单
3. **TTPOS 调用失败**: 返回 "menu not found" 错误并记录日志

### 测试验收

1. **单元测试**: `fetchMenuFromTTpos` 方法测试通过
2. **集成测试**: 端到端 GetMenu 流程测试通过
3. **错误场景**: 各类错误返回正确错误码

### 文档验收

1. **技术文档**: design.md 包含详细实现方案
2. **配置文档**: 说明 `app.ttposEndpoint` 配置项

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 使用 `g.Client()` 发起 HTTP 请求
- 使用 `gerror` 包装所有错误
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

### 业务约束

- TTPOS 主模块 `/api/v1/takeout/menu/export` 接口需支持 `X-TTPOS-SECRET` 认证
- `shopUUID` 与 `companyUuid` 的映射关系需确认

### 资源约束

- 开发时间: 1-2 天
- Story Point: 3 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2/crypto/gmd5` - MD5 加密
- `github.com/gogf/gf/v2/errors/gerror` - 错误包装
- `github.com/grab/grabfood-api-sdk-go` - Grab SDK

### 服务依赖

- **BMP → Main**: HTTP API 调用 `/api/v1/takeout/menu/export`

### 业务依赖

- TTPOS 主模块菜单导出接口已就绪
- `X-TTPOS-SECRET` 认证机制已实现

---

## 风险和缓解

### 风险 1: TTPOS 接口超时

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 设置合理超时时间（10s）
- 超时后返回友好错误信息

### 风险 2: 认证机制不兼容

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 开发前与 main 模块确认认证规则
- 复用现有 `getCallBackAuth` 逻辑

### 风险 3: shopUUID 与 companyUUID 映射

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 查阅数据模型确认关系
- 必要时增加映射查询

---

## 时间表

- **Phase 1 - 开发**: 1 天
- **Phase 2 - 测试**: 0.5 天
- **Phase 3 - 文档**: 0.5 天
- **总计**: 2 天（SP = 3）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范

### 相关文件

- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go` - 现有实现
- `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/skootar.go` - 认证参考
- `main/app/api/v1/takeout/menu_handler.go` - TTPOS 导出接口

### 相关文档

- `docs/shared/integrations/grab/grab-menu-integration.md` - Grab 菜单集成文档

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-12/2025-12-18.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-18  
**作者**: rikugun  
**审核者**: 待定

