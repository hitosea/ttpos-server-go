package service

import (
	"time"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"
)

// IOrderSourceSrv 外卖来源服务接口
//
// 任务: story-order-source-nationality Phase 2.3
// 需求: R1.1-R1.6
//
// @version v2.10.0
type IOrderSourceSrv interface {
	GetList(ctx context.Context) (resp.OrderSourceListResp, error)
	Create(ctx context.Context, req req.OrderSourceCreateReq) (resp.OrderSourceCreateResp, error)
	Update(ctx context.Context, req req.OrderSourceUpdateReq) error
	Delete(ctx context.Context, req req.OrderSourceDeleteReq) error
	CheckCanDelete(ctx context.Context, uuid uint64) error
}

// orderSourceSrv 外卖来源服务实现
type orderSourceSrv struct {
	dbm *database.DBManager // 数据库管理器
}

// NewOrderSourceSrv 创建外卖来源服务
func NewOrderSourceSrv(dbm *database.DBManager) IOrderSourceSrv {
	return NewOrderSourceSrvImpl(dbm)
}

// NewOrderSourceSrvImpl 创建外卖来源服务实现
func NewOrderSourceSrvImpl(dbm *database.DBManager) IOrderSourceSrv {
	return &orderSourceSrv{
		dbm: dbm,
	}
}

// GetList 获取外卖来源列表
func (s *orderSourceSrv) GetList(ctx context.Context) (resp.OrderSourceListResp, error) {
	db := ctx.GetDB()

	// 获取所有外卖来源
	orderSourceRepo := repository.NewOrderSourceRepo(db)
	orderSources, err := orderSourceRepo.FindList()
	if err != nil {
		return resp.OrderSourceListResp{}, errors.WithMessage(err)
	}

	// 构建响应数据
	list := make([]resp.OrderSourceItem, 0, len(orderSources))
	for _, orderSource := range orderSources {
		list = append(list, resp.OrderSourceItem{
			Uuid:       orderSource.Uuid,
			LocaleName: orderSource.MultiLanguageName.GetNames(),
			Sort:       orderSource.Sort,
			Status:     orderSource.Status,
		})
	}

	return resp.OrderSourceListResp{
		List: list,
	}, nil
}

// Create 创建外卖来源
func (s *orderSourceSrv) Create(ctx context.Context, req req.OrderSourceCreateReq) (resp.OrderSourceCreateResp, error) {
	// 验证请求
	if err := req.Validate(); err != nil {
		return resp.OrderSourceCreateResp{}, err
	}

	db := ctx.GetDB()
	currentTime := int64(time.Now().Unix())

	// 生成UUID
	orderSourceUuid, err := utils.GetID()
	if err != nil {
		return resp.OrderSourceCreateResp{}, errors.WithMessage(err, "生成UUID失败")
	}

	multiLangUuid, err := utils.GetID()
	if err != nil {
		return resp.OrderSourceCreateResp{}, errors.WithMessage(err, "生成UUID失败")
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

	if err := db.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error; err != nil {
		return resp.OrderSourceCreateResp{}, errors.WithMessage(err, "创建多语言名称失败")
	}

	// 创建外卖来源
	orderSource := model.OrderSource{
		BaseModel: model.BaseModel{
			Uuid:       orderSourceUuid,
			CreateTime: currentTime,
			UpdateTime: currentTime,
		},
		MultiLanguageNameUuid: multiLangUuid,
		Sort:                  req.Sort,
		Status:                1, // 默认启用
	}

	orderSourceRepo := repository.NewOrderSourceRepo(db)
	uuid, err := orderSourceRepo.Create(orderSource)
	if err != nil {
		return resp.OrderSourceCreateResp{}, errors.WithMessage(err, "创建外卖来源失败")
	}

	return resp.OrderSourceCreateResp{
		Uuid: uuid,
	}, nil
}

// Update 更新外卖来源
func (s *orderSourceSrv) Update(ctx context.Context, req req.OrderSourceUpdateReq) error {
	// 验证请求
	if err := req.Validate(); err != nil {
		return err
	}

	db := ctx.GetDB()
	currentTime := int64(time.Now().Unix())

	// 查找外卖来源
	orderSourceRepo := repository.NewOrderSourceRepo(db)
	orderSource, err := orderSourceRepo.FindByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查找外卖来源失败")
	}
	if orderSource == nil {
		return errors.New("外卖来源不存在")
	}

	// 更新多语言名称
	err = db.Model(&model.MultiLanguageName{}).
		Where("uuid = ?", orderSource.MultiLanguageNameUuid).
		Updates(map[string]interface{}{
			"zh_name":     req.LocaleName.ZH,
			"en_name":     req.LocaleName.EN,
			"th_name":     req.LocaleName.TH,
			"zh_tw_name":  req.LocaleName.ZHTW,
			"my_name":     req.LocaleName.MY,
			"ja_name":     req.LocaleName.JA,
			"ko_name":     req.LocaleName.KO,
			"tr_name":     req.LocaleName.TR,
			"sv_name":     req.LocaleName.SV,
			"update_time": currentTime,
		}).Error
	if err != nil {
		return errors.WithMessage(err, "更新多语言名称失败")
	}

	// 更新外卖来源
	orderSource.Sort = req.Sort
	orderSource.Status = req.Status
	orderSource.UpdateTime = currentTime

	err = orderSourceRepo.Update(*orderSource)
	if err != nil {
		return errors.WithMessage(err, "更新外卖来源失败")
	}

	return nil
}

// Delete 删除外卖来源
func (s *orderSourceSrv) Delete(ctx context.Context, req req.OrderSourceDeleteReq) error {
	// 校验是否可删除
	if err := s.CheckCanDelete(ctx, req.Uuid); err != nil {
		return err
	}

	db := ctx.GetDB()

	// 软删除外卖来源
	orderSourceRepo := repository.NewOrderSourceRepo(db)
	err := orderSourceRepo.SoftDelete(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "删除外卖来源失败")
	}

	return nil
}

// CheckCanDelete 校验是否可删除
func (s *orderSourceSrv) CheckCanDelete(ctx context.Context, uuid uint64) error {
	db := ctx.GetDB()

	// 统计使用该外卖来源的订单数量
	orderSourceRepo := repository.NewOrderSourceRepo(db)
	count, err := orderSourceRepo.CountOrdersBySourceUuid(uuid)
	if err != nil {
		return errors.WithMessage(err, "统计订单数量失败")
	}

	if count > 0 {
		return errors.New("该外卖来源已被订单使用，无法删除")
	}

	return nil
}
