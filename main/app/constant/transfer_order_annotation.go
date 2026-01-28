package constant

import "ttpos-server-go/app/dto"

const (
	// TransferOrderAnnotationTypeResubmit 批注类型-重新提交
	TransferOrderAnnotationTypeResubmit = 1
	// TransferOrderAnnotationTypeShopApprove 批注类型-门店通过
	TransferOrderAnnotationTypeShopApprove = 2
	// TransferOrderAnnotationTypeShopReject 批注类型-门店驳回
	TransferOrderAnnotationTypeShopReject = 3
	// TransferOrderAnnotationTypeParentApprove 批注类型-上级门店通过
	TransferOrderAnnotationTypeParentApprove = 4
	// TransferOrderAnnotationTypeParentReject 批注类型-上级门店驳回
	TransferOrderAnnotationTypeParentReject = 5
	// TransferOrderAnnotationTypeShipperApprove 批注类型-发货门店通过（调入）
	TransferOrderAnnotationTypeShipperApprove = 6
	// TransferOrderAnnotationTypeShipperReject 批注类型-发货门店驳回（调入）
	TransferOrderAnnotationTypeShipperReject = 7
	// TransferOrderAnnotationTypeReceiverApprove 批注类型-收货门店通过（调出）
	TransferOrderAnnotationTypeReceiverApprove = 8
	// TransferOrderAnnotationTypeReceiverReject 批注类型-收货门店驳回（调出）
	TransferOrderAnnotationTypeReceiverReject = 9
)

// TransferOrderAnnotationTypeLocaleNameMap 批注类型多语言名称映射
var TransferOrderAnnotationTypeLocaleNameMap = map[int]dto.LocaleResponse{
	TransferOrderAnnotationTypeResubmit:        {ZH: "重新提交", EN: "Resubmit", TH: "ส่งใหม่", ZHTW: "重新提交", JA: "再提出", KO: "재제출", MY: "ပြန်လည်တင်သွင်းရန်", TR: "Yeniden Gönder", SV: "Skicka igen"},
	TransferOrderAnnotationTypeShopApprove:     {ZH: "门店通过", EN: "Store Approved", TH: "ร้านค้าอนุมัติ", ZHTW: "門店通過", JA: "店舗承認", KO: "매장 승인", MY: "ဆိုင်အတည်ပြု", TR: "Mağaza Onayladı", SV: "Butik godkänd"},
	TransferOrderAnnotationTypeShopReject:      {ZH: "门店驳回", EN: "Store Rejected", TH: "ร้านค้าปฏิเสธ", ZHTW: "門店駁回", JA: "店舗却下", KO: "매장 거부", MY: "ဆိုင်ငြင်းပယ်", TR: "Mağaza Reddetti", SV: "Butik avvisad"},
	TransferOrderAnnotationTypeParentApprove:   {ZH: "上级门店通过", EN: "Parent Store Approved", TH: "ร้านค้าระดับบนอนุมัติ", ZHTW: "上級門店通過", JA: "上位店舗承認", KO: "상위 매장 승인", MY: "မိခင်ဆိုင်အတည်ပြု", TR: "Üst Mağaza Onayladı", SV: "Överordnad butik godkänd"},
	TransferOrderAnnotationTypeParentReject:    {ZH: "上级门店驳回", EN: "Parent Store Rejected", TH: "ร้านค้าระดับบนปฏิเสธ", ZHTW: "上級門店駁回", JA: "上位店舗却下", KO: "상위 매장 거부", MY: "မိခင်ဆိုင်ငြင်းပယ်", TR: "Üst Mağaza Reddetti", SV: "Överordnad butik avvisad"},
	TransferOrderAnnotationTypeShipperApprove:  {ZH: "发货门店通过", EN: "Shipper Approved", TH: "ร้านค้าผู้ส่งอนุมัติ", ZHTW: "發貨門店通過", JA: "発送店舗承認", KO: "발송 매장 승인", MY: "ပို့ဆောင်သူဆိုင်အတည်ပြု", TR: "Gönderici Onayladı", SV: "Avsändare godkänd"},
	TransferOrderAnnotationTypeShipperReject:   {ZH: "发货门店驳回", EN: "Shipper Rejected", TH: "ร้านค้าผู้ส่งปฏิเสธ", ZHTW: "發貨門店駁回", JA: "発送店舗却下", KO: "발송 매장 거부", MY: "ပို့ဆောင်သူဆိုင်ငြင်းပယ်", TR: "Gönderici Reddetti", SV: "Avsändare avvisad"},
	TransferOrderAnnotationTypeReceiverApprove: {ZH: "收货门店通过", EN: "Receiver Approved", TH: "ร้านค้าผู้รับอนุมัติ", ZHTW: "收貨門店通過", JA: "受取店舗承認", KO: "수령 매장 승인", MY: "လက်ခံသူဆိုင်အတည်ပြု", TR: "Alıcı Onayladı", SV: "Mottagare godkänd"},
	TransferOrderAnnotationTypeReceiverReject:  {ZH: "收货门店驳回", EN: "Receiver Rejected", TH: "ร้านค้าผู้รับปฏิเสธ", ZHTW: "收貨門店駁回", JA: "受取店舗却下", KO: "수령 매장 거부", MY: "လက်ခံသူဆိုင်ငြင်းပယ်", TR: "Alıcı Reddetti", SV: "Mottagare avvisad"},
}

// GetTransferOrderAnnotationTypeLocaleName 获取批注类型多语言名称
func GetTransferOrderAnnotationTypeLocaleName(annotationType int) dto.LocaleResponse {
	if name, ok := TransferOrderAnnotationTypeLocaleNameMap[annotationType]; ok {
		return name
	}
	return dto.LocaleResponse{}
}
