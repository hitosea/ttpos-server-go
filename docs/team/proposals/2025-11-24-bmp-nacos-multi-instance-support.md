# BMP Nacos 多实例支持优化 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-11-24   |
| **目标版本** | v2.11.0 |
| **状态**   | 已创建 Spec   |
| **关联任务** | -      |
| **关联 Spec** | [task-bmp-nacos-multi-instance-support](../../shared/specs/task-bmp-nacos-multi-instance-support/)      |

---

## 🎯 背景和动机

### 问题描述

当前 `ttpos-bmp` 各子模块（ttpos-erp、ttpos-message、ttpos-websocket、ttpos-shop、ttpos-takeout、ttpos-manager）对接 Nacos 时，仅支持配置单个 Nacos 实例地址，存在以下问题：

1. **单点故障风险**
   - 当唯一的 Nacos 实例宕机时，所有依赖 Nacos 的服务都会受到影响
   - 无法实现高可用部署

2. **配置限制**
   - 当前配置格式：`nacos.server.ip` + `nacos.server.port`，只能配置一个地址
   - `GetNacosAddress()` 函数返回单个地址字符串，不支持多实例

3. **扩展性不足**
   - 无法支持 Nacos 集群部署场景
   - 无法实现负载均衡和故障转移

**示例场景**：
> 生产环境部署了 Nacos 集群（3个节点），但当前代码只能连接其中一个节点。当该节点故障时，所有服务都无法从 Nacos 获取配置和服务发现信息，导致服务不可用。

### 业务价值

解决这个问题能带来以下业务价值：

- **提升系统可用性**：支持多实例配置，实现故障自动转移，降低单点故障风险
- **支持集群部署**：支持 Nacos 集群模式，提升系统扩展性
- **提高运维灵活性**：运维人员可以配置多个 Nacos 实例，实现负载均衡
- **符合生产环境最佳实践**：满足高可用架构要求

### 目标用户

- [ ] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] **运维人员**：需要配置和管理 Nacos 集群
- [x] **开发人员**：需要实现高可用架构
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

优化 `ttpos-bmp` 各子模块的 Nacos 对接机制，支持配置多个 Nacos 实例地址，实现故障转移和负载均衡。具体包括：

1. **配置格式优化**：支持配置多个 Nacos 实例地址（类似 Redis 集群配置）
2. **客户端改造**：修改 `GetNacosAddress()` 函数，支持返回多个地址
3. **故障转移机制**：实现自动故障检测和切换逻辑
4. **向后兼容**：保持单实例配置格式的兼容性

**参考实现**：
> 参考 Redis 集群配置方式，支持逗号分隔的多个地址：`host1:port1,host2:port2,host3:port3`

### 核心功能点

1. **多实例配置支持**
   - 支持配置多个 Nacos 实例地址（逗号分隔或数组格式）
   - 兼容现有单实例配置格式

2. **故障转移机制**
   - 实现健康检查机制，自动检测 Nacos 实例可用性
   - 当主实例故障时，自动切换到备用实例

3. **负载均衡**
   - 支持多个可用实例之间的负载均衡
   - 实现连接池管理

4. **配置统一管理**
   - 统一各子模块的 Nacos 配置格式
   - 提供统一的配置解析和初始化逻辑

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

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [ ] 数据模型
- [x] **业务逻辑**：Nacos 客户端初始化和服务发现逻辑
- [ ] 第三方集成
- [x] **基础设施**：Nacos 配置和服务注册发现
- [x] **配置管理**：配置文件格式和解析逻辑

**涉及子模块**：
- [x] `ttpos-bmp/app/ttpos-erp`
- [x] `ttpos-bmp/app/ttpos-message`
- [x] `ttpos-bmp/app/ttpos-websocket`
- [x] `ttpos-bmp/app/ttpos-shop`
- [x] `ttpos-bmp/app/ttpos-takeout`
- [x] `ttpos-bmp/app/ttpos-manager`
- [x] `ttpos-bmp/internal/pkg/nacos`

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：
- 需要修改 Nacos 客户端初始化逻辑
- 需要实现故障转移和健康检查机制
- 需要统一各子模块的配置格式
- 需要保持向后兼容性

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 3-5 天
- **预估 SP**: 3-5（待技术评审确认）

**任务分解**：
1. 配置格式设计和实现（1天）
2. Nacos 客户端改造（1-2天）
3. 故障转移机制实现（1天）
4. 测试和文档（1天）

### 风险识别

**潜在风险**：
1. **向后兼容性问题**：修改配置格式可能影响现有部署
2. **GoFrame Nacos 客户端限制**：需要确认 GoFrame 的 Nacos 客户端是否支持多实例
3. **测试复杂度**：需要模拟 Nacos 故障场景进行测试

**缓解措施**：
1. **保持向后兼容**：优先支持单实例配置，多实例配置作为可选增强
2. **技术预研**：先调研 GoFrame Nacos 客户端的能力，确认是否支持多实例
3. **分阶段实施**：先实现配置解析，再实现故障转移，最后优化负载均衡
4. **充分测试**：在测试环境部署 Nacos 集群，验证故障转移机制

---

## 🔗 相关资源

### 参考需求

- 类似功能: Redis 集群配置（支持多个地址）
- 竞品分析: Spring Cloud Alibaba Nacos 多实例配置

### 相关文档

- GoFrame Nacos 客户端文档: https://goframe.org/pages/viewpage.action?pageId=1114320
- Nacos 集群部署文档: https://nacos.io/docs/latest/guide/admin/cluster-mode-quick-start/
- 当前实现: `ttpos-bmp/internal/pkg/nacos/service/rpc_service.go`

### 相关代码

- Nacos 配置解析: `ttpos-bmp/internal/pkg/nacos/service/rpc_service.go:GetNacosAddress()`
- 各子模块配置: `ttpos-bmp/app/*/manifest/config/config.tpl.yaml`

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

- [x] 创建 Spec：`task-bmp-nacos-multi-instance-support`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 运维人员  
**我想** 配置多个 Nacos 实例地址  
**以便于** 实现高可用部署，避免单点故障导致服务不可用

### AC 验收标准（初稿）

1. **WHEN** 配置多个 Nacos 实例地址 **THEN** 系统 **SHALL** 能够连接到所有配置的实例
2. **IF** 主 Nacos 实例故障 **THEN** 系统 **SHALL** 自动切换到备用实例
3. **WHEN** 使用单实例配置格式 **THEN** 系统 **SHALL** 保持向后兼容，正常工作
4. **WHEN** 所有 Nacos 实例都不可用 **THEN** 系统 **SHALL** 记录错误日志并降级处理

### 技术方案（初稿）

#### 配置格式设计

**方案一：逗号分隔字符串（推荐）**
```yaml
nacos:
  server:
    addresses: "$NACOS_SERVER_ADDRESSES"  # 格式：host1:port1,host2:port2,host3:port3
    # 兼容旧配置
    ip: $NACOS_SERVER_IP
    port: $NACOS_SERVER_PORT
```

**方案二：数组格式**
```yaml
nacos:
  server:
    instances:
      - ip: $NACOS_SERVER_IP_1
        port: $NACOS_SERVER_PORT_1
      - ip: $NACOS_SERVER_IP_2
        port: $NACOS_SERVER_PORT_2
```

#### 实现要点

1. **配置解析优先级**：
   - 优先使用 `addresses` 配置（多实例）
   - 如果不存在，使用 `ip` + `port`（单实例，向后兼容）

2. **客户端初始化**：
   - 解析多个地址，创建多个 Nacos 客户端连接
   - 实现连接池管理

3. **故障转移**：
   - 定期健康检查（心跳检测）
   - 主实例故障时自动切换
   - 记录故障日志

---

**版本**: v1.0.0  
**创建日期**: 2025-11-24  
**维护者**: 后端开发组  
**相关规范**: `.cursor/rules/go-bmp.mdc`, `.cursor/rules/specs.mdc`

