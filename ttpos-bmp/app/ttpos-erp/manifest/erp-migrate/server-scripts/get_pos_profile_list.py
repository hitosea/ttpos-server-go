# 查询Pos Profile列表，只返回名称

company_abbr = frappe.form_dict.company_abbr

company = frappe.get_last_doc("Company", filters={"abbr": ("like", company_abbr)})


profiles = frappe.get_list("POS Profile", fields=["name", "company"], filters={"company": ("like", company.name)})
# 返回结果
frappe.response['data'] = profiles
