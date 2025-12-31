# 任务分解: Grab OAuth2 Token 接口实现

> **关联 Spec**: [requirements.md](./requirements.md)
> **状态**: Pending
> **负责人**: rikugun

## Phase 1: 配置与基础重构

- [ ] **Task 1.1**: 更新配置文件模板
  - 文件: `ttpos-bmp/app/ttpos-takeout/manifest/config/config.tpl.yaml`
  - 内容: 添加 `takeout.grab` 配置段 (client_id, client_secret, env)。
  - 验证: 生成的 `config.yaml` 包含新配置。

- [ ] **Task 1.2**: 定义 Config 结构体
  - 文件: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/client.go`
  - 内容: 确保 `ClientConfig` 结构体能映射新的配置项。

## Phase 2: 核心逻辑实现

- [ ] **Task 2.1**: 引入 Redis 依赖与常量定义
  - 文件: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/client.go`
  - 内容: 定义 Redis Key 常量，引入 `g.Redis()`。

- [ ] **Task 2.2**: 实现远程获取 Token 方法
  - 文件: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/client.go`
  - 内容: 提取原 `getAccessToken` 中的 HTTP 请求逻辑为独立私有方法 `fetchTokenFromGrab`。

- [ ] **Task 2.3**: 实现带缓存的 Token 获取逻辑
  - 文件: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/client.go`
  - 内容: 重写 `getAccessToken`，实现 Redis Get -> Fetch -> Redis Set 流程。

## Phase 3: 验证与清理

- [ ] **Task 3.1**: 集成测试
  - 内容: 启动服务，调用触发 Grab 逻辑的接口，验证 Redis 中是否有 Key 生成。
  - 验证: Token 能正常使用，API 调用成功。

- [ ] **Task 3.2**: 代码清理
  - 内容: 移除旧的内存缓存逻辑（`tokenMu`, `accessToken`, `tokenExpiry` 字段）。

