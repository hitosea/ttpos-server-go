package resp

type TakeoutDistanceResp struct {
	Distance     float64 // 预估距离。单位km
	TripDuration int     // 预估时间，单位秒
}

type CreateTakeoutOrderResp struct {
	TakeoutJobUuid string // 外送服务订单
	Status         string // 订单状态
	ShopOrderUuid  string // 商户订单
	TakeoutRefNo   string // 外送渠道订单，比如skootar订单
	FinishTime     string // 预计送达时间
}

type GetDriverInfoResp struct {
	Name   string  // 骑手姓名
	Phone  string  // 骑手电话
	Avatar string  // 骑手头像
	Rating float64 // 骑手评分
	Lat    float64 // 骑手纬度
	Lng    float64 // 骑手经度
}

// TakeoutMenuExportResp 外卖菜单导出响应
type TakeoutMenuExportResp struct {
	Platform string      `json:"platform"` // 平台名称
	MenuData interface{} `json:"menuData"` // 平台格式的菜单数据
}

// GrabProductImportFailure 失败明细
type GrabProductImportFailure struct {
	GrabProductId string `json:"grabProductId"`
	Message       string `json:"message"`
}

// ProductImportError 商品导入错误详情
type ProductImportError struct {
	ProductID   string `json:"product_id"`   // 商品ID
	ProductName string `json:"product_name"` // 商品名称
	CategoryID  string `json:"category_id"`  // 分类ID
	Error       string `json:"error"`        // 错误信息
}

// GrabMenuImportResp Grab 商品导入响应
type GrabMenuImportResp struct {
	SuccessCount int                        `json:"successCount"`
	FailureCount int                        `json:"failureCount"`
	Failures     []GrabProductImportFailure `json:"failures"`
	ErrorList    []ProductImportError       `json:"error_list"` // 详细错误列表
}
