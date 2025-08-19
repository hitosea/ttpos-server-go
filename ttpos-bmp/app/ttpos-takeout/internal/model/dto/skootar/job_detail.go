package skootar

type JobDetailInp struct {
	JobId string `json:"jobId"`
}

// JobDetailReq  skootar job 接口请求
type JobDetailReq struct {
	ReqBase
	JobId string `json:"jobId"`
}

// JobDetail skootar job 接口响应
type JobDetail struct {
	JobId           string         `json:"jobId"`
	JobDate         string         `json:"jobDate"`
	JobStatus       int            `json:"jobStatus"`
	JobStatusEn     string         `json:"jobStatusEn"`
	JobStatusTh     string         `json:"jobStatusTh"`
	JobDesc         string         `json:"jobDesc"`
	StartTime       string         `json:"startTime"`
	FinishTime      string         `json:"finishTime"`
	JobType         string         `json:"jobType"`
	Option          string         `json:"option"`
	TotalDistance   float32        `json:"totalDistance"` // not support in this version
	TotalWeight     float32        `json:"totalWeight"`   // not support in this version
	TotalSize       float32        `json:"totalSize"`     // not support in this version
	Remark          string         `json:"remark"`
	UserType        int            `json:"userType"`
	NormalPrice     float32        `json:"normalPrice"`
	NetPrice        float32        `json:"netPrice"`
	Discount        float32        `json:"discount"`
	Rating          int            `json:"rating"`
	LocationList    []LocationList `json:"locationList"`
	HaveReturn      bool           `json:"haveReturn"` // not support in this version
	SkootarId       string         `json:"skootarId"`
	SkootarName     string         `json:"skootarName"`
	SkootarPhone    string         `json:"skootarPhone"`
	SkootarImageUrl string         `json:"skootarImageUrl"`
	SkootarRating   float32        `json:"skootarRating"`
	TrackingUrl     string         `json:"trackingUrl"`
}

type JobDetailResp struct {
	RespBase
	JobDetail JobDetail `json:"jobDetail"`
}
