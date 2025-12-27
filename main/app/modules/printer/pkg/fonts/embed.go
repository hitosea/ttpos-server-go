package fonts

import (
	"embed"
)

//go:embed en/NotoSans-Regular.ttf
//go:embed ja/NotoSansJP-Regular.ttf
//go:embed ko/NotoSansKR-Regular.ttf
//go:embed my/NotoSansMyanmar-Regular.ttf
//go:embed my/Zawgyi-One_v4.ttf
//go:embed th/NotoSansThai-Regular.ttf
//go:embed th/NotoSerifThai_ExtraCondensed-Regular.ttf
//go:embed tr/NotoSans-Regular.ttf
//go:embed zh/NotoSansSC-Regular.ttf
//go:embed lo/NotoSerifLao-Regular.ttf
var FontFS embed.FS

// 获取字体数据
func GetFontData(fontPath string) ([]byte, error) {
	return FontFS.ReadFile(fontPath)
}

// 字体路径常量
const (
	FontEN    = "en/NotoSans-Regular.ttf"
	FontJA    = "ja/NotoSansJP-Regular.ttf"
	FontKO    = "ko/NotoSansKR-Regular.ttf"
	FontMY    = "my/NotoSansMyanmar-Regular.ttf"
	FontMY2   = "my/Zawgyi-One_v4.ttf"
	FontTH    = "th/NotoSansThai-Regular.ttf"
	FontTHExt = "th/NotoSerifThai_ExtraCondensed-Regular.ttf"
	FontTR    = "tr/NotoSans-Regular.ttf"
	FontZH    = "zh/NotoSansSC-Regular.ttf"
	FontLO    = "lo/NotoSerifLao-Regular.ttf"
)
