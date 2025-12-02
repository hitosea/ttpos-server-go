# Main Nacos 多实例支持优化 任务分解

> 本文档定义 Main Nacos 多实例支持优化的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-2 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 10  
**已完成**: 4  
**进行中**: -  
**完成率**: 40%

**说明**：由于 Nacos SDK 原生支持多 `ServerConfig`，无需实现客户端池、健康检查和负载均衡逻辑，任务数量较少。

---

## Phase 1: 配置结构扩展

- [x] 1.1 扩展 NacosConf 结构体，添加 Addresses 字段

  - File: `main/config/types.go`
  - Purpose: 扩展配置结构体，支持多实例地址配置
  - Requirements: Requirement 1
  - Leverage: 现有 `NacosConf` 结构体
  - Prompt: Role: Go Developer | Task: 扩展 NacosConf 结构体，添加 Addresses 字段 | Context: 参考 design.md 中的结构体设计，添加 Addresses 字段用于多实例配置，保持 Host 和 Port 字段用于向后兼容 | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持向后兼容 | Success: 结构体能够正确表示多实例和单实例配置
  - Code Example:
    ```go
    // NacosConf Nacos配置结构体
    type NacosConf struct {
        Addresses string // 多实例配置（格式：host1:port1,host2:port2,host3:port3），优先使用
        Host      string // nacos服务地址（兼容旧配置）
        Port      int    // nacos端口（兼容旧配置）
        Namespace string // 命名空间
        Username  string // 用户名
        Password  string // 密码
        DataId    string // 配置DataId
        Group     string // 配置Group
    }
    ```

- [x] 1.2 修改配置加载逻辑，支持读取 NACOS_SERVER_ADDRESSES 环境变量

  - File: `main/config/config.go`
  - Purpose: 扩展配置加载逻辑，支持读取多实例配置
  - Requirements: Requirement 1, Requirement 4
  - Leverage: 现有 `nacosConf()` 函数，Viper 配置读取
  - Prompt: Role: Go Developer | Task: 修改 nacosConf() 函数，支持读取 NACOS_SERVER_ADDRESSES 环境变量 | Context: 参考 design.md 中的配置加载逻辑，优先读取 NACOS_SERVER_ADDRESSES，兼容 NACOS_SERVER_IP 和 NACOS_SERVER_PORT | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持向后兼容 | Success: 函数能够正确读取多实例和单实例配置
  - Code Example:
    ```go
    func nacosConf(opt copier.Option) {
        Nacos = NacosConf{
            Addresses: "",
            Host:      "localhost",
            Port:      8848,
            // ... 其他字段
        }
        copier.CopyWithOption(&Nacos, NacosConf{
            Addresses: viper.GetString("NACOS_SERVER_ADDRESSES"),  // 多实例配置（优先）
            Host:      viper.GetString("NACOS_SERVER_IP"),        // 兼容旧配置
            Port:      viper.GetInt("NACOS_SERVER_PORT"),         // 兼容旧配置
            // ... 其他字段
        }, opt)
    }
    ```

---

## Phase 2: 客户端初始化改造

- [x] 2.1 修改 NewNacosClient() 函数，支持解析多个地址并创建多个 ServerConfig

  - File: `main/pkg/nacos/nacos.go`
  - Purpose: 扩展客户端初始化逻辑，支持多实例配置
  - Requirements: Requirement 1, Requirement 2, Requirement 3
  - Leverage: 现有 `NewNacosClient()` 函数，Nacos SDK 多 `ServerConfig` 支持
  - Prompt: Role: Go Developer | Task: 修改 NewNacosClient() 函数，支持解析多个地址并创建多个 ServerConfig | Context: 参考 design.md 中的实现方案，优先使用 Addresses 配置，解析逗号分隔的地址，创建多个 ServerConfig，兼容单实例配置 | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持向后兼容，错误处理完善 | Success: 函数能够正确解析多实例和单实例配置，创建多个或单个 ServerConfig
  - Code Example:
    ```go
    func NewNacosClient(nacosConfig config.NacosConf) (*NacosClient, error) {
        // ... ClientConfig 创建逻辑 ...
        
        // 解析多个地址，创建多个 ServerConfig
        var serverConfigs []constant.ServerConfig
        if len(nacosConfig.Addresses) > 0 {
            // 多实例配置
            addresses := strings.Split(nacosConfig.Addresses, ",")
            for _, addr := range addresses {
                addr = strings.TrimSpace(addr)
                parts := strings.Split(addr, ":")
                if len(parts) == 2 {
                    host := strings.TrimSpace(parts[0])
                    portStr := strings.TrimSpace(parts[1])
                    if len(host) > 0 && len(portStr) > 0 {
                        port, err := strconv.ParseUint(portStr, 10, 64)
                        if err == nil {
                            serverConfigs = append(serverConfigs, 
                                *constant.NewServerConfig(host, port))
                        }
                    }
                }
            }
        } else {
            // 单实例配置（向后兼容）
            serverConfigs = []constant.ServerConfig{
                *constant.NewServerConfig(nacosConfig.Host, uint64(nacosConfig.Port)),
            }
        }
        
        // 创建配置客户端和服务发现客户端
        // ...
    }
    ```

- [x] 2.2 实现地址格式验证和错误处理

  - File: `main/pkg/nacos/nacos.go`
  - Purpose: 验证地址格式，处理无效地址
  - Requirements: Requirement 1
  - Leverage: Go 标准库字符串处理
  - Prompt: Role: Go Developer | Task: 实现地址格式验证和错误处理逻辑 | Context: 验证地址格式为 host:port，过滤无效地址并记录警告日志，如果所有地址都无效则回退到单实例配置 | Restrictions: 使用标准库，错误处理完善 | Success: 函数能够正确验证地址格式，处理无效地址
  - Code Example:
    ```go
    // 在 NewNacosClient() 中实现地址验证
    validCount := 0
    for _, addr := range addresses {
        addr = strings.TrimSpace(addr)
        parts := strings.Split(addr, ":")
        if len(parts) == 2 {
            host := strings.TrimSpace(parts[0])
            portStr := strings.TrimSpace(parts[1])
            if len(host) > 0 && len(portStr) > 0 {
                port, err := strconv.ParseUint(portStr, 10, 64)
                if err == nil {
                    serverConfigs = append(serverConfigs, 
                        *constant.NewServerConfig(host, port))
                    validCount++
                } else {
                    logger.Logger.Warn("无效的 Nacos 端口格式，将忽略", 
                        zap.String("address", addr), zap.Error(err))
                }
            }
        } else {
            logger.Logger.Warn("无效的 Nacos 地址格式，将忽略", 
                zap.String("address", addr))
        }
    }
    ```

---

## Phase 3: 测试和文档

- [ ] 3.1 单元测试：配置解析逻辑

  - File: `main/config/config_test.go`
  - Purpose: 测试配置加载逻辑
  - Requirements: Requirement 1
  - Leverage: Go 测试框架
  - Prompt: Role: QA Engineer | Task: 编写单元测试，测试配置加载逻辑 | Context: 测试多实例配置（NACOS_SERVER_ADDRESSES）、单实例配置（NACOS_SERVER_IP + NACOS_SERVER_PORT）、配置优先级等场景 | Restrictions: 使用 Go 测试框架 | Success: 所有测试用例通过
  - Test Cases:
    - 测试多实例配置：`NACOS_SERVER_ADDRESSES=host1:port1,host2:port2,host3:port3`
    - 测试单实例配置：`NACOS_SERVER_IP + NACOS_SERVER_PORT`
    - 测试配置优先级：`NACOS_SERVER_ADDRESSES` 优先于 `NACOS_SERVER_IP + NACOS_SERVER_PORT`

- [ ] 3.2 单元测试：客户端初始化逻辑

  - File: `main/pkg/nacos/nacos_test.go`
  - Purpose: 测试客户端初始化逻辑
  - Requirements: Requirement 1, Requirement 2
  - Leverage: Go 测试框架
  - Prompt: Role: QA Engineer | Task: 编写单元测试，测试客户端初始化逻辑 | Context: 测试多实例配置解析、单实例配置解析、无效地址处理等场景 | Restrictions: 使用 Go 测试框架 | Success: 所有测试用例通过
  - Test Cases:
    - 测试多实例配置解析：解析逗号分隔的地址，创建多个 ServerConfig
    - 测试单实例配置解析：使用 Host + Port，创建单个 ServerConfig
    - 测试无效地址处理：过滤无效地址，记录警告日志

- [ ] 3.3 集成测试：多实例配置

  - File: 测试环境
  - Purpose: 在测试环境验证多实例配置功能
  - Requirements: Requirement 1
  - Leverage: 测试环境、Nacos 集群
  - Prompt: Role: QA Engineer | Task: 在测试环境验证多实例配置功能 | Context: 部署 Nacos 集群（3个节点），配置多个实例地址，验证系统能够连接到所有实例 | Restrictions: 使用测试环境 | Success: 系统能够正确连接到所有配置的实例，Nacos SDK 自动处理多实例连接

- [ ] 3.4 集成测试：故障转移（由 Nacos SDK 自动处理）

  - File: 测试环境
  - Purpose: 在测试环境验证故障转移功能（由 Nacos SDK 自动处理）
  - Requirements: Requirement 2
  - Leverage: 测试环境、Nacos 集群
  - Prompt: Role: QA Engineer | Task: 在测试环境验证故障转移功能 | Context: 模拟主实例故障，验证 Nacos SDK 自动切换到备用实例 | Restrictions: 使用测试环境 | Success: Nacos SDK 能够自动检测故障并切换

- [ ] 3.5 集成测试：向后兼容性

  - File: 测试环境
  - Purpose: 验证单实例配置格式的向后兼容性
  - Requirements: Requirement 1
  - Leverage: 测试环境
  - Prompt: Role: QA Engineer | Task: 验证向后兼容性，确保单实例配置格式继续有效 | Context: 使用单实例配置格式（NACOS_SERVER_IP + NACOS_SERVER_PORT），验证系统正常工作 | Restrictions: 使用测试环境 | Success: 单实例配置格式继续有效，功能正常

- [ ] 3.6 更新相关文档

  - Files: README.md、配置文档
  - Purpose: 更新文档，说明多实例配置格式和使用方法
  - Requirements: Requirement 4
  - Leverage: 现有文档
  - Prompt: Role: Technical Writer | Task: 更新相关文档，说明多实例配置格式和使用方法 | Context: 更新配置说明、使用指南、故障排查等文档，说明 Nacos SDK 自动处理多实例、故障转移和负载均衡 | Restrictions: 保持文档格式一致 | Success: 文档已更新，说明清晰

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 代码符合 Go Main 规范

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 向后兼容性验证通过
- [ ] 配置格式与 BMP 模块保持一致

---

## 进度追踪

### 执行流程

1. **配置扩展**: 完成 Phase 1，扩展配置结构体和加载逻辑
2. **客户端改造**: 完成 Phase 2，修改客户端初始化逻辑
3. **测试文档**: 完成 Phase 3，测试和文档更新

**简化说明**：由于 Nacos SDK 原生支持多 `ServerConfig`，无需实现客户端池、健康检查和负载均衡逻辑。

### 预计时间

- Phase 1: 0.5 天（4 小时）
- Phase 2: 0.5 天（4 小时）
- Phase 3: 1 天（8 小时）
- **总计**: 2 天（16 小时）= **SP 2**

**说明**：由于 Nacos SDK 已支持多实例，无需实现客户端池、健康检查和负载均衡，工作量较少。

---

## 附录：AI Prompt 示例

### 客户端初始化实现

```
Role: Go Developer with Nacos SDK expertise

Task: 修改 NewNacosClient() 函数，支持解析多个地址并创建多个 ServerConfig

Context:
- File: main/pkg/nacos/nacos.go
- Leverage: design.md 中的实现方案，现有 NewNacosClient() 函数，Nacos SDK 多 ServerConfig 支持
- Requirements: requirements.md Requirement 1, Requirement 2, Requirement 3
- Dependencies: strings, strconv, logger

Implementation Steps:
1. 优先检查 nacosConfig.Addresses 配置（多实例）
2. 如果存在，解析逗号分隔的地址，验证格式，创建多个 ServerConfig
3. 如果不存在，使用 Host+Port 配置（单实例，兼容），创建单个 ServerConfig
4. Nacos SDK 会自动处理多实例连接、故障转移和负载均衡

Restrictions:
- 保持向后兼容（单实例配置格式完全兼容）
- 遵循 Go Main 规范
- 错误处理完善（无效地址记录警告，不中断）

Success Criteria:
- 代码通过 go fmt 和 go vet
- 能够正确解析多实例和单实例配置
- 向后兼容性验证通过
- Nacos SDK 自动处理多实例连接、故障转移和负载均衡
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

