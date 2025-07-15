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
	JobTypeFood        JobType = "3" //外送服务默认
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
	EstimateOptionFood        EstimateOption = "10" //外送服务默认
)

type PaymentType string

const (
	PaymentTypeCash       PaymentType = "cash"
	PaymentTypeInvoice    PaymentType = "invoice"
	PaymentTypeCreditcard PaymentType = "creditcard"
	PaymentTypePrepaid    PaymentType = "prepaid" //外送服务默认
)

type Vehicle string

const (
	VehicleMotorcycle Vehicle = "Motorcycle" //外送目前只有一种
)

// 默认送餐为当前时间
const DEFAULT_START_TIME = "now"

type JobStatus int

const (
	JobStatusCanceled       JobStatus = iota
	JobStatusNew            JobStatus = iota
	JobStatusUnderReview    JobStatus = iota
	JobStatusMissionCreated JobStatus = iota
	JobStatusMatching       JobStatus = iota
	JobStatusAssigned       JobStatus = iota
	JobStatusOTW_Pickup     JobStatus = iota
	JobStatusOTW_Delivery   JobStatus = iota
	JobStatusPending        JobStatus = iota
	JobStatusCompleted      JobStatus = iota
	JobStatusInvoiceCreated JobStatus = iota
	JobStatusPaid           JobStatus = iota
)
