package event

// init 初始化 takeout 模块事件处理器（基础设施层）
func init() {
	RegisterOrderCancelledEventHandler()
}
