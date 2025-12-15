# MQ Topic 清单（ttpos-bmp）

> 本目录按 **模块** 维护 `ttpos-bmp` 中使用到的 MQ Topic 说明，便于运维初始化、联调排查与跨服务协作。

## 目录

- `ttpos-erp.md`：ERP 模块 topic（回调异步、POS 发票异步、商品同步等）
- `ttpos-message.md`：消息中心模块 topic（异步发送邮件/短信等）
- `ttpos-takeout.md`：外卖模块 topic（Grab 菜单更新、门店集成状态、订单事件等）
- `ttpos-websocket.md`：WebSocket 模块 topic（LAN 打印机上报等，当前以文档约定为主）
- `common.md`：通用/基础设施 topic（如连通性探测）

## 约定

- **topic 命名**：业务 topic 优先使用 `kebab-case` 或 `snake_case`，且保持稳定。
- **消息体序列化**：`ttpos-bmp/internal/pkg/queue` 通过 `gjson.EncodeString` 序列化；建议结构体字段补齐 `json` tag，避免字段名大小写差异导致跨语言解析困难。
- **生产/消费定位**：文档中给出“生产位置/消费位置”的源码路径，便于快速跳转。
