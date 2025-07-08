package utils

import (
	"fmt"
	"strings"
)

// GenerateProductPackageSign 生成商品包签名。格式：商品规格uuid:属性uuid1,属性uuid2,属性uuid3:加料uuid1,加料uuid2,加料uuid3
func GenerateProductPackageSign(flavorUuid uint64, attributeUuidList, sauceUuidList []string) string {
	return fmt.Sprintf("%d:%s:%s", flavorUuid, strings.Join(attributeUuidList, ","), strings.Join(sauceUuidList, ","))
}
