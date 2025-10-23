package template_json

import (
	"embed"
)

//go:embed 结账单_adv_config.json
//go:embed 结账单_config.json
//go:embed 结账单_data.json
//go:embed 结账单_tmp.json
//go:embed 预结账单_adv_config.json
//go:embed 预结账单_config.json
//go:embed 预结账单_data.json
//go:embed 预结账单_tmp.json
var TemplateJsonFS embed.FS

// 获取模板JSON数据
func GetTemplateJsonData(templateJsonPath string) ([]byte, error) {
	return TemplateJsonFS.ReadFile(templateJsonPath)
}
