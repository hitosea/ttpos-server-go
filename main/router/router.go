package router

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"ttpos-server-go/app/api/v1/admin"
	"ttpos-server-go/app/api/v1/assistant"
	"ttpos-server-go/app/api/v1/callboard"
	"ttpos-server-go/app/api/v1/cashier"
	"ttpos-server-go/app/api/v1/h5"
	"ttpos-server-go/app/api/v1/kitchen"
	"ttpos-server-go/app/api/v1/member"
	"ttpos-server-go/app/api/v1/menu"
	"ttpos-server-go/app/api/v1/passport"
	"ttpos-server-go/app/api/v1/shop"
	"ttpos-server-go/app/api/v1/tablet"
	_ "ttpos-server-go/app/event" // 注册事件
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/rpc"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, dbm *database.DBManager, cache cache.Cache) {
	// 判活接口
	r.GET("api/health", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})
	r.GET("api/v1/testrpc", middleware.Internal(), func(c *gin.Context) {
		rpc.TestCompanyList()
		rpc.TestEstimateDistance()
		c.String(http.StatusOK, "Success")
	})
	r.GET("api/add_multi_language_name_uuid", func(c *gin.Context) {
		uuids := strings.Split(c.Query("uuids"), ",")
		var uuidUint64s []uint64
		for _, uuid := range uuids {
			uuidUint64, _ := strconv.ParseUint(strings.TrimSpace(uuid), 10, 64)
			if uuidUint64 > 0 && !slices.Contains(uuidUint64s, uuidUint64) {
				uuidUint64s = append(uuidUint64s, uuidUint64)
			}
		}
		companyUuid, _ := strconv.ParseUint(c.Query("company_uuid"), 10, 64)
		if companyUuid > 0 && len(uuidUint64s) > 0 && dbm.GetDB(companyUuid) != nil {
			translateSrv := service.NewTranslateSrv(dbm, cache)
			translateSrv.AddMultiLanguageNameUuidToSet(companyUuid, uuidUint64s...)
		} else {
			c.String(http.StatusBadRequest, "参数错误")
		}
		c.String(http.StatusOK, "ok")
	})

	apiV1 := r.Group("api/v1")
	{
		adminGroup := apiV1.Group("/admin")
		{
			admin.RegisterHandlers(adminGroup, dbm, cache)
		}

		// 通用接口
		passportGroup := apiV1.Group("/passport")
		{
			passport.RegisterHandlers(passportGroup, dbm, cache)
		}

		// 内部接口
		internalGroup := apiV1.Group("/internal")
		{
			internalGroup.Use(middleware.Internal())
			passport.RegisterInternalHandlers(internalGroup, dbm, cache)
		}

		// 商家端/移动管理端
		shopGroup := apiV1.Group("/shop")
		{
			shop.RegisterBaseHandlers(shopGroup, dbm, cache)
			shop.RegisterOrderHandlers(shopGroup, dbm, cache)
			shop.RegisterOrderImportHandlers(shopGroup, dbm, cache)
			shop.RegisterRechargeOrderHandlers(shopGroup, dbm, cache)
			shop.RegisterStatisticsHandlers(shopGroup, dbm, cache)
			shop.RegisterMemberOrderHandlers(shopGroup, dbm, cache)
			shop.RegisterAuthHandlers(shopGroup, dbm, cache)         // 认证
			shop.RegisterStaffHandlers(shopGroup, dbm, cache)        // 管理员管理
			shop.RegisterSettingHandlers(shopGroup, dbm, cache)      // 设置
			shop.RegisterProductHandlers(shopGroup, dbm, cache)      // 商品
			shop.RegisterProductLabelHandlers(shopGroup, dbm, cache) // 商品标签
			shop.RegisterMaterialHandlers(shopGroup, dbm, cache)     // 物品管理
			shop.RegisterMiscHandlers(shopGroup, dbm, cache)         // 杂项
			shop.RegisterPurchaseHandlers(shopGroup, dbm, cache)     // 采购
			shop.RegisterSupplierHandlers(shopGroup, dbm, cache)     // 供应商
			shop.RegisterCallBoardHandlers(shopGroup, dbm, cache)    // 叫号展示
			shop.RegisterWarehouseHandlers(shopGroup, dbm, cache)    // 仓库管理
			shop.RegisterPrintHandlers(shopGroup, dbm, cache)        // 打印管理

			shop.RegisterStockReconciliationHandlers(shopGroup, dbm, cache) // 盘点

			shop.RegisterBatchProductHandlers(shopGroup, dbm, cache) // 分批商品

			shop.RegisterTransferOrderHandlers(shopGroup, dbm, cache) // 调拨单

			shop.RegisterExportRecordHandlers(shopGroup, dbm, cache) // 导出记录
		}
		// 收银端
		cashierGroup := apiV1.Group("/cashier")
		{
			cashier.RegisterAuthHandlers(cashierGroup, dbm, cache)
			cashier.RegisterProductHandlers(cashierGroup, dbm, cache)
			cashier.RegisterDeskHandlers(cashierGroup, dbm, cache)
			cashier.RegisterBuffetHandlers(cashierGroup, dbm, cache)
			cashier.RegisterInstantHandlers(cashierGroup, dbm, cache)
			cashier.RegisterOrderHandlers(cashierGroup, dbm, cache)
			cashier.RegisterOrderOldHandler(cashierGroup, dbm, cache)
			cashier.RegisterMemberHandlers(cashierGroup, dbm, cache)
			cashier.RegisterSoldOutHandlers(cashierGroup, dbm, cache)
			cashier.RegisterBaseHandlers(cashierGroup, dbm, cache)
			cashier.RegisterCallHandlers(cashierGroup, dbm, cache)
			cashier.RegisterH5OrderHandlers(cashierGroup, dbm, cache)
			cashier.RegisterMemberOrderHandlers(cashierGroup, dbm, cache)
			cashier.RegisterMemberOrderManageHandlers(cashierGroup, dbm, cache)
			cashier.RegisterRechargeOrderHandlers(cashierGroup, dbm, cache)
			cashier.RegisterPrinterHandlers(cashierGroup, dbm, cache)
			cashier.RegisterStatisticsHandlers(cashierGroup, dbm, cache)
		}
		// 点餐助手端
		assistantGroup := apiV1.Group("/assistant")
		{
			assistant.RegisterDeskHandlers(assistantGroup, dbm, cache)
			assistant.RegisterProductHandlers(assistantGroup, dbm, cache)
			assistant.RegisterBuffetHandlers(assistantGroup, dbm, cache)
			assistant.RegisterAuthHandlers(assistantGroup, dbm, cache)
			assistant.RegisterBaseHandlers(assistantGroup, dbm, cache)
			assistant.RegisterMemberHandlers(assistantGroup, dbm, cache)
			assistant.RegisterCallHandlers(assistantGroup, dbm, cache)
			assistant.RegisterOrderHandlers(assistantGroup, dbm, cache)
		}
		// H5扫码端
		h5Group := apiV1.Group("/h5")
		{
			h5.RegisterH5Handlers(h5Group, dbm, cache)
		}
		// 电子菜单菜单
		menuGroup := apiV1.Group("/menu")
		{
			menu.RegisterMenuHandlers(menuGroup, dbm, cache)
		}
		// 厨房端
		kitchenGroup := apiV1.Group("/kitchen")
		{
			kitchen.RegisterAuthHandlers(kitchenGroup, dbm, cache)
			kitchen.RegisterBaseHandlers(kitchenGroup, dbm, cache)
			kitchen.RegisterCallHandlers(kitchenGroup, dbm, cache)
			kitchen.RegisterProductHandlers(kitchenGroup, dbm, cache)
		}
		// 平板端
		tabletGroup := apiV1.Group("/tablet")
		{
			tablet.RegisterAuthHandlers(tabletGroup, dbm, cache)
			tablet.RegisterBaseHandlers(tabletGroup, dbm, cache)
			tablet.RegisterDeskHandlers(tabletGroup, dbm, cache)
			tablet.RegisterCallHandlers(tabletGroup, dbm, cache)
			tablet.RegisterProductHandlers(tabletGroup, dbm, cache)
			tablet.RegisterBuffetHandlers(tabletGroup, dbm, cache)
		}
		// 会员端
		memberGroup := apiV1.Group("/member")
		{
			member.RegisterAuthHandlers(memberGroup, dbm, cache)
			member.RegisterProductHandlers(memberGroup, dbm, cache)
			member.RegisterOrderHandlers(memberGroup, dbm, cache)
			member.RegisterMarketingHandlers(memberGroup, dbm, cache)
			member.RegisterAddressHandlers(memberGroup, dbm, cache)
			member.RegisterBaseHandlers(memberGroup, dbm, cache)
			member.RegisterBenefitHandlers(memberGroup, dbm, cache)
			member.RegisterMemberCallbackHandlers(memberGroup, dbm, cache)
		}

		// 叫号展示设备端
		callBoardGroup := apiV1.Group("/callboard")
		{
			callboard.RegisterHandlers(callBoardGroup, dbm, cache)
		}
	}
}
