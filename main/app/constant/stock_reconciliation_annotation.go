package constant

import "ttpos-server-go/app/dto"

const (
	// StockReconciliationAnnotationTypeResubmit 批注类型-重新发起
	StockReconciliationAnnotationTypeResubmit = 1
	// StockReconciliationAnnotationTypeReject 批注类型-驳回
	StockReconciliationAnnotationTypeReject = 2
	// StockReconciliationAnnotationTypeApprove 批注类型-通过
	StockReconciliationAnnotationTypeApprove = 3
)

// StockReconciliationAnnotationTypeLocaleNameMap 批注类型多语言名称映射
var StockReconciliationAnnotationTypeLocaleNameMap = map[int]dto.LocaleResponse{
	StockReconciliationAnnotationTypeResubmit: {ZH: "重新发起", EN: "Resubmit", TH: "ส่งใหม่", ZHTW: "重新發起", JA: "再提出", KO: "재제출", MY: "ပြန်လည်တင်သွင်းရန်", TR: "Yeniden Gönder", SV: "Skicka igen"},
	StockReconciliationAnnotationTypeReject:   {ZH: "门店驳回", EN: "Store Rejected", TH: "ร้านค้าปฏิเสธ", ZHTW: "門店駁回", JA: "店舗却下", KO: "매장 거부", MY: "ဆိုင်ငြင်းပယ်", TR: "Mağaza Reddetti", SV: "Butik avvisad"},
	StockReconciliationAnnotationTypeApprove:  {ZH: "门店通过", EN: "Store Approved", TH: "ร้านค้าอนุมัติ", ZHTW: "門店通過", JA: "店舗承認", KO: "매장 승인", MY: "ဆိုင်အတည်ပြု", TR: "Mağaza Onayladı", SV: "Butik godkänd"},
}

// GetStockReconciliationAnnotationTypeLocaleName 获取批注类型多语言名称
func GetStockReconciliationAnnotationTypeLocaleName(annotationType int) dto.LocaleResponse {
	if name, ok := StockReconciliationAnnotationTypeLocaleNameMap[annotationType]; ok {
		return name
	}
	return dto.LocaleResponse{}
}
