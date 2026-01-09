# TTPOS HTTP Client 工具类抽取 任务分解

> 本文档定义 TTPOS HTTP Client 工具类抽取的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 3
**已完成**: 3
**进行中**: -
**完成率**: 100%

---

## Phase 1: 工具类开发

- [x] 1.1 创建 TTPOS Client 工具类

  - File: `ttpos-bmp/app/ttpos-takeout/utility/ttpos_client.go`
  - Purpose: 提供统一的 TTPOS HTTP Client 工厂方法
  - Requirements: 1.1, 1.2, 1.3, 1.4, 2.1, 2.2
  - Leverage: 
    - 参考实现: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/erpnext.go` (GetClient)
    - 认证生成: `ttpos-bmp/app/ttpos-takeout/utility/ttpos_auth.go`
  - 实现内容:
    - `GetTtposClient(ctx)`: 基础 Client，自动配置 prefix、超时、ContentJson、dump 中间件
    - `GetTtposClientWithAuth(ctx, identifier)`: 带认证头的 Client

---

## Phase 2: 重构现有代码

- [x] 2.1 重构 fetchMenuFromTTpos 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 使用新工具类替换原有的 Client 创建逻辑
  - Requirements: 1.1, 2.1
  - Leverage: Task 1.1 的工具类
  - 修改内容:
    - 移除手动创建 `g.Client()` 的代码
    - 使用 `utility.GetTtposClientWithAuth()` 获取 Client
    - 移除手动设置 header、timeout、ContentJson 的代码

---

## Phase 3: 配置更新

- [x] 3.1 更新配置模板

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/config/config.tpl.yaml`
  - Purpose: 添加 dump 开关配置项
  - Requirements: 1.4
  - 配置项:
    ```yaml
    app:
      ttpos-client:
        dump: false  # 是否打印请求/响应详情
    ```

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [x] 所有任务标记为 `[x]`
- [x] Go 代码通过 `go fmt` 和 `go vet`
- [x] 功能测试通过

### 功能完整性

- [x] requirements.md 中的所有需求已满足
- [x] design.md 中的设计已实现

### 规范遵循

- [x] 遵循 `.cursor/rules/go-bmp.mdc`

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-18  
**维护者**: 开发组

