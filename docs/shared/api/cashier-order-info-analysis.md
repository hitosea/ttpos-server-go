# 收银端订单信息获取逻辑分析

> 本文档分析收银端（cashier）订单信息获取的完整逻辑流程，包括订单列表和订单详情的实现细节。

**文件位置**: `main/app/api/v1/cashier/cashier_order.go`  
**创建时间**: 2025-01-27  
**维护者**: TTPOS Team

---

## 目录

- [API 端点](#api-端点)
- [1. 获取订单列表](#1-获取订单列表)
- [2. 获取订单详情](#2-获取订单详情)
- [数据流转图](#数据流转图)
- [关键逻辑说明](#关键逻辑说明)
- [数据结构说明](#数据结构说明)

---

## API 端点

### 1. 获取订单列表

**端点**: `GET /cashier/order/list`

**路由注册**: ```494:494:main/app/api/v1/cashier/cashier_order.go
privateApi.GET("/order/list", wrapper.GetCashierOrderList)
```

**Handler**: `GetCashierOrderList`

### 2. 获取订单详情

**端点**: `GET /cashier/order/info`

**路由注册**: ```495:495:main/app/api/v1/cashier/cashier_order.go
privateApi.GET("/order/info", wrapper.GetOrderInfo)
```

**Handler**: `GetOrderInfo`

---

## 1. 获取订单列表

### 1.1 Handler 层处理

**文件**: `main/app/api/v1/cashier/cashier_order.go`

**处理流程**:

```37:54:main/app/api/v1/cashier/cashier_order.go
func (h *OrderHandler) GetCashierOrderList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, dto.PageReqMessage)
		return
	}
	// 获取产品列表
	res, err := h.orderSrv.GetOrderLists(ctx, req)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}
```

**步骤说明**:

1. **获取上下文**: 从 Gin Context 中提取业务上下文（包含公司信息、员工信息、时区等）
2. **参数绑定**: 使用 `ShouldBindQuery` 绑定查询参数到 `req.OrderListReq` 结构体
3. **参数验证**: 如果绑定失败，返回验证错误信息
4. **调用服务层**: 调用 `orderSrv.GetOrderLists` 获取订单列表数据
5. **错误处理**: 如果服务层返回错误，统一处理并返回错误响应
6. **返回结果**: 成功时返回订单列表数据

### 1.2 请求参数结构

**文件**: `main/app/dto/req/order.go`

```22:36:main/app/dto/req/order.go
type OrderListReq struct {
	dto.PageReq                // 分页参数
	OrderNo             string `form:"order_no"`                         // 订单编号
	DateType            int    `form:"date_type,default=-1"`             // 日期类型 -1=全都、 0=今天、 1=昨天、 2=本周、3=本月、4=本年、5=近7天、6=上个月
	EnableCreateTime    bool   `form:"enable_create_time"`               // 启用开台时间 false-不启用，true-启用
	EnablePayTime       bool   `form:"enable_pay_time"`                  // 启用支付时间 false-不启用，true-启用
	QueryStartTime      uint   `form:"query_start_time"`                 // 查询开始时间戳
	QueryEndTime        uint   `form:"query_end_time"`                   // 查询结束时间戳
	Status              int    `form:"status,default=-1"`                // 账单状态, -1=全都、 0=待付款、1=已完成、2=已取消
	BillType            int    `form:"bill_type,default=-1"`             // 账单类型, -1=全都、 0=Desk桌台订单、1=OrderingFood点餐订单
	DiningMethod        int    `form:"dining_method,default=-1"`         // 用餐方式, -1=全都、 0-堂食 1-打包
	SaleBillUuids       string `form:"sale_bill_uuids"`                  // 销售账单UUID列表，多个UUID用逗号分隔
	IsOnlyDataManage    int    `form:"is_only_data_manage,default=0"`    // 是否只包含数据管理, 0-不包含、1-包含
	IsContainDataManage int    `form:"is_contain_data_manage,default=0"` // 是否包含数据管理, 0-不包含、1-包含
}
```

**参数说明**:

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `page_no` | int | 页码 | 1 |
| `page_size` | int | 每页数量 | 20 |
| `order_no` | string | 订单编号（模糊查询） | "" |
| `date_type` | int | 日期类型：-1=全部、0=今天、1=昨天、2=本周、3=本月、4=本年、5=近7天、6=上个月 | -1 |
| `enable_create_time` | bool | 是否启用开台时间筛选 | false |
| `enable_pay_time` | bool | 是否启用支付时间筛选 | false |
| `query_start_time` | uint | 查询开始时间戳 | 0 |
| `query_end_time` | uint | 查询结束时间戳 | 0 |
| `status` | int | 账单状态：-1=全部、0=待付款、1=已完成、2=已取消 | -1 |
| `bill_type` | int | 账单类型：-1=全部、0=桌台订单、1=点餐订单 | -1 |
| `dining_method` | int | 用餐方式：-1=全部、0=堂食、1=打包 | -1 |
| `sale_bill_uuids` | string | 销售账单UUID列表（逗号分隔） | "" |
| `is_only_data_manage` | int | 是否只包含数据管理订单 | 0 |
| `is_contain_data_manage` | int | 是否包含数据管理订单 | 0 |

### 1.3 Service 层处理

**文件**: `main/app/service/order_manage.go`

**核心方法**: `GetOrderLists`

**处理流程**:

```38:213:main/app/service/order_manage.go
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
		num, _ := orderRepo.GetOrderNum(
			repository.CommonRepo.WhereByStatus(status),
			repository.CommonRepo.WhereBySoftDelete(),
			repository.CommonRepo.WhereByCooking(),
			repository.CommonRepo.WhereInBillType([]uint{constant.SaleBillTypeDesk, constant.SaleBillTypeInstant}),
			dbOption,
		)
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
```

**关键处理步骤**:

1. **初始化 Repository**: 根据上下文中的数据库ID获取对应的数据库连接，创建订单 Repository
2. **参数转换**: 使用 `copier` 将请求参数转换为 Repository 层需要的参数类型
3. **查询数据**: 调用 `GetCashierOrderListWithPagination` 从数据库获取订单列表数据
4. **数据组装**: 遍历订单列表，组装响应数据结构
   - **拆单处理**: 如果订单被拆单（`len(bill.SaleOrders) > 1`），需要单独处理每个子订单
   - **支付方式汇总**: 收集所有支付方式名称，去重后拼接
   - **会员信息汇总**: 收集所有会员UUID
   - **操作权限判断**: 根据订单状态和业务规则判断是否可退款、取消、反结账、打印等
5. **统计信息**: 统计待付款、已完成、已取消订单数量
6. **返回结果**: 组装分页响应对象并返回

### 1.4 Repository 层处理

**文件**: `main/app/repository/order.go`

**核心方法**: `GetCashierOrderListWithPagination`

**查询逻辑**:

```381:456:main/app/repository/order.go
// GetCashierOrderListWithPagination 获取收银台订单列表
func (r *orderRepo) GetCashierOrderListWithPagination(param GetCashierOrderListWithPaginationType, tz string) (lists []model.SaleBill, total int64, dbOption DBOption, err error) {
	// 额外条件
	dbOption = r.getOrderListDBOption(param, tz)

	opts := []DBOption{
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders.PaymentMethod",
			},
			WithPreload{
				Query: "SaleOrders.ReturnOrders",
			},
			WithPreload{
				Query: "SaleOrders.Member",
			},
			WithPreload{
				Query: "DataManage",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereByType(model.DataManageTypeOrder)),
				},
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByCooking(),
		CommonRepo.WhereInBillType([]uint{constant.SaleBillTypeDesk, constant.SaleBillTypeInstant}),
		CommonRepo.SortWithID("DESC"),
		dbOption,
		//
		func() DBOption {
			return func(db *gorm.DB) *gorm.DB {
				//  账单状态
				if param.Status != -1 {
					db = db.Where("status = ?", uint(param.Status))
				}
				//
				return db
			}
		}(),
	}
	if param.IsOnlyDataManage == 1 {
		uuidList := strings.Split(param.SaleBillUuids, ",")
		uuids := []uint64{}
		for _, uuid := range uuidList {
			uuid, _ := strconv.ParseUint(uuid, 10, 64)
			uuids = append(uuids, uint64(uuid))
		}
		opts = append(opts, CommonRepo.WhereInUuids(uuids))
	}
	if param.IsOnlyDataManage == 0 && param.IsContainDataManage == 0 {
		opts = append(opts,
			func() DBOption {
				return func(db *gorm.DB) *gorm.DB {
					return db.Where("uuid NOT IN (SELECT data_uuid FROM ttpos_data_manage WHERE type = ?)", model.DataManageTypeOrder)
				}
			}(),
		)
	}

	//
	lists, total, err = r.GetOrderListWithPagination(
		param.PageNo,
		param.PageSize,
		opts...,
	)
	if err != nil {
		return nil, 0, dbOption, fmt.Errorf("GetCashierOrderListWithPagination: %v", err)
	}
	return lists, total, dbOption, nil
}
```

**查询条件构建** (`getOrderListDBOption`):

```304:379:main/app/repository/order.go
func (r *orderRepo) getOrderListDBOption(param GetCashierOrderListWithPaginationType, tz string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		// 订单编号
		if param.OrderNo != "" {
			db = db.Where("order_no like ?", "%"+param.OrderNo+"%")
		}
		// 账单类型
		if param.BillType != -1 {
			db = db.Where("bill_type = ?", param.BillType)
		}
		if param.DiningMethod != -1 {
			db = db.Where("dining_method = ?", param.DiningMethod)
		}
		//  日期类型 -1-全都 1-今天 2-昨天 3-本周 4-本月 5-本年 6-近7天 7-上个月
		if param.DateType >= 0 && param.DateType <= 3 {
			var startTime, endTime int64
			switch param.DateType {
			case constant.OrderDateTypeToday: // 今天
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeToday)
			case constant.OrderDateTypeYesterday: // 昨天
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeYesterday)
			case constant.OrderDateTypeWeek: // 本周
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeThisWeek)
			case constant.OrderDateTypeMonth: // 本月
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeThisMonth)
			case constant.OrderDateTypeYear: // 本年
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeThisYear)
			case constant.OrderDateTypeLastWeek: // 近7天
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeLastWeek)
			case constant.OrderDateTypeLastMonth: // 上个月
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeLastMonth)
			}
			// db = db.Where("create_time BETWEEN ? AND ?", startTime, endTime)
			param.QueryStartTime = uint(startTime)
			param.QueryEndTime = uint(endTime)
		}
		// 日期范围
		if param.QueryStartTime != 0 || param.QueryEndTime != 0
```

**预加载关联数据**:

- `SaleOrders`: 销售订单列表（已软删除的除外）
- `SaleOrders.PaymentOrders.PaymentMethod`: 支付订单及支付方式
- `SaleOrders.ReturnOrders`: 退款订单
- `SaleOrders.Member`: 会员信息
- `DataManage`: 数据管理信息（仅订单类型）

**查询条件**:

- 软删除过滤: `delete_time = 0`
- 送厨状态过滤: 排除未送厨订单
- 账单类型过滤: 仅查询桌台订单（`SaleBillTypeDesk`）和点餐订单（`SaleBillTypeInstant`）
- 排序: 按ID降序（最新的在前）
- 状态过滤: 如果指定了状态，则按状态过滤
- 数据管理过滤: 根据 `IsOnlyDataManage` 和 `IsContainDataManage` 参数决定是否包含数据管理订单

### 1.5 响应结构

**文件**: `main/app/dto/resp/order.go`

```65:68:main/app/dto/resp/order.go
type OrderListPaginationResp struct {
	List []BillLists   `json:"list"` // 订单列表
	Meta OrderListMeta `json:"meta"` // Meta信息
}
```

```37:53:main/app/dto/resp/order.go
// BillLists 订单列表响应
type BillLists struct {
	SaleBillUuid  uint64           `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64           `json:"sale_order_uuid"` // 销售订单UUID,第一个销售订单的uuid
	BillType      uint             `json:"bill_type"`       // 订单类型	0:桌台订单 1:点餐订单
	IsSplit       bool             `json:"is_split"`        // 是否拆单	false:否 true:是
	SerialNo      string           `json:"serial_no"`       // 桌位编号 (点餐流水号)
	OrderNo       string           `json:"order_no"`        // 订单编号
	Status        uint             `json:"status"`          // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	FinishTime    int64            `json:"finish_time"`     // 完成时间（支付时间）（时间戳）
	OrderAmount   float64          `json:"order_amount"`    // 订单总金额
	PaymentAmount float64          `json:"payment_amount"`  // 支付金额
	PayTypeName   string           `json:"pay_type_name"`     // 支付类型名称
	ConsumerUuids string           `json:"consumer_uuids"`  // 会员id
	SaleOrders    []BillListsOrder `json:"sale_orders"`     // 订单列表
	Extra         BillListsExtra   `json:"extra,omitempty"` // 通过当前数据控制按钮是否显示
}
```

```55:62:main/app/dto/resp/order.go
type OrderListMeta struct {
	dto.PageResponse
	TotalNum      int64   `json:"total_num"`      // 总数量
	UnpaidNum     int64   `json:"unpaid_num"`     // 待付款数量
	CompleteNum   int64   `json:"complete_num"`   // 已完成数量
	CancelNum     int64   `json:"cancel_num"`     // 已取消数量
	PaymentAmount float64 `json:"payment_amount"` // 实付金额
}
```

---

## 2. 获取订单详情

### 2.1 Handler 层处理

**文件**: `main/app/api/v1/cashier/cashier_order.go`

**处理流程**:

```67:84:main/app/api/v1/cashier/cashier_order.go
func (h *OrderHandler) GetOrderInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderInfoReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 获取收银产品列表
	res, err := h.orderSrv.GetOrderInfos(ctx, req)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}
```

**步骤说明**:

1. **获取上下文**: 从 Gin Context 中提取业务上下文
2. **参数绑定**: 使用 `ShouldBindQuery` 绑定查询参数
3. **调用服务层**: 调用 `orderSrv.GetOrderInfos` 获取订单详情
4. **错误处理**: 统一处理错误并返回
5. **返回结果**: 成功时返回订单详情数据

### 2.2 请求参数结构

**文件**: `main/app/dto/req/order.go`

```38:42:main/app/dto/req/order.go
// OrderInfoReq 订单信息查询
type OrderInfoReq struct {
	SaleBillUuid  uint64 `form:"sale_bill_uuid" json:"sale_bill_uuid"`   // 销售账单UUID
	SaleOrderUuid uint64 `form:"sale_order_uuid" json:"sale_order_uuid"` // 销售订单UUID 当查看子订单信息的时候才需要传
}
```

**参数说明**:

| 参数 | 类型 | 说明 | 必填 |
|------|------|------|------|
| `sale_bill_uuid` | uint64 | 销售账单UUID | 是 |
| `sale_order_uuid` | uint64 | 销售订单UUID（查看子订单时必填） | 否 |

**使用场景**:

- **查看主单**: 只传 `sale_bill_uuid`，返回整个账单的所有订单信息
- **查看子订单**: 同时传 `sale_bill_uuid` 和 `sale_order_uuid`，只返回指定子订单的信息

### 2.3 Service 层处理

**文件**: `main/app/service/order_manage.go`

**核心方法**: `GetOrderInfos`

**处理流程**:

```394:693:main/app/service/order_manage.go
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
					// 如果是套餐商品，则获取各个子商品的名称、数量、规格、属性，如："牛排*1（标准，黑椒汁）；可乐*2（大杯，少冰）；沙拉*1（大份，沙拉酱，蜂蜜酱）"
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
			NationalityName: func() string {
				if saleBill.Nationality != nil {
					return saleBill.Nationality.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
				}
				return ""
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
```

**关键处理步骤**:

1. **获取订单数据**: 调用 `GetSaleBillInfo` 获取销售账单的完整信息（包含所有关联数据）
2. **判断查询类型**: 
   - `isMain`: 是否查询主单（`SaleOrderUuid == 0`）
   - `isSplit`: 是否拆单（`len(SaleOrders) > 1`）
3. **遍历销售订单**: 根据 `SaleOrderUuid` 参数决定返回所有订单还是单个订单
4. **组装商品列表**: 按顺序添加三类商品
   - **自助餐顾客**: `SaleOrderBuffetCustomerTypes`
   - **加钟商品**: `SaleOrderBuffetDelayProducts`
   - **正常商品**: `SaleOrderProducts`（需要过滤套餐子商品、未接单商品、取消订单时的未送厨商品）
5. **处理套餐商品**: 如果是套餐商品，需要组装子商品信息到属性名称中
6. **组装订单信息**: 构建 `OrderInfo` 对象，包含订单基本信息、金额、支付方式、会员信息、商品列表等
7. **获取操作日志**: 调用 `GetRecordList` 获取订单操作记录
8. **判断操作权限**: 根据订单状态和业务规则判断可执行的操作
9. **返回结果**: 组装完整的 `OrderInfosResp` 响应对象

### 2.4 Repository 层处理

**文件**: `main/app/repository/order.go`

**核心方法**: `GetSaleBillInfo`

**查询逻辑**:

```534:615:main/app/repository/order.go
// GetSaleBillInfo 获取销售账单详细信息
func (r *orderRepo) GetSaleBillInfo(saleBillUuid uint64, saleOrderUuid uint64) (model.SaleBill, error) {
	info, err := r.GetSaleBill(
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleBillSetting",
			},
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						db = db.Where("delete_time = ?", constant.NotDeleted)
						if saleOrderUuid > constant.OptionalUuid {
							db = db.Where("uuid = ?", saleOrderUuid)
						}
						return db
					},
				},
			},
			WithPreload{
				Query: "SaleOrders.ReturnOrders",
			},
			WithPreload{
				Query: "SaleOrders.FreeReasons.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.Member",
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ImageFile",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.CancelReasons.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ReturnOrderProducts",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetDelayProducts.ReturnOrderProducts",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.BuffetPackage.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.ReturnOrderProducts",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.BuffetCustomerTypePrice.BuffetCustomerType",
			},
			WithPreload{
				Query: "Desk",
			},
			WithPreload{
				Query: "OrderSource.MultiLanguageName",
			},
			WithPreload{
				Query: "Nationality.MultiLanguageName",
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByUuid(saleBillUuid),
	)
	if err != nil {
		return model.SaleBill{}, fmt.Errorf("GetSaleBillInfo: %v", err)
	}
	return info, nil
}
```

**预加载关联数据**:

- `SaleBillSetting`: 账单设置
- `SaleOrders`: 销售订单列表（可选择性过滤单个订单）
- `SaleOrders.ReturnOrders`: 退款订单
- `SaleOrders.FreeReasons.MultiLanguageName`: 免单原因（多语言）
- `SaleOrders.Member`: 会员信息
- `SaleOrders.PaymentOrders`: 支付订单
- `SaleOrders.SaleOrderProducts.*`: 订单商品及其关联数据（图片、取消原因、退款、BOM、属性等）
- `SaleOrders.SaleOrderBuffetDelayProducts.ReturnOrderProducts`: 加钟商品退款
- `SaleOrders.SaleOrderBuffetCustomerTypes.*`: 自助餐顾客类型及其关联数据
- `Desk`: 桌台信息
- `OrderSource.MultiLanguageName`: 订单来源（多语言）
- `Nationality.MultiLanguageName`: 国籍（多语言）

### 2.5 响应结构

**文件**: `main/app/dto/resp/order.go`

```180:186:main/app/dto/resp/order.go
type OrderInfosResp struct {
	Detail       OrderInfos `json:"detail"` // 明细
	OperationLog struct {
		List []OrderOperationLog `json:"list"`
	} `json:"operation_log"` // 操作日志
	Extra BillListsExtra `json:"extra"` // 通过当前数据控制按钮是否显示
}
```

```124:151:main/app/dto/resp/order.go
// OrderInfos 订单信息响应
type OrderInfos struct {
	SaleBillUuid    uint64              `json:"sale_bill_uuid"`    // 销售账单UUID
	BillType        uint                `json:"bill_type"`         // 订单类型	0:桌台订单 1:点餐订单 2:外送订单
	DiningMethod    uint                `json:"dining_method"`     // 用餐方式,0-堂食(店内就餐) 1-打包
	IsSplit         bool                `json:"is_split"`          // 是否拆单	false:否 true:是
	IsBuffet        bool                `json:"is_buffet"`         // 是否自助餐	false:否 true:是
	SerialNo        string              `json:"serial_no"`         // 桌位编号 (点餐流水号)
	OrderNo         string              `json:"order_no"`          // 订单编号
	Status          uint                `json:"status"`            // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	CreateTime      int64               `json:"create_time"`       // 创建时间
	FinishTime      int64               `json:"finish_time"`       // 完成时间（支付时间）（时间戳）
	OrderAmount     float64             `json:"order_amount"`      // 订单总金额
	PaymentAmount   float64             `json:"payment_amount"`    // 支付金额
	RefundAmount    float64             `json:"refund_amount"`     // 退款金额
	MemberNames     string              `json:"member_names"`      // 会员名称
	MemberUuids     string              `json:"member_uuids"`      // 会员名称
	BuffetNames     string              `json:"buffet_names"`      // 自助餐名称
	CancelReason    string              `json:"cancel_reason"`      // 取消原因
	CashierName     string              `json:"cashier_name"`      // 收银员名称
	Remark          string              `json:"remark"`            // 备注
	OrderSourceUuid uint64              `json:"order_source_uuid"` // 订单来源UUID（0=店内）
	OrderSourceName string              `json:"order_source_name"` // 订单来源名称
	NationalityUuid uint64              `json:"nationality_uuid"`  // 国籍UUID（0=未记录）
	NationalityName string              `json:"nationality_name"`  // 国籍名称
	PayTypes        []OrderInfoPayTypes `json:"pay_types"`         // 支付类型
	SaleOrders      []OrderInfo         `json:"sale_orders"`       // 订单列表
}
```

```104:122:main/app/dto/resp/order.go
// OrderInfo 订单信息响应
type OrderInfo struct {
	SaleOrderUuid uint64             `json:"sale_order_uuid"` // 销售订单UUID
	BillType      uint               `json:"bill_type"`       // 订单类型	0:桌台订单 1:点餐订单
	DiningMethod  uint               `json:"dining_method"`   // 用餐方式,0-堂食 1-打包
	SerialNo      string             `json:"serial_no"`       // 桌位编号 (点餐流水号)
	OrderNo       string             `json:"order_no"`        // 订单编号
	Status        uint               `json:"status"`          // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	FinishTime    int64              `json:"finish_time"`     // 完成时间（支付时间）（时间戳）
	OrderAmount   float64            `json:"order_amount"`    // 订单总金额
	PaymentAmount float64            `json:"payment_amount"`  // 支付金额
	RefundAmount  float64            `json:"refund_amount"`   // 退款金额
	PayTypeName   string             `json:"pay_type_name"`   // 支付类型名称
	MemberName    string             `json:"member_name"`     // 会员名称
	MemberUuid    uint64             `json:"member_uuid"`     // 会员名称
	IsFree        bool               `json:"is_free"`         // 是否免单
	FreeReason    dto.LocaleResponse `json:"free_reason"`     // 免单原因
	Products      []OrderProduct     `json:"products"`        // 产品列表
}
```

```82:102:main/app/dto/resp/order.go
type OrderProduct struct {
	Uuid                uint64             `json:"uuid"`                  // 销售订单商品ID
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 产品名称
	LocaleAttributeName dto.LocaleResponse `json:"locale_attribute_name"` // 口味名称
	Price               float64            `json:"price"`                 // 单价. 折前价
	Num                 float64            `json:"num"`                   // 数量
	SalePrice           float64            `json:"sale_price"`            // 销售价 (折前价) (划线价格) 当 sale_price 不等于 total_price 时才显示
	TotalPrice          float64            `json:"total_price"`           // 最终总价(折后价)
	RefundAmount        float64            `json:"refund_amount"`         // 退款金额
	Status              uint               `json:"status"`                // 状态, 0-正常 1-退菜
	Remark              string             `json:"remark"`                // 备注
	IsGift              bool               `json:"is_gift"`               // 是否赠品, false-否 true-是
	IsWrap              bool               `json:"is_wrap"`               // 是否打包, false-否 true-是
	IsBuffet            bool               `json:"is_buffet"`               // 是否自助餐, false-否 true-是
	IsBuffetCustomer    bool               `json:"is_buffet_customer"`     // 是否自助餐顾客, false-否 true-是
	IsDelay             bool               `json:"is_delay"`              // 是否加钟, false-否 true-是
	IsMust              bool               `json:"is_must"`               // 是否必点, false-否 true-是
	GiftReason          string             `json:"gift_reason"`           // 赠品原因
	ImageUrl            string             `json:"image_url"`             // 图片地址
	CancelReason        string             `json:"refund_reason"`         // 退菜原因
}
```

---

## 数据流转图

```
┌─────────────────────────────────────────────────────────────────┐
│                        客户端请求                                │
│  GET /cashier/order/list?page_no=1&page_size=20&status=-1      │
│  GET /cashier/order/info?sale_bill_uuid=123&sale_order_uuid=0  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Handler 层 (cashier_order.go)               │
│  • GetCashierOrderList()                                       │
│  • GetOrderInfo()                                              │
│  • 参数绑定与验证                                               │
│  • 错误处理                                                     │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Service 层 (order_manage.go)                │
│  • GetOrderLists()                                             │
│  • GetOrderInfos()                                             │
│  • 业务逻辑处理                                                 │
│  • 数据组装与转换                                               │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Repository 层 (order.go)                       │
│  • GetCashierOrderListWithPagination()                          │
│  • GetSaleBillInfo()                                            │
│  • 数据库查询                                                   │
│  • 关联数据预加载                                               │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                        数据库 (MySQL)                           │
│  • ttpos_sale_bill                                             │
│  • ttpos_sale_order                                            │
│  • ttpos_sale_order_product                                    │
│  • ttpos_payment_order                                         │
│  • ...                                                          │
└─────────────────────────────────────────────────────────────────┘
```

---

## 关键逻辑说明

### 1. 拆单处理

**场景**: 一个销售账单（`SaleBill`）可能包含多个销售订单（`SaleOrder`），这种情况称为"拆单"。

**判断条件**: `len(bill.SaleOrders) > 1`

**处理逻辑**:

- **订单列表**: 拆单时，主账单的 `SaleOrders` 字段包含所有子订单信息
- **订单详情**: 可以通过 `SaleOrderUuid` 参数查看单个子订单，或不传参数查看所有子订单
- **序列号**: 拆单时，子订单的序列号为 `主序列号-序号`（如：`A01-1`, `A01-2`）
- **操作权限**: 拆单时，某些操作（如打印、发票）只能在子订单级别执行

### 2. 商品类型处理

订单详情中包含三种类型的商品：

1. **自助餐顾客** (`SaleOrderBuffetCustomerTypes`)
   - 记录自助餐套餐的顾客类型和数量
   - 例如：老人2人、儿童1人

2. **加钟商品** (`SaleOrderBuffetDelayProducts`)
   - 自助餐的加钟服务商品
   - 拆单后数量可能不等于桌台人数，但同一加钟商品的总数等于桌台人数

3. **正常商品** (`SaleOrderProducts`)
   - 普通点餐商品
   - 需要过滤：
     - 套餐子商品（`IsPackageSubProduct()`）
     - 未接单商品（`!IsAcceptOrderProduct()`）
     - 取消订单时的未送厨商品（`IsUnCookingProduct()`）

### 3. 套餐商品处理

**特殊处理**: 套餐商品需要将子商品信息组装到属性名称中。

**格式**: `子商品1*数量（规格，属性）；子商品2*数量（规格，属性）；...`

**示例**: `牛排*1（标准，黑椒汁）；可乐*2（大杯，少冰）；沙拉*1（大份，沙拉酱，蜂蜜酱）`

### 4. 操作权限判断

**字段**: `BillListsExtra`

**判断规则**:

| 操作 | 判断条件 |
|------|----------|
| `IsCellRefund` (可退款) | 非免单 && 未全退款 && 已完成状态 |
| `IsCellCancel` (可取消) | 待付款状态 |
| `IsCellReverseSettle` (可反结账) | 已完成 && 当前员工 && 在班次时间内 |
| `IsCellPrint` (可打印) | 非拆单（拆单时在子订单级别打印） |
| `IsCellInvoice` (可打印发票) | 非拆单 && 已完成状态 |
| `IsCellDelete` (可删除) | 已取消状态 |
| `IsCellShow` (可显示) | 非数据管理订单 |

### 5. 时间处理

**时区**: 使用公司设置的时区（`ctx.GetCompanySetting().Timezone`）

**日期类型**: 
- `-1`: 全部
- `0`: 今天
- `1`: 昨天
- `2`: 本周
- `3`: 本月
- `4`: 本年
- `5`: 近7天
- `6`: 上个月

**时间筛选**: 根据 `EnableCreateTime` 和 `EnablePayTime` 决定使用创建时间还是支付时间进行筛选。

### 6. 数据管理订单

**概念**: 数据管理订单是指被标记为"数据管理"的订单，通常用于数据统计或特殊处理。

**过滤逻辑**:

- `IsOnlyDataManage = 1`: 只返回数据管理订单（通过 `SaleBillUuids` 参数指定）
- `IsContainDataManage = 0`: 不包含数据管理订单
- `IsContainDataManage = 1`: 包含数据管理订单

---

## 数据结构说明

### SaleBill (销售账单)

**核心字段**:

- `Uuid`: 账单UUID
- `OrderNo`: 订单编号
- `SerialNo`: 序列号（桌位编号或点餐流水号）
- `BillType`: 账单类型（0=桌台订单，1=点餐订单）
- `DiningMethod`: 用餐方式（0=堂食，1=打包）
- `Status`: 状态（0=待付款，1=已完成，2=已取消）
- `OriginAmount`: 订单总金额
- `SaleOrders`: 销售订单列表

### SaleOrder (销售订单)

**核心字段**:

- `Uuid`: 订单UUID
- `OrderNo`: 订单编号
- `SaleBillUuid`: 所属账单UUID
- `Status`: 状态
- `OriginAmount`: 订单金额
- `PaymentAmount`: 支付金额
- `IsFree`: 是否免单（0=否，1=是）
- `ConsumerUuid`: 会员UUID
- `PaymentOrders`: 支付订单列表
- `SaleOrderProducts`: 订单商品列表
- `SaleOrderBuffetCustomerTypes`: 自助餐顾客列表
- `SaleOrderBuffetDelayProducts`: 加钟商品列表

### SaleOrderProduct (订单商品)

**核心字段**:

- `Uuid`: 商品UUID
- `Num`: 数量
- `SalePrice`: 销售单价
- `Status`: 状态（0=正常，1=退菜）
- `IsGift`: 是否赠品
- `IsWrap`: 是否打包
- `IsBuffet`: 是否自助餐商品
- `Remark`: 备注

---

## 注意事项

1. **性能优化**: 
   - 使用 GORM 的 `Preload` 预加载关联数据，避免 N+1 查询问题
   - 订单列表查询时，只加载必要的关联数据

2. **数据一致性**:
   - 所有金额计算都使用模型方法（如 `GetPaymentAmount()`, `GetTotalRefundAmount()`），确保计算逻辑一致

3. **多语言支持**:
   - 商品名称、属性名称、取消原因等都支持多语言
   - 使用 `LocaleResponse` 结构存储多语言数据

4. **软删除处理**:
   - 所有查询都过滤软删除的数据（`delete_time = 0`）
   - 关联数据也需要过滤软删除

5. **错误处理**:
   - 使用 `errors.WithMessage` 包装错误，保留错误堆栈信息
   - Handler 层统一处理错误并返回标准错误响应

---

## 相关文件

- **Handler**: `main/app/api/v1/cashier/cashier_order.go`
- **Service**: `main/app/service/order_manage.go`
- **Repository**: `main/app/repository/order.go`
- **DTO Request**: `main/app/dto/req/order.go`
- **DTO Response**: `main/app/dto/resp/order.go`
- **Model**: `main/app/model/sale_bill.go`, `main/app/model/sale_order.go`

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

