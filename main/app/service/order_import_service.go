package service

import (
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// IOrderImportSrv 订单导入服务接口
type IOrderImportSrv interface {
	// Import 导入订单数据
	// file: Excel 文件
	// 返回: 成功数量、失败数量、失败列表
	Import(ctx context.Context, file *multipart.FileHeader) (successCount int, failCount int, failList []OrderImportFailItem, err error)
}

// OrderImportFailItem 导入失败项
type OrderImportFailItem struct {
	Row     int    // Excel 行号
	OrderNo string // 订单号
	Reason  string // 失败原因
}

// orderImportSrv 订单导入服务实现
type orderImportSrv struct {
	dbm        *database.DBManager
	orderSrv   IOrderSrv
	productSrv IProductSrv
	memberSrv  IMemberSrv
}

// NewOrderImportSrv 创建订单导入服务
func NewOrderImportSrv(dbm *database.DBManager, orderSrv IOrderSrv, productSrv IProductSrv, memberSrv IMemberSrv) IOrderImportSrv {
	return &orderImportSrv{
		dbm:        dbm,
		orderSrv:   orderSrv,
		productSrv: productSrv,
		memberSrv:  memberSrv,
	}
}

// Import 导入订单数据
func (s *orderImportSrv) Import(ctx context.Context, file *multipart.FileHeader) (successCount int, failCount int, failList []OrderImportFailItem, err error) {
	// 1. 打开 Excel 文件
	f, err := file.Open()
	if err != nil {
		return 0, 0, nil, errors.WithMessage(err, "打开文件失败")
	}
	defer f.Close()

	excelFile, err := excelize.OpenReader(f)
	if err != nil {
		return 0, 0, nil, errors.WithMessage(err, "解析 Excel 文件失败")
	}
	defer excelFile.Close()

	// 2. 解析订单基本信息（Sheet1）
	orderBasicList, err := s.parseOrderBasicSheet(excelFile, "订单基本信息")
	if err != nil {
		return 0, 0, nil, errors.WithMessage(err, "解析订单基本信息失败")
	}

	// 3. 解析订单明细（Sheet2）
	orderDetailList, err := s.parseOrderDetailSheet(excelFile, "订单明细")
	if err != nil {
		return 0, 0, nil, errors.WithMessage(err, "解析订单明细失败")
	}

	// 4. 限制数量
	if len(orderBasicList) > 5000 {
		return 0, 0, nil, errors.New("单次导入不能超过 5000 条，请分批导入")
	}

	// 5. 开始事务
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 6. 批量处理
	batchSize := 500
	for i := 0; i < len(orderBasicList); i += batchSize {
		end := i + batchSize
		if end > len(orderBasicList) {
			end = len(orderBasicList)
		}
		batch := orderBasicList[i:end]

		for _, orderBasic := range batch {
			// 获取该订单的明细
			details := s.getOrderDetailsByOrderNo(orderDetailList, orderBasic.OrderNo)

			// 校验数据
			if err := s.validateOrderData(ctx, tx, &orderBasic, details); err != nil {
				failCount++
				failList = append(failList, OrderImportFailItem{
					Row:     orderBasic.Row,
					OrderNo: orderBasic.OrderNo,
					Reason:  err.Error(),
				})
				continue
			}

			// 检查订单号是否已存在
			orderRepo := repository.NewOrderRepo(tx)
			notExists, err := orderRepo.IsOrderNoExists(orderBasic.OrderNo)
			if err != nil {
				failCount++
				failList = append(failList, OrderImportFailItem{
					Row:     orderBasic.Row,
					OrderNo: orderBasic.OrderNo,
					Reason:  "检查订单号失败: " + err.Error(),
				})
				continue
			}
			if !notExists {
				// notExists 为 false 表示订单号已存在
				failCount++
				failList = append(failList, OrderImportFailItem{
					Row:     orderBasic.Row,
					OrderNo: orderBasic.OrderNo,
					Reason:  "订单号已存在，跳过",
				})
				continue
			}

			// 创建订单
			if err := s.createOrder(ctx, tx, &orderBasic, details); err != nil {
				failCount++
				failList = append(failList, OrderImportFailItem{
					Row:     orderBasic.Row,
					OrderNo: orderBasic.OrderNo,
					Reason:  err.Error(),
				})
				continue
			}

			successCount++
		}
	}

	// 7. 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return 0, 0, nil, errors.WithMessage(err, "提交事务失败")
	}

	// 8. 限制失败列表数量（最多返回 10 条）
	if len(failList) > 10 {
		failList = failList[:10]
	}

	return successCount, failCount, failList, nil
}

// OrderBasicData 订单基本信息数据
type OrderBasicData struct {
	Row           int     // Excel 行号
	OrderNo       string  // 订单号
	CreateTime    int64   // 下单时间（时间戳）
	Status        uint    // 订单状态
	BillType      uint    // 订单类型
	DiningMethod  uint    // 用餐方式
	Amount        float64 // 订单金额（折后价）
	OriginAmount  float64 // 订单原价（折前价）
	ShopName      string  // 门店名称
	DeskName      string  // 桌台名称
	MemberNo      string  // 会员编号
	CashierName   string  // 收银员姓名
	MealNum       uint    // 就餐人数
	DutyNo        string  // 当班编号
	SerialNo      string  // 桌位编号
	ProductAmount float64 // 商品金额
	ServiceFee    float64 // 服务费
	TaxFee        float64 // 税费
	Remark        string  // 备注
}

// OrderDetailData 订单明细数据
type OrderDetailData struct {
	Row         int     // Excel 行号
	OrderNo     string  // 订单号
	ProductName string  // 商品名称
	Num         uint    // 数量
	Price       float64 // 单价
	Amount      float64 // 小计
	Remark      string  // 备注
}

// parseOrderBasicSheet 解析订单基本信息工作表
func (s *orderImportSrv) parseOrderBasicSheet(excelFile *excelize.File, sheetName string) ([]OrderBasicData, error) {
	rows, err := excelFile.GetRows(sheetName)
	if err != nil {
		return nil, errors.WithMessage(err, "读取工作表失败")
	}

	if len(rows) < 2 {
		return nil, errors.New("工作表数据为空")
	}

	var orderList []OrderBasicData
	for i := 1; i < len(rows); i++ { // 跳过表头
		row := rows[i]
		if len(row) < 18 {
			continue // 跳过空行或数据不完整的行
		}

		// 解析下单时间
		createTimeStr := strings.TrimSpace(row[1])
		createTime, err := s.parseDateTime(createTimeStr)
		if err != nil {
			continue // 跳过时间格式错误的行
		}

		// 解析数字字段
		status, _ := strconv.ParseUint(safeGetString(row, 2), 10, 32)
		billType, _ := strconv.ParseUint(safeGetString(row, 3), 10, 32)
		diningMethod, _ := strconv.ParseUint(safeGetString(row, 4), 10, 32)
		amount, _ := strconv.ParseFloat(safeGetString(row, 5), 64)
		originAmount, _ := strconv.ParseFloat(safeGetString(row, 6), 64)
		mealNum, _ := strconv.ParseUint(safeGetString(row, 11), 10, 32)
		productAmount, _ := strconv.ParseFloat(safeGetString(row, 14), 64)
		serviceFee, _ := strconv.ParseFloat(safeGetString(row, 15), 64)
		taxFee, _ := strconv.ParseFloat(safeGetString(row, 16), 64)

		order := OrderBasicData{
			Row:           i + 1,
			OrderNo:       strings.TrimSpace(row[0]),
			CreateTime:    createTime,
			Status:        uint(status),
			BillType:      uint(billType),
			DiningMethod:  uint(diningMethod),
			Amount:        amount,
			OriginAmount:  originAmount,
			ShopName:      strings.TrimSpace(row[7]),
			DeskName:      strings.TrimSpace(row[8]),
			MemberNo:      strings.TrimSpace(row[9]),
			CashierName:   strings.TrimSpace(row[10]),
			MealNum:       uint(mealNum),
			DutyNo:        strings.TrimSpace(row[12]),
			SerialNo:      strings.TrimSpace(row[13]),
			ProductAmount: productAmount,
			ServiceFee:    serviceFee,
			TaxFee:        taxFee,
			Remark:        strings.TrimSpace(row[17]),
		}

		// 跳过订单号为空的记录
		if order.OrderNo == "" {
			continue
		}

		orderList = append(orderList, order)
	}

	return orderList, nil
}

// parseOrderDetailSheet 解析订单明细工作表
func (s *orderImportSrv) parseOrderDetailSheet(excelFile *excelize.File, sheetName string) ([]OrderDetailData, error) {
	rows, err := excelFile.GetRows(sheetName)
	if err != nil {
		return nil, errors.WithMessage(err, "读取工作表失败")
	}

	if len(rows) < 2 {
		return nil, errors.New("工作表数据为空")
	}

	var detailList []OrderDetailData
	for i := 1; i < len(rows); i++ { // 跳过表头
		row := rows[i]
		if len(row) < 6 {
			continue // 跳过空行或数据不完整的行
		}

		// 解析数字字段
		num, _ := strconv.ParseUint(safeGetString(row, 2), 10, 32)
		price, _ := strconv.ParseFloat(safeGetString(row, 3), 64)
		amount, _ := strconv.ParseFloat(safeGetString(row, 4), 64)

		detail := OrderDetailData{
			Row:         i + 1,
			OrderNo:     strings.TrimSpace(row[0]),
			ProductName: strings.TrimSpace(row[1]),
			Num:         uint(num),
			Price:       price,
			Amount:      amount,
			Remark:      strings.TrimSpace(row[5]),
		}

		// 跳过订单号或商品名称为空的记录
		if detail.OrderNo == "" || detail.ProductName == "" {
			continue
		}

		detailList = append(detailList, detail)
	}

	return detailList, nil
}

// parseDateTime 解析日期时间字符串为时间戳
func (s *orderImportSrv) parseDateTime(dateTimeStr string) (int64, error) {
	if dateTimeStr == "" {
		return 0, errors.New("时间不能为空")
	}

	// 尝试解析格式：YYYY-MM-DD HH:mm:ss
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, dateTimeStr)
	if err != nil {
		return 0, errors.WithMessage(err, "时间格式错误，应为 YYYY-MM-DD HH:mm:ss")
	}

	return t.Unix(), nil
}

// safeGetString 安全获取字符串，避免索引越界
func safeGetString(row []string, index int) string {
	if index < len(row) {
		return row[index]
	}
	return ""
}

// getOrderDetailsByOrderNo 根据订单号获取订单明细
func (s *orderImportSrv) getOrderDetailsByOrderNo(detailList []OrderDetailData, orderNo string) []OrderDetailData {
	var details []OrderDetailData
	for _, detail := range detailList {
		if detail.OrderNo == orderNo {
			details = append(details, detail)
		}
	}
	return details
}

// validateOrderData 校验订单数据
func (s *orderImportSrv) validateOrderData(_ context.Context, db *gorm.DB, orderBasic *OrderBasicData, details []OrderDetailData) error {
	// 1. 校验必填字段
	if orderBasic.OrderNo == "" {
		return errors.New("订单号不能为空")
	}
	if orderBasic.CreateTime == 0 {
		return errors.New("下单时间不能为空")
	}
	if orderBasic.ShopName == "" {
		return errors.New("门店名称不能为空")
	}

	// 2. 校验订单状态
	if orderBasic.Status > 2 {
		return errors.New("订单状态无效，应为 0-待付款、1-已完成、2-已取消")
	}

	// 3. 校验订单类型
	if orderBasic.BillType > 2 {
		return errors.New("订单类型无效，应为 0-桌台订单、1-点餐订单、2-会员端订单")
	}

	// 4. 校验用餐方式
	if orderBasic.DiningMethod > 1 {
		return errors.New("用餐方式无效，应为 0-堂食、1-打包")
	}

	// 5. 校验金额
	if orderBasic.Amount < 0 {
		return errors.New("订单金额不能为负数")
	}
	if orderBasic.OriginAmount < orderBasic.Amount {
		return errors.New("订单原价必须大于等于订单金额")
	}

	// 6. 校验门店是否存在
	// companyRepo := repository.NewCompanyRepo(db)
	// _, err := companyRepo.GetCompany(
	// 	repository.CommonRepo.WhereBySoftDelete(),
	// 	companyRepo.WhereName(orderBasic.ShopName),
	// )
	// if err != nil {
	// 	if strings.Contains(err.Error(), "record not found") {
	// 		return errors.New(fmt.Sprintf("门店 %s 不存在", orderBasic.ShopName))
	// 	}
	// 	return errors.WithMessage(err, "查询门店失败")
	// }

	// 7. 校验桌台（如果提供）
	if orderBasic.DeskName != "" && orderBasic.DiningMethod == 0 {
		deskRepo := repository.NewDeskRepo(db)
		_, err := deskRepo.GetDesk(
			repository.CommonRepo.WhereBySoftDelete(),
			func(db *gorm.DB) *gorm.DB {
				return db.Where("desk_no = ?", orderBasic.DeskName)
			},
		)
		if err != nil && strings.Contains(err.Error(), "record not found") {
			return errors.New(fmt.Sprintf("桌台 %s 不存在", orderBasic.DeskName))
		}
	}

	// 8. 校验会员（如果提供）
	if orderBasic.MemberNo != "" {
		memberRepo := repository.NewMemberRepo(db)
		member := memberRepo.GetMember(
			repository.CommonRepo.WhereBySoftDelete(),
			func(db *gorm.DB) *gorm.DB {
				return db.Where("member_no = ?", orderBasic.MemberNo)
			},
		)
		if member.Uuid == 0 {
			return errors.New(fmt.Sprintf("会员 %s 不存在", orderBasic.MemberNo))
		}
	}

	// 9. 校验订单明细
	if len(details) == 0 {
		return errors.New("订单明细不能为空")
	}

	// 10. 校验订单明细基本字段
	for _, detail := range details {
		if detail.Num == 0 {
			return errors.New(fmt.Sprintf("商品 %s 数量不能为0", detail.ProductName))
		}
		if detail.Price < 0 {
			return errors.New(fmt.Sprintf("商品 %s 单价不能为负数", detail.ProductName))
		}
		if detail.Amount < 0 {
			return errors.New(fmt.Sprintf("商品 %s 小计不能为负数", detail.ProductName))
		}
		// 商品存在性校验在 createOrder 中进行，避免重复查询
	}

	return nil
}

// createOrder 创建订单
func (s *orderImportSrv) createOrder(_ context.Context, db *gorm.DB, orderBasic *OrderBasicData, details []OrderDetailData) error {
	// 1. 查找门店 UUID（已在 validateOrderData 中校验，这里再次确认）
	// companyRepo := repository.NewCompanyRepo(db)
	// _, err := companyRepo.GetCompany(
	// 	repository.CommonRepo.WhereBySoftDelete(),
	// 	companyRepo.WhereName(orderBasic.ShopName),
	// )
	// if err != nil {
	// 	return errors.WithMessage(err, "查找门店失败")
	// }

	// 2. 查找桌台 UUID（如果提供）
	var deskUuid uint64
	if orderBasic.DeskName != "" {
		deskRepo := repository.NewDeskRepo(db)
		desk, err := deskRepo.GetDesk(
			repository.CommonRepo.WhereBySoftDelete(),
			func(db *gorm.DB) *gorm.DB {
				return db.Where("desk_no = ?", orderBasic.DeskName)
			},
		)
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("查找桌台 %s 失败", orderBasic.DeskName))
		}
		deskUuid = desk.Uuid
	}

	// 3. 查找会员 UUID（如果提供）
	var consumerUuid uint64
	if orderBasic.MemberNo != "" {
		memberRepo := repository.NewMemberRepo(db)
		member := memberRepo.GetMember(
			repository.CommonRepo.WhereBySoftDelete(),
			func(db *gorm.DB) *gorm.DB {
				return db.Where("member_no = ?", orderBasic.MemberNo)
			},
		)
		if member.Uuid == 0 {
			return errors.New(fmt.Sprintf("会员 %s 不存在", orderBasic.MemberNo))
		}
		consumerUuid = member.Uuid
	}

	// 4. 生成 UUID
	saleBillUuid, _ := utils.GetID()

	// 5. 创建 SaleBill
	saleBill := model.SaleBill{
		BaseModel: model.BaseModel{
			Uuid:       saleBillUuid,
			CreateTime: orderBasic.CreateTime,
			UpdateTime: orderBasic.CreateTime,
		},
		OrderNo:       orderBasic.OrderNo,
		DutyNo:        orderBasic.DutyNo,
		SerialNo:      orderBasic.SerialNo,
		Status:        orderBasic.Status,
		BillType:      orderBasic.BillType,
		DiningMethod:  orderBasic.DiningMethod,
		Amount:        orderBasic.Amount,
		OriginAmount:  orderBasic.OriginAmount,
		ProductAmount: orderBasic.ProductAmount,
		ServiceFee:    orderBasic.ServiceFee,
		TaxFee:        orderBasic.TaxFee,
		CashierName:   orderBasic.CashierName,
		MealNum:       orderBasic.MealNum,
		Remark:        orderBasic.Remark,
		ConsumerUuid:  consumerUuid,
		DeskUuid:      deskUuid,
		Source:        constant.SaleBillSourceDefault, // 导入订单使用默认值 0
		// ShopUuid 字段不存在，通过 DeskUuid 关联到门店
	}

	orderRepo := repository.NewOrderRepo(db)
	_, err := orderRepo.CreateSaleBill(saleBill)
	if err != nil {
		return errors.WithMessage(err, "创建销售账单失败")
	}

	// 6. 创建 SaleOrder
	saleOrderUuid, _ := utils.GetID()
	saleOrder := model.SaleOrder{
		BaseModel: model.BaseModel{
			Uuid:       saleOrderUuid,
			CreateTime: orderBasic.CreateTime,
			UpdateTime: orderBasic.CreateTime,
		},
		OrderNo:       orderBasic.OrderNo,
		Status:        constant.SaleOrderStatusFinish, // 已结账
		SaleBillUuid:  saleBillUuid,
		ConsumerUuid:  consumerUuid,
		Amount:        orderBasic.Amount,
		OriginAmount:  orderBasic.OriginAmount,
		ProductAmount: orderBasic.ProductAmount,
		ServiceFee:    orderBasic.ServiceFee,
		TaxFee:        orderBasic.TaxFee,
		CashierName:   orderBasic.CashierName,
	}

	_, err = orderRepo.CreateSaleOrder(saleOrder)
	if err != nil {
		return errors.WithMessage(err, "创建销售订单失败")
	}

	// 7. 创建 SaleOrderProduct
	saleOrderProductRepo := repository.NewSaleOrderProductRepo(db)
	for _, detail := range details {
		// 通过商品名称查找商品 UUID（通过多语言名称的各个字段匹配）
		productRepo := repository.NewProductRepo(db)
		product, err := productRepo.GetProduct(
			repository.CommonRepo.WhereBySoftDelete(),
			func(db *gorm.DB) *gorm.DB {
				return db.Joins("JOIN ttpos_multi_language_name ON ttpos_product_package.multi_language_name_uuid = ttpos_multi_language_name.uuid").
					Where("ttpos_multi_language_name.zh_name = ? OR ttpos_multi_language_name.en_name = ? OR ttpos_multi_language_name.zh_tw_name = ?",
						detail.ProductName, detail.ProductName, detail.ProductName)
			},
		)
		if err != nil {
			if strings.Contains(err.Error(), "record not found") {
				return errors.New(fmt.Sprintf("商品 %s 不存在", detail.ProductName))
			}
			return errors.WithMessage(err, fmt.Sprintf("查找商品 %s 失败", detail.ProductName))
		}

		productUuid, _ := utils.GetID()
		saleOrderProduct := model.SaleOrderProduct{
			BaseModel: model.BaseModel{
				Uuid:       productUuid,
				CreateTime: orderBasic.CreateTime,
				UpdateTime: orderBasic.CreateTime,
			},
			SaleOrderUuid:      saleOrderUuid,
			ProductPackageUuid: product.Uuid,
			Num:                float64(detail.Num),
			Price:              detail.Price,
			TotalPrice:         detail.Amount,
			OriginTotalPrice:   detail.Amount,
			SalePrice:          detail.Price,
			ProductPrice:       detail.Price,
			Remark:             detail.Remark,
		}

		_, err = saleOrderProductRepo.CreateSaleOrderProduct(&saleOrderProduct)
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("创建订单商品失败: %s", detail.ProductName))
		}
	}

	return nil
}
