payment_type = frappe.form_dict.payment_type
company_abbr = frappe.form_dict.company_abbr

company = frappe.get_last_doc("Company", filters={"abbr": ("like", company_abbr)})
payment = frappe.get_doc('Mode of Payment',payment_type)
payment.append('accounts',{
    'company': company.name ,
    "default_account": "Cash - "+ company_abbr,
})
#默认用现金账号
payment.save()
frappe.response['message'] = "ok"