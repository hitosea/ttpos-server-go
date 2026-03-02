// Package consts 定义常量
package consts

// 源类型常量
const (
	SourceAll       = "*"         // 所有来源
	SourceShop      = "shop"      // 商家端
	SourceCashier   = "cashier"   // 收银机端
	SourceTablet    = "tablet"    // 平板端
	SourceKitchen   = "kitchen"   // 厨显端
	SourceAssistant = "assistant" // 点餐助手
	SourceH5        = "H5"        // H5端
)

// 消息类型常量
const (
	// 更新订单
	// 刷新购物车和桌台列表都用它（desk_uuid不等于0代表是桌台订单，desk_uuid等于0代表是点餐订单）
	// data = {"update_time": 1742971471,"sale_bill_uuid": 3655262269341697,"desk_uuid": 3655262269341699}
	MessageTypeUpdateOrder = "update_order"

	// 客户呼叫
	// data = {"update_time": 1742971471,"customer_call_uuid": 3655262269341697,"desk_uuid": 3655262269341699}
	MessageTypeCustomerCall = "customer_call"

	// 打印数据
	// data = {"update_time": 1742971471,"print_log_uuid": 3655262269341697}
	MessageTypePrintData = "print_data"

	// H5订单
	// data = {"update_time": 1742971471,"h5_order_uuid": 3655262269341697, "desk_uuid": 3655262269341699}
	MessageTypeH5Order = "h5_order"

	// 更新配置
	// 所有后台配置相关变动
	// data = {"update_time": 1742971471}
	MessageTypeUpdateConfig = "update_config"

	// 更新权限
	// 编辑角色的时候
	// data = {"update_time": 1742971471}
	MessageTypeUpdatePermission = "update_permission"

	// 更新用户
	// 用户名称、头像等信息（也可能会切换角色，所以也要更新权限）
	// data = {"update_time": 1742971471, "staff_uuid": 3655262269341697}
	MessageTypeUpdateUser = "update_user"

	// 更新商品
	// data = {"update_time": 1742971471, "product_uuid": 1, "type": "update | delete"}
	MessageTypeUpdateProduct = "update_product"

	// 更新分类
	// data = {"update_time": 1742971471, "category_uuid": 1, "type": "update | delete"}
	MessageTypeUpdateCategory = "update_category"

	// 更新自助餐
	// data = {"update_time": 1742971471, "buffet_uuid": 1, "type": "update | delete"}
	MessageTypeUpdateBuffet = "update_buffet"

	// 更新桌台
	// data = {"update_time": 1742971471, "desk_uuid": 1, "type": "update | delete"}
	MessageTypeUpdateDesk = "update_desk"

	// 更新桌台类型
	// data = {"update_time": 1742971471, "type_uuid": 1, "type": "update | delete"}
	MessageTypeUpdateDeskType = "update_desk_type"

	// 更新退款状态
	// data = {"update_time": 1742971471, "uuid": 1, "type": "update | delete"}
	MessageTypeUpdateRefundState = "update_refund_state"

	// 更新厨显
	// data = {"update_time": 1742971471}
	MessageTypeUpdateKitchen = "update_kitchen"

	// 更新打印机
	// data = {"update_time": 1742971471, "printer_uuid": 1, "type": "update | delete"}
	MessageTypeUpdateSelectedPrinter = "update_selected_printer"

	// 更新会员订单
	// data = {"update_time": 1742971471, "status": 1, "member_sale_order_uuid": 1, "type": "update | delete"}
	MessageTypeUpdateMemberSaleOrder = "update_member_sale_order"

	// 移动管理端获取最新数据
	// data= {"sync_time": 1742971471}
	MessageTypeSyncData = "sync_data"

	// 导入商品
	// data= {"time": 1742971471, "status": "finish", "error": ""}
	MessageTypeImportProduct = "import_product"

	// 导入物品
	// data= {"time": 1742971471, "status": "finish", "error": ""}
	MessageTypeImportMaterial = "import_material"
)

// WebSocket 连接状态代码
const (
	CodeSuccess = 200 // 成功
	CodeFail    = 500 // 失败
)

// WebSocket 客户端消息类型
const (
	ClientMessageTypeHeartbeat      = "heartbeat"        // 心跳消息
	ClientMessageTypeReply          = "reply"            // 已读回复
	ClientMessageTypeUsbPrintReport = "usb_print_report" // USB打印机上报
	ClientMessageTypeLanPrintReport = "lan_print_report" // LAN打印机上报
)

// websocket_msg_push 频道
const (
	ChannelWebsocketMsgPush = "websocket_msg_push" // 推送消息频道
)

// Redis Key
const (
	RedisKeyDeviceHeartbeat = "device:heartbeat" // 设备心跳缓存 (HASH: connection_key -> timestamp)
)
