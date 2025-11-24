# Main Nacos 多实例支持优化 需求文档

> 本文档定义 Main Nacos 多实例支持优化的详细需求和验收标准。

## 📋 基本信息

| 项目          | 内容                                     |
| ------------- | ---------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11-24-main-nacos-multi-instance-support.md](../../../team/proposals/2025-11-24-main-nacos-multi-instance-support.md) |
| **创建日期**  | 2025-11-24                             |
| **负责人**    | {待分配}                                   |
| **目标 Sprint** | {待分配}                             |
| **涉及技术栈** | [x] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/) |

---

## 📋 概述

Main Nacos 多实例支持优化旨在解决当前 `ttpos-main` 模块仅支持单个 Nacos 实例配置的问题，通过支持多实例配置、实现故障转移和负载均衡，提升系统的高可用性和扩展性。

本功能主要涉及 Nacos 配置结构体扩展、客户端初始化逻辑改造、配置加载逻辑优化等。

## 🎯 产品对齐

该功能支持公司2025年Q4的核心目标：
- **提升系统可用性**: 支持多实例配置，实现故障自动转移，降低单点故障风险
- **支持集群部署**: 支持 Nacos 集群模式，提升系统扩展性
- **提高运维灵活性**: 运维人员可以配置多个 Nacos 实例，实现负载均衡
- **符合生产环境最佳实践**: 满足高可用架构要求
- **与 BMP 模块保持一致**: `ttpos-bmp` 模块已支持多实例，`ttpos-main` 也应保持一致

## 📝 用户故事

**作为** 运维人员  
**我想** 配置多个 Nacos 实例地址  
**以便于** 实现高可用部署，避免单点故障导致 `ttpos-main` 模块无法进行服务发现

---

## 功能需求

### Requirement 1: 多实例配置支持

**用户故事**: 作为运维人员，我想配置多个 Nacos 实例地址，以便于实现高可用部署

#### 验收标准

1. **WHEN** 配置多个 Nacos 实例地址（逗号分隔） **THEN** `ttpos-main` 模块 **SHALL** 能够解析并连接到所有配置的实例
2. **WHEN** 使用单实例配置格式（`NACOS_SERVER_IP` + `NACOS_SERVER_PORT`） **THEN** 系统 **SHALL** 保持向后兼容，正常工作
3. **IF** 配置格式错误（如地址格式不正确） **THEN** 系统 **SHALL** 返回清晰的错误提示或记录警告日志
4. **WHEN** 配置多个实例地址 **THEN** 系统 **SHALL** 支持环境变量方式配置

#### 具体要求

- [ ] 1.1 扩展 `NacosConf` 结构体，添加 `Addresses` 字段
- [ ] 1.2 保持单实例配置格式（`NACOS_SERVER_IP` + `NACOS_SERVER_PORT`）的向后兼容
- [ ] 1.3 配置解析优先级：优先使用 `NACOS_SERVER_ADDRESSES` 环境变量，不存在时使用 `NACOS_SERVER_IP` + `NACOS_SERVER_PORT`
- [ ] 1.4 支持逗号分隔的多实例地址格式：`host1:port1,host2:port2,host3:port3`

---

### Requirement 2: 故障转移机制

**用户故事**: 作为开发人员，我想系统能够自动检测 Nacos 实例故障并切换到备用实例，以便于提高系统可用性

#### 验收标准

1. **IF** 主 Nacos 实例故障 **THEN** 系统 **SHALL** 自动切换到备用实例（由 Nacos SDK 自动处理）
2. **WHEN** 主实例恢复 **THEN** 系统 **SHALL** 能够自动切换回主实例（由 Nacos SDK 自动处理）
3. **WHEN** 所有 Nacos 实例都不可用 **THEN** 系统 **SHALL** 记录错误日志并降级处理
4. **WHEN** 实例故障切换时 **THEN** 系统 **SHALL** 记录切换日志，便于运维排查（由 Nacos SDK 自动记录）

#### 具体要求

- [ ] 2.1 Nacos SDK 自动处理多实例连接和故障转移（无需额外实现）
- [ ] 2.2 确保 `NewNacosClient()` 函数正确创建多个 `ServerConfig`
- [ ] 2.3 记录详细的故障和切换日志（由 Nacos SDK 自动记录）

---

### Requirement 3: 负载均衡

**用户故事**: 作为运维人员，我想多个可用 Nacos 实例之间能够负载均衡，以便于提高系统性能和可用性

#### 验收标准

1. **WHEN** 配置多个可用 Nacos 实例 **THEN** 系统 **SHALL** 在多个实例之间进行负载均衡（由 Nacos SDK 自动处理）
2. **WHEN** 执行服务发现操作 **THEN** 系统 **SHALL** 选择合适的实例进行连接（由 Nacos SDK 自动处理）

#### 具体要求

- [ ] 3.1 Nacos SDK 自动处理多实例负载均衡（无需额外实现）
- [ ] 3.2 确保 `NewNacosClient()` 函数正确创建多个 `ServerConfig`

---

### Requirement 4: 配置统一管理

**用户故事**: 作为开发人员，我想 `ttpos-main` 模块的 Nacos 配置格式与 `ttpos-bmp` 模块保持一致，以便于维护和管理

#### 验收标准

1. **WHEN** 查看 `ttpos-main` 模块的配置 **THEN** 系统 **SHALL** 使用与 `ttpos-bmp` 模块一致的配置格式
2. **WHEN** 修改 Nacos 配置解析逻辑 **THEN** 系统 **SHALL** 统一在 `main/pkg/nacos` 中实现
3. **IF** 新增功能需要对接 Nacos **THEN** 系统 **SHALL** 复用统一的配置解析逻辑

#### 具体要求

- [ ] 4.1 统一配置格式，与 `ttpos-bmp` 模块保持一致（环境变量：`NACOS_SERVER_ADDRESSES`）
- [ ] 4.2 在 `main/pkg/nacos/nacos.go` 中实现统一的配置解析和初始化逻辑
- [ ] 4.3 复用统一的配置解析和初始化逻辑

---

## 非功能需求

### NFR 1: 性能要求

- **配置解析时间**: < 100ms
- **故障检测响应时间**: < 5秒（由 Nacos SDK 自动处理）
- **故障切换时间**: < 10秒（由 Nacos SDK 自动处理）

### NFR 2: 可维护性

- **代码结构清晰**: 配置解析、客户端初始化逻辑分离
- **易于测试**: 支持模拟 Nacos 故障场景进行测试
- **文档完善**: 更新相关文档，说明配置格式和使用方法

### NFR 3: 兼容性

- **向后兼容**: 保持现有单实例配置格式完全兼容
- **Nacos SDK 兼容**: 确保与 Nacos SDK 多实例配置兼容
- **与 BMP 模块一致**: 配置格式与 `ttpos-bmp` 模块保持一致

### NFR 4: 可观测性

- **日志记录**: 记录配置解析、实例连接等关键操作
- **错误处理**: 提供清晰的错误信息和处理建议
- **Nacos SDK 日志**: Nacos SDK 自动记录故障转移和负载均衡日志

---

## 验收标准总结

### 功能验收

- [ ] ✅ 配置多个 Nacos 实例地址时，`ttpos-main` 模块能够解析并连接到所有实例
- [ ] ✅ 使用单实例配置格式时，系统保持向后兼容，正常工作
- [ ] ✅ 主实例故障时，系统自动切换到备用实例（由 Nacos SDK 自动处理）
- [ ] ✅ 所有实例都不可用时，系统记录错误日志并降级处理
- [ ] ✅ 多个可用实例之间能够进行负载均衡（由 Nacos SDK 自动处理）

### 非功能验收

- [ ] ✅ 代码通过 `go fmt` 和 `go vet` 检查
- [ ] ✅ 配置解析和客户端初始化逻辑正确，无回归问题
- [ ] ✅ 配置格式与 `ttpos-bmp` 模块保持一致
- [ ] ✅ 相关文档已更新（配置说明、使用指南等）

---

## 相关资源

### 参考文档

- Nacos SDK 文档: https://github.com/nacos-group/nacos-sdk-go
- Nacos 集群部署文档: https://nacos.io/docs/latest/guide/admin/cluster-mode-quick-start/
- BMP 模块实现: `ttpos-bmp/internal/pkg/nacos/service/rpc_service.go`
- BMP 模块提案: `docs/team/proposals/2025-11-24-bmp-nacos-multi-instance-support.md`

### 相关代码

- Nacos 配置结构: `main/config/types.go:NacosConf`
- 配置加载: `main/config/config.go:nacosConf()`
- Nacos 客户端初始化: `main/pkg/nacos/nacos.go:NewNacosClient()`
- Nacos 客户端使用: `main/app/cloud/nacos.go`

---

**版本**: v1.0.0  
**创建日期**: 2025-11-24  
**维护者**: 后端开发组

