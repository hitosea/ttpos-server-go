# ttpos 与 LINE MAN 集成用例图

本文档提供了 ttpos 系统与 LINE MAN 平台集成用例的 PlantUML 图示。图示从 ttpos 系统视角和 LINE MAN 系统视角展示了主要参与者、用例及其关系。
在线渲染UML网站:https://www.planttext.com

## 1. ttpos 系统视角用例图 (ttpos System Perspective Use Case Diagram)

此图重点展示 `ttpos 系统` 作为中心，它所执行或直接参与的用例，以及与 `商家操作员` 和 `LINE MAN 系统` 的交互。

```plantuml
@startuml
left to right direction

actor "商家操作员" as Merchant
actor "LINE MAN 系统" as LINEMAN

rectangle "ttpos 系统" {
  usecase "获取访问令牌" as UC_TTPOS_AUTH
  usecase "推送完整菜单快照至 LINE MAN" as UC_TTPOS_MENU_SYNC
  usecase "更新单个菜单项状态至 LINE MAN" as UC_TTPOS_MENU_STATUS_UPDATE
  usecase "接收并保存新订单" as UC_TTPOS_RECEIVE_ORDER
  usecase "更新订单状态" as UC_TTPOS_UPDATE_ORDER_STATUS
  usecase "更新订单详情" as UC_TTPOS_UPDATE_ORDER_DETAIL
  usecase "控制 LINE MAN 门店营业状态" as UC_TTPOS_RESTAURANT_CONTROL

  Merchant -- UC_TTPOS_MENU_SYNC
  Merchant -- UC_TTPOS_MENU_STATUS_UPDATE
  Merchant -- UC_TTPOS_RESTAURANT_CONTROL

  LINEMAN -- UC_TTPOS_AUTH
  LINEMAN -- UC_TTPOS_RECEIVE_ORDER
  LINEMAN -- UC_TTPOS_UPDATE_ORDER_STATUS
  LINEMAN -- UC_TTPOS_UPDATE_ORDER_DETAIL
  LINEMAN -- UC_TTPOS_MENU_SYNC : <<triggered by LINE MAN>>
}
@enduml
```

## 2. LINE MAN 系统视角用例图 (LINE MAN System Perspective Use Case Diagram)

此图重点展示 `LINE MAN 系统` 作为中心，它所执行或直接参与的用例，以及与 `ttpos 系统` 的交互。

```plantuml
@startuml
left to right direction

actor "ttpos 系统" as TTPOS

rectangle "LINE MAN 系统" {
  usecase "获取令牌并验证 ttpos API 请求" as UC_LINEMAN_AUTH
  usecase "IP 白名单验证" as UC_IP_WHITELIST <<extend>>

  usecase "接收并处理菜单" as UC_LINEMAN_PROCESS_MENU
  usecase "通知菜单同步结果" as UC_LINEMAN_MENU_NOTIFY
  usecase "请求 ttpos 重新同步菜单" as UC_LINEMAN_TRIGGER_SYNC
  usecase "更新菜单项状态" as UC_LINEMAN_UPDATE_MENU_STATUS

  usecase "推送新订单详情至 ttpos" as UC_LINEMAN_ORDER_PLACE
  usecase "推送订单状态更新通知至 ttpos" as UC_LINEMAN_ORDER_STATUS
  usecase "推送订单编辑更新通知至 ttpos" as UC_LINEMAN_ORDER_EDIT

  usecase "更新门店营业状态" as UC_LINEMAN_UPDATE_RESTAURANT_STATUS

  TTPOS -- UC_LINEMAN_AUTH
  UC_LINEMAN_AUTH <.. UC_IP_WHITELIST

  TTPOS -- UC_LINEMAN_PROCESS_MENU
  UC_LINEMAN_PROCESS_MENU ..> UC_LINEMAN_MENU_NOTIFY
  TTPOS -- UC_LINEMAN_UPDATE_MENU_STATUS

  TTPOS -- UC_LINEMAN_TRIGGER_SYNC : <<response to ttpos>>

  UC_LINEMAN_ORDER_PLACE -- TTPOS
  UC_LINEMAN_ORDER_STATUS -- TTPOS
  UC_LINEMAN_ORDER_EDIT -- TTPOS

  TTPOS -- UC_LINEMAN_UPDATE_RESTAURANT_STATUS
}
@enduml
```
