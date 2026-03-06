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

// INationalitySrv 国籍服务接口
//
// 任务: story-order-source-nationality Phase 2.6
// 需求: R2.1-R2.7
//
// @version v2.10.0
type INationalitySrv interface {
	GetList(ctx context.Context) (resp.NationalityListResp, error)
	Create(ctx context.Context, req req.NationalityCreateReq) (resp.NationalityCreateResp, error)
	Update(ctx context.Context, req req.NationalityUpdateReq) error
	Delete(ctx context.Context, req req.NationalityDeleteReq) error
	CheckCanDelete(ctx context.Context, uuid uint64) error
}

// nationalitySrv 国籍服务实现
type nationalitySrv struct {
	dbm *database.DBManager // 数据库管理器
}

// NewNationalitySrv 创建国籍服务
func NewNationalitySrv(dbm *database.DBManager) INationalitySrv {
	return NewNationalitySrvImpl(dbm)
}

// NewNationalitySrvImpl 创建国籍服务实现
func NewNationalitySrvImpl(dbm *database.DBManager) INationalitySrv {
	return &nationalitySrv{
		dbm: dbm,
	}
}

// GetList 获取国籍列表
func (s *nationalitySrv) GetList(ctx context.Context) (resp.NationalityListResp, error) {
	db := ctx.GetDB()

	// 获取所有国籍
	nationalityRepo := repository.NewNationalityRepo(db)
	nationalities, err := nationalityRepo.FindList()
	if err != nil {
		return resp.NationalityListResp{}, errors.WithMessage(err)
	}

	// 构建响应数据
	list := make([]resp.NationalityItem, 0, len(nationalities))
	for _, nationality := range nationalities {
		list = append(list, resp.NationalityItem{
			Uuid:       nationality.Uuid,
			LocaleName: nationality.MultiLanguageName.GetNames(),
			Sort:       nationality.Sort,
			Status:     nationality.Status,
		})
	}

	return resp.NationalityListResp{
		List: list,
	}, nil
}

// Create 创建国籍
func (s *nationalitySrv) Create(ctx context.Context, req req.NationalityCreateReq) (resp.NationalityCreateResp, error) {
	// 验证请求
	if err := req.Validate(); err != nil {
		return resp.NationalityCreateResp{}, err
	}

	db := ctx.GetDB()
	currentTime := int64(time.Now().Unix())

	// 生成UUID
	nationalityUuid, err := utils.GetID()
	if err != nil {
		return resp.NationalityCreateResp{}, errors.WithMessage(err, "生成UUID失败")
	}

	multiLangUuid, err := utils.GetID()
	if err != nil {
		return resp.NationalityCreateResp{}, errors.WithMessage(err, "生成UUID失败")
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

	multiLanguageNameUuid, err := repository.NewMultiLanguageNameRepo(db).CreateMultiLanguageName(multiLanguageName)
	if err != nil {
		return resp.NationalityCreateResp{}, errors.WithMessage(err, "创建多语言名称失败")
	}

	// 创建国籍
	nationality := model.Nationality{
		BaseModel: model.BaseModel{
			Uuid:       nationalityUuid,
			CreateTime: currentTime,
			UpdateTime: currentTime,
		},
		MultiLanguageNameUuid: multiLanguageNameUuid,
		Sort:                  req.Sort,
		Status:                1, // 默认启用
	}

	nationalityRepo := repository.NewNationalityRepo(db)
	uuid, err := nationalityRepo.Create(nationality)
	if err != nil {
		return resp.NationalityCreateResp{}, errors.WithMessage(err, "创建国籍失败")
	}

	return resp.NationalityCreateResp{
		Uuid: uuid,
	}, nil
}

// Update 更新国籍
func (s *nationalitySrv) Update(ctx context.Context, req req.NationalityUpdateReq) error {
	// 验证请求
	if err := req.Validate(); err != nil {
		return err
	}

	db := ctx.GetDB()
	currentTime := int64(time.Now().Unix())

	// 查找国籍
	nationalityRepo := repository.NewNationalityRepo(db)
	nationality, err := nationalityRepo.FindByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查找国籍失败")
	}
	if nationality == nil {
		return errors.New("国籍不存在")
	}

	// 更新多语言名称
	err = repository.NewMultiLanguageNameRepo(db).UpdateMultiLanguageNameData(
		req.LocaleName.ToMultiLanguageUpdateMap(currentTime),
		repository.CommonRepo.WhereByUuid(nationality.MultiLanguageNameUuid),
	)
	if err != nil {
		return errors.WithMessage(err, "更新多语言名称失败")
	}

	// 更新国籍
	nationality.Sort = req.Sort
	nationality.Status = req.Status
	nationality.UpdateTime = currentTime

	err = nationalityRepo.Update(*nationality)
	if err != nil {
		return errors.WithMessage(err, "更新国籍失败")
	}

	return nil
}

// Delete 删除国籍
func (s *nationalitySrv) Delete(ctx context.Context, req req.NationalityDeleteReq) error {
	db := ctx.GetDB()

	// 软删除国籍（移除CheckCanDelete检查，允许删除已使用的配置）
	// 历史订单仍可通过 GetByUuidWithDeleted 查询到配置名称
	nationalityRepo := repository.NewNationalityRepo(db)
	err := nationalityRepo.SoftDelete(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "删除国籍失败")
	}

	return nil
}

// CheckCanDelete 校验是否可删除
func (s *nationalitySrv) CheckCanDelete(ctx context.Context, uuid uint64) error {
	db := ctx.GetDB()

	// 统计使用该国籍的订单数量
	nationalityRepo := repository.NewNationalityRepo(db)
	count, err := nationalityRepo.CountOrdersByNationalityUuid(uuid)
	if err != nil {
		return errors.WithMessage(err, "统计订单数量失败")
	}

	if count > 0 {
		return errors.New("该国籍已被订单使用，无法删除")
	}

	return nil
}
