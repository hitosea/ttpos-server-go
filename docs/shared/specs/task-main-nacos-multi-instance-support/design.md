# Main Nacos 多实例支持优化 设计文档

> 本文档定义 Main Nacos 多实例支持优化的技术设计和实现方案。

## 📋 概述

Main Nacos 多实例支持优化通过扩展配置格式，利用 Nacos SDK 原生支持多 `ServerConfig` 的能力，实现多个 Nacos 实例的高可用部署。

核心实现包括：
- 配置结构体扩展：添加 `Addresses` 字段支持多实例地址配置（逗号分隔）
- 利用 Nacos SDK 能力：Nacos SDK 原生支持多个 `ServerConfig`，自动处理故障转移和负载均衡
- 向后兼容：保持单实例配置格式完全兼容

**参考资源**：
- [Nacos SDK Go 文档](https://github.com/nacos-group/nacos-sdk-go)
- Nacos SDK 的 `ServerConfigs` 参数支持数组，可以传入多个 `ServerConfig`
- BMP 模块实现参考：`ttpos-bmp/internal/pkg/nacos/service/rpc_service.go`

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- ✅ 遵循 Go 标准库和常用第三方库规范
- ✅ 保持代码结构清晰，职责分离
- ✅ 错误处理使用标准 error
- ✅ 不使用 panic，返回 error
- ✅ 使用中文注释

---

## 🔄 代码复用分析

### 可复用的现有组件

- **配置结构**: `main/config/types.go:NacosConf` - 扩展添加 `Addresses` 字段
- **配置加载**: `main/config/config.go:nacosConf()` - 扩展支持读取 `NACOS_SERVER_ADDRESSES` 环境变量
- **Nacos 客户端**: `main/pkg/nacos/nacos.go:NewNacosClient()` - 修改支持多个 `ServerConfig`
- **BMP 模块参考**: `ttpos-bmp/internal/pkg/nacos/service/rpc_service.go` - 参考配置格式和解析逻辑

### Nacos SDK 多实例支持确认

根据 [Nacos SDK Go 文档](https://github.com/nacos-group/nacos-sdk-go)，`ServerConfigs` 参数支持数组，可以传入多个 `ServerConfig`：

```go
// Nacos SDK 支持多个 ServerConfig
sc := []constant.ServerConfig{
    *constant.NewServerConfig("host1", 8848),
    *constant.NewServerConfig("host2", 8848),
    *constant.NewServerConfig("host3", 8848),
}

// 创建客户端时传入多个 ServerConfig
configClient, err := clients.NewConfigClient(
    vo.NacosClientParam{
        ClientConfig:  &cc,
        ServerConfigs: sc,  // 支持多个 ServerConfig
    },
)
```

**关键发现**：
- ✅ Nacos SDK 的 `ServerConfigs` 参数支持数组，可以传入多个 `ServerConfig`
- ✅ 底层 Nacos SDK 自动处理多实例连接和故障转移
- ✅ 无需实现客户端池和健康检查，SDK 已内置支持

### 集成点

- **配置结构体**: 扩展 `NacosConf` 结构体，添加 `Addresses` 字段
- **配置加载**: 修改 `nacosConf()` 函数，支持读取 `NACOS_SERVER_ADDRESSES` 环境变量
- **客户端初始化**: 修改 `NewNacosClient()` 函数，支持解析多个地址并创建多个 `ServerConfig`

---

## 🏗️ 架构设计

### 当前架构

```
main/config/config.go
  └── nacosConf()
      └── 读取环境变量：NACOS_SERVER_IP + NACOS_SERVER_PORT
          └── 创建 NacosConf{Host, Port}
              └── main/pkg/nacos/nacos.go
                  └── NewNacosClient()
                      └── 创建单个 ServerConfig
                          └── Nacos SDK 客户端
```

### 目标架构

```
main/config/config.go
  └── nacosConf()
      └── 优先读取：NACOS_SERVER_ADDRESSES（多实例）
      └── 兼容读取：NACOS_SERVER_IP + NACOS_SERVER_PORT（单实例）
          └── 创建 NacosConf{Addresses, Host, Port}
              └── main/pkg/nacos/nacos.go
                  └── NewNacosClient()
                      └── 解析 Addresses，创建多个 ServerConfig
                          └── Nacos SDK 客户端（多实例）
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
- ✅ 实现简单（只需修改配置解析和客户端初始化）

---

## 🔧 实现方案

### 方案：扩展配置格式 + 利用 Nacos SDK 原生多实例支持（推荐）

**实现思路**：
1. 扩展 `NacosConf` 结构体，添加 `Addresses` 字段
2. 修改配置加载逻辑，支持读取 `NACOS_SERVER_ADDRESSES` 环境变量
3. 修改 `NewNacosClient()` 函数，支持解析多个地址并创建多个 `ServerConfig`
4. **无需实现客户端池和健康检查**，Nacos SDK 已内置支持

**参考实现**：
- [Nacos SDK Go 文档](https://github.com/nacos-group/nacos-sdk-go) - `ServerConfigs` 参数支持数组
- Nacos SDK 自动处理多实例连接、故障转移和负载均衡

**配置格式设计**：

环境变量配置：
```bash
# 多实例配置（优先使用）
NACOS_SERVER_ADDRESSES=host1:port1,host2:port2,host3:port3

# 兼容旧配置（单实例）
NACOS_SERVER_IP=host1
NACOS_SERVER_PORT=8848
```

**代码实现**：

### 1. 扩展配置结构体

```go
// main/config/types.go

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

### 2. 修改配置加载逻辑

```go
// main/config/config.go

func nacosConf(opt copier.Option) {
    Nacos = NacosConf{
        Addresses: "",
        Host:      "localhost",
        Port:      8848,
        Namespace: "",
        Username:  "",
        Password:  "",
        DataId:    "",
        Group:     "",
    }
    copier.CopyWithOption(&Nacos, NacosConf{
        Addresses: viper.GetString("NACOS_SERVER_ADDRESSES"),  // 多实例配置（优先）
        Host:      viper.GetString("NACOS_SERVER_IP"),        // 兼容旧配置
        Port:      viper.GetInt("NACOS_SERVER_PORT"),         // 兼容旧配置
        Namespace: viper.GetString("NACOS_NAMESPACE"),
        Username:  viper.GetString("NACOS_USERNAME"),
        Password:  viper.GetString("NACOS_PASSWORD"),
        DataId:    viper.GetString("NACOS_DATAID"),
        Group:     viper.GetString("NACOS_GROUP"),
    }, opt)
}
```

### 3. 修改客户端初始化逻辑

```go
// main/pkg/nacos/nacos.go

import (
    "strings"
    "strconv"
    // ... 其他导入
)

// NewNacosClient 创建nacos客户端
// 参数：nacosConfig nacos配置
// 返回：NacosClient实例，错误信息
func NewNacosClient(nacosConfig config.NacosConf) (*NacosClient, error) {
    // 创建nacos客户端配置
    cc := *constant.NewClientConfig(
        constant.WithNamespaceId(nacosConfig.Namespace),
        constant.WithTimeoutMs(5000),
        constant.WithNotLoadCacheAtStart(true),
        constant.WithLogDir("./log/nacos"),
        constant.WithCacheDir("./tmp/nacos/cache"),
        constant.WithLogLevel("info"),
        constant.WithUsername(nacosConfig.Username),
        constant.WithPassword(nacosConfig.Password),
    )

    // 解析多个地址，创建多个 ServerConfig
    var serverConfigs []constant.ServerConfig
    if len(nacosConfig.Addresses) > 0 {
        // 多实例配置
        addresses := strings.Split(nacosConfig.Addresses, ",")
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
        if validCount > 0 {
            logger.Logger.Info("使用多实例 Nacos 配置", 
                zap.Int("count", validCount), 
                zap.String("addresses", nacosConfig.Addresses))
        } else {
            logger.Logger.Warn("多实例配置中没有有效地址，将使用单实例配置")
            // 回退到单实例配置
            serverConfigs = []constant.ServerConfig{
                *constant.NewServerConfig(nacosConfig.Host, uint64(nacosConfig.Port)),
            }
        }
    } else {
        // 单实例配置（向后兼容）
        serverConfigs = []constant.ServerConfig{
            *constant.NewServerConfig(nacosConfig.Host, uint64(nacosConfig.Port)),
        }
        logger.Logger.Debug("使用单实例 Nacos 配置", 
            zap.String("host", nacosConfig.Host), 
            zap.Int("port", nacosConfig.Port))
    }

    // 创建配置客户端
    configClient, err := clients.NewConfigClient(
        vo.NacosClientParam{
            ClientConfig:  &cc,
            ServerConfigs: serverConfigs,  // 传入多个 ServerConfig
        },
    )
    if err != nil {
        return nil, err
    }

    // 创建服务发现客户端
    namingClient, err := clients.NewNamingClient(
        vo.NacosClientParam{
            ClientConfig:  &cc,
            ServerConfigs: serverConfigs,  // 传入多个 ServerConfig
        },
    )
    if err != nil {
        logger.Logger.Error("创建nacos服务发现客户端失败:", zap.Error(err))
        return nil, err
    }

    return &NacosClient{
        configClient: configClient,
        namingClient: namingClient,
        nacosConfig:  nacosConfig,
    }, nil
}
```

---

## 📊 数据流设计

### 配置解析流程

```
环境变量读取
  ├── NACOS_SERVER_ADDRESSES (优先)
  │   └── 解析逗号分隔的地址
  │       └── 验证格式 (host:port)
  │           └── 创建多个 ServerConfig
  │
  └── NACOS_SERVER_IP + NACOS_SERVER_PORT (兼容)
      └── 创建单个 ServerConfig
```

### 客户端初始化流程

```
NewNacosClient()
  ├── 创建 ClientConfig
  ├── 解析 ServerConfigs
  │   ├── 多实例：解析 Addresses → 多个 ServerConfig
  │   └── 单实例：Host + Port → 单个 ServerConfig
  ├── 创建 ConfigClient (传入多个 ServerConfig)
  └── 创建 NamingClient (传入多个 ServerConfig)
      └── Nacos SDK 自动处理多实例连接、故障转移、负载均衡
```

---

## 🔒 错误处理

### 配置解析错误

- **无效地址格式**: 记录警告日志，忽略该地址，继续处理其他地址
- **无效端口格式**: 记录警告日志，忽略该地址，继续处理其他地址
- **所有地址都无效**: 记录警告日志，回退到单实例配置（如果存在）

### 客户端初始化错误

- **ConfigClient 创建失败**: 返回错误，中断初始化
- **NamingClient 创建失败**: 返回错误，中断初始化
- **部分实例不可用**: 由 Nacos SDK 自动处理，记录日志

---

## 🧪 测试策略

### 单元测试

- **配置解析测试**: 测试多实例配置解析、单实例配置解析、无效地址处理
- **客户端初始化测试**: 测试多实例客户端创建、单实例客户端创建

### 集成测试

- **多实例配置测试**: 在测试环境部署 Nacos 集群，验证多实例配置功能
- **故障转移测试**: 模拟主实例故障，验证自动切换到备用实例
- **向后兼容性测试**: 使用单实例配置格式，验证功能正常

---

## 📝 实现检查清单

- [ ] 扩展 `NacosConf` 结构体，添加 `Addresses` 字段
- [ ] 修改 `nacosConf()` 函数，支持读取 `NACOS_SERVER_ADDRESSES` 环境变量
- [ ] 修改 `NewNacosClient()` 函数，支持解析多个地址并创建多个 `ServerConfig`
- [ ] 实现地址格式验证逻辑
- [ ] 添加日志记录（多实例配置、单实例配置、无效地址警告）
- [ ] 单元测试：配置解析逻辑
- [ ] 集成测试：多实例配置、故障转移、向后兼容性
- [ ] 更新相关文档

---

**版本**: v1.0.0  
**创建日期**: 2025-11-24  
**维护者**: 后端开发组  
**相关规范**: `.cursor/rules/go-main.mdc`, `.cursor/rules/specs.mdc`

