package utils

import (
	"encoding/json"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
)

type Names struct {
	Zh   string `json:"zh"`
	En   string `json:"en"`
	Th   string `json:"th"`
	Ja   string `json:"ja"`
	Ko   string `json:"ko"`
	My   string `json:"my"`
	Tr   string `json:"tr"`
	ZhTw string `json:"zhtw"`
}

func (n *Names) GetNames(jsonString string) error {
	err := json.Unmarshal([]byte(jsonString), &n)
	if err != nil {
		return errors.WithMessage(err, "解析json失败")
	}
	return nil
}

func (n *Names) CreateMultiLanguageName(multiLanguageNameUuid uint64) *model.MultiLanguageName {
	multiLanguageName := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid: multiLanguageNameUuid,
		},
		EnName:   n.En,
		ZhName:   n.Zh,
		ZhTwName: n.ZhTw,
		ThName:   n.Th,
		MyName:   n.My,
		JaName:   n.Ja,
		KoName:   n.Ko,
		TrName:   n.Tr,
	}
	return &multiLanguageName
}
