package constant

// 呼叫类型(1服务员,2结账)
const (
	CallTypeWaiter   = 1 // 服务员
	CallTypeCheckout = 2 // 结账
)

// 呼叫状态
const (
	CallStatusUnprocessed = 0 // 未处理
	CallStatusProcessed   = 1 // 已处理
)
