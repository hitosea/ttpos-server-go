# Lineman Menu Sync V2 需求文档

> 本文档定义 Lineman Menu Sync V2 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                     |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/v2.13.1-takeout-lineman-menu-sync.md](../../../../team/proposals/2026-01/v2.13.1-takeout-lineman-menu-sync.md) |
| **创建日期**      | 2026-01-08                                                                                                               |
| **负责人**        | rikugun                                                                                                                  |
| **目标版本**      | v2.13.1                                                                                                                  |
| **目标 Sprint**   | Sprint TBD                                                                                                               |
| **涉及技术栈**    | [ ] Go (main/) [x] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                               |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过                   |
| **审核人**   | rikugun                  |
| **审核日期** | 2026-01-08               |
| **审核意见** | 需求明确，技术方案可行，复用现有表和Token管理，工作量合理 |

---

## 📋 概述

TTPOS 需要与泰国外卖平台 Lineman 进行菜单数据同步，支持商家在 Lineman 平台上展示和销售商品。本功能实现 Lineman Menu Sync V2 API 的集成，将 TTPOS 的菜单数据（分类、商品、规格、属性、价格等）自动同步到 Lineman 平台。

**核心价值**：
- **扩展销售渠道**：接入泰国主流外卖平台，帮助商家覆盖更多线上用户
- **提升运营效率**：菜单统一管理，自动同步，减少人工维护成本
- **保证数据一致性**：实时同步确保两个平台菜单信息一致，避免订单错误
- **增强产品竞争力**：支持主流外卖平台集成，提升 TTPOS 市场竞争力

## 🎯 产品对齐

该功能支持 TTPOS 的国际化战略和多渠道销售愿景：
- **市场拓展**：进入泰国外卖市场，支持商家多渠道运营
- **生态整合**：与主流外卖平台深度集成，构建开放生态
- **用户价值**：商家只需维护一份菜单，自动同步到多个平台

## 📝 用户故事

**作为** 商户管理员  
**我想** 在 TTPOS 中一键同步菜单到 Lineman 平台  
**以便于** 快速上线外卖业务，无需在多个平台重复维护菜单

---

## 功能需求

### Requirement 1: 菜单数据映射与转换

**用户故事**: 作为系统，我需要将 TTPOS 的菜单结构准确转换为 Lineman API V2 格式，以便确保数据兼容性和完整性。

#### 验收标准

1. **WHEN** 执行菜单同步 **THEN** 系统 **SHALL** 将 TTPOS categories 映射为 Lineman menuGroups
2. **WHEN** 转换商品数据 **THEN** 系统 **SHALL** 正确映射商品名称、价格、图片、状态等字段
3. **WHEN** 转换规格/加料组 **THEN** 系统 **SHALL** 将 modifierGroups 映射为 properties，modifiers 映射为 values
4. **WHEN** 处理多语言数据 **THEN** 系统 **SHALL** 提供 nameTranslation 对象（支持泰语、英语等）
5. **WHEN** 转换价格 **THEN** 系统 **SHALL** 将单位从元转换为分（乘以 100）
6. **IF** 商品有图片 **THEN** 系统 **SHALL** 将第一张图片映射到 photoUrl 字段
7. **WHEN** 处理商品状态 **THEN** 系统 **SHALL** 将 TTPOS 状态映射为 Lineman menuStatus（AVAILABLE/SUSPENDED/SOLD_OUT_TODAY）

#### 具体要求

- [x] 1.1 实现 TTPOS → Lineman 数据结构映射层
- [x] 1.2 支持分类（categories）→ 菜单组（menuGroups）转换
- [x] 1.3 支持商品/套餐（goods/packages）→ 菜单项（menuItems）转换
- [x] 1.4 支持规格组/加料组（modifierGroups）→ 属性（properties）转换
- [x] 1.5 支持属性值（modifiers）→ 选项值（values）转换
- [x] 1.6 支持多语言名称映射（中文 → 泰语 + 英语）
- [x] 1.7 支持价格单位转换（元 → 分）
- [x] 1.8 支持商品状态映射（可用、售罄、暂停销售）
- [x] 1.9 生成唯一 ID（格式：`TTPOS-CAT-{id}`, `TTPOS-ITEM-{id}`, `TTPOS-PROP-{id}`）

**字段映射表**（核心字段）：

| TTPOS 字段 | Lineman 字段 | 数据类型 | 说明 |
| --- | --- | --- | --- |
| `categories[].id` | `menuGroups[].id` | String(30) | 分类 ID |
| `categories[].name` | `menuGroups[].name.thai` | String | 分类名称（泰语） |
| `categories[].items[].id` | `menuItems[].id` | String(30) | 商品 ID |
| `categories[].items[].name` | `menuItems[].name.thai` | String | 商品名称（泰语） |
| `categories[].items[].price` | `menuItems[].price` | double | 价格（分） |
| `categories[].items[].photos[0]` | `menuItems[].photoUrl` | String | 商品图片 |
| `categories[].items[].status` | `menuItems[].menuStatus` | String | 商品状态 |
| `modifierGroups[].id` | `properties[].id` | String(160) | 属性组 ID |
| `modifierGroups[].name` | `properties[].name.thai` | String | 属性组名称 |
| `modifierGroups[].selectionRangeMin` | `properties[].min` | int | 最小选择数 |
| `modifierGroups[].selectionRangeMax` | `properties[].max` | int | 最大选择数 |
| `modifiers[].id` | `values[].id` | String(160) | 属性值 ID |
| `modifiers[].name` | `values[].name.thai` | String | 属性值名称 |
| `modifiers[].price` | `values[].price` | double | 属性值价格 |

**固定值映射**：
- `useSellingTime`: 固定 `false`（TTPOS 全时段销售，不支持时段限制）
- `salesChannelsAvailability.delivery`: 固定 `true`（配送渠道开启）
- `salesChannelsAvailability.pickup`: 固定 `true`（自提渠道开启）
- `properties[].type`: 根据 `selectionRangeMax` 判断（`>1` 为 `"2"` 复选框，`=1` 为 `"1"` 单选）

---

### Requirement 2: Lineman API Client 封装

**用户故事**: 作为系统，我需要封装 Lineman RESTful API 的调用逻辑，以便与 Lineman 服务进行可靠的通信。

#### 验收标准

1. **WHEN** 调用 Lineman API **THEN** 系统 **SHALL** 使用 HTTPS 协议和正确的 Endpoint
2. **WHEN** 发送请求 **THEN** 系统 **SHALL** 携带 `Authorization: Bearer {access_token}` 认证头
3. **WHEN** 发送请求 **THEN** 系统 **SHALL** 设置 `Content-Type: application/json` 头
4. **WHEN** API 返回 2xx **THEN** 系统 **SHALL** 解析响应并返回成功状态
5. **WHEN** API 返回 4xx/5xx **THEN** 系统 **SHALL** 记录错误并返回失败状态
6. **IF** 网络超时或连接失败 **THEN** 系统 **SHALL** 执行重试策略（最多 3 次）
7. **WHEN** 所有重试失败 **THEN** 系统 **SHALL** 记录详细错误日志并返回最终失败

#### 具体要求

- [x] 2.1 实现 HTTP Client 封装（使用 GoFrame g.Client()）
- [x] 2.2 支持 `PUT /v2/partners/{partnerId}/stores/{storeId}/menus` 接口
- [x] 2.3 **复用现有的 Access Token 管理**（`internal/logic/lineman_token`）
  - 使用 `service.LinemanToken().GetAuthorizationHeader(ctx)` 获取 Bearer Token
  - 无需重新实现 Token 获取、缓存、刷新逻辑
  - 现有实现已支持：Redis 缓存、双重检查锁、自动过期刷新
- [x] 2.4 实现请求重试机制（指数退避策略）
- [x] 2.5 实现超时控制（默认 30 秒）
- [x] 2.6 实现响应解析与错误码处理
- [x] 2.7 记录请求和响应日志（便于排查问题）
- [x] 2.8 支持自定义 partnerId 和 storeId

**API Endpoint**：
```
PUT https://api.lineman.co.th/v2/partners/{partnerId}/stores/{storeId}/menus
```

**请求头**：
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**响应格式**（成功）：
```json
{
  "status": "ok",
  "code": "SUCCESS",
  "menuSyncRequestId": "unique-request-id"
}
```

**响应格式**（失败）：
```json
{
  "status": "error",
  "code": "ERROR_CODE",
  "message": "Error description"
}
```

**重构后的使用示例**：

```go
package lineman

import (
    "context"
    "github.com/gogf/gf/v2/frame/g"
)

// sLineman Lineman 服务（重构后统一管理 Token 和菜单同步）
type sLineman struct {
    // Token 管理（从 lineman_token 迁移过来）
    cfgLoader *PartnerConfigLoader
    secretKey string
    tokenLock sync.Mutex
    
    // 菜单同步（新增）
    menuSyncClient *MenuSyncClient
}

// MenuSyncClient Lineman 菜单同步客户端
type MenuSyncClient struct {
    endpoint string
}

// SyncMenu 同步菜单到 Lineman
func (s *sLineman) SyncMenu(ctx context.Context, partnerId, storeId string, menuData interface{}) error {
    // 1. 获取 Authorization Header（复用同包内的 Token 管理）
    authHeader, err := s.GetAuthorizationHeader(ctx)
    if err != nil {
        return err
    }

    // 2. 构造 API URL
    url := cfg.Endpoint + "/v2/partners/" + partnerId + "/stores/" + storeId + "/menus"

    // 3. 发送 PUT 请求
    client := g.Client()
    resp, err := client.
        SetHeader("Authorization", authHeader).
        SetHeader("Content-Type", "application/json").
        ContentJson().
        Put(ctx, url, menuData)
    
    if err != nil {
        return err
    }
    defer resp.Close()

    // 4. 记录同步日志（使用 takeout_menu_log 表）
    logErr := s.recordMenuLog(ctx, partnerId, storeId, "FULL", "SUCCESS", resp)
    if logErr != nil {
        g.Log().Warningf(ctx, "记录菜单同步日志失败: %v", logErr)
    }
    
    return nil
}

// recordMenuLog 记录菜单同步日志到 takeout_menu_log 表
func (s *sLineman) recordMenuLog(ctx context.Context, partnerId, storeId, syncType, status string, resp interface{}) error {
    // 使用 DAO 层插入 takeout_menu_log
    // ...
}
```

---

### Requirement 3: 同步流程与策略

**用户故事**: 作为商户管理员，我想要触发菜单同步并查看同步状态，以便掌控菜单数据的一致性。

#### 验收标准

1. **WHEN** 商家在 Shop 后台点击"同步到 Lineman"按钮 **THEN** 系统 **SHALL** 立即开始全量同步
2. **WHEN** 商家修改菜单数据 **THEN** 系统 **SHALL** 支持增量同步（仅同步变更部分）
3. **WHEN** 同步成功 **THEN** 系统 **SHALL** 记录同步时间、请求 ID 和同步状态
4. **WHEN** 同步失败 **THEN** 系统 **SHALL** 记录错误原因和重试次数
5. **WHEN** 部分商品同步失败 **THEN** 系统 **SHALL** 标记失败的商品并提示商家
6. **IF** 同步过程中数据变更 **THEN** 系统 **SHALL** 使用最新数据重新同步

#### 具体要求

- [x] 3.1 实现全量同步（推送门店完整菜单结构）
- [x] 3.2 实现增量同步（仅推送变更的分类/商品/属性）
- [x] 3.3 记录同步状态（成功/失败/部分成功）
- [x] 3.4 记录同步日志（时间戳、请求 ID、错误信息）
- [x] 3.5 支持手动触发同步（Shop 后台按钮）
- [ ] 3.6 支持自动触发同步（商品变更后自动同步，可选）
- [ ] 3.7 支持定时同步（每天定时全量同步，可选）
- [x] 3.8 实现同步状态查询接口

---

### Requirement 4: 多语言字段处理

**用户故事**: 作为系统，我需要为 Lineman API 提供符合要求的 nameTranslation 字段，以便平台正确展示菜单信息。

#### 验收标准

1. **WHEN** 生成 menuGroups/menuItems/properties/values **THEN** 系统 **SHALL** 提供 nameTranslation 对象
2. **IF** TTPOS 中已有多语言字段（泰语、英语） **THEN** 系统 **SHALL** 直接使用已有翻译
3. **IF** TTPOS 中仅有中文 **THEN** 系统 **SHALL** 使用中文填充 nameTranslation（降级处理）
4. **WHEN** 处理描述字段 **THEN** 系统 **SHALL** 同样提供 descriptionTranslation 对象

#### 具体要求

- [x] 4.1 在 `data_mapper.go` 中实现多语言字段映射逻辑
- [x] 4.2 优先从 TTPOS 多语言字段读取泰语和英语翻译
- [x] 4.3 如无翻译，使用中文作为默认值（`thai: 中文, english: 中文`）
- [x] 4.4 支持以下字段的多语言：
  - `menuGroups[].name` → `nameTranslation`
  - `menuItems[].name` → `nameTranslation`
  - `menuItems[].description` → `descriptionTranslation`
  - `properties[].name` → `nameTranslation`
  - `values[].name` → `nameTranslation`

**实现示例**：
```go
// data_mapper.go 中的多语言处理
func buildNameTranslation(nameCN, nameTH, nameEN string) map[string]string {
    // 如果有泰语和英语翻译，使用翻译
    if nameTH != "" && nameEN != "" {
        return map[string]string{
            "thai":    nameTH,
            "english": nameEN,
        }
    }
    
    // 否则使用中文作为默认值（降级处理）
    return map[string]string{
        "thai":    nameCN,
        "english": nameCN,
    }
}
```

---

### Requirement 5: Partner ID 和 Store ID 管理

**用户故事**: 作为商户管理员，我需要配置 Lineman 的 Partner ID 和 Store ID，以便系统识别我的门店身份。

#### 验收标准

1. **WHEN** 商家首次配置 Lineman 集成 **THEN** 系统 **SHALL** 要求输入 Partner ID 和 Store ID
2. **WHEN** 配置完成 **THEN** 系统 **SHALL** 验证 ID 的有效性（调用 Lineman API 测试连接）
3. **WHEN** 配置成功 **THEN** 系统 **SHALL** 保存配置到数据库并加密存储
4. **IF** ID 配置错误 **THEN** 系统 **SHALL** 显示错误提示并拒绝保存

#### 具体要求

- [x] 5.1 在 Shop 后台添加"Lineman 配置"页面（或在现有外卖配置页面扩展）
- [x] 5.2 支持输入 Partner ID 和 Store ID
- [x] 5.3 支持测试连接（验证 ID 有效性）
- [x] 5.4 **使用 `takeout_shop_provider_cfg` 表保存配置**
  - `provider_name` = `'lineman'`
  - `provider_merchant_id` = `'{partnerId}_{storeId}'` 或 JSON 格式
  - `provider_shop_status` = `'ACTIVE'` / `'INACTIVE'` / `'SYNCING'` / `'FAILED'`
- [x] 5.5 Access Token 由 Token 管理模块自动管理（无需手动存储）

---

### Requirement 6: 错误处理与监控

**用户故事**: 作为运营人员，我需要监控菜单同步的状态和错误，以便及时发现和解决问题。

#### 验收标准

1. **WHEN** 同步失败 **THEN** 系统 **SHALL** 记录详细错误日志（错误码、错误信息、请求参数）
2. **WHEN** 网络异常 **THEN** 系统 **SHALL** 执行重试并记录重试次数
3. **WHEN** API 限流 **THEN** 系统 **SHALL** 实现退避重试策略
4. **WHEN** 同步失败超过阈值 **THEN** 系统 **SHALL** 发送告警通知（邮件/短信/系统通知）
5. **WHEN** 商家查询同步历史 **THEN** 系统 **SHALL** 展示最近 N 次同步记录

#### 具体要求

- [x] 6.1 **使用 `takeout_menu_log` 表记录同步日志**
  - `provider_name` = `'lineman'`
  - `merchant_id` = `'{partnerId}_{storeId}'`
  - `sync_type` = `'FULL'` / `'PARTIAL'`
  - `status` = `'QUEUED'` / `'PROCESSING'` / `'SUCCESS'` / `'FAIL'`
  - `request_id` = Lineman 返回的 `menuSyncRequestId`
  - `menu_snapshot` = 菜单数据 JSON 快照
  - `error_code` / `error_msg` = 错误信息
- [x] 6.2 实现错误码映射（Lineman 错误码 → TTPOS 错误码）
- [ ] 6.3 实现告警机制（失败率超过阈值时触发，可选）
- [x] 6.4 支持同步历史查询（通过 `takeout_menu_log` 表查询）
- [x] 6.5 记录关键指标（同步成功率、平均响应时间、失败原因分布）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Logic → Client 分层（BMP 模块架构）
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Logic 和 Client 应独立且可复用
- **依赖管理**: Logic 只能依赖 Client 接口，不能直接依赖具体实现
- **遵循规范**:
  - `.cursor/rules/go-bmp.mdc` - Go BMP 微服务规范
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go 代码开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

**代码结构重构方案**：
```
ttpos-bmp/app/ttpos-takeout/
├── internal/
│   ├── logic/
│   │   └── lineman/                  # 【重构】统一 Lineman 逻辑
│   │       ├── token.go              # 【迁移自 lineman_token/】OAuth Token 管理
│   │       ├── token_config.go       # 【迁移自 lineman_token/】配置加载
│   │       ├── menu_sync.go          # 【新增】菜单同步业务逻辑
│   │       └── data_mapper.go        # 【新增】数据映射转换
│   ├── client/
│   │   └── lineman/                  # 【新增】API Client
│   │       ├── menu_sync_client.go   # Menu Sync API Client
│   │       └── retry.go              # 重试策略
│   ├── dao/                          # 【复用】数据访问层
│   │   ├── shop_provider_cfg.go     # 门店配置 DAO（已存在）
│   │   └── menu_log.go              # 菜单日志 DAO（已存在）
│   └── model/
│       ├── entity/
│       │   ├── shop_provider_cfg.go  # 【复用】门店配置实体
│       │   └── menu_log.go           # 【复用】菜单日志实体
│       └── dto/lineman/
│           ├── menu_sync_request.go  # 菜单同步请求 DTO
│           └── menu_sync_response.go # 菜单同步响应 DTO
```

**重构步骤**：
1. ✅ **迁移 Token 管理**：
   - 将 `internal/logic/lineman_token/` 下的文件移动到 `internal/logic/lineman/`
   - 重命名：`lineman_token.go` → `token.go`
   - 更新包名：`package lineman_token` → `package lineman`
   - 更新服务注册：`service.RegisterLinemanToken()` → `service.RegisterLineman()`

2. ✅ **复用现有表**：
   - `takeout_shop_provider_cfg` - 存储 Lineman 门店配置
   - `takeout_menu_log` - 记录菜单同步日志

3. ✅ **新增菜单同步逻辑**：
   - 在 `logic/lineman/` 下新增菜单同步相关代码
   - 复用同一目录下的 Token 管理逻辑
   
4. ✅ **多语言处理**：
   - 直接在 `data_mapper.go` 中处理多语言字段
   - 使用中文作为默认值填充 `nameTranslation` 字段
   - 无需单独的翻译服务模块

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/lineman/menu_sync`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

**复用现有表，不创建新表**：

1. **门店配置**：使用现有的 `takeout_shop_provider_cfg` 表
```sql
-- 已存在的表结构
CREATE TABLE IF NOT EXISTS `takeout_shop_provider_cfg` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一标识',
    `shop_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '门店UUID',
    `provider_name` VARCHAR(32) NOT NULL DEFAULT 'grab' COMMENT '第三方名称，如 grab, lineman',
    `provider_merchant_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '第三方商户ID（Lineman的partnerId_storeId）',
    `provider_shop_status` ENUM('INACTIVE','ACTIVE','SYNCING','FAILED') NOT NULL DEFAULT 'INACTIVE' COMMENT '门店集成状态',
    `created_at` INT NOT NULL DEFAULT 0 COMMENT '创建时间',
    `updated_at` INT NOT NULL DEFAULT 0 COMMENT '更新时间',
    `deleted_at` INT NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_shop_provider` (`shop_uuid`, `provider_name`),
    KEY `idx_provider_name` (`provider_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='门店第三方集成配置';

-- Lineman 使用方式：
-- provider_name = 'lineman'
-- provider_merchant_id = '{partnerId}_{storeId}' 或存储为 JSON
```

2. **菜单同步日志**：使用现有的 `takeout_menu_log` 表
```sql
-- 已存在的表结构
CREATE TABLE IF NOT EXISTS `takeout_menu_log` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
    `uuid` varchar(100) NOT NULL COMMENT '唯一ID',
    `merchant_id` varchar(100) NOT NULL COMMENT '商户ID（partnerId_storeId）',
    `provider_name` varchar(50) NOT NULL COMMENT '渠道: grab, lineman',
    `sync_type` varchar(50) DEFAULT 'FULL' COMMENT '同步类型: FULL, PARTIAL, NOTIFY',
    `request_id` varchar(100) DEFAULT NULL COMMENT '请求ID (来自第三方平台)',
    `status` varchar(20) DEFAULT NULL COMMENT '同步状态: QUEUED, SUCCESS, FAIL, PROCESSING',
    `menu_snapshot` json DEFAULT NULL COMMENT '菜单快照(JSON)',
    `error_code` varchar(50) DEFAULT NULL COMMENT '错误代码',
    `error_msg` text DEFAULT NULL COMMENT '错误信息',
    `created_at` int NOT NULL DEFAULT 0 COMMENT '创建时间',
    `updated_at` int NOT NULL DEFAULT 0 COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uuid` (`uuid`),
    KEY `idx_merchant` (`merchant_id`, `provider_name`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单同步记录表';

-- Lineman 使用方式：
-- provider_name = 'lineman'
-- merchant_id = '{partnerId}_{storeId}'
-- request_id = Lineman API 返回的 menuSyncRequestId
```

**数据库规范**：
- [x] 必须包含: `id`, `uuid`, `created_at`, `updated_at`, `deleted_at`
- [x] 时间字段使用 int 类型，_at 结尾，默认值 0
- [x] UUID 字段使用 bigint unsigned 或 varchar(100)
- [x] 表名使用 takeout_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [x] 本地响应时间 < 200ms（不含 Lineman API 调用）
- [x] Lineman API 调用超时设置为 30 秒
- [x] 缓存 Access Token（Redis，过期时间根据 Lineman 规定）
- [x] 使用连接池优化 HTTP 请求性能
- [ ] 大菜单（>100 商品）考虑分批同步

### 测试要求

- [x] Logic 层测试覆盖率 ≥ 70%
- [x] Client 层测试覆盖率 ≥ 80%
- [x] 数据映射逻辑测试覆盖率 100%（关键逻辑）
- [x] Mock Lineman API 进行集成测试
- [ ] 在 Sandbox 环境进行真实 API 测试
- [x] 参考: `ttpos-bmp/.cursor/rules/go-rules.mdc` - 测试规范

### 国际化要求

- [x] 所有错误提示使用多语言实现（系统错误提示）
- [x] 菜单数据的多语言字段使用中文作为默认值（降级处理）
- [x] 参考: `ttpos-bmp/i18n/` - 国际化配置

### 安全要求

- [x] Access Token 加密存储（AES-256）
- [x] 所有 API 需要身份验证
- [x] HTTPS 通信（强制）
- [x] 参数验证（防止注入攻击）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级（记录失败日志，不影响主流程）
- [x] 事务管理（同步日志记录与业务操作原子性）
- [x] 错误日志记录（使用 glog 或 GoFrame Logger）
- [x] 故障恢复机制（支持手动重试）

---

## 验收标准

### 功能验收

1. **数据映射正确性**: TTPOS 菜单数据能够正确转换为 Lineman API V2 格式，所有必填字段完整
2. **API 调用成功**: 能够成功调用 Lineman API 并接收到正确的响应
3. **同步状态可追溯**: 商家可以查看同步历史和状态，了解同步成功/失败原因
4. **多语言支持**: 商品名称支持泰语和英语展示（至少支持中文降级）
5. **错误处理完善**: 网络异常、API 错误、数据异常等情况都有明确的错误提示和日志
6. **性能达标**: 同步 100 个商品的完整菜单在 5 秒内完成（不含网络延迟）

### 测试验收

1. **单元测试**: Logic 层和 Client 层覆盖率达标
2. **集成测试**: Mock Lineman API 测试通过
3. **真实环境测试**: 在 Lineman Sandbox 环境测试通过（如有）
4. **边界测试**: 大菜单（>100 商品）、空菜单、特殊字符等场景测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确（使用 `/spec-design` 创建）
2. **API 文档**: Lineman API Client 的使用文档完整
3. **数据库文档**: 迁移脚本和表结构文档完整
4. **配置文档**: Partner ID / Store ID 配置指南完整

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 遵循 BMP 模块的分层架构（Controller → Logic → Client）
- 不使用 panic，返回 error
- gRPC 服务（如需要）必须注册到 Nacos
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`

#### HTTP Client

- 使用 Go 标准库 net/http 或第三方库（如 resty）
- 支持超时控制、重试、连接池
- 支持 Bearer Token 认证
- 记录请求和响应日志

### 业务约束

- **时段销售**: TTPOS 不支持时段销售，useSellingTime 固定为 false
- **销售渠道**: TTPOS 统一开启配送和自提，salesChannelsAvailability 固定为 {delivery: true, pickup: true}
- **统一定价**: TTPOS 所有渠道统一价格，不支持 salesChannelsPrice
- **多语言**: 如商家未提供翻译，使用中文作为默认值

### 资源约束

- 开发时间: 4 天（单人，由于复用现有 Token 管理而减少）
- Story Point: 3-5（待技术评审确认，从 5-8 调整为 3-5）

---

## 依赖关系

### 技术依赖

- `github.com/gogf/gf/v2` - GoFrame 框架
- `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman` - **Lineman 统一逻辑（重构后）**
- `ttpos-bmp/app/ttpos-takeout/internal/service` - 服务接口
- `ttpos-bmp/app/ttpos-takeout/internal/dao` - 数据访问层（复用 shop_provider_cfg, menu_log）
- `ttpos-bmp/app/ttpos-takeout/internal/model/entity` - 实体模型（复用现有表结构）

### 服务依赖

- **Lineman API**: 外部 RESTful API（https://api.lineman.co.th）
- **TTPOS Main/Admin**: 菜单数据来源（通过 BMP 内部接口获取）
- **Redis**: Access Token 缓存
- **MySQL**: 同步日志和配置存储

### 业务依赖

- 商家必须先在 Lineman 平台完成入驻并获取 Partner ID 和 Store ID
- 商家必须在 TTPOS Shop 后台配置 Lineman 集成信息
- 菜单数据必须在 TTPOS 中已完整录入（分类、商品、规格等）

---

## 风险和缓解

### 风险 1: Lineman API 文档不完整

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 联系 Lineman 技术支持，获取完整 API 文档、错误码表、最佳实践指南
- 申请 Sandbox 测试环境和测试账号
- 参考其他 Partner 的集成经验（如有公开资料）
- 实现灵活的错误处理和降级策略，应对未知错误

### 风险 2: 认证机制不明确

**影响**: ~~高~~ → **低**（已解决）  
**概率**: ~~中~~ → **无**  
**缓解措施**:

- ✅ **已有现成实现**：`internal/logic/lineman_token` 已实现完整的 OAuth 2.0 认证流程
- ✅ **Token 管理完善**：支持 client_credentials 授权模式、Redis 缓存、自动刷新
- ✅ **生产验证**：该 Token 管理逻辑已在生产环境使用，稳定可靠
- 📝 **开发建议**：直接复用 `service.LinemanToken().GetAuthorizationHeader(ctx)` 即可

### 风险 3: 数据映射不兼容

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 设计灵活的映射层，支持配置化扩展
- 对无法映射的字段做降级处理（如特殊属性组类型跳过）
- 记录映射失败的详细日志，便于后续优化
- 在测试阶段充分验证各种菜单场景（简单菜单、复杂菜单、特殊字符等）

### 风险 4: 多语言字段问题

**影响**: ~~中~~ → **低**（简化方案）  
**概率**: ~~高~~ → **低**  
**缓解措施**:

- ✅ **简化方案**：不集成自动翻译服务，使用中文作为默认值
- ✅ **降级处理**：如商家未提供泰语/英语翻译，直接使用中文填充 nameTranslation
- 📝 **后续优化**：可在 Shop 后台提供"菜单翻译"功能，让商家手动维护多语言菜单
- 📝 **Lineman 支持**：确认 Lineman 平台是否支持中文显示（如支持则无需翻译）

### 风险 5: 网络稳定性和 API 限流

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 实现请求重试、超时控制、熔断机制
- 实现同步队列，控制并发数，避免触发限流
- 监控 API 限流情况并实现退避重试
- 异步同步避免阻塞主流程，不影响商家正常操作

### 风险 6: 测试环境缺失

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 优先向 Lineman 申请 Sandbox 环境
- 如无法获取，在 Mock Server 上模拟 API 响应进行测试
- 使用 Postman 或类似工具手动测试 API
- 在生产环境上线前进行小范围灰度测试

---

## 时间表

- **Phase 1 - 代码重构**: 0.5 天
  - 迁移 `logic/lineman_token/` 到 `logic/lineman/`
  - 更新包名和服务注册
  - 更新依赖引用
  
- **Phase 2 - 数据映射与 Client 封装**: 1.5 天
  - 实现 TTPOS → Lineman 数据结构映射
  - 实现 Lineman Menu Sync API Client 封装
  - 集成现有 Token 管理（同包内直接调用）
  - 实现重试机制
  
- **Phase 3 - 同步流程与日志记录**: 1 天（优化）
  - 实现菜单同步业务逻辑
  - 集成 `takeout_shop_provider_cfg` 配置管理
  - 集成 `takeout_menu_log` 日志记录
  
- **Phase 4 - 测试与联调**: 1 天
  - 单元测试和集成测试
  - 与 Lineman Sandbox 环境联调（如有）
  - 边界测试和性能测试
  
- **总计**: 4 天（SP = 3-5，复用现有表和 Token 管理）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go 代码开发规范
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- **Lineman API 定义**: [Google Sheets - Menu Sync API](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=352934607#gid=352934607)
- **Lineman 开发者文档**: [待补充]
- **Lineman Partner Portal**: [待补充]

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-08.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-08  
**作者**: rikugun  
**审核者**: 待审核
