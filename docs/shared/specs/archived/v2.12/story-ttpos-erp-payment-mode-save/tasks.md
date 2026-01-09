# story-ttpos-erp-payment-mode-save / SaveModeOfPayment 任务清单

## 开发
- [ ] 更新 `selling.proto`：新增 `SaveModeOfPaymentReq/Resp` 与 `rpc SaveModeOfPayment`，字段含 `company_abbr/branch/channel/pay_type`。
- [ ] 生成/提交 proto 代码（goframe/golang）并确保编译通过。
- [ ] 服务端实现：名称规则 `{channel}-{pay_type}-{NNNN} - {company_abbr}`，序号按 `channel+pay_type` 递增，`0000` 预留系统，用户自建从 `0001` 起，冲突顺延。
- [ ] 校验：必填字段、字符集限制；缺失/非法返回错误码。
- [ ] 并发与唯一性：序号生成加唯一约束或锁，失败重试并返回重复错误。
- [ ] 同步：TTPOS → ERP 创建/更新；ERP → TTPOS 变更回流；失败记录审计日志。
- [ ] 数据模型检查：确认存储表字段（渠道/序号/名称）及唯一键 `(channel, pay_type, company_abbr, seq)`。

## 测试
- [ ] 正向：四必填字段生成名称并落库/同步成功。
- [ ] 序号：同组合多次创建得到 `0000`（系统默认）与 `0001+` 自建；并发无重复。
- [ ] 校验：缺失必填/非法字符返回校验错误，不写入。
- [ ] 同步：TTPOS 创建后 ERP 可见；ERP 更新后 TTPOS 名称一致。
- [ ] 兼容：存量数据查询/同步不受影响。

## 文档与交付
- [ ] 更新接口文档/README，说明命名规则与错误码。
- [ ] 提交测试报告与示例请求/响应。
