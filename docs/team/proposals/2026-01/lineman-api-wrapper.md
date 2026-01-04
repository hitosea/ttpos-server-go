# Lineman API 包装实现 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2026-01-04   |
| **目标版本** | - |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | -      |

---

## 🎯 背景和动机

### 问题描述

当前 ttpos-takeout 模块已经集成了 GrabFood 和 Skootar 等第三方外卖平台的 API 包装服务，但尚未实现 Lineman 平台的 API 集成。Lineman 是泰国重要的外卖配送平台之一，为了完善系统的外卖平台覆盖，需要实现 Lineman API 的包装层。

根据 `docs/others/lineman/API Spec (Master).xlsx` 中的 API 规格说明，需要在 `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman` 目录下实现 Lineman API 的包装服务。

### 业务价值

- 扩展外卖平台覆盖范围，支持 Lineman 平台订单处理
- 统一第三方平台 API 调用接口，降低业务层集成复杂度
- 提升系统对泰国市场的支持能力
- 为未来更多第三方平台集成提供参考实现

### 目标用户

- [ ] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: 系统内部服务（不对外直接提供服务）

---

## 💡 解决方案概述

### 方案描述

参考现有的 GrabFood 和 Skootar API 包装实现模式，在 `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman` 目录下实现 Lineman API 的包装服务。

实现范围：
1. **仅实现到 Service 层**：不对外提供 gRPC 服务，仅在 `internal/logic/lineman` 中实现业务逻辑
2. **参考现有实现**：遵循 Grab 和 Skootar 的实现模式和代码风格
3. **配置管理**：在配置文件中添加 Lineman 相关配置（API Key、Endpoint 等）
4. **错误处理**：统一使用 GoFrame 的 `gerror` 进行错误处理
5. **日志记录**：使用 GoFrame 的日志系统记录关键操作

### 核心功能点

根据 API Spec 文档，预计需要实现以下功能模块：

1. **认证与配置管理**
   - 配置读取（API Key、Endpoint、环境等）
   - 认证信息管理
   - 请求签名/Token 管理

2. **订单管理 API**
   - 接收订单（Webhook 处理）
   - 接受/拒绝订单
   - 取消订单
   - 更新订单状态
   - 标记订单准备完成

3. **菜单管理 API**
   - 获取菜单
   - 同步菜单
   - 更新菜单状态

4. **门店管理 API**
   - 获取门店状态
   - 暂停/恢复门店营业
   - 更新门店信息

5. **通用工具**
   - HTTP 请求封装
   - 响应解析
   - 错误处理
   - 日志记录

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端
- [x] 其他: 内部服务（ttpos-takeout 模块）

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（内部 Service 层）
- [x] 数据模型（DTO 定义）
- [x] 业务逻辑（Logic 层实现）
- [x] 第三方集成（Lineman API）
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [x] **中**：需要实现完整的 API 包装层，包含认证、订单、菜单、门店等多个模块

### 工作量预估

**基于现有 Grab 和 Skootar 实现参考**：

- **预计天数**: 3-5 天
- **预估 SP**: 5-8（待技术评审确认）

**工作量分解**：
1. API 规格文档分析和 DTO 定义：0.5 天
2. 配置管理和认证模块：0.5 天
3. 订单管理 API 实现：1-1.5 天
4. 菜单管理 API 实现：0.5-1 天
5. 门店管理 API 实现：0.5 天
6. 单元测试和集成测试：1 天

### 风险识别

**潜在风险**：
1. API Spec 文档为 Excel 格式，可能存在理解偏差
2. Lineman API 的认证机制可能与 Grab/Skootar 不同
3. Webhook 签名验证机制需要详细了解
4. 缺少 Lineman 官方 SDK，需要手动实现 HTTP 请求

**缓解措施**：
1. 仔细阅读 API Spec 文档，必要时与产品/业务确认
2. 参考 Grab 的认证实现，设计灵活的认证接口
3. 预留 Webhook 签名验证的扩展接口
4. 封装通用的 HTTP 请求工具，便于维护和测试

---

## 🔗 相关资源

### 参考需求

- 类似功能: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/`
- 类似功能: `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/`

### 相关文档

- API 规格文档: `docs/others/lineman/API Spec (Master).xlsx`
- GoFrame 官方文档: https://goframe.org.cn
- ttpos-bmp 开发规范: `ttpos-bmp/.cursor/rules/go-rules.mdc`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | -      |           |
| 技术负责人   | -      |           |
| 开发代表     | rikugun |           |
| 测试代表     | -      |           |
| UI/UX 设计师 | -      |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`feature-takeout-lineman-api-wrapper`
- [ ] 分配负责人：rikugun
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 外卖系统开发者  
**我想** 实现 Lineman API 的包装服务  
**以便于** 业务层能够统一调用 Lineman 平台的订单、菜单、门店等功能

### AC 验收标准（初稿）

1. **WHEN** 调用 Lineman 订单接收 API **THEN** 系统 **SHALL** 正确解析订单数据并返回结构化响应
2. **WHEN** 调用接受/拒绝订单 API **THEN** 系统 **SHALL** 向 Lineman 发送正确的请求并处理响应
3. **WHEN** 调用菜单同步 API **THEN** 系统 **SHALL** 正确构造菜单数据并发送到 Lineman
4. **WHEN** 调用门店状态管理 API **THEN** 系统 **SHALL** 正确更新门店状态并记录日志
5. **IF** API 调用失败 **THEN** 系统 **SHALL** 使用 `gerror` 包装错误并记录详细日志
6. **IF** 配置缺失或无效 **THEN** 系统 **SHALL** 在启动时或首次调用时返回明确的错误信息

### 技术实现要点

#### 1. 目录结构

```
ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/
├── lineman.go          # 主服务入口，配置管理
├── config.go           # 配置结构定义
├── auth.go             # 认证相关（如需要）
├── order.go            # 订单管理 API
├── menu.go             # 菜单管理 API
├── store.go            # 门店管理 API
└── lineman_test.go     # 单元测试
```

#### 2. 配置定义

```go
// internal/model/conf/lineman.go
type Lineman struct {
    Endpoint     string `json:"endpoint"`     // API 端点
    ApiKey       string `json:"apiKey"`       // API Key
    SecretKey    string `json:"secretKey"`    // Secret Key（如需要）
    Environment  string `json:"environment"`  // staging/production
    Timeout      int    `json:"timeout"`      // 请求超时时间（秒）
}
```

#### 3. 服务注册模式

```go
// internal/logic/lineman/lineman.go
var (
    Lineman = new(sLineman)
    config  *conf.Lineman
)

type sLineman struct{}

func init() {
    service.RegisterLineman(Lineman)
}

func (s *sLineman) MustConf() *conf.Lineman {
    // 配置读取逻辑（参考 Grab 实现）
}
```

#### 4. DTO 定义位置

- 请求/响应数据结构：`internal/model/dto/lineman/`
- 根据 API Spec 定义对应的结构体
- 使用 JSON 标签进行序列化/反序列化

#### 5. 错误处理规范

```go
import "github.com/gogf/gf/v2/errors/gerror"

// 统一错误处理
if err != nil {
    return gerror.Wrapf(err, "[Lineman] 操作失败: %s", operation)
}
```

#### 6. 日志记录规范

```go
import "github.com/gogf/gf/v2/frame/g"

// 关键操作日志
g.Log().Infof(ctx, "[Lineman] 订单已接受: %s", orderID)
g.Log().Errorf(ctx, "[Lineman] API 调用失败: %v", err)
```

### 线框图/原型（可选）

N/A - 内部服务实现，无 UI 界面

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2026-01-04  
**维护者**: rikugun  
**相关规范**: `ttpos-bmp/.cursor/rules/go-rules.mdc`, `ttpos-bmp/.cursor/rules/go-ttpos-takeout.mdc`

