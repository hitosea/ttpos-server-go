# BMP Nacos 多实例支持优化 任务分解

> 本文档定义 BMP Nacos 多实例支持优化的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-2 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 5  
**已完成**: 4  
**进行中**: -  
**完成率**: 80%

**说明**：由于 GoFrame `nacos.New()` 已支持多实例，无需实现客户端池、健康检查和负载均衡逻辑，任务数量大幅减少。

---

## Phase 1: 技术预研（已完成）

- [x] 1.1 调研 GoFrame Nacos 客户端多实例支持能力

  - File: `ttpos-bmp/internal/pkg/nacos/service/rpc_service.go`
  - Purpose: 确认 GoFrame Nacos 客户端是否支持多实例配置
  - Requirements: Requirement 1
  - Leverage: [GoFrame Nacos 源码](https://github.com/gogf/gf/tree/master/contrib/registry/nacos)
  - Result: ✅ **已确认 GoFrame `nacos.New()` 支持逗号分隔的多实例地址**
  - Code Example:
    ```go
    // GoFrame nacos.New() 源码已支持：
    nacos.New("host1:port1,host2:port2,host3:port3")
    // 会自动解析逗号分隔的地址，创建多实例连接
    // Nacos SDK 自动处理故障转移和负载均衡
    ```

---

## Phase 2: 配置格式扩展

- [x] 2.1 修改 GetNacosAddress() 函数支持多实例

  - File: `ttpos-bmp/internal/pkg/nacos/service/rpc_service.go`
  - Purpose: 扩展配置解析，支持返回逗号分隔的多实例地址
  - Requirements: Requirement 1, Requirement 4
  - Leverage: 现有 `GetNacosAddress()` 函数，[GoFrame Nacos 源码](https://github.com/gogf/gf/tree/master/contrib/registry/nacos)
  - Prompt: Role: Go Developer | Task: 修改 GetNacosAddress() 函数，支持返回逗号分隔的多实例地址 | Context: 参考 design.md 中的实现方案，优先使用 addresses 配置，兼容 ip+port 配置，返回逗号分隔的地址字符串 | Restrictions: 遵循 .cursor/rules/go-bmp.mdc，保持向后兼容 | Success: 函数能够正确解析多实例和单实例配置，返回逗号分隔的地址字符串
  - Code Example:
    ```go
    func GetNacosAddress(ctx context.Context) string {
        // 优先使用 addresses 配置（多实例）
        if addresses := g.Cfg().MustGetWithEnv(ctx, "nacos.server.addresses", "").String(); len(addresses) > 0 {
            // 验证地址格式
            addressList := gstr.Split(addresses, ",")
            for _, addr := range addressList {
                addr = gstr.Trim(addr)
                if !isValidAddress(addr) {
                    g.Log().Warnf(ctx, "无效的 Nacos 地址格式，将忽略: %s", addr)
                    continue
                }
            }
            g.Log().Infof(ctx, "使用多实例 Nacos 配置: %s", addresses)
            return addresses
        }
        
        // 兼容单实例配置
        nacosServerIp := g.Cfg().MustGetWithEnv(ctx, "nacos.server.ip")
        nacosServerPort := g.Cfg().MustGetWithEnv(ctx, "nacos.server.port")
        address := fmt.Sprintf("%v:%v", nacosServerIp, nacosServerPort)
        g.Log().Debugf(ctx, "使用单实例 Nacos 配置: %s", address)
        return address
    }
    ```

- [x] 2.2 实现地址格式验证函数

  - File: `ttpos-bmp/internal/pkg/nacos/service/rpc_service.go`
  - Purpose: 验证 Nacos 地址格式是否正确
  - Requirements: Requirement 1
  - Leverage: GoFrame 字符串处理函数
  - Prompt: Role: Go Developer | Task: 实现 isValidAddress() 函数，验证 Nacos 地址格式 | Context: 地址格式应为 host:port，需要验证格式正确性 | Restrictions: 使用 gstr 包处理字符串 | Success: 函数能够正确验证地址格式，返回 bool 值
  - Code Example:
    ```go
    func isValidAddress(addr string) bool {
        parts := gstr.Split(addr, ":")
        return len(parts) == 2 && len(parts[0]) > 0 && len(parts[1]) > 0
    }
    ```

- [x] 2.3 更新各子模块配置文件格式

  - Files: `ttpos-bmp/app/*/manifest/config/config.tpl.yaml`
  - Purpose: 统一各子模块的 Nacos 配置格式，添加多实例支持
  - Requirements: Requirement 4
  - Leverage: 现有配置文件格式，Redis 集群配置格式
  - Prompt: Role: DevOps Engineer | Task: 更新各子模块的 Nacos 配置格式，添加 addresses 配置项 | Context: 参考 design.md 中的配置格式，添加 addresses 配置，保持 ip+port 配置兼容 | Restrictions: 保持向后兼容，不破坏现有配置 | Success: 所有子模块配置文件格式统一，支持多实例配置
  - Code Example:
    ```yaml
    nacos:
      server:
        addresses: "$NACOS_SERVER_ADDRESSES"  # 格式：host1:port1,host2:port2,host3:port3
        # 兼容旧配置
        ip: $NACOS_SERVER_IP
        port: $NACOS_SERVER_PORT
      config:
        dataId: ttpos-erp.yaml
        group: DEFAULT_GROUP
    ```

---

**注意**：由于 GoFrame `nacos.New()` 已支持多实例，无需实现客户端池、健康检查和负载均衡逻辑。只需修改 `GetNacosAddress()` 函数即可。

---

## Phase 3: 测试和文档

- [ ] 3.1 单元测试：配置解析逻辑

  - File: `ttpos-bmp/internal/pkg/nacos/service/rpc_service_test.go`
  - Purpose: 测试 GetNacosAddress() 函数的配置解析逻辑
  - Requirements: Requirement 1
  - Leverage: Go 测试框架
  - Prompt: Role: QA Engineer | Task: 编写单元测试，测试配置解析逻辑 | Context: 测试多实例配置（逗号分隔）、单实例配置、错误配置等场景 | Restrictions: 使用 Go 测试框架 | Success: 所有测试用例通过
  - Test Cases:
    - 测试多实例配置：`"host1:port1,host2:port2,host3:port3"`
    - 测试单实例配置：`ip + port`
    - 测试配置优先级：`addresses` 优先于 `ip+port`
    - 测试地址格式验证：无效格式应被忽略或警告

- [ ] 3.2 集成测试：多实例配置

  - File: 测试环境
  - Purpose: 在测试环境验证多实例配置功能
  - Requirements: Requirement 1
  - Leverage: 测试环境、Nacos 集群
  - Prompt: Role: QA Engineer | Task: 在测试环境验证多实例配置功能 | Context: 部署 Nacos 集群（3个节点），配置多个实例地址，验证系统能够连接到所有实例 | Restrictions: 使用测试环境 | Success: 系统能够正确连接到所有配置的实例，GoFrame 自动处理多实例连接

- [ ] 3.3 集成测试：故障转移（由 Nacos SDK 自动处理）

  - File: 测试环境
  - Purpose: 在测试环境验证故障转移功能（由 Nacos SDK 自动处理）
  - Requirements: Requirement 2
  - Leverage: 测试环境、Nacos 集群
  - Prompt: Role: QA Engineer | Task: 在测试环境验证故障转移功能 | Context: 模拟主实例故障，验证 Nacos SDK 自动切换到备用实例 | Restrictions: 使用测试环境 | Success: Nacos SDK 能够自动检测故障并切换

- [ ] 3.4 集成测试：向后兼容性

  - File: 测试环境
  - Purpose: 验证单实例配置格式的向后兼容性
  - Requirements: Requirement 1
  - Leverage: 测试环境
  - Prompt: Role: QA Engineer | Task: 验证向后兼容性，确保单实例配置格式继续有效 | Context: 使用单实例配置格式（ip+port），验证系统正常工作 | Restrictions: 使用测试环境 | Success: 单实例配置格式继续有效，功能正常

- [ ] 3.5 更新相关文档

  - Files: README.md、配置文档
  - Purpose: 更新文档，说明多实例配置格式和使用方法
  - Requirements: Requirement 4
  - Leverage: 现有文档
  - Prompt: Role: Technical Writer | Task: 更新相关文档，说明多实例配置格式和使用方法 | Context: 更新配置说明、使用指南、故障排查等文档，说明 GoFrame 自动处理多实例、故障转移和负载均衡 | Restrictions: 保持文档格式一致 | Success: 文档已更新，说明清晰

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 代码符合 Go BMP 规范

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-bmp.mdc`
- [ ] 向后兼容性验证通过
- [ ] 各子模块配置格式统一

---

## 进度追踪

### 执行流程

1. ✅ **技术预研**: Phase 1 已完成，确认 GoFrame 支持多实例
2. **配置扩展**: 完成 Phase 2，修改 `GetNacosAddress()` 函数，支持多实例配置
3. **测试文档**: 完成 Phase 3，测试和文档更新

**简化说明**：由于 GoFrame `nacos.New()` 已支持多实例，无需实现客户端池、健康检查和负载均衡逻辑。

### 预计时间

- Phase 1: 0.5 天（4 小时）- ✅ 已完成
- Phase 2: 1 天（8 小时）
- Phase 3: 0.5 天（4 小时）
- **总计**: 2 天（16 小时）= **SP 2**

**说明**：由于 GoFrame 已支持多实例，无需实现客户端池、健康检查和负载均衡，工作量大幅减少。

---

## 附录：AI Prompt 示例

### 配置解析实现

```
Role: Go Developer with GoFrame expertise

Task: 修改 GetNacosAddress() 函数，支持返回逗号分隔的多实例地址

Context:
- File: ttpos-bmp/internal/pkg/nacos/service/rpc_service.go
- Leverage: design.md 中的实现方案，现有 GetNacosAddress() 函数，[GoFrame Nacos 源码](https://github.com/gogf/gf/tree/master/contrib/registry/nacos)
- Requirements: requirements.md Requirement 1
- Dependencies: g.Cfg(), gstr.Split(), gerror, fmt

Implementation Steps:
1. 优先检查 nacos.server.addresses 配置（多实例）
2. 如果存在，验证地址格式，返回逗号分隔的地址字符串
3. 如果不存在，使用 ip+port 配置（单实例，兼容），返回单个地址
4. GoFrame nacos.New() 会自动解析逗号分隔的地址，创建多实例连接

Restrictions:
- 保持向后兼容（单实例配置格式完全兼容）
- 遵循 Go BMP 规范
- 错误处理完善（无效地址记录警告，不中断）

Success Criteria:
- 代码通过 go fmt 和 go vet
- 能够正确解析多实例和单实例配置
- 向后兼容性验证通过
- GoFrame 自动处理多实例连接、故障转移和负载均衡
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/2025-11/2025-11-24.md`
- 当执行任务中形成复盘/优化建议时，及时沉淀 Episode 并在本节更新名称。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-24  
**维护者**: 后端开发组

