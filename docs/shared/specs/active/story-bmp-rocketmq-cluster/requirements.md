# 需求规格说明书: BMP RocketMQ 集群配置支持

> 对应提案: [bmp-rocketmq-cluster-config](../../../team/proposals/2025-12/bmp-rocketmq-cluster-config.md)

---

## 1. 📋 概览

| 属性 | 内容 |
| :--- | :--- |
| **Spec ID** | `story-bmp-rocketmq-cluster` |
| **版本** | v1.1.0 |
| **状态** | Approved |
| **负责人** | rikugun |
| **目标版本** | v2.10.0 |
| **最后更新** | 2025-12-02 |

---

## 2. 🎯 背景与目标

### 2.1 背景
目前 BMP 模块（如 `ttpos-message`, `ttpos-takeout`, `ttpos-manager`, `ttpos-erp` 等）中的 RocketMQ 服务配置 `nameSrvAdders` 仅支持单例模式，硬编码为 `[ "$ROCKETMQ_NAME_SRV_ADDR" ]`。这导致无法配置 RocketMQ 集群（多 NameServer），存在单点故障风险，不符合生产环境高可用部署的最佳实践。

### 2.2 目标
1.  **消除单点故障**：支持配置多个 NameServer 地址，当部分节点宕机时服务仍能正常运行。
2.  **支持集群部署**：使 BMP 服务能够连接生产环境的 RocketMQ 集群（多 Master 多 Slave）。
3.  **支持 `make conf` 生成集群配置**：修改配置生成脚本，使其能正确处理包含多个地址的环境变量。

---

## 3. 📝 需求详情

### 3.1 用户故事 (User Stories)

| ID | 角色 | 行为 | 目的/价值 | 优先级 |
| :--- | :--- | :--- | :--- | :--- |
| US-01 | 运维工程师 | 在 `.env` 文件中配置 `ROCKETMQ_NAME_SRV_ADDR` 为逗号分隔的多个地址，并运行 `make conf` | 自动生成包含正确 YAML 数组格式的配置文件，无需手动修改生成的 YAML 文件。 | P0 |
| US-02 | 开发人员 | 在本地开发时，仍然可以使用单个 NameServer 地址进行配置 | 保持开发环境的简洁和现有习惯的兼容。 | P1 |

### 3.2 功能需求 (Functional Requirements)

#### FR-01: 配置模板更新
- **描述**: 更新 BMP 服务的配置文件模板 `manifest/config/config.tpl.yaml`。
- **要求**: 将 `nameSrvAdders: [ "$ROCKETMQ_NAME_SRV_ADDR" ]` 修改为 `nameSrvAdders: [ $ROCKETMQ_NAME_SRV_ADDR ]`，以便支持注入带引号的列表字符串。

#### FR-02: `make conf` 脚本增强
- **描述**: 修改 `ttpos-bmp/hack/init_conf.sh` 脚本。
- **逻辑**:
    1. 读取 `ROCKETMQ_NAME_SRV_ADDR` 环境变量。
    2. 检查是否包含逗号 `,` 或分号 `;`。
    3. 如果包含分隔符，将其分割并重组为 YAML 列表项格式（例如 `"192.168.1.1:9876", "192.168.1.2:9876"`）。
    4. 如果是单个地址，确保其被双引号包围（例如 `"127.0.0.1:9876"`）。
    5. 导出处理后的变量供 `envsubst` 使用。

#### FR-03: RocketMQ 客户端适配
- **描述**: 确保 BMP 服务的 RocketMQ 初始化代码能够接收并使用 `[]string` 类型的 `nameSrvAdders` 配置。
- **注意**: 由于配置生成阶段已经将其转为标准的 YAML 列表，GoFrame 应该能自动将其映射为字符串切片，代码层面可能不需要做 `strings.Split`，但需验证。

### 3.3 非功能需求 (Non-Functional Requirements)

- **NFR-01 兼容性**: 必须兼容旧有的单地址配置方式。
- **NFR-02 健壮性**: 脚本应能处理地址前后的空格。

---

## 4. ✅ 验收标准 (Acceptance Criteria)

### AC-01: 单节点配置生成
- **前置条件**: `.env` 中 `ROCKETMQ_NAME_SRV_ADDR=127.0.0.1:9876`。
- **操作**: 运行 `make conf`。
- **预期结果**: `manifest/config/config.yaml` 中包含 `nameSrvAdders: [ "127.0.0.1:9876" ]`。

### AC-02: 多节点配置生成
- **前置条件**: `.env` 中 `ROCKETMQ_NAME_SRV_ADDR=192.168.1.1:9876,192.168.1.2:9876`。
- **操作**: 运行 `make conf`。
- **预期结果**: `manifest/config/config.yaml` 中包含 `nameSrvAdders: [ "192.168.1.1:9876", "192.168.1.2:9876" ]`。

### AC-03: 服务启动验证
- **前置条件**: 生成了包含多节点的配置文件。
- **操作**: 启动服务（如 `make run.message`）。
- **预期结果**: 服务成功启动，且连接到 RocketMQ 集群（可通过日志或抓包验证）。

---

## 5. ⚠️ 风险与假设

- **风险**: Shell 脚本处理字符串可能存在兼容性问题（如不同 OS 的 shell 行为差异），需在 Linux/Bash 环境下验证。
- **假设**: `envsubst` 替换后的文本能被正确识别为 YAML 语法。
