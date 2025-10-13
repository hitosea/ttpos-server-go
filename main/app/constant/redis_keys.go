package constant

import "fmt"

const (
	RedisKeyProductPrinterListV2 = "PRODUCT_PRINTER_LIST_v2:%d:%d"
)

func GetRedisKeyProductPrinterListV2(companyUuid uint64, printMode int) string {
	return fmt.Sprintf(RedisKeyProductPrinterListV2, companyUuid, printMode)
}
