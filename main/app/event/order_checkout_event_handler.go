package event

import (
	"fmt"
	"sync"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/printer/printer_model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var once_checkout_sale_order_event_handler sync.Once

// checkoutSaleOrderEventHandler "结账"事件处理器
func checkoutSaleOrderEventHandler() {
	once_checkout_sale_order_event_handler.Do(func() {

		// 创建菜单-付款打印
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			saleOrder := payload.SaleBill.GetSaleOrder(payload.SaleOrderUuid)
			if saleOrder == nil {
				return
			}
			products := printer_model.Products{}
			for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
				if saleOrderProduct.IsCancelProduct() || !saleOrderProduct.IsSendKitchen() {
					continue
				}
				products = append(products, printer_model.OrderProduct{
					OrderProductId:        saleOrderProduct.Uuid,
					BatchTagUuid:          saleOrderProduct.BatchTagUuid,
					ProductId:             saleOrderProduct.ProductPackageUuid,
					ProductName:           saleOrderProduct.MultiLanguageName.GetNames(),
					ProductType:           saleOrderProduct.ProductType,
					ProductAttr:           saleOrderProduct.GetAttributeName(),
					ProductAttrList:       saleOrderProduct.GetAttributeNameList(),
					Attr:                  saleOrderProduct.GetPureAttributeName(),
					AttrList:              saleOrderProduct.GetPureAttributeNameList(),
					FlavorName:            saleOrderProduct.GetFlavorName(),
					ProductSauceNamesList: saleOrderProduct.GetSauceNamesList(),
					TotalNum:              saleOrderProduct.Num,
					NumType:               saleOrderProduct.NumType,
					IsBuffet:              saleOrderProduct.IsBuffet == 1,
					IsWrap:                saleOrderProduct.IsWrapProduct(),
					IsGift:                saleOrderProduct.IsGiftProduct(),
					Remark:                saleOrderProduct.Remark,
				})
			}
			if len(products) > 0 {
				printer.NewPrinterRepo(payload.Ctx, "").PrintingDishes(
					constant.PrinterProductTypePay,
					payload.SaleBillUuid,
					payload.SaleOrderUuid,
					products,
				)
			}
		})

		// 创建结账单打印
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			_, err := printer.NewPrinterRepo(payload.Ctx).PrintingStatementOrder(
				constant.PrinterTemplateBilling,
				payload.SaleBill,
				payload.SaleOrderUuid,
				0,
				0,
			)
			if err != nil {
				fmt.Println("CheckoutSaleOrderEvent process, PrintingStatementOrder failed ", err)
			}
		})

		// 创建操作记录
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
			record := model.SaleOrderOperationRecord{
				Source:        payload.Source,
				Action:        constant.OrderSettle,
				Remark:        "结账",
				SaleBillUuid:  payload.SaleBillUuid,
				SaleOrderUuid: payload.SaleOrderUuid,
				OperatorUuid:  payload.GetOperatorUuid(),
			}
			payload.IsSplitOrder = payload.SaleBill.IsSplit()
			for i, saleOrder := range payload.SaleBill.SaleOrders {
				if saleOrder.Uuid == payload.SaleOrderUuid {
					payload.Index = i + 1
				}
			}
			record.Data = payload.ToJsonString()
			record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
			uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
			if err != nil {
				logger.Logger.Error("SubscribeCheckoutZeroSaleOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
				return
			}
			logger.Logger.Info(fmt.Sprintf("操作记录:结账 %+v", payload), zap.Uint64("record", uuid))
		})

		// 扣减库存
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			ReduceStock(db, payload.SaleBillUuid)
		})

		// 发放积分
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			if payload.SaleOrderUuid == 0 {
				return
			}
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			saleOrder := payload.SaleBill.GetSaleOrder(payload.SaleOrderUuid)
			// 如果订单有会员且订单的赠送积分大于0，则发放积分
			if saleOrder.ConsumerUuid != 0 && saleOrder.GiftPoints > 0 {
				time.Sleep(time.Second)
				// 加锁, 避免并发问题
				lock.NewSystemLock().LockUuid(saleOrder.ConsumerUuid)
				defer lock.NewSystemLock().UnlockUuid(saleOrder.ConsumerUuid)

				// 获取最新的会员信息
				member, err := repository.NewMemberRepo(db).GetMemberByUuid(saleOrder.ConsumerUuid)
				if err != nil {
					logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, GetMemberRecord failed", zap.Any("payload", payload), zap.Error(err))
					return
				}
				// 创建积分发放记录. // 累计会员的消费金额、消费次数
				saleOrder.HandleMemberPoints(member)
				if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
					// 更新会员积分
					if err := repository.NewMemberRepo(tx).Update(member.Uuid, map[string]any{
						"frozen_point": member.FrozenPoint,
					}); err != nil {
						return errors.WithMessage(err)
					}
					// 创建积分变动记录
					memberPointLog := saleOrder.NewMemberPointLog()
					if _, err := repository.NewMemberPointLogRepo(tx).Create(*memberPointLog); err != nil {
						return errors.WithMessage(err)
					}
					return nil
				}); err != nil {
					logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, Transaction failed", zap.Any("payload", payload), zap.Error(err))
					return
				}

				go func() {
					//  处理"积分变动"事件
					HandleMemberPoints(db)

					// 处理会员升级
					memberSrv := service.NewMemberSrv(database.GetDBManager(config.DatabaseConf{}), cache.Global)
					go memberSrv.HandleMemberUpgrade(payload.CompanyUuid, saleOrder.ConsumerUuid)
				}()
			}
		})

		// 处理会员余额
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			go HandleMemberBalance(db)
		})

		// 处理高峰时段
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			// 如果订单未完成，不处理
			if !payload.SaleBill.IsFinish() {
				return
			}
			// 获取门店设置
			setting := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global)
			storeSetting, err := setting.GetStoreSetting(payload.Ctx)
			if err != nil {
				logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, GetStoreSetting failed", zap.Error(err))
				return
			}
			//
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			err = repository.NewSaleOrderPeakTimeRepo(db).Record("inc", payload.SaleBill, 0.0, storeSetting.TimeZone)
			if err != nil {
				fmt.Println("SubscribeCheckoutSaleOrderEvent process, Record failed", payload, err)
				logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, Record failed", zap.Any("payload", payload), zap.Error(err))
			}
		})

		// 邀请有礼活动-统计获奖
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			HandleActivityConsumption(payload)
		})

		// 增加销量
		// 增加产品销量
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			payload.Ctx.SetDB(db)
			HandleAddSalesVolume(payload)
		})
		// 增加材料销量
		event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
			payload.Ctx.SetDB(db)
			HandleAddMaterialSalesVolume(payload)
		})

		// 结账订单后，统计订单原料用量。 TODO 影响功能
		// event.NewSystemBus().SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
		// 	db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
		// 	payload.Ctx.SetDB(db)

		// 	saleOrder := payload.SaleBill.GetSaleOrder(payload.SaleOrderUuid)
		// 	if saleOrder == nil {
		// 		return
		// 	}
		// 	materialStocks := saleOrder.GetValidSaleOrderProductMaterialList()
		// 	saleOrderMaterials := make([]*model.SaleOrderMaterial, 0)
		// 	for _, materialStock := range materialStocks {
		// 		saleOrderMaterials = append(saleOrderMaterials, &model.SaleOrderMaterial{
		// 			BaseModel: model.BaseModel{
		// 				CreateTime: saleOrder.FinishTime, // 原料使用时间=销售订单完成时间
		// 			},
		// 			SaleOrderUuid:     saleOrder.Uuid,
		// 			SaleBillUuid:      payload.SaleBillUuid,
		// 			MaterialUuid:      materialStock.MaterialUuid,
		// 			Num:               materialStock.StockNum,
		// 			StaffShiftLogUuid: saleOrder.StaffShiftLogUuid,
		// 		})
		// 	}
		// 	if err := repository.NewSaleOrderMaterialRepo(db).BatchInsertSaleOrderMaterial(saleOrderMaterials); err != nil {
		// 		logger.Logger.Error("HandleAddMaterialSalesVolume process, BatchInsertSaleOrderMaterial failed", zap.Any("saleOrderMaterials", saleOrderMaterials), zap.Error(err))
		// 		return
		// 	}
		// })
	})
}

// 增加销量
func HandleAddSalesVolume(payload event.CheckoutSaleOrderPayload) {
	ProductBoms, ProducttPackages := getSalesVolume(payload.SaleBill)

	for productBomUuid, saleNum := range ProductBoms {
		if err := repository.NewProductBomRepo(payload.Ctx.GetDB()).AddActualSaleNum(productBomUuid, saleNum); err != nil {
			logger.Logger.Error("HandleAddSalesVolume process, AddActualSaleNum failed", zap.Any("productBomUuid", productBomUuid), zap.Any("saleNum", saleNum), zap.Error(err))
			continue
		}
	}
	for productPackageUuid, saleNum := range ProducttPackages {
		if err := repository.NewProductPackageRepo(payload.Ctx.GetDB()).AddActualSaleNum(productPackageUuid, saleNum); err != nil {
			logger.Logger.Error("HandleAddSalesVolume process, AddActualSaleNum failed", zap.Any("productPackageUuid", productPackageUuid), zap.Any("saleNum", saleNum), zap.Error(err))
			continue
		}
	}
}

// 增加材料销量
func HandleAddMaterialSalesVolume(payload event.CheckoutSaleOrderPayload) {
	MaterialSalesVolume := getMaterialSalesVolume(payload.CompanyUuid, payload.SaleOrderUuid)
	for materialUuid, saleNum := range MaterialSalesVolume {
		if err := repository.NewMaterialRepo(payload.Ctx.GetDB()).AddActualSaleNum(materialUuid, saleNum); err != nil {
			logger.Logger.Error("HandleAddMaterialSalesVolume process, AddActualSaleNum failed", zap.Any("materialUuid", materialUuid), zap.Any("saleNum", saleNum), zap.Error(err))
			continue
		}
	}
}

// 获取销量
func getSalesVolume(saleBill *model.SaleBill) (map[uint64]float64, map[uint64]float64) {
	ProductBoms := make(map[uint64]float64)      // 规格商品销量 map[规格商品UUID]销量
	ProducttPackages := make(map[uint64]float64) // 套餐商品销量 map[套餐商品UUID]销量
	for _, saleOrder := range saleBill.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			// 删除商品、取消商品、未送厨商品、套餐子商品不增加销量
			if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() || !saleOrderProduct.IsSendKitchen() || saleOrderProduct.IsPackageSubProduct() {
				continue
			}
			ProductBoms[saleOrderProduct.GetFlavorBomUuid()] = decimal.NewFromFloat(ProductBoms[saleOrderProduct.GetFlavorBomUuid()]).Add(decimal.NewFromFloat(saleOrderProduct.Num)).InexactFloat64()           // 增加实际销量
			ProducttPackages[saleOrderProduct.ProductPackageUuid] = decimal.NewFromFloat(ProducttPackages[saleOrderProduct.ProductPackageUuid]).Add(decimal.NewFromFloat(saleOrderProduct.Num)).InexactFloat64() // 增加实际销量
		}
	}
	return ProductBoms, ProducttPackages
}

// 获取材料销量
func getMaterialSalesVolume(companyUuid uint64, saleOrderUuid uint64) map[uint64]float64 {
	db := database.GetDBManager(config.DatabaseConf{}).GetDB(companyUuid)
	// 通过sale_order_uuid查询出库单明细中有效的出库材料,然后统计每个材料的销量
	warehouseOutFormItems, err := repository.NewWarehouseFormRepo(db).GetWarehouseOutFormItemBySaleOrderUuid(saleOrderUuid)
	if err != nil {
		logger.Logger.Error("getMaterialSalesVolume process, GetWarehouseOutFormItemsBySaleOrderUuid failed", zap.Error(err))
		return nil
	}
	MaterialSalesVolume := make(map[uint64]float64) // 材料销量 map[材料UUID]销量
	for _, warehouseOutFormItem := range warehouseOutFormItems {
		MaterialSalesVolume[warehouseOutFormItem.MaterialUuid] = decimal.NewFromFloat(MaterialSalesVolume[warehouseOutFormItem.MaterialUuid]).Add(decimal.NewFromFloat(warehouseOutFormItem.Num)).Round(4).InexactFloat64()
	}
	return MaterialSalesVolume
}

// 处理积分变动
func HandleMemberPoints(db *gorm.DB) {
	// 加锁, 避免并发问题
	lock.NewSystemLock().LockUuid(constant.LockNameMemberPoints)
	defer lock.NewSystemLock().UnlockUuid(constant.LockNameMemberPoints)

	// 查询积分变动
	memberPointLogRepo := repository.NewMemberPointLogRepo(db)
	memberPointLogs, err := memberPointLogRepo.GetMemberPointLogNotProcessed()
	if err != nil {
		logger.Logger.Info("HandleMemberPoints process, GetMemberPointLogNotProcessed failed", zap.Error(err))
		return
	}

	type MemberUuid uint64
	memberChangePoint := make(map[MemberUuid]decimal.Decimal)            // 各个会员的积分变动
	memberChangePointConsumption := make(map[MemberUuid]decimal.Decimal) // 各个会员的消费赠送积分变动
	for _, memberPointLog := range memberPointLogs {
		// 累计同一个会员的积分变动
		pre := memberChangePoint[MemberUuid(memberPointLog.MemberUuid)]
		memberChangePoint[MemberUuid(memberPointLog.MemberUuid)] = pre.Add(decimal.NewFromFloat(memberPointLog.Value))
		// 累计同一个会员的消费赠送积分变动
		if memberPointLog.Scene == constant.MemberPointLogSceneConsume || memberPointLog.Scene == constant.MemberPointLogSceneReverse {
			preConsumption := memberChangePointConsumption[MemberUuid(memberPointLog.MemberUuid)]
			memberChangePointConsumption[MemberUuid(memberPointLog.MemberUuid)] = preConsumption.Add(decimal.NewFromFloat(memberPointLog.Value))
		}
	}

	memberUuids := make([]uint64, 0) // 积分有变动的会员
	for memberUuid := range memberChangePoint {
		memberUuids = append(memberUuids, uint64(memberUuid))
	}

	// 获取会员信息
	members, err := repository.NewMemberRepo(db).GetMembersByUuids(memberUuids)
	if err != nil {
		logger.Logger.Info("HandleMemberPoints process, GetMembersByUuids failed", zap.Any("memberUuids", memberUuids), zap.Error(err))
		return
	}

	// 更新会员积分
	logMemberInfoMap := make(map[string][]float64)
	for _, member := range members {
		beforePoint := member.GetPoints()
		memberPoint := 0.0
		memberPointConsumption := 0.0
		if memberChangePointMap, ok := memberChangePoint[MemberUuid(member.Uuid)]; ok {
			memberPoint = memberChangePointMap.InexactFloat64()
		}
		if memberChangePointConsumptionMap, ok := memberChangePointConsumption[MemberUuid(member.Uuid)]; ok {
			memberPointConsumption = memberChangePointConsumptionMap.InexactFloat64()
		}
		member.UpdatePoint(memberPoint, memberPointConsumption)
		logMemberInfoMap[member.Nickname] = []float64{beforePoint, memberPoint, member.GetPoints()}
	}

	uuids := make([]uint64, 0)
	for _, memberPointLog := range memberPointLogs {
		uuids = append(uuids, memberPointLog.Uuid)
	}

	// 更新会员积分,更新到数据库
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		for _, member := range members {
			if err := repository.NewMemberRepo(tx).Update(member.Uuid, map[string]any{
				"point":                             member.GetPoints(),
				"frozen_point":                      member.FrozenPoint,
				"accumulated_get_point":             member.AccumulatedGetPoint,
				"accumulated_consumption_get_point": member.AccumulatedConsumptionGetPoint,
			}); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 标记所有记录为已经处理
		if err := repository.NewMemberPointLogRepo(tx).UpdateProcessed(uuids); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		logger.Logger.Info("HandleMemberPoints process, Transaction failed", zap.Any("members", members), zap.Error(err))
		return
	}
	// 记录日志
	for nickname, info := range logMemberInfoMap {
		logger.Logger.Info("HandleMemberPoints process, UpdateMemberPoint", zap.Any("member", nickname), zap.Any("before point", info[0]), zap.Any("change point", info[1]), zap.Any("after point", info[2]))
	}
}

// 处理会员余额
func HandleMemberBalance(db *gorm.DB) {
	// 加锁, 避免并发问题
	lock.NewSystemLock().LockUuid(constant.LockNameMemberBalance)
	defer lock.NewSystemLock().UnlockUuid(constant.LockNameMemberBalance)

	// 查询余额变动
	memberBalanceLogRepo := repository.NewMemberBalanceLogRepo(db)
	memberBalanceLogs, err := memberBalanceLogRepo.GetMemberBalanceLogNotProcessed()
	if err != nil {
		logger.Logger.Info("HandleMemberBalance process, GetMemberBalanceLogNotProcessed failed", zap.Error(err))
		return
	}

	type MemberUuid uint64
	memberChangeBalance := make(map[MemberUuid]decimal.Decimal)     // 各个会员的余额变动
	memberChangeBalanceGift := make(map[MemberUuid]decimal.Decimal) // 各个会员的赠送帐户余额变动
	for _, memberBalanceLog := range memberBalanceLogs {
		// 累计同一个会员的余额变动
		pre := memberChangeBalance[MemberUuid(memberBalanceLog.MemberUuid)]
		money := decimal.NewFromFloat(memberBalanceLog.Money).Sub(decimal.NewFromFloat(memberBalanceLog.GiftMoney)) // 主余额变动金额=余额变动金额-赠送余额变动金额
		memberChangeBalance[MemberUuid(memberBalanceLog.MemberUuid)] = pre.Add(money)
		// 累计同一个会员的赠送帐户余额变动
		preGift := memberChangeBalanceGift[MemberUuid(memberBalanceLog.MemberUuid)]
		memberChangeBalanceGift[MemberUuid(memberBalanceLog.MemberUuid)] = preGift.Add(decimal.NewFromFloat(memberBalanceLog.GiftMoney))
	}

	memberUuids := make([]uint64, 0) // 余额有变动的会员
	for memberUuid := range memberChangeBalance {
		memberUuids = append(memberUuids, uint64(memberUuid))
	}
	memberUuidsGift := make([]uint64, 0) // 赠送帐户余额有变动的会员
	for memberUuid := range memberChangeBalanceGift {
		memberUuidsGift = append(memberUuidsGift, uint64(memberUuid))
	}
	memberUuids = append(memberUuids, memberUuidsGift...)
	// memberUuids 去重
	memberUuids = utils.RemoveDuplicates(memberUuids)

	// 获取会员信息
	members, err := repository.NewMemberRepo(db).GetMembersByUuids(memberUuids)
	if err != nil {
		logger.Logger.Info("HandleMemberBalance process, GetMembersByUuids failed", zap.Any("memberUuids", memberUuids), zap.Error(err))
		return
	}
	// 更新会员余额
	logMemberInfoMap := make(map[string][]float64)
	for _, member := range members {
		beforeBalance := member.Balance
		beforeBalanceGift := member.GiftBalance
		memberChangeBalance := memberChangeBalance[MemberUuid(member.Uuid)].InexactFloat64()
		memberChangeBalanceGift := memberChangeBalanceGift[MemberUuid(member.Uuid)].InexactFloat64()
		member.UpdateBalance(memberChangeBalance, memberChangeBalanceGift)
		logMemberInfoMap[member.Nickname] = []float64{beforeBalance, memberChangeBalance, member.Balance, beforeBalanceGift, memberChangeBalanceGift, member.GiftBalance}
	}

	uuids := make([]uint64, 0) // 未处理的余额变动记录
	for _, memberBalanceLog := range memberBalanceLogs {
		uuids = append(uuids, memberBalanceLog.Uuid)
	}

	// 更新会员余额,更新到数据库
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		for _, member := range members {
			if err := repository.NewMemberRepo(tx).Update(member.Uuid, map[string]any{
				"balance":             member.GetBalance(),
				"gift_balance":        member.GetGiftBalance(),
				"frozen_balance":      0,
				"frozen_gift_balance": 0,
			}); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 标记所有记录为已经处理
		if err := repository.NewMemberRepo(tx).UpdateProcessed(uuids); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		logger.Logger.Info("HandleMemberBalance process, Transaction failed", zap.Any("members", members), zap.Error(err))
		return
	}
	// 记录日志
	for nickname, info := range logMemberInfoMap {
		logger.Logger.Info("HandleMemberBalance process, UpdateMemberBalance", zap.Any("member", nickname), zap.Any("before balance", info[0]), zap.Any("change balance", info[1]), zap.Any("after balance", info[2]), zap.Any("before balance gift", info[3]), zap.Any("change balance gift", info[4]), zap.Any("after balance gift", info[5]))
	}
}

// 处理邀请有礼活动-统计获奖
func HandleActivityConsumption(payload event.CheckoutSaleOrderPayload) {
	// 已经进行过反结账的不处理
	if payload.SaleBill.IsReverseSettle() {
		return
	}

	// 加锁, 避免并发问题
	lock.NewSystemLock().LockUuid(constant.LockNameActivityConsumption)
	defer lock.NewSystemLock().UnlockUuid(constant.LockNameActivityConsumption)

	// 设置当前DB、公司设置
	db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)

	// 产品： 关闭营销活动后，不限制会员的登录行为。系统需停止该商家的营销活动活动，不进行营销活动消费累积计算和奖励发放。
	companySetting := repository.NewCompanySettingRepo(db).Get()
	if companySetting.IsOpenMarketing != 1 {
		return
	}

	// 设置当前上下文
	payload.Ctx.SetDB(db)
	payload.Ctx.SetCompanySetting(companySetting)
	if payload.Ctx.GetCompany().Uuid == 0 {
		company, err := repository.NewCompanyRepo(db).GetCompany(repository.CommonRepo.WhereByUuid(payload.CompanyUuid))
		if err != nil {
			logger.Logger.Error("发送奖励-获取当前公司数据失败", zap.Error(err))
			return
		}
		payload.Ctx.SetCompany(company)
	}

	// 处理邀请有礼活动-统计获奖
	for _, saleOrder := range payload.SaleBill.SaleOrders {
		if saleOrder.ConsumerUuid != 0 && saleOrder.Member != nil && saleOrder.Member.IsExistActivityAndReferrer() {

			if !saleOrder.IsSettled() {
				logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, SaleOrder not settled", zap.Any("saleOrder", saleOrder))
				continue
			}

			activity, err := repository.NewMarketingActivityRepo(db).GetActivity(saleOrder.Member.ActivityUuid)
			if err != nil || activity == nil {
				logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, GetActivity failed", zap.Any("activityUuid", saleOrder.Member.ActivityUuid), zap.Error(err))
				continue
			}
			if !activity.IsValid() {
				// 活动无效则跳过
				continue
			}

			consumptionAmount := saleOrder.GetFinalNoFeeAmount()

			// 发送奖励一
			{
				// 记录消费金额
				err = repository.NewMarketingActivityConsumptionRepo(db).CreateOrUpdateConsumption(
					activity.Uuid,
					saleOrder.Member.ReferrerUuid,
					saleOrder.ConsumerUuid,
					consumptionAmount,
				)
				if err != nil {
					logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, CreateConsumption failed", zap.Any("saleOrder", saleOrder), zap.Error(err))
				}
				// 发放奖励
				err := HandleActivitySendReward(payload, db, activity.Uuid, saleOrder.Member.ReferrerUuid)
				if err != nil {
					fmt.Println(err)
					logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, SendReward failed - 01", zap.Any("activityUuid", activity.Uuid), zap.Error(err))
				}
			}

			// 发送奖励二
			{
				referrer, err := repository.NewMemberRepo(db).GetMemberByReferrerUuid(saleOrder.Member.ReferrerUuid)
				if err != nil {
					fmt.Println(err)
					logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, GetMemberByReferrerUuid failed", zap.Any("referrerUuid", saleOrder.Member.ReferrerUuid), zap.Error(err))
					continue
				}
				if referrer.IsExistActivityAndReferrer() {
					// 记录消费金额
					err = repository.NewMarketingActivityConsumptionRepo(db).CreateOrUpdateConsumption(
						referrer.ActivityUuid,
						referrer.ReferrerUuid,
						saleOrder.ConsumerUuid,
						consumptionAmount,
					)
					if err != nil {
						logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, CreateConsumption failed", zap.Any("saleOrder", saleOrder), zap.Error(err))
					}
					// 发放奖励
					err = HandleActivitySendReward(payload, db, referrer.ActivityUuid, referrer.ReferrerUuid)
					if err != nil {
						fmt.Println(err)
						logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, SendReward failed - 02", zap.Any("activityUuid", referrer.ActivityUuid), zap.Error(err))
					}
				}
			}
		}
	}
}

// HandleActivitySendReward 发放奖励
func HandleActivitySendReward(payload event.CheckoutSaleOrderPayload, db *gorm.DB, activityUuid, memberUuid uint64) error {
	member, err := repository.NewMemberRepo(db).GetMemberRecord(repository.CommonRepo.WhereByUuid(memberUuid))
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("member not found")
	}

	// 获取活动信息
	activity, err := repository.NewMarketingActivityRepo(db).GetActivityAndPrizes(activityUuid)
	if err != nil {
		return err
	}
	if activity == nil {
		return fmt.Errorf("activity not found")
	}

	// 检查活动是否有效
	if !activity.IsValid() {
		return fmt.Errorf("activity is invalid")
	}

	// 检查是否开启奖励次数限制, 如果已经达到奖励次数限制，则不再发放奖励
	rewardCount, err := repository.NewMarketingActivityRecordRepo(db).GetRewardCount(activityUuid, memberUuid)
	if err != nil {
		return err
	}
	if activity.IsOpenRewardLimit == 1 && rewardCount >= activity.RewardLimit {
		return fmt.Errorf("reward limit reached")
	}

	// 获取推荐人的总消费金额
	consumptionAmount, err := repository.NewMarketingActivityConsumptionRepo(db).GetByActivityAndReferrerConsumptionAmount(activityUuid, memberUuid)
	if err != nil {
		return err
	}

	// 计算应该发放的奖励次数
	rewardCountToGive := utils.IfInt(activity.RewardConditionAmount <= 0, 0, int(consumptionAmount/activity.RewardConditionAmount))
	if rewardCountToGive <= 0 {
		return fmt.Errorf("no reward to give")
	}
	// 最多等于奖励次数限制
	if activity.IsOpenRewardLimit == 1 {
		if rewardCountToGive > int(activity.RewardLimit) {
			rewardCountToGive = int(activity.RewardLimit)
		}
	}
	// 计算应该发放的奖励次数，减去已经发放的奖励次数
	rewardCountToGive = rewardCountToGive - int(rewardCount)
	if rewardCountToGive <= 0 {
		return nil
	}

	// 获取数据库管理器
	dbm := database.GetDBManager(config.DatabaseConf{})

	// 发放奖励
	if activity.RewardType == 0 {

		// 发放优惠券
		marketingCoupon := &model.MarketingCoupon{}
		if len(activity.Prizes) > 0 && activity.Prizes[0] != nil && activity.Prizes[0].Coupon != nil {
			marketingCoupon = activity.Prizes[0].Coupon
		}
		if marketingCoupon == nil || marketingCoupon.Uuid == 0 {
			return nil
		}

		for i := 0; i < rewardCountToGive; i++ {
			err = db.Transaction(func(tx *gorm.DB) error {
				// 创建优惠券
				err = repository.NewMarketingCouponRepo(tx).DecreaseCouponQuantity(marketingCoupon.Uuid, memberUuid, activityUuid)
				if err != nil {
					return err
				}
				// 创建奖励记录
				err = repository.NewMarketingActivityRecordRepo(tx).CreateRecord(&model.MarketingActivityRecord{
					ActivityUuid:   activityUuid,
					MemberUuid:     memberUuid,
					RewardCount:    1,
					LastRewardTime: time.Now().Unix(),
					PrizeUuid: func() uint64 {
						if len(activity.Prizes) > 0 && activity.Prizes[0] != nil {
							return activity.Prizes[0].PrizeUuid
						}
						return 0
					}(),
				})
				if err != nil {
					return err
				}
				// 创建会员优惠券
				err = repository.NewMemberCouponRepo(tx).CreateMemberCoupon(&model.MemberCoupon{
					MemberUuid:     memberUuid,
					Type:           "deduction",
					Status:         0,
					CouponUuid:     marketingCoupon.Uuid,
					Name:           marketingCoupon.Name,
					DeductionType:  marketingCoupon.DeductionType,
					DayStartTime:   marketingCoupon.DayStartTime,
					DayEndTime:     marketingCoupon.DayEndTime,
					ValidStartTime: marketingCoupon.ValidStartTime,
					ValidEndTime:   marketingCoupon.ValidEndTime,
					Amount:         marketingCoupon.Amount,
					StartTime:      time.Now().Unix(),
					EndTime:        time.Now().AddDate(0, 0, marketingCoupon.ValidDays).Unix(),
				})
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		// 发送短信
		if rewardCountToGive > 0 && activity.IsSendSms == 1 {
			go func() {
				err := service.NewSMSSrv(dbm).SendMemberCouponSMS(payload.Ctx, member.Phone, &sms.MemberCouponRequest{CouponNum: uint64(rewardCountToGive)})
				if err != nil {
					fmt.Println("HandleActivitySendReward process, SendMemberCouponSMS failed", zap.Any("activityUuid", activityUuid), zap.Any("phone", member.Phone), zap.Error(err))
					logger.Logger.Info("HandleActivitySendReward process, SendMemberCouponSMS failed", zap.Any("activityUuid", activityUuid), zap.Any("phone", member.Phone), zap.Error(err))
				}
			}()
		}

	} else if activity.RewardType == 1 {

		// 发放积分
		if activity.RewardValue == 0 {
			return nil
		}

		points := decimal.NewFromFloat(activity.RewardValue).Mul(decimal.NewFromFloat(float64(rewardCountToGive))).InexactFloat64()

		err = db.Transaction(func(tx *gorm.DB) error {

			// 创建奖励记录
			err = repository.NewMarketingActivityRecordRepo(tx).CreateRecord(&model.MarketingActivityRecord{
				ActivityUuid:   activityUuid,
				MemberUuid:     memberUuid,
				RewardCount:    rewardCountToGive,
				LastRewardTime: time.Now().Unix(),
				PrizeUuid:      0,
				RewardValue:    points,
			})
			if err != nil {
				return err
			}

			// 更新会员积分
			if err := repository.NewMemberRepo(tx).Update(memberUuid, map[string]any{
				"frozen_point": gorm.Expr("frozen_point + ?", points),
			}); err != nil {
				return err
			}

			// 创建积分记录
			_, err := repository.NewMemberPointLogRepo(tx).Create(model.MemberPointLog{
				MemberUuid:  memberUuid,
				Scene:       constant.MemberPointLogSceneMarketingActivity,
				Value:       points,
				Describe:    "邀请消费有礼",
				RelatedUuid: activityUuid,
			})
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		// 发布"积分变动"事件
		go HandleMemberPoints(db)

		// 发送短信
		if activity.IsSendSms == 1 {
			go func() {
				err := service.NewSMSSrv(dbm).SendMemberPointsSMS(payload.Ctx, member.Phone, &sms.MemberPointsRequest{Points: points})
				if err != nil {
					fmt.Println("HandleActivitySendReward process, SendMemberPointsSMS failed", zap.Any("activityUuid", activityUuid), zap.Any("phone", member.Phone), zap.Error(err))
					logger.Logger.Info("HandleActivitySendReward process, SendMemberPointsSMS failed", zap.Any("activityUuid", activityUuid), zap.Any("phone", member.Phone), zap.Error(err))
				}
			}()
		}
	}

	return nil
}
