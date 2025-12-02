package service

import (
	"encoding/json"
	builtinerrors "errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"github.com/duke-git/lancet/v2/slice"

	"go.uber.org/zap"

	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// GetOrderLists 获取订单列表
func (s *orderSrv) GetOrderLists(ctx context.Context, req req.OrderListReq) (resp.OrderListPaginationResp, error) {
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(ctx.GetDbId()))
	// 获取列表源数据
	var reqs repository.GetCashierOrderListWithPaginationType
	_ = copier.Copy(&reqs, req)
	lists, total, dbOption, err := orderRepo.GetCashierOrderListWithPagination(reqs, ctx.GetCompanySetting().Timezone)
	if err != nil {
		return resp.OrderListPaginationResp{}, errors.WithMessage(err)
	}

	// 组合列表源数据
	billList := make([]resp.BillLists, len(lists))
	for i, bill := range lists {
		consumerUuids := []string{}
		totalPayTypeNames := []string{}
		isSplit := len(bill.SaleOrders) > 1 // 拆单
		orderList := make([]resp.BillListsOrder, 0)
		var paymentAmounts float64
		//
		billListsExtra := resp.BillListsExtra{
			IsCellRefund:        false,
			IsCellCancel:        bill.Status == constant.SaleBillStatusPending,
			IsCellReverseSettle: bill.IsCellReverseSettle(ctx.GetStaff().Uuid, ctx.GetStaff().CashierLoginTime),
			IsCellPrint:         !isSplit,
			IsCellInvoice:       !isSplit && bill.Status == constant.SaleBillStatusComplete,
			IsCellDelete:        bill.Status == constant.SaleBillStatusCanceled,
			IsCellShow:          bill.DataManage == nil,
		}
		// 拆单
		if isSplit {
			for k, order := range bill.SaleOrders {
				if order.IsDelete() {
					continue
				}
				// 获取支付方式
				payTypeNames := []string{}
				if order.IsFree == 1 {
					totalPayTypeNames = append(totalPayTypeNames, i18n.Translate(ctx.GetLanguage(), "免单"))
					payTypeNames = append(payTypeNames, i18n.Translate(ctx.GetLanguage(), "免单"))
				} else {
					for _, payment := range order.PaymentOrders {
						if payment.IsDelete() {
							continue
						}
						totalPayTypeNames = append(totalPayTypeNames, payment.PaymentMethodName)
						payTypeNames = append(payTypeNames, payment.PaymentMethodName)
					}
				}

				orderExtra := resp.BillListsExtra{
					IsCellRefund:        false,
					IsCellCancel:        false,
					IsCellReverseSettle: false,
					IsCellPrint:         true,
					IsCellInvoice:       order.Status == constant.SaleBillStatusComplete,
					IsCellDelete:        order.Status == constant.SaleBillStatusCanceled,
				}
				// 不等于免单 && 未全退款 && 完成
				if order.IsFree == 0 && order.GetTotalRefundAmount() < order.PaymentAmount && order.Status == constant.SaleBillStatusComplete {
					orderExtra.IsCellRefund = true
				}
				//
				paymentAmount := order.GetActualPaymentAmount()
				paymentAmounts += paymentAmount
				//
				orderList = append(orderList, resp.BillListsOrder{
					SaleBillUuid:  order.SaleBillUuid,
					SaleOrderUuid: order.Uuid,
					BillType:      bill.BillType,
					SerialNo:      bill.SerialNo + "-" + strconv.Itoa(k+1),
					ConsumerUuids: func() string {
						if order.ConsumerUuid == 0 {
							return ""
						}
						return strconv.FormatUint(uint64(order.Member.ID), 10)
					}(),
					OrderNo:       order.OrderNo,
					Status:        order.Status,
					FinishTime:    order.FinishTime,
					OrderAmount:   order.OriginAmount,
					PaymentAmount: paymentAmount,
					PayTypeName:   strings.Join(utils.RemoveDuplicates(payTypeNames), ","),
					Extra:         orderExtra,
				})
				//
				if order.ConsumerUuid > 0 {
					consumerUuids = append(consumerUuids, strconv.FormatUint(uint64(order.Member.ID), 10))
				}
			}
		} else {
			// 没有拆单
			if len(bill.SaleOrders) > 0 {
				order := bill.SaleOrders[0]
				if order.ConsumerUuid > 0 {
					if order.Member == nil {
						logger.Logger.Info("member is nil", zap.Any("order", order))
					}
					consumerUuids = append(consumerUuids, strconv.FormatUint(uint64(order.Member.ID), 10))
				}
				if order.IsFree == 1 {
					totalPayTypeNames = append(totalPayTypeNames, i18n.Translate(ctx.GetLanguage(), "免单"))
				} else {
					for _, payment := range order.PaymentOrders {
						if payment.IsDelete() {
							continue
						}
						totalPayTypeNames = append(totalPayTypeNames, payment.PaymentMethodName)
					}
				}
				// 不等于免单 && 未退款 && 完成
				if order.IsFree == 0 && order.GetTotalRefundAmount() < order.PaymentAmount && order.Status == constant.SaleBillStatusComplete {
					billListsExtra.IsCellRefund = true
				}
				// 等于主单 && 完成 && 等于当前用户 && 在班次时间内
				billListsExtra.IsCellReverseSettle = bill.IsCellReverseSettle(ctx.GetStaff().Uuid, ctx.GetStaff().CashierLoginTime)
			}
		}
		//
		saleOrderUuid := uint64(0)
		if !isSplit && len(bill.SaleOrders) > 0 {
			saleOrderUuid = bill.SaleOrders[0].Uuid
		}
		//
		billList[i] = resp.BillLists{
			SaleBillUuid:  bill.Uuid,
			SaleOrderUuid: saleOrderUuid,
			BillType:      bill.BillType,
			IsSplit:       len(bill.SaleOrders) > 1,
			SerialNo:      bill.SerialNo,
			OrderNo:       bill.OrderNo,
			Status:        bill.Status,
			FinishTime:    bill.FinishTime,
			OrderAmount:   bill.OriginAmount,
			PaymentAmount: bill.GetPaymentAmount(),
			ConsumerUuids: strings.Join(consumerUuids, ","),
			PayTypeName:   strings.Join(utils.RemoveDuplicates(totalPayTypeNames), ","),
			SaleOrders:    orderList,
			Extra:         billListsExtra,
		}
	}
	// 获取数量
	getOrderNum := func(status uint) int64 {
		opts := []repository.DBOption{
			repository.CommonRepo.WhereByStatus(status),
			repository.CommonRepo.WhereBySoftDelete(),
			repository.CommonRepo.WhereByCooking(),
			repository.CommonRepo.WhereInBillType([]uint{constant.SaleBillTypeDesk, constant.SaleBillTypeInstant}),
			dbOption,
		}
		if reqs.IsOnlyDataManage == 1 {
			uuidList := strings.Split(reqs.SaleBillUuids, ",")
			uuids := []uint64{}
			for _, uuid := range uuidList {
				uuid, _ := strconv.ParseUint(uuid, 10, 64)
				uuids = append(uuids, uint64(uuid))
			}
			opts = append(opts, repository.CommonRepo.WhereInUuids(uuids))
		}
		if reqs.IsOnlyDataManage == 0 && reqs.IsContainDataManage == 0 {
			opts = append(opts,
				func() repository.DBOption {
					return func(db *gorm.DB) *gorm.DB {
						return db.Where("uuid NOT IN (SELECT data_uuid FROM ttpos_data_manage WHERE type = ?)", model.DataManageTypeOrder)
					}
				}(),
			)
		}
		num, _ := orderRepo.GetOrderNum(opts...)
		return num
	}
	// 获取数量
	unpaidNum := getOrderNum(constant.SaleBillStatusPending)
	completeNum := getOrderNum(constant.SaleBillStatusComplete)
	cancelNum := getOrderNum(constant.SaleBillStatusCanceled)
	// 获取实付金额
	paymentAmountDec := decimal.NewFromFloat(0)

	// 返回响应对象
	return resp.OrderListPaginationResp{
		List: billList,
		Meta: resp.OrderListMeta{
			PageResponse: dto.PageResponse{
				PageNo:   req.PageNo,
				PageSize: req.PageSize,
				Total:    total,
			},
			TotalNum:      unpaidNum + completeNum + cancelNum,
			UnpaidNum:     unpaidNum,
			CompleteNum:   completeNum,
			CancelNum:     cancelNum,
			PaymentAmount: paymentAmountDec.Round(2).InexactFloat64(),
		},
	}, nil
}

// ExportOrderLists 导出订单列表
func (s *orderSrv) ExportOrderLists(ctx context.Context, req req.OrderListReq) (resp.OrderExportListPaginationResp, error) {
	language := ctx.GetLanguage()

	// 获取列表源数据
	var reqs repository.GetCashierOrderListWithPaginationType
	_ = copier.Copy(&reqs, req)
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(ctx.GetDbId()))
	lists, total, _, err := orderRepo.GetCashierOrderExportListWithPagination(reqs, ctx.GetCompanySetting().Timezone)
	if err != nil {
		return resp.OrderExportListPaginationResp{}, errors.WithMessage(err)
	}

	statusText := map[uint]string{
		constant.SaleBillStatusPending:  i18n.Translate(language, "待付款"),
		constant.SaleBillStatusComplete: i18n.Translate(language, "已完成"),
		constant.SaleBillStatusCanceled: i18n.Translate(language, "已取消"),
	}

	// 组合列表源数据
	saleBillUuids := []uint64{}
	exportLists := make([]resp.OrderExportInfo, 0)
	for _, bill := range lists {
		saleBillUuids = append(saleBillUuids, bill.Uuid)
		isSplit := len(bill.SaleOrders) > 1
		// 拆单
		for index, saleOrder := range bill.SaleOrders {

			var products []*resp.OrderExportInfoProduct
			// 添加自助餐顾客
			for _, orderBuffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
				if orderBuffetCustomer.IsDelete() {
					continue
				}
				products = append(products, &resp.OrderExportInfoProduct{
					Name:       orderBuffetCustomer.BuffetPackage.MultiLanguageName.GetNameByLang(language),
					AttrName:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					Num:        float64(orderBuffetCustomer.Num),
					TotalPrice: orderBuffetCustomer.GetDiscountPriceWithVAT(),
				})
			}
			// 添加加钟商品
			for _, delayProduct := range saleOrder.SaleOrderBuffetDelayProducts {
				if delayProduct.IsDelete() {
					continue
				}
				products = append(products, &resp.OrderExportInfoProduct{
					Name:       delayProduct.Name,
					AttrName:   "",
					Num:        float64(delayProduct.Num),
					TotalPrice: delayProduct.GetAmount(),
				})
			}
			// 添加正常商品
			for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
				if saleOrderProduct.IsDelete() {
					continue
				}
				products = append(products, &resp.OrderExportInfoProduct{
					Name:       saleOrderProduct.MultiLanguageName.GetNameByLang(language),
					AttrName:   saleOrderProduct.GetAttributeNamesByLang(language),
					Num:        saleOrderProduct.Num,
					TotalPrice: saleOrderProduct.GetTotalPrice(),
				})
			}
			//
			exportLists = append(exportLists, resp.OrderExportInfo{
				CreateTime:    saleOrder.CreateTime,
				BillUuid:      bill.Uuid,
				BillType:      utils.IfString(bill.IsInstantBill(), i18n.Translate(language, "点餐订单"), i18n.Translate(language, "桌台订单")),
				Products:      products,
				SerialNo:      utils.IfString(isSplit, bill.SerialNo+"-"+strconv.Itoa(index+1), bill.SerialNo),
				OrderNo:       saleOrder.OrderNo,
				Status:        bill.Status,
				StatusText:    statusText[bill.Status],
				FinishTime:    saleOrder.FinishTime,
				OrderAmount:   saleOrder.OriginAmount,
				ServiceFee:    saleOrder.ServiceFee,
				DiscountFee:   saleOrder.CustomDiscountFee,
				MemberFee:     saleOrder.MemberDiscountFee,
				PaymentAmount: saleOrder.GetActualPaymentAmount(),
				RefundAmount:  saleOrder.GetTotalRefundAmount(),
				MemberNames:   saleOrder.GetMemberName(),
				MemberIds: func() string {
					if saleOrder.Member == nil {
						return ""
					}
					return strconv.FormatUint(uint64(saleOrder.Member.ID), 10)
				}(),
				PayTypeName:  saleOrder.GetPayTypeNames(ctx.GetLanguage()),
				DiningMethod: bill.DiningMethod,
				CashierName:  bill.CashierName,
			})
		}
		// 拆单
		if isSplit && len(exportLists) > 0 {
			mainOrder := exportLists[len(exportLists)-1]
			// 收集当前账单所有会员名称并去重
			var allMemberNames []string
			var allMemberUuids []string
			var allProducts []*resp.OrderExportInfoProduct
			var allPayTypeNames []string
			for _, orderExportInfo := range exportLists {
				if orderExportInfo.BillUuid == bill.Uuid {
					// 添加产品和支付方式
					allProducts = append(allProducts, orderExportInfo.Products...)
					// 处理MemberNames，将字符串拆分为数组并逐个添加
					if orderExportInfo.MemberNames != "" {
						memberNames := strings.Split(orderExportInfo.MemberNames, ",")
						for _, name := range memberNames {
							if name != "" {
								allMemberNames = append(allMemberNames, name)
							}
						}
					}
					if orderExportInfo.MemberIds != "" {
						memberIds := strings.Split(orderExportInfo.MemberIds, ",")
						for _, id := range memberIds {
							if id != "" {
								allMemberUuids = append(allMemberUuids, id)
							}
						}
					}
					if orderExportInfo.PayTypeName != "" {
						payTypeNames := strings.Split(orderExportInfo.PayTypeName, ",")
						for _, name := range payTypeNames {
							if name != "" {
								allPayTypeNames = append(allPayTypeNames, name)
							}
						}
					}
				}
			}
			//
			mainOrder.Products = allProducts
			mainOrder.SerialNo = bill.SerialNo
			mainOrder.OrderNo = bill.OrderNo
			mainOrder.Status = bill.Status
			mainOrder.StatusText = statusText[bill.Status]
			mainOrder.FinishTime = bill.FinishTime
			mainOrder.OrderAmount = bill.OriginAmount
			mainOrder.ServiceFee = bill.ServiceFee
			mainOrder.DiscountFee = bill.CustomDiscountFee
			mainOrder.MemberFee = bill.MemberDiscountFee
			mainOrder.PaymentAmount = bill.GetPaymentAmount()
			mainOrder.RefundAmount = bill.GetTotalRefundAmount()
			mainOrder.MemberIds = strings.Join(utils.RemoveDuplicates(allMemberUuids), ",")
			mainOrder.MemberNames = strings.Join(utils.RemoveDuplicates(allMemberNames), ",")
			mainOrder.PayTypeName = strings.Join(utils.RemoveDuplicates(allPayTypeNames), ",")
			mainOrder.DiningMethod = bill.DiningMethod
			mainOrder.CashierName = bill.CashierName
			exportLists = append(exportLists, mainOrder)
		}
	}
	rankLists, err := orderRepo.GetMonthlyOrderRanks(saleBillUuids)
	if err != nil {
		return resp.OrderExportListPaginationResp{}, errors.WithMessage(err)
	}
	for i, exportList := range exportLists {
		result, ok := slice.FindBy(rankLists, func(index int, rankList repository.MonthlyOrderRank) bool {
			return rankList.OrderNo == exportList.OrderNo
		})
		if ok {
			exportLists[i].OrderID = fmt.Sprintf("OID%s%05d", result.MonthYear, result.MonthlyOrderNumber)
		}
	}
	//
	return resp.OrderExportListPaginationResp{
		List: exportLists,
		Meta: resp.OrderExportMeta{
			PageResponse: dto.PageResponse{
				PageNo:   req.PageNo,
				PageSize: req.PageSize,
				Total:    total,
			},
		},
	}, errors.WithMessage(err)
}

// GetOrderInfos 获取收银端订单信息
func (s *orderSrv) GetOrderInfos(ctx context.Context, req req.OrderInfoReq) (resp.OrderInfosResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)

	// 获取信息源
	saleBill, err := orderRepo.GetSaleBillInfo(req.SaleBillUuid, 0)
	if err != nil {
		return resp.OrderInfosResp{}, errors.WithMessage(err)
	}
	isMain := req.SaleOrderUuid == 0        // 是否查询主单
	isSplit := len(saleBill.SaleOrders) > 1 // 是否拆单
	isCellCancel := isMain

	// 组合信息
	totalMemberNames := []string{}
	totalMemberUuids := []string{}
	orderList := make([]resp.OrderInfo, 0)
	for i, saleOrder := range saleBill.SaleOrders {
		if req.SaleOrderUuid > 0 && req.SaleOrderUuid != saleOrder.Uuid {
			continue
		}
		if saleOrder.GetMemberName() != "" && !slices.Contains(totalMemberNames, saleOrder.GetMemberName()) {
			totalMemberNames = append(totalMemberNames, saleOrder.GetMemberName())
		}
		if saleOrder.ConsumerUuid != 0 {
			totalMemberUuids = append(totalMemberUuids, strconv.FormatUint(uint64(saleOrder.Member.ID), 10))
		}
		//
		products := make([]resp.OrderProduct, 0)

		// 添加自助餐顾客
		{
			for _, orderBuffetCustomer := range saleOrder.SaleOrderBuffetCustomerTypes {
				if orderBuffetCustomer.IsDelete() {
					continue
				}
				// 自助餐顾客价格收费列表
				products = append(products, resp.OrderProduct{
					Uuid:       orderBuffetCustomer.Uuid,
					LocaleName: orderBuffetCustomer.BuffetPackage.MultiLanguageName.GetNames(),
					LocaleAttributeName: dto.LocaleResponse{
						ZH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						TH:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						EN:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						ZHTW: orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						JA:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						KO:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						MY:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						TR:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
						SV:   orderBuffetCustomer.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					},
					Price:            orderBuffetCustomer.SalePrice,
					Num:              float64(orderBuffetCustomer.Num), // 这种类型顾客多少个，如老人这个类型2人
					SalePrice:        orderBuffetCustomer.GetDiscountPriceWithVAT(model.WithOriginPrice()),
					TotalPrice:       orderBuffetCustomer.GetDiscountPriceWithVAT(),
					RefundAmount:     -orderBuffetCustomer.GetReturnPrice(),
					Status:           1,
					Remark:           "",
					IsMust:           false,
					IsGift:           false,
					IsBuffetCustomer: true,
				})
			}
		}

		// 添加加钟商品
		{
			for _, delayProduct := range saleOrder.SaleOrderBuffetDelayProducts {
				if delayProduct.IsDelete() {
					continue
				}
				products = append(products, resp.OrderProduct{
					Uuid: delayProduct.Uuid,
					LocaleName: dto.LocaleResponse{
						ZH:   delayProduct.Name,
						TH:   delayProduct.Name,
						EN:   delayProduct.Name,
						ZHTW: delayProduct.Name,
						JA:   delayProduct.Name,
						KO:   delayProduct.Name,
						MY:   delayProduct.Name,
						TR:   delayProduct.Name,
						SV:   delayProduct.Name,
					},
					LocaleAttributeName: dto.LocaleResponse{},
					Num:                 float64(delayProduct.Num), // 拆单后不等于桌台人数，但同一个加钟商品的总数等于桌台人数
					Price:               delayProduct.Price,
					SalePrice:           delayProduct.GetAmount(),
					TotalPrice:          delayProduct.GetAmount(),
					RefundAmount:        -delayProduct.GetReturnPrice(),
					Status:              1,  // 添加后标记送厨状态，不可修改
					Remark:              "", // 加钟商品没有备注
					IsMust:              false,
					IsGift:              false,
					IsBuffet:            false,
					IsDelay:             true,
				})
			}
		}

		// 添加正常商品
		{
			for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
				if saleOrderProduct.IsDelete() {
					continue
				}
				// 取消订单时，过滤掉未送厨的商品
				if saleBill.IsCanceled() {
					if saleOrderProduct.IsUnCookingProduct() {
						continue
					}
				}
				// 过滤掉套餐子商品
				if saleOrderProduct.IsPackageSubProduct() {
					continue
				}

				// 过滤掉未接单的商品
				if !saleOrderProduct.IsAcceptOrderProduct() {
					continue
				}
				imageUrl := ""
				if saleOrderProduct.ImageFile != nil {
					imageUrl = saleOrderProduct.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request))
				}
				cancelReason := saleOrderProduct.GetCancelReason()
				giftReason := saleOrderProduct.GetGiftReason()

				attributeName := saleOrderProduct.GetAttributeName()
				if saleOrderProduct.IsPackageProduct() {
					// 如果是套餐商品，则获取各个子商品的名称、数量、规格、属性，如：“牛排*1（标准，黑椒汁）；可乐*2（大杯，少冰）；沙拉*1（大份，沙拉酱，蜂蜜酱）”
					subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
					zh := ""
					th := ""
					en := ""
					zhtw := ""
					ja := ""
					ko := ""
					my := ""
					tr := ""
					sv := ""
					for _, subProduct := range subProducts {
						zh += subProduct.GetProductNameAttributes(string(constant.LocaleZH)) + "；"
						th += subProduct.GetProductNameAttributes(string(constant.LocaleTH)) + "；"
						en += subProduct.GetProductNameAttributes(string(constant.LocaleEN)) + "；"
						zhtw += subProduct.GetProductNameAttributes(string(constant.LocaleZHTW)) + "；"
						ja += subProduct.GetProductNameAttributes(string(constant.LocaleJA)) + "；"
						ko += subProduct.GetProductNameAttributes(string(constant.LocaleKO)) + "；"
						my += subProduct.GetProductNameAttributes(string(constant.LocaleMY)) + "；"
						tr += subProduct.GetProductNameAttributes(string(constant.LocaleTR)) + "；"
						sv += subProduct.GetProductNameAttributes(string(constant.LocaleSV)) + "；"
					}
					attributeName = dto.LocaleResponse{
						ZH:   zh,
						TH:   th,
						EN:   en,
						ZHTW: zhtw,
						JA:   ja,
						KO:   ko,
						MY:   my,
						TR:   tr,
						SV:   sv,
					}
				}

				products = append(products, resp.OrderProduct{
					Uuid:                saleOrderProduct.Uuid,
					LocaleName:          saleOrderProduct.MultiLanguageName.GetNames(),
					LocaleAttributeName: attributeName,
					Price:               saleOrderProduct.SalePrice,
					Num:                 saleOrderProduct.Num,
					SalePrice:           saleOrderProduct.GetTotalPriceOrigin(),
					TotalPrice:          saleOrderProduct.GetTotalPrice(),
					Status:              saleOrderProduct.Status,
					Remark:              saleOrderProduct.Remark,
					IsMust:              saleOrderProduct.IsMustProduct(),
					IsGift:              saleOrderProduct.IsGiftProduct(),
					IsWrap:              saleOrderProduct.IsWrapProduct(),
					IsBuffet:            saleOrderProduct.IsBuffetProduct(),
					ImageUrl:            imageUrl,
					CancelReason:        cancelReason.GetLocale(ctx.GetLanguage()),
					GiftReason:          giftReason.GetLocale(ctx.GetLanguage()),
					RefundAmount:        -saleOrderProduct.GetReturnPrice(),
				})
			}
		}

		//
		orderList = append(orderList, resp.OrderInfo{
			SaleOrderUuid: saleOrder.Uuid,
			BillType:      saleBill.BillType,
			DiningMethod:  saleBill.DiningMethod,
			SerialNo:      saleBill.SerialNo + "-" + strconv.Itoa(i+1),
			OrderNo:       saleOrder.OrderNo,
			Status:        saleOrder.Status,
			IsFree:        saleOrder.IsFree == 1,
			FreeReason:    saleOrder.GetFreeReason(),
			OrderAmount:   saleOrder.OriginAmount,
			PaymentAmount: saleOrder.GetActualPaymentAmount(),
			RefundAmount:  saleOrder.GetTotalRefundAmount(),
			PayTypeName:   saleOrder.GetPayTypeNames(ctx.GetLanguage()),
			MemberName:    saleOrder.GetMemberName(),
			MemberUuid: func() uint64 {
				if saleOrder.Member == nil {
					return uint64(0)
				}
				return uint64(saleOrder.Member.ID)
			}(),
			Products: products,
		})
		//
		if saleOrder.Status != constant.SaleBillStatusPending {
			isCellCancel = false
		}
	}

	// 处理额外信息
	var order *model.SaleOrder
	if len(saleBill.SaleOrders) > 0 {
		order = saleBill.SaleOrders[0]
	}
	orderExtra := resp.BillListsExtra{
		IsCellRefund:        false,
		IsCellCancel:        isCellCancel,
		IsCellReverseSettle: saleBill.IsCellReverseSettle(ctx.GetStaff().Uuid, ctx.GetStaff().CashierLoginTime),
		IsCellPrint:         true,
		IsCellDelete:        order.Status == constant.SaleBillStatusCanceled,
		IsCellInvoice:       false,
	}
	if (!isSplit || !isMain) && order.IsFree == 0 && order.GetTotalRefundAmount() < order.PaymentAmount && order.Status == constant.SaleBillStatusComplete {
		orderExtra.IsCellRefund = true
	}

	// 返回响应对象
	return resp.OrderInfosResp{
		Detail: resp.OrderInfos{
			SaleBillUuid: saleBill.Uuid,
			IsSplit:      isSplit,
			BillType:     saleBill.BillType,
			DiningMethod: saleBill.DiningMethod,
			SerialNo:     saleBill.SerialNo,
			OrderNo: func() string {
				if isMain {
					return saleBill.OrderNo
				}
				return order.OrderNo
			}(),
			Status:        saleBill.Status,
			CreateTime:    saleBill.CreateTime,
			FinishTime:    saleBill.FinishTime,
			OrderAmount:   saleBill.OriginAmount,
			PaymentAmount: saleBill.GetPaymentAmount(),
			RefundAmount:  saleBill.GetTotalRefundAmount(),
			MemberNames:   strings.Join(totalMemberNames, ","),
			MemberUuids:   strings.Join(totalMemberUuids, ","),
			CashierName:   saleBill.CashierName,
			IsBuffet:      saleBill.IsBuffet == constant.SaleBillIsBuffetYes,
			BuffetNames:   saleBill.GetBuffetNames(ctx.GetLanguage()),
			CancelReason:  saleBill.Reason,
			OrderSourceUuid: func() uint64 {
				if saleBill.OrderSource != nil {
					return saleBill.OrderSource.Uuid
				}
				return 0
			}(),
			OrderSourceName: func() string {
				if saleBill.OrderSource != nil {
					return saleBill.OrderSource.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
				}
				return ""
			}(),
			NationalityUuid: func() uint64 {
				if saleBill.Nationality != nil {
					return saleBill.Nationality.Uuid
				}
				return 0
			}(),
			// 使用快照数据（优先），降级使用关联表（兼容历史）
			// Requirement: story-main-nationality-snapshot-fix (JSON 方案)
			NationalityName: func() string {
				nationalityName := saleBill.GetLocaleNationalityName()
				if nationalityName.IsNull() {
					return ""
				}
				return nationalityName.GetLocale(ctx.GetLanguage())
			}(),
			PayTypes:   saleBill.GetPayTypes(ctx.GetLanguage(), req.SaleOrderUuid),
			SaleOrders: orderList,
			Remark:     saleBill.Remark,
		},
		OperationLog: struct {
			List []resp.OrderOperationLog `json:"list"`
		}{
			List: func() []resp.OrderOperationLog {
				logs, err := s.GetRecordList(ctx, req.SaleBillUuid, 0)
				if err != nil {
					return []resp.OrderOperationLog{}
				}
				return logs
			}(),
		},
		Extra: orderExtra,
	}, nil
}

// GetRecordList 获取操作记录
func (s *orderSrv) GetRecordList(ctx context.Context, saleBillUuid uint64, h5OrderUuid uint64) ([]resp.OrderOperationLog, error) {
	orderRecordRepo := repository.NewOrderOperationRecordRepo(ctx.GetDB())
	var dbOptions []repository.DBOption
	dbOptions = append(dbOptions, orderRecordRepo.WithSaleBillUuid(saleBillUuid))
	if h5OrderUuid > 0 {
		dbOptions = append(dbOptions, orderRecordRepo.WithH5OrderUuid(h5OrderUuid))
	}
	orderRecordLists, err := orderRecordRepo.GetRecordLists(dbOptions...)
	if err != nil {
		return []resp.OrderOperationLog{}, errors.WithMessage(err)
	}
	logs := make([]resp.OrderOperationLog, 0)
	language := ctx.GetLanguage()

	for _, record := range orderRecordLists {
		actionDescription := s.getActionDescription(ctx, record, language)
		if actionDescription.HideLog {
			// 隐藏. 日志关联的订单已经被删除，故因此该订单的操作记录
			continue
		}

		// 获取操作描述
		actionText := s.getActionText(record, language)
		if record.Action == constant.OrderCheckoutDiscount && actionDescription.IsAutoCheckoutZero {
			actionText = i18n.Translate(language, "结账自动抹零")
		}
		if actionDescription.SplitMessage != "" {
			actionText = actionDescription.SplitMessage + actionText
		}
		if actionDescription.Desc != "" {
			if record.Source == constant.SourceMember || record.Action == constant.OrderCancelMemberSaleOrder {
				actionText = actionText + actionDescription.Desc
			} else {
				actionText = actionText + ": " + actionDescription.Desc
			}
		}
		realName := record.Operator.RealName
		email := record.Operator.Username

		// 获取授权人信息
		if strings.Contains(record.Data, "authorized_staff") { // 授权操作时，记录了授权人信息
			var authorizedStaffInfo event.AuthorizedStaffInfo
			if err := utils.ExtractNestedFieldToStruct(record.Data, "authorized_staff", &authorizedStaffInfo); err == nil {
				realName = authorizedStaffInfo.Name
				email = authorizedStaffInfo.Email
			} else {
				ctx.Log().Info("解析订单操作记录时，获取授权人信息失败", zap.Any("companyUuid", ctx.GetCompanyUuid()), zap.Any("record", record), zap.Error(err))
			}
		}

		if record.Source == constant.SourceH5 {
			realName = i18n.Translate(language, "用户")
		}

		// 如果是骑手端的操作
		if record.Source == constant.SourceRider {
			prefix := i18n.Translate(language, "骑手")
			riderName := "--"
			var payload event.RiderAcceptMemberSaleOrderPayload
			err := json.Unmarshal([]byte(record.Data), &payload)
			if err == nil {
				riderName = payload.RiderName
			}
			realName = fmt.Sprintf("%s(%s)", prefix, riderName)
		}

		// 如果是会员端的操作
		if record.Source == constant.SourceMember {
			prefix := i18n.Translate(language, "顾客")
			realName = fmt.Sprintf("%s(%s)", prefix, record.Member.Nickname)
		}

		// 如果是系统自动操作
		if record.Source == "" {
			realName = i18n.Translate(language, "系统自动")
		}

		logs = append(logs, resp.OrderOperationLog{
			Uuid:       record.Uuid,
			RealName:   realName,
			Email:      email,
			Source:     i18n.Translate(language, constant.SourceTextMap[record.Source]),
			CreateTime: record.CreateTime,
			RefundType: func() int {
				var refundType int
				if record.Action == constant.OrderRefund {
					var refundPayload event.ReturnOrderPayload
					json.Unmarshal([]byte(record.Data), &refundPayload)
					refundType = refundPayload.RefundType
				}
				return refundType
			}(),
			Description: actionText, // 获取描述
			PayType:     s.getRefundPayType(ctx, record, language),
		})
	}
	return logs, nil
}

// CancelOrder 取消订单
func (s *orderSrv) CancelOrder(ctx context.Context, req req.OrderCancelReq) error {
	dbId := ctx.GetDbId()
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取订单信息
	billInfo, err := s.IsCellCancelOrder(ctx, req.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if billInfo.ID == 0 {
		return errors.New("找不到订单")
	}

	// 验证高级密码
	if s.settingSrv == nil {
		return errors.New("找不到 settingSrv")
	}
	// 如果不需要验证高级密码，则跳过
	if !req.NotNeedPassword {
		if err := s.settingSrv.VerifyAdvancedPassword(ctx, req.Password); err != nil {
			return errors.WithMessage(err)
		}
	}

	// 获取信息源
	db := s.dbm.GetDB(dbId)

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 如果发生恐慌，回滚事务
		}
	}()

	orderRepo := repository.NewOrderRepo(tx)
	deskRepo := repository.NewDeskRepo(tx)

	// 退回商品库存
	{
		// 获取销售账单信息
		saleBill, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		s.returnInventory(ctx, saleBill)
	}

	// 如果是桌台订单
	if billInfo.IsDeskBill() && billInfo.DeskUuid > 0 {
		// 拒绝所有待接单
		ctx.SetDB(tx)
		if err := s.RejectAllH5Order(ctx, billInfo.Uuid); err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
		}
		// 关闭桌台
		err = deskRepo.CloseDesk(ctx, billInfo.DeskUuid, billInfo.Uuid, req.CancelReason)
		if err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
		}
	} else {
		err = orderRepo.CancelOrder(ctx, req.SaleBillUuid, 0, req.CancelReason)
		if err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
		}
	}

	// 标记送厨单、送厨商品为删除
	productionRepo := repository.NewProductionRepo(tx)
	saleBillUuidOpt := productionRepo.WhereSaleBillUuid(billInfo.Uuid)
	err = productionRepo.UpdateOrder([]repository.DBOption{saleBillUuidOpt}, map[string]any{
		"delete_time": time.Now().Unix(),
	})
	if err != nil {
		tx.Rollback()
		return errors.WithMessage(builtinerrors.New("删除送厨单失败"), err.Error())
	}
	// 修改送厨商品数量为0，在确认整单退菜时、确认该菜品全退时，再标记为删除
	err = productionRepo.UpdateProduct([]repository.DBOption{saleBillUuidOpt}, map[string]any{
		"num": 0,
	})
	if err != nil {
		tx.Rollback()
		return errors.WithMessage(builtinerrors.New("删除送厨单商品失败"), err.Error())
	}

	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err, "销售账单不存在")
	}
	saleBill.SetCanceled()
	// 计算订单商品、订单、账单金额并更新或创建
	if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill, model.WithCanceled()); err != nil {
		tx.Rollback()
		return errors.WithMessage(err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.WithMessage(err)
	}

	// 发布"整单取消"操作事件
	utils.Go(func() {
		s.bus.PublishCancelOrderEvent(event.CancelOrderPayload{
			BasePayload: event.BasePayload{ // 整单取消
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: billInfo.Uuid,
				OperatorUuid: int64(ctx.GetStaffUuid()),
			},
		})
	})

	// 成功后，推送到厨显端更新订单
	utils.Go(func() {
		websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]interface{}{
			"update_time": time.Now().Unix(),
		})
	})

	return nil
}

// DeleteOrder 删除订单, saleOrderUuid = 等于0的时候删除主单，并且主单下的子单也会被删除， saleOrderUuid > 0 的时候删除子单
func (s *orderSrv) DeleteOrder(ctx context.Context, dbId uint64, saleBillUuid uint64, saleOrderUuid uint64) error {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(saleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(saleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	db := s.dbm.GetDB(dbId)
	orderRepo := repository.NewOrderRepo(db)

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillInfo(saleBillUuid, constant.OptionalUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if billInfo.ID == 0 {
		return errors.New("找不到订单")
	}

	if billInfo.Status != constant.SaleBillStatusCanceled {
		return errors.New("订单状态不允许删除")
	}

	err = orderRepo.DeleteOrder(saleBillUuid, saleOrderUuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 检查是否有已送厨的商品，如果有，则标记production_order_product.status为消单退菜（制作中消单退菜、制作完成消单退菜）
	// 如果已送厨商品还在制作中，通知厨房取消制作
	doingProductList := make([]uint64, 0) // 制作中的商品uuid列表 sale_order_product_uuid
	// TODO 获取还在制作中的商品

	// 发布事件，通知厨房取消制作
	event.NewSystemBus().PublishCancelDoingProductEvent(event.CancelDoingProductPayload{SaleOrderProductUuids: doingProductList})
	return nil
}

// ReturnOrder 退款订单
func (s *orderSrv) ReturnOrder(ctx context.Context, request req.OrderReturnReq) (error, int) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(request.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	// 版本判断：从请求头获取客户端版本
	// 如果版本 >= v2.10.0，进行权限验证；否则不进行权限验证（向后兼容）
	var authorizedStaff *model.Staff
	if ctx.Version(context.GTE, constant.ClientVersionV2100) {
		// 版本 >= v2.10.0，进行授权验证（退款操作）
		var err error
		authorizedStaff, err = s.AuthorizeSensitiveOperation(ctx, SensitiveOperationTypeRefund, request.AuthorizedStaffAccount, request.AuthorizedStaffPassword)
		if err != nil {
			return errors.WithMessage(err), constant.CodeFail
		}
	}

	// 获取门店设置
	storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
	if err != nil {
		logger.Logger.Info("ReturnOrder process, GetStoreSetting failed", zap.Error(err))
		return errors.WithMessage(err), constant.CodeFail
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)
	// 获取销售账单信息
	saleBill, err := orderRepo.GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}

	// 获取销售订单信息
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return errors.WithMessage(errors.New("找不到销售订单")), constant.CodeFail
	}

	if request.Points > saleOrder.GetManualReturnPoints() {
		return errors.WithMessage(errors.New("退款积分不能大于最大可退积分")), constant.CodeFail
	}

	returnType := constant.ReturnOrderRefundTypeTotal
	saleOrderProducts := make([]*model.SaleOrderProduct, 0)                       // 退款商品列表
	saleOrderBuffetCustomerTypes := make([]*model.SaleOrderBuffetCustomerType, 0) // 退款自助餐顾客列表
	saleOrderBuffetDelayProducts := make([]*model.SaleOrderBuffetDelayProduct, 0) // 退款自助餐延迟商品列表
	numMap := make(map[uint64]float64)                                            // 每个退款商品的退货数量
	// 整单退款
	if len(request.Products) == 0 {
		returnType = constant.ReturnOrderRefundTypeTotal
		// 整单退款，退款商品列表为销售订单商品列表.
		// 注意：要判断订单商品是否还有可退货数量
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			canReturnNum := saleOrderProduct.GetCanReturnNum() // 可退货数量
			if canReturnNum > 0 {
				saleOrderProducts = append(saleOrderProducts, saleOrderProduct)
				numMap[saleOrderProduct.Uuid] = canReturnNum
			}
		}

		for _, saleOrderProduct := range saleOrder.SaleOrderBuffetCustomerTypes {
			canReturnNum := saleOrderProduct.GetCanReturnNum() // 可退货数量
			if canReturnNum > 0 {
				saleOrderBuffetCustomerTypes = append(saleOrderBuffetCustomerTypes, saleOrderProduct)
				numMap[saleOrderProduct.Uuid] = float64(canReturnNum)
			}
		}

		for _, saleOrderProduct := range saleOrder.SaleOrderBuffetDelayProducts {
			canReturnNum := saleOrderProduct.GetCanReturnNum() // 可退货数量
			if canReturnNum > 0 {
				saleOrderBuffetDelayProducts = append(saleOrderBuffetDelayProducts, saleOrderProduct)
				numMap[saleOrderProduct.Uuid] = float64(canReturnNum)
			}
		}
	}
	// 部分退款
	isPartReturn := false // 部分退款时，如果有赠菜，先不传给erp
	if len(request.Products) > 0 {
		returnType = constant.ReturnOrderRefundTypePart
		isPartReturn = true
		// 获取退款商品列表
		saleOrderProductUuids := make([]uint64, 0)
		for _, product := range request.Products {
			saleOrderProductUuids = append(saleOrderProductUuids, product.SaleOrderProductUuid)
			numMap[product.SaleOrderProductUuid] = product.Num
		}
		// 注意：要判断订单商品是否还有可退货数量
		saleOrderProductList := saleOrder.GetSaleOrderProductList(saleOrderProductUuids)
		for _, saleOrderProduct := range saleOrderProductList {
			canReturnNum := saleOrderProduct.GetCanReturnNum() // 可退货数量
			if canReturnNum > 0 {
				num := numMap[saleOrderProduct.Uuid] // 退货数量
				if num <= canReturnNum {
					saleOrderProducts = append(saleOrderProducts, saleOrderProduct)
				} else {
					return errors.WithMessage(errors.New("1退货数量超过可退货数量")), constant.CodeFail
				}
			}
		}

		saleOrderBuffetCustomerTypeList := saleOrder.GetSaleOrderBuffetComstomerTypeList(saleOrderProductUuids)
		for _, saleOrderProduct := range saleOrderBuffetCustomerTypeList {
			canReturnNum := saleOrderProduct.GetCanReturnNum() // 可退货数量
			if canReturnNum > 0 {
				num := numMap[saleOrderProduct.Uuid] // 退货数量
				if uint(num) <= canReturnNum {
					saleOrderBuffetCustomerTypes = append(saleOrderBuffetCustomerTypes, saleOrderProduct)
				} else {
					return errors.WithMessage(errors.New("2退货数量超过可退货数量")), constant.CodeFail
				}
			}
		}

		saleOrderBuffetDelayProductsList := saleOrder.GetSaleOrderBuffetDelayList(saleOrderProductUuids)
		for _, saleOrderProduct := range saleOrderBuffetDelayProductsList {
			canReturnNum := saleOrderProduct.GetCanReturnNum() // 可退货数量
			if canReturnNum > 0 {
				num := numMap[saleOrderProduct.Uuid] // 退货数量
				if num <= float64(canReturnNum) {
					saleOrderBuffetDelayProducts = append(saleOrderBuffetDelayProducts, saleOrderProduct)
				} else {
					return errors.WithMessage(errors.New("3退货数量超过可退货数量")), constant.CodeFail
				}
			}
		}

	}

	// 如果退款类型为部分退款，则必须有可退货的商品。整单退款则可以没有可退货的商品，可能已经退完商品了但还有手续费没有退
	if len(saleOrderProducts) == 0 && len(saleOrderBuffetCustomerTypes) == 0 && len(saleOrderBuffetDelayProducts) == 0 && returnType == constant.ReturnOrderRefundTypePart {
		return errors.WithMessage(errors.New("没有可退货的商品")), constant.CodeFail
	}

	// 可退款金额
	canReturnAmount := saleOrder.GetCanReturnAmount()
	// 如果是会员端订单，则需要根据配送费计算可退款金额
	deliveryFee := 0.0
	if ctx.GetScene() == constant.SceneMemberOrder {
		memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecordOnlyBySaleBillUuid(request.SaleBillUuid)
		if err != nil {
			return errors.WithMessage(err), constant.CodeFail
		}
		deliveryFee = memberSaleOrder.DeliveryFeeAmount
		canReturnAmount = saleOrder.GetCanReturnAmountWithDeliveryFee(deliveryFee)
	}
	// 可退的会员消费金额
	canReturnMemberConsumptionAmount := saleOrder.GetCanReturnMemberConsumptionAmount()

	// 获取当前员工班次信息
	var staffShiftLogUuid uint64
	if ctx.GetStaffUuid() != 0 {
		staffShiftLog, err := GetCurrentStaffShiftLog(db, ctx.GetStaffUuid())
		if err == nil {
			staffShiftLogUuid = staffShiftLog.Uuid
		}
	}

	// 创建退款单
	returnOrder, err := saleOrder.NewReturnOrder(ctx.GetScene(), deliveryFee, ctx.GetStaff().DutyNo, ctx.GetLanguage(), saleOrderProducts, saleOrderBuffetCustomerTypes, saleOrderBuffetDelayProducts, numMap, returnType, canReturnAmount, staffShiftLogUuid)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}

	// 本次退款的会员累计消费金额。=退款金额
	returnOrderMemberConsumptionAmount := returnOrder.RefundAmount
	if returnOrderMemberConsumptionAmount > canReturnMemberConsumptionAmount {
		// 本次退款的会员累计消费金额不能大于可退的会员消费金额
		returnOrderMemberConsumptionAmount = canReturnMemberConsumptionAmount
	}

	// 是否存在QrPromptPay支付
	if returnOrder.IsExistQrPromptPay() {
		if request.BankCode == "" || request.AccountNo == "" || request.AccountName == "" {
			return errors.WithMessage(errors.New("请选择银行")), constant.CodeReturnOrderBank
		}
		returnOrder.BankCode = request.BankCode
		returnOrder.AccountNo = request.AccountNo
		returnOrder.AccountName = request.AccountName
	}

	lianLianPayCount := returnOrder.GetLianLianPayCount()

	var publishChangeMemberBalance, publishChangeMemberPoints, isExistCashPay bool
	// 创建
	err = repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		ctx.SetDB(db) // 否则 s.memberSrv.HandleMemberBalance会事务失效

		// 创建退货单
		if _, err = repository.NewReturnOrderRepo(db).CreateReturnOrderRecord(*returnOrder); err != nil {
			return errors.WithMessage(err)
		}
		// 创建连连退款订单
		for _, returnOrderAmount := range returnOrder.ReturnOrderAmounts {
			if lianLianPayCount > 0 && returnOrderAmount.PaymentMethod.IsLianLianPay() {
				paymentServiceRefundReq := PaymentServiceRefundReq{
					RelatedType:           constant.PaymentOrderRelatedTypeSaleOrder,
					PaymentOrderUuid:      returnOrderAmount.PaymentOrderUuid,
					MerchantRefundOrderNo: returnOrderAmount.MerchantRefundOrderNo,
					RefundAmount:          returnOrderAmount.Amount,
					BankCode:              returnOrder.BankCode,
					AccountNo:             returnOrder.AccountNo,
					AccountName:           returnOrder.AccountName,
				}
				if lianLianPayCount > 1 {
					utils.Go(func() {
						payment, err := NewPaymentRepo(ctx, s.dbm).Refund(paymentServiceRefundReq)
						if err != nil {
							returnOrderAmount.RefundStatus = 2
							returnOrderAmount.LlReturnOrderid = "0"
						} else {
							returnOrderAmount.LlReturnOrderid = payment.RefundOrderId
						}
						// 更新退款状态
						returnOrderRepo := repository.NewReturnOrderRepo(s.dbm.GetDB(ctx.GetDbId()))
						returnOrderRepo.UpdateReturnOrderAmount([]repository.DBOption{
							returnOrderRepo.WhereUuid(returnOrderAmount.Uuid),
						}, returnOrderAmount)
						if err != nil {
							fmt.Println("更新退款状态失败", err)
							logger.Logger.Error("更新退款状态失败", zap.Error(err))
						}
					})
				} else {
					payment, err := NewPaymentRepo(ctx, s.dbm).Refund(paymentServiceRefundReq)
					if err != nil {
						return errors.WithMessage(err)
					}
					// 设置连连退款订单ID
					returnOrderAmount.LlReturnOrderid = payment.RefundOrderId
				}
			} else {
				returnOrderAmount.RefundStatus = 1
			}
			// 创建退款金额
			if err = repository.NewReturnOrderRepo(db).CreateReturnOrderAmount([]model.ReturnOrderAmount{returnOrderAmount}); err != nil {
				return errors.WithMessage(err)
			}
			// 如果退款金额为余额，则退回余额，创建余额变动记录
			if returnOrderAmount.PaymentMethod.Code == constant.PaymentMethodCodeBalance {
				if err := s.memberSrv.HandleMemberBalance(ctx, MemberBalanceChangeReq{
					MemberUuid:  returnOrderAmount.MemberBalanceLog.MemberUuid,
					GiftMoney:   returnOrderAmount.MemberBalanceLog.GiftMoney, // 退款金额。余额退款都是退回到赠送帐户
					Scene:       returnOrderAmount.MemberBalanceLog.Scene,
					Describe:    returnOrderAmount.MemberBalanceLog.Describe,
					RelatedUuid: returnOrderAmount.MemberBalanceLog.RelatedUuid,
				}); err != nil {
					return errors.WithMessage(err)
				}
				publishChangeMemberBalance = true
			}
			// 如果退款金额为现金，则更新钱箱
			if returnOrderAmount.PaymentMethod.Code == constant.PaymentMethodCodeCash {
				isExistCashPay = true
				// 存现金，更新钱箱
				ctx.SetDB(db)
				if err := s.cashBoxSrv.UpdateBalance(ctx, UpdateCashBalanceParam{
					Amount:    -returnOrderAmount.Amount,
					Scene:     constant.CashBoxLogSceneRefund,
					OrderUuid: returnOrderAmount.Uuid,
				}); err != nil {
					return errors.WithMessage(err)
				}
			}
		}
		// 创建退货单商品
		if err = repository.NewReturnOrderRepo(db).CreateReturnOrderProduct(returnOrder.ReturnOrderProducts); err != nil {
			return errors.WithMessage(err)
		}
		// 更新高峰时段
		if err := repository.NewSaleOrderPeakTimeRepo(db).Record("dec", saleBill, returnOrder.RefundAmount, storeSetting.TimeZone); err != nil {
			return errors.WithMessage(err)
		}
		// 退积分
		if saleOrder.ConsumerUuid > 0 {
			// 手动退积分
			if saleOrder.CanManualReturnPoints() {
				if request.Points > 0 {
					points := request.Points
					// 开始手动退积分
					member, err := repository.NewMemberRepo(db).GetMemberByUuid(saleOrder.ConsumerUuid)
					if err != nil {
						return errors.WithMessage(err)
					}
					if err := repository.NewMemberRepo(db).Update(saleOrder.ConsumerUuid, map[string]any{
						"frozen_point": member.FrozenPoint - points, // 扣减积分
					}); err != nil {
						return errors.WithMessage(err)
					}
					// 创建积分变动记录
					memberPointLog := saleOrder.NewRefundMemberPointLog(-points)
					if _, err := repository.NewMemberPointLogRepo(db).Create(*memberPointLog); err != nil {
						return errors.WithMessage(err)
					}
				}
			} else {
				// 自动退积分
				refundAmount := returnOrder.RefundAmount // 退款金额
				// 积分赠送比例
				integralGiveRate := saleOrder.GiftPointsRate
				//  部分退款时。退积分=退款金额*积分赠送比例
				points := decimal.NewFromFloat(refundAmount).Mul(decimal.NewFromFloat(integralGiveRate)).Round(2).InexactFloat64()
				// 退还积分不能超过可退积分
				if points > saleOrder.GetManualReturnPoints() {
					points = saleOrder.GetManualReturnPoints()
				}
				// 如果退款类型为整单退款，则退还积分剩余未退的积分
				if len(request.Products) == 0 {
					points = saleOrder.GetManualReturnPoints()
				}
				member, err := repository.NewMemberRepo(db).GetMemberByUuid(saleOrder.ConsumerUuid)
				if err != nil {
					return errors.WithMessage(err)
				}
				// 更新会员积分
				if points > 0 {
					if err := repository.NewMemberRepo(db).Update(saleOrder.ConsumerUuid, map[string]any{
						"frozen_point": member.FrozenPoint - points, // 扣减积分
					}); err != nil {
						return errors.WithMessage(err)
					}
					// 创建积分变动记录
					memberPointLog := saleOrder.NewRefundMemberPointLog(-points)
					if _, err := repository.NewMemberPointLogRepo(db).Create(*memberPointLog); err != nil {
						return errors.WithMessage(err)
					}
				}
			}
			publishChangeMemberPoints = true
		}

		// 退会员的累计消费金额。退款后减少会员的累计消费金额
		if saleOrder.ConsumerUuid > 0 {
			// 减少会员累计消费金额
			if err := repository.NewMemberRepo(db).DecConsumptionAmount(saleOrder.ConsumerUuid, returnOrderMemberConsumptionAmount); err != nil {
				return errors.WithMessage(err)
			}
			// 如果是整单退款，则减少会员累计消费次数
			if len(request.Products) == 0 {

			}
		}

		// 保存发票到erp
		company := ctx.GetCompany()
		companySetting := ctx.GetCompanySetting()
		if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
			res, err := s.ReturnPosInvoice(ctx, saleOrder, returnOrder, db, returnType, isPartReturn)
			if err != nil {
				return errors.WithMessage(err)
			}
			returnOrder.ErpInvoiceName = res.InvoiceName
		}
		// 更新退货单erp发票名
		if err := repository.NewReturnOrderRepo(db).UpdateReturnOrderRecordErpInvoiceName(returnOrder.Uuid, returnOrder.ErpInvoiceName); err != nil {
			return errors.WithMessage(err)
		}

		return nil
	})

	// 发送短信
	utils.Go(func() {
		// 获取最新的会员信息
		member, err := repository.NewMemberRepo(db).GetMemberByUuid(saleOrder.ConsumerUuid)
		if err != nil {
			ctx.Log().Info("停止发送短信（退款），获取会员失败", zap.Error(errors.WithMessage(err)))
		} else {
			refundAmount := float64(0)
			for _, returnOrderAmount := range returnOrder.ReturnOrderAmounts {
				if returnOrderAmount.PaymentMethod.Code == constant.PaymentMethodCodeBalance {
					refundAmount = returnOrderAmount.Amount
					break
				}
			}
			if refundAmount > 0 {
				if member != nil {
					smsReq := sms.MemberOrderRefundRequest{
						Company:       ctx.GetCompany().Name,
						OrderRefund:   refundAmount,
						Balance:       member.GetBalanceAll(),
						PointsBalance: member.GetPoints(),
					}
					if err := s.smsSrv.SendMemberOrderRefundSMS(ctx, member.Phone, &smsReq); err != nil {
						ctx.Log().Info("发送退款短信失败", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq), zap.Error(errors.WithMessage(err)))
					} else {
						ctx.Log().Info("发送退款短信成功", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq))
					}
				}
			}
		}
	})

	if publishChangeMemberBalance {
		// 发布"会员余额变动"事件
		utils.Go(func() {
			s.bus.PublishChangeMemberBalanceEvent(event.ChangeMemberBalancePayload{
				BasePayload: event.BasePayload{ // 会员余额变动
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: request.SaleBillUuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
			})
		})
	}

	if publishChangeMemberPoints {
		// 发布"会员积分变动"事件
		utils.Go(func() {
			s.bus.PublishChangeMemberPointsEvent(event.ChangeMemberPointsPayload{
				BasePayload: event.BasePayload{ // 会员积分变动
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: request.SaleBillUuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
			})
		})
	}

	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}
	// 发布"退款"事件
	products := make(event.Products, 0)
	for _, saleOrderProduct := range saleOrderProducts {
		if num, exists := numMap[saleOrderProduct.Uuid]; exists && num > 0 {
			products = append(products, event.OrderProduct{
				OrderProductId:  saleOrderProduct.Uuid,
				ProductId:       saleOrderProduct.ProductPackageUuid,
				ProductName:     saleOrderProduct.MultiLanguageName.GetNames(),
				ProductAttr:     saleOrderProduct.GetAttributeName(),
				ProductAttrList: saleOrderProduct.GetAttributeNameList(),
				TotalNum:        num,
				NumType:         saleOrderProduct.NumType,
				IsBuffet:        saleOrderProduct.IsBuffet == 1,
				IsWrap: func() bool {
					if saleBill.IsTakeout() && saleBill.MemberSaleOrderUuid == 0 {
						return true
					}
					return saleOrderProduct.IsWrapProduct()
				}(),
				Remark: saleOrderProduct.Remark,
			})
		}
	}
	for _, saleOrderProduct := range saleOrderBuffetCustomerTypes {
		if num, exists := numMap[saleOrderProduct.Uuid]; exists && num > 0 {
			products = append(products, event.OrderProduct{
				OrderProductId: saleOrderProduct.Uuid,
				ProductId:      saleOrderProduct.BuffetCustomerTypePriceUuid,
				ProductName:    saleOrderProduct.BuffetPackage.MultiLanguageName.GetNames(),
				ProductAttr: dto.LocaleResponse{
					ZH:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					TH:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					EN:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					ZHTW: saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					JA:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					KO:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					MY:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					TR:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
					SV:   saleOrderProduct.BuffetCustomerTypePrice.BuffetCustomerType.Name,
				},
				TotalNum: num,
			})
		}
	}
	for _, saleOrderProduct := range saleOrderBuffetDelayProducts {
		if num, exists := numMap[saleOrderProduct.Uuid]; exists && num > 0 {
			products = append(products, event.OrderProduct{
				OrderProductId: saleOrderProduct.Uuid,
				ProductId:      saleOrderProduct.BuffetDelayUuid,
				ProductName: dto.LocaleResponse{
					ZH:   saleOrderProduct.Name,
					TH:   saleOrderProduct.Name,
					EN:   saleOrderProduct.Name,
					ZHTW: saleOrderProduct.Name,
					JA:   saleOrderProduct.Name,
					KO:   saleOrderProduct.Name,
					MY:   saleOrderProduct.Name,
					TR:   saleOrderProduct.Name,
					SV:   saleOrderProduct.Name,
				},
				TotalNum: num,
			})
		}
	}
	var payTypes []event.RefundPayType
	for _, amount := range returnOrder.ReturnOrderAmounts {
		payTypes = append(payTypes, event.RefundPayType{
			Name:              amount.PaymentMethod.PaymentName,
			Code:              amount.PaymentMethod.Code,
			Amount:            amount.Amount,
			RefundStatus:      amount.RefundStatus,
			ReturnAmountUuid:  amount.Uuid,
			ReturnOrderUuid:   amount.ReturnOrderUuid,
			PaymentOrderUuid:  amount.PaymentOrderUuid,
			PaymentMethodUuid: amount.PaymentMethodUuid,
		})
	}

	// 记录外送订单的退款金额
	if saleBill.MemberSaleOrderUuid > 0 {
		utils.Go(func() {
			if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderRefundAmount(saleBill.MemberSaleOrderUuid, returnOrder.RefundAmount); err != nil {
				ctx.Log().Error("记录外送订单的退款金额失败", zap.Error(errors.WithMessage(err)))
			}
		})
	}

	// 构建授权员工信息（如果使用了授权验证）
	var authorizedStaffInfo *event.AuthorizedStaffInfo
	if authorizedStaff != nil {
		authorizedStaffInfo = &event.AuthorizedStaffInfo{
			Uuid:  authorizedStaff.Uuid,
			Name:  authorizedStaff.RealName,
			Email: authorizedStaff.Username,
		}
	}

	utils.Go(func() {
		s.bus.PublishReturnOrderEvent(event.ReturnOrderPayload{
			SaleBill: saleBill,
			BasePayload: event.BasePayload{ // 退款
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  saleBill.Uuid,
				SaleOrderUuid: saleOrder.Uuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			Products:        products,
			PayTypes:        payTypes,
			RefundType:      returnType,
			AuthorizedStaff: authorizedStaffInfo,
		})
	})
	// 发布"统计"事件
	utils.Go(func() {
		s.bus.PublishStatisticsSaleEvent(event.StatisticsSalePayload{
			BasePayload: event.BasePayload{ // 统计
				Ctx: ctx,
			},
			SaleBillUuid: saleBill.Uuid,
		})
	})

	if isExistCashPay {
		return nil, constant.CodeSuccessOpenCashBox
	}

	return nil, 0
}

// ReReturnOrder 重新退款
func (s *orderSrv) ReReturnOrder(ctx context.Context, req req.OrderReReturnReq) (error, int) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.ReturnOrderUuid)
		defer lock.NewSystemLock().UnlockUuid(req.ReturnOrderUuid)
		ctx.AddLock()
	}
	// 获取退款订单信息
	returnOrderRepo := repository.NewReturnOrderRepo(ctx.GetDB())
	orderAmount, err := returnOrderRepo.GetReturnOrderAmount(
		returnOrderRepo.WithReturnOrder(),
		returnOrderRepo.WithPaymentMethod(),
		returnOrderRepo.WhereUuid(req.ReturnAmountUuid),
	)
	if err != nil || orderAmount.ReturnOrder.Uuid != req.ReturnOrderUuid {
		return errors.New("找不到订单"), constant.CodeFail
	}
	if orderAmount.RefundStatus == 1 {
		return errors.New("该订单已成功退款，无法重复退款"), constant.CodeFail
	}
	if !orderAmount.PaymentMethod.IsLianLianPay() {
		return errors.New("该订单无法重新退款"), constant.CodeFail
	}
	// 判断订单是否正在退款
	if orderAmount.RefundStatus == 0 {
		return errors.New("该订单正在进行退款，无法重复操作"), constant.CodeFail
	}

	refundReq := PaymentServiceRefundReq{
		RelatedType:           constant.PaymentOrderRelatedTypeSaleOrder,
		PaymentOrderUuid:      orderAmount.PaymentOrderUuid,
		MerchantRefundOrderNo: orderAmount.MerchantRefundOrderNo,
		RefundAmount:          orderAmount.Amount,
		RefundOrderId:         orderAmount.LlReturnOrderid,
	}

	// 是否存在QrPromptPay支付
	isChangeBankCode := false
	if orderAmount.PaymentMethod.IsQrPromptPay() {
		if req.BankCode == "" || req.AccountNo == "" || req.AccountName == "" {
			return errors.WithMessage(errors.New("请选择银行")), constant.CodeReturnOrderBank
		}
		if req.BankCode != orderAmount.ReturnOrder.BankCode || req.AccountNo != orderAmount.ReturnOrder.AccountNo || req.AccountName != orderAmount.ReturnOrder.AccountName {
			isChangeBankCode = true
		}
		refundReq.BankCode = orderAmount.ReturnOrder.BankCode
		refundReq.AccountNo = orderAmount.ReturnOrder.AccountNo
		refundReq.AccountName = orderAmount.ReturnOrder.AccountName
	}

	// 发起退款
	refund, err := NewPaymentRepo(ctx, s.dbm).Refund(refundReq)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}
	if refund.RefundStatus == "RP" {
		return errors.New("该订单正在进行退款，无法重复操作"), constant.CodeFail
	}
	if refund.RefundStatus == "RS" {
		orderAmount.RefundStatus = 1
		err = returnOrderRepo.UpdateReturnOrderAmount([]repository.DBOption{returnOrderRepo.WhereUuid(orderAmount.Uuid)}, orderAmount)
		if err != nil {
			return errors.WithMessage(err), constant.CodeFail
		}
		return errors.New("该订单已成功退款，无法重复退款"), constant.CodeFail
	}
	// 更新银行信息 - 重新发起退款
	if isChangeBankCode {
		orderAmount.RefundStatus = 1
		orderAmount.MerchantRefundOrderNo = utils.GenerateMerchantOrderNo("RE")
		// 更新退款订单号
		refundReq.MerchantRefundOrderNo = orderAmount.MerchantRefundOrderNo
		refundReq.BankCode = req.BankCode
		refundReq.AccountNo = req.AccountNo
		refundReq.AccountName = req.AccountName
		// 更新银行信息
		orderAmount.ReturnOrder.BankCode = req.BankCode
		orderAmount.ReturnOrder.AccountNo = req.AccountNo
		orderAmount.ReturnOrder.AccountName = req.AccountName
	}
	// 重新发起退款
	refund, err = NewPaymentRepo(ctx, s.dbm).Refund(refundReq)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}
	// 更新退款订单号
	orderAmount.LlReturnOrderid = refund.RefundOrderId
	err = returnOrderRepo.UpdateReturnOrder([]repository.DBOption{returnOrderRepo.WhereUuid(orderAmount.ReturnOrder.Uuid)}, *orderAmount.ReturnOrder)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}
	err = returnOrderRepo.UpdateReturnOrderAmount([]repository.DBOption{returnOrderRepo.WhereUuid(orderAmount.Uuid)}, orderAmount)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}
	//
	return nil, 0
}

// GetReturnOrderInfo 获取退款信息
func (s *orderSrv) GetReturnOrderInfo(ctx context.Context, req req.OrderReturnInfoReq) (*resp.OrderReturnInfoResp, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)
	// 获取销售账单信息
	saleBill, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("找不到销售订单")
	}

	products := make([]resp.OrderReturnProduct, 0)

	// 获取销售订单的每个付款单的可退款金额
	// 要求排好序：退款顺序优先退会员、不够退则到现金、再到记录支付（多个时，哪个先后都行）、再到lianlian（多个时，哪个先后都行）
	paymentRecords, currencyUnit := saleOrder.GetPaymentOrderCanReturnAmount()

	// 获取销售订单的自助餐顾客列表
	buffetCustomers := saleOrder.GetCustomerList()
	for _, buffetCustomer := range buffetCustomers {
		products = append(products, resp.OrderReturnProduct{
			SaleOrderProductUuid: buffetCustomer.Uuid,
			LocaleName:           buffetCustomer.LocaleName,
			LocaleAttributeName:  buffetCustomer.LocaleAttributeName,
			Num:                  float64(buffetCustomer.CanReturnNum), // 自助餐顾客类型可退货数量
			Price:                buffetCustomer.TotalPrice,            // 自助餐顾客类型总价（单个商品、折后）
			CanReturnAmount:      buffetCustomer.CanReturnAmount,       // 自助餐顾客类型可退款金额
			CurrencyUnit:         currencyUnit,
		})
	}

	// 获取销售订单的加钟商品列表
	delayProducts := saleOrder.GetDelayProductList()
	for _, delayProduct := range delayProducts {
		products = append(products, resp.OrderReturnProduct{
			SaleOrderProductUuid: delayProduct.Uuid,
			LocaleName:           delayProduct.LocaleName,
			LocaleAttributeName:  delayProduct.LocaleAttributeName,
			Num:                  float64(delayProduct.CanReturnNum), // 加钟商品可退货数量
			Price:                delayProduct.UnitPrice,             // 加钟商品单价
			CanReturnAmount:      delayProduct.CanReturnAmount,       // 加钟商品可退款金额
			CurrencyUnit:         currencyUnit,
		})
	}

	// 获取销售订单商品列表
	for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
		if saleOrderProduct.IsCancelProduct() || saleOrderProduct.IsGiftProduct() || saleOrderProduct.Status == constant.OrderProductStatusUnSending {
			continue
		}
		if saleOrderProduct.IsPackageSubProduct() {
			continue
		}
		products = append(products, resp.OrderReturnProduct{
			SaleOrderProductUuid: saleOrderProduct.Uuid,
			LocaleName:           saleOrderProduct.MultiLanguageName.GetNames(),
			LocaleAttributeName:  saleOrderProduct.GetAttributeName(),
			Num:                  saleOrderProduct.GetCanReturnNum(), // 可退货数量=订单商品数量-已退货数量
			NumType:              saleOrderProduct.NumType,
			Price:                saleOrderProduct.TotalPrice,
			CanReturnAmount:      saleOrderProduct.GetCanReturnPrice(),
			CurrencyUnit:         currencyUnit,
		})
	}

	// 过滤掉单价为0的商品
	productList := make([]resp.OrderReturnProduct, 0)
	for _, product := range products {
		if product.Price == 0 {
			continue
		}
		productList = append(productList, product)
	}

	// 获取销售订单付款单列表
	// 可退款金额
	canReturnAmount := saleOrder.GetCanReturnAmount()
	res := &resp.OrderReturnInfoResp{
		ManualReturnPoints: saleOrder.CanManualReturnPoints(), // 是否可以手动退款积分。订单是按比例赠送积分且未发生积分抵扣时，不自动退款。
		DeductiblePoints:   saleOrder.GetManualReturnPoints(), // 可扣除积分。订单赠送的积分-已经退回的积分
		CanReturnAmount:    canReturnAmount,                   // 可退款金额. 可退款金额=订单最终应收金额-已退款金额
		PaymentRecords:     paymentRecords,
		Products:           productList,
	}

	return res, nil
}

// GetReverseSettleInfo 获取反结账信息
func (s *orderSrv) GetReverseSettleInfo(ctx context.Context, req req.OrderReverseSettleInfoReq) (*resp.OrderReverseSettleInfoResp, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)
	// 获取销售账单信息
	saleBill, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	var resDesks *resp.OrderReverseSettleDeskList
	if saleBill.IsDeskSaleBill() {
		desk := saleBill.Desk
		desks := make([]resp.OrderReverseSettleDesk, 0)
		// 如果原桌台空闲
		if desk.IsAvailableDesk() {
			desks = append(desks, resp.OrderReverseSettleDesk{
				Uuid:     desk.Uuid,
				SerialNo: desk.DeskNo,
			})
		}
		// 如果原桌台不空闲
		if !desk.IsAvailableDesk() {
			// 获取所有空闲的桌台
			freeDesks, err := repository.NewDeskRepo(db).GetAvailableDeskList()
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			for _, freeDesk := range freeDesks {
				desks = append(desks, resp.OrderReverseSettleDesk{
					Uuid:     freeDesk.Uuid,
					SerialNo: freeDesk.DeskNo,
				})
			}
		}
		resDesks = &resp.OrderReverseSettleDeskList{
			OriginDeskAvailable: desk.IsAvailableDesk(),
			List:                desks,
		}
	}

	var hasInstantOrder *bool
	if !saleBill.IsDeskSaleBill() {
		// 判断该收银机是否有未挂单的点餐订单
		_, hasInstantOrderBool, err := HasInstantOrder(ctx, db)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		hasInstantOrder = &hasInstantOrderBool
	}

	// 获取支付方式名称列表
	payMethods := saleBill.GetPaymentMethodNameList(ctx.GetLanguage())

	return &resp.OrderReverseSettleInfoResp{
		SaleBillUuid:    saleBill.Uuid,
		SaleBillNo:      saleBill.OrderNo,
		SaleBillType:    saleBill.BillType,
		OrderAmount:     saleBill.OriginAmount,
		PaymentAmount:   saleBill.PaymentAmount,
		PayMethods:      payMethods,
		Desks:           resDesks,
		HasInstantOrder: hasInstantOrder,
	}, nil
}

// ReverseSettle 处理反结账
func (s *orderSrv) ReverseSettle(ctx context.Context, request req.OrderReverseSettleReq) error {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(request.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取门店设置
	storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
	if err != nil {
		logger.Logger.Info("SubscribeCheckoutSaleOrderEvent process, GetStoreSetting failed", zap.Error(err))
		return errors.WithMessage(err)
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	orderRepo := repository.NewOrderRepo(db)
	// 获取销售账单信息
	saleBill, err := orderRepo.GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if saleBill.IsDeskSaleBill() {
		if request.DeskUuid == 0 {
			return errors.WithMessage(errors.New("桌台UUID不能为0"))
		}
	}

	// 销售账单状态变为未结账状态
	// 销售订单状态变为未结账状态
	// 销售订单的所有付款单都退款，并生成退款单
	// 反结账次数+1
	saleBill.SetReverseSettle()

	// 如果销售账单是桌台订单，则开桌
	// 开桌
	var desk *model.Desk
	if saleBill.IsDeskSaleBill() {
		deskRepo := repository.NewDeskRepo(db)
		desk, err = deskRepo.GetDeskRecord(request.DeskUuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		if !desk.IsAvailableDesk() {
			return errors.WithMessage(errors.New("桌台非空闲"))
		}
		desk.SetOpenDesk(saleBill.Uuid)
		saleBill.DeskUuid = desk.Uuid
		saleBill.SerialNo = desk.DeskNo
	}

	// 如果销售账单是点餐订单，则如果存在未挂单的点餐订单，根据参数决定是否挂单
	var hideSaleBill *model.SaleBill
	if !saleBill.IsDeskSaleBill() {
		saleBill, hasInstantOrder, err := HasInstantOrder(ctx, db)
		if err != nil {
			return errors.WithMessage(err)
		}
		// 如果存在未挂单的点餐订单，则根据参数决定是否挂单
		if hasInstantOrder {
			if request.HideOrder {
				hideSaleBill = saleBill
				hideSaleBill.SetHideSaleBill()
			} else {
				// 如果存在未挂单的点餐订单
				return errors.WithMessage(errors.New("存在未挂单的点餐订单，请先挂单"))
			}
		}
	}

	// 构建入库单，将账单的商品重新入库.
	// 出库记录标记为已撤销，并生成入库单将库存退还
	// 构建出库单，将账单下单减库存的商品出库
	if err := s.returnInventory(ctx, saleBill, WithReverseSettle()); err != nil {
		return errors.WithMessage(err)
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 删除销售订单原料
		if err := repository.NewSaleOrderMaterialRepo(db).DeleteSaleOrderMaterial(saleBill.Uuid); err != nil {
			return errors.WithMessage(err)
		}
		// 如果销售订单是免单，删除免单原因
		for _, saleOrder := range saleBill.SaleOrders {
			if saleOrder.IsFreeSaleOrder() {
				saleOrder.SetCancelFreeOrder()
				// 删除销售订单的免单原因
				if err := repository.NewSaleOrderProductReasonRepo(db).DeleteFreeReason(saleOrder.Uuid); err != nil {
					return errors.WithMessage(err)
				}
			}
		}

		// 如果存在需要挂单的销售账单，则更新该销售账单
		if hideSaleBill != nil {
			if err := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*hideSaleBill); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 更新高峰时段
		if err := repository.NewSaleOrderPeakTimeRepo(db).Record("dec", saleBill, 0, storeSetting.TimeZone); err != nil {
			return errors.WithMessage(err)
		}
		// 更新销售账单
		if err := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); err != nil {
			return errors.WithMessage(err)
		}

		giftPointsMap := make(map[uint64]float64) // sale_order_uuid -> gift_points
		// 更新销售订单
		for _, saleOrder := range saleBill.SaleOrders {
			giftPointsMap[saleOrder.Uuid] = saleOrder.GiftPoints // 清空之前记录的赠送积分
			// 将结账才记录的值清空
			saleOrder.ClearSettleInfo()
			if err := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 更新支付订单,状态为已退款
		for _, saleOrder := range saleBill.SaleOrders {
			for _, paymentOrder := range saleOrder.PaymentOrders {
				if err := repository.NewPaymentOrderRepo(db).UpdatePaymentOrderRecord(*paymentOrder); err != nil {
					return errors.WithMessage(err)
				}
			}
		}
		// 生成退款单
		for _, saleOrder := range saleBill.SaleOrders {
			isUseMember := saleOrder.ConsumerUuid != 0
			for _, paymentOrder := range saleOrder.PaymentOrders {
				refundOrder := paymentOrder.RefundOrder
				if err := repository.NewPaymentOrderRepo(db).CreateRefundOrderRecord(*refundOrder); err != nil {
					return errors.WithMessage(err)
				}
				// 如果是余额支付，则退款到余额
				if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeBalance {
					s.memberSrv.HandleMemberBalance(ctx, MemberBalanceChangeReq{
						MemberUuid:  saleOrder.ConsumerUuid,
						Money:       paymentOrder.BalanceAmount,
						GiftMoney:   paymentOrder.GiftBalanceAmount,
						Scene:       constant.MemberBalanceLogReverse,
						Describe:    fmt.Sprintf("订单反结账：%s", saleOrder.OrderNo),
						RelatedUuid: saleOrder.Uuid,
					})
				}
				// 取现金，更新钱箱
				if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeCash {
					if err := s.cashBoxSrv.UpdateBalance(ctx, UpdateCashBalanceParam{
						Amount:    -paymentOrder.Amount,
						Scene:     constant.CashBoxLogSceneRefund,
						OrderUuid: saleOrder.Uuid,
					}); err != nil {
						return errors.WithMessage(err)
					}
				}
			}
			// 退会员的累计消费金额
			if isUseMember {
				// 减少会员累计消费金额
				if err := repository.NewMemberRepo(db).DecConsumptionAmount(saleOrder.ConsumerUuid, saleOrder.GetCanReturnMemberConsumptionAmountMax()); err != nil {
					return errors.WithMessage(err)
				}
				// 减少会员累计消费次数
				if err := repository.NewMemberRepo(db).DecConsumptionCount(saleOrder.ConsumerUuid); err != nil {
					return errors.WithMessage(err)
				}
			}
			// 退积分
			if isUseMember {
				points := giftPointsMap[saleOrder.Uuid]
				member, err := repository.NewMemberRepo(db).GetMemberByUuid(saleOrder.ConsumerUuid)
				if err != nil {
					return errors.WithMessage(err)
				}
				// 如果订单有积分抵扣，则已抵扣的积分
				if saleBill.SaleBillSetting.IsOpenPointsExchange() && saleOrder.PayPoints > 0 {
					// 更新会员积分
					member.FrozenPoint = member.FrozenPoint + saleOrder.PayPoints // 退回已抵扣的积分
					if err := repository.NewMemberRepo(db).Update(saleOrder.ConsumerUuid, map[string]any{
						"frozen_point": member.FrozenPoint, // 退回已抵扣的积分
					}); err != nil {
						return errors.WithMessage(err)
					}
					// 创建积分变动记录
					memberPointLog := saleOrder.NewReverseSettleExchangeMemberPointLog(saleOrder.PayPoints)
					if _, err := repository.NewMemberPointLogRepo(db).Create(*memberPointLog); err != nil {
						return errors.WithMessage(err)
					}
				}
				if points > 0 {
					// 如果会员积分余额不足时，仅扣完余额
					if member.GetPoints() < points {
						points = member.GetPoints()
					}
					// 更新会员积分
					if err := repository.NewMemberRepo(db).Update(saleOrder.ConsumerUuid, map[string]any{
						"frozen_point": member.FrozenPoint - points, // 扣减积分
					}); err != nil {
						return errors.WithMessage(err)
					}
					// 创建积分变动记录
					memberPointLog := saleOrder.NewReverseSettleMemberPointLog(-points)
					if _, err := repository.NewMemberPointLogRepo(db).Create(*memberPointLog); err != nil {
						return errors.WithMessage(err)
					}
				}
			}

			// 退优惠券。如果订单使用了优惠券，需要将优惠券退还给会员。如果使用了通用优惠券，则通用优惠券余量+1并生成记录
			if saleOrder.HasCoupon() {
				// 加锁, 避免并发问题,避免数量+1重复或失效
				lock.NewSystemLock().LockUuid(constant.LockNameActivityConsumption)
				defer lock.NewSystemLock().UnlockUuid(constant.LockNameActivityConsumption)
				for _, coupon := range saleOrder.Coupons {
					if !coupon.IsDelete() {
						if coupon.IsCommonCoupon() {
							commonCoupon, err := repository.NewMarketingCouponRepo(db).GetCouponByUuid(coupon.MarketingCouponUuid)
							if err != nil {
								return errors.WithMessage(err)
							}
							commonCoupon.Count = commonCoupon.Count + 1
							// 取消核销通用优惠券，数量+1
							if err := repository.NewMarketingCouponRepo(db).UpdateCommonCouponCountCancel(coupon.MarketingCouponUuid); err != nil {
								return errors.WithMessage(err)
							}
							// 创建通用优惠券记录，记录类型：反结账退还
							if err := repository.NewMarketingCouponRepo(db).CreateCommonCouponRecordCancel(coupon.MarketingCouponUuid, commonCoupon.Count); err != nil {
								return errors.WithMessage(err)
							}
						}
						if coupon.IsMemberCoupon() {
							// 取消核销会员优惠券
							if err := repository.NewMemberCouponRepo(db).CancelVerifyMemberCoupon(coupon.MemberCouponUuid); err != nil {
								return errors.WithMessage(err)
							}
							// 删除会员优惠券使用记录
							if err := repository.NewMemberCouponRepo(db).DeleteMemberCouponRecord(coupon.MemberCouponUuid); err != nil {
								return errors.WithMessage(err)
							}
						}
					}
				}
			}

			// 发布"会员余额变动"事件
			utils.Go(func() {
				s.bus.PublishChangeMemberBalanceEvent(event.ChangeMemberBalancePayload{
					BasePayload: event.BasePayload{ // 会员余额变动
						Ctx:          ctx,
						CompanyUuid:  ctx.GetCompanyUuid(),
						Source:       ctx.GetSource(),
						SaleBillUuid: request.SaleBillUuid,
						OperatorUuid: int64(ctx.GetStaffUuid()),
					},
				})
			})

			// 发布"会员积分变动"事件
			utils.Go(func() {
				s.bus.PublishChangeMemberPointsEvent(event.ChangeMemberPointsPayload{
					BasePayload: event.BasePayload{ // 会员积分变动
						Ctx:          ctx,
						CompanyUuid:  ctx.GetCompanyUuid(),
						Source:       ctx.GetSource(),
						SaleBillUuid: request.SaleBillUuid,
						OperatorUuid: int64(ctx.GetStaffUuid()),
					},
				})
			})

			// 发送短信
			if isUseMember {
				utils.Go(func() {
					newCtx := ctx.Copy()
					newCtx.SetDB(s.dbm.GetDB(newCtx.GetDbId()))
					// 获取最新的会员信息
					member, err := repository.NewMemberRepo(newCtx.GetDB()).GetMemberByUuid(saleOrder.ConsumerUuid)
					if err != nil {
						ctx.Log().Info("停止发送短信（消费反结账），获取会员失败", zap.Error(errors.WithMessage(err)))
					} else {
						refundAmount := float64(0)
						for _, paymentOrder := range saleOrder.PaymentOrders {
							// 如果是余额支付，则退款到余额
							if paymentOrder.PaymentMethod.Code == constant.PaymentMethodCodeBalance {
								refundAmount = decimal.NewFromFloat(paymentOrder.BalanceAmount).Add(decimal.NewFromFloat(paymentOrder.GiftBalanceAmount)).Truncate(2).InexactFloat64()
							}
						}
						if refundAmount > 0 {
							if member != nil {
								smsReq := sms.MemberOrderRefundRequest{
									Company:       ctx.GetCompany().Name,
									OrderRefund:   refundAmount,
									Balance:       member.GetBalanceAll(),
									PointsBalance: member.GetPoints(),
								}
								if err := s.smsSrv.SendMemberOrderRefundSMS(newCtx, member.Phone, &smsReq); err != nil {
									ctx.Log().Info("发送退款短信失败（消费反结账）", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq), zap.Error(errors.WithMessage(err)))
								} else {
									ctx.Log().Info("发送退款短信成功（消费反结账）", zap.String("phone", member.Phone), zap.Any("smsReq", smsReq))
								}
							}
						}
					}
				})
			}
		}

		utils.Go(func() {
			// 发布"反结账"操作事件
			var payTypes []event.PayType
			for _, order := range saleBill.SaleOrders {
				if order.IsFree == 1 {
					payTypes = append(payTypes, event.PayType{
						Name:  "免单",
						Value: constant.PaymentMethodCodeFreePay,
						Price: order.GetAmount(),
					})
				} else {
					if infoResp, err := s.InstantOrderPaymentInfo(ctx, saleBill, request.SaleBillUuid, order.Uuid); err == nil {
						for _, paymentOrder := range infoResp.PaymentOrders.List {
							payTypes = append(payTypes, event.PayType{
								Name:           paymentOrder.PaymentMethodName,
								Value:          paymentOrder.PaymentMethodCode,
								DisabledCancel: utils.BoolToUint(paymentOrder.DisabledCancel),
								Price:          paymentOrder.Amount,
								FeeMoney:       paymentOrder.PaymentCommissionFee,
							})
						}
					}
				}
			}
			s.bus.PublishOrderReverseSettleEvent(event.OrderReverseSettlePayload{
				BasePayload: event.BasePayload{ // 订单反结账
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: saleBill.Uuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
				PayTypes: payTypes,
			})
		})

		// 更新桌台
		if saleBill.IsDeskSaleBill() {
			if err := repository.NewDeskRepo(db).UpdateDeskRecord(*desk); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 更新自助餐销量
		if saleBill.IsBuffetSaleBill() {
			if saleBill.BuffetPackage1Uuid != 0 {
				saleNum := saleBill.GetBuffetSaleNum(saleBill.BuffetPackage1Uuid)
				if err := repository.NewBuffetRepo(db).SubActualSaleNum(saleBill.BuffetPackage1Uuid, saleNum); err != nil {
					fmt.Println(err)
					ctx.Log().Error("SubActualSaleNum", zap.Error(fmt.Errorf("%s %s", ctx.GetRequestUuid(), err)))
				}
			}
			if saleBill.BuffetPackage2Uuid != 0 {
				saleNum := saleBill.GetBuffetSaleNum(saleBill.BuffetPackage2Uuid)
				if err := repository.NewBuffetRepo(db).SubActualSaleNum(saleBill.BuffetPackage2Uuid, saleNum); err != nil {
					ctx.Log().Error("SubActualSaleNum", zap.Error(fmt.Errorf("%s %s", ctx.GetRequestUuid(), err)))
				}
			}
		}

		utils.Go(func() {
			// 发布"统计"操作事件
			s.bus.PublishStatisticsSaleEvent(event.StatisticsSalePayload{
				BasePayload: event.BasePayload{ // 统计
					Ctx: ctx,
				},
				SaleBillUuid: saleBill.Uuid,
				OnlyDelete:   true,
			})
		})

		// 在ERP取消发票
		company := ctx.GetCompany()
		companySetting := ctx.GetCompanySetting()
		if company.IsOpenErpPhase3() && companySetting.ErpnextSiteCode != "" {
			staff := ctx.GetStaff()
			shiftLogRepo := repository.NewShiftLogRepo(db)
			shiftLog, err := shiftLogRepo.GetShiftLog(
				repository.CommonRepo.WhereByStaffUuid(staff.Uuid),
				repository.CommonRepo.WhereByShiftNo(staff.DutyNo),
			)
			if err != nil {
				return errors.WithMessage(err)
			}
			if shiftLog.IsHandedOver() {
				return errors.New("当前班次已交班，无法保存发票")
			}
			erpSrv := erp.NewIErpSrv(s.dbm)
			for _, saleOrder := range saleBill.SaleOrders {
				if saleOrder.IsDelete() {
					continue
				}
				err := erpSrv.CancelPosInvoice(ctx, req.CancelPosInvoiceReq{
					ProductsInvoiceName: saleOrder.ErpProductsInvoiceName,
					MaterialInvoiceName: saleOrder.ErpMaterialInvoiceName,
					OpenPosEntryName:    shiftLog.ErpnextOpenPosEntryName, //异步模式必填
					OrderNo:             saleOrder.OrderNo,                //异步模式必填
				})
				if err != nil {
					return errors.WithMessage(err)
				}
			}
		}

		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

// OrderRemark 修改订单备注
func (s *orderSrv) OrderRemark(ctx context.Context, req req.OrderRemarkReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	if req.SaleBillUuid == 0 {
		return nil, errors.New("请先下单再整单备注")
	}
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 判断订单状态
	if err := billInfo.ValidateOrderStatus(ctx.GetSource(), constant.OrderOrderRemark, 0); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 查询整单备注信息
	orderRemarkList, err := base.NewOrderRemarkRepo(db).GetOrderRemarkListByUuids(req.RemarkUuids)
	if err != nil {
		return nil, errors.WithMessage(err, "查询整单备注信息失败")
	}
	if len(orderRemarkList) != len(req.RemarkUuids) {
		return nil, errors.WithMessage(errors.New("整单备注信息不存在"), "整单备注信息不存在")
	}

	// 整单备注信息
	orderRemark, err := billInfo.GetOrderRemark()
	if err != nil {
		return nil, errors.WithMessage(err, "获取整单备注信息失败")
	}

	if req.IsNullRemark() && orderRemark != nil { // 历史备注信息存在，但是没有选择整单备注或输入整单备注文本
		// 划线所有历史备注
		for i := range orderRemark.List {
			orderRemark.List[i].IsLatest = false
		}
		billInfo.OrderRemark = orderRemark.ToJson()
	} else {
		// 创建新的备注信息
		orderRemarkItem := resp.OrderRemarkItem{
			IsLatest: true,
			Uuids:    req.RemarkUuids,
			Remark:   req.Remark,
			Remarks: func() []dto.LocaleResponse {
				remarks := make([]dto.LocaleResponse, 0)
				for _, remark := range orderRemarkList {
					remarks = append(remarks, remark.MultiLanguageName.GetNames())
				}
				return remarks
			}(),
			CreateTime: time.Now().Unix(),
		}
		if orderRemark != nil {
			// 有历史备注信息
			// 修改历史备注信息为不是最新
			for i := range orderRemark.List {
				orderRemark.List[i].IsLatest = false
			}
			orderRemark.List = append(orderRemark.List, orderRemarkItem)
			billInfo.OrderRemark = orderRemark.ToJson()
		} else {
			// 没有历史备注信息
			orderRemarkInfo := &resp.OrderRemarkInfo{
				List: []resp.OrderRemarkItem{orderRemarkItem},
			}
			billInfo.OrderRemark = orderRemarkInfo.ToJson()
		}
	}
	// 修改订单备注
	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if err := repository.NewOrderRepo(db).UpdateSaleBillOrderRemark(req.SaleBillUuid, billInfo.OrderRemark); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// CreateSaleBillSetting 创建销售账单设置
// isMember 是否是会员端订单
func (s *orderSrv) CreateSaleBillSetting(ctx context.Context, db *gorm.DB, saleBillUuid uint64, deskUuid uint64, isMember bool) (*model.SaleBillSetting, error) {
	saleBillSetting, err := s.NewSaleBillSetting(ctx, saleBillUuid, deskUuid, isMember)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	newSaleBillSetting, err := repository.NewOrderRepo(db).CreateSaleBillSetting(*saleBillSetting)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &newSaleBillSetting, nil
}

// CheckAuthorization 检查授权
// 检查当前员工是否有权限进行敏感操作（折扣或退款）
// operationType: "discount" - 折扣操作（整单改价、打折、抹零），"refund" - 退款操作
func (s *orderSrv) CheckAuthorization(ctx context.Context, operationType string) (bool, error) {
	// 1. 获取当前员工信息
	currentStaff := ctx.GetStaff()
	if currentStaff.Uuid == 0 {
		return false, errors.New("未找到当前员工信息")
	}
	// 如果是助手端,需要获取当前助手端登陆的员账号. 当前使用的token是收银机账户的
	if ctx.GetSource() == constant.SourceAssistant {
		// 1. 获取当前员工信息
		currentStaffUuid := ctx.GetAssistantUuid()
		staff, err := repository.NewStaffRepo(ctx.GetDB()).GetStaff(repository.CommonRepo.WhereByUuid(currentStaffUuid))
		if err != nil {
			return false, errors.WithMessage(err, "获取当前员工信息失败")
		}
		currentStaff = staff
	}

	// 2. 获取业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return false, errors.WithMessage(err, "获取业务设置失败")
	}

	// 3. 根据操作类型选择对应的配置
	var needPassword bool
	var authorizedStaffIds []uint64

	if operationType == string(SensitiveOperationTypeRefund) {
		// 退款操作：使用退款授权列表
		needPassword = businessSetting.RefundNeedPassword == "1"
		authorizedStaffIds = businessSetting.RefundAuthorizedStaffIds
	} else {
		// 折扣操作：使用折扣授权列表（默认）
		needPassword = businessSetting.DiscountNeedPassword == "1"
		authorizedStaffIds = businessSetting.DiscountAuthorizedStaffIds
	}

	// 4. 如果未开启密码验证，返回有权限
	if !needPassword {
		return true, nil
	}

	// 5. 检查当前员工是否在授权名单中
	for _, staffId := range authorizedStaffIds {
		if staffId == currentStaff.Uuid {
			return true, nil
		}
	}

	// 6. 不在授权名单中，返回无权限
	return false, nil
}

// VerifyPassword 密码验证
// 验证授权员工账号和密码（根据操作类型选择折扣操作或退款操作）
func (s *orderSrv) VerifyPassword(ctx context.Context, req req.VerifyPasswordForSensitiveOperationReq) (bool, error) {
	// 根据操作类型选择不同的验证方法
	if req.OperationType == string(SensitiveOperationTypeRefund) {
		return s.verifyPasswordForRefund(ctx, req)
	}
	// 默认使用折扣操作的验证逻辑
	return s.verifyPasswordForDiscount(ctx, req)
}

// verifyPasswordForDiscount 折扣操作密码验证
// 验证授权员工账号和密码（用于折扣操作：整单改价、打折、抹零）
func (s *orderSrv) verifyPasswordForDiscount(ctx context.Context, req req.VerifyPasswordForSensitiveOperationReq) (bool, error) {
	// 1. 根据账号（邮箱或手机号）查找员工
	db := s.dbm.GetDB(ctx.GetDbId())
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUsername(req.AuthorizedStaffAccount))
	if err != nil || staff.Uuid == 0 {
		if err != nil {
			ctx.Log().Info("verifyPasswordForDiscount", zap.Any("companyUuid", ctx.GetCompanyUuid()), zap.Any("req", req), zap.Error(errors.WithMessage(err)))
		}
		return false, errors.WithMessage(errors.New("不是权限员工，请确认信息"), err.Error())
	}

	// 2. 获取业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return false, errors.WithMessage(err, "获取业务设置失败")
	}

	// 3. 检查员工是否在授权名单中（折扣操作的授权名单）
	authorizedStaffIds := businessSetting.DiscountAuthorizedStaffIds

	isAuthorized := false
	for _, staffId := range authorizedStaffIds {
		if staffId == staff.Uuid {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		return false, errors.WithMessage(errors.New("不是权限员工，请确认信息"), "不是权限员工，请确认信息")
	}

	// 4. 验证密码（从 staff 表中读取权限密码 permission_password，加密后比较）
	encryptedPassword := utils.EncryptPassword(req.Password)
	if staff.PermissionPassword != encryptedPassword {
		return false, errors.WithMessage(errors.New("密码错误"), "密码错误")
	}

	// 5. 返回验证结果
	return true, nil
}

// VerifyPasswordForRefund 退款密码验证
// 验证授权员工账号和密码（用于退款操作）
func (s *orderSrv) verifyPasswordForRefund(ctx context.Context, req req.VerifyPasswordForSensitiveOperationReq) (bool, error) {
	// 1. 根据账号（邮箱或手机号）查找员工
	db := s.dbm.GetDB(ctx.GetDbId())
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUsername(req.AuthorizedStaffAccount))
	if err != nil || staff.Uuid == 0 {
		if err != nil {
			ctx.Log().Info("VerifyPasswordForRefund", zap.Any("companyUuid", ctx.GetCompanyUuid()), zap.Any("req", req), zap.Error(errors.WithMessage(err)))
		}
		return false, errors.New("不是权限员工，请确认信息")
	}

	// 2. 获取业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return false, errors.WithMessage(err, "获取业务设置失败")
	}

	// 3. 检查员工是否在授权名单中（退款操作的授权名单）
	authorizedStaffIds := businessSetting.RefundAuthorizedStaffIds

	isAuthorized := false
	for _, staffId := range authorizedStaffIds {
		if staffId == staff.Uuid {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		return false, errors.New("不是权限员工，请确认信息")
	}

	// 4. 验证密码（从 staff 表中读取权限密码 permission_password，加密后比较）
	encryptedPassword := utils.EncryptPassword(req.Password)
	if staff.PermissionPassword != encryptedPassword {
		return false, errors.New("密码错误")
	}

	// 5. 返回验证结果
	return true, nil
}

// SensitiveOperationType 敏感操作类型
type SensitiveOperationType string

const (
	SensitiveOperationTypeDiscount SensitiveOperationType = "discount" // 折扣操作（整单改价、打折、抹零）
	SensitiveOperationTypeRefund   SensitiveOperationType = "refund"   // 退款操作
)

// AuthorizeSensitiveOperation 授权敏感操作
// 统一的授权验证方法，用于折扣操作和退款操作
// 返回授权员工信息，如果当前员工有权限则返回当前员工信息
func (s *orderSrv) AuthorizeSensitiveOperation(ctx context.Context, operationType SensitiveOperationType, authorizedStaffAccount, authorizedStaffPassword string) (*model.Staff, error) {
	// 如果提供了授权参数，需要验证
	if authorizedStaffAccount != "" && authorizedStaffPassword != "" {
		verifyReq := req.VerifyPasswordForSensitiveOperationReq{
			OperationType:          string(operationType),
			AuthorizedStaffAccount: authorizedStaffAccount,
			Password:               authorizedStaffPassword,
		}

		// 使用统一的验证方法，内部会根据 operation_type 选择不同的验证逻辑
		verified, err := s.VerifyPassword(ctx, verifyReq)

		if err != nil {
			return nil, errors.WithMessage(err, "授权验证失败")
		}
		if !verified {
			return nil, errors.New("授权验证失败")
		}

		// 获取授权员工信息
		db := s.dbm.GetDB(ctx.GetDbId())
		staffRepo := repository.NewStaffRepo(db)
		staff, err := staffRepo.GetStaff(staffRepo.WhereUsername(authorizedStaffAccount))
		if err != nil || staff.Uuid == 0 {
			return nil, errors.New("获取授权员工信息失败")
		}
		return &staff, nil
	}

	// 如果没有提供授权参数，检查当前员工是否在授权名单中
	var hasPermission bool
	var err error

	if operationType == SensitiveOperationTypeRefund {
		hasPermission, err = s.CheckAuthorization(ctx, string(SensitiveOperationTypeRefund))
	} else {
		hasPermission, err = s.CheckAuthorization(ctx, string(SensitiveOperationTypeDiscount))
	}

	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if !hasPermission {
		return nil, errors.New("需要授权验证")
	}

	// 当前员工有权限，返回当前员工信息
	staff := ctx.GetStaff()
	// 如果是助手端,需要获取当前助手端登陆的员账号. 当前使用的token是收银机账户的
	if ctx.GetSource() == constant.SourceAssistant {
		// 1. 获取当前员工信息
		assistantStaffUuid := ctx.GetAssistantUuid()
		assistantStaff, err := repository.NewStaffRepo(ctx.GetDB()).GetStaff(repository.CommonRepo.WhereByUuid(assistantStaffUuid))
		if err != nil {
			return nil, errors.WithMessage(err, "获取当前员工信息失败")
		}
		staff = assistantStaff
	}
	return &staff, nil
}

// GetPaymentAmount 获取实付金额。实付金额=订单金额-退款金额
func (s *orderSrv) GetPaymentAmount(ctx context.Context, req req.OrderPaymentAmountReq) resp.GetPaymentAmountResp {
	if len(req.SaleBillUuids) == 0 {
		return resp.GetPaymentAmountResp{PaymentAmount: 0}
	}

	db := ctx.GetDB()
	if db == nil {
		db = s.dbm.GetDB(ctx.GetDbId())
	}

	type amountAgg struct {
		Amount float64 `gorm:"column:amount"`
	}

	var (
		payAgg    amountAgg
		refundAgg amountAgg
	)

	// 统计销售订单的支付金额
	if err := db.Model(&model.SaleOrder{}).
		Select("COALESCE(SUM(payment_amount), 0) AS amount").
		Where("sale_bill_uuid IN (?)", req.SaleBillUuids).
		Where("delete_time = ?", constant.NotDeleted).
		Where("status = ?", constant.SaleOrderStatusFinish).
		Scan(&payAgg).Error; err != nil {
		logger.Logger.Error("GetPaymentAmount sum payment failed", zap.Error(err))
		return resp.GetPaymentAmountResp{PaymentAmount: 0}
	}

	// 统计退款金额
	if err := db.Table("ttpos_return_order AS ro").
		Select("COALESCE(SUM(ro.refund_amount), 0) AS amount").
		Joins("INNER JOIN ttpos_sale_order AS so ON ro.related_order_uuid = so.uuid AND so.delete_time = ? AND so.status = ?", constant.NotDeleted, constant.SaleOrderStatusFinish).
		Where("ro.delete_time = ?", constant.NotDeleted).
		Where("ro.related_order_type = ?", constant.ReturnOrderRelatedOrderTypeSaleOrder).
		Where("so.sale_bill_uuid IN (?)", req.SaleBillUuids).
		Scan(&refundAgg).Error; err != nil {
		logger.Logger.Error("GetPaymentAmount sum refund failed", zap.Error(err))
		return resp.GetPaymentAmountResp{PaymentAmount: 0}
	}

	paymentAmount := decimal.NewFromFloat(payAgg.Amount).
		Sub(decimal.NewFromFloat(refundAgg.Amount)).
		Round(2).InexactFloat64()

	if paymentAmount < 0 {
		paymentAmount = 0
	}

	return resp.GetPaymentAmountResp{
		PaymentAmount: paymentAmount,
	}
}
