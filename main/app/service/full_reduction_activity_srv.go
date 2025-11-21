package service

import (
	"sort"
	"time"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"
)

// IFullReductionActivitySrv 满减活动服务接口
type IFullReductionActivitySrv interface {
	Create(ctx context.Context, req *req.FullReductionActivityCreateReq) (*resp.FullReductionActivityResp, error)
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
func (s *fullReductionActivitySrv) Create(ctx context.Context, req *req.FullReductionActivityCreateReq) (*resp.FullReductionActivityResp, error) {
	db := ctx.GetDB()
	currentTime := time.Now().Unix()

	// 生成多语言名称UUID
	multiLangUuid, err := utils.GetID()
	if err != nil {
		return nil, errors.WithMessage(err, "生成UUID失败")
	}

	// 创建多语言名称
	multiLanguageName := model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid:       multiLangUuid,
			CreateTime: currentTime,
			UpdateTime: currentTime,
		},
	}
	multiLanguageName.InitByLocaleResponseJson(req.Name)

	if err := db.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error; err != nil {
		return nil, errors.WithMessage(err, "创建多语言名称失败")
	}

	// 生成活动UUID
	activityUuid, err := utils.GetID()
	if err != nil {
		return nil, errors.WithMessage(err, "生成UUID失败")
	}

	// 创建活动
	activity := &model.FullReductionActivity{
		BaseModel: model.BaseModel{
			Uuid:       activityUuid,
			CreateTime: currentTime,
			UpdateTime: currentTime,
		},
		Name:                  req.Name,
		MultiLanguageNameUuid: multiLangUuid,
		StartDate:             req.StartDate,
		EndDate:               req.EndDate,
		StartTime:             req.StartTime,
		EndTime:               req.EndTime,
		IsAllDay:              req.IsAllDay,
		ReductionType:         req.ReductionType,
		IsDisabled:            0,
	}

	activityRepo := repository.NewFullReductionActivityRepo(db)
	if err := activityRepo.Create(activity); err != nil {
		return nil, errors.WithMessage(err, "创建活动失败")
	}

	// 创建规则
	// 如果是阶梯满减，需要排序
	rules := req.Rules
	if req.ReductionType == 0 { // 阶梯满减
		sort.Slice(rules, func(i, j int) bool {
			return rules[i].Threshold < rules[j].Threshold
		})
	}

	ruleRepo := repository.NewFullReductionActivityRuleRepo(db)
	for _, ruleReq := range rules {
		ruleUuid, err := utils.GetID()
		if err != nil {
			return nil, errors.WithMessage(err, "生成UUID失败")
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
			return nil, errors.WithMessage(err, "创建规则失败")
		}
	}

	// 重新加载活动（包含规则）
	activity, err = activityRepo.GetByUuid(activityUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取活动失败")
	}
	if activity == nil {
		return nil, errors.New("活动不存在")
	}

	// 返回响应
	return s.buildResp(ctx, activity)
}

// Update 更新满减活动
func (s *fullReductionActivitySrv) Update(ctx context.Context, req *req.FullReductionActivityUpdateReq) error {
	db := ctx.GetDB()
	currentTime := time.Now().Unix()

	// 查找活动
	activityRepo := repository.NewFullReductionActivityRepo(db)
	activity, err := activityRepo.GetByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查找活动失败")
	}
	if activity == nil {
		return errors.New("活动不存在")
	}

	// 更新多语言名称
	multiLanguageName := model.NewMultiLanguageName(req.Name)
	err = db.Model(&model.MultiLanguageName{}).
		Where("uuid = ?", activity.MultiLanguageNameUuid).
		Updates(map[string]interface{}{
			"zh_name":     multiLanguageName.ZhName,
			"en_name":     multiLanguageName.EnName,
			"th_name":     multiLanguageName.ThName,
			"zh_tw_name":  multiLanguageName.ZhTwName,
			"my_name":     multiLanguageName.MyName,
			"ja_name":     multiLanguageName.JaName,
			"ko_name":     multiLanguageName.KoName,
			"tr_name":     multiLanguageName.TrName,
			"sv_name":     multiLanguageName.SvName,
			"update_time": currentTime,
		}).Error
	if err != nil {
		return errors.WithMessage(err, "更新多语言名称失败")
	}

	// 更新活动
	activity.Name = req.Name
	activity.StartDate = req.StartDate
	activity.EndDate = req.EndDate
	activity.StartTime = req.StartTime
	activity.EndTime = req.EndTime
	activity.IsAllDay = req.IsAllDay
	activity.ReductionType = req.ReductionType
	activity.UpdateTime = currentTime

	if err := activityRepo.Update(activity); err != nil {
		return errors.WithMessage(err, "更新活动失败")
	}

	// 删除旧规则
	ruleRepo := repository.NewFullReductionActivityRuleRepo(db)
	if err := ruleRepo.DeleteByFullReductionActivityUuid(req.Uuid); err != nil {
		return errors.WithMessage(err, "删除旧规则失败")
	}

	// 创建新规则
	// 如果是阶梯满减，需要排序
	rules := req.Rules
	if req.ReductionType == 0 { // 阶梯满减
		sort.Slice(rules, func(i, j int) bool {
			return rules[i].Threshold < rules[j].Threshold
		})
	}

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
		return nil, errors.New("活动不存在")
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
	if req.Status != "" && req.Status != "all" {
		opts = append(opts, activityRepo.WhereStatus(req.Status, now))
	}

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
		return errors.New("活动不存在")
	}

	// 检查活动状态，进行中的活动不可删除
	now := time.Now().Unix()
	status := activity.GetStatus(now, "")
	if status == "ongoing" {
		return errors.New("进行中的活动不可删除")
	}

	// 删除活动
	if err := activityRepo.Delete(uuid); err != nil {
		return errors.WithMessage(err, "删除活动失败")
	}

	// 删除规则
	ruleRepo := repository.NewFullReductionActivityRuleRepo(db)
	if err := ruleRepo.DeleteByFullReductionActivityUuid(uuid); err != nil {
		return errors.WithMessage(err, "删除规则失败")
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
		return errors.New("活动不存在")
	}

	// 检查活动状态，已结束的活动不可失效
	now := time.Now().Unix()
	status := activity.GetStatus(now, "")
	if status == "ended" {
		return errors.New("已结束的活动不可失效")
	}

	// 更新活动状态
	activity.IsDisabled = 1
	activity.UpdateTime = currentTime

	if err := activityRepo.Update(activity); err != nil {
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

	// 获取活动状态
	now := time.Now().Unix()
	timezone := "" // 可以从 ctx 获取时区
	status := activity.GetStatus(now, timezone)

	// 获取满减方式名称
	reductionTypeName := "阶梯满减"
	if activity.ReductionType == 1 {
		reductionTypeName = "循环满减"
	}

	return &resp.FullReductionActivityResp{
		Uuid:              activity.Uuid,
		Name:              name, // ✅ 使用 LocaleResponse
		StartDate:         activity.StartDate,
		EndDate:           activity.EndDate,
		StartTime:         activity.StartTime,
		EndTime:           activity.EndTime,
		IsAllDay:          activity.IsAllDay,
		ReductionType:     activity.ReductionType,
		ReductionTypeName: reductionTypeName,
		IsDisabled:        activity.IsDisabled,
		Status:            status,
		Rules:             rules,
		CreateTime:        activity.CreateTime,
		UpdateTime:        activity.UpdateTime,
	}, nil
}
