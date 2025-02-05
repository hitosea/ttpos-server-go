package req

// 收银机解锁
type ReqCashierUnlock struct {
	Password string `json:"password"`
}
