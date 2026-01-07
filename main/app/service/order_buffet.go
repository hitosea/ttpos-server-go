package service

import (
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetOrderChangeBuffet 桌台订单自助餐信息
func (s *orderSrv) GetOrderChangeBuffet(ctx context.Context, saleBillUuid, saleOrderUuid uint64) (resp.OrderBuffetResp, error) {
	res := resp.OrderBuffetResp{
		BuffetUuids:         make([]uint64, 0),
		BuffetCustomerTypes: make([]resp.DeskBuffetCustomerType, 0),
	}
	saleBill, err := repository.NewOrderRepo(ctx.GetDB()).GetOrderBuffetInfo(saleBillUuid, saleOrderUuid)
	if err != nil {
		return res, errors.ErrInternal
	}
	if !saleBill.IsBuffetSaleBill() {
		return res, nil
	}
	if len(saleBill.SaleOrders) == 0 {
		return res, nil
	}
	var customerTypes []resp.DeskBuffetCustomerType
	var buffetUuids []uint64
	for _, customerType := range saleBill.SaleOrders[0].SaleOrderBuffetCustomerTypes {
		customerTypes = append(customerTypes, resp.DeskBuffetCustomerType{
			Uuid:    customerType.BuffetCustomerTypePrice.BuffetCustomerType.Uuid,
			MealNum: customerType.Num,
		})
		buffetUuids = append(buffetUuids, customerType.BuffetPackageUuid)
	}
	return resp.OrderBuffetResp{
		BuffetUuids:         buffetUuids,
		BuffetCustomerTypes: customerTypes,
	}, nil
}

// OrderChangeBuffet 调整自助餐
func (s *orderSrv) OrderChangeBuffet(ctx context.Context, req req.OrderChangeBuffetReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	if err := req.Validate(); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}
	if !saleBill.IsBuffetSaleBill() {
		return nil, errors.New("当前订单不是自助餐类型，无法调整自助餐")
	}
	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.New("当前订单已拆单，请前去收银机操作")
	}
	if saleBill.IsSplit() {
		return nil, errors.New("当前订单已拆单，无法调整自助餐")
	}
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderUpdateMealNum, 0); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前的所有商品
	saleOrderProductAll := saleBill.GetSaleOrderProductAll()
	// 获取自助餐信息
	buffetList, err := repository.NewBuffetRepo(db).GetBuffetListByUuids(req.BuffetUuids)
	if err != nil || len(buffetList) != len(req.BuffetUuids) {
		return nil, errors.New("自助餐不存在")
	}
	// 销售订单
	saleOrder := saleBill.SaleOrders[0]

	// 是否新增
	oldBuffetIds := []uint64{}
	if saleBill.BuffetPackage1Uuid != 0 {
		oldBuffetIds = append(oldBuffetIds, saleBill.BuffetPackage1Uuid)
	}
	if saleBill.BuffetPackage2Uuid != 0 {
		oldBuffetIds = append(oldBuffetIds, saleBill.BuffetPackage2Uuid)
	}
	addBuffetIds := slice.Difference(req.BuffetUuids, oldBuffetIds)
	removeBuffetIds := slice.Difference(oldBuffetIds, req.BuffetUuids)

	// 获取自助餐顾客
	customerTypes := []model.BuffetUuidMapBuffetCustomerTypes{}
	copier.Copy(&customerTypes, req.BuffetCustomerTypes)
	saleOrderCustomerTypes, buffetUuids, mealNum, maxTimeLimit, _, _ := saleOrder.GetSaleOrderBuffetCustomerTypes(
		buffetList,
		req.BuffetUuids,
		customerTypes,
		saleBill.SaleBillSetting,
	)

	// 修改
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 删除原来的 CustomerType
		if err := repository.NewOrderRepo(tx).DeleteSaleOrderBuffetCustomerType(saleOrder.Uuid); err != nil {
			return errors.WithMessage(err)
		}
		saleBill.DeleteSaleOrderBuffetCustomerTypeAll(saleOrder.Uuid)

		// 创建新的顾客
		if len(saleOrderCustomerTypes) > 0 {
			for _, customer := range saleOrderCustomerTypes {
				saleBill.AddSaleOrderBuffetCustomerType(saleOrder.Uuid, customer)
			}
		}

		// 如果改了套餐信息，需要处理时间 和调整商品是否自助餐
		if len(addBuffetIds) != 0 || len(removeBuffetIds) != 0 {
			// 设置自助餐的开启时间和时长
			saleBill.SetBuffetStartTimeAndDuration(maxTimeLimit)
			// 更新商品为自助餐商品
			buffetProductUuids := []uint64{}
			for _, buffet := range buffetList {
				for _, buffetProduct := range buffet.BuffetProducts {
					buffetProductUuids = append(buffetProductUuids, buffetProduct.ProductPackageUuid)
				}
			}
			for _, product := range saleOrderProductAll {
				if product.IsBuffetProduct() {
					return errors.New("请先清除自助餐套餐内商品")
				}
				if slices.Contains(buffetProductUuids, product.ProductPackageUuid) {
					product.IsBuffet = 1
					if err := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProduct(product); err != nil {
						return errors.WithMessage(err)
					}
				}
			}
		}

		// 保存账单。不能用这个方法来创建销售账单，故不使用UpdateOrCreate
		saleBill.SetMealNum(mealNum)
		saleBill.SetBuffetPackage(buffetUuids)
		saleBill.SetDelayProductMealNum(mealNum)

		// 设置自助餐名称快照（JSON 方案）
		// Requirement: story-main-buffet-package-name-snapshot-fix
		// 创建 map 确保按照 buffetUuids 的顺序匹配（buffetUuids 是从 GetSaleOrderBuffetCustomerTypes 返回的，已去重且保持顺序）
		buffetMap := make(map[uint64]*model.BuffetPackage)
		for _, buffet := range buffetList {
			buffetMap[buffet.Uuid] = buffet
		}
		// 按照 buffetUuids 的顺序设置快照
		if len(buffetUuids) >= 1 {
			if buffet1, ok := buffetMap[buffetUuids[0]]; ok && !buffet1.MultiLanguageName.IsNullName() {
				if err := saleBill.SetBuffetPackage1NameSnapshot(buffet1.MultiLanguageName); err != nil {
					logger.Logger.Error("保存自助餐套餐1名称快照失败", zap.Error(err))
				}
			}
		}
		if len(buffetUuids) >= 2 {
			if buffet2, ok := buffetMap[buffetUuids[1]]; ok && !buffet2.MultiLanguageName.IsNullName() {
				if err := saleBill.SetBuffetPackage2NameSnapshot(buffet2.MultiLanguageName); err != nil {
					logger.Logger.Error("保存自助餐套餐2名称快照失败", zap.Error(err))
				}
			}
		}

		//
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 重新计算销售账单.
	{
		newSaleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
		if err != nil {
			return nil, errors.WithMessage(err, "repository.NewOrderRepo(db).GetSaleBillAllInfo failed")
		}
		if err := s.CalcAndSaveSaleBill(ctx, db, newSaleBill); err != nil {
			return nil, errors.WithMessage(err)
		}
	}

	if adapter.IsObjectStorageCacheEnabled(ctx.GetCompanyUuid()) {
		if ctx.GetSource() == constant.SourceAssistant {
			return nil, nil // 如果开启对象存储缓存且是助手端，则不返回购物车商品数据
		}
	}

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderChangeBuffetClock 调整自助餐加钟
func (s *orderSrv) OrderChangeBuffetClock(ctx context.Context, req req.OrderChangeBuffetClockReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 系统级验证
	companySetting, err := s.settingSrv.GetCompanySetting(ctx)
	if err != nil {
		return nil, err
	}
	buffetSetting, buffetErr := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if buffetErr != nil {
		return nil, buffetErr
	}
	if buffetSetting.IsAddClock != "1" {
		return nil, errors.New("未开启加钟")
	}

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.New("销售账单不存在")
	}
	if !saleBill.IsBuffetSaleBill() {
		return nil, errors.New("当前不是自助餐类订单，无法添加加钟")
	}
	if saleBill.GetBuffetRemainingSeconds() == -1 {
		return nil, errors.New("当前套餐已经是无限时，无法添加加钟")
	}
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderClock, req.SaleOrderUuid); err != nil {
		return nil, err
	}

	// 获取销售订单
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 获取加钟
	delays, err := base.NewBuffetDelayRepo(db).GetBuffetDelayListByUuids(req.DelayUuids)
	if err != nil || len(delays) != len(req.DelayUuids) {
		return nil, errors.New("加钟不存在")
	}

	// 修改
	if err := db.Transaction(func(tx *gorm.DB) error {
		//
		totalDelayTime := int64(0)
		for _, delay := range delays {
			delayProduct := model.SaleOrderBuffetDelayProduct{
				SaleOrderUuid:   saleOrder.Uuid,
				BuffetDelayUuid: delay.Uuid,
				Price:           delay.Price,
				Name:            delay.Name,
				Num:             saleBill.MealNum,
				DelayTime:       delay.DelayTime,
				Sign:            uuid.New().String(),
			}
			//
			totalDelayTime += delay.DelayTime * 60
			//
			saleBill.AddSaleOrderBuffetDelayProduct(saleOrder.Uuid, delayProduct)
		}

		// 设置加钟时长
		saleBill.AddBuffetDelayStartTimeAndDuration(int(totalDelayTime))

		// 计算销售账单金额
		if err = s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// OrderDeskBuffetProductList 获取桌台的自助餐商品列表. 实现思路：通过账单uuid获取该账单的自助餐套餐，然后通过自助餐套餐uuid获取该自助餐套餐的商品列表
func (s *orderSrv) OrderDeskBuffetProductList(ctx context.Context, req req.OrderChangeBuffetProductListReq) (*resp.BuffetProductList, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	// 获取销售账单
	saleBill, err := repository.NewSaleBillRepo(db).GetSaleBillBuffetProductList(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	buffetProductList := saleBill.GetBuffetProductList()

	return &buffetProductList, nil
}
