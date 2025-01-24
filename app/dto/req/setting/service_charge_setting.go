package setting

type ServiceCharge struct {
	IsOpen            string `json:"is_open"`
	ChargeType        string `json:"charge_type"`
	ServiceCharge     string `json:"service_charge"`
	ServiceChargeRate string `json:"service_charge_rate"`
	IsOpenTax         string `json:"is_open_tax"`
}
