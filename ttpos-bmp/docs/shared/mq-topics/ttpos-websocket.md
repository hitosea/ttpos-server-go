# ttpos-websocket MQ Topic

## 总览

- **模块**：`app/ttpos-websocket`
- **说明**：当前在该模块代码中未发现 `queue.Push/Subscribe` 的实际落地；但在 `ttpos-bmp/manifest/topics.txt` 与架构文档中存在 topic 约定。

## Topic 清单

### 1) `lan-printer-report`

- **用途（文档约定）**：LAN 打印机信息上报，供其他服务/消费者订阅获取门店 LAN 打印机的在线/变更情况。
- **来源**：`docs/human/architecture/modules/websocket/features/lan-printer-report.md`
- **代码状态**：
  - 在 `app/ttpos-websocket` 内暂未检索到该 topic 的生产/消费代码。
  - WebSocket 客户端消息类型当前存在 `lan_print_report`（见 `app/ttpos-websocket/internal/consts/consts.go`），与 MQ topic 命名不同。

> 建议：若要真正通过 MQ 广播 LAN 打印机上报，需要明确由哪个服务生产 `lan-printer-report`，并补齐消息体 schema 与订阅者列表。
