package websocket

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"ttpos-server-go/config"
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
)

// Push sends a POST request to the WebSocket server with specific parameters.
func PushClient(company_uuid uint64, source_client, device_id, message_type string, data map[string]interface{}) error {
	var jsonData []byte
	// 计算包含关键参数的MD5值
	if message_type == UPDATE_ORDER {
		jsonData = fmt.Appendf(nil, "%d", data["sale_bill_uuid"])
	}

	// 创建缓存键
	var cacheKey string
	// 更新桌台的时候
	if message_type == UPDATE_DESK {
		key := fmt.Sprintf("%d:%s:%s:%s", company_uuid, source_client, device_id, message_type)
		md5Sum := fmt.Sprintf("%x", md5.Sum([]byte(key)))
		cacheKey = fmt.Sprintf("ws_msg:%s", md5Sum)
	} else if message_type == PRINT_DATA {
		key := fmt.Sprintf("%d:%s:%s:%s", company_uuid, source_client, device_id, message_type)
		var updateTime string
		// 安全地获取更新时间，支持多种可能的键名和类型
		if updateTimeVal, exists := data["update_times"]; exists {
			switch v := updateTimeVal.(type) {
			case int64:
				updateTime = strconv.FormatInt(v, 10)
			case int:
				updateTime = strconv.FormatInt(int64(v), 10)
			case string:
				updateTime = v
			default:
				updateTime = fmt.Sprintf("%v", v)
			}
		}
		md5Sum := fmt.Sprintf("%x", md5.Sum(append([]byte(key), []byte(updateTime)...)))
		cacheKey = fmt.Sprintf("ws_msg:%s", md5Sum)
	} else {
		key := fmt.Sprintf("%d:%s:%s:%s", company_uuid, source_client, device_id, message_type)
		md5Sum := fmt.Sprintf("%x", md5.Sum(append([]byte(key), jsonData...)))
		cacheKey = fmt.Sprintf("ws_msg:%s", md5Sum)
	}

	// 判断当前是否在容器内执行
	url := fmt.Sprintf("http://127.0.0.1:%s/ws/push", os.Getenv("NGINX_PORT"))
	if _, err := os.Stat("/.dockerenv"); err == nil {
		url = "http://nginx/ws/push"
	}

	// 构建请求体
	payload := map[string]interface{}{
		"company_uuid":  company_uuid,
		"source_client": source_client,
		"device_id":     device_id,
		"message_type":  message_type,
		"message_key":   cacheKey,
		"data":          data,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Failed to marshal payload: %v", err)
		log.Printf("Failed to marshal payload: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Printf("Failed to create request: %v", err)
		log.Printf("Failed to create request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", config.JWT.Secret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v", err)
		log.Printf("Failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Received non-OK response: %s %s", resp.Status, url)
		log.Printf("Received non-OK response: %s %s", resp.Status, url)
		return fmt.Errorf("received non-OK response: %s", resp.Status)
	}

	return nil
}
