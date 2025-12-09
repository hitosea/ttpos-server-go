# KDS 等待时长颜色配置 需求文档（后端）

> 本文档定义 Shop 管理端和 KDS 厨显端等待时长颜色配置功能的后端详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                      |
| ----------------- | --------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11/kds-wait-time-color-config.md](../../../../team/proposals/2025-11/kds-wait-time-color-config.md) |
| **关联前端 Spec** | [前端仓库: docs/shared/specs/active/story-shop-kds-wait-time-color-config/requirements.md](../../../../../ttpos-flutter/docs/shared/specs/active/story-shop-kds-wait-time-color-config/requirements.md) |
| **关联任务**      | DooTask #37471 - 新管理端-厨显设置支持自定义配置等待时长、颜色 |
| **创建日期**      | 2025-12-09                                                                                              |
| **负责人**        | {待分配}                                                                                                    |
| **目标 Sprint**   | Sprint {待定}                                                                                                |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [x] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过 |
| **审核人**   | {审核人姓名}             |
| **审核日期** | 2025-12-09             |
| **审核意见** | 需求已通过审核，可进入技术设计阶段         |

---

## 📋 概述

当前厨显（KDS）系统的等待时长颜色配置是固定的，无法根据不同门店的实际经营场景进行调整。本需求旨在为新管理端（Shop）提供等待时长颜色配置功能的后端支持，允许门店管理员自定义配置等待时长的颜色方案，配置通过配置中心同步到新旧两个后台，并实时下发到 KDS 终端生效。

**核心价值**：
- 提升厨房效率：门店可根据自身出餐节奏自定义时间阈值，更精准地提醒厨房人员
- 改善可视化管理：自定义颜色方案可适配不同厨房环境光线，提高颜色辨识度
- 降低运营成本：门店管理员可自主配置，减少对技术支持的依赖
- 统一配置管理：新旧后台配置同步，确保数据一致性

**实现范围**：
- 后端 API 接口（Go Main 模块）
- 配置中心下发机制（WebSocket 推送）
- 数据迁移脚本（PHP 文件，初始化默认配置）
- 权限管理（管理APP权限，厨显设置）

## 🎯 产品对齐

- **提升厨房效率**：门店可根据自身出餐节奏自定义时间阈值，更精准地提醒厨房人员
- **改善可视化管理**：自定义颜色方案可适配不同厨房环境光线，提高颜色辨识度
- **降低运营成本**：门店管理员可自主配置，减少对技术支持的依赖
- **统一配置管理**：新旧后台配置同步，确保数据一致性

## 📝 用户故事

**作为** 商户管理员  
**我想** 在新管理端配置厨显的等待时长颜色方案  
**以便于** 根据门店实际经营场景（快餐、正餐、咖啡等）调整时间阈值和颜色，提高厨房效率。

**作为** 系统  
**我想** 在新管理端配置保存后同步到旧商家后台和 KDS 终端  
**以便于** 确保新旧系统数据一致并实时生效。

---

## 功能需求

### Requirement 1: 厨显设置 API 接口设计

**用户故事**: 作为前端开发者，我想调用后端 API 获取和保存厨显等待时长颜色配置，以便于实现配置界面功能。

#### 验收标准

1. **WHEN** 前端调用获取厨显设置接口 **THEN** 系统 **SHALL** 返回完整的厨显设置信息，包括等待时长颜色配置
2. **WHEN** 前端调用保存厨显设置接口 **AND** 传递等待时长颜色配置参数 **THEN** 系统 **SHALL** 验证参数合法性，保存配置到数据库，同步到旧后台，并通过 WebSocket 下发到 KDS 终端
3. **WHEN** 前端传递的时间区间重叠（第三区间起点 ≤ 第二区间） **THEN** 系统 **SHALL** 返回错误提示："区间不可重叠，第三区间起点必须大于第二区间"
4. **WHEN** 前端传递的时间区间不在有效范围（1-60 分钟） **THEN** 系统 **SHALL** 返回错误提示："时间区间必须在 1-60 分钟之间"
5. **WHEN** 前端传递的颜色格式不正确（非 RGB 格式） **THEN** 系统 **SHALL** 返回错误提示："颜色格式不正确，请使用 RGB 格式（如 #100A05）"

#### 具体要求

- [ ] 1.1 在 `main/app/api/v1/shop/shop_setting.go` 中新增 `GetKitchenSetting` 接口（如不存在则创建）
  - URL: `GET /api/v1/shop/setting/kitchen`
  - 返回完整的厨显设置信息，包括等待时长颜色配置
- [ ] 1.2 在 `main/app/api/v1/shop/shop_setting.go` 中新增 `SaveKitchenSetting` 接口
  - URL: `POST /api/v1/shop/setting/kitchen`
  - 接收等待时长颜色配置参数
  - 验证参数合法性（时间区间、颜色格式）
  - 保存配置到数据库
  - 同步到旧后台（调用 PHP API）
  - 通过 WebSocket 推送配置更新
- [ ] 1.3 在 `main/app/dto/req/setting/kitchen_setting.go` 中新增请求 DTO
  - `WaitTimeColorConfig` 结构体，包含：
    - `IsWaitColor` (string): 是否开启等待时长颜色 0-关闭 1-开启
    - `WaitTimeColorRanges` ([]WaitTimeColorRange): 等待时长颜色区间配置
      - `WaitTimeColorRange` 结构体：
        - `Minute` (int): 时间阈值（分钟）
        - `Color` (string): 颜色值（RGB 格式，统一使用 #xxxxxx 格式）
          - 黑色：`#100A05`
          - 黄色：`#FFBE00`
          - 红色：`#E50028`
- [ ] 1.4 在 `main/app/dto/resp/setting/kitchen_setting.go` 中更新响应 DTO
  - 在 `KitchenResp` 结构体中新增 `WaitTimeColorRanges` ([]WaitTimeColorRange) 字段
  - 保留 `WaitColor` ([]string) 字段，保持向后兼容
- [ ] 1.5 在 `main/app/service/setting/setting.go` 中实现 `SaveKitchenSetting` 方法
  - 参数验证（时间区间、颜色格式）
  - 保存配置到 `setting` 表（key = `kitchen`）
  - 调用旧后台同步方法
  - 推送 WebSocket 配置更新
- [ ] 1.6 在 `main/app/service/setting/setting.go` 中更新 `GetKitchenSetting` 方法
  - 返回新的等待时长颜色配置格式（`WaitTimeColorRanges`）
  - 同时返回旧格式数据（`WaitColor`），保持向后兼容

---

### Requirement 2: 等待时长颜色配置数据模型

**用户故事**: 作为系统，我想存储和读取等待时长颜色配置，以便于持久化配置数据。

#### 验收标准

1. **WHEN** 保存等待时长颜色配置 **THEN** 系统 **SHALL** 将配置以 JSON 格式存储到 `setting` 表的 `values` 字段中
2. **WHEN** 读取等待时长颜色配置 **THEN** 系统 **SHALL** 正确解析 JSON 数据并返回配置信息
3. **WHEN** 配置数据不存在 **THEN** 系统 **SHALL** 返回默认配置：
   - 第一区间：0 分钟（黑色，#100A05）
   - 第二区间：10 分钟（颜色从旧后台配置读取，如无则使用黄色 #FFBE00）
   - 第三区间：20 分钟及以上（颜色从旧后台配置读取，如无则使用红色 #E50028）

#### 具体要求

- [ ] 2.1 定义等待时长颜色配置数据结构
  - 在 `main/app/dto/resp/setting/kitchen_setting.go` 中定义 `WaitTimeColorRange` 结构体
  - 字段：`Minute` (int), `Color` (string)
- [ ] 2.2 更新 `Kitchen` 结构体
  - 新增 `WaitTimeColorRanges` ([]WaitTimeColorRange) 字段
  - 保留 `WaitColor` ([]string) 字段，保持向后兼容
- [ ] 2.3 实现配置数据转换逻辑
  - 旧格式转新格式：`WaitColor` []string（如 `["red", "yellow"]`）→ `WaitTimeColorRanges` []WaitTimeColorRange（读取时转换）
    - `"red"` → `#E50028`
    - `"yellow"` → `#FFBE00`
    - 第一区间固定：0 分钟 → `#100A05`（黑色）
  - 新格式转旧格式：`WaitTimeColorRanges` → `WaitColor`（保存时同时保存，保持兼容）
    - RGB 格式转换为 red/yellow（如 `#E50028` → `"red"`，`#FFBE00` → `"yellow"`）
    - 其他 RGB 颜色保持原值（不转换）
- [ ] 2.4 实现默认配置逻辑
  - 在 `main/app/service/setting/default.go` 中更新 `getDefaultKitchen` 方法
  - 设置默认等待时长颜色配置

---

### Requirement 3: 权限管理集成

**用户故事**: 作为系统管理员，我想在角色权限系统中配置「厨显设置」权限，以便于控制哪些员工可以访问和配置厨显等待时长颜色。

#### 验收标准

1. **WHEN** 系统管理员在角色权限系统中查看权限树 **THEN** 系统 **SHALL** 在「管理 APP → 工作台 → 其他 → 各端设置」节点下显示「厨显设置」权限项
2. **WHEN** 系统管理员创建或编辑角色 **AND** 配置功能权限 **THEN** 系统 **SHALL** 允许勾选/取消勾选「厨显设置」权限项
3. **WHEN** 用户角色包含「厨显设置」权限 **AND** 用户访问「各端设置」模块 **THEN** 系统 **SHALL** 显示「厨显设置」入口
4. **WHEN** 用户角色不包含「厨显设置」权限 **AND** 用户访问「各端设置」模块 **THEN** 系统 **SHALL** 不显示「厨显设置」入口

#### 具体要求

- [ ] 3.1 创建权限迁移文件（PHP）
  - 文件路径：`admin/database/migrations/{YYYYMMDDHHMMSS}_add_kitchen_wait_time_color_access.php`
  - 在 `access` 表中新增「厨显设置」权限项
  - 权限路径：管理 APP → 工作台 → 其他 → 各端设置 → 厨显设置
  - parent_uuid: 2859290595328000（各端设置）
  - sort: 2（排号取餐是1，厨显设置是2）
- [ ] 3.2 权限赋给所有角色
  - 查询所有角色（`role` 表）
  - 为每个角色在 `role_access` 表中添加「厨显设置」权限
- [ ] 3.3 实现权限校验逻辑
  - 根据用户角色权限动态显示/隐藏「厨显设置」入口
  - 无权限用户访问时显示无权限提示页面

---

### Requirement 4: 配置中心下发机制（保留原编号）

**用户故事**: 作为系统，我想在配置保存后通过配置中心下发到 KDS 终端，以便于 KDS 终端实时应用新配置。

#### 验收标准

1. **WHEN** 新管理端保存厨显设置成功 **THEN** 系统 **SHALL** 通过 WebSocket 推送配置更新事件
2. **WHEN** WebSocket 推送成功 **THEN** 系统 **SHALL** KDS 终端在 5 秒内接收到配置更新事件
3. **WHEN** KDS 终端接收到配置更新事件 **THEN** 系统 **SHALL** KDS 终端立即应用新配置，无需重启应用

#### 具体要求

- [ ] 4.1 在 `main/app/service/setting/setting.go` 的 `SaveKitchenSetting` 方法中实现 WebSocket 推送
  - 使用 `websocket.PushClient` 推送配置更新事件
  - 事件类型：`websocket.UPDATE_CONFIG`
  - 目标：`websocket.SourceKitchen`（厨显端）
  - 数据：包含配置更新时间和配置内容
- [ ] 4.2 实现异步推送机制
  - 使用 `utils.Go` 异步推送，不阻塞主流程
  - 推送失败不影响配置保存
- [ ] 4.3 推送数据格式
  - 包含 `update_time` (int64): 更新时间戳
  - 包含 `config_type` (string): "kitchen_wait_time_color"
  - 包含 `config_data` (map): 配置数据

---

### Requirement 5: 数据迁移脚本

**用户故事**: 作为系统管理员，我想在版本上线时为所有门店初始化默认配置，以便于确保所有门店都有正确的初始配置。

#### 验收标准

1. **WHEN** 版本上线后执行迁移脚本 **THEN** 系统 **SHALL** 为所有门店初始化默认等待时长颜色配置
2. **WHEN** 门店已有配置 **THEN** 系统 **SHALL** 保持原配置不变，不覆盖
3. **WHEN** 门店无配置 **THEN** 系统 **SHALL** 初始化为默认配置：
   - 第一区间：0 分钟（黑色，#000000）
   - 第二区间：10 分钟（颜色从旧后台配置读取，如无则使用黄色 #ffff00）
   - 第三区间：20 分钟及以上（颜色从旧后台配置读取，如无则使用红色 #ff0000）
4. **WHEN** 迁移脚本重复执行 **THEN** 系统 **SHALL** 幂等性保证，可重复执行而不产生副作用

#### 具体要求

- [ ] 5.1 创建数据迁移脚本（PHP）
  - 文件路径：`admin/database/migrations/{YYYYMMDDHHMMSS}_add_kitchen_wait_time_color_config.php`
  - 脚本内容：
    - 查询所有门店的厨显设置（`setting` 表，key = `kitchen`）
    - 检查是否已有等待时长颜色配置（`wait_time_color_ranges` 字段）
    - 如无配置，则初始化默认配置
    - 如有配置，则保持原配置不变
- [ ] 5.2 实现迁移逻辑（PHP）
  - 使用 ThinkPHP Migration 框架
  - 读取所有门店配置
  - 检查并初始化默认配置
  - 记录迁移日志
- [ ] 5.3 实现幂等性保证
  - 检查配置是否已存在
  - 已存在则跳过，不重复初始化
- [ ] 5.4 从旧后台读取默认颜色
  - 查询旧后台 `setting` 表中的 `wait_color` 配置
  - 解析旧格式：`["red", "yellow"]` 或 `["yellow", "red"]`（第一个元素对应第二区间，第二个元素对应第三区间）
  - 转换为新格式并初始化：
    - 第一区间：0 分钟 → `#100A05`（黑色）
    - 第二区间：10 分钟 → `"red"` → `#E50028`，`"yellow"` → `#FFBE00`
    - 第三区间：20 分钟 → `"red"` → `#E50028`，`"yellow"` → `#FFBE00`

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

- [ ] URL 使用 snake_case 命名（如：`/api/v1/shop/setting/kitchen`）
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 错误信息使用多语言
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 配置存储在 `setting` 表的 `values` 字段（JSON 格式）
- [ ] 不新增表结构，复用现有 `setting` 表
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 配置保存响应时间 < 2 秒（包含旧后台同步）
- [ ] 配置同步到 KDS 终端延迟 < 5 秒
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] API 测试覆盖所有接口
- [ ] 集成测试覆盖核心流程（配置保存、同步、KDS 应用）
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有错误提示使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [ ] 所有 API 需要身份验证（JWT Token）
- [ ] 参数验证（时间区间、颜色格式）
- [ ] SQL 注入防护（使用参数化查询）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级（旧后台同步失败不影响新后台保存）
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 重试机制（旧后台同步最多重试 3 次）

---

## 验收标准

### 功能验收

1. **API 接口**: 获取和保存厨显设置接口正常工作，参数验证正确
2. **数据模型**: 等待时长颜色配置数据结构正确，保留旧格式字段，新增新格式字段
3. **权限管理**: 权限项正确创建，所有角色已赋权限，权限校验逻辑正确
4. **配置下发**: WebSocket 推送配置更新成功，KDS 终端能接收到配置更新事件
5. **数据迁移**: 版本上线后所有门店默认配置正确，已有配置不覆盖

### 测试验收

1. **单元测试**: 时间区间校验逻辑测试通过，颜色配置测试通过
2. **API 测试**: 获取和保存厨显设置接口测试通过
3. **集成测试**: 配置保存 → KDS 应用端到端流程测试通过
4. **手动测试**: 多场景验证（不同时间区间、不同颜色组合、权限控制场景）

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: 配置保存/获取 API 接口文档完整（Swagger）
3. **数据库文档**: 迁移脚本和表结构文档完整
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- WebSocket 推送使用 `websocket.PushClient` 方法
- HTTP 客户端调用使用 `utils.HttpPost` 或 `net/http`

### 业务约束

- 第一区间固定为 0 分钟，不可修改
- 第二、三区间时间范围：1-60 分钟整数
- 时间区间不可重叠，第三区间起点必须大于第二区间
- 颜色可重叠（三个区间可选择相同颜色）
- 配置同步到旧商家后台，确保新旧系统数据一致
- 版本上线后默认配置：
  - 第一区间：0 分钟（黑色 `#100A05`）
  - 第二区间：10 分钟（颜色从旧后台配置读取，如无则使用 `#FFBE00`）
  - 第三区间：20 分钟及以上（颜色从旧后台配置读取，如无则使用 `#E50028`）
- 旧格式 `wait_color` 只有 `"red"` 和 `"yellow"` 两个取值，第一个元素对应第二区间，第二个元素对应第三区间
- 新格式 `wait_time_color_ranges` 统一使用 RGB 格式（`#xxxxxx`），颜色值限定：
  - 黑色：`#100A05`
  - 黄色：`#FFBE00`
  - 红色：`#E50028`

### 资源约束

- 开发时间: 3-4 天
- Story Point: 5-8（待技术评审确认）

---

## 依赖关系

### 技术依赖

- `main/app/service/setting` - 设置服务
- `main/pkg/websocket` - WebSocket 推送服务
- `main/pkg/utils` - HTTP 客户端工具
- `main/app/dto` - 数据传输对象

### 服务依赖

- **Main → WebSocket**: gRPC 调用（推送配置更新）

### 业务依赖

- WebSocket 服务（配置下发到 KDS 终端）
- 角色权限系统（权限项配置、权限校验）

---

## 风险和缓解

### 风险 1: 时间区间校验逻辑复杂

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 后端实现严格的时间区间校验逻辑
- 编写详细的单元测试覆盖所有边界条件
- 前端和后端双重校验，确保数据合法性

### 风险 2: WebSocket 推送失败

**影响**: 中  
**概率**: 低  
**缓解措施**:

- WebSocket 推送异步执行，不阻塞主流程
- 推送失败不影响配置保存
- KDS 终端支持主动拉取配置（作为备用方案）

### 风险 3: 数据格式兼容性问题

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 实现新旧格式自动转换逻辑
- 保持向后兼容，支持旧格式数据
- 迁移脚本确保数据格式正确
- 充分测试新旧格式转换逻辑

### 风险 4: 默认配置迁移风险

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 版本上线时执行数据迁移脚本，为所有门店初始化默认配置
- 迁移脚本需幂等性保证，可重复执行
- 迁移后验证所有门店配置数据正确
- 从旧后台读取默认颜色配置

---

## 时间表

- **Phase 1 - API 接口和数据模型开发**: 1-1.5 天
- **Phase 2 - 权限管理集成开发**: 0.5 天
- **Phase 3 - 配置中心下发机制开发**: 0.5 天
- **Phase 4 - 数据迁移脚本开发**: 0.5 天
- **Phase 5 - 测试与优化**: 0.5-1 天
- **总计**: 3-4 天（SP = 5-8，待技术评审确认）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 相关代码

- `main/app/api/v1/shop/shop_setting.go` - Shop 设置 API
- `main/app/service/setting/setting.go` - 设置服务
- `main/app/dto/resp/setting/kitchen_setting.go` - 厨显设置 DTO
- `main/app/service/setting/default.go` - 默认设置
- `main/pkg/websocket/websocket.go` - WebSocket 推送
- `admin/app/shop/controller/setting/Terminal.php` - 旧后台厨显设置 API

### 前端需求文档

- [前端仓库: docs/shared/specs/active/story-shop-kds-wait-time-color-config/requirements.md](../../../../../ttpos-flutter/docs/shared/specs/active/story-shop-kds-wait-time-color-config/requirements.md)

### 外部参考

- [提案文档](../../../../team/proposals/2025-11/kds-wait-time-color-config.md) - 完整需求背景和方案

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: {团队/个人}  
**审核者**: {审核者}
