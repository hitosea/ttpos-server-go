package dto

type ConfirmOrderInp struct {
	JobId string
}

type CancelOrderInp struct {
	JobId  string
	Reason string
}

type GetDriverInfoInp struct {
	SKootarId string
	Name      string
	Phone     string
	Avatar    string
	Rating    float64
}
