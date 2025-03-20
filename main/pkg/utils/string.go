package utils

import (
	"fmt"
	"math/rand"
	"time"
)

/**
* 生成订单号
* @return string
 */
func GenerateMerchantOrderNo(prefix string) string {
	datePart := time.Now().Format("20060102150405")
	randomPart := fmt.Sprintf("%08d", rand.Intn(100000000))
	return prefix + datePart + randomPart
}
