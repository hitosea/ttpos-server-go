package service

import (
	builtinerrors "errors"
	"sort"
	"strconv"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// GetUnOrderedH5ProductList 获取扫码h5购物车未下单商品列表
func (s *orderSrv) GetUnOrderedH5ProductList(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart, opts ...repository.OrderCartInfoOptionFunc) (*resp.UnsentKitchen, error) {
	res, err := s.GetUnsentKitchen(ctx, saleBillUuid, shopCart, nil, opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 重新计算商品金额。商品金额=商品列表中各个商品的金额之和
	productAmount := decimal.NewFromFloat(0)
	for _, product := range res.Products.List {
		productAmount = productAmount.Add(decimal.NewFromFloat(product.GetPrice()))
	}
	res.AmountInfo.ProductAmount = productAmount.InexactFloat64()

	return &res, nil
}

// GetOrderedH5ProductList 获取扫码h5购物车已下单商品列表
func (s *orderSrv) GetOrderedH5ProductList(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart, opts ...repository.OrderCartInfoOptionFunc) (*resp.H5CartSendProduct, error) {
	if shopCart == nil {
		var err error
		shopCart, err = s.GetOrderCartInfo(ctx, saleBillUuid, opts...)
		if err != nil {
			return nil, errors.WithMessage(errors.New("获取点餐购物车信息"), err.Error())
		}
	}
	productGroup := make(map[int64][]resp.Product)
	for _, saleOrder := range shopCart.SaleOrderList {
		for _, product := range saleOrder.ProductList {
			productGroup[product.SendKitchenTime] = append(productGroup[product.SendKitchenTime], product)
		}
	}
	groups := make([]resp.H5Group, 0)
	for sendKitchenTime, products := range productGroup {
		groups = append(groups, resp.H5Group{
			SentKitchenProductGroup: resp.SentKitchenProductGroup{
				SendKitchenTime: sendKitchenTime,
				Products: resp.GroupProductList{
					List: products,
				},
			},
			AcceptTime:         products[0].AcceptTime,
			IsAccept:           products[0].IsAccept,
			H5OrderTime:        products[0].H5OrderTime,
			IsH5OrderNeedAudit: products[0].IsH5OrderNeedAudit,
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].SendKitchenTime > groups[j].SendKitchenTime
	})

	return &resp.H5CartSendProduct{
		Groups: resp.H5GroupList{List: groups},
	}, nil
}

// ConfirmH5Order 下单扫码h5订单
func (s *orderSrv) ConfirmH5Order(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64, ignoreMust bool) (any, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(saleBillUuid)
		defer s.lock.UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	res := make(map[string]any)
	// 获取销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return res, errors.WithMessage(err, "查询销售账单失败")
	}
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderH5Confirm, saleOrderUuid); err != nil {
		return res, errors.WithMessage(err)
	}

	// h5下单限制
	h5Setting, _ := s.settingSrv.GetH5Setting(ctx, []dto.LanguageItem{})
	if h5Setting.IsBuffetOrderLimit == "1" || h5Setting.IsOrderLimit == "1" {
		// 读取上一单下单时间
		h5Repo := repository.NewH5OrderRepo(db)
		lastH5Order, _ := h5Repo.GetH5Order(h5Repo.WhereSaleBillUuid(saleBillUuid))
		var lastOrderTime int64 = 0
		if lastH5Order != nil {
			lastOrderTime = lastH5Order.CreateTime
		}
		if h5Setting.IsBuffetOrderLimit == "1" && saleBill.IsBuffetSaleBill() { // 自助餐下单限制
			if h5Setting.BuffetOrderLimit.IsLimitTime == "1" && lastOrderTime > 0 { // 限制下单间隔
				interval, err := strconv.Atoi(h5Setting.BuffetOrderLimit.LimitTime)
				if err != nil {
					return res, errors.WithMessage(err, "解析H5设置失败")
				}
				// 小于间隔时间，不可下单
				nextTime := time.Unix(lastOrderTime, 0).Add(time.Duration(interval) * time.Minute).Unix()
				now := time.Now().Unix()
				if nextTime-now > 0 {
					return gin.H{"value": nextTime - now}, errors.NewWithCode(constant.CodeH5OrderTimeLimit, "时间限制")
				}
			}
			if h5Setting.BuffetOrderLimit.IsLimitNum == "1" { // 限制下单最大商品总数
				numLimit, err := strconv.Atoi(h5Setting.BuffetOrderLimit.LimitNum)
				if err != nil {
					return res, errors.WithMessage(err, "解析H5设置失败")
				}
				if saleBill.GetUnOrderH5OrderProductNum() > float64(numLimit) {
					return gin.H{"value": numLimit}, errors.NewWithCode(constant.CodeH5OrderNumLimit, "数量限制")
				}
			}
		}
		if h5Setting.IsOrderLimit == "1" && !saleBill.IsBuffetSaleBill() { // 非自助餐下单限制
			if h5Setting.OrderLimit.IsLimitTime == "1" && lastOrderTime > 0 { // 限制下单间隔
				interval, err := strconv.Atoi(h5Setting.OrderLimit.LimitTime)
				if err != nil {
					return res, errors.WithMessage(err, "解析H5设置失败")
				}
				// 小于间隔时间，不可下单
				nextTime := time.Unix(lastOrderTime, 0).Add(time.Duration(interval) * time.Minute).Unix()
				now := time.Now().Unix()
				if nextTime-now > 0 {
					return gin.H{"value": nextTime - now}, errors.NewWithCode(constant.CodeH5OrderTimeLimit, "时间限制")
				}
			}
			if h5Setting.OrderLimit.IsLimitNum == "1" { // 限制下单最大商品总数
				numLimit, err := strconv.Atoi(h5Setting.OrderLimit.LimitNum)
				if err != nil {
					return res, errors.WithMessage(err, "解析H5设置失败")
				}
				if saleBill.GetUnOrderH5OrderProductNum() > float64(numLimit) {
					return gin.H{"value": numLimit}, errors.NewWithCode(constant.CodeH5OrderNumLimit, "数量限制")
				}
			}
		}
	}

	h5Order, err := saleBill.NewH5Order(ctx.GetCompanySetting())
	if err != nil {
		return res, errors.WithMessage(err)
	}
	// 获取未下单的h5订单商品
	h5OrderProducts := saleBill.GetUnOrderH5OrderProduct()
	subProductList := make([]*model.SaleOrderProduct, 0)
	// 将未下单的h5订单商品变为已下单的h5订单商品
	for _, h5OrderProduct := range h5OrderProducts {
		h5OrderProduct.SetH5OrderProduct(h5Order.Uuid)
		// 如果商品是套餐商品，则设置套餐子商品变为已下单的h5订单商品
		if h5OrderProduct.IsPackageProduct() {
			subProducts := saleBill.GetSubProducts(h5OrderProduct.Uuid)
			for _, subProduct := range subProducts {
				subProduct.SetH5OrderProduct(h5Order.Uuid)
			}
			subProductList = append(subProductList, subProducts...)
		}
	}

	// 检查超时不能加购
	if err := s.checkTimeoutAndCannotAddPurchase(ctx, saleBill, h5OrderProducts); err != nil {
		return res, errors.WithMessage(err)
	}

	// 只检查本次下单的商品是否超过限购
	uuids := make([]uint64, 0)
	for _, h5OrderProduct := range h5OrderProducts {
		uuids = append(uuids, h5OrderProduct.Uuid)
	}

	// 检查必点
	saleOrderProductAll := saleBill.GetSaleOrderProductAll(model.WithH5CheckLimit())
	checkServiceRes, errCheck := s.checkOrder(ctx, ignoreMust, db, saleBill.Uuid, saleBill.DeskUuid, saleOrderProductAll, WithCheckTypeCooking(), WithIsH5Check(), WithSaleOrderProductUuid(uuids...))
	if errCheck != nil {
		ctx.Log().Error("检查商品失败", zap.Error(errCheck))
		return nil, errors.New("检查商品失败")
	}
	if checkServiceRes != nil {
		if checkServiceRes.Code == constant.CodeOrderCheckProductMust && ignoreMust {
			// 必点方案未选择，且忽略必点方案
		} else {
			return checkServiceRes.OrderCheckRes, errors.WithMessage(errors.New(constant.ParseCodeOrderCheck(checkServiceRes.Code, constant.WithIsH5())))
		}
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 创建h5订单
		if _, err := repository.NewH5OrderRepo(tx).CreateH5Order(*h5Order); err != nil {
			return errors.WithMessage(err, "保存h5订单失败")
		}
		// 创建h5订单商品
		for _, h5OrderProduct := range h5Order.H5OrderProducts {
			h5OrderProduct.H5OrderUuid = h5Order.Uuid
			if _, err := repository.NewH5OrderRepo(tx).CreateH5OrderProduct(*h5OrderProduct); err != nil {
				return errors.WithMessage(err, "保存h5订单商品失败")
			}
		}
		// 更新销售订单商品，将未下单商品改为已下单商品。记录上saleOrderProduct的h5OrderUuid
		for _, saleOrderProduct := range h5OrderProducts {
			if err := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProductRecord(*saleOrderProduct); err != nil {
				return errors.WithMessage(err, "更新销售订单商品失败")
			}
		}
		// 更新销售订单商品，将套餐子商品改为已下单商品。记录上saleOrderProduct的h5OrderUuid
		for _, saleOrderProduct := range subProductList {
			if err := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProductRecord(*saleOrderProduct); err != nil {
				return errors.WithMessage(err, "更新销售订单商品失败")
			}
		}

		// 关闭h5接单功能时,自动接单
		{
			companySetting := ctx.GetCompanySetting()
			if !companySetting.GetIsOpenH5Order() {
				// 如果关闭h5接单功能
				// 自动接单
				ctx.SetDB(tx) // 使用当前事务
				if result, err := s.AcceptH5Order(ctx, h5Order.Uuid, true); err != nil {
					ctx.Log().Error("关闭h5接单功能，自动接单失败", zap.Error(err))
					return errors.WithMessage(err, "关闭h5接单功能，自动接单失败/异常")
				} else if result != nil {
					ctx.Log().Error("关闭h5接单功能，自动接单异常", zap.Any("result", result))
					return errors.WithMessage(errors.New("关闭h5接单功能，自动接单异常"), "关闭h5接单功能，自动接单失败/异常")
				}
				ctx.SetDB(nil) // 清空db,避免影响后续的逻辑
			}
		}
		return nil
	}); err != nil {
		return res, errors.WithMessage(err, "下单扫码h5订单失败")
	}

	// 自动接单
	{
		companySetting := ctx.GetCompanySetting()
		if companySetting.GetIsOpenH5Order() {
			// 判断
			acceptOrderSetting, err := s.settingSrv.GetAcceptOrderSetting(ctx)
			if err != nil {
				ctx.Log().Error("获取接单设置失败", zap.Error(err))
			}
			totalPrice := saleBill.GetUnAcceptH5OrderProductTotalPrice(h5OrderProducts) // 未接单的h5订单商品的商品金额之和
			if acceptOrderSetting.CanAutoOrder(totalPrice) {
				// 自动接单
				if result, err := s.AcceptH5Order(ctx, h5Order.Uuid, true); err != nil {
					ctx.Log().Error("自动接单失败", zap.Error(err))
				} else if result != nil {
					ctx.Log().Error("自动接单异常", zap.Any("result", result))
				}
			}
		}
	}

	return res, nil
}
func (s *orderSrv) AcceptH5Order(ctx context.Context, h5OrderUuid uint64, isAutoOrder bool) (*resp.OrderCheckServiceRes, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	if ctx.GetDB() != nil && isAutoOrder { // 关闭h5接单功能时，使用当前事务
		db = ctx.GetDB()
	}
	companySetting := ctx.GetCompanySetting()
	h5OrderRepo := repository.NewH5OrderRepo(db)
	// 获取h5订单
	h5Order, err := h5OrderRepo.GetH5OrderDetail(h5OrderUuid, companySetting.GetIsOpenH5Order())
	if err != nil {
		return nil, errors.WithMessage(errors.New("获取h5订单失败"), err.Error())
	}
	if h5Order.Status != constant.H5OrderStatusOrder {
		return nil, nil
	}

	if ctx.NoLock() {
		s.lock.LockUuid(h5Order.SaleOrder.SaleBillUuid)
		defer s.lock.UnlockUuid(h5Order.SaleOrder.SaleBillUuid)
		ctx.AddLock()
	}

	// 接单,保证h5订单的商品快照信息
	h5Order.Accept(ctx.GetStaffUuid(), ctx.GetLanguage())
	// 将已下单的h5订单商品变为已接单单的h5订单商品
	h5Order.ChangeToAccepted()

	// 标记为是自动接单
	if isAutoOrder {
		h5Order.IsAutoAccept = 1
	}

	{
		ignoreMust := true // 接单，送厨忽略必点方案
		// 获取销售账单信息
		saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(h5Order.SaleOrder.SaleBillUuid)
		if errSaleBill != nil {
			return nil, errors.WithMessage(errSaleBill, "repository.NewOrderRepo(db).GetSaleBillAllInfo")
		}
		ctx.Log().Debug("获取销售账单信息")

		// 获取本次接单的商品列表
		unCookingSaleOrderProducts := h5Order.SaleOrderProducts

		// 检查超时不能加购
		if !isAutoOrder {
			if err := s.checkTimeoutAndCannotAddPurchase(ctx, saleBill, unCookingSaleOrderProducts); err != nil {
				return nil, errors.WithMessage(err)
			}
		}

		// 将h5订单商品插入到销售订单中
		saleOrder := saleBill.GetFirstSaleOrder()
		saleOrder.InsertSaleOrderProduct(unCookingSaleOrderProducts)

		// 送厨
		checkServiceRes, err := s.ActionCooking(ctx, ignoreMust, saleBill, unCookingSaleOrderProducts, h5OrderUuid, isAutoOrder, withCalcAndSaveSaleBill()) // 接单
		if err != nil {
			return nil, errors.WithMessage(err, "ActionCooking")
		}
		if checkServiceRes != nil {
			return checkServiceRes, nil
		}
	}
	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 更新h5订单
		if err := repository.NewH5OrderRepo(db).UpdateH5OrderRecord(*h5Order); err != nil {
			return errors.WithMessage(err, "更新h5订单失败")
		}
		// 更新h5订单商品列表
		for _, h5OrderProduct := range h5Order.H5OrderProducts {
			// 更新h5订单商品
			if err := repository.NewH5OrderRepo(db).UpdateH5OrderProductRecord(*h5OrderProduct); err != nil {
				return errors.WithMessage(err, "更新h5订单商品失败")
			}
		}
		// 更新销售订单商品.将该h5订单的商品变为已接单
		// for _, saleOrderProduct := range h5Order.SaleOrderProducts {
		// 	if err := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProductRecord(*saleOrderProduct); err != nil {
		// 		return errors.WithMessage(err, "将已下单的h5订单商品变为已接单单的h5订单商品失败")
		// 	}
		// }

		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "接单失败")
	}
	return nil, nil
}
func (s *orderSrv) RejectH5Order(ctx context.Context, h5OrderUuid uint64) error {
	db := ctx.GetDB()
	h5OrderRepo := repository.NewH5OrderRepo(db)
	// 获取h5订单
	h5Order, err := h5OrderRepo.GetH5OrderDetail(h5OrderUuid, true)
	if err != nil {
		if builtinerrors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return errors.WithMessage(errors.New("获取h5订单失败"), err.Error())
	}
	// 非待处理状态不可操作
	if h5Order.Status != constant.H5OrderStatusOrder {
		return errors.WithMessage(errors.New("当前状态不可操作"))
	}

	// 拒单,保证h5订单的商品快照信息
	h5Order.Reject(ctx.GetStaffUuid(), ctx.GetLanguage())
	// 删除销售订单商品
	h5Order.DeleteSaleOrderProduct()

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 更新h5订单
		if err := repository.NewH5OrderRepo(db).UpdateH5OrderRecord(*h5Order); err != nil {
			return errors.WithMessage(err, "更新h5订单失败")
		}
		// 更新h5订单商品列表
		for _, h5OrderProduct := range h5Order.H5OrderProducts {
			// 更新h5订单商品
			if err := repository.NewH5OrderRepo(db).UpdateH5OrderProductRecord(*h5OrderProduct); err != nil {
				return errors.WithMessage(err, "更新h5订单商品失败")
			}
		}
		// 删除销售订单商品.将该h5订单的商品删除
		if err := repository.NewSaleOrderProductRepo(db).DeleteSaleOrderProductList(h5Order.SaleOrderProducts); err != nil {
			return errors.WithMessage(err, "删除销售订单商品失败")
		}

		// 发布"拒单"操作事件
		utils.Go(func() {
			s.bus.PublishRejectH5OrderEvent(event.RejectH5OrderPayload{
				BasePayload: event.BasePayload{ // 拒单
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: h5Order.SaleBillUuid,
					H5OrderUuid:  h5OrderUuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
			})
		})
		return nil
	}); err != nil {
		return errors.WithMessage(err, "拒单失败")
	}
	return nil
}

// RejectAllH5Order 拒绝某销售账单的所有未接单h5订单
func (s *orderSrv) RejectAllH5Order(ctx context.Context, saleBillUuid uint64) error {
	db := ctx.GetDB()
	// 获取所有待接单的h5订单
	h5Orders, err := repository.NewH5OrderRepo(db).GetH5OrderListBySaleBillUuid(saleBillUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	for _, h5Order := range h5Orders {
		err := s.RejectH5Order(ctx, h5Order.Uuid)
		if err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}

// RejectAllH5OrderInShop 将商家的所有待接单的h5订单都拒单
func (s *orderSrv) RejectAllH5OrderInShop(ctx context.Context) error {
	db := ctx.GetDB()
	// 查询所有还在进行中的桌台账单
	saleBills, err := repository.NewSaleBillRepo(db).GetDeskSaleBillUnPay()
	if err != nil {
		return errors.WithMessage(err)
	}
	for _, saleBill := range saleBills {
		s.RejectAllH5Order(ctx, saleBill.Uuid)
	}
	return nil
}
