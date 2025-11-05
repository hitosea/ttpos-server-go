package transfer_order

import (
	"context"
	"fmt"
	"time"
	"ttpos-server-go/app/constant"
	respSetting "ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	ttposContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

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
	companySetting := model.CompanySetting{}
	for _, setting := range allCompanySettings {
		if setting.CompanyUuid == companyUuid {
			companySetting = setting
		}
	}
	if companySetting.Uuid == 0 {
		return company, companySetting, respSetting.Business{}, errors.WithMessage(errors.New("获取公司设置失败"), "公司不存在")
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
	_, _, headquarterBusinessSetting, err := h.GetCompanySetting(ctx, dbm, allCompanySettings, settingSrv, headquarterUuid)
	if err != nil {
		return err
	}

	// 1. 我的公司和设置
	myCompanyUuid := utils.IfUint64(transferOrder.TransferType == 1, transferOrder.ReceiverCompanyUuid, transferOrder.SenderCompanyUuid)
	myCompany, myCompanySetting, _, err := h.GetCompanySetting(ctx, dbm, allCompanySettings, settingSrv, myCompanyUuid)
	if err != nil {
		return err
	}

	// 2. 我的上级公司和设置
	myParentCompanyUuid := utils.IfUint64(myCompanyUuid == headquarterUuid, headquarterUuid, myCompanySetting.GetParentCompanyUuid(1))
	if myParentCompanyUuid == 0 {
		return errors.WithMessage(errors.New("获取我的上级门店失败，请同步层级数据"))
	}
	myParentCompany, myParentCompanySetting, myParentBusinessSetting, err := h.GetCompanySetting(ctx, dbm, allCompanySettings, settingSrv, myParentCompanyUuid)
	if err != nil {
		return err
	}

	// 3. 对方公司和设置
	theOtherCompanyUuid := utils.IfUint64(transferOrder.TransferType == 1, transferOrder.SenderCompanyUuid, transferOrder.ReceiverCompanyUuid)
	theOtherCompany, theOtherCompanySetting, _, err := h.GetCompanySetting(ctx, dbm, allCompanySettings, settingSrv, theOtherCompanyUuid)
	if err != nil {
		return err
	}

	// 4. 获取对方的上级公司和设置
	theOtherParentCompanyUuid := utils.IfUint64(theOtherCompanyUuid == headquarterUuid, headquarterUuid, theOtherCompanySetting.GetParentCompanyUuid(1))
	if theOtherParentCompanyUuid == 0 {
		return errors.WithMessage(errors.New("获取对方的上级门店失败，请同步层级数据"))
	}
	theOtherParentCompany, theOtherParentCompanySetting, theOtherParentBusinessSetting, err := h.GetCompanySetting(ctx, dbm, allCompanySettings, settingSrv, theOtherParentCompanyUuid)
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
			"company_uuid":             myParentCompanyUuid,
			"company":                  myParentCompany,
			"erpnext_company_abbr":     myParentCompanySetting.ErpnextCompanyAbbr,
			"is_via_company_warehouse": myParentBusinessSetting.ViaParentCompanyWarehouse,
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
			"company_uuid":             theOtherParentCompanyUuid,
			"company":                  theOtherParentCompany,
			"erpnext_company_abbr":     theOtherParentCompanySetting.ErpnextCompanyAbbr,
			"is_via_company_warehouse": theOtherParentBusinessSetting.ViaParentCompanyWarehouse,
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
