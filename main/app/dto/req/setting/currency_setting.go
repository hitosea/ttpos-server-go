package setting

type Currency struct {
	Unit             string `json:"unit"`
	PrintUnit        string `json:"print_unit"`
	UnitPosition     string `json:"unit_position"`
	IsOpen           string `json:"is_open"`
	ViceUnit         string `json:"vice_unit"`
	ViceUnitPosition string `json:"vice_unit_position"`
	UnitRate         string `json:"unit_rate"`
}
