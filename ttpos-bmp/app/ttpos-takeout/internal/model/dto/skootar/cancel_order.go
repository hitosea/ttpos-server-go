package skootar

type CancelOrderInp struct {
	ReqBase
	JobId        string `json:"jobId"`
	CancelReason string `json:"cancelReason"`
}

type CancelOrderOut struct {
	ResponseCode string `json:"responseCode"`
	ResponseDesc string `json:"responseDesc"`
}
