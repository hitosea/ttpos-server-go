# Main Nacos 多实例支持优化 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目          | 内容                                                                                                           |
| ------------- | -------------------------------------------------------------------------------------------------------------- |
| **提案人**    | rikugun                                                                                                        |
| **日期**      | 2025-11-24                                                                                                     |
| **目标版本**  | v2.11.0                                                                                                        |
| **状态**      | 已创建 Spec                                                                                                    |
| **关联任务**  | -                                                                                                              |
| **关联 Spec** | [task-main-nacos-multi-instance-support](../../../shared/specs/active/task-main-nacos-multi-instance-support/) |

---

## 🎯 背景和动机

### 问题描述

当前 `ttpos-main` 模块对接 Nacos 时，仅支持配置单个 Nacos 实例地址，存在以下问题：

1. **单点故障风险**

   - 当唯一的 Nacos 实例宕机时，`ttpos-main` 模块无法进行服务发现和配置获取
   - 无法实现高可用部署

2. **配置限制**

   - 当前配置格式：`NACOS_SERVER_IP` + `NACOS_SERVER_PORT`，只能配置一个地址
   - `NacosConf` 结构体只有 `Host` 和 `Port` 字段，不支持多实例
   - `NewNacosClient()` 函数只创建单个 `ServerConfig`

3. **扩展性不足**
   - 无法支持 Nacos 集群部署场景
   - 无法实现故障转移和负载均衡

**示例场景**：

> 生产环境部署了 Nacos 集群（3 个节点），但当前代码只能连接其中一个节点。当该节点故障时，`ttpos-main` 无法从 Nacos 获取服务发现信息，导致无法调用 `ttpos-bmp` 子模块的 gRPC 服务，整个系统不可用。

### 业务价值

解决这个问题能带来以下业务价值：

- **提升系统可用性**：支持多实例配置，实现故障自动转移，降低单点故障风险
- **支持集群部署**：支持 Nacos 集群模式，提升系统扩展性
- **提高运维灵活性**：运维人员可以配置多个 Nacos 实例，实现负载均衡
- **符合生产环境最佳实践**：满足高可用架构要求
- **与 BMP 模块保持一致**：`ttpos-bmp` 模块已支持多实例，`ttpos-main` 也应保持一致

### 目标用户

- [ ] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] **运维人员**：需要配置和管理 Nacos 集群
- [x] **开发人员**：需要实现高可用架构
- [ ] 其他: **\_\_\_\_**

---

## 💡 解决方案概述

### 方案描述

优化 `ttpos-main` 模块的 Nacos 对接机制，支持配置多个 Nacos 实例地址，实现故障转移和负载均衡。具体包括：

1. **配置格式优化**：支持配置多个 Nacos 实例地址（逗号分隔格式）
2. **结构体扩展**：扩展 `NacosConf` 结构体，添加 `Addresses` 字段
3. **客户端改造**：修改 `NewNacosClient()` 函数，支持解析多个地址并创建多个 `ServerConfig`
4. **配置加载优化**：修改配置加载逻辑，支持从环境变量读取多个地址
5. **向后兼容**：保持单实例配置格式的兼容性

**参考实现**：

> 参考 `ttpos-bmp` 模块的多实例支持实现，以及 Nacos SDK 原生支持多 `ServerConfig` 的能力

### 核心功能点

1. **多实例配置支持**

   - 支持配置多个 Nacos 实例地址（逗号分隔格式：`host1:port1,host2:port2,host3:port3`）
   - 兼容现有单实例配置格式（`NACOS_SERVER_IP` + `NACOS_SERVER_PORT`）

2. **故障转移机制**

   - Nacos SDK 自动处理多实例连接和故障转移
   - 当主实例故障时，自动切换到备用实例

3. **负载均衡**

   - Nacos SDK 自动在多个可用实例之间进行负载均衡

4. **配置统一管理**
   - 统一配置格式，与 `ttpos-bmp` 模块保持一致
   - 提供统一的配置解析和初始化逻辑

### 影响范围

**涉及终端**：

- [x] **POS 收银端**：依赖 `ttpos-main` 的服务发现能力
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：

- [ ] UI 组件
- [ ] API 接口
- [ ] 数据模型
- [ ] 业务逻辑
- [ ] 第三方集成
- [x] **基础设施**：Nacos 配置和服务发现
- [x] **配置管理**：配置文件格式和解析逻辑

**涉及文件**：

- [x] `main/config/types.go`：`NacosConf` 结构体定义
- [x] `main/config/config.go`：配置加载逻辑
- [x] `main/pkg/nacos/nacos.go`：Nacos 客户端初始化逻辑
- [x] `main/app/cloud/nacos.go`：Nacos 客户端使用逻辑

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要修改配置结构和客户端初始化逻辑，保持向后兼容
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：

- 需要修改 `NacosConf` 结构体，添加 `Addresses` 字段
- 需要修改 `NewNacosClient()` 函数，支持解析多个地址
- 需要修改配置加载逻辑，支持从环境变量读取多个地址
- 需要保持向后兼容性（单实例配置格式继续有效）
- Nacos SDK 原生支持多 `ServerConfig`，无需实现额外的故障转移逻辑

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 1-2 天
- **预估 SP**: 1-2（待技术评审确认）

**任务分解**：

1. 配置结构体扩展（0.5 天）
2. 客户端初始化逻辑改造（0.5 天）
3. 配置加载逻辑优化（0.5 天）
4. 测试和文档（0.5 天）

### 风险识别

**潜在风险**：

1. **向后兼容性问题**：修改配置格式可能影响现有部署
2. **Nacos SDK 使用方式**：需要确认 Nacos SDK 多实例配置的正确使用方式
3. **环境变量配置**：需要确认环境变量配置格式和解析逻辑

**缓解措施**：

1. **保持向后兼容**：优先支持单实例配置，多实例配置作为可选增强
2. **技术预研**：先调研 Nacos SDK 的多实例支持能力，参考 `ttpos-bmp` 模块的实现
3. **分阶段实施**：先实现配置解析，再实现客户端初始化，最后测试验证
4. **充分测试**：在测试环境部署 Nacos 集群，验证故障转移机制

---

## 🔗 相关资源

### 参考需求

- 类似功能: `ttpos-bmp` 模块的 Nacos 多实例支持（已完成）
- 竞品分析: Spring Cloud Alibaba Nacos 多实例配置

### 相关文档

- Nacos SDK 文档: https://github.com/nacos-group/nacos-sdk-go
- Nacos 集群部署文档: https://nacos.io/docs/latest/guide/admin/cluster-mode-quick-start/
- BMP 模块实现: `ttpos-bmp/internal/pkg/nacos/service/rpc_service.go`
- BMP 模块提案: `docs/team/proposals/2025-11/bmp-nacos-multi-instance-support.md`

### 相关代码

- Nacos 配置结构: `main/config/types.go:NacosConf`
- 配置加载: `main/config/config.go:nacosConf()`
- Nacos 客户端初始化: `main/pkg/nacos/nacos.go:NewNacosClient()`
- Nacos 客户端使用: `main/app/cloud/nacos.go`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`task-main-nacos-multi-instance-support`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 运维人员  
**我想** 配置多个 Nacos 实例地址  
**以便于** 实现高可用部署，避免单点故障导致 `ttpos-main` 模块无法进行服务发现

### AC 验收标准（初稿）

1. **WHEN** 配置多个 Nacos 实例地址 **THEN** `ttpos-main` 模块 **SHALL** 能够连接到所有配置的实例
2. **IF** 主 Nacos 实例故障 **THEN** `ttpos-main` 模块 **SHALL** 自动切换到备用实例
3. **WHEN** 使用单实例配置格式（`NACOS_SERVER_IP` + `NACOS_SERVER_PORT`） **THEN** 系统 **SHALL** 保持向后兼容，正常工作
4. **WHEN** 所有 Nacos 实例都不可用 **THEN** 系统 **SHALL** 记录错误日志并降级处理

### 技术方案（初稿）

#### 配置格式设计

**方案：逗号分隔字符串（推荐，与 BMP 模块保持一致）**

环境变量配置：

```bash
# 多实例配置（优先使用）
NACOS_SERVER_ADDRESSES=host1:port1,host2:port2,host3:port3

# 兼容旧配置（单实例）
NACOS_SERVER_IP=host1
NACOS_SERVER_PORT=8848
```

配置结构体扩展：

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

#### 实现要点

1. **配置解析优先级**：

   - 优先使用 `NACOS_SERVER_ADDRESSES` 环境变量（多实例）
   - 如果不存在，使用 `NACOS_SERVER_IP` + `NACOS_SERVER_PORT`（单实例，向后兼容）

2. **客户端初始化**：

   - 解析多个地址，创建多个 `ServerConfig`
   - Nacos SDK 自动处理多实例连接、故障转移和负载均衡

3. **地址格式验证**：
   - 验证地址格式为 `host:port`
   - 过滤无效地址并记录警告日志

**参考实现**：

```go
// main/pkg/nacos/nacos.go

func NewNacosClient(nacosConfig config.NacosConf) (*NacosClient, error) {
    // ... ClientConfig 创建逻辑 ...

    // 解析多个地址
    var serverConfigs []constant.ServerConfig
    if len(nacosConfig.Addresses) > 0 {
        // 多实例配置
        addresses := strings.Split(nacosConfig.Addresses, ",")
        for _, addr := range addresses {
            addr = strings.TrimSpace(addr)
            parts := strings.Split(addr, ":")
            if len(parts) == 2 {
                port, _ := strconv.ParseUint(parts[1], 10, 64)
                serverConfigs = append(serverConfigs,
                    *constant.NewServerConfig(parts[0], port))
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

---

**版本**: v1.0.0  
**创建日期**: 2025-11-24  
**维护者**: 后端开发组  
**相关规范**: `.cursor/rules/go-main.mdc`, `.cursor/rules/specs.mdc`
