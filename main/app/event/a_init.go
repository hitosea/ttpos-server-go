package event

// init 初始化事件
func init() {
	// 自动注册"接单"事件处理器
	acceptH5OrderEventHandler()
	// 自动注册"拒单"事件处理器
	rejectH5OrderEventHandler()
}
