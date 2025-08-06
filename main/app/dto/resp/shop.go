package resp

type CheckNameItem struct {
	Lang      string `json:"lang"`
	TextExist bool   `json:"text_exist"`
}

type CheckNameResp struct {
	List []CheckNameItem `json:"list"`
}
