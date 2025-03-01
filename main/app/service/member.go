package service

import (
	"errors"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/duke-git/lancet/v2/cryptor"
)

// IMemberSrv 定义会员服务接口
type IMemberSrv interface {
	GetLevels(companyUuid uint64) resp.MemberLevelList                                                             // 获取等级列表
	SearchMember(companyUuid uint64, keyword string) resp.SearchMemberList                                         // 模糊搜索
	AddMember(ctx context.Context, addMemberReq req.AddMemberReq) error                                            // 添加会员
	GetRechargeMember(companyUuid uint64, memberUuid uint64) resp.RechargeMember                                   // 获取充值会员信息
	GetMemberDiscount(ctx context.Context, discountReq req.GetMemberDiscountReq) (*resp.MemberDiscountResp, error) // 获取会员折扣
	CheckMemberPassword(ctx context.Context, discountReq req.CheckMemberPasswordReq) error                         // 使用会员优惠验证密码
	HandleMemberUpgrade(companyUuid uint64, memberUuid uint64)                                                     // 处理会员升级
}

// memberSrv 会员服务结构体
type memberSrv struct {
	dbm *database.DBManager // 数据库管理器
}

// NewMemberSrv 创建新的会员服务
func NewMemberSrv(dbm *database.DBManager) IMemberSrv {
	return NewMemberSrvImpl(dbm)
}

// NewMemberSrvImpl 创建新的会员服务实现
func NewMemberSrvImpl(dbm *database.DBManager) IMemberSrv {
	return &memberSrv{
		dbm: dbm,
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
	// 判断等级是否存在
	if !memberRepo.CheckLevelExists(addMemberReq.LevelUuid) {
		return errors.New("会员等级不存在")
	}
	// 判断是否存在
	if memberRepo.CheckMemberExists(addMemberReq.Phone) {
		return errors.New("会员已存在")
	}
	if addMemberReq.Password != "" {
		addMemberReq.Password = cryptor.Md5String(addMemberReq.Password)
	}
	if err := memberRepo.CreateMember(model.Member{
		MemberNo:        utils.RandomNumber(5), // 5位数字
		Nickname:        addMemberReq.Nickname,
		Phone:           addMemberReq.Phone,
		Password:        addMemberReq.Password,
		MemberLevelUuid: addMemberReq.LevelUuid,
	}); err != nil {
		ctx.Log().Error("添加会员失败", zap.Error(err))
		return errors.New("添加会员失败")
	}
	return nil
}

// GetRechargeMember 获取充值会员信息
func (s *memberSrv) GetRechargeMember(companyUuid uint64, memberUuid uint64) resp.RechargeMember {
	memberRepo := repository.NewMemberRepo(s.dbm.GetDB(companyUuid))
	member := memberRepo.GetMember(memberRepo.WhereUuid(memberUuid),
		memberRepo.WithMemberCard(), memberRepo.WithMemberCardType(), memberRepo.WithMemberLevel())

	var cardName string
	if member.MemberCard != nil && member.MemberCard.MemberCardType != nil {
		cardName = member.MemberCard.MemberCardType.Name
	}
	var level string
	if member.MemberLevel != nil {
		level = member.MemberLevel.Name
	}
	return resp.RechargeMember{
		Uuid:     member.Uuid,
		Nickname: member.Nickname,
		Card:     resp.Card{Name: cardName},
		Level:    resp.Level{Name: level},
		Balance:  member.Balance + member.GiftBalance,
		Points:   member.Point,
		Phone:    member.Phone,
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
	memberLevels := repository.NewMemberRepo(db).GetMemberLevels()
	if len(memberLevels) == 0 {
		return
	}
	var upgradeLevel model.MemberLevel
	for _, level := range memberLevels {
		if level.IsDefault == 1 {
			continue
		}
		if s.checkCanUpgrade(member, level) {
			upgradeLevel = level
			break
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

// checkCanUpgrade 检查会员是否可以升级
func (s *memberSrv) checkCanUpgrade(member model.Member, level model.MemberLevel) bool {
	if (level.OpenMoney == 1 && level.OpenPoint == 1) &&
		(member.AccumulatedConsumptionAmount >= level.UpgradeMoney && member.Point >= level.UpgradePoint) {
		return true
	}
	if level.OpenMoney == 1 && member.AccumulatedConsumptionAmount >= level.UpgradeMoney {
		return true
	}
	if level.OpenPoint == 1 && member.Point >= level.UpgradePoint {
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
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(discountReq.SaleBillUuid)
	if errSaleBill != nil {
		ctx.Log().Error("查询销售账单失败", zap.Error(errSaleBill))
		return nil, errors.New("查询销售账单失败")
	}
	// 获取销售账单信息
	saleOrder := saleBill.GetSaleOrder(discountReq.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 计算销售订单的金额
	memberDiscountFee := s.calcOrderAmount(ctx, member, saleBill, saleOrder)

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
			Uuid:     member.Uuid,
			Nickname: member.Nickname,
			Card:     resp.Card{Name: cardName},
			Level:    resp.Level{Name: levelName},
			Balance:  member.Balance,
			Points:   member.Point,
			Phone:    member.Phone,
		},
		HasPassword:     member.HasPassword(),
		DiscountedPrice: memberDiscountFee,
	}, nil
}

func (s *memberSrv) calcOrderAmount(ctx context.Context, member *model.Member, saleBill *model.SaleBill, saleOrder *model.SaleOrder) float64 {
	var cardDiscountRate float64
	if member.MemberCard != nil {
		cardDiscountRate = member.MemberCard.Discount
	}
	var levelDiscountRate float64
	if member.MemberLevel != nil {
		levelDiscountRate = member.MemberLevel.Discount
	}
	saleOrder.ConsumerUuid = member.Uuid
	saleOrder.MemberCardDiscountRate = cardDiscountRate
	saleOrder.MemberDiscountRate = levelDiscountRate

	taxFeeType := saleBill.SaleBillSetting.GetTaxFeeType()
	serviceFeeType := saleBill.SaleBillSetting.GetServiceFeeType()
	serviceFeeValue := saleBill.SaleBillSetting.ServiceFeeValue
	for i, _ := range saleOrder.SaleOrderProducts {
		saleOrderProduct := saleOrder.SaleOrderProducts[i]
		if saleOrderProduct.IsDelete() || saleOrderProduct.IsCancelProduct() || !saleOrderProduct.IsAcceptOrderBool() {
			continue
		}
		saleOrderProduct.MemberCardDiscountRate = cardDiscountRate
		saleOrderProduct.MemberDiscountRate = levelDiscountRate
		calc := saleOrderProduct.CalcSaleOrderProduct(serviceFeeValue, taxFeeType, serviceFeeType)
		ctx.Log().Debug("商品会员优惠金额", zap.Any("discount", calc.MemberDiscountFee))
	}
	ctx.Log().Debug("获取会员优惠金额", zap.Any("levelDiscountRate", levelDiscountRate), zap.Any("cardDiscountRate", cardDiscountRate))
	calc := saleOrder.CalcSaleOrder(serviceFeeType, serviceFeeValue, taxFeeType)
	memberDiscountFee := calc.MemberDiscountFee
	return memberDiscountFee
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
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(discountReq.SaleBillUuid)
	if errSaleBill != nil {
		ctx.Log().Error("查询销售账单失败", zap.Error(errSaleBill))
		return errors.New("查询销售账单失败")
	}

	// 获取销售账单信息
	saleOrder := saleBill.GetSaleOrder(discountReq.SaleOrderUuid)
	if saleOrder == nil {
		return errors.New("销售订单不存在")
	}

	// 重新计算销售订单的金额
	memberDiscountFee := s.calcOrderAmount(ctx, member, saleBill, saleOrder)
	ctx.Log().Debug("选定会员,优惠金额", zap.Any("memberDiscountFee", memberDiscountFee), zap.Any("saleOrder.MemberDiscountFee", saleOrder.MemberDiscountFee))

	errUpdate := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {

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
	if errUpdate != nil {
		ctx.Log().Error("更新销售订单失败", zap.Error(errUpdate))
		return errors.New("更新销售订单失败")
	}

	return nil
}
