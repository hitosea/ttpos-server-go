---
name: bmp-autogen
description: 自动生成 BMP 模块 service 和 protobuf 代码。当 AI Agent 在 ttpos-bmp 模块中修改 logic 或 protobuf 文件后，自动触发相应代码生成流程。
---

# BMP 模块代码自动生成

## 触发条件

**AI Agent 触发**（自动执行）：

当 AI Agent 修改以下文件时，**必须立即执行代码生成**：

- **Logic 文件修改**: `ttpos-bmp/app/*/internal/logic/*.go`
  → 执行 `make service` 生成服务接口代码

- **Protobuf 文件修改**: `ttpos-bmp/app/*/manifest/protobuf/*.proto`
  → 执行 `make pb` 生成 Go 代码

**用户手动触发**（可选）：

当用户明确要求时：
- "重新生成 service 代码"
- "重新生成 pb 代码"
- "生成所有模块代码"

## Agent 执行规范

### 1. Logic 文件修改后的自动执行

**触发场景**: Agent 使用 `edit`/`write` 工具修改了 `internal/logic/` 下的 Go 文件

**执行步骤**:

```bash
# 1. 识别模块名（从文件路径提取）
# 例如: ttpos-bmp/app/ttpos-manager/internal/logic/user.go
# 模块名: ttpos-manager

# 2. 进入对应模块目录
cd ttpos-bmp/app/{module-name}

# 3. 执行生成命令
make service

# 4. 验证生成的代码
ls -la internal/service/
```

**关键点**:
- ✅ 必须在修改 logic 文件后**立即**执行
- ✅ 只在对应模块目录执行，不需要遍历所有模块
- ✅ 执行前检查 Makefile 是否存在
- ✅ 检查生成命令是否成功

**示例**:

```go
// Agent 修改了这个文件:
// ttpos-bmp/app/ttpos-manager/internal/logic/user.go

// Agent 必须立即执行:
cd ttpos-bmp/app/ttpos-manager && make service

// 生成: internal/service/ 目录下的服务接口代码
```

### 2. Protobuf 文件修改后的自动执行

**触发场景**: Agent 使用 `edit`/`write` 工具修改了 `manifest/protobuf/` 下的 `.proto` 文件

**执行步骤**:

```bash
# 1. 识别模块名（从文件路径提取）
# 例如: ttpos-bmp/app/ttpos-manager/manifest/protobuf/user.proto
# 模块名: ttpos-manager

# 2. 进入对应模块目录
cd ttpos-bmp/app/{module-name}

# 3. 执行生成命令
make pb

# 4. 验证生成的代码
ls -la api/
ls -la internal/controller/rpc/
```

**关键点**:
- ✅ 必须在修改 protobuf 文件后**立即**执行
- ✅ 只在对应模块目录执行
- ✅ 执行前检查 Makefile 是否存在
- ✅ 检查生成命令是否成功

**示例**:

```protobuf
// Agent 修改了这个文件:
// ttpos-bmp/app/ttpos-manager/manifest/protobuf/user.proto

// Agent 必须立即执行:
cd ttpos-bmp/app/ttpos-manager && make pb

// 生成: api/ 和 internal/controller/rpc/ 下的代码
```

## 代码生成规则

### Logic 修改 → 生成 Service 代码

**监控目录**: `ttpos-bmp/app/*/internal/logic/`

**执行命令**:
```bash
cd ttpos-bmp/app/{module-name} && make service
```

**生成的文件**:
- `internal/service/` - 服务接口定义（自动生成，**禁止手动修改**）

**失败处理**:
- 如果 `gf` 命令不存在 → 提示安装 GoFrame CLI
- 如果 Makefile 不存在 → 提示检查模块路径
- 如果生成失败 → 显示错误日志，但不阻止任务继续

### Protobuf 修改 → 生成 Go 代码

**监控目录**: `ttpos-bmp/app/*/manifest/protobuf/`

**执行命令**:
```bash
cd ttpos-bmp/app/{module-name} && make pb
```

**生成的文件**:
- `api/` - API 定义文件
- `internal/controller/rpc/` - gRPC 控制器

**失败处理**:
- 如果 `protoc` 不存在 → 提示安装 protobuf 工具
- 如果生成失败 → 显示错误日志，但不阻止任务继续

## 模块列表

| 模块名称 | Logic 目录 | Protobuf 目录 |
|---------|-----------|--------------|
| ttpos-erp | `app/ttpos-erp/internal/logic/` | `app/ttpos-erp/manifest/protobuf/` |
| ttpos-manager | `app/ttpos-manager/internal/logic/` | `app/ttpos-manager/manifest/protobuf/` |
| ttpos-message | `app/ttpos-message/internal/logic/` | `app/ttpos-message/manifest/protobuf/` |
| ttpos-shop | `app/ttpos-shop/internal/logic/` | `app/ttpos-shop/manifest/protobuf/` |
| ttpos-takeout | `app/ttpos-takeout/internal/logic/` | `app/ttpos-takeout/manifest/protobuf/` |
| ttpos-websocket | `app/ttpos-websocket/internal/logic/` | `app/ttpos-websocket/manifest/protobuf/` |

## 手动执行方式（仅当用户要求时）

### 生成单个模块的代码

```bash
# 生成 ttpos-manager 的 service 代码
cd ttpos-bmp/app/ttpos-manager && make service

# 生成 ttpos-manager 的 protobuf 代码
cd ttpos-bmp/app/ttpos-manager && make pb
```

### 生成所有模块的代码

```bash
# 在 ttpos-bmp 根目录执行
cd ttpos-bmp

# 生成所有模块的 service 代码
for module in app/*/; do
    echo "Generating service for $module"
    (cd "$module" && make service)
done

# 生成所有模块的 protobuf 代码
for module in app/*/; do
    echo "Generating pb for $module"
    (cd "$module" && make pb)
done
```

### 使用 watch.sh 脚本（可选）

`watch.sh` 是一个辅助脚本，**仅在需要持续监控时使用**（不推荐，应依赖 Agent 自动执行）：

```bash
cd ttpos-bmp
./watch.sh watch      # 监控文件变化
./watch.sh service    # 生成所有 service
./watch.sh pb         # 生成所有 protobuf
./watch.sh all        # 生成所有代码
```

详见 [watch.sh 使用文档](../../ttpos-bmp/docs/watch-script.md)

## Agent 工作流示例

### 场景 1: 添加新的 Logic 方法

```
Agent 操作:
1. 读取 ttpos-bmp/app/ttpos-manager/internal/logic/user.go
2. 添加新的业务方法
3. 使用 edit 工具修改文件

自动执行:
→ cd ttpos-bmp/app/ttpos-manager && make service

结果:
→ internal/service/ 目录下生成新的服务接口代码
```

### 场景 2: 修改 Protobuf 定义

```
Agent 操作:
1. 读取 ttpos-bmp/app/ttpos-manager/manifest/protobuf/user.proto
2. 修改或添加 message/service 定义
3. 使用 edit/write 工具修改文件

自动执行:
→ cd ttpos-bmp/app/ttpos-manager && make pb

结果:
→ api/ 和 internal/controller/rpc/ 目录下生成新的 Go 代码
```

### 场景 3: 批量修改多个 Logic 文件

```
Agent 操作:
1. 修改 app/ttpos-manager/internal/logic/user.go
2. 修改 app/ttpos-manager/internal/logic/order.go
3. 修改 app/ttpos-manager/internal/logic/product.go

自动执行:
→ cd ttpos-bmp/app/ttpos-manager && make service

注意:
→ 只需执行一次，不是每个文件修改后都执行
→ 在所有文件修改完成后，执行一次即可
```

## 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| `gf` 命令找不到 | GoFrame CLI 未安装或 PATH 未配置 | 参考 `bmp-tools-install` skill 安装 |
| `make service` 后没有生成代码 | logic 文件缺少必要的注释或方法签名 | 检查 logic 文件格式，确保遵循 GoFrame 规范 |
| `make pb` 报错 | protoc 或相关插件未安装 | 安装 protobuf 编译器和 Go 插件 |
| 生成的代码和预期不一致 | GoFrame 版本问题 | 确保使用 GoFrame v2.x |
| 执行 `make service` 没有反应 | Makefile 路径错误或 gf 命令失败 | 检查 Makefile 是否存在，确认 gf 命令可用 |

## 验证生成的代码

### Service 代码验证

```bash
# 检查 service 文件是否生成
ls -la ttpos-bmp/app/{module}/internal/service/

# 检查是否有包含 "Code generated and maintained by GoFrame CLI tool" 的注释
head -n 5 ttpos-bmp/app/{module}/internal/service/*.go
```

### Protobuf 代码验证

```bash
# 检查 api 文件是否生成
ls -la ttpos-bmp/app/{module}/api/

# 检查 rpc controller 是否生成
ls -la ttpos-bmp/app/{module}/internal/controller/rpc/
```

## Agent 注意事项

### ✅ 必须执行

- [ ] 修改任何 `internal/logic/*.go` 文件后，立即在对应模块执行 `make service`
- [ ] 修改任何 `manifest/protobuf/*.proto` 文件后，立即在对应模块执行 `make pb`
- [ ] 验证生成命令是否成功（检查退出码）
- [ ] 如果生成失败，记录错误并提示用户

### ❌ 不要执行

- [ ] 不需要在每次保存文件后执行（只在实际修改内容后执行）
- [ ] 不需要遍历所有模块执行（只在修改的模块执行）
- [ ] 不需要自动安装依赖工具（失败时提示用户手动安装）
- [ ] 不需要运行 `watch.sh` 脚本（除非用户明确要求）

### 🔄 执行时机

1. **单个文件修改**: 修改后立即执行
2. **批量修改**: 所有文件修改完成后执行一次
3. **跨模块修改**: 每个模块分别执行

## 最佳实践

### Agent 角色最佳实践

1. **修改 Logic 文件时**:
   - 完成所有修改后，**立即**执行 `make service`
   - 验证生成的 service 文件是否包含新方法
   - 如果有编译错误，检查 logic 文件格式

2. **修改 Protobuf 文件时**:
   - 完成所有修改后，**立即**执行 `make pb`
   - 验证生成的 api 和 rpc 文件是否符合预期
   - 如果有编译错误，检查 proto 语法

3. **生成失败时**:
   - 检查错误日志
   - 提示用户安装缺失的工具
   - 不要阻止后续任务继续

4. **性能考虑**:
   - 只在修改的模块执行
   - 批量修改时只执行一次
   - 避免不必要的重复执行

## 相关文档

- [GoFrame CLI 文档](https://goframe.org/docs/cli)
- [Go 代码开发规范](../../ttpos-bmp/.cursor/rules/go-rules.mdc)
- [Protobuf 开发规范](../../ttpos-bmp/.cursor/rules/proto-rules.mdc)
- [BMP 工具安装](../bmp-tools-install/)
- [watch.sh 使用文档](../../ttpos-bmp/docs/watch-script.md)
