# Bug-251127-005: BMP UUID 生成器在多应用实例间存在 ID 重复问题

## 基本信息

| 字段       | 值                               |
| ---------- | -------------------------------- |
| Bug ID     | bug-251127-005                   |
| 模块       | ttpos-bmp（UUID 工具包）         |
| 严重程度   | High                             |
| 发现版本   | 当前版本                         |
| 发现日期   | 2025-11-27                       |
| 发现者     | rikugun                          |
| 状态       | 🟢 开发完成                      |
| 修复方案   | [solution.md](./solution.md)     |
| 任务清单   | [tasks.md](./tasks.md)           |

---

## 问题描述

### 问题概述

`ttpos-bmp/utility/uuid/uuid.go` 中的 `GetID()` 方法在不同应用实例调用时存在 ID 重复问题。

**根本原因**：
- 使用 `go-id` 库（Snowflake 算法）生成 ID
- 节点 ID 仅通过 `SERVER_ID` 配置或默认值 1 设置
- 4 位 nodeBits 只支持 1-15 个节点
- **多个应用（ttpos-erp、ttpos-message 等）可能配置相同的 SERVER_ID 或都使用默认值 1**

---

### 问题详情

**现状代码**：

```go
// ttpos-bmp/utility/uuid/uuid.go
func InitIdGenerator(ctx context.Context) {
    idGenerator = goid.NewID()
    serverID := uint32(g.Cfg().MustGetWithEnv(ctx, "SERVER_ID", "1").Int())
    if serverID < 1 || serverID > 15 {
        serverID = 1  // 默认值
    }
    idGenerator.SetNode(serverID, 4)  // 4 位 nodeBits
}
```

**问题场景**：

1. **场景 1：不同应用使用相同 SERVER_ID**
   - ttpos-erp 配置 `SERVER_ID=1`
   - ttpos-message 配置 `SERVER_ID=1`
   - 两个应用在同一毫秒生成 ID 时会产生相同 ID

2. **场景 2：未配置 SERVER_ID 使用默认值**
   - ttpos-erp 未配置 `SERVER_ID`（默认 1）
   - ttpos-message 未配置 `SERVER_ID`（默认 1）
   - 结果同上

3. **场景 3：同一应用多实例**
   - ttpos-erp 实例 A：`SERVER_ID=1`
   - ttpos-erp 实例 B：`SERVER_ID=1`（水平扩展）
   - 两个实例在同一毫秒生成 ID 时会产生相同 ID

---

### Snowflake 算法说明

Snowflake ID 结构（64 位）：

```
| 1 bit | 41 bits    | 4 bits  | 18 bits |
| sign  | timestamp  | node_id | sequence |
```

- **timestamp**：毫秒级时间戳
- **node_id**：节点标识（当前 4 位，1-15）
- **sequence**：同毫秒内序列号

**重复条件**：当 `timestamp` 和 `node_id` 相同时，不同实例的 `sequence` 会从 0 开始递增，导致 ID 重复。

---

## 影响范围

### 影响应用

| 应用 | 使用 InitIdGenerator | 风险 |
| ---- | -------------------- | ---- |
| ttpos-erp | ✅ `internal/boot/boot.go` | 🔴 高 |
| ttpos-message | ✅ `internal/boot/boot.go` | 🔴 高 |
| ttpos-manager | ❓ 待确认 | ⚠️ 中 |
| ttpos-shop | ❓ 待确认 | ⚠️ 中 |
| ttpos-takeout | ❓ 待确认 | ⚠️ 中 |
| ttpos-websocket | ❓ 待确认 | ⚠️ 中 |

### 影响业务

- **数据完整性**：唯一键冲突导致数据插入失败
- **业务连续性**：关键业务流程可能中断
- **数据一致性**：ID 冲突可能导致数据覆盖或错乱

### 影响数据表

所有使用 `uuid.GetID()` 或 `uuid.MustGetID()` 生成主键的表都受影响。

---

## 复现步骤

1. 启动 ttpos-erp 实例（`SERVER_ID=1` 或默认）
2. 启动 ttpos-message 实例（`SERVER_ID=1` 或默认）
3. 在两个应用中同时（或接近同时）调用 `uuid.GetID()`
4. 观察生成的 ID，存在重复可能

---

## 环境信息

- **语言**: Go 1.23+
- **框架**: GoFrame 2.x
- **依赖库**: `github.com/ace-zhaoy/go-id`
- **部署方式**: Docker 容器 / K8s

---

## 初步分析

### 问题根因

1. **nodeBits 不足**：4 位只能区分 15 个节点，无法支持多应用 + 多实例
2. **缺少应用类型区分**：不同应用共享 `SERVER_ID` 配置空间
3. **无自动节点分配**：依赖手动配置，容易出错
4. **默认值风险**：未配置时默认为 1，多个应用默认冲突

### 修复方向

1. **方案 A：扩展 nodeBits + 应用类型前缀**（推荐）
   - 扩展 nodeBits 为 8-10 位
   - 高位为应用类型，低位为实例 ID
   
2. **方案 B：使用 Redis 分配节点 ID**
   - 通过 Redis 自增分配唯一节点 ID
   - 支持动态扩缩容
   
3. **方案 C：切换为 ULID/UUID**
   - 放弃 Snowflake，使用不依赖节点 ID 的方案
   - 需评估对现有系统的影响

---

## 相关链接

### 相关代码

- `ttpos-bmp/utility/uuid/uuid.go`
- `ttpos-bmp/app/ttpos-erp/internal/boot/boot.go`
- `ttpos-bmp/app/ttpos-message/internal/boot/boot.go`

### 参考资料

- [Snowflake 算法原理](https://en.wikipedia.org/wiki/Snowflake_ID)
- [go-id 库文档](https://github.com/ace-zhaoy/go-id)

---

## 下一步行动

1. ✅ **问题分析**（当前阶段）
   - [x] 确认问题根因
   - [x] 评估影响范围
   
2. 🚧 **制定修复方案**
   - [x] 设计技术方案
   - [ ] 评估各方案优缺点
   - [ ] 选择最终方案
   
3. ⏸️ **实施修复**
   - 按照任务清单逐项实现
   
4. ⏸️ **测试验证**
   - 验证 ID 唯一性
   - 压力测试
   
5. ⏸️ **归档**
   - 使用 `/bug-archive` 归档已修复的 Bug

---

**创建时间**: 2025-11-27  
**最后更新**: 2025-11-27  
**创建者**: rikugun  
**当前阶段**: 分析中 → 制定修复方案


