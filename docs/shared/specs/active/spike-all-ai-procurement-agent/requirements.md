# AI 智能采购 Agent 技术调研 需求文档

## 基本信息

| 项目              | 内容                              |
| ----------------- | --------------------------------- |
| **Spec ID**       | spike-all-ai-procurement-agent    |
| **Level**         | spike（技术调研）                  |
| **来源 Proposal** | 无（技术驱动）                     |
| **创建日期**      | 2026-03-05                        |
| **负责人**        | weifashi                          |
| **目标 Sprint**   | -                                 |

## 审核状态

| 项目         | 内容   |
| ------------ | ------ |
| **审核状态** | 待审核 |
| **审核人**   | -      |
| **审核日期** | -      |

---

## 调研背景

### 现状

TTPOS 已有完整的采购链路：供应商管理 → 原料管理 → 采购单创建 → 审批 → 收货入库 → 库存管理，并集成 ERPNext 实现供应商/仓库/库存的双向同步。

**当前痛点**：所有采购决策完全依赖人工判断，包括：
- 何时需要补货（靠经验看库存）
- 补多少货（凭感觉估算）
- 向哪个供应商下单（人工比价）
- 异常消耗无法及时发现

### 目标

验证基于 LangGraph 框架构建 AI 智能采购 Agent 的技术可行性，评估其在 TTPOS 餐饮采购场景的实际表现，为后续正式开发提供决策依据。

---

## 用户故事

**作为** 技术团队
**我想** 验证 LangGraph 与 TTPOS 现有采购系统的集成可行性
**以便于** 判断是否值得投入资源正式开发 AI 智能采购功能，并为后续 AI 能力（点餐助手、报表分析等）铺路

---

## 调研目标

### Objective 1: LangGraph 与 TTPOS API 集成可行性

**描述**: 验证 Python LangGraph Agent 能否通过现有 REST API 读取库存数据、供应商信息，并创建采购单。

#### 验收标准

1. **WHEN** Agent 启动 **THEN** Agent **SHALL** 通过 `/warehouse/material/list` 成功获取当前库存数据
2. **WHEN** Agent 获取库存数据后 **THEN** Agent **SHALL** 通过 `/supplier/list` 获取可用供应商列表
3. **WHEN** Agent 生成采购建议并获得人工确认 **THEN** Agent **SHALL** 通过 `/purchase/order/create` 成功创建采购单
4. **WHEN** Agent 调用 TTPOS API **THEN** Agent **SHALL** 正确携带 JWT Token 和 company_uuid 完成多租户认证

### Objective 2: 需求预测准确性

**描述**: 验证基于历史销售数据 + 原料 BOM，LLM + 时序分析能否有效预测原料用量。

#### 验收标准

1. **WHEN** 提供过去 30 天的销售数据 **THEN** 预测模型 **SHALL** 输出未来 7 天各原料的预计消耗量
2. **WHEN** 对比预测值与实际消耗 **THEN** 预测偏差 **SHALL** 记录在评估报告中（目标 ≤ 20%）
3. **IF** 存在节假日或特殊活动 **THEN** 预测模型 **SHALL** 能够通过 prompt 注入日历因子进行调整

### Objective 3: 端到端流程跑通

**描述**: 验证从库存检测到采购单创建的完整链路可正常运作。

#### 验收标准

1. **WHEN** 某原料库存低于安全库存 **THEN** Agent **SHALL** 自动触发采购建议生成流程
2. **WHEN** Agent 生成采购建议 **THEN** 建议 **SHALL** 包含：原料名称、建议采购量、推荐供应商、预估金额
3. **WHEN** 建议生成后 **THEN** Agent **SHALL** 暂停等待人工确认（Human-in-the-Loop）
4. **WHEN** 人工确认通过 **THEN** Agent **SHALL** 按建议内容调用 TTPOS API 创建采购单
5. **WHEN** 人工驳回 **THEN** Agent **SHALL** 终止流程并记录驳回原因

### Objective 4: LLM 成本与响应速度

**描述**: 评估单次采购决策的 Token 消耗、延迟、费用是否在可接受范围内。

#### 验收标准

1. **WHEN** Agent 完成一次完整的采购建议生成 **THEN** 评估报告 **SHALL** 记录：总 Token 消耗、LLM 调用次数、端到端延迟
2. **WHEN** 对比不同 LLM（Claude / GPT-4o / GPT-4o-mini）**THEN** 评估报告 **SHALL** 记录各模型的效果差异和成本对比
3. **WHEN** 评估完成 **THEN** 报告 **SHALL** 给出单次决策的预估费用和月度运营成本估算

---

## 调研范围

### In Scope

| 范围项 | 说明 |
|--------|------|
| LangGraph PoC 开发 | 构建最小可运行原型，演示核心流程 |
| TTPOS API 集成 | 复用现有 REST API（库存、供应商、采购单） |
| 需求预测验证 | 基于真实/模拟销售数据验证预测能力 |
| 成本评估 | 不同 LLM 的效果和成本对比 |
| 技术评估报告 | 架构方案、性能数据、风险评估、Go/No-Go 建议 |

### Out of Scope

| 排除项 | 原因 |
|--------|------|
| 生产环境部署 | Spike 阶段仅做验证，不上生产 |
| 前端 UI | 不涉及 Flutter/Web 前端开发 |
| 新数据表设计 | 完全复用现有表结构和 API |
| 多租户全面支持 | PoC 用单一测试商户验证 |
| 自然语言交互 UI | 仅验证后端 Agent 能力 |

---

## 技术方案概要

### 架构

```
┌────────────────────────────────────────────────────┐
│  AI Procurement Agent Service (Python)             │
│                                                     │
│  ┌───────────────────────────────────────────┐     │
│  │          LangGraph Workflow Engine         │     │
│  │                                            │     │
│  │  Nodes:                                    │     │
│  │    - collect_data      (采集库存/销售数据)   │     │
│  │    - forecast_demand   (需求预测)           │     │
│  │    - compare_stock     (库存对比)           │     │
│  │    - match_supplier    (供应商匹配+限额)     │     │
│  │    - generate_proposal (生成采购建议)        │     │
│  │    - human_review      (人工审批节点)        │     │
│  │    - create_po         (创建采购单)          │     │
│  │    - detect_anomaly    (异常检测)           │     │
│  │                                            │     │
│  │  Tools:                                    │     │
│  │    - ttpos_api_client  (TTPOS REST API)    │     │
│  │    - forecaster        (时序预测)           │     │
│  │    - notifier          (消息推送)           │     │
│  └───────────┬───────────────────────────────┘     │
│              │                                      │
│         ┌────▼────┐                                 │
│         │   LLM   │  Claude / GPT-4o               │
│         └─────────┘                                 │
└──────────┬─────────────────────────────────────────┘
           │ REST API (JWT Auth)
           ▼
┌──────────────────────┐
│  TTPOS Main (Go/Gin) │
│  现有采购相关 API：    │
│  /warehouse/material/ │
│  /supplier/           │
│  /purchase/order/     │
│  /purchase/limit/     │
└──────────────────────┘
```

### LangGraph 工作流

```
START
  │
  ▼
[collect_data] ── 调 TTPOS API 获取库存、销售、BOM 数据
  │
  ▼
[forecast_demand] ── LLM + 时序分析：预测未来 7 天用量
  │
  ▼
[compare_stock] ── 预测量 vs 当前库存 vs 安全库存
  │
  ├─ 不需要采购 ──► [detect_anomaly] ──► END
  │
  ▼ 需要采购
[match_supplier] ── 匹配供应商 + 校验采购限额
  │
  ▼
[generate_proposal] ── 按供应商分组，生成采购建议单
  │
  ▼
[human_review] ── Human-in-the-Loop：暂停等待人工确认
  │
  ├─ 驳回 ──► END (记录原因)
  │
  ▼ 通过
[create_po] ── 调 TTPOS /purchase/order/create
  │
  ▼
END
```

### 依赖的现有 TTPOS API

| API | 用途 | 模块 |
|-----|------|------|
| `GET /shop/warehouse/material/list` | 获取库存水位 | Main |
| `GET /shop/supplier/list` | 获取供应商列表 | Main |
| `GET /shop/supplier/select` | 供应商选择器 | Main |
| `GET /shop/purchase/limit/scheme/list` | 获取采购限额 | Main |
| `POST /shop/purchase/order/create` | 创建采购单 | Main |
| `GET /shop/purchase/order/list` | 查历史采购单 | Main |
| 销售数据接口（待确认） | 获取历史销售数据 | Main |

### 技术栈

| 组件 | 技术选型 | 版本 |
|------|---------|------|
| Agent 框架 | LangGraph (Python) | 1.0+ |
| LLM | Claude API / OpenAI API | - |
| HTTP 客户端 | httpx | - |
| 运行时 | Python | 3.11+ |
| 容器化 | Docker | - |
| 与 TTPOS 通信 | REST API (JWT Auth) | - |

---

## 交付物

### 1. PoC Demo（可运行原型）

- 独立 Python 项目，Docker 化部署
- 演示完整流程：库存检测 → 需求预测 → 采购建议 → 人工确认 → 创建采购单
- 使用测试商户数据（dev 环境）
- 提供命令行 / 简单 Web UI 触发和交互

### 2. 技术评估报告

| 章节 | 内容 |
|------|------|
| 架构方案 | 详细技术架构、部署方案、与 TTPOS 集成方式 |
| 性能数据 | API 调用延迟、LLM 响应时间、端到端耗时 |
| 预测评估 | 需求预测准确率、偏差分析 |
| 成本分析 | Token 消耗、单次决策费用、月度运营成本估算 |
| LLM 对比 | Claude vs GPT-4o vs GPT-4o-mini 效果/成本对比 |
| 风险评估 | 技术风险、业务风险、安全风险 |
| Go/No-Go 建议 | 是否推荐正式开发、推荐的实施路线 |

---

## 非功能需求

### 安全要求

- [ ] Agent 调用 TTPOS API 必须使用 JWT Token 认证
- [ ] 不直接访问数据库，仅通过 API 层交互
- [ ] LLM 调用不泄露商户敏感数据（价格、供应商联系方式等需脱敏）

### 测试要求

- [ ] PoC 代码包含基本单元测试
- [ ] 提供可复现的 Demo 运行脚本

### 平台兼容性

- 不涉及前端终端（纯后端服务）
- Docker 容器化，支持 Linux 部署

---

## 约束条件

### 技术约束

- Agent 服务使用 Python 3.11+ 和 LangGraph 1.0+
- 不新建业务数据表，完全复用现有 TTPOS REST API
- 通过 JWT Token + company_uuid 实现多租户认证
- Human-in-the-Loop：采购建议需人工确认后才创建采购单

### 资源约束

- Spike 阶段不设 Story Point（调研性质，产出为报告和 PoC）
- 预计调研周期：2-3 周

---

## 风险和缓解

### 风险 1: LLM 预测准确性不足

**影响**: 高
**概率**: 中
**缓解措施**: PoC 阶段对比多种策略（纯 LLM 推理 vs LLM + 统计模型混合），如果 LLM 预测效果不佳，可降级为基于规则的安全库存补货策略

### 风险 2: LLM 成本过高

**影响**: 中
**概率**: 中
**缓解措施**: 评估不同 LLM（Claude Haiku / GPT-4o-mini 等低成本模型），优化 prompt 减少 Token 消耗，对于确定性逻辑（库存对比、限额校验）使用代码而非 LLM

### 风险 3: TTPOS API 不满足 Agent 需求

**影响**: 中
**概率**: 低
**缓解措施**: PoC 阶段记录缺失的 API 能力，如需扩展可在正式开发阶段补充接口（如销售数据聚合 API）

### 风险 4: Python 微服务引入运维复杂度

**影响**: 低
**概率**: 中
**缓解措施**: Docker 化部署，与现有容器编排一致；Spike 阶段仅在 dev 环境运行

---

## 分阶段路线（后续参考）

Spike 通过后的建议实施路线：

| 阶段 | 内容 | 前置条件 |
|------|------|---------|
| **P0** | 库存预警 Agent — 低于安全库存自动告警 | Spike 通过 |
| **P1** | 智能建单 — 基于库存 + 安全库存自动生成采购建议 | P0 完成 |
| **P2** | 需求预测 — 接入销售数据，预测驱动采购建议 | P1 完成 |
| **P3** | 自然语言交互 — 对话式采购管理 | P2 完成 |

---

**版本**: v1.0.0
**创建日期**: 2026-03-05
