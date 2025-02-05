package setting

type Printer struct {
	CashierOpen        string         `json:"cashier_open"`
	CashierPrinterID   string         `json:"cashier_printer_id"`
	CashierPrinter     []interface{}  `json:"cashier_printer"`
	LanguageList       []LanguageItem `json:"language_list"`
	LanguageMethod     string         `json:"language_method"`
	DefaultLanguage    string         `json:"default_language"`
	PrintMethod        int            `json:"print_method"`
	KitchenLanguage    string         `json:"kitchen_language"`
	KitchenPrintMethod int            `json:"kitchen_print_method"`
	ConsumptionTax     string         `json:"consumption_tax"`
	BuffetSignOpen     string         `json:"buffet_sign_open"`
	MonetaryUnitOpen   string         `json:"monetary_unit_open"`
	CalendarList       []CalendarItem `json:"calendar_list"`
	PrintList          []PrintItem    `json:"print_list"`
	DefaultCalendar    string         `json:"default_calendar"`
}

type CalendarItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type PrintItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}
