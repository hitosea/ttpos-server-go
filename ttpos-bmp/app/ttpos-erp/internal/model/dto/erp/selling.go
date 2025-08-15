package erp

type PosProfile struct {
	Name      string `json:"name,omitempty"`
	Company   string `json:"company,omitempty"`
	Warehouse string `json:"warehouse,omitempty"`
	Branch    string `json:"branch,omitempty"`
}
