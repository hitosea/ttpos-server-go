# LINE MAN 对接梳理摘要

## 概览
- 文档版本：Partner Integration Workflow V2。
- 主体内容包含认证、菜单、订单、门店四大模块和十个 API。
- 目标：建立 LINE MAN 与 ttpos 之间的菜单同步与订单互通能力。

## 数据方向

**LINE MAN 推送给 ttpos**
- `place order notification`（必选）：推送新订单详情。
- `status update notification`（可选）：推送订单完成或取消状态。
- `order update notification`（可选）：推送订单编辑后的变更。
- `trigger sync menu`（必选）：请求 ttpos 重新同步菜单。
- `menu sync notification`（可选）：反馈菜单同步成功或失败。

**ttpos 提供给 LINE MAN**
- `authentication`：通过 OAuth2 Client Credentials 申请访问令牌。
- `menu sync`（必选）：推送完整菜单快照。
- `update menu item status`（可选）：同步单个菜品的可售状态。
- `update menu propertyValue status`（可选）：同步菜单选项状态。
- `force close/open restaurant`（可选）：控制 LINE MAN 门店营业状态。

## 交互模式
- **认证方式**：OAuth2 Client Credentials；可辅以 IP 白名单。
- **调用方向**：双方均暴露 HTTPS API，携带 access token 进行 REST 调用。
- **通知机制**：所有 “*notification*” 接口均为 LINE MAN 主动回调 ttpos。
- **触发机制**：收到 `trigger sync menu` 后，ttpos 需立即调用 `menu sync` 推送最新菜单。

## 核心流程说明

1. **认证流程**
   - ttpos 使用 Client ID/Secret 申请 access token。
   - LINE MAN 验证后返回 token，双方后续调用需携带。

2. **菜单同步流程**
   - ttpos 监控内部菜单变化，调用 `menu sync` 推送完整快照。
   - （可选）LINE MAN 完成处理后，通过 `menu sync notification` 回传结果。

3. **触发同步流程**
   - LINE MAN 监测到菜单异常时，调用 `trigger sync menu` 请求 ttpos 重新推送。
   - ttpos 再次执行菜单同步流程，确保数据一致。

4. **菜单状态更新流程（可选）**
   - ttpos 针对菜品或选项的销售状态更改调用对应接口。
   - 状态可选：`AVAILABLE`、`SOLD_OUT_TODAY`、`SUSPENDED`。

5. **订单通知流程**
   - LINE MAN 在接单时通过 `place order notification` 推送完整信息。
   - （可选）状态变化或订单编辑事件触发对应通知接口。
   - 平板仍作为兜底接单终端。

6. **门店控制流程（可选）**
   - ttpos 在 POS 端发起开店/关店请求，调用 `force close/open restaurant`。
   - LINE MAN 更新门店营业状态并生效。

以上流程覆盖文档中所有必选与可选场景，可作为后续系统设计与接口联调的基础。


