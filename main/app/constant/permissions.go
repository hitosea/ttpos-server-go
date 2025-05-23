package constant

// 收银机权限
var CashierPermissions = map[string]string{
	"/api/v1/cashier/desk/open":  "cashier_table_open",
	"/api/v1/cashier/desk/close": "cashier_table_delete",
}
