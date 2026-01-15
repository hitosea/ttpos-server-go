package websocket

import (
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

// 源类型
const (
	SourceAll       = "*"         // 所有
	SourceShop      = "shop"      // 商家
	SourceCashier   = "cashier"   // 收银机
	SourceTablet    = "tablet"    // 平板端
	SourceKitchen   = "kitchen"   // 厨显端
	SourceAssistant = "assistant" // 点餐助手
	SourceH5        = "H5"        // H5
)

// 消息类型
const (
	// 更新订单，刷新购物车和桌台列表都用它（desk_uuid不等于0代表是桌台订单，desk_uuid等于0代表是点餐订单），data = {"update_time": 1742971471,"sale_bill_uuid": 3655262269341697,"desk_uuid": 3655262269341699}
	UPDATE_ORDER = "update_order"
	// 客户呼叫 data = {"update_time": 1742971471,"customer_call_uuid": 3655262269341697,"desk_uuid": 3655262269341699}
	CUSTOMER_CALL = "customer_call"
	// 打印数据 data = {"update_time": 1742971471,"print_log_uuid": 3655262269341697}
	PRINT_DATA = "print_data"
	// H5订单 data = {"update_time": 1742971471,"h5_order_uuid": 3655262269341697, "desk_uuid": 3655262269341699}
	H5_ORDER = "h5_order"
	// 更新配置 （所有后台配置相关变动） data = {"update_time": 1742971471}
	UPDATE_CONFIG = "update_config"
	// 更新权限 （编辑角色的时候） data = {"update_time": 1742971471}
	UPDATE_PERMISSION = "update_permission"
	// 更新用户 （用户名称、头像等信息）（也可能会切换角色，所以也要更新权限） data = {"update_time": 1742971471, "staff_uuid": 3655262269341697}
	UPDATE_USER = "update_user"
	// 更新商品 data = {"update_time": 1742971471, "product_uuid": 1, "type": "update | delete"}
	UPDATE_PRODUCT = "update_product"
	// 更新分类 data = {"update_time": 1742971471, "category_uuid": 1, "type": "update | delete"}
	UPDATE_CATEGORY = "update_category"
	// 更新自助餐 data = {"update_time": 1742971471, "buffet_uuid": 1, "type": "update | delete"}
	UPDATE_BUFFET = "update_buffet"
	// 更新桌台 data = {"update_time": 1742971471, "desk_uuid": 1, "type": "update | delete"}
	UPDATE_DESK = "update_desk"
	// 更新桌台类型 data = {"update_time": 1742971471, "type_uuid": 1, "type": "update | delete"}
	UPDATE_DESK_TYPE = "update_desk_type"
	// 更新退款状态 data = {"update_time": 1742971471, "uuid": 1, "type": "update | delete"}
	UPDATE_REFUND_STATE = "update_refund_state"
	// 更新厨显 data = {"update_time": 1742971471}
	UPDATE_KITCHEN = "update_kitchen"
	// 更新打印机 data = {"update_time": 1742971471, "printer_uuid": 1, "type": "update | delete"}
	UPDATE_SELECTED_PRINTER = "update_selected_printer"
	// 更新会员订单 data = {"update_time": 1742971471, "status": 1, "member_sale_order_uuid": 1, "type": "update | delete"}
	UPDATE_MEMBER_SALE_ORDER = "update_member_sale_order"
	// 移动管理端获取最新数据 data= {"sync_time": 1742971471}
	SYNC_DATA = "sync_data"
	// 导入商品 data= {"time": 1742971471, "status": "finish", "error": ""}
	IMPORT_PRODUCT = "import_product"
	// 导入物品 data= {"time": 1742971471, "status": "finish", "error": ""}
	IMPORT_MATERIAL = "import_material"
	// 更新外卖订单 data = {"update_time": 1742971471, "order_uuid": 1, "takeout_order_uuid": "xxx", "platform": "grab", "type": "create | update | delete"}
	UPDATE_TAKEOUT_ORDER = "update_takeout_order"
)

// Push sends a POST request to the WebSocket server with specific parameters.
func PushClient(companyUuid uint64, sourceClient, deviceId, messageType string, data map[string]any) error {
	utils.Go(func() {
		err := PushClients(companyUuid, sourceClient, deviceId, messageType, data)
		if err != nil {
			logger.Logger.Error("推送消息失败", zap.Error(err))
		}
	})
	return nil
}
