# Bug-251127-005 修复任务清单

> **Bug**: BMP UUID 生成器在多应用实例间存在 ID 重复问题  
> **当前状态**: 🟢 开发完成  
> **开始时间**: 2025-11-27  
> **完成时间**: 2025-11-27

---

## 📋 任务列表

### 阶段 1: 修改 UUID 工具包（1 小时）

#### 1.1 修改 uuid.go

- [x] **定义应用类型常量** `ttpos-bmp/utility/uuid/uuid.go`
  - **内容**:
    - 添加 `AppTypeERP`, `AppTypeMessage`, `AppTypeManager` 等常量
    - 添加 `appTypeBits`, `instanceIDBits`, `totalNodeBits` 常量
  - **实际时间**: 5 分钟
  - **负责人**: AI Agent

- [x] **修改 InitIdGenerator 函数** `ttpos-bmp/utility/uuid/uuid.go`
  - **修改点**:
    - 添加 `appType uint32` 参数
    - 实现节点 ID 组合逻辑：`(appType << 6) | instanceID`
    - 将 nodeBits 从 4 改为 10
    - 添加日志记录
  - **实际时间**: 10 分钟
  - **负责人**: AI Agent

- [x] **添加向后兼容函数（可选）** `ttpos-bmp/utility/uuid/uuid.go`
  - **说明**: 未添加，直接修改所有调用点更简洁
  - **负责人**: AI Agent 

---

### 阶段 2: 更新各应用初始化代码（30 分钟）

#### 2.1 更新 ttpos-erp

- [x] **修改 boot.go** `ttpos-bmp/app/ttpos-erp/internal/boot/boot.go`
  - **修改点**: `uuid.InitIdGenerator(ctx)` → `uuid.InitIdGenerator(ctx, uuid.AppTypeERP)`
  - **实际时间**: 2 分钟
  - **负责人**: AI Agent

#### 2.2 更新 ttpos-message

- [x] **修改 boot.go** `ttpos-bmp/app/ttpos-message/internal/boot/boot.go`
  - **修改点**: `uuid.InitIdGenerator(ctx)` → `uuid.InitIdGenerator(ctx, uuid.AppTypeMessage)`
  - **实际时间**: 2 分钟
  - **负责人**: AI Agent

#### 2.3 检查并更新其他应用

- [x] **检查 ttpos-manager** 是否使用 uuid 包
  - **结果**: 未使用，无需修改
  - **负责人**: AI Agent

- [x] **检查 ttpos-shop** 是否使用 uuid 包
  - **结果**: 未使用，无需修改
  - **负责人**: AI Agent

- [x] **检查 ttpos-takeout** 是否使用 uuid 包
  - **结果**: 未使用，无需修改
  - **负责人**: AI Agent

- [x] **检查 ttpos-websocket** 是否使用 uuid 包
  - **结果**: 未使用，无需修改
  - **负责人**: AI Agent

---

### 阶段 3: 单元测试（30 分钟）

#### 3.1 编写测试用例

- [x] **编写节点 ID 计算测试** `ttpos-bmp/utility/uuid/uuid_test.go`
  - **测试内容**:
    - ✅ `TestNodeIdCalculation` - 验证节点 ID 计算正确
    - ✅ `TestBitsConfiguration` - 验证位数配置
    - ✅ `TestAppTypeConstants` - 验证常量定义
  - **实际时间**: 10 分钟
  - **负责人**: AI Agent

- [x] **编写 ID 唯一性测试** `ttpos-bmp/utility/uuid/uuid_test.go`
  - **测试内容**:
    - ✅ `TestIdUniquenessWithinApp` - 同应用内 10000 个 ID 唯一
    - ✅ `TestIdUniquenessAcrossApps` - 跨应用 ID 无冲突
    - ✅ `TestInitIdGeneratorWithInvalidAppType` - 无效参数处理
    - ✅ `BenchmarkMustGetID` - 性能基准测试
  - **实际时间**: 10 分钟
  - **负责人**: AI Agent 

---

### 阶段 4: 集成测试（30 分钟）

- [ ] **本地集成测试**
  - **测试步骤**:
    1. 启动 ttpos-erp（SERVER_ID=1）
    2. 启动 ttpos-message（SERVER_ID=1）
    3. 调用两个应用的 ID 生成接口
    4. 验证无重复 ID
  - **预计时间**: 30 分钟
  - **负责人**: 

---

### 阶段 5: 文档更新（15 分钟）

- [ ] **更新配置文档**
  - **内容**:
    - 说明 SERVER_ID 配置范围（1-63）
    - 说明多实例部署时的配置要求
  - **预计时间**: 15 分钟
  - **负责人**: 

---

### 阶段 6: 部署上线（1 小时）

- [ ] **Code Review**
  - **检查点**:
    - 节点 ID 计算逻辑正确
    - 所有应用调用点已更新
    - 测试覆盖率
  - **预计时间**: 30 分钟
  - **负责人**: 

- [ ] **逐应用发布**
  - **发布顺序**:
    1. ttpos-websocket（低流量）
    2. ttpos-shop
    3. ttpos-manager
    4. ttpos-takeout
    5. ttpos-message
    6. ttpos-erp（高流量，最后）
  - **预计时间**: 30 分钟
  - **负责人**: 

---

### 阶段 7: Bug 归档（10 分钟）

- [ ] **更新 Bug 状态为「已修复」**
- [ ] **执行 `/bug-archive`**

---

## 📊 任务统计

| 阶段 | 任务数 | 预计时间 | 状态 |
| ---- | ------ | -------- | ---- |
| **阶段 1: UUID 工具包** | 3 | 40 分钟 | ⏸️ 待开始 |
| **阶段 2: 应用初始化** | 6 | 30 分钟 | ⏸️ 待开始 |
| **阶段 3: 单元测试** | 2 | 30 分钟 | ⏸️ 待开始 |
| **阶段 4: 集成测试** | 1 | 30 分钟 | ⏸️ 待开始 |
| **阶段 5: 文档更新** | 1 | 15 分钟 | ⏸️ 待开始 |
| **阶段 6: 部署上线** | 2 | 60 分钟 | ⏸️ 待开始 |
| **阶段 7: Bug 归档** | 2 | 10 分钟 | ⏸️ 待开始 |
| **总计** | **17** | **~3.5 小时** | ⏸️ 待开始 |

---

## 📈 进度跟踪

- **总任务数**: 17
- **已完成**: 0
- **进行中**: 0
- **未开始**: 17
- **完成率**: 0%

---

## 🔗 相关链接

- **Bug 报告**: `bug.md`
- **修复方案**: `solution.md`
- **UUID 工具包**: `ttpos-bmp/utility/uuid/uuid.go`

---

**创建时间**: 2025-11-27  
**创建者**: rikugun


