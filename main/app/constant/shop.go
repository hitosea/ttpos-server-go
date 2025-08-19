package constant

type LanguageItem struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var Languages = []LanguageItem{
	{Name: "th", Value: "ภาษาไทย"},
	{Name: "zh", Value: "简体中文"},
	{Name: "zhtw", Value: "繁體中文"},
	{Name: "en", Value: "English"},
	{Name: "ja", Value: "日本語"},
	{Name: "ko", Value: "한국어"},
	{Name: "my", Value: "မြန်မာဘာသာ"},
	{Name: "tr", Value: "Türkçe"},
	{Name: "sv", Value: "Svenska"},
}

const (
	CheckNameSourceUnit           = "unit"
	CheckNameSourceProduct        = "product"
	CheckNameSourceCategory       = "category"
	CheckNameSourceSauce          = "sauce"
	CheckNameSourceAttribute      = "attribute"
	CheckNameSourceAttributeGroup = "attribute_group"
	CheckNameSourceFlavor         = "flavor"
)
