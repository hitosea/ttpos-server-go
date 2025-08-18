def setup_company_default_warehouse(doc):
    warehouse = frappe.get_doc({
            "doctype": "Warehouse",
            "warehouse_name": "All Warehouses",
            "company": doc.name,
            "is_group": 1
        })
    warehouse.custom_aliasname = 'Default'
    warehouse.save()

setup_company_default_warehouse(doc)