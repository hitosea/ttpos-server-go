# story-ttpos-erp-payment-mode-save / SaveModeOfPayment 设计说明

## 1. 技术方案概述

- 在 `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto` 新增 `SaveModeOfPayment` RPC 及请求/响应消息。
- 服务端生成支付方式名称：`{channel}-{pay_type}-{4位序号} - {company_abbr}`，序号按 `channel+pay_type` 维度递增；`0000` 预留给系统/同步默认，用户自建从 `0001` 起。
- 序号生成采用唯一性保护（DB 唯一键或分布式锁），失败重试并返回明确错误码。
- 同步策略：TTPOS 发起创建/更新时写 ERP；ERP 变更回流 TTPOS。存量数据不改名。

## 2. 接口定义（proto 草案）

文件：`ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`

```
message SaveModeOfPaymentReq {
  string company_abbr = 1; // 商户简称，必填
  string branch = 2;       // 分支，必填
  string channel = 3;      // 渠道，如 LianLianPay，必填
  string pay_type = 4;     // 支付类型（TTPOS 定义），必填
}

message SaveModeOfPaymentResp {
  string name = 1; // 规范化名称 {channel}-{pay_type}-{NNNN} - {company_abbr}
  string id = 2;   // ERP/TTPOS 唯一标识（若有）
}

service SellingService {
  rpc SaveModeOfPayment (SaveModeOfPaymentReq) returns (erp.ResponseInfo);
}
```

> 若 `erp.ResponseInfo` 需携带 `SaveModeOfPaymentResp`，保持现有封装方式（data 字段 JSON）。

## 3. 名称与序号策略

- 维度：`channel + pay_type`。
- 序号生成：
  - 默认/系统：`0000`（保留位）。
  - 自建：从 `0001` 起递增；存在同名冲突则 +1 直至成功。
- 校验：
  - 必填：`company_abbr`、`branch`、`channel`、`pay_type`。
  - 字符集：仅允许字母、数字、连字符及空格（防止非法字符同步失败）。
- 存量数据：不改名，仅新建遵循规则。

## 4. 同步与容错

- TTPOS → ERP：创建/更新成功后推送 ERP；若 ERP 失败返回错误并记录审计日志。
- ERP → TTPOS：ERP 侧新增/更新触发同步回 TTPOS；名称沿用 ERP 生成或现有规则。
- 并发：序号生成使用唯一索引或锁；失败重试有限次数后返回重复错误码。
- 监控：记录同步失败原因、原始入参、生成名称。

## 5. 数据模型与存储

- 复用现有支付方式表（ERP/TTPOS），新增字段无需变更；如需存储渠道和序号，使用扩展字段（待与现有表结构核对）。
- 保证 `(channel, pay_type, company_abbr, seq)` 唯一。

## 6. 测试要点

- 正向：必填四字段 → 名称生成符合格式并成功写入。
- 序号：同组合多次创建得到 `0000`（系统/默认）与 `0001+` 自建；并发无重复。
- 校验：缺少必填/非法字符返回校验错误。
- 同步：TTPOS 创建后 ERP 可见；ERP 更新后 TTPOS 更新名称一致。
- 回归：存量数据查询不受影响。
