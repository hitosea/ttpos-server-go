# Grab 菜单获取回退 TTPOS 导出接口 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-18   |
| **目标版本** | v2.11.0 |
| **状态**   | ✅ 已创建 Spec   |
| **关联任务** | - |
| **关联 Spec** | [story-takeout-grab-menu-fallback-ttpos](../../../shared/specs/active/story-takeout-grab-menu-fallback-ttpos/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

当商家绑定 Grab 外卖平台时，可以选择"跳过导出菜单"。在这种情况下：

1. 商家选择跳过导出菜单后，TTPOS 不会主动推送菜单到 `channel_menu_snapshot.ttpos_menu_data` 字段
2. 当 Grab 调用 Partner Endpoint (`HandleGetMenu`) 获取菜单时，因为 `ttpos_menu_data` 为空会返回 "menu not found" 错误
3. 这导致跳过导出菜单的商家无法正常使用 Grab 外卖服务

**现有逻辑问题**：
```go
// grab_menu.go HandleGetMenu 当前逻辑
menuJSON, err := service.ChannelMenu().GetTtposMenu(ctx, shopUUID, string(consts.ProviderGrab))
if menuJSON == "" {
    // 直接返回 "menu not found" 错误
    return nil, gerror.NewCode(gcode.CodeNotFound, "menu not found")
}
```

### 业务价值

- 支持商家灵活选择菜单同步方式（主动推送 vs 按需拉取）
- 提升商家绑定 Grab 的用户体验
- 减少商家因菜单同步问题导致的服务中断
- 保持与 TTPOS 主系统的菜单数据一致性

### 目标用户

- [x] 商户管理员
- [x] 其他: TTPOS 后台运维人员

---

## 💡 解决方案概述

### 方案描述

在 `HandleGetMenu` 函数中增加回退逻辑：当本地菜单快照为空时，实时调用 TTPOS 主模块的 `/api/v1/takeout/menu/export` 接口获取菜单数据并返回。

请求时需要携带认证头 `X-TTPOS-SECRET`，认证方式参考已有的 Skootar 回调认证逻辑（使用 MD5 加密生成）。

### 核心功能点

1. **菜单快照检查**：优先从 `channel_menu_snapshot.ttpos_menu_data` 读取已缓存的菜单
2. **回退 TTPOS 导出**：当本地快照为空时，调用 main 模块 `/api/v1/takeout/menu/export` 接口
3. **认证机制**：请求携带 `X-TTPOS-SECRET` 头，使用 `shopUUID + callbackSecret` 的 MD5 加密值
4. **响应转换**：将 TTPOS 导出接口的响应转换为 Grab SDK 的 `GetMenuNewResponse` 格式
5. **可选缓存**：获取成功后可选择性保存到本地快照，避免重复调用

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口
- [ ] 数据模型
- [x] 业务逻辑
- [x] 第三方集成
- [ ] 其他: ________

---

## 📐 技术方案概要

### 涉及文件

| 文件路径 | 修改内容 |
| -------- | -------- |
| `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go` | `HandleGetMenu` 增加回退逻辑 |
| `ttpos-bmp/app/ttpos-takeout/internal/consts/consts.go` | 新增 TTPOS API 相关常量（可选） |
| `ttpos-bmp/app/ttpos-takeout/manifest/config/config.tpl.yaml` | 新增 TTPOS 主模块 endpoint 配置（可选） |

### 伪代码逻辑

```go
func (s *sGrabMenu) HandleGetMenu(ctx context.Context, partnerMerchantID string) (*grabfood.GetMenuNewResponse, error) {
    // 1. 解析 partnerMerchantID -> shopUUID
    shopUUID := g.NewVar(partnerMerchantID).Uint64()
    if shopUUID == 0 {
        return nil, gerror.NewCode(gcode.CodeInvalidParameter, "invalid partnerMerchantID format")
    }
    
    // 2. 优先读取本地菜单快照
    menuJSON, err := service.ChannelMenu().GetTtposMenu(ctx, shopUUID, string(consts.ProviderGrab))
    if err != nil {
        return nil, gerror.Wrap(err, "failed to get channel menu")
    }
    
    // 3. 如果本地快照为空，回退调用 TTPOS 导出接口
    if menuJSON == "" {
        g.Log().Infof(ctx, "[Grab] Local menu snapshot not found, fallback to TTPOS export API: shopUUID=%d", shopUUID)
        menuJSON, err = s.fetchMenuFromTTpos(ctx, shopUUID)
        if err != nil {
            g.Log().Errorf(ctx, "[Grab] Failed to fetch menu from TTPOS: shopUUID=%d, error=%v", shopUUID, err)
            return nil, gerror.Wrap(err, "failed to fetch menu from TTPOS")
        }
    }
    
    // 4. 仍然为空，返回菜单未找到错误
    if menuJSON == "" {
        return nil, gerror.NewCode(gcode.CodeNotFound, "menu not found")
    }
    
    // 5. 解析并返回
    // ...
}

func (s *sGrabMenu) fetchMenuFromTTpos(ctx context.Context, shopUUID uint64) (string, error) {
    // 1. 获取 TTPOS endpoint 配置
    ttposEndpoint := g.Cfg().MustGet(ctx, "app.ttposEndpoint").String()
    if ttposEndpoint == "" {
        return "", gerror.NewCode(gcode.CodeMissingConfiguration, "TTPOS endpoint not configured")
    }
    
    // 2. 构建请求
    url := fmt.Sprintf("%s/api/v1/takeout/menu/export", ttposEndpoint)
    reqBody := g.Map{
        "platform":    "grab",
        "companyUuid": shopUUID, // 需要确认是 shopUUID 还是 companyUUID
    }
    
    // 3. 生成认证头（参考 Skootar 回调认证）
    callbackSecret := g.Cfg().MustGet(ctx, "app.callbackSecret").String()
    auth, err := gmd5.EncryptString(fmt.Sprintf("%d%s", shopUUID, callbackSecret))
    if err != nil {
        return "", gerror.Wrap(err, "failed to generate auth header")
    }
    
    // 4. 发起请求
    resp, err := g.Client().
        SetHeader("X-TTPOS-SECRET", auth).
        ContentJson().
        Post(ctx, url, reqBody)
    if err != nil {
        return "", gerror.Wrap(err, "failed to call TTPOS export API")
    }
    defer resp.Close()
    
    // 5. 检查 HTTP 状态码
    if !resp.IsSuccess() {
        return "", gerror.Newf("TTPOS export API returned status %d: %s", resp.StatusCode, resp.ReadAllString())
    }
    
    // 6. 解析响应
    var result struct {
        Code int `json:"code"`
        Data struct {
            Platform string      `json:"platform"`
            MenuData interface{} `json:"menuData"`
        } `json:"data"`
        Message string `json:"message"`
    }
    if err := resp.UnmarshalJSON(&result); err != nil {
        return "", gerror.Wrap(err, "failed to parse TTPOS export API response")
    }
    if result.Code != 200 {
        return "", gerror.Newf("TTPOS export API error: code=%d, message=%s", result.Code, result.Message)
    }
    
    // 7. 序列化 menuData 返回
    menuJSON, err := json.Marshal(result.Data.MenuData)
    if err != nil {
        return "", gerror.Wrap(err, "failed to marshal menu data")
    }
    
    return string(menuJSON), nil
}
```

### 接口定义参考

**TTPOS 导出接口**：
- 路径：`POST /api/v1/takeout/menu/export`
- 认证：Header `X-TTPOS-SECRET`

**请求体**：
```json
{
  "platform": "grab",
  "companyUuid": 123456789,
  "categoryIds": [],      // 可选
  "sellingTimeIds": []    // 可选
}
```

**响应体**：
```json
{
  "code": 200,
  "data": {
    "platform": "grab",
    "menuData": { ... }  // Grab 格式菜单数据
  }
}
```

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

- **预计天数**: 1-2 天
- **预估 SP**: 3（待技术评审确认）

### 风险识别

**潜在风险**：
1. TTPOS 主模块接口超时可能导致 Grab GetMenu 请求失败
2. `X-TTPOS-SECRET` 认证机制需要与 main 模块协调确认
3. shopUUID 与 companyUUID 的映射关系需要确认

**缓解措施**：
1. 设置合理的请求超时时间（建议 10s），超时后返回友好错误
2. 与 main 模块开发人员确认认证方式，或复用现有认证机制
3. 查阅数据模型确认 shopUUID 与 companyUUID 的关系

---

## 🔗 相关资源

### 参考需求

- 现有实现: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
- 回调认证参考: `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/skootar.go` (`getCallBackAuth`)
- TTPOS 导出接口: `main/app/api/v1/takeout/menu_handler.go`

### 相关文档

- Grab 菜单集成文档: `docs/shared/integrations/grab/grab-menu-integration.md`
- 数据库加载完成报告: `docs/shared/integrations/grab/DATABASE_LOADING_COMPLETED.md`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     |        |           |
| 技术负责人   |        |           |
| 开发代表     |        |           |
| 测试代表     |        |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-takeout-grab-menu-fallback`
- [ ] 分配负责人：
- [ ] 目标 Sprint：

---

## 📝 附录

### User Story（初稿）

**作为** Grab 外卖平台  
**我想** 在调用 GetMenu 接口时获取商家菜单  
**以便于** 在 Grab App 上展示商家的外卖菜单

### AC 验收标准（初稿）

1. **WHEN** Grab 调用 GetMenu 且本地菜单快照存在 **THEN** 系统 **SHALL** 返回本地快照数据
2. **WHEN** Grab 调用 GetMenu 且本地菜单快照为空 **THEN** 系统 **SHALL** 调用 TTPOS 导出接口获取菜单
3. **IF** TTPOS 导出接口调用失败 **THEN** 系统 **SHALL** 返回 "menu not found" 错误并记录日志
4. **WHEN** 从 TTPOS 成功获取菜单 **THEN** 系统 **SHALL** 正确转换为 Grab SDK 格式返回

### 待确认问题

1. `X-TTPOS-SECRET` 认证头的生成规则是否与 `X-TTPOS-Callback-Auth` 一致？
2. 请求 TTPOS 接口时使用的是 `shopUUID` 还是需要转换为 `companyUUID`？
3. 从 TTPOS 获取的菜单是否需要缓存到本地 `ttpos_menu_data` 字段？
4. TTPOS 主模块的 endpoint 地址如何配置（通过 Nacos 还是配置文件）？

---

**版本**: v1.0.0  
**创建日期**: 2025-12-18  
**维护者**: rikugun

