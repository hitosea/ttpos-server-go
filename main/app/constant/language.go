package constant

type LocaleType string

const (
	LocaleZH   LocaleType = "zh"
	LocaleTH   LocaleType = "th"
	LocaleEN   LocaleType = "en"
	LocaleZHTW LocaleType = "zhtw"
	LocaleJA   LocaleType = "ja"
	LocaleKO   LocaleType = "ko"
	LocaleMY   LocaleType = "my"
	LocaleTR   LocaleType = "tr"
	LocaleSV   LocaleType = "sv"
)

type Locales []LocaleType

var LocaleList = Locales{LocaleZH, LocaleTH, LocaleEN, LocaleZHTW, LocaleJA, LocaleKO, LocaleMY, LocaleTR}

func (l Locales) GetLocaleType(language string) LocaleType {
	switch language {
	case "zh":
		return LocaleZH
	case "th":
		return LocaleTH
	case "en":
		return LocaleEN
	case "zhtw":
		return LocaleZHTW
	case "ja":
		return LocaleJA
	case "ko":
		return LocaleKO
	case "my":
		return LocaleMY
	case "tr":
		return LocaleTR
	}
	return LocaleZHTW
}
