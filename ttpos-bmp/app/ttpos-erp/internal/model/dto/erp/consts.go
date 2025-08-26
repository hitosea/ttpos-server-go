package erp

const (
	ApiMethodMakeMappedDoc = "frappe.model.mapper.make_mapped_doc"
	//ApiSaveCancel 取消已提交记录
	ApiSaveCancel = "frappe.desk.form.save.cancel"
	//ApiIsDocumentAmend 记录是否已修订
	ApiIsDocumentAmend = "frappe.client.is_document_amended"
)

// 文档类型
const (
	DocTypePosProfile      = "POS Profile"
	DocTypePosInvoice      = "POS Invoice"
	DocTypePosOpeningEntry = "POS Opening Entry"
	DocTypePosClosingEntry = "POS Closing Entry"
	DocTypeModeOfPayment   = "Mode of Payment"
)
