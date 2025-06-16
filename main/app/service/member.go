package service

import (
	"fmt"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/duke-git/lancet/v2/cryptor"
)

// IMemberSrv 定义会员服务接口
type IMemberSrv interface {
	GetLevels(companyUuid uint64) resp.MemberLevelList                                                             // 获取等级列表
	GetCardTypes(companyUuid uint64) resp.MemberCardTypeList                                                       // 获取会员开类型
	SearchMember(companyUuid uint64, keyword string) resp.SearchMemberList                                         // 模糊搜索
	AddMember(ctx context.Context, addMemberReq req.AddMemberReq) error                                            // 添加会员
	GetRechargeMember(companyUuid uint64, memberUuid uint64) resp.RechargeMember                                   // 获取充值会员信息
	GetMemberDiscount(ctx context.Context, discountReq req.GetMemberDiscountReq) (*resp.MemberDiscountResp, error) // 获取会员折扣
	CheckMemberPassword(ctx context.Context, discountReq req.CheckMemberPasswordReq) error                         // 使用会员优惠验证密码
	HandleMemberUpgrade(companyUuid uint64, memberUuid uint64)                                                     // 处理会员升级
	HandleMemberPoints(ctx context.Context, changeReq MemberPointsChangeReq) error                                 // 处理会员积分
	HandleMemberBalance(ctx context.Context, changeReq MemberBalanceChangeReq) error                               // 处理会员余额
}

// memberSrv 会员服务结构体
type memberSrv struct {
	dbm *database.DBManager // 数据库管理器
	bus *event.SystemEventBus
}

// NewMemberSrv 创建新的会员服务
func NewMemberSrv(dbm *database.DBManager) IMemberSrv {
	return NewMemberSrvImpl(dbm)
}

// NewMemberSrvImpl 创建新的会员服务实现
func NewMemberSrvImpl(dbm *database.DBManager) IMemberSrv {
	return &memberSrv{
		dbm: dbm,
		bus: event.NewSystemBus(),
	}
}

// GetLevels 获取等级列表
func (s *memberSrv) GetLevels(companyUuid uint64) resp.MemberLevelList {
	memberLevels := repository.NewMemberRepo(s.dbm.GetDB(companyUuid)).GetMemberLevels()
	respMemberLevels := make([]resp.MemberLevel, 0)
	for _, memberLevel := range memberLevels {
		var respMemberLevel resp.MemberLevel
		copier.Copy(&respMemberLevel, memberLevel)
		respMemberLevels = append(respMemberLevels, respMemberLevel)
	}
	return resp.MemberLevelList{
		List: respMemberLevels,
	}
}

// GetCardTypes 获取会员卡
func (s *memberSrv) GetCardTypes(companyUuid uint64) resp.MemberCardTypeList {
	memberCardTypes := repository.NewMemberRepo(s.dbm.GetDB(companyUuid)).GetCardTypes()
	respMemberCardTypes := make([]resp.MemberCardType, 0)
	for _, memberLevel := range memberCardTypes {
		var respMemberLevel resp.MemberCardType
		copier.Copy(&respMemberLevel, memberLevel)
		respMemberCardTypes = append(respMemberCardTypes, respMemberLevel)
	}
	return resp.MemberCardTypeList{
		List: respMemberCardTypes,
	}
}

// SearchMember 模糊搜索会员
func (s *memberSrv) SearchMember(companyUuid uint64, keyword string) resp.SearchMemberList {
	searchMembers := make([]resp.SearchMember, 0)
	if keyword == "" {
		return resp.SearchMemberList{List: searchMembers}
	}
	members := repository.NewMemberRepo(s.dbm.GetDB(companyUuid)).SearchMember(keyword)
	for _, member := range members {
		var searchMember resp.SearchMember
		copier.Copy(&searchMember, member)
		searchMembers = append(searchMembers, searchMember)
	}
	return resp.SearchMemberList{
		List: searchMembers,
	}
}

// AddMember 添加会员
func (s *memberSrv) AddMember(ctx context.Context, addMemberReq req.AddMemberReq) error {
	companyUuid := ctx.GetCompanyUuid()
	memberRepo := repository.NewMemberRepo(s.dbm.GetDB(companyUuid))

	// 判断手机号是否存在
	if memberRepo.CheckMemberExists(addMemberReq.Phone) {
		return errors.New("会员已存在")
	}

	// 判断等级是否存在
	if !memberRepo.CheckLevelExists(addMemberReq.LevelUuid) {
		return errors.New("会员等级不存在")
	}

	// 判断会员卡
	var cardType model.MemberCardType
	if addMemberReq.CardTypeUuid != 0 {
		cardType = memberRepo.GetCheckCardType(addMemberReq.CardTypeUuid)
		if cardType.ID == 0 {
			return errors.New("会员卡不存在")
		}
	}

	// 判断推荐人，活动
	var referrer model.Member
	//activityUuid := addMemberReq.ActivityUuid
	// 默认为0，不接收前端传递的活动ID
	var activityUuid uint64
	if addMemberReq.ReferrerUuid != 0 {
		referrer = memberRepo.GetMember(memberRepo.WhereUuid(addMemberReq.ReferrerUuid))
		if referrer.ID == 0 {
			return errors.New("推荐人不存在")
		}
		//if addMemberReq.ActivityUuid == 0 {
		activities, _ := repository.NewMarketingActivityRepo(ctx.GetDB()).GetActivityListByNow()
		if len(activities) != 0 {
			activityUuid = activities[0].Uuid
		}
		//}
	}

	// 处理会员密码
	if addMemberReq.Password != "" {
		addMemberReq.Password = cryptor.Md5String(addMemberReq.Password)
	}

	// 卡号不能重复
	if addMemberReq.CardNo != "" {
		sameCardNoMember := memberRepo.GetMember(memberRepo.WhereCardNo(addMemberReq.CardNo))
		if sameCardNoMember.ID != 0 {
			return errors.New("卡号重复")
		}
	}

	var memberPointsChanged, memberBalanceChanged bool
	member := &model.Member{
		MemberNo:        utils.RandomNumber(5), // 5位数字
		Nickname:        addMemberReq.Nickname,
		Phone:           addMemberReq.Phone,
		Password:        addMemberReq.Password,
		MemberLevelUuid: addMemberReq.LevelUuid,
		Gender:          constant.MemberGenderUnknown,
		MemberCardNo:    addMemberReq.CardNo,
		ReferrerUuid:    referrer.Uuid,
		ActivityUuid:    activityUuid,
	}
	err := ctx.GetDB().Transaction(func(tx *gorm.DB) error {
		memberRepo = repository.NewMemberRepo(tx)
		if err := memberRepo.CreateMember(member); err != nil {
			return err
		}
		// 未发送会员卡，则返回
		if cardType.ID == 0 {
			return nil
		}
		var discount float64
		if cardType.Discount > 0 {
			discount = cardType.Discount
		}
		var expire int64
		if cardType.Expire > 0 { // 单位是月
			expire = time.Now().Add(time.Duration(expire*30*24) * time.Hour).Unix()
		}
		// 会员卡记录
		_, err := base.NewMemberCardLogRepo(tx).CreateMemberCardLog(model.MemberCardLog{
			Price:              cardType.Price,
			Discount:           discount,
			Expire:             expire,
			MemberName:         member.Nickname,
			MemberPhone:        member.Phone,
			MemberNo:           member.MemberNo,
			MemberCardTypeName: cardType.Name,
			MemberCardTypeUuid: cardType.Uuid,
			MemberUuid:         member.Uuid,
		})
		if err != nil {
			return err
		}
		// 给会员发卡
		memberCard := &model.MemberCard{
			CardTypeUuid: cardType.Uuid,
			MemberUuid:   member.Uuid,
			ExpireTime:   expire,
			Discount:     discount,
		}
		err = memberRepo.CreateMemberCard(memberCard)
		if err != nil {
			return err
		}
		if err = memberRepo.Update(member.Uuid, map[string]any{"member_card_uuid": memberCard.Uuid}); err != nil {
			return err
		}

		textMap := map[string]string{
			constant.SourceCashier:   "收银机",
			constant.SourceAssistant: "点餐助手",
		}
		staffName := ctx.GetStaff().RealName
		if ctx.GetSource() == constant.SourceAssistant {
			staffRepo := repository.NewStaffRepo(tx)
			staff, _ := staffRepo.GetStaff(staffRepo.WhereUuid(ctx.GetAssistantUuid()))
			staffName = staff.RealName
		}
		ctx.SetDB(tx)
		// 开卡赠送积分
		if cardType.OpenPoint == 1 && cardType.OpenPointNum > 0 {
			memberPointsChanged = true
			if err = s.HandleMemberPoints(ctx, MemberPointsChangeReq{
				Uuid:     member.Uuid,
				Points:   cardType.OpenPointNum,
				Scene:    constant.MemberPointLogSceneCashierOrAssistant,
				Describe: fmt.Sprintf("%s管理员添加会员发卡赠送操作 [%s]", textMap[ctx.GetSource()], staffName),
			}); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 开卡赠送余额
		if cardType.OpenMoney == 1 && cardType.OpenMoneyNum > 0 {
			memberBalanceChanged = true
			if err = s.HandleMemberBalance(ctx, MemberBalanceChangeReq{
				MemberUuid: member.Uuid,
				Money:      0,
				GiftMoney:  cardType.OpenMoneyNum,
				Scene:      constant.MemberBalanceLogCashierOrAssistant,
				Describe:   fmt.Sprintf("%s管理员添加会员发卡赠送操作 [%s]", textMap[ctx.GetSource()], staffName),
			}); err != nil {
				return errors.WithMessage(err)
			}
		}
		return nil
	})

	// 发卡可能会送积分，赠送余额
	if memberPointsChanged || memberBalanceChanged {
		go func() {
			if memberBalanceChanged {
				// 发布“会员余额变动”事件
				s.bus.PublishChangeMemberBalanceEvent(event.ChangeMemberBalancePayload{
					BasePayload: event.BasePayload{ // 会员余额变动
						Ctx:          ctx,
						CompanyUuid:  ctx.GetCompanyUuid(),
						Source:       ctx.GetSource(),
						OperatorUuid: int64(ctx.GetStaffUuid()),
					},
				})
			}
			if memberPointsChanged {
				s.HandleMemberUpgrade(companyUuid, member.Uuid)
				// 发布“会员积分变动”事件
				s.bus.PublishChangeMemberPointsEvent(event.ChangeMemberPointsPayload{
					BasePayload: event.BasePayload{ // 会员积分变动
						Ctx:          ctx,
						CompanyUuid:  ctx.GetCompanyUuid(),
						Source:       ctx.GetSource(),
						OperatorUuid: int64(ctx.GetStaffUuid()),
					},
				})
			}
		}()
	}

	// 存在推荐人
	if referrer.Uuid > 0 {
		// todo 处理后续逻辑
	}

	if err != nil {
		ctx.Log().Error("添加会员失败", zap.Error(err))
		return errors.WithMessage(err, "添加会员失败")
	}
	return nil
}

// GetRechargeMember 获取充值会员信息
func (s *memberSrv) GetRechargeMember(companyUuid uint64, memberUuid uint64) resp.RechargeMember {
	memberRepo := repository.NewMemberRepo(s.dbm.GetDB(companyUuid))
	member := memberRepo.GetMember(memberRepo.WhereUuid(memberUuid),
		memberRepo.WithMemberCard(), memberRepo.WithMemberCardType(), memberRepo.WithMemberLevel(), memberRepo.WithMemberBalanceLog())

	var cardName string
	if member.MemberCard != nil && member.MemberCard.MemberCardType != nil {
		cardName = member.MemberCard.MemberCardType.Name
	}
	var level string
	if member.MemberLevel != nil {
		level = member.MemberLevel.Name
	}

	return resp.RechargeMember{
		Uuid:          member.Uuid,
		Nickname:      member.Nickname,
		Card:          resp.Card{Name: cardName},
		Level:         resp.Level{Name: level},
		Balance:       member.GetBalanceAll(),
		Points:        member.GetPoints(),
		Phone:         member.Phone,
		RechargeMoney: member.GetRechargeMoney(),
	}
}

// HandleMemberUpgrade 处理会员升级
func (s *memberSrv) HandleMemberUpgrade(companyUuid uint64, memberUuid uint64) {
	db := s.dbm.GetDB(companyUuid)
	memberRepo := repository.NewMemberRepo(db)
	member := memberRepo.GetMember(memberRepo.WhereUuid(memberUuid), memberRepo.WithMemberLevel())
	if member.Uuid == 0 || member.MemberLevel == nil {
		return
	}
	memberLevels := repository.NewMemberRepo(db).GetMemberLevelsAllColumns()
	if len(memberLevels) == 0 {
		return
	}

	priority := member.MemberLevel.Priority
	var upgradeLevel model.MemberLevel
	for _, level := range memberLevels {
		if level.IsDefault == 1 {
			continue
		}
		// 如果当前会员等级优先级大于等于升级等级优先级，则跳过. 只找比当前会员等级优先级高的等级
		// 注意：这个判断会导致会员只能升级，不能降级
		if priority >= level.Priority {
			continue
		}
		if s.checkCanUpgrade(member, level) {
			upgradeLevel = level
		}
	}
	if upgradeLevel.Uuid == 0 ||
		member.MemberLevel.Priority > upgradeLevel.Priority ||
		member.MemberLevel.Uuid == upgradeLevel.Uuid {
		return
	}
	// 更新会员等级ID
	if err := memberRepo.Update(member.Uuid, map[string]any{"member_level_uuid": upgradeLevel.Uuid}); err != nil {
		return
	}
	// 添加等级变动日志
	if _, err := repository.NewMemberLevelLogRepo(db).Create(model.MemberLevelLog{
		MemberUuid: memberUuid,
		OldLevelId: member.MemberLevelUuid,
		NewLevelId: upgradeLevel.Uuid,
		ChangeType: constant.MemberLevelLogTypeAutoUpgrade,
	}); err != nil {
		return
	}
}

type MemberPointsChangeReq struct {
	Uuid     uint64  `json:"uuid"`
	Points   float64 `json:"points"`
	Scene    int     `json:"scene"`
	Describe string  `json:"describe"`
}

// HandleMemberPoints 处理会员积分
func (s *memberSrv) HandleMemberPoints(ctx context.Context, changeReq MemberPointsChangeReq) error {
	tx := ctx.GetDB()
	memberRepo := repository.NewMemberRepo(tx)
	member := memberRepo.GetMember(memberRepo.WhereUuid(changeReq.Uuid))
	if member.Uuid == 0 {
		return errors.New("会员不存在")
	}
	// 处理会员积分
	if err := memberRepo.Update(changeReq.Uuid, map[string]any{
		"frozen_point": utils.DecimalAdd(member.FrozenPoint, changeReq.Points),
	}); err != nil {
		return errors.WithMessage(err, "处理会员积分失败")
	}
	// 添加积分日志
	if _, err := repository.NewMemberPointLogRepo(tx).Create(model.MemberPointLog{
		MemberUuid: changeReq.Uuid,
		Scene:      changeReq.Scene,
		Value:      changeReq.Points,
		Describe:   changeReq.Describe,
	}); err != nil {
		return errors.WithMessage(err, "处理会员积分失败")
	}
	return nil
}

type MemberBalanceChangeReq struct {
	MemberUuid  uint64  `json:"uuid"`
	Money       float64 `json:"money"`        // 变动的金额。 正数为增加，负数为减少
	GiftMoney   float64 `json:"gift_money"`   // 变动的赠送金额。 正数为增加，负数为减少
	Scene       int     `json:"scene"`        // 场景
	Describe    string  `json:"describe"`     // 描述
	RelatedUuid uint64  `json:"related_uuid"` // 关联的ID。比如退款的时候，关联的是退款单金额的ID; 用餐订单反结账的时候，关联的是用餐订单的ID
}

// HandleMemberBalance 处理会员余额
func (s *memberSrv) HandleMemberBalance(ctx context.Context, changeReq MemberBalanceChangeReq) error {
	tx := ctx.GetDB()
	memberRepo := repository.NewMemberRepo(tx)
	member := memberRepo.GetMember(memberRepo.WhereUuid(changeReq.MemberUuid))
	if member.Uuid == 0 {
		return errors.New("会员不存在")
	}
	updateData := map[string]any{
		"frozen_balance":      utils.DecimalAdd(member.FrozenBalance, changeReq.Money),
		"frozen_gift_balance": utils.DecimalAdd(member.FrozenGiftBalance, changeReq.GiftMoney),
	}
	if err := memberRepo.Update(changeReq.MemberUuid, updateData); err != nil {
		return errors.WithMessage(err, "更新会员余额失败")
	}

	logRepo := repository.NewMemberBalanceLogRepo(tx)
	// 余额明细
	if _, err := logRepo.Create(model.MemberBalanceLog{
		MemberUuid:  changeReq.MemberUuid,
		Scene:       changeReq.Scene,
		Money:       utils.DecimalAdd(changeReq.Money, changeReq.GiftMoney),
		GiftMoney:   changeReq.GiftMoney,
		Describe:    changeReq.Describe,
		RelatedUuid: changeReq.RelatedUuid,
	}); err != nil {
		return errors.WithMessage(err, "记录明细失败")
	}
	return nil
}

// checkCanUpgrade 检查会员是否可以升级
func (s *memberSrv) checkCanUpgrade(member model.Member, level model.MemberLevel) bool {
	if (level.OpenMoney == 1 && level.OpenPoint == 1) &&
		(member.AccumulatedConsumptionAmount >= level.UpgradeMoney && member.GetPoints() >= level.UpgradePoint) {
		return true
	}
	if level.OpenMoney == 1 && member.AccumulatedConsumptionAmount >= level.UpgradeMoney {
		return true
	}
	if level.OpenPoint == 1 && member.GetPoints() >= level.UpgradePoint {
		return true
	}
	return false
}

func (s *memberSrv) GetMemberDiscount(ctx context.Context, discountReq req.GetMemberDiscountReq) (*resp.MemberDiscountResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取会员信息
	member, errMember := repository.NewMemberRepo(db).GetMemberInfoForSaleOrder(ctx, discountReq.MemberUuid)
	if errMember != nil {
		return nil, errMember
	}
	// 获取账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(discountReq.SaleBillUuid)
	if err != nil {
		ctx.Log().Error("查询销售账单失败", zap.Error(err))
		return nil, errors.WithMessage(err, "查询销售账单失败")
	}
	// 获取销售账单信息
	saleOrder := saleBill.GetSaleOrder(discountReq.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.WithMessage(err, "销售订单不存在")
	}

	// 设置会员折扣并重新计算订单的金额
	saleOrder.SetMemberDiscount(*member)
	saleBill.CalcAll()
	memberDiscountFee := saleOrder.GetAmount()

	var cardName string
	var levelName string
	if member.MemberCard != nil {
		cardName = member.MemberCard.MemberCardType.Name
	}
	if member.MemberLevel != nil {
		levelName = member.MemberLevel.Name
	}

	return &resp.MemberDiscountResp{
		Member: resp.RechargeMember{
			Uuid:          member.Uuid,
			Nickname:      member.Nickname,
			Card:          resp.Card{Name: cardName},
			Level:         resp.Level{Name: levelName},
			Balance:       member.GetBalanceAll(),
			Points:        member.GetPoints(),
			Phone:         member.Phone,
			RechargeMoney: member.GetRechargeMoney(),
		},
		HasPassword:     member.HasPassword(),
		DiscountedPrice: memberDiscountFee,
	}, nil
}

func (s *memberSrv) calcOrderAmount(ctx context.Context, member *model.Member, saleBill *model.SaleBill, saleOrder *model.SaleOrder) float64 {
	saleOrder.SetMemberDiscount(*member)

	for i, _ := range saleOrder.SaleOrderProducts {
		saleOrderProduct := saleOrder.SaleOrderProducts[i]
		if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() || !saleOrderProduct.IsAcceptOrderBool() {
			continue
		}
		calc := saleOrderProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)
		ctx.Log().Debug("商品会员优惠金额", zap.Any("discount", calc.MemberDiscountFee))
	}
	calc := saleOrder.CalcSaleOrder(*saleBill.SaleBillSetting)
	return decimal.NewFromFloat(calc.MemberDiscountFee).Round(2).InexactFloat64()
}

func (s *memberSrv) CheckMemberPassword(ctx context.Context, discountReq req.CheckMemberPasswordReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取会员信息
	member, errMember := repository.NewMemberRepo(db).GetMemberInfoForSaleOrder(ctx, discountReq.MemberUuid)
	if errMember != nil {
		return errMember
	}

	// 如果会员有密码的话，验证会员密码
	if member.HasPassword() {
		md5Password := cryptor.Md5String(discountReq.Password)
		ctx.Log().Debug("验证密码", zap.Any("md5Password", md5Password), zap.Any("member.Password", member.Password))
		if member.Password != md5Password {
			ctx.Log().Debug("验证密码", zap.Any("md5Password", md5Password), zap.Any("member.Password", member.Password))
			return errors.New("密码错误")
		}
	}

	// 获取账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(discountReq.SaleBillUuid)
	if err != nil {
		ctx.Log().Error("查询销售账单失败", zap.Error(err))
		return errors.WithMessage(err, "查询销售账单失败")
	}

	// 获取销售账单信息
	saleOrder := saleBill.GetSaleOrder(discountReq.SaleOrderUuid)
	if saleOrder == nil {
		return errors.New("销售订单不存在")
	}

	// 重新计算销售订单的金额
	memberDiscountFee := s.calcOrderAmount(ctx, member, saleBill, saleOrder)
	ctx.Log().Debug("选定会员,优惠金额", zap.Any("memberDiscountFee", memberDiscountFee), zap.Any("saleOrder.MemberDiscountFee", saleOrder.MemberDiscountFee))

	err = repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {

		for index, _ := range saleOrder.SaleOrderProducts {
			saleOrderProduct := saleOrder.SaleOrderProducts[index]
			if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() || !saleOrderProduct.IsAcceptOrderBool() {
				continue
			}
			if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProduct(saleOrderProduct); errUpdate != nil {
				return errUpdate
			}
			ctx.Log().Debug("更新销售订单商品成功")
		}

		// 更新完整个销售订单
		if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrder(saleOrder); errUpdate != nil {
			return errUpdate
		}
		return nil
	})
	if err != nil {
		ctx.Log().Error("更新销售订单失败", zap.Error(err))
		return errors.WithMessage(err, "更新销售订单失败")
	}

	return nil
}
