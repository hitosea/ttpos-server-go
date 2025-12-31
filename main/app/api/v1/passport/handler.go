package passport

import (
	"runtime"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	objectStorageAdapter "ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"

	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	printerPkg "ttpos-server-go/app/modules/printer/pkg"
	"ttpos-server-go/app/service"
	"ttpos-server-go/config"
)

type Handler struct {
	dbm        *database.DBManager
	captchaSrv service.ICaptchaSrv
	encryptSrv service.IEncryptSrv
}

// GetCaptcha 获取验证码
// @Summary 获取验证码
// @Tags 通用
// @Access json
// @Produce json
// @Success 200 {object} dto.Response
// @Router /passport/captcha [get]
func (h *Handler) GetCaptcha(c *gin.Context) {
	captcha, err := h.captchaSrv.Generate()
	if err != nil {
		logger.Logger.Error("生成验证码失败", zap.Error(err))
		helper.Fail(c, constant.CodeFail, "生成验证码失败")
		return
	}
	helper.Success(c, captcha)
}

// GetServerPublicKey 获取服务端公钥
// @Summary 获取服务端公钥
// @Tags 通用
// @Access json
// @Produce json
// @param client_id query string true "客户端Id"
// @param type query string true "加密类型: jsencrypt"
// @Success 200 {object} dto.Response
// @Router /passport/server_public_key [get]
func (h *Handler) GetServerPublicKey(c *gin.Context) {
	var getKeyReq req.GetServerPublicKeyRequest
	if err := c.ShouldBindQuery(&getKeyReq); err != nil {
		helper.HandleValidationError(c, err, getKeyReq, req.GetServerPublicKeyRequestMessage)
		return
	}
	resp, err := h.encryptSrv.GetServerPublicKey(getKeyReq.ClientId, getKeyReq.Type)
	if err != nil {
		helper.Fail(c, constant.CodeFail, "获取服务端公钥失败")
		return
	}
	helper.Success(c, resp)
}

// LianLianCallback 连连支付回调
func (h *Handler) LianLianCallback(c *gin.Context) {
	ctx := helper.GetContext(c)
	sign := c.GetHeader("sign")
	//
	var callbackReq req.LianLianCallbackRequest
	if err := c.ShouldBind(&callbackReq); err != nil {
		helper.HandleValidationError(c, err, callbackReq, req.GetServerPublicKeyRequestMessage)
		return
	}
	// 处理支付回调
	err := service.NewPaymentRepo(ctx, h.dbm).HandleCallback(sign, callbackReq)
	if err != nil {
		helper.Fail(c, constant.CodeFail, err.Error())
		return
	}

	c.String(200, "success")
}

// LianLianRefundCallback 连连支付退款回调
func (h *Handler) LianLianRefundCallback(c *gin.Context) {
	ctx := helper.GetContext(c)
	sign := c.GetHeader("sign")
	//
	var callbackReq req.LianLianRefundCallbackRequest
	if err := c.ShouldBind(&callbackReq); err != nil {
		helper.HandleValidationError(c, err, callbackReq, req.GetServerPublicKeyRequestMessage)
		return
	}
	callbackReq.MerchantRefundId = c.Query("merchant_refund_id")

	// 处理支付回调
	err := service.NewPaymentRepo(ctx, h.dbm).HandleRefundCallback(sign, callbackReq)
	if err != nil {
		helper.Fail(c, constant.CodeFail, err.Error())
		return
	}

	c.String(200, "success")
}

// GetMemoryStats 获取内存统计（内部接口）
func (h *Handler) GetMemoryStats(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 构建内存统计响应数据
	memStats := map[string]interface{}{
		"goroutines":       runtime.NumGoroutine(), // 协程数量
		"alloc_kb":         m.Alloc / 1024,         // 系统内存分配
		"total_alloc_kb":   m.TotalAlloc / 1024,    // 总分配内存
		"sys_kb":           m.Sys / 1024,           // 系统内存
		"num_gc":           m.NumGC,                // 垃圾回收次数
		"last_gc_ns":       m.LastGC,               // 上次GC时间
		"next_gc_kb":       m.NextGC / 1024,        // 下次GC目标
		"heap_alloc_kb":    m.HeapAlloc / 1024,     // 堆内存大小
		"heap_sys_kb":      m.HeapSys / 1024,       // 堆内存系统
		"heap_idle_kb":     m.HeapIdle / 1024,      // 堆内存空闲
		"heap_inuse_kb":    m.HeapInuse / 1024,     // 堆内存使用
		"heap_released_kb": m.HeapReleased / 1024,  // 堆内存释放
		"heap_objects":     m.HeapObjects,          // 堆内存对象
		"stack_inuse_kb":   m.StackInuse / 1024,    // 栈内存
		"stack_sys_kb":     m.StackSys / 1024,      // 栈内存系统
		"gc_cpu_fraction":  m.GCCPUFraction,        // GC CPU使用率
		"enable_gc":        m.EnableGC,             // 是否启用GC
		"debug_gc":         m.DebugGC,              // 是否调试GC
		"pause_total_ns":   m.PauseTotalNs,         // 总暂停时间
		"num_forced_gc":    m.NumForcedGC,          // 强制GC次数
		"gc_cycles":        m.NumGC,                // GC周期（使用NumGC代替）
		"other_sys_kb":     m.OtherSys / 1024,      // 其他系统内存
		"gcsys_kb":         m.GCSys / 1024,         // GC系统内存
		"mcache_inuse_kb":  m.MCacheInuse / 1024,   // MCache使用
		"mcache_sys_kb":    m.MCacheSys / 1024,     // MCache系统
		"mspan_inuse_kb":   m.MSpanInuse / 1024,    // MSpan使用
		"mspan_sys_kb":     m.MSpanSys / 1024,      // MSpan系统
		"buck_hash_sys_kb": m.BuckHashSys / 1024,   // 哈希表系统
		"lookups":          m.Lookups,              // 查找次数
		"mallocs":          m.Mallocs,              // 分配次数
		"frees":            m.Frees,                // 释放次数
		"imgFontSemaphore": printerPkg.GetImgFontStatus(),
	}

	helper.Success(c, memStats)
}

func RegisterHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	wrapper := &Handler{
		dbm:        dbm,
		captchaSrv: service.NewCaptchaSrv(cache),
		encryptSrv: service.NewEncryptSrv(cache),
	}
	router.GET("/captcha", wrapper.GetCaptcha)
	router.GET("/server_public_key", wrapper.GetServerPublicKey)
	router.POST("/lianlian/callback", wrapper.LianLianCallback)
	router.POST("/lianlian/refund/callback", wrapper.LianLianRefundCallback)
}

func (h *Handler) ClearL1Cache(c *gin.Context) {
	// 使用 CommitSHA 作为身份验证，确保只有同一个构建版本才能调用
	apiKey := c.GetHeader("X-API-KEY")
	if apiKey != config.CommitSHA && config.Server.Mode != "debug" {
		helper.Fail(c, constant.CodeAccessDenied, "Unauthorized: Invalid API Key")
		c.Abort()
		return
	}

	clearedCount := objectStorageAdapter.ClearAllL1Cache()

	helper.Success(c, map[string]interface{}{
		"cleared_count": clearedCount,
		"message":       "L1 缓存已清空",
	})
}

// RegisterInternalHandlers 注册内部接口（需要API Key验证）
func RegisterInternalHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	wrapper := &Handler{
		dbm:        dbm,
		captchaSrv: service.NewCaptchaSrv(cache),
		encryptSrv: service.NewEncryptSrv(cache),
	}
	router.GET("/memory/stats", wrapper.GetMemoryStats)
	router.POST("/cache/l1/clear", wrapper.ClearL1Cache)
}
