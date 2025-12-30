package cache

import (
	"context"
	"fmt"
	"time"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

const (
	// PrinterTemplateTmpDataCacheKeyPrefix 打印机模板 TmpData 缓存键前缀
	PrinterTemplateTmpDataCacheKeyPrefix = "printer_template_data:%d:%d" // company_uuid:template_id
	// PrinterTemplateTmpDataCacheTTL 缓存过期时间（1小时）
	PrinterTemplateTmpDataCacheTTL = 1 * time.Hour
)

// PrinterTemplateTmpDataCache 打印机模板 TmpData 缓存工具
type PrinterTemplateTmpDataCache struct {
	cache cache.Cache
}

// NewPrinterTemplateTmpDataCache 创建缓存工具实例
func NewPrinterTemplateTmpDataCache(cache cache.Cache) *PrinterTemplateTmpDataCache {
	return &PrinterTemplateTmpDataCache{
		cache: cache,
	}
}

// Get 获取缓存的 TmpData
func (c *PrinterTemplateTmpDataCache) Get(companyUuid, templateID uint64) (string, bool) {
	if c.cache == nil {
		return "", false
	}

	cacheKey := c.getCacheKey(companyUuid, templateID)
	if cacheData, ok := c.cache.GetBytes(cacheKey); ok {
		return string(cacheData), true
	}

	return "", false
}

// Set 设置缓存
func (c *PrinterTemplateTmpDataCache) Set(companyUuid, templateID uint64, tmpData string) {
	if c.cache == nil || tmpData == "" {
		return
	}

	cacheKey := c.getCacheKey(companyUuid, templateID)
	if err := c.cache.Set(cacheKey, []byte(tmpData), PrinterTemplateTmpDataCacheTTL); err != nil {
		logger.Logger.Error("设置打印机模板数据缓存失败",
			zap.Uint64("template_id", templateID),
			zap.Uint64("company_uuid", companyUuid),
			zap.Error(err))
		return
	}

	logger.Logger.Debug("已缓存打印机模板数据",
		zap.Uint64("template_id", templateID),
		zap.Uint64("company_uuid", companyUuid),
		zap.Int("data_size", len(tmpData)))
}

// Delete 删除缓存
func (c *PrinterTemplateTmpDataCache) Delete(companyUuid, templateID uint64) {
	if c.cache == nil {
		return
	}

	cacheKey := c.getCacheKey(companyUuid, templateID)
	c.cache.Del(cacheKey)

	logger.Logger.Info("已清除打印机模板数据缓存",
		zap.Uint64("template_id", templateID),
		zap.Uint64("company_uuid", companyUuid))
}

// getCacheKey 获取缓存键
func (c *PrinterTemplateTmpDataCache) getCacheKey(companyUuid, templateID uint64) string {
	return fmt.Sprintf(PrinterTemplateTmpDataCacheKeyPrefix, companyUuid, templateID)
}

// InvalidateByContext 通过 context 清除缓存（用于更新操作）
func (c *PrinterTemplateTmpDataCache) InvalidateByContext(ctx context.Context, templateID uint64) {
	// 从 context 中获取 company_uuid（如果有的话）
	if companyUuid, ok := ctx.Value("company_uuid").(uint64); ok && companyUuid > 0 {
		c.Delete(companyUuid, templateID)
	}
}
