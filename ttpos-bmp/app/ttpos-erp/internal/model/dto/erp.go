package dto

type ErpReq struct {
	DocType string `json:"docType"`
	Name    string `json:"name"`
	Method  string `json:"method"`
}
