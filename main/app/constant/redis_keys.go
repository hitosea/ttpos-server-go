package constant

import "fmt"

const (
	RedisKeyProductPrinterListV2 = "PRODUCT_PRINTER_LIST_v2:%d:%d"
	// RedisKeySyncTask 同步任务锁的 Redis Key，格式: sync:task:{companyUuid}
	RedisKeySyncTask = "sync:task:%d"
)

func GetRedisKeyProductPrinterListV2(companyUuid uint64, printMode int) string {
	return fmt.Sprintf(RedisKeyProductPrinterListV2, companyUuid, printMode)
}

func GetRedisKeySyncTask(companyUuid uint64) string {
	return fmt.Sprintf(RedisKeySyncTask, companyUuid)
}
