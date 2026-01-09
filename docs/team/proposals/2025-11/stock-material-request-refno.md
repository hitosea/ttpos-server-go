# stock.SaveMaterialRequestReq 增加 RefNo 字段 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目         | 内容                           |
| ------------ | ------------------------------ |
| **提案人**   | rikugun                        |
| **日期**     | 2025-11-27                     |
| **目标版本** | 待定                           |
| **状态**     | ✅ 已创建 Spec                 |
| **关联任务** | -                              |
| **关联 Spec**| [task-erp-material-request-refno](../../../shared/specs/archived/v2.12/task-erp-material-request-refno/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

当前 `stock.SaveMaterialRequestReq` 消息结构中缺少来源单据号字段，导致以下问题：

1. **排错困难**：当 ttpos 调用 ERP 创建物料申请单时，无法在 ERP 侧追溯到原始的 ttpos 订单号
2. **数据关联断裂**：ttpos 与 ERP 之间的单据无法建立明确的对应关系
3. **问题定位耗时**：出现问题时，需要通过时间、金额等间接条件反向匹配，效率低下

### 业务价值

- **提升排错效率**：通过 RefNo 可直接定位 ttpos 原始订单
- **增强数据可追溯性**：建立 ttpos 订单与 ERP 物料申请单的明确关联
- **降低运维成本**：减少问题排查时间，提高系统可维护性

### 目标用户

- [x] 开发人员（排错定位）
- [x] 运维人员（问题追踪）
- [x] 商户管理员（单据查询）

---

## 💡 解决方案概述

### 方案描述

在 `stock.SaveMaterialRequestReq` protobuf 消息中新增 `ref_no` 字段，用于存储 ttpos 传递的原始订单号。该字段为可选字段，不影响现有调用方。

### 核心功能点

1. 在 `SaveMaterialRequestReq` 消息中新增 `string ref_no = 10;` 字段
2. 字段用途：存储 ttpos 原始订单号，用于跟踪和排错
3. 字段属性：可选，不传时不影响现有业务逻辑

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
- [x] 内部服务调用（ttpos → ttpos-erp）

**涉及模块**：
- [ ] UI 组件
- [x] API 接口（gRPC protobuf）
- [ ] 数据模型
- [x] 业务逻辑（可选：日志记录）
- [ ] 第三方集成

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：仅 protobuf 字段新增，无业务逻辑变更

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1 SP

### 涉及文件

1. `ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto` - 新增字段定义
2. `ttpos-bmp/app/ttpos-erp/api/stock/stock.pb.go` - 自动生成
3. `ttpos-bmp/app/ttpos-erp/api/stock/stock_grpc.pb.go` - 自动生成
4. ttpos 调用方代码 - 传入 ref_no 参数

### 风险识别

**潜在风险**：
1. 无明显风险，字段为可选，向后兼容

**缓解措施**：
1. 保持字段可选，不影响现有调用方

---

## 🔗 相关资源

### 参考代码

当前 `SaveMaterialRequestReq` 定义（`stock.proto` 第 28-38 行）：

```protobuf
message SaveMaterialRequestReq {
  int64 transaction_date = 1;      // 单据日期,必填
  string company_abbr = 2;         // 公司缩写,必填
  string branch = 3;               // 分支名称 必填
  int64 required_by = 4;           // 需求时间,必填
  string source_warehouse = 5;     // 来源仓库，必填
  string target_warehouse = 6;     // 目标仓库，必填
  string purpose = 7;              // 申请目的,可选 默认 Purchase
  string supplier = 8;             // 供应商名称, purpose 为 Purchases时 必填
  repeated MaterialRequestItem items = 9;  // 物品列表
}
```

### 建议修改

```protobuf
message SaveMaterialRequestReq {
  int64 transaction_date = 1;      // 单据日期,必填
  string company_abbr = 2;         // 公司缩写,必填
  string branch = 3;               // 分支名称 必填
  int64 required_by = 4;           // 需求时间,必填
  string source_warehouse = 5;     // 来源仓库，必填
  string target_warehouse = 6;     // 目标仓库，必填
  string purpose = 7;              // 申请目的,可选 默认 Purchase
  string supplier = 8;             // 供应商名称, purpose 为 Purchases时 必填
  repeated MaterialRequestItem items = 9;  // 物品列表
  string ref_no = 10;              // 来源单据号，可选，用于跟踪 ttpos 原始订单号
}
```

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 技术负责人   |        |           |
| 开发代表     |        |           |

### 评审结论

- [ ] ✅ **批准**：进入开发阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 修改 protobuf 文件
- [ ] 执行 `gf gen pb` 重新生成代码
- [ ] 更新调用方代码传入 ref_no
- [ ] 测试验证

---

**版本**: v1.0.0  
**创建日期**: 2025-11-27  
**维护者**: rikugun

