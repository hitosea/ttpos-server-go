package transfer_order

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"ttpos-bmp/app/ttpos-erp/api/material_transfer"
	"ttpos-server-go/app/constant"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	ttposContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// transferOrderHelper 调拨单辅助方法
type transferOrderHelper struct{}

// GenerateOrderNo 生成调拨单订单编号
// 格式：TR+12位数字（全平台唯一）
// 组成：8位日期(YYYYMMDD) + 4位序号
func (h *transferOrderHelper) GenerateOrderNo(db *gorm.DB) string {
	now := time.Now()
	dateStr := now.Format("20060102") // 8位日期

	// Redis key: transfer_order:seq:YYYYMMDD
	redisKey := fmt.Sprintf("transfer_order:seq:%s", dateStr)
	ctx := context.Background()

	// 获取Redis客户端（兼容集群和单机模式）
	var seq int64
	var err error

	if clusterClient := cache.Global.GetClusterClient(); clusterClient != nil {
		// 集群模式
		seq, err = clusterClient.Incr(ctx, redisKey).Result()
		if err == nil {
			clusterClient.Expire(ctx, redisKey, 3*24*time.Hour)
			if seq > 9999 {
				clusterClient.Set(ctx, redisKey, 1, 3*24*time.Hour)
				seq = 1
			}
		}
	} else if client := cache.Global.GetClient(); client != nil {
		// 单机模式
		seq, err = client.Incr(ctx, redisKey).Result()
		if err == nil {
			client.Expire(ctx, redisKey, 3*24*time.Hour)
			if seq > 9999 {
				client.Set(ctx, redisKey, 1, 3*24*time.Hour)
				seq = 1
			}
		}
	}

	// Redis失败时降级使用时间戳
	if err != nil {
		return fmt.Sprintf("TR%s%04d", dateStr, now.Unix()%10000)
	}

	// 生成12位数字：TR + 8位日期 + 4位序号
	orderNo := fmt.Sprintf("TR%s%04d", dateStr, seq)
	return orderNo
}

// CreateLog 创建操作日志
func (h *transferOrderHelper) CreateLog(
	ctx ttposContext.Context,
	db *gorm.DB,
	transferOrderUuid uint64,
	action,
	actionDesc string,
	oldStatus,
	newStatus int,
) error {
	logRepo := repository.NewTransferOrderLogRepo(db)

	log := &model.TransferOrderLog{
		TransferOrderUuid: transferOrderUuid,
		CompanyUuid:       ctx.GetCompanyUuid(),
		Action:            action,
		ActionDesc:        actionDesc,
		OldStatus:         oldStatus,
		NewStatus:         newStatus,
		OperatorUuid:      ctx.GetStaffUuid(),
		OperatorName: func() string {
			if realName := ctx.GetStaff().RealName; realName != "" {
				return realName
			}
			return ctx.GetStaff().Username
		}(),
	}

	return logRepo.Create(log)
}

// 获取总部下所有公司的设置
func (h *transferOrderHelper) GetCompanySetting(
	ctx ttposContext.Context,
	allCompanySettings []model.CompanySetting,
	companyUuid uint64,
) (model.CompanySetting, error) {
	// 获取公司设置
	companySetting := model.CompanySetting{}
	for _, setting := range allCompanySettings {
		if setting.CompanyUuid == companyUuid {
			companySetting = setting
		}
	}
	if companySetting.Uuid == 0 {
		return model.CompanySetting{}, errors.WithMessage(errors.New("获取公司设置失败"), "公司不存在")
	}
	// 返回公司设置和业务设置
	return companySetting, nil
}

// 获取总部下所有公司的设置
func (h *transferOrderHelper) GetCompanySettings(
	ctx ttposContext.Context,
	dbm *database.DBManager,
	allCompanySettings []model.CompanySetting,
	settingSrv *setting.Srv,
	companyUuid uint64,
) (model.Company, model.CompanySetting, respSetting.Business, error) {
	// 获取公司信息
	db := dbm.GetDB(companyUuid)
	if db == nil {
		return model.Company{}, model.CompanySetting{}, respSetting.Business{}, errors.WithMessage(errors.New("获取公司失败"), "数据库连接失败")
	}
	company, err := repository.NewCompanyRepo(db).GetCompany()
	if err != nil {
		return model.Company{}, model.CompanySetting{}, respSetting.Business{}, errors.WithMessage(errors.New("获取公司失败"), err.Error())
	}
	// 获取公司设置
	companySetting, err := h.GetCompanySetting(ctx, allCompanySettings, companyUuid)
	if err != nil {
		return model.Company{}, model.CompanySetting{}, respSetting.Business{}, errors.WithMessage(errors.New("获取公司设置失败"), err.Error())
	}
	// 获取公司业务设置
	businessSetting, err := func(copyCtx ttposContext.Context) (respSetting.Business, error) {
		copyCtx.SetCompanyUuid(companyUuid)
		businessSetting, err := settingSrv.GetBusinessSetting(copyCtx)
		if err != nil {
			return respSetting.Business{}, err
		}
		return businessSetting, nil
	}(ctx.Copy())
	if err != nil {
		return company, companySetting, respSetting.Business{}, errors.WithMessage(errors.New("获取公司业务设置失败"), err.Error())
	}
	// 返回公司设置和业务设置
	return company, companySetting, businessSetting, nil
}

// CreateApproval 创建审批全体流程
func (h *transferOrderHelper) CreateApproval(
	ctx ttposContext.Context,
	db *gorm.DB,
	dbm *database.DBManager,
	transferOrder *model.TransferOrder,
) error {
	approvalRepo := repository.NewTransferOrderApprovalRepo(db)
	settingSrv := setting.NewSrvImpl(dbm, cache.Global)

	// 当前公司的设置
	companySetting := ctx.GetCompanySetting()

	// 获取当前公司总部下所有公司的设置
	allCompanySettings, err := repository.NewCompanySettingRepo(dbm.GetDB(0)).GetAllByHeadquarterUuid(companySetting.HeadquarterUuid)
	if err != nil {
		return errors.WithMessage(errors.New("获取总部下所有公司的设置失败"), err.Error())
	}

	// 获取总部门店业务设置
	headquarterUuid := companySetting.HeadquarterUuid
	_, _, headquarterBusinessSetting, err := h.GetCompanySettings(ctx, dbm, allCompanySettings, settingSrv, headquarterUuid)
	if err != nil {
		return err
	}

	// 1. 我的公司和设置
	myCompanyUuid := utils.IfUint64(transferOrder.TransferType == 1, transferOrder.ReceiverCompanyUuid, transferOrder.SenderCompanyUuid)
	myCompany, myCompanySetting, _, err := h.GetCompanySettings(ctx, dbm, allCompanySettings, settingSrv, myCompanyUuid)
	if err != nil {
		return err
	}

	// 2. 我的上级公司和设置
	myParentCompanyUuid := utils.IfUint64(myCompanyUuid == headquarterUuid, headquarterUuid, myCompanySetting.GetParentCompanyUuid(1))
	if myParentCompanyUuid == 0 {
		return errors.WithMessage(errors.New("获取我的上级门店失败，请同步层级数据"))
	}
	myParentCompany, myParentCompanySetting, myParentBusinessSetting, err := h.GetCompanySettings(ctx, dbm, allCompanySettings, settingSrv, myParentCompanyUuid)
	if err != nil {
		return err
	}

	// 3. 对方公司和设置
	theOtherCompanyUuid := utils.IfUint64(transferOrder.TransferType == 1, transferOrder.SenderCompanyUuid, transferOrder.ReceiverCompanyUuid)
	theOtherCompany, theOtherCompanySetting, _, err := h.GetCompanySettings(ctx, dbm, allCompanySettings, settingSrv, theOtherCompanyUuid)
	if err != nil {
		return err
	}

	// 4. 获取对方的上级公司和设置
	theOtherParentCompanyUuid := utils.IfUint64(theOtherCompanyUuid == headquarterUuid, headquarterUuid, theOtherCompanySetting.GetParentCompanyUuid(1))
	if theOtherParentCompanyUuid == 0 {
		return errors.WithMessage(errors.New("获取对方的上级门店失败，请同步层级数据"))
	}
	theOtherParentCompany, theOtherParentCompanySetting, theOtherParentBusinessSetting, err := h.GetCompanySettings(ctx, dbm, allCompanySettings, settingSrv, theOtherParentCompanyUuid)
	if err != nil {
		return err
	}

	// 创建审批列表
	var approvalLists map[int]map[string]interface{} = map[int]map[string]interface{}{
		1: {
			"company_uuid":             myCompanyUuid,
			"company":                  myCompany,
			"erpnext_company_abbr":     myCompanySetting.ErpnextCompanyAbbr,
			"is_via_company_warehouse": "1",
			"approval_type": func() string {
				if transferOrder.TransferType == 1 {
					return constant.TransferApprovalTypeReceiver
				}
				return constant.TransferApprovalTypeSender
			}(),
			"status": constant.TransferApprovalPending,
		},
		2: {
			"company_uuid":         myParentCompanyUuid,
			"company":              myParentCompany,
			"erpnext_company_abbr": myParentCompanySetting.ErpnextCompanyAbbr,
			"is_via_company_warehouse": func() string {
				if headquarterBusinessSetting.IsViaParentCompanyWarehouse() || myParentBusinessSetting.IsViaParentCompanyWarehouse() {
					return "1"
				}
				return "0"
			}(),
			"approval_type": func() string {
				if transferOrder.TransferType == 1 {
					return constant.TransferApprovalTypeReceiverParent
				}
				return constant.TransferApprovalTypeSenderParent
			}(),
			"status": func() int {
				if headquarterBusinessSetting.IsRequiredParentCompanyApproval() || myParentBusinessSetting.IsRequiredParentCompanyApproval() {
					return constant.TransferApprovalPending
				}
				return constant.TransferApprovalSkipped
			}(),
		},
		3: {
			"company_uuid":         theOtherParentCompanyUuid,
			"company":              theOtherParentCompany,
			"erpnext_company_abbr": theOtherParentCompanySetting.ErpnextCompanyAbbr,
			"is_via_company_warehouse": func() string {
				if headquarterBusinessSetting.IsViaParentCompanyWarehouse() || theOtherParentBusinessSetting.IsViaParentCompanyWarehouse() {
					return "1"
				}
				return "0"
			}(),
			"approval_type": func() string {
				if transferOrder.TransferType == 1 {
					return constant.TransferApprovalTypeSenderParent
				}
				return constant.TransferApprovalTypeReceiverParent
			}(),
			"status": func() int {
				if headquarterBusinessSetting.IsRequiredParentCompanyApproval() || theOtherParentBusinessSetting.IsRequiredParentCompanyApproval() {
					return constant.TransferApprovalPending
				}
				return constant.TransferApprovalSkipped
			}(),
		},
		4: {
			"company_uuid":             theOtherCompanyUuid,
			"company":                  theOtherCompany,
			"erpnext_company_abbr":     theOtherCompanySetting.ErpnextCompanyAbbr,
			"is_via_company_warehouse": "1",
			"approval_type": func() string {
				if transferOrder.TransferType == 1 {
					return constant.TransferApprovalTypeSender
				}
				return constant.TransferApprovalTypeReceiver
			}(),
			"status": constant.TransferApprovalPending,
		},
	}

	// 创建审批
	var approvals []*model.TransferOrderApproval
	for key, sequence := range approvalLists {
		approvals = append(approvals, &model.TransferOrderApproval{
			CompanyUuid:         myCompanyUuid,
			HeadquarterUuid:     headquarterUuid,
			TransferOrderUuid:   transferOrder.Uuid,
			ApprovalType:        sequence["approval_type"].(string),
			ApprovalCompanyUuid: sequence["company_uuid"].(uint64),
			ApprovalCompanyName: sequence["company"].(model.Company).Name,
			Status:              sequence["status"].(int),
			IsRequired:          1,
			IsViaCompanyWarehouse: func() int {
				if sequence["is_via_company_warehouse"].(string) == "1" {
					return 1
				}
				return 0
			}(),
			ErpnextCompanyAbbr: sequence["erpnext_company_abbr"].(string),
			Sequence:           key,
		})
	}

	return approvalRepo.CreateBatch(approvals)
}

// GetOrderDb 获取调拨单数据库
func (h *transferOrderHelper) GetOrderDb(
	ctx ttposContext.Context,
	dbm *database.DBManager,
	transferOrderUuid uint64,
) (*gorm.DB, error) {
	// 查询调拨单详情
	companyUuid := ctx.GetCompanyUuid()
	headquarterTransferOrder, err := repository.NewTransferOrderRepo(dbm.GetDB(0)).GetByUuid(transferOrderUuid)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Logger.Error("查询调拨单详情失败", zap.Error(err))
		return nil, errors.WithMessage(errors.New("获取调拨单详情失败"), err.Error())
	}
	if headquarterTransferOrder != nil {
		companyUuid = headquarterTransferOrder.CompanyUuid
	}
	db := dbm.GetDB(companyUuid)
	if db == nil {
		return nil, errors.WithMessage(errors.New("获取调拨单数据库失败"), "数据库不存在")
	}
	return db, nil
}

// CopyDataToHeadquarter 复制数据到总部
// 将调拨单及其所有关联数据从门店数据库复制到总部数据库
func (h *transferOrderHelper) CopyDataToHeadquarter(
	ctx ttposContext.Context,
	dbm *database.DBManager,
	db *gorm.DB,
	transferOrderUuid uint64,
) error {
	// 获取总部数据库连接
	headquarterDb := dbm.GetDB(0)
	if headquarterDb == nil {
		return errors.New("获取总部数据库失败")
	}

	// 1. 查询调拨单主表数据（包含所有关联数据）
	transferOrderRepo := repository.NewTransferOrderRepo(db)
	transferOrder, err := transferOrderRepo.GetByUuid(
		transferOrderUuid,
		transferOrderRepo.WithApprovals(),
	)
	if err != nil {
		return errors.WithMessage(errors.New("查询调拨单失败"), err.Error())
	}

	// 3. 在总部数据库中开启事务
	err = headquarterDb.Transaction(func(hqTx *gorm.DB) error {
		// 3.1 检查总部数据库中是否已存在该调拨单
		hqTransferOrderRepo := repository.NewTransferOrderRepo(hqTx)
		existingOrder, err := hqTransferOrderRepo.GetByUuid(transferOrderUuid)
		if err != nil && err != gorm.ErrRecordNotFound {
			return errors.WithMessage(errors.New("查询总部调拨单失败"), err.Error())
		}

		// 3.2 复制或更新调拨单主表
		transferOrderCopy := *transferOrder
		transferOrderCopy.SetNil() // 清空关联数据，避免重复插入
		if existingOrder != nil {
			// 已存在，更新数据（不更新 ID）
			transferOrderCopy.ID = existingOrder.ID
			if err := hqTx.Model(&model.TransferOrder{}).
				Where("uuid = ?", transferOrderUuid).
				Updates(&transferOrderCopy).Error; err != nil {
				return errors.WithMessage(errors.New("更新总部调拨单失败"), err.Error())
			}
		} else {
			// 不存在，插入数据（清空 ID，让数据库自动生成）
			transferOrderCopy.ID = 0
			if err := hqTx.Create(&transferOrderCopy).Error; err != nil {
				return errors.WithMessage(errors.New("创建总部调拨单失败"), err.Error())
			}
		}

		// 3.3 复制审批记录
		if len(transferOrder.Approvals) > 0 {
			hqApprovalRepo := repository.NewTransferOrderApprovalRepo(hqTx)
			for _, approval := range transferOrder.Approvals {
				approvalCopy := *approval

				// 检查审批记录是否已存在
				existingApproval, err := hqApprovalRepo.GetByUuid(approval.Uuid)
				if err != nil && err != gorm.ErrRecordNotFound {
					return errors.WithMessage(errors.New("查询总部调拨单审批记录失败"), err.Error())
				}

				if existingApproval != nil {
					// 已存在，更新（不更新 ID）
					approvalCopy.ID = existingApproval.ID
					if err := hqTx.Model(&model.TransferOrderApproval{}).
						Where("uuid = ?", approval.Uuid).
						Updates(&approvalCopy).Error; err != nil {
						return errors.WithMessage(errors.New("更新总部调拨单审批记录失败"), err.Error())
					}
				} else {
					// 不存在，插入（清空 ID，让数据库自动生成）
					approvalCopy.ID = 0
					if err := hqTx.Create(&approvalCopy).Error; err != nil {
						return errors.WithMessage(errors.New("创建总部调拨单审批记录失败"), err.Error())
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("复制数据到总部失败",
			zap.Uint64("transfer_order_uuid", transferOrderUuid),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// extractName 从错误信息中提取名称
func (h *transferOrderHelper) extractName(name, after, errorMsg string) string {
	// 转义正则表达式中的特殊字符
	escapedName := regexp.QuoteMeta(name)
	escapedAfter := regexp.QuoteMeta(after)

	// 使用正则表达式匹配，支持HTML标签和普通文本
	var re *regexp.Regexp
	if strings.Contains(name, "<") || strings.Contains(after, "<") {
		// HTML标签模式：不要求空格分隔
		re = regexp.MustCompile(escapedName + `(.+?)` + escapedAfter)
	} else {
		// 普通文本模式：要求空格分隔
		re = regexp.MustCompile(escapedName + `\s+(.+?)\s+` + escapedAfter)
	}

	matches := re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		supplierInfo := strings.TrimSpace(matches[1])
		// 如果包含编码#名称格式，提取名称部分
		if strings.Contains(supplierInfo, "#") {
			parts := strings.SplitN(supplierInfo, "#", 2)
			if len(parts) == 2 {
				return parts[1] // 返回供应商名称
			}
		}
		return supplierInfo
	}
	return ""
}

// handleErpError 处理ERP错误
func (h *transferOrderHelper) handleErpError(ctx ttposContext.Context, err error, transferOrder *model.TransferOrder) error {
	// 记录日志，方便排查问题
	if transferOrder != nil {
		logger.Logger.Error("handleErpError", zap.Any("transferOrder-uuid", transferOrder.Uuid), zap.Any("err", err))
	} else {
		logger.Logger.Error("handleErpError", zap.Any("err", err))
	}

	// 检查物品状态
	if strings.Contains(err.Error(), "Item") && strings.Contains(err.Error(), "is disabled") {
		// 提取物品名称
		itemName := h.extractName("Item", "is disabled", err.Error())
		if itemName != "" {
			return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, fmt.Sprintf(
				i18n.Translate(ctx.GetLanguage(), "物品 %s 已禁用，请修改物品状态"), itemName),
			)
		}
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, "有物品已禁用，请修改物品状态")
	}
	// 检查仓库状态
	if strings.Contains(err.Error(), "Warehouse") && strings.Contains(err.Error(), "is disabled") {
		// 提取仓库名称
		warehouseName := h.extractName("Warehouse", "is disabled", err.Error())
		if warehouseName != "" {
			return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, fmt.Sprintf("仓库 %s 已禁用，请修改仓库状态", warehouseName))
		}
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, "仓库已禁用，请修改仓库状态")
	}
	// 检查采购数量
	if strings.Contains(err.Error(), "cannot be less than minimum order qty") {
		itemName := h.extractName("Item", ":Ordered qty", err.Error())
		if itemName != "" {
			num := h.extractName("minimum order qty", "(defined in Item).", err.Error())
			if num != "" {
				return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled,
					itemName+" "+i18n.Translate(ctx.GetLanguage(), "采购数量不能小于ERP中设置的最小采购数量")+" "+num,
				)
			}
			return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled,
				itemName+" "+i18n.Translate(ctx.GetLanguage(), "采购数量不能小于ERP中设置的最小采购数量"),
			)
		}
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, "采购数量不能小于ERP中设置的最小采购数量")
	}
	// before Transaction Date
	if strings.Contains(err.Error(), "before Transaction Date") {
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, i18n.Translate(ctx.GetLanguage(), "期望到货日期不能小于今天"))
	}
	// 检查单位是否为整数
	if strings.Contains(err.Error(), "Must be Whole Number") {
		itemName := h.extractName("UOM <strong>", "</strong>", err.Error())
		if itemName != "" {
			return errors.NewWithCode(
				constant.CodePurchaseOrderSupplierDisabled,
				fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "单位 %s 只能使用整数"), itemName),
			)
		}
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, "单位只能使用整数")
	}
	// 检查发货库存不足
	if strings.Contains(err.Error(), "创建发货单失败") && strings.Contains(err.Error(), "NegativeStockError") {
		itemName := h.extractName("Item ", "</a> needed in", err.Error())
		if itemName != "" {
			return errors.NewWithCode(
				constant.CodePurchaseOrderSupplierDisabled,
				fmt.Sprintf(i18n.Translate(ctx.GetLanguage(), "物品 %s 库存不足，请补充库存"), itemName),
			)
		}
		return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled, "物品库存不足，请补充库存")
	}
	// 未知错误
	return errors.NewWithCode(constant.CodePurchaseOrderSupplierDisabled,
		i18n.Translate(ctx.GetLanguage(), "操作失败")+": "+transferOrder.OrderNo,
	)
}

// ConvertTransferOrderItemsToMaterialTransferItems 转换调拨单物品列表为ERP接口物品列表
func (h *transferOrderHelper) ConvertTransferOrderItemsToMaterialTransferItems(items []*model.TransferOrderItem) []*material_transfer.MaterialTransferItem {
	materialTransferItems := make([]*material_transfer.MaterialTransferItem, 0, len(items))
	for _, item := range items {
		if item.DeleteTime > 0 {
			continue
		}
		for _, unit := range item.Units {
			if unit.DeleteTime > 0 {
				continue
			}
			materialTransferItems = append(materialTransferItems, &material_transfer.MaterialTransferItem{
				ItemCode: item.MaterialCode,
				Qty:      unit.Num,
				Uom:      unit.ErpnextUom,
			})
		}
	}
	return materialTransferItems
}

// 调用erp接口保存调拨单
func (h *transferOrderHelper) SaveMaterialTransfer(ctx ttposContext.Context, dbm *database.DBManager, db *gorm.DB, transferOrder *model.TransferOrder) (*material_transfer.MaterialTransferResp, error) {
	// 获取当前公司总部下所有公司的设置
	allCompanySettings, err := repository.NewCompanySettingRepo(dbm.GetDB(0)).GetAllByHeadquarterUuid(ctx.GetCompanySetting().HeadquarterUuid)
	if err != nil {
		return nil, errors.WithMessage(errors.New("获取总部下所有公司的设置失败"), err.Error())
	}

	// 发货门店
	senderCompanySetting := model.CompanySetting{}
	if transferOrder.SenderCompanyUuid != 0 {
		companySetting, err := h.GetCompanySetting(ctx, allCompanySettings, transferOrder.SenderCompanyUuid)
		if err != nil {
			logger.Logger.Error("获取公司设置失败", zap.Error(err))
			return nil, errors.WithMessage(errors.New("获取公司设置失败"), err.Error())
		}
		senderCompanySetting = companySetting
	}

	// 收货门店
	receiverCompanySetting := model.CompanySetting{}
	if transferOrder.ReceiverCompanyUuid != 0 {
		companySetting, err := h.GetCompanySetting(ctx, allCompanySettings, transferOrder.ReceiverCompanyUuid)
		if err != nil {
			logger.Logger.Error("获取公司设置失败", zap.Error(err))
			return nil, errors.WithMessage(errors.New("获取公司设置失败"), err.Error())
		}
		receiverCompanySetting = companySetting
	}

	senderParentCompanyAbbr := ""
	senderParentBranch := ""
	receiverParentCompanyAbbr := ""
	receiverParentBranch := ""

	// 查询审批流程列表
	approvals, err := repository.NewTransferOrderApprovalRepo(db).GetListByTransferOrderUuid(transferOrder.Uuid)
	if err != nil {
		return nil, errors.WithMessage(errors.New("查询审批流程失败"), err.Error())
	}
	for _, approval := range approvals {
		if approval.ApprovalType == constant.TransferApprovalTypeSenderParent && approval.IsViaCompanyWarehouseBool() {
			companySetting, err := h.GetCompanySetting(ctx, allCompanySettings, approval.ApprovalCompanyUuid)
			if err == nil {
				senderParentCompanyAbbr = companySetting.ErpnextCompanyAbbr
				senderParentBranch = companySetting.ErpnextBranchName
			}
		}
		if approval.ApprovalType == constant.TransferApprovalTypeReceiverParent && approval.IsViaCompanyWarehouseBool() {
			companySetting, err := h.GetCompanySetting(ctx, allCompanySettings, approval.ApprovalCompanyUuid)
			if err != nil {
				receiverParentCompanyAbbr = companySetting.ErpnextCompanyAbbr
				receiverParentBranch = companySetting.ErpnextBranchName
			}
		}
	}

	materialTransferReq := &material_transfer.MaterialTransferReq{
		FromCompanyAbbr:       senderCompanySetting.ErpnextCompanyAbbr,
		FromBranch:            senderCompanySetting.ErpnextBranchName,
		ToCompanyAbbr:         receiverCompanySetting.ErpnextCompanyAbbr,
		ToBranch:              receiverCompanySetting.ErpnextBranchName,
		FromWarehouse:         transferOrder.OutWarehouseErpCode,
		ToWarehouse:           transferOrder.InWarehouseErpCode,
		Items:                 h.ConvertTransferOrderItemsToMaterialTransferItems(transferOrder.Items),
		FromParentCompanyAbbr: senderParentCompanyAbbr,
		FromParentBranch:      senderParentBranch,
		ToParentCompanyAbbr:   receiverParentCompanyAbbr,
		ToParentBranch:        receiverParentBranch,
	}
	erpResp, err := erp.NewIErpSrv(dbm).SaveMaterialTransfer(ctx, senderCompanySetting, materialTransferReq)
	if err != nil {
		logger.Logger.Error("调用erp接口失败 - 审批通过调拨单", zap.Error(err), zap.Any("materialTransferReq", utils.ToJsonString(materialTransferReq)))
		return nil, h.handleErpError(ctx, err, transferOrder)
	}

	return erpResp, nil
}

// 更新在途仓库存
func (h *transferOrderHelper) UpdateStockInTransit(
	ctx ttposContext.Context,
	dbm *database.DBManager,
	db *gorm.DB,
	transferOrder *model.TransferOrder,
) error {
	approvals, err := repository.NewTransferOrderApprovalRepo(db).GetListByTransferOrderUuid(transferOrder.Uuid)
	if err != nil {
		return errors.WithMessage(errors.New("查询审批流程失败"), err.Error())
	}

	// TODO：优化代码，减少重复代码
	// TODO：同一个父亲的处理的逻辑

	// 遍历审批流程
	for _, approval := range approvals {
		if !approval.IsViaCompanyWarehouseBool() {
			continue
		}
		// 获取数据库连接
		targetDb := dbm.GetDB(approval.ApprovalCompanyUuid)
		if targetDb == nil {
			return errors.WithMessage(errors.New("获取数据库连接失败"), "获取数据库连接失败")
		}
		err = targetDb.Transaction(func(tx *gorm.DB) error {

			// 获取Repository
			warehouseRepo := repository.NewWarehouseRepo(tx)
			warehouseItemRepo := repository.NewWarehouseItemRepo(tx)
			materialRepo := repository.NewMaterialRepo(tx)
			warehouseLogRepo := repository.NewWarehouseInOutLogRepo(tx)

			// 发货门店
			if approval.ApprovalType == constant.TransferApprovalTypeSender {

				// 获取目标仓库信息（通过ERP编码查找）
				targetWarehouse, err := warehouseRepo.GetByErpCode(transferOrder.OutWarehouseErpCode)
				if err != nil {
					logger.Logger.Error("recordErpStockInLog-GetByErpCode", zap.Any("targetWarehouseErpCode", transferOrder.OutWarehouseErpCode), zap.Any("err", err))
					return errors.WithMessage(errors.New("获取目标仓库信息失败"), err.Error())
				}

				// 减少库存
				for _, item := range transferOrder.Items {
					actualNum := item.GetUnitsTotalConversionRateNum()

					// 获取物品信息
					material, err := materialRepo.GetMaterialByCode(item.MaterialCode,
						materialRepo.WithUnit(),
						materialRepo.WithRelatedMaterialList(),
					)
					if err != nil {
						logger.Logger.Error("reduceHeadquarterStockAndLog-GetMaterialByCode", zap.Any("materialCode", item.MaterialCode), zap.Any("err", err))
						return errors.WithMessage(errors.New("获取物品失败"), err.Error())
					}

					// 获取仓库物品
					warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(targetWarehouse.Uuid, material.Uuid)
					if err != nil {
						logger.Logger.Error("查询仓库存失败", zap.Error(err), zap.Any("transferOrder", transferOrder), zap.Any("item", item))
						return errors.WithMessage(errors.New("查询仓库存失败"), err.Error())
					}

					// 减少仓库物品库存
					err = warehouseItemRepo.ReduceStock(warehouseItem.Uuid, actualNum)
					if err != nil {
						logger.Logger.Error("reduceHeadquarterStockAndLog-ReduceStock", zap.Any("warehouseItem", warehouseItem), zap.Any("actualNum", actualNum), zap.Any("err", err))
						return errors.WithMessage(errors.New("减少总部库存失败"), err.Error())
					}

					// 更新规格/加料关联材料库存
					relatedMaterialUuids := material.GetRelatedMaterialUuids()
					err = materialRepo.UpdateRelatedMaterialStock(relatedMaterialUuids)
					if err != nil {
						logger.Logger.Error("reduceHeadquarterStockAndLog-updateRelatedMaterialStock", zap.Any("relatedMaterialUuids", relatedMaterialUuids), zap.Any("err", err))
						return errors.WithMessage(errors.New("更新规格/加料关联材料库存失败"), err.Error())
					}

					// 记录出库日志
					warehouseLog := &model.WarehouseInOutLog{
						LogType:              constant.WarehouseInOutLogLogTypeOut,    // 出库
						Scene:                constant.WarehouseInOutLogSceneDelivery, // 发货出库
						WarehouseUuid:        targetWarehouse.Uuid,
						MaterialUuid:         material.Uuid,
						MaterialName:         material.Name,
						MaterialBaseUnitUuid: material.UnitUuid,
						MaterialBaseUnitName: func() string {
							if material.Unit != nil {
								return material.Unit.Name
							}
							return ""
						}(),
						Num:          actualNum,
						Price:        item.Valuation,
						Amount:       decimal.NewFromFloat(item.Valuation).Mul(decimal.NewFromFloat(actualNum)).InexactFloat64(),
						OrderNo:      transferOrder.OrderNo,
						OtherOrgType: 1,
						OtherOrgUuid: approval.ApprovalCompanyUuid,
						OtherOrgName: approval.ApprovalCompanyName,
					}
					err = warehouseLogRepo.Create(warehouseLog)
					if err != nil {
						logger.Logger.Error("reduceHeadquarterStockAndLog-Create", zap.Any("warehouseLog", warehouseLog), zap.Any("err", err))
						return errors.WithMessage(errors.New("记录出库日志失败"), err.Error())
					}
				}

			}
			// 发货门店上级
			if approval.ApprovalType == constant.TransferApprovalTypeSenderParent {
				// // 获取在途仓库
				// transitWarehouse, err := repository.NewWarehouseRepo(db).GetTransitWarehouse()
				// if err != nil {
				// 	logger.Logger.Error("查询在途仓仓库失败", zap.Error(err), zap.Any("transferOrder", transferOrder))
				// 	return errors.WithMessage(errors.New("查询在途仓仓库失败"), err.Error())
				// }
				// warehouse, err := repository.NewWarehouseRepo(db).GetByErpCode(transferOrder.OutParentWarehouseErpCode)
				// if err != nil {
				// 	return errors.WithMessage(errors.New("查询在途仓仓库失败"), err.Error())
				// }
			}
			// 收货门店上级
			if approval.ApprovalType == constant.TransferApprovalTypeReceiverParent {
				// warehouse, err := repository.NewWarehouseRepo(db).GetByErpCode(transferOrder.InParentWarehouseErpCode)
				// if err != nil {
				// 	return errors.WithMessage(errors.New("查询在途仓仓库失败"), err.Error())
				// }
			}
			// 收货门店
			if approval.ApprovalType == constant.TransferApprovalTypeReceiver {
				// warehouse, err := repository.NewWarehouseRepo(db).GetByErpCode(transferOrder.InWarehouseErpCode)
				// if err != nil {
				// 	return errors.WithMessage(errors.New("查询在途仓仓库失败"), err.Error())
				// }
			}
			return nil
		})
		if err != nil {
			logger.Logger.Error("更新在途仓库存失败", zap.Error(err), zap.Any("transferOrder", transferOrder))
			return errors.WithMessage(errors.New("更新在途仓库存失败"), err.Error())
		}
	}
	return nil
}
