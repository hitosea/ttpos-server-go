package skootar

type CreateOrderInp struct {
	ReqBase
	Vehicle         string     `json:"vehicle"`         // 车型，固定: Motorcycle
	JobType         string     `json:"jobType"`         // 订单类型，固定 3 Food
	JobDate         string     `json:"jobDate"`         // 订单日期: 格式 YYYY-MM-DD
	StartTime       string     `json:"startTime"`       // 订单开始时间: 格式 HH:MM 或 now
	LocationList    []Location `json:"locationList"`    // 定位，包含商家和会员地址
	PaymentType     string     `json:"paymentType"`     // 支付类型，固定 prepaid 预付费
	MerchantConfirm int        `json:"merchantConfirm"` // 商家确认，固定 1
	CallbackUrl     string     `json:"callbackUrl"`     // 回调地址，skootar 通知外送服务的 URL
	Option          string     `json:"option"`          // 可选参数，固定 10 Food box
	Remark          string     `json:"remark,omitempty"`
	RefNo           string     `json:"refNo,omitempty"`
	PromCode        string     `json:"promCode,omitempty"`
}

type CreateOrderOut struct {
	JobDetail    JobDetail `json:"jobDetail"`
	ResponseCode string    `json:"responseCode"`
	ResponseDesc string    `json:"responseDesc"`
}

type LocationList struct {
	Seq          int    `json:"seq"`
	Type         string `json:"type"`
	AddressID    int    `json:"addressId"`
	AddressName  string `json:"addressName"`
	Address      string `json:"address"`
	Lat          string `json:"lat"`
	Lng          string `json:"lng"`
	ContactName  string `json:"contactName"`
	ContactPhone string `json:"contactPhone"`
}

//
//type JobDetail struct {
//	JobID         string         `json:"jobId"`
//	JobDate       string         `json:"jobDate"`
//	JobStatus     int            `json:"jobStatus"`
//	JobStatusEn   string         `json:"jobStatusEn"`
//	JobStatusTh   string         `json:"jobStatusTh"`
//	JobDesc       string         `json:"jobDesc"`
//	StartTime     string         `json:"startTime"`
//	FinishTime    string         `json:"finishTime"`
//	HaveReturn    bool           `json:"haveReturn"`
//	JobType       string         `json:"jobType"`
//	Option        string         `json:"option"`
//	TotalDistance float64        `json:"totalDistance"`
//	TotalWeight   float64        `json:"totalWeight"`
//	TotalSize     float64        `json:"totalSize"`
//	Remark        string         `json:"remark"`
//	UserType      int            `json:"userType"`
//	NormalPrice   float64        `json:"normalPrice"`
//	NetPrice      float64        `json:"netPrice"`
//	Discount      float64        `json:"discount"`
//	Rating        int            `json:"rating"`
//	LocationList  []LocationList `json:"locationList"`
//}
