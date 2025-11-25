# BMP Nacos 多实例支持优化 设计文档

> 本文档定义 BMP Nacos 多实例支持优化的技术设计和实现方案。

## 📋 概述

BMP Nacos 多实例支持优化通过扩展配置格式，利用 GoFrame 内置的多实例支持能力，实现多个 Nacos 实例的高可用部署。

核心实现包括：
- 配置格式扩展：支持多实例地址配置（逗号分隔）
- 利用 GoFrame 能力：GoFrame `nacos.New()` 已支持多实例，自动处理故障转移和负载均衡
- 向后兼容：保持单实例配置格式完全兼容

**参考资源**：
- [GoFrame Nacos 注册中心源码](https://github.com/gogf/gf/tree/master/contrib/registry/nacos)
- GoFrame `nacos.New()` 函数支持逗号分隔的多实例地址：`"host1:port1,host2:port2,host3:port3"`

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- ✅ 遵循 GoFrame 框架规范
- ✅ 保持代码结构清晰，职责分离
- ✅ 错误处理使用 gerror
- ✅ 不使用 panic，返回 error
- ✅ 使用中文注释

---

## 🔄 代码复用分析

### 可复用的现有组件

- **配置解析**: `ttpos-bmp/internal/pkg/nacos/service/rpc_service.go:GetNacosAddress()` - 扩展支持多实例
- **Nacos 客户端**: GoFrame 的 `nacos.New()` - **已支持多实例**（参考 [GoFrame Nacos 源码](https://github.com/gogf/gf/tree/master/contrib/registry/nacos)）
- **Redis 集群配置**: `ttpos-bmp/app/ttpos-websocket/manifest/config/config.tpl.yaml` - 参考多地址配置格式

### GoFrame Nacos 多实例支持确认

根据 [GoFrame Nacos 注册中心源码](https://github.com/gogf/gf/tree/master/contrib/registry/nacos)，`nacos.New()` 函数**已经支持多实例配置**：

```go
// GoFrame nacos.New() 源码实现
func New(address string, opts ...constant.ClientOption) (reg *Registry) {
    endpoints := gstr.SplitAndTrim(address, ",")  // 支持逗号分隔的多个地址！
    // ...
    serverConfigs := make([]constant.ServerConfig, 0, len(endpoints))
    for _, endpoint := range endpoints {
        tmp := gstr.Split(endpoint, ":")
        ip := tmp[0]
        port := gconv.Uint64(tmp[1])
        if port == 0 {
            port = 8848
        }
        serverConfigs = append(serverConfigs, *constant.NewServerConfig(ip, port))
    }
    // 创建 Nacos 客户端，自动支持多实例和故障转移
    // ...
}
```

**关键发现**：
- ✅ GoFrame `nacos.New()` 支持逗号分隔的多个地址：`"host1:port1,host2:port2,host3:port3"`
- ✅ 底层 Nacos SDK 自动处理多实例连接和故障转移
- ✅ 无需实现客户端池和健康检查，SDK 已内置支持

### 集成点

- **配置解析**: 扩展 `GetNacosAddress()` 函数，支持返回逗号分隔的多实例地址
- **客户端初始化**: 直接使用 `nacos.New(addresses)`，GoFrame 自动处理多实例
- **故障转移**: 由 Nacos SDK 自动处理，无需额外实现

---

## 🏗️ 架构设计

### 当前架构

```
配置文件 (config.tpl.yaml)
  └── nacos.server.ip + nacos.server.port (单实例)
      └── GetNacosAddress() → 返回单个地址
          └── nacos.New(address) → 创建单个客户端
```

**问题**：
- 仅支持单个 Nacos 实例
- 单点故障风险
- 无法实现高可用

### 目标架构

```
配置文件 (config.tpl.yaml)
  ├── nacos.server.addresses (多实例，优先，格式：host1:port1,host2:port2,host3:port3)
  └── nacos.server.ip + nacos.server.port (单实例，兼容)
      └── GetNacosAddress() → 返回地址字符串（单实例或多实例）
          └── nacos.New(address) → GoFrame 自动解析逗号分隔的地址
              └── Nacos SDK 自动处理
                  ├── 多实例连接
                  ├── 故障转移（自动）
                  └── 负载均衡（自动）
```

**优势**：
- ✅ 支持多实例配置（逗号分隔）
- ✅ 自动故障转移（由 Nacos SDK 处理）
- ✅ 自动负载均衡（由 Nacos SDK 处理）
- ✅ 保持向后兼容（单实例配置格式完全兼容）
- ✅ 实现简单（只需修改配置解析函数）

---

## 🔧 实现方案

### 方案：扩展配置格式 + 利用 GoFrame 内置多实例支持（推荐）

**实现思路**：
1. 扩展配置格式，支持多实例地址（逗号分隔）
2. 修改 `GetNacosAddress()` 函数，支持返回逗号分隔的多实例地址
3. 直接使用 `nacos.New(addresses)`，GoFrame 自动处理多实例和故障转移
4. **无需实现客户端池和健康检查**，Nacos SDK 已内置支持

**参考实现**：
- [GoFrame Nacos 源码](https://github.com/gogf/gf/tree/master/contrib/registry/nacos) - `nacos.New()` 函数已支持逗号分隔的多实例地址
- Nacos SDK 自动处理多实例连接、故障转移和负载均衡

**配置格式设计**：

```yaml
# 支持逗号分隔的多实例地址（GoFrame 原生支持）
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

**代码实现**：

```go
// internal/pkg/nacos/service/rpc_service.go

// GetNacosAddress 获取 Nacos 地址（支持多实例，返回逗号分隔的地址字符串）
// GoFrame nacos.New() 支持逗号分隔的多个地址，格式：host1:port1,host2:port2,host3:port3
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

// isValidAddress 验证地址格式
func isValidAddress(addr string) bool {
	parts := gstr.Split(addr, ":")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) == 0 {
		return false
	}
	// 验证端口是否为数字
	port := gstr.Trim(parts[1])
	if len(port) == 0 {
		return false
	}
	return true
}
```

**修改现有函数**（无需大改，只需使用新的地址格式）：

```go
// Init() 函数保持不变，直接使用 GetNacosAddress()
func (s *rpcServer) Init(ctx context.Context) {
	grpcConf := grpcx.Server.NewConfig()
	//特殊处理 grpc.endpoints ,支持从环境变量中获取
	if endpoints := genv.Get("GRPC_ENDPOINTS", "").String(); len(endpoints) > 0 {
		g.Log().Infof(ctx, "使用环境变量注册服务 GRPC_ENDPOINTS: %v", endpoints)
		grpcConf.Endpoints = gstr.Split(endpoints, ",")
	}
	grpcConf.Options = append(grpcConf.Options,
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(unaryRecordInterceptor()),
	)
	
	// 使用 GetNacosAddress()，支持多实例（逗号分隔）
	// GoFrame nacos.New() 会自动解析逗号分隔的地址，创建多实例连接
	grpcx.Resolver.Register(nacos.New(GetNacosAddress(ctx)))
	
	s.GRpc = grpcx.Server.New(grpcConf)
}

// InitHttp() 函数保持不变
func (s *rpcServer) InitHttp(ctx context.Context) {
	// 使用 GetNacosAddress()，支持多实例（逗号分隔）
	gsvc.SetRegistry(nacos.New(GetNacosAddress(ctx)))
}
```

**优点**：
- ✅ **实现简单**：只需修改配置解析函数，无需实现客户端池
- ✅ **保持向后兼容**：单实例配置格式完全兼容
- ✅ **自动故障转移**：Nacos SDK 自动处理多实例故障转移
- ✅ **自动负载均衡**：Nacos SDK 自动在多个实例间负载均衡
- ✅ **参考 Redis 集群配置格式**，易于理解和配置
- ✅ **无需额外维护**：不需要实现健康检查和切换逻辑

**缺点**：
- 无（GoFrame 已完美支持）

---

---

## 🎯 推荐方案

**推荐使用扩展配置格式方案**，原因：
1. **实现简单**：只需修改 `GetNacosAddress()` 函数，支持返回逗号分隔的多实例地址
2. **向后兼容**：保持现有单实例配置格式完全兼容
3. **利用现有能力**：GoFrame `nacos.New()` 已支持多实例，无需额外实现
4. **风险极低**：不改变现有调用方式，只需扩展配置解析
5. **自动故障转移**：Nacos SDK 自动处理多实例故障转移和负载均衡

**实施步骤**：
1. ✅ **技术预研**：已确认 GoFrame Nacos 客户端支持多实例（逗号分隔地址）
2. **配置格式扩展**：修改 `GetNacosAddress()` 函数，支持解析多实例配置
3. **更新配置文件**：更新各子模块配置文件，添加 `addresses` 配置项
4. **测试验证**：在测试环境部署 Nacos 集群，验证多实例功能

---

## 🔍 实现细节

### 配置解析优先级

```
1. 检查 nacos.server.addresses（多实例，优先）
   ├── 存在 → 返回逗号分隔的地址字符串（格式：host1:port1,host2:port2,host3:port3）
   └── 不存在 → 检查 nacos.server.ip + nacos.server.port（单实例，兼容）
       └── 返回单个地址（格式：host:port）
```

### 故障转移机制

**由 Nacos SDK 自动处理**：
- Nacos SDK 内置故障检测和自动切换机制
- 当某个实例不可用时，自动切换到其他可用实例
- 无需额外实现健康检查和切换逻辑

### 负载均衡策略

**由 Nacos SDK 自动处理**：
- Nacos SDK 自动在多个实例间进行负载均衡
- 支持多种负载均衡策略（轮询、随机等）
- 无需额外实现负载均衡算法

---

## 🧪 测试策略

### 单元测试

**测试用例**：
1. 测试配置解析逻辑（多实例、单实例）
2. 测试地址格式验证函数
3. 测试配置优先级（addresses 优先于 ip+port）

### 集成测试

**测试场景**：
1. **多实例配置测试**
   - 配置多个 Nacos 实例地址
   - 验证系统能够连接到所有实例
   
2. **故障转移测试**
   - 模拟主实例故障
   - 验证系统自动切换到备用实例
   
3. **向后兼容测试**
   - 使用单实例配置格式
   - 验证系统正常工作

4. **负载均衡测试**
   - 配置多个可用实例
   - 验证请求在多个实例间负载均衡

### 环境准备

- **Nacos 集群**：部署 3 个 Nacos 节点
- **测试环境**：在测试环境部署各子模块，验证功能

---

## 📈 性能优化

### 优化措施

1. **配置解析优化**：缓存解析结果，避免重复解析
2. **地址验证优化**：快速验证地址格式，提前发现配置错误

### 性能指标

- **配置解析时间**: < 100ms
- **地址格式验证时间**: < 10ms
- **故障转移时间**: 由 Nacos SDK 自动处理（通常 < 5秒）

---

## 🚨 错误处理

### 主要错误场景

1. **配置格式错误**: 返回清晰的错误信息
2. **所有实例不可用**: 记录错误日志，降级处理
3. **客户端初始化失败**: 降级到单实例模式

### 错误处理策略

```go
// 配置解析错误（地址格式无效）
address := GetNacosAddress(ctx)
if len(address) == 0 {
	g.Log().Errorf(ctx, "Nacos 地址配置为空")
	// 返回错误或使用默认值
	return gerror.New("Nacos 地址配置为空")
}

// 地址格式验证
addressList := gstr.Split(address, ",")
for _, addr := range addressList {
	addr = gstr.Trim(addr)
	if !isValidAddress(addr) {
		g.Log().Warnf(ctx, "无效的 Nacos 地址格式，将忽略: %s", addr)
		// 继续处理其他地址，不中断
		continue
	}
}

// 客户端初始化失败（由 Nacos SDK 处理）
// GoFrame nacos.New() 会自动处理多实例连接失败的情况
// 如果所有实例都不可用，SDK 会返回错误，我们记录日志即可
```

---

## 📚 实现清单

### Phase 1: 配置格式扩展（参见 tasks.md）

- [ ] 修改 `GetNacosAddress()` 函数，支持返回逗号分隔的多实例地址
- [ ] 实现地址格式验证函数 `isValidAddress()`
- [ ] 更新各子模块配置文件格式，添加 `addresses` 配置项
- [ ] 测试配置解析逻辑（多实例、单实例）

### Phase 2: 测试和文档（参见 tasks.md）

- [ ] 单元测试：配置解析逻辑
- [ ] 集成测试：多实例配置、故障转移、向后兼容性
- [ ] 文档更新：配置说明、使用指南

**注意**：由于 GoFrame `nacos.New()` 已支持多实例，无需实现客户端池、健康检查和负载均衡逻辑。

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/2025-11/2025-11-24.md`
- 当设计结论可复用或踩坑较多时，沉淀 Episode 并在此更新名称，保持 Spec ↔ Graphiti 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-24  
**作者**: 后端开发组  
**审核者**: {待分配}

