package consts

// 支付方式创建来源
const (
	// PaymentAddedBySystem 系统创建的支付方式标识
	// 使用此标识时，支付方式序号固定为 0000
	PaymentAddedBySystem = "sys"
)

// 支付方式序号
const (
	// PaymentSeqSystem 系统支付方式固定序号
	// 系统创建的支付方式使用此序号，避免与用户创建的支付方式冲突
	PaymentSeqSystem = 0
)
