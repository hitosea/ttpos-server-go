package constant

// 收银机权限
var CashierPermissions = map[string]string{
	"/api/v1/cashier/desk/open":  "cashier_table_open",
	"/api/v1/cashier/desk/close": "cashier_table_delete",
}

// 点餐助手权限
var AssistantPermissions = map[string]string{
	"/api/v1/assistant/desk/close":        "clear_table",
	"/api/v1/assistant/desk/order/cancel": "clear_table",
}
