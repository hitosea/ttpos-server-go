package service

import (
	"sort"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// IFullReductionActivitySrv 满减活动服务接口
type IFullReductionActivitySrv interface {
	Create(ctx context.Context, req *req.FullReductionActivityCreateReq) error
	Update(ctx context.Context, req *req.FullReductionActivityUpdateReq) error
	GetByUuid(ctx context.Context, uuid uint64) (*resp.FullReductionActivityResp, error)
	GetList(ctx context.Context, req *req.FullReductionActivityListReq) (*resp.FullReductionActivityListResp, error)
	Delete(ctx context.Context, uuid uint64) error
	Disable(ctx context.Context, uuid uint64) error
}

// fullReductionActivitySrv 满减活动服务实现
type fullReductionActivitySrv struct {
	dbm *database.DBManager
}

// NewFullReductionActivitySrv 创建满减活动服务
func NewFullReductionActivitySrv(dbm *database.DBManager) IFullReductionActivitySrv {
	return &fullReductionActivitySrv{
		dbm: dbm,
	}
}

// Create 创建满减活动
func (s *fullReductionActivitySrv) Create(ctx context.Context, req *req.FullReductionActivityCreateReq) error {
	// 获取商家时区
	companySetting := ctx.GetCompanySetting()
	timezone := (&companySetting).GetTimezone()
	// 验证请求并初始化日期字段
	if err := req.Validate(timezone); err != nil {
		return err
	}

	db := ctx.GetDB()
	currentTime := time.Now().Unix()

	// 生成活动UUID
	activityUuid, err := utils.GetID()
	if err != nil {
		return errors.WithMessage(err, "生成UUID失败")
	}

	// 生成多语言名称UUID
	multiLangUuid, err := utils.GetID()
	if err != nil {
		return errors.WithMessage(err, "生成UUID失败")
	}

	// 创建多语言名称
	multiLanguageName := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid:       multiLangUuid,
			CreateTime: currentTime,
			UpdateTime: currentTime,
		},
	}
	multiLanguageName.InitByLocaleResponse(req.LocaleName)

	// 如果是阶梯满减，需要排序
	rules := req.Rules
	if req.ReductionType == constant.FullReductionTypeStep {
		sort.Slice(rules, func(i, j int) bool {
			return rules[i].Threshold < rules[j].Threshold
		})
	}

	// 多个数据库操作，必须使用事务
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 1. 创建多语言名称（使用 tx 创建 Repository）
		multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
		if _, err := multiLanguageNameRepo.CreateMultiLanguageName(multiLanguageName); err != nil {
			return errors.WithMessage(err, "创建多语言名称失败")
		}

		// 2. 创建活动（使用 tx 创建 Repository）
		activity := &model.FullReductionActivity{
			BaseModel: model.BaseModel{
				Uuid:       activityUuid,
				CreateTime: currentTime,
				UpdateTime: currentTime,
			},
			Name:                  req.LocaleName.ToJson(),
			MultiLanguageNameUuid: multiLangUuid,
			StartDate:             req.StartDate,
			EndDate:               req.EndDate,
			StartTime:             req.StartTime,
			EndTime:               req.EndTime,
			IsAllDay:              req.IsAllDay,
			ReductionType:         req.ReductionType,
			IsDisabled:            constant.No,
		}

		activityRepo := repository.NewFullReductionActivityRepo(tx)
		if err := activityRepo.Create(activity); err != nil {
			return errors.WithMessage(err, "创建活动失败")
		}

		// 3. 创建规则（使用 tx 创建 Repository）
		ruleRepo := repository.NewFullReductionActivityRuleRepo(tx)
		for _, ruleReq := range rules {
			ruleUuid, err := utils.GetID()
			if err != nil {
				return errors.WithMessage(err, "生成UUID失败")
			}

			rule := &model.FullReductionActivityRule{
				BaseModel: model.BaseModel{
					Uuid:       ruleUuid,
					CreateTime: currentTime,
					UpdateTime: currentTime,
				},
				FullReductionActivityUuid: activityUuid,
				Threshold:                 ruleReq.Threshold,
				ReductionAmount:           ruleReq.ReductionAmount,
			}
			if err := ruleRepo.Create(rule); err != nil {
				return errors.WithMessage(err, "创建规则失败")
			}
		}

		// 返回 nil 自动提交，返回 error 自动回滚
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// Update 更新满减活动
func (s *fullReductionActivitySrv) Update(ctx context.Context, req *req.FullReductionActivityUpdateReq) error {
	// 获取商家时区
	companySetting := ctx.GetCompanySetting()
	timezone := (&companySetting).GetTimezone()
	// 验证请求并初始化日期字段
	if err := req.Validate(timezone); err != nil {
		return err
	}

	db := ctx.GetDB()
	currentTime := time.Now().Unix()

	// 查找活动
	activityRepo := repository.NewFullReductionActivityRepo(db)
	activity, err := activityRepo.GetByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查找活动失败")
	}
	if activity == nil {
		return errors.WithMessage(errors.New("活动不存在"), "活动不存在")
	}

	// 检查是否为总部来源数据
	if !isEditable(ctx, activity.HeadquarterUuid) {
		return errors.New("活动不可编辑")
	}

	// 检查活动状态：已结束和已失效的活动不可编辑
	if activity.IsDisabled == constant.Yes {
		return errors.WithMessage(errors.New("已失效的活动不可编辑"), "已失效的活动不可编辑")
	}
	now := time.Now().Unix()
	status := activity.GetStatus(now, "")
	if status == constant.ActivityStatusEnded {
		return errors.WithMessage(errors.New("已结束的活动不可编辑"), "已结束的活动不可编辑")
	}

	// 如果是阶梯满减，需要排序
	rules := req.Rules
	if req.ReductionType == constant.FullReductionTypeStep {
		sort.Slice(rules, func(i, j int) bool {
			return rules[i].Threshold < rules[j].Threshold
		})
	}

	// 多个数据库操作，必须使用事务
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 1. 更新多语言名称（使用 tx 创建 Repository）
		multiLanguageName := model.MultiLanguageName{
			BaseModel: model.BaseModel{
				Uuid:       activity.MultiLanguageNameUuid,
				UpdateTime: currentTime,
			},
		}
		multiLanguageName.InitByLocaleResponse(req.LocaleName)

		multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
		if err := multiLanguageNameRepo.UpdateMultiLanguageName(activity.MultiLanguageNameUuid, multiLanguageName); err != nil {
			return errors.WithMessage(err, "更新多语言名称失败")
		}

		// 2. 更新活动（使用 tx 创建 Repository）
		activity.Name = req.LocaleName.ToJson()
		activity.StartDate = req.StartDate
		activity.EndDate = req.EndDate
		activity.StartTime = req.StartTime
		activity.EndTime = req.EndTime
		activity.IsAllDay = req.IsAllDay
		activity.ReductionType = req.ReductionType
		activity.UpdateTime = currentTime

		if req.IsAllDay == constant.Yes {
			activity.StartTime = ""
			activity.EndTime = ""
		}

		activityRepoTx := repository.NewFullReductionActivityRepo(tx)
		if err := activityRepoTx.UpdateActivity(activity); err != nil {
			return errors.WithMessage(err, "更新活动失败")
		}

		// 3. 删除旧规则（使用 tx 创建 Repository）
		ruleRepo := repository.NewFullReductionActivityRuleRepo(tx)
		if err := ruleRepo.DeleteByFullReductionActivityUuid(req.Uuid); err != nil {
			return errors.WithMessage(err, "删除旧规则失败")
		}

		// 4. 创建新规则（使用 tx 创建 Repository）
		for _, ruleReq := range rules {
			ruleUuid, err := utils.GetID()
			if err != nil {
				return errors.WithMessage(err, "生成UUID失败")
			}

			rule := &model.FullReductionActivityRule{
				BaseModel: model.BaseModel{
					Uuid:       ruleUuid,
					CreateTime: currentTime,
					UpdateTime: currentTime,
				},
				FullReductionActivityUuid: req.Uuid,
				Threshold:                 ruleReq.Threshold,
				ReductionAmount:           ruleReq.ReductionAmount,
			}
			if err := ruleRepo.Create(rule); err != nil {
				return errors.WithMessage(err, "创建规则失败")
			}
		}

		// 返回 nil 自动提交，返回 error 自动回滚
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// GetByUuid 根据UUID获取满减活动
func (s *fullReductionActivitySrv) GetByUuid(ctx context.Context, uuid uint64) (*resp.FullReductionActivityResp, error) {
	db := ctx.GetDB()

	activityRepo := repository.NewFullReductionActivityRepo(db)
	activity, err := activityRepo.GetByUuid(uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取活动失败")
	}
	if activity == nil {
		return nil, errors.WithMessage(errors.New("活动不存在"), "活动不存在")
	}

	return s.buildResp(ctx, activity)
}

// GetList 获取满减活动列表
func (s *fullReductionActivitySrv) GetList(ctx context.Context, req *req.FullReductionActivityListReq) (*resp.FullReductionActivityListResp, error) {
	db := ctx.GetDB()
	now := time.Now().Unix()

	activityRepo := repository.NewFullReductionActivityRepo(db)

	// 构建查询选项
	opts := []repository.DBOption{
		repository.CommonRepo.WhereBySoftDelete(),
	}

	// 根据状态筛选
	if req.Status != "" && req.Status != constant.ActivityStatusAll {
		opts = append(opts, activityRepo.WhereStatus(req.Status, now))
	}

	// 添加分页选项
	offset := (req.PageNo - 1) * req.PageSize
	opts = append(opts, activityRepo.Offset(offset))
	opts = append(opts, activityRepo.Limit(req.PageSize))

	// 获取列表
	activities, total, err := activityRepo.GetList(opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "获取活动列表失败")
	}

	// 构建响应
	list := make([]*resp.FullReductionActivityResp, 0, len(activities))
	for _, activity := range activities {
		respItem, err := s.buildResp(ctx, activity)
		if err != nil {
			return nil, errors.WithMessage(err, "构建响应失败")
		}
		list = append(list, respItem)
	}

	return &resp.FullReductionActivityListResp{
		List: list,
		Meta: &resp.PageMeta{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// Delete 删除满减活动
func (s *fullReductionActivitySrv) Delete(ctx context.Context, uuid uint64) error {
	db := ctx.GetDB()

	activityRepo := repository.NewFullReductionActivityRepo(db)
	activity, err := activityRepo.GetByUuid(uuid)
	if err != nil {
		return errors.WithMessage(err, "查找活动失败")
	}
	if activity == nil {
		return errors.WithMessage(errors.New("活动不存在"), "活动不存在")
	}

	// 检查是否为总部来源数据
	if !isEditable(ctx, activity.HeadquarterUuid) {
		return errors.New("活动不可删除")
	}

	// 检查活动状态，进行中的活动不可删除
	now := time.Now().Unix()
	status := activity.GetStatus(now, "")
	if status == constant.ActivityStatusInProgress {
		return errors.WithMessage(errors.New("进行中的活动不可删除"), "进行中的活动不可删除")
	}

	// 多个数据库操作，必须使用事务
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 1. 删除活动（使用 tx 创建 Repository）
		activityRepoTx := repository.NewFullReductionActivityRepo(tx)
		if err := activityRepoTx.Delete(uuid); err != nil {
			return errors.WithMessage(err, "删除活动失败")
		}

		// 2. 删除规则（使用 tx 创建 Repository）
		ruleRepo := repository.NewFullReductionActivityRuleRepo(tx)
		if err := ruleRepo.DeleteByFullReductionActivityUuid(uuid); err != nil {
			return errors.WithMessage(err, "删除规则失败")
		}

		// 返回 nil 自动提交，返回 error 自动回滚
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// Disable 失效满减活动
func (s *fullReductionActivitySrv) Disable(ctx context.Context, uuid uint64) error {
	db := ctx.GetDB()
	currentTime := time.Now().Unix()

	activityRepo := repository.NewFullReductionActivityRepo(db)
	activity, err := activityRepo.GetByUuid(uuid)
	if err != nil {
		return errors.WithMessage(err, "查找活动失败")
	}
	if activity == nil {
		return errors.WithMessage(errors.New("活动不存在"), "活动不存在")
	}

	// 检查活动状态，已结束的活动不可失效
	now := time.Now().Unix()
	status := activity.GetStatus(now, "")
	if status == constant.ActivityStatusEnded {
		return errors.WithMessage(errors.New("已结束的活动不可失效"), "已结束的活动不可失效")
	}

	// 更新活动状态
	activity.IsDisabled = constant.Yes
	activity.UpdateTime = currentTime

	if err := activityRepo.UpdateDisabled(activity); err != nil {
		return errors.WithMessage(err, "失效活动失败")
	}

	return nil
}

// buildResp 构建响应对象
func (s *fullReductionActivitySrv) buildResp(ctx context.Context, activity *model.FullReductionActivity) (*resp.FullReductionActivityResp, error) {
	// 从 MultiLanguageName 转换为 LocaleResponse（必须使用 LocaleResponse）
	var name dto.LocaleResponse
	if activity.MultiLanguageName.Uuid > 0 {
		name = activity.MultiLanguageName.GetNames() // ✅ 使用 GetNames() 方法转换为 LocaleResponse
	}

	// 构建规则响应
	rules := make([]resp.FullReductionActivityRuleResp, 0, len(activity.Rules))
	for _, rule := range activity.Rules {
		rules = append(rules, resp.FullReductionActivityRuleResp{
			Uuid:            rule.Uuid,
			Threshold:       rule.Threshold,
			ReductionAmount: rule.ReductionAmount,
		})
	}

	// 获取商家时区
	companySetting := ctx.GetCompanySetting()
	timezone := (&companySetting).GetTimezone()

	// 将时间戳转换为日期字符串
	timeUtil := utils.Timezone(timezone)
	startDateStr := ""
	endDateStr := ""
	if activity.StartDate > 0 {
		startDateStr = timeUtil.FormatUnixTime(activity.StartDate, "2006/01/02")
	}
	if activity.EndDate > 0 {
		endDateStr = timeUtil.FormatUnixTime(activity.EndDate, "2006/01/02")
	}

	// 获取活动状态
	now := time.Now().Unix()
	status := activity.GetStatus(now, timezone)

	// 获取满减方式名称
	reductionTypeName := constant.FullReductionTypeNames[activity.ReductionType]
	if reductionTypeName == "" {
		reductionTypeName = "未知类型"
	}

	return &resp.FullReductionActivityResp{
		Uuid:              activity.Uuid,
		LocaleName:        name, // ✅ 使用 LocaleResponse
		StartDate:         activity.StartDate,
		EndDate:           activity.EndDate,
		StartDateStr:      startDateStr,
		EndDateStr:        endDateStr,
		StartTime:         activity.StartTime,
		EndTime:           activity.EndTime,
		IsAllDay:          activity.IsAllDay,
		ReductionType:     activity.ReductionType,
		ReductionTypeName: reductionTypeName,
		IsDisabled:        activity.IsDisabled,
		Status:            status,
		Rules:             rules,
		IsEditable:        isEditable(ctx, activity.HeadquarterUuid),
	}, nil
}
