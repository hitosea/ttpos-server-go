package skootar

import "ttpos-bmp/app/ttpos-takeout/internal/consts"

type EstimatePriceInp struct {
	ReqBase
	JobType      consts.JobType        `json:"jobType"`
	LocationList []Location            `json:"locationList"`
	Option       consts.EstimateOption `json:"option,omitempty"` //可以多选 如 1,2,4,6
	PromoCode    string                `json:"promoCode,omitempty" `
}

type EstimatePriceOut struct {
	NormalPrice  float32 //Normal price
	NetPrice     float32 //Net price after discount
	Discount     float32 //Discount from promo code
	Distance     float32 //Distance in km.
	ResponseCode string
	ResponseDesc string
	TripDuration int32
}
