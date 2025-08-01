package consts

// ProviderName 外送供应商名称
type ProviderName string

const (
	ProviderSkootar ProviderName = "skootar"
	ProviderGrab    ProviderName = "grab"
)

// TTPOS_HEADER_CALLBACK_AUTH TTPOS回调Auth
const TTPOS_HEADER_CALLBACK_AUTH = "X-TTPOS-Callback-Auth"
