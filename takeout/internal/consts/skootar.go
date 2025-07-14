package consts

type CashFee string

const (
	Yes CashFee = "Y"
	No  CashFee = "N"
)

type JobType string

const (
	JobTypeDocument    JobType = "1"
	JobTypeParcelGoods JobType = "2"
	JobTypeFood        JobType = "3"
)

type EstimateOption string

/**
Extra optional for delivery

1 is Delivery Document, Collect cheque, Deliver invoice
2 is Deposit cheque after collecting (+40 baht)
3 is Return trip (+50% max 200 baht)
4 is Collect cash on delivery and return immediatly (+50% max 200 baht)
6 is Send post
*/
const (
	EstimateOptionDocument    EstimateOption = "1"
	EstimateOptionDeposit     EstimateOption = "2"
	EstimateOptionReturn      EstimateOption = "3"
	EstimateOptionCollectCash EstimateOption = "4"
	EstimateOptionSendPost    EstimateOption = "6"
)
