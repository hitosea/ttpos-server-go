package service

import (
	contexts "context"
	"fmt"
	"slices"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"go.uber.org/zap"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CreateInstantOrder 创建点餐订单
func (s *orderSrv) CreateInstantOrder(ctx context.Context) (resp.CreateInstantOrderResp, error) {
	dbId := ctx.GetDbId()
	var billUuid uint64
	var orderUuid uint64
	db := s.dbm.GetDB(dbId)

	// 判断是否有待支付、未挂单的订单
	_, hasInstantOrder, err := HasInstantOrder(ctx, db)
	if err != nil {
		return resp.CreateInstantOrderResp{}, errors.WithMessage(err)
	}
	if hasInstantOrder {
		return resp.CreateInstantOrderResp{}, errors.New("有待支付、未挂单的订单")
	}
	if err := repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
		// 创建订单编号
		orderNo, err := s.createOrderNo(tx, constant.OrderSourceInstant)
		if err != nil {
			ctx.Log().Error("订单编号生成失败", zap.Error(err))
			return errors.WithMessage(err, "订单编号生成失败")
		}

		serialNo, err := s.createInstantOrderSerialNo(ctx, tx)
		if err != nil {
			ctx.Log().Error("订单序号生成失败", zap.Error(err))
			return errors.WithMessage(err, "订单序号生成失败")
		}
		// 创建销售账单
		// ⚠️ 重要：source 和 client_version 必须一起设置，确保数据一致性
		saleBill, err := repository.NewOrderRepo(tx).CreateSaleBill(model.SaleBill{
			OrderNo:       orderNo,
			SerialNo:      serialNo,
			BillType:      constant.OrderSourceMapToBillType[constant.OrderSourceInstant],
			DiningMethod:  constant.SaleBillDiningMethodDineIn,
			DeviceUuid:    ctx.GetDeviceUuid(),
			Source:        constant.MapJwtSourceToSaleBillSource(ctx.GetSource()),
			ClientVersion: constant.NormalizeClientVersion(ctx.GetVersion()),
		})
		if err != nil {
			return errors.WithMessage(err)
		}

		// 创建销售账单设置
		saleBillSetting, err := s.CreateSaleBillSetting(ctx, tx, saleBill.Uuid, saleBill.DeskUuid, false)
		if err != nil {
			return errors.WithMessage(err)
		}

		// 创建销售订单
		saleOrder, errCreateSaleOrder := createSaleOrder(ctx, tx, saleBillSetting, saleBill.Uuid, orderNo)
		if errCreateSaleOrder != nil {
			return errCreateSaleOrder
		}

		billUuid = saleBill.Uuid
		orderUuid = saleOrder.Uuid

		return nil
	}); err != nil {
		return resp.CreateInstantOrderResp{}, errors.WithMessage(err)
	}

	return resp.CreateInstantOrderResp{
		SaleBillUuid:  billUuid,
		SaleOrderUuid: orderUuid,
	}, nil
}

// CreateDeskOrder 创建桌台订单
func (s *orderSrv) CreateDeskOrder(ctx context.Context, req req.DeskOrderCreateReq) (resp.CreateDeskOrderResp, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.DeskUuid)
		defer lock.NewSystemLock().UnlockUuid(req.DeskUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	desk, err := repository.NewDeskRepo(db).GetDeskRecord(req.DeskUuid)
	if err != nil {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err, "无法找到空闲桌台")
	}
	if !desk.IsAvailableDesk() {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err, "该桌台非空闲桌台")
	}
	saleBillUuid, _ := utils.GetID()
	desk.SetOpenDesk(saleBillUuid)

	// 创建订单编号
	orderNo, err := s.createOrderNo(db, constant.OrderSourceDesk)
	if err != nil {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err, "订单编号生成失败")
	}

	// 构建销售账单
	staff := ctx.GetStaff()
	saleBill := model.NewDeskSaleBill(saleBillUuid, orderNo, req.BuffetUuids, req.GetMealNum(), req.Remark, req.DeskUuid, desk.DeskNo, staff.DutyNo, staff.Uuid, staff.GetUserName())

	// 构建销售账单设置
	saleBillSetting, err := s.NewSaleBillSetting(ctx, saleBill.Uuid, req.DeskUuid, false)
	if err != nil {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err)
	}
	// 构建销售订单
	saleOrder := model.NewSaleOrder(ctx.GetDeviceSn(), saleBill.Uuid, saleBill.OrderNo, *saleBillSetting)
	staffShiftLogUuid := uint64(0)
	{
		staffShiftLog, err := GetCurrentStaffShiftLog(db, staff.Uuid)
		if err != nil {
			logger.Logger.Error("获取当前员工班次信息失败", zap.Error(err))
		} else {
			staffShiftLogUuid = staffShiftLog.Uuid
		}
	}

	saleOrder.StaffShiftLogUuid = staffShiftLogUuid

	// 获取自助餐信息
	buffetList, err := repository.NewBuffetRepo(db).GetBuffetListByUuids(req.BuffetUuids)
	if err != nil {
		return resp.CreateDeskOrderResp{}, nil
	}

	// 设置自助餐名称快照（JSON 方案）
	// Requirement: story-main-buffet-package-name-snapshot-fix
	// 创建 map 确保按照 req.BuffetUuids 的顺序匹配
	buffetMap := make(map[uint64]*model.BuffetPackage)
	for _, buffet := range buffetList {
		buffetMap[buffet.Uuid] = buffet
	}
	// 按照 req.BuffetUuids 的顺序设置快照
	if len(req.BuffetUuids) >= 1 {
		if buffet1, ok := buffetMap[req.BuffetUuids[0]]; ok && !buffet1.MultiLanguageName.IsNullName() {
			if err := saleBill.SetBuffetPackage1NameSnapshot(buffet1.MultiLanguageName); err != nil {
				ctx.Log().Error("保存自助餐套餐1名称快照失败", zap.Error(err))
			}
		}
	}
	if len(req.BuffetUuids) >= 2 {
		if buffet2, ok := buffetMap[req.BuffetUuids[1]]; ok && !buffet2.MultiLanguageName.IsNullName() {
			if err := saleBill.SetBuffetPackage2NameSnapshot(buffet2.MultiLanguageName); err != nil {
				ctx.Log().Error("保存自助餐套餐2名称快照失败", zap.Error(err))
			}
		}
	}

	// 构建自助餐顾客列表
	buffetCustomerTypes := []model.BuffetUuidMapBuffetCustomerTypes{}
	copier.Copy(&buffetCustomerTypes, req.BuffetCustomerTypes)
	saleOrderBuffetCustomerTypes, _, _, maxTimeLimit, nonOrderingTime, reminderOrderTime := saleOrder.GetSaleOrderBuffetCustomerTypes(buffetList, req.BuffetUuids, buffetCustomerTypes, saleBillSetting)

	// 开始事务
	if err := db.Transaction(func(tx *gorm.DB) error {

		// 标记
		tx = tx.WithContext(contexts.WithValue(contexts.Background(), constant.OrderOperateSource, constant.OrderOpenTable))

		// 如果是自助餐，有顾客列表的话，创建顾客
		if len(saleOrderBuffetCustomerTypes) > 0 {
			for _, customer := range saleOrderBuffetCustomerTypes {
				if _, err = repository.NewOrderRepo(tx).CreateSaleOrderBuffetCustomerType(*customer); err != nil {
					return errors.WithMessage(err)
				}
			}
			if maxTimeLimit == -1 {
				saleBill.BuffetDuration = 0
			} else {
				saleBill.BuffetDuration = uint(maxTimeLimit)
				saleBill.NonOrderingTime = nonOrderingTime
				saleBill.ReminderOrderTime = reminderOrderTime
			}
		}

		// ⚠️ 重要：source 和 client_version 必须一起设置，确保数据一致性
		// 设置订单来源和客户端版本
		saleBill.Source = constant.MapJwtSourceToSaleBillSource(ctx.GetSource())
		saleBill.ClientVersion = constant.NormalizeClientVersion(ctx.GetVersion())

		// 创建销售账单
		if _, errCreateSaleBill := repository.NewOrderRepo(tx).CreateSaleBill(*saleBill); errCreateSaleBill != nil {
			return errCreateSaleBill
		}

		// 创建销售账单设置
		if _, errCreateSaleBillSetting := repository.NewOrderRepo(db).CreateSaleBillSetting(*saleBillSetting); errCreateSaleBillSetting != nil {
			return errCreateSaleBillSetting
		}

		// 创建销售订单
		if _, errCreateSaleOrder := repository.NewOrderRepo(tx).CreateSaleOrder(*saleOrder); errCreateSaleOrder != nil {
			return errCreateSaleOrder
		}

		// 新桌台的状态
		if errUpdate := repository.NewDeskRepo(tx).UpdateDesk(req.DeskUuid, *desk); errUpdate != nil {
			return errUpdate
		}

		// 提交完订单后，重新查询并计算金额。 todo 改为sale_bill保存数据库前计算好金额。
		{
			// 获取销售账单信息
			saleBill, err := repository.NewOrderRepo(tx).GetSaleBillAllInfo(saleBill.Uuid)
			if err != nil {
				return errors.WithMessage(err)
			}
			// 计算销售账单金额
			if err = s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
				return errors.WithMessage(err)
			}
		}
		return nil
	}); err != nil {
		return resp.CreateDeskOrderResp{}, errors.WithMessage(err)
	}

	return resp.CreateDeskOrderResp{
		SaleBillUuid:  saleBill.Uuid,
		SaleOrderUuid: saleOrder.Uuid,
	}, nil
}

// SetOrderSource 设置销售账单订单来源
func (s *orderSrv) SetOrderSource(ctx context.Context, saleBillUuid uint64, orderSourceUuid uint64) (resp.ShopCart, error) {
	if saleBillUuid == 0 {
		return resp.ShopCart{}, errors.New("销售账单 UUID 不能为空")
	}
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(saleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	err := repository.NewSaleBillRepo(db).UpdateOrderSource(saleBillUuid, orderSourceUuid)
	if err != nil {
		return resp.ShopCart{}, errors.WithMessage(err)
	}
	shopCart, err := s.GetOrderCartInfo(ctx, saleBillUuid)
	if err != nil {
		return resp.ShopCart{}, errors.WithMessage(err)
	}
	return *shopCart, nil
}

// SetNationality 设置销售账单国籍
func (s *orderSrv) SetNationality(ctx context.Context, saleBillUuid uint64, nationalityUuid uint64) (resp.ShopCart, error) {
	if saleBillUuid == 0 {
		return resp.ShopCart{}, errors.New("销售账单 UUID 不能为空")
	}
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(saleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	err := repository.NewSaleBillRepo(db).UpdateNationality(saleBillUuid, nationalityUuid)
	if err != nil {
		return resp.ShopCart{}, errors.WithMessage(err)
	}
	shopCart, err := s.GetOrderCartInfo(ctx, saleBillUuid)
	if err != nil {
		return resp.ShopCart{}, errors.WithMessage(err)
	}
	return *shopCart, nil
}

// IsCellCancelOrder 判断订单是否可以取消
func (s *orderSrv) IsCellCancelOrder(ctx context.Context, saleBillUuid uint64) (model.SaleBill, error) {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	orderRepo := repository.NewOrderRepo(db)
	billInfo, err := orderRepo.GetSaleBillInfo(saleBillUuid, constant.OptionalUuid)
	if err != nil {
		return model.SaleBill{}, errors.WithMessage(err)
	}
	if slices.Contains([]string{constant.SourceShop}, ctx.GetSource()) {
		if orderRepo.IsPartiallyPaid(saleBillUuid) {
			return model.SaleBill{}, errors.New("当前订单已被部分支付，不支持取消")
		}
	}
	if err := billInfo.ValidateOrderStatus(ctx.GetSource(), constant.OrderOrderCancel, 0); err != nil {
		return model.SaleBill{}, errors.WithMessage(err)
	}
	if !slices.Contains([]string{constant.SourceShop}, ctx.GetSource()) {
		if orderRepo.IsPartiallyPaid(saleBillUuid) {
			return model.SaleBill{}, errors.New("当前订单已被部分支付，不支持取消")
		}
	}
	return billInfo, nil
}

// HideOrder 隐藏订单（挂单）
func (s *orderSrv) HideOrder(ctx context.Context, saleBillUuid uint64) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(saleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillRecord(saleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if billInfo.ID == 0 {
		return nil, errors.New("找不到订单")
	}
	if billInfo.Status != constant.SaleBillStatusPending {
		return nil, errors.New("订单状态不允许挂单")
	}

	// 隐藏
	err = orderRepo.HideOrder(saleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"挂单"事件
	utils.Go(func() {
		event.NewSystemBus().PublishHideSaleBillEvent(event.HideSaleBillPayload{
			BasePayload: event.BasePayload{ // 挂单
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: saleBillUuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
	})
	// 获取新的数据
	info, err := s.GetOrderCartInfoByDeviceSn(ctx, ctx.GetDeviceSn())
	if err != nil {
		fmt.Println("获取订单信息失败", err)
		return &resp.ShopCart{SaleOrderList: make([]resp.SaleOrder, 0)}, nil
	}

	return info, nil
}

// ShowOrder 显示订单
func (s *orderSrv) ShowOrder(ctx context.Context, req req.OrderShowReq) (*resp.ShopCart, error) {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	ctx.Log().Debug("取单", zap.Any("request", req))
	db := s.dbm.GetDB(ctx.GetDbId())

	// 判断是否有未挂单的点餐账单
	currentSaleBillUuid, err := repository.NewOrderRepo(db).HasShowOrder(ctx.GetDeviceUuid())
	if err != nil {
		ctx.Log().Error("判断是否有未挂单的点餐账单失败", zap.Error(err))
		return nil, errors.WithMessage(err, "判断是否有未挂单的点餐账单失败")
	}
	if currentSaleBillUuid != 0 {
		// 如果未挂单的点餐账单没有商品，则删除该订单，并允许取单
		// 当前销售账单数据
		currentSaleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(currentSaleBillUuid)
		if errSaleBill != nil {
			return nil, errSaleBill
		}
		if len(currentSaleBill.GetSaleOrderProductAll()) == 0 {
			// 软删除sale_bill和sale_order
			repository.NewSaleBillRepo(db).DeleteSaleBill(currentSaleBill.Uuid)
			for _, saleOrder := range currentSaleBill.SaleOrders {
				repository.NewSaleOrderRepo(db).DeleteSaleOrder(saleOrder.Uuid)
			}
		} else {
			return nil, errors.New("该设备有未挂单的点餐账单，禁止取单")
		}
	}

	saleBill, err := repository.NewOrderRepo(db).GetSaleBillRecord(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}

	if saleBill.IsShowSaleBill() {
		return nil, errors.New("该账单已取出")
	}

	// 修改销售账单信息，标记账单取出
	saleBill.SetShowSaleBill(ctx.GetDeviceUuid())
	// 更新销售账单
	if err := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); err != nil {
		return nil, errors.WithMessage(err, "更新销售账单失败", fmt.Sprintf("NewSaleBillRepo(db).UpdateSaleBillRecord failed, sale_bill uuid:%d", saleBill.Uuid))
	}

	// 发布"取单"事件
	utils.Go(func() {
		event.NewSystemBus().PublishShowSaleBillEvent(event.ShowSaleBillPayload{
			BasePayload: event.BasePayload{ // 取单
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: req.SaleBillUuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
	})

	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return info, nil
}

// InstantHideOrderList 获取挂单订单列表
func (s *orderSrv) InstantHideOrderList(ctx context.Context, req req.HideSaleBillListReq) (*resp.InstantHideOrderListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBillRepo := repository.NewSaleBillRepo(db)

	// 查询所有已挂单的点餐销售账单
	saleBills, total, err := saleBillRepo.GetHideSaleBillList(req.PageNo, req.PageSize, ctx.GetDeviceUuid())
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	hideSaleBills := make([]resp.InstantHideSaleBill, 0)
	for _, saleBill := range saleBills {
		if saleBill.IsShowSaleBill() || saleBill.IsDeskSaleBill() || saleBill.IsDelete() {
			continue
		}
		listMap := make(map[string]resp.Product) // 商品列表，key为商品sign
		for _, saleOrder := range saleBill.SaleOrders {
			for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
				if saleOrderProduct.IsPackageSubProduct() {
					// 套餐子商品不显示
					continue
				}
				if saleOrderProduct.IsDelete() || saleOrderProduct.Num == 0 {
					continue
				}
				if product, ok := listMap[saleOrderProduct.Sign]; !ok {
					productPrice := decimal.NewFromFloat(saleOrderProduct.Price).Mul(saleOrderProduct.GetNumDecimal()).InexactFloat64()
					newProduct := resp.Product{
						LocaleName:    saleOrderProduct.MultiLanguageName.GetNames(),
						Num:           saleOrderProduct.Num,
						SalePrice:     productPrice,
						DiscountPrice: productPrice,
					}
					// 如果是套餐商品，则设置套餐商品列表
					if saleOrderProduct.IsPackageProduct() {
						subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
						packageProductList := make([]resp.PackageProduct, 0)
						for _, subProduct := range subProducts {
							packageProductList = append(packageProductList, resp.PackageProduct{
								Uuid:       subProduct.Uuid,
								LocaleName: subProduct.MultiLanguageName.GetNames(),
								Num:        subProduct.Num,
								UnitNum:    subProduct.GetProductNum(),
								AddPrice:   subProduct.AddPrice, // 子商品加价金额
							})
						}
						newProduct.PackageProductList = resp.PackageProductList{
							List: packageProductList,
						}
					}
					listMap[saleOrderProduct.Sign] = newProduct
				} else {
					productPrice := decimal.NewFromFloat(saleOrderProduct.Price).Mul(saleOrderProduct.GetNumDecimal())
					price := productPrice.Add(decimal.NewFromFloat(product.SalePrice)).InexactFloat64()
					product.Num += saleOrderProduct.Num
					product.SalePrice = price
					product.DiscountPrice = price
					// 如果是套餐商品，则更新套餐商品列表
					if saleOrderProduct.IsPackageProduct() {
						for index := range product.PackageProductList.List {
							unitNum := decimal.NewFromFloat(saleOrderProduct.GetUnitNum())                  // 每份套餐的子商品数量
							num := decimal.NewFromFloat(product.Num).Mul(unitNum).Round(3).InexactFloat64() // 套餐数量*每份套餐的子商品数量= 子商品的数量
							product.PackageProductList.List[index].Num = num
						}
					}
				}
			}
		}
		list := make([]resp.Product, 0)
		for sign := range listMap {
			list = append(list, listMap[sign])
		}
		productList := resp.InstantHideSaleProductList{List: list}
		hideSaleBill := resp.InstantHideSaleBill{
			SaleBillUuid: saleBill.Uuid,
			SerialNo:     saleBill.SerialNo,
			Amount:       saleBill.Amount,
			HideBillTime: saleBill.HideBillTime,
			Products:     productList,
		}
		hideSaleBills = append(hideSaleBills, hideSaleBill)
	}

	res := &resp.InstantHideOrderListResp{
		List: hideSaleBills,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}
	return res, nil
}

// OrderTakeout 打包
func (s *orderSrv) OrderTakeout(ctx context.Context, req req.OrderTakeoutReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 当不填销售账单ID时，表示要新建一个销售账单
	if req.SaleBillUuid == 0 {
		billInfo, hasInstantOrder, err := HasInstantOrder(ctx, s.dbm.GetDB(ctx.GetDbId()))
		if err != nil {
			return nil, err
		}
		if billInfo != nil && hasInstantOrder {
			req.SaleBillUuid = billInfo.Uuid
		} else {
			order, err := s.CreateInstantOrder(ctx)
			if err != nil {
				ctx.Log().Info("添加商品时点餐订单创建失败", zap.Any("err", err.Error()))
				return nil, errors.WithMessage(err)
			}
			ctx.Log().Debug("添加商品时点餐订单创建成功", zap.Any("order info", order))
			req.SaleBillUuid = order.SaleBillUuid
		}
	}

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取操作的销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderTakeout, 0); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 修改销售账单状态
	if req.Takeout {
		saleBill.SetTakeoutSaleBill(constant.SaleBillDiningMethodTakeout)
	} else {
		saleBill.SetTakeoutSaleBill(constant.SaleBillDiningMethodDineIn)
	}

	saleBill.CalcAll()

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"整单打包"或"取消整单打包"事件
	utils.Go(func() {
		if req.Takeout {
			s.bus.PublishWrapSaleBillEvent(event.WrapSaleBillPayload{
				BasePayload: event.BasePayload{ // 整单打包
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: req.SaleBillUuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
			})
		} else {
			s.bus.PublishUnwrapSaleBillEvent(event.UnwrapSaleBillPayload{
				BasePayload: event.BasePayload{ // 取消整单打包
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: req.SaleBillUuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
			})
		}
	})

	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return info, nil
}

// OrderChangePopulation  修改订单人数
func (s *orderSrv) OrderChangePopulation(ctx context.Context, req req.OrderChangePopulationReq) (*resp.ShopCart, error) {
	if req.Population < 0 || req.Population > 999 {
		return nil, errors.New("人数错误")
	}

	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillInfo(req.SaleBillUuid, constant.OptionalUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 点餐助手，拆单后不可以修改人数
	if ctx.GetSource() == constant.SourceAssistant && billInfo.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	oldMealNum := billInfo.MealNum

	// 判断订单状态
	if err := billInfo.ValidateOrderStatus(ctx.GetSource(), constant.OrderUpdateMealNum, 0); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 如果发生恐慌，回滚事务
		}
	}()

	// 修改订单商品人数
	if err := orderRepo.ChangePopulation(req.SaleBillUuid, req.Population); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"修改桌台就餐人数"事件
	utils.Go(func() {
		event.NewSystemBus().PublishChangeMealNumSaleBillEvent(event.ChangeMealNumSaleBillPayload{
			BasePayload: event.BasePayload{ // 修改桌台人数
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: req.SaleBillUuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
			OldMealNum: oldMealNum,
			NewMealNum: uint(req.Population),
		})
	})

	// 推送桌台更新
	utils.Go(func() {
		websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_DESK, map[string]interface{}{
			"desk_uuid":   billInfo.DeskUuid,
			"update_time": time.Now().Unix(),
		})
	})

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// InstantOrderSaleOrderCreate 给销售账单创建一个销售订单。（创建新拆单）
func (s *orderSrv) InstantOrderSaleOrderCreate(ctx context.Context, req req.InstantOrderSaleOrderCreateReq) (*resp.ShopCart, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 加锁
	saleBillUuid := req.SaleBillUuid
	if ctx.NoLock() {
		s.lock.LockUuid(saleBillUuid)
		defer s.lock.UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	// 最大只能创建10个
	if len(saleBill.SaleOrders) == 10 {
		return nil, errors.New("销售账单最多只能创建10个销售订单")
	}

	// 如果销售账单目前只有一个销售订单，增加一个销售订单后要求撤销订单1的优惠折扣
	// 这是产品的特殊要求，可能后续会改。
	// 撤销订单的优惠折扣
	if len(saleBill.SaleOrders) == 1 {
		saleOrder := saleBill.GetFirstSaleOrder()
		// 撤销订单1的优惠折扣
		if saleOrder.IsManualDiscount(uint8(saleBill.SaleBillSetting.ZeroRule)) {
			saleOrder.SetAllDiscountCancel()
		}
		// 撤销订单1的会员折扣
		saleOrder.SetMemberDiscountCancel()
	}

	// 计算并保存销售账单
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 生成订单编号
	var orderSourceType string
	if saleBill.IsDeskSaleBill() {
		orderSourceType = constant.OrderSourceDesk
	} else {
		orderSourceType = constant.OrderSourceInstant
	}
	orderNo, err := s.createOrderNo(db, orderSourceType)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 设置拆单
	saleBill.SetIsSplitOrder(true)

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 创建销售订单
		if _, errCreateSaleOrder := createSaleOrder(ctx, db, saleBill.SaleBillSetting, saleBill.Uuid, orderNo); errCreateSaleOrder != nil {
			return errors.WithMessage(errCreateSaleOrder, fmt.Sprintf("新建拆单失败,saleBill.Uuid:%v, orderNo:%v", saleBill.Uuid, orderNo))
		}

		// 计算并保存销售账单
		if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
			return errors.WithMessage(err)
		}

		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	cartInfo, errCartInfo := s.GetOrderCartInfo(ctx, saleBillUuid)
	if errCartInfo != nil {
		ctx.Log().Error("查询购物车信息失败", zap.Any("errCartInfo", errCartInfo))
		return nil, errors.WithMessage(errCartInfo, "查询购物车信息失败")
	}

	var orders []event.Order
	for i, order := range cartInfo.SaleOrderList {
		orders = append(orders, event.Order{
			SaleOrderUuid: order.Uuid,
			OrderName:     fmt.Sprintf("%d", i+1),
			Amount:        order.AmountInfo.Amount,
		})
	}

	// 发布"拆单"操作事件
	utils.Go(func() {
		s.bus.PublishSplitOrderEvent(event.SplitOrderPayload{
			BasePayload: event.BasePayload{ // 拆单
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: saleBill.Uuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
			Orders: orders,
		})
	})

	return cartInfo, nil
}

// SaleOrderMoveProduct 从一个销售订单移动商品到另一个销售订单
// 第一种移动方式：原销售订单商品数量大于移动数量，则原销售订单商品数量减少移动数量，目标销售订单中有签名一样的商品，该商品数量增加移动数量
// 第二种移动方式：原销售订单商品数量小于移动数量，则原销售订单商品数量减少移动数量，目标销售订单中没有签名一样的商品，则新建一个销售订单商品，该商品数量为移动数量
// 第三种移动方式：原销售订单商品数量等于移动数量，则原销售订单商品从原销售订单中移除，目标销售订单中有签名一样的商品，该商品数量增加移动数量
// 第四种移动方式：原销售订单商品数量等于移动数量，则原销售订单商品从原销售订单中移除，目标销售订单中没有签名一样的商品，则新建一个销售订单商品，该商品数量为移动数量
// 数据处理：
// 第一种移动方式：修改原销售订单商品数量，更新记录，重新计算订单金额；修改目标销售订单商品数量，更新记录，重新计算订单金额
// 第二种移动方式：修改原销售订单商品数量，更新记录，重新计算订单金额；新建目标销售订单商品，计算金额，表插入记录，数组增加这条记录，计算订单金额
// 第三种移动方式：删除原销售订单商品，更新表记录，重新计算原订单金额；修改目标销售订单商品数量，更新记录，重新计算订单金额
// 第四种移动方式：修改原销售订单商品的销售订单uuid为目标销售订单的uuid，使用目标销售订单的折扣优惠，更新记录，重新计算原订单金额；目标销售订单的商品数组增加这条记录，重新计算订单金额
func (s *orderSrv) SaleOrderMoveProduct(ctx context.Context, req req.InstantOrderSaleOrderMoveProductReq, needDeleteSaleOrder bool) (*resp.ShopCart, error) {
	saleBillUuid := req.SaleBillUuid
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(saleBillUuid)
		defer s.lock.UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	// 获取销售订单信息
	saleOrderFrom := saleBill.GetSaleOrder(req.From)
	saleOrderTo := saleBill.GetSaleOrder(req.To)

	// 构建移动到订单商品的map结构
	moveProductMap := make(map[uint64]float64)
	for _, moveProduct := range req.Products {
		moveProductMap[moveProduct.Uuid] = moveProduct.Num
	}

	saleOrderProducts, saleOrderBuffetCustomers, buffetDelayProducts, err := s.getMoveProductInfo(ctx, saleOrderFrom, req)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	moveNumMap := make(map[uint64]float64)
	for _, moveProduct := range req.Products {
		moveNumMap[moveProduct.Uuid] = moveProduct.Num
	}

	waitUpdateSaleOrderProductMap, waitCreateSaleOrderProductMap, err := s.moveSaleOrderProduct(ctx, saleBill, saleOrderFrom, saleOrderTo, saleOrderProducts, moveNumMap)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	waitUpdateBuffetCustomerMap, waitCreateBuffetCustomerMap, err := s.moveBuffetCustomer(ctx, saleBill, saleOrderFrom, saleOrderTo, saleOrderBuffetCustomers, moveNumMap)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	waitUpdateBuffetDelayProductMap, waitCreateBuffetDelayProductMap, err := s.moveBuffetDelayProduct(ctx, saleBill, saleOrderFrom, saleOrderTo, buffetDelayProducts, moveNumMap)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	if len(waitUpdateSaleOrderProductMap) == 0 && len(waitUpdateBuffetCustomerMap) == 0 && len(waitUpdateBuffetDelayProductMap) == 0 {
		ctx.Log().Debug("移动商品失败，没有需要更新的销售订单商品")
		return nil, errors.WithMessage(errors.New("移动商品失败"))
	}

	// 计算订单金额
	ctx.Log().Debug("移动商品前,销售订单信息", zap.Any("saleOrderTo calc", saleOrderTo.BeforeCalc()))
	afterSaleOrderCalc := saleOrderTo.CalcSaleOrder(*saleBill.SaleBillSetting)
	ctx.Log().Debug("移动商品后,销售订单信息", zap.Any("saleOrderTo calc", afterSaleOrderCalc))

	ctx.Log().Debug("移动商品前,销售订单信息", zap.Any("saleOrderFrom calc", saleOrderFrom.BeforeCalc()))
	afterSaleOrderFromCalc := saleOrderFrom.CalcSaleOrder(*saleBill.SaleBillSetting)
	ctx.Log().Debug("移动商品后,销售订单信息", zap.Any("saleOrderFrom calc", afterSaleOrderFromCalc))
	// 计算账单金额
	saleBill.CalcSaleBill()

	var cartInfo *resp.ShopCart
	errUpdateDB := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		ctx.Log().Debug("更新销售订单商品", zap.Any("waitUpdateSaleOrderProductMap len", len(waitUpdateSaleOrderProductMap)))
		for _, saleOrderProduct := range waitUpdateSaleOrderProductMap {
			ctx.Log().Debug("更新销售订单商品", zap.Any("saleOrderProduct saleOrder uuid", saleOrderProduct.SaleOrderUuid), zap.Any("saleOrderProduct uuid", saleOrderProduct.Uuid), zap.Any("saleOrderProduct", saleOrderProduct.MultiLanguageName.GetNameByLang(ctx.GetLanguage())))

			if err := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProductRecord(*saleOrderProduct); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 创建销售订单商品及BOM、属性
		for _, saleOrderProduct := range waitCreateSaleOrderProductMap {
			ctx.Log().Debug("新建销售订单商品", zap.Any("saleOrderProduct saleOrder uuid", saleOrderProduct.SaleOrderUuid), zap.Any("saleOrderProduct uuid", saleOrderProduct.Uuid), zap.Any("saleOrderProduct", saleOrderProduct.MultiLanguageName.GetNameByLang(ctx.GetLanguage())))
			if _, err := repository.NewSaleOrderProductRepo(tx).CreateSaleOrderProductAndBomAndAttribute(*saleOrderProduct); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 更新自助餐顾客
		for _, buffetCustomer := range waitUpdateBuffetCustomerMap {
			if err := repository.NewSaleOrderBuffetCustomerTypeRepo(tx).UpdateSaleOrderBuffetCustomerTypeRecord(*buffetCustomer); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 创建自助餐顾客
		for _, buffetCustomer := range waitCreateBuffetCustomerMap {
			if err := repository.NewSaleOrderBuffetCustomerTypeRepo(tx).CreateSaleOrderBuffetCustomerTypeRecord(*buffetCustomer); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 更新自助餐加钟商品
		for _, buffetDelayProduct := range waitUpdateBuffetDelayProductMap {
			if err := repository.NewOrderRepo(tx).UpdateSaleOrderBuffetDelayProductRecord(*buffetDelayProduct); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 创建自助餐加钟商品
		for _, buffetDelayProduct := range waitCreateBuffetDelayProductMap {
			if _, err := repository.NewOrderRepo(tx).CreateSaleOrderBuffetDelayProduct(*buffetDelayProduct); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 当删除拆单时. needDeleteSaleOrder使用场景：1.删除某个子单，移动完商品后，需要删除该子单；2.撤销拆单，移动完商品后，需要删除所有子单
		if needDeleteSaleOrder {
			if err := repository.NewSaleOrderRepo(tx).UpdateSaleOrderSoftDeleteByUuid(saleOrderFrom.Uuid); err != nil {
				return errors.WithMessage(err)
			}
		} else {
			if err := repository.NewSaleOrderRepo(tx).UpdateSaleOrderRecord(*saleOrderFrom); err != nil {
				return errors.WithMessage(err)
			}
		}
		if err := repository.NewSaleOrderRepo(tx).UpdateSaleOrderRecord(*saleOrderTo); err != nil {
			return errors.WithMessage(err)
		}

		// 更新账单
		if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
			return errUpdateSaleBill
		}
		return nil
	})
	if errUpdateDB != nil {
		return nil, errors.WithMessage(errUpdateDB, "更新数据失败")
	}
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	cartInfo = info

	return cartInfo, nil
}

// InstantOrderSaleOrderDelete 删除一个销售订单(删除拆单)
func (s *orderSrv) InstantOrderSaleOrderDelete(ctx context.Context, request req.InstantOrderSaleOrderDeleteReq) (*resp.ShopCart, error) {
	var shopCart *resp.ShopCart
	if ctx.NoLock() {
		// 加锁
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}
	ctx.Log().Debug("删除一个销售订单(删除拆单)", zap.Any("request", request))
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		ctx.Log().Error("获取销售账单信息失败", zap.Error(errSaleBill))
		return nil, errors.New("获取销售账单信息失败")
	}

	// 不能删除第一个销售订单
	if len(saleBill.SaleOrders) > 0 {
		if saleBill.SaleOrders[0].Uuid == request.SaleOrderUuid {
			return nil, errors.New("不能删除第一个销售订单")
		}
	}

	firstSaleOrder := saleBill.GetSaleOrder(saleBill.SaleOrders[0].Uuid)

	saleOrderFrom := saleBill.GetSaleOrder(request.SaleOrderUuid)

	// 获取要移动的商品列表
	moveProductList := s.getMoveProductList(saleOrderFrom)

	// 如果第一个销售订单已经结账且要删除的订单还有商品的话，则提示"拆单1已结账，请结账当前拆单或删除商品后再删除拆单"。已送厨的商品也要先退菜再删除后才能删除拆单
	if firstSaleOrder.IsSettled() && len(moveProductList) > 0 {
		return nil, errors.New("拆单1已结账，请结账当前拆单或删除商品后再删除拆单")
	}

	// 如果第一个销售订单已经结账且要删除的订单没有商品且销售订单数量大于2时，则删除该拆单
	if firstSaleOrder.IsSettled() && len(moveProductList) == 0 && len(saleBill.SaleOrders) > 2 {
		// 如果销售订单中没有商品，则直接删除订单
		if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderSoftDeleteByUuid(saleOrderFrom.Uuid); err != nil {
			ctx.Log().Error("删除订单失败", zap.Error(err))
			return nil, errors.New("删除订单失败")
		}

		var err error
		shopCart, err = s.GetOrderCartInfo(ctx, request.SaleBillUuid, repository.FilterEndStatus())
		if err != nil {
			ctx.Log().Error("获取购物车信息失败", zap.Error(err))
			return nil, errors.WithMessage(err, "获取购物车信息失败")
		}
	}

	// 如果要删除的订单没有商品，且删除后剩余订单全部已结账，则删除该拆单并完成该销售账单
	if len(moveProductList) == 0 && saleBill.ShouldFinishBillAfterDelete(saleOrderFrom.Uuid) {
		// 如果销售订单中没有商品，且剩余订单全部已结账，则直接删除订单并完成账单
		saleOrderFrom.SetDelete()
		if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
			if err := repository.NewSaleOrderRepo(tx).UpdateSaleOrderSoftDeleteByUuid(saleOrderFrom.Uuid); err != nil {
				ctx.Log().Error("删除订单失败", zap.Error(err))
				return errors.New("删除订单失败")
			}

			// 获取门店业务设置
			businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
			if err != nil {
				return errors.WithMessage(err)
			}

			// 更新销售账单. 如果可以结束销售账单的话
			if err := s.FinishSaleBill(ctx, saleBill, businessSetting, tx); err != nil {
				return errors.WithMessage(err)
			}
			return nil
		}); err != nil {
			return nil, errors.WithMessage(err)
		}
		var err error
		shopCart, err = s.GetOrderCartInfo(ctx, request.SaleBillUuid, repository.FilterEndStatus())
		if err != nil {
			ctx.Log().Error("获取购物车信息失败", zap.Error(err))
			return nil, errors.WithMessage(err, "获取购物车信息失败")
		}
	}

	// 如果销售订单中有商品，则先移动商品到第一个销售订单再删除该子单
	if !firstSaleOrder.IsSettled() && len(moveProductList) > 0 {
		moveProductReq := req.InstantOrderSaleOrderMoveProductReq{
			SaleBillUuid: request.SaleBillUuid,
			From:         request.SaleOrderUuid,
			To:           firstSaleOrder.Uuid,
			Products:     moveProductList,
		}
		var err error
		shopCart, err = s.SaleOrderMoveProduct(ctx, moveProductReq, true)
		if err != nil {
			ctx.Log().Error("移动商品失败", zap.Error(err))
			return nil, errors.WithMessage(err)
		}
	}

	if !firstSaleOrder.IsSettled() && len(moveProductList) == 0 {
		// 如果销售订单中没有商品，则直接删除订单
		if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderSoftDeleteByUuid(saleOrderFrom.Uuid); err != nil {
			ctx.Log().Error("删除订单失败", zap.Error(err))
			return nil, errors.New("删除订单失败")
		}

		var err error
		shopCart, err = s.GetOrderCartInfo(ctx, request.SaleBillUuid, repository.FilterEndStatus())
		if err != nil {
			ctx.Log().Error("获取购物车信息失败", zap.Error(err))
			return nil, errors.WithMessage(err, "获取购物车信息失败")
		}
	}

	// 更新销售账单
	saleBill.SetIsSplitOrder(len(saleBill.SaleOrders)-1 > 1)
	if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
		return nil, errors.WithMessage(errUpdateSaleBill)
	}

	// 发布"拆单"操作事件
	utils.Go(func() {
		var orders []event.Order
		for i, order := range shopCart.SaleOrderList {
			orders = append(orders, event.Order{
				SaleOrderUuid: order.Uuid,
				OrderName:     fmt.Sprintf("%d", i+1),
				Amount:        order.AmountInfo.Amount,
			})
		}
		if len(orders) == 1 {
			// 发布"撤销拆单"操作事件
			s.bus.PublishCancelSplitOrderEvent(event.CancelSplitOrderPayload{
				BasePayload: event.BasePayload{ // 撤销拆单
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: saleBill.Uuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
			})
		} else {
			// 发布"拆单"操作事件
			s.bus.PublishSplitOrderEvent(event.SplitOrderPayload{
				BasePayload: event.BasePayload{ // 拆单
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: saleBill.Uuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
				Orders: orders,
			})
		}
	})

	return shopCart, nil

}

// InstantOrderSaleOrderDeleteAll 删除所有子销售订单(撤销拆单)
func (s *orderSrv) InstantOrderSaleOrderDeleteAll(ctx context.Context, request req.InstantOrderSaleOrderDeleteAllReq) (*resp.ShopCart, error) {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		ctx.Log().Error("获取销售账单信息失败", zap.Error(errSaleBill))
		return nil, errors.WithMessage(errSaleBill, "获取销售账单信息失败")
	}

	firstSaleOrder := saleBill.GetSaleOrder(saleBill.SaleOrders[0].Uuid)

	// 如果第一个销售订单已经结账，则提示"当前订单已结账，无法撤销"
	if firstSaleOrder.IsSettled() {
		return nil, errors.New("当前订单已结账，无法撤销")
	}

	// 判断订单是否已被部分支付
	if repository.NewOrderRepo(db).IsPartiallyPaid(request.SaleBillUuid) {
		return nil, errors.New("当前订单已被部分支付，不支持撤销拆单")
	}

	saleOrderFromList := make([]*model.SaleOrder, 0)
	for _, saleOrder := range saleBill.SaleOrders {
		if saleOrder.Uuid == firstSaleOrder.Uuid {
			continue
		}
		saleOrderFromList = append(saleOrderFromList, saleOrder)
	}

	for _, saleOrderFrom := range saleOrderFromList {
		moveProductList := s.getMoveProductList(saleOrderFrom)
		moveProductReq := req.InstantOrderSaleOrderMoveProductReq{
			SaleBillUuid: request.SaleBillUuid,
			From:         saleOrderFrom.Uuid,
			To:           firstSaleOrder.Uuid,
			Products:     moveProductList,
		}

		if len(moveProductList) > 0 {
			// NOTE 优化减少重复查询
			_, err := s.SaleOrderMoveProduct(ctx, moveProductReq, true)
			if err != nil {
				ctx.Log().Error("移动商品失败", zap.Error(err))
				return nil, errors.WithMessage(err)
			}
		} else {
			// 如果销售订单中没有商品，则直接删除订单
			if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderSoftDeleteByUuid(saleOrderFrom.Uuid); err != nil {
				ctx.Log().Error("删除订单失败", zap.Error(err))
				return nil, errors.WithMessage(err, "删除订单失败")
			}
		}
	}

	// 发布"撤销拆单"操作事件
	utils.Go(func() {
		s.bus.PublishCancelSplitOrderEvent(event.CancelSplitOrderPayload{
			BasePayload: event.BasePayload{ // 撤销拆单
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: saleBill.Uuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
	})

	// 获取账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}
	// 获取销售订单信息
	saleOrder := saleBill.GetSaleOrder(saleBill.SaleOrders[0].Uuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}
	if request.MemberUuid > 0 {
		// 设置会员折扣
		member, errMember := repository.NewMemberRepo(db).GetMemberInfoForSaleOrder(ctx, request.MemberUuid)
		if errMember != nil {
			return nil, errors.WithMessage(errMember)
		}
		saleOrder.SetMemberDiscount(*member)
	}
	saleOrder.SetAllDiscountCancel()
	// 更新销售账单是否拆单的字段
	saleBill.SetIsSplitOrder(len(saleBill.SaleOrders)-1 > 1)

	// 重新计算销售订单金额
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err, "s.CalcAndSaveSaleBill failed")
	}

	info, err := s.GetOrderCartInfo(ctx, request.SaleBillUuid)
	if err != nil {
		ctx.Log().Error("获取购物车信息失败", zap.Error(err))
		return nil, errors.WithMessage(err, "获取购物车信息失败")
	}

	return info, nil
}

// OrderUnlock 订单解锁
func (s *orderSrv) OrderUnlock(ctx context.Context, saleBillUuid uint64) error {
	// 加锁
	if ctx.NoLock() {
		s.lock.LockUuid(saleBillUuid)
		defer s.lock.UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return errors.WithMessage(err, "查询销售账单失败")
	}

	// 验证订单是否可操作
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderUnlock, 0); err != nil {
		return errors.WithMessage(err)
	}

	// 验证销售账单是否已锁定
	if !saleBill.IsLockStatus() {
		return errors.New("销售账单未锁定")
	}

	// 保存销售账单
	if err := repository.NewOrderRepo(db).SetLock(saleBill.Uuid, false); err != nil {
		return errors.WithMessage(err, "解锁失败")
	}

	// 推送桌台更新
	utils.Go(func() {
		websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_DESK, map[string]interface{}{
			"desk_uuid":   saleBill.DeskUuid,
			"update_time": time.Now().Unix(),
		})
	})

	return nil
}

// GetSaleBillByDeskId  获取桌台账单信息
func (s *orderSrv) GetSaleBillByDeskId(ctx context.Context) (model.SaleBill, error) {
	dbId := ctx.GetDbId()
	deskUuid := ctx.GetDeskUuid()

	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))

	// 通过桌台查找到当前桌台的正在进行销售账单
	billInfo, err := orderRepo.GetSaleBillInfoByDesk(deskUuid, constant.OptionalUuid)
	if err != nil {
		return model.SaleBill{}, errors.WithMessage(err)
	}
	return billInfo, nil
}
func (s *orderSrv) GetOrderCartInfoByDeviceSn(ctx context.Context, deviceSn string) (*resp.ShopCart, error) {
	// 通过deviceSn获取saleBillUuid
	saleBillUuid, errUuid := s.getSaleBillUuidByDeviceSn(ctx)
	if errUuid != nil {
		return nil, errors.WithMessage(errUuid)
	}
	// 没有找到销售账单
	if saleBillUuid == 0 {
		// 收银机点餐页面没有销售账单时，检查是否有自动加购的必点方案，如果有，则创建一个销售账单并自动加购商品
		res, err := s.InstantOrderMustPlan2(ctx, deviceSn)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		if res == nil {
			return nil, nil
		}
		if res.BatchCookingMode == "" {
			// 获取门店业务设置
			businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			// 没开启分批时，才返回 BatchCookingMode
			if businessSetting.OpenIsBatch() {
				res.BatchCookingMode = businessSetting.BatchCookingMode
			}
		}
		return res, nil
	}
	// 查询购物车信息
	cartInfo, errInfo := s.GetOrderCartInfo(ctx, saleBillUuid)
	if errInfo != nil {
		return nil, errInfo
	}
	return cartInfo, nil
}

// GetSaleBillUuidAndSaleOrderUuid 获取销售账单uuid和销售订单uuid
func (s *orderSrv) GetSaleBillUuidAndSaleOrderUuid(ctx context.Context, deskUuid uint64) (uint64, uint64, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	// 获取桌台信息
	saleBillUuid, saleOrderUuid, err := repository.NewDeskRepo(db).GetSaleBillUuidAndSaleOrderUuid(deskUuid)
	if err != nil {
		return 0, 0, errors.WithMessage(err)
	}

	return saleBillUuid, saleOrderUuid, nil
}

// GetProductPackageDetail 获取商品包详情
func (s *orderSrv) GetProductPackageDetail(ctx context.Context, req req.GetProductPackageDetailReq) (*resp.ProductPackageDetailRes, error) {
	db := ctx.GetDB()
	// 获取销售订单中h5未下单的销售订单商品
	saleOrderProducts, err := repository.NewSaleOrderProductRepo(db).GetProductPackageDetail(req.SaleBillUuid, req.SaleOrderUuid, req.ProductPackageUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取商品包详情失败")
	}

	productPackageDetailList := make([]resp.ProductPackageDetail, 0)

	for _, saleOrderProduct := range saleOrderProducts {
		// 过滤掉已经送厨的商品
		if saleOrderProduct.IsCookingProduct() {
			continue
		}
		productPackageDetail := saleOrderProduct.GetProductPackageDetail()
		productPackageDetailList = append(productPackageDetailList, productPackageDetail)
	}

	return &resp.ProductPackageDetailRes{List: productPackageDetailList}, nil
}
