package skootar

type ConfirmOrderInp struct {
	ReqBase
	JobId string `json:"jobId"`
}

type ConfirmOrderOut struct {
	ResponseCode string `json:"responseCode"`
	ResponseDesc string `json:"responseDesc"`
}
