package service

import (
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// IDeskManageSrv 桌台管理服务接口
type IDeskManageSrv interface {
	// 区域管理
	GetDeskRegionList(ctx context.Context) (*resp.DeskRegionListResp, error)
	AddDeskRegion(ctx context.Context, addReq req.DeskRegionAddReq) error
	EditDeskRegion(ctx context.Context, editReq req.DeskRegionEditReq) error
	DeleteDeskRegion(ctx context.Context, deleteReq req.DeskRegionDeleteReq) error

	// 类型管理
	GetDeskTypeList(ctx context.Context) (*resp.DeskTypeListResp, error)
	AddDeskType(ctx context.Context, addReq req.DeskTypeAddReq) error
	EditDeskType(ctx context.Context, editReq req.DeskTypeEditReq) error
	DeleteDeskType(ctx context.Context, deleteReq req.DeskTypeDeleteReq) error

	// 桌台管理
	GetDeskManageList(ctx context.Context, listReq req.DeskManageListReq) (*resp.DeskManageListResp, error)
	AddDesk(ctx context.Context, addReq req.DeskAddReq) error
	EditDesk(ctx context.Context, editReq req.DeskEditReq) error
	DeleteDesk(ctx context.Context, deleteReq req.DeskDeleteReq) error
}

// NewDeskManageSrv 创建桌台管理服务
func NewDeskManageSrv(dbm *database.DBManager) IDeskManageSrv {
	return &DeskManageSrvImpl{dbm: dbm}
}

// DeskManageSrvImpl 桌台管理服务实现
type DeskManageSrvImpl struct {
	dbm *database.DBManager
}

// ==================== 区域管理 ====================

// GetDeskRegionList 获取区域列表
func (s *DeskManageSrvImpl) GetDeskRegionList(ctx context.Context) (*resp.DeskRegionListResp, error) {
	db := ctx.GetDB()
	regionRepo := repository.NewDeskRegionRepo(db)
	deskRepo := repository.NewDeskRepo(db)

	// 获取区域列表
	regions, err := regionRepo.GetDeskRegionList()
	if err != nil {
		return nil, errors.WithMessage(err, "获取区域列表失败")
	}

	// 获取各区域桌台数量
	countMap, err := deskRepo.GetDeskCountsByRegion()
	if err != nil {
		return nil, errors.WithMessage(err, "获取桌台数量失败")
	}

	// 构建响应
	list := make([]resp.DeskRegionItem, 0, len(regions))
	for _, region := range regions {
		list = append(list, resp.DeskRegionItem{
			Uuid:      region.Uuid,
			Name:      region.Name,
			Sort:      region.Sort,
			DeskCount: countMap[region.Uuid],
		})
	}

	return &resp.DeskRegionListResp{List: list}, nil
}

// AddDeskRegion 添加区域
func (s *DeskManageSrvImpl) AddDeskRegion(ctx context.Context, addReq req.DeskRegionAddReq) error {
	db := ctx.GetDB()
	regionRepo := repository.NewDeskRegionRepo(db)

	region := model.DeskRegion{
		Name: addReq.Name,
		Sort: addReq.Sort,
	}

	_, err := regionRepo.CreateDeskRegion(region)
	if err != nil {
		return errors.WithMessage(err, "创建区域失败")
	}

	return nil
}

// EditDeskRegion 编辑区域
func (s *DeskManageSrvImpl) EditDeskRegion(ctx context.Context, editReq req.DeskRegionEditReq) error {
	db := ctx.GetDB()
	regionRepo := repository.NewDeskRegionRepo(db)

	region := model.DeskRegion{
		Name: editReq.Name,
		Sort: editReq.Sort,
	}

	err := regionRepo.UpdateDeskRegion(uint(editReq.Uuid), region)
	if err != nil {
		return errors.WithMessage(err, "更新区域失败")
	}

	return nil
}

// DeleteDeskRegion 删除区域
func (s *DeskManageSrvImpl) DeleteDeskRegion(ctx context.Context, deleteReq req.DeskRegionDeleteReq) error {
	db := ctx.GetDB()
	regionRepo := repository.NewDeskRegionRepo(db)
	deskRepo := repository.NewDeskRepo(db)

	// 检查该区域下是否有桌台
	countMap, err := deskRepo.GetDeskCountsByRegion()
	if err != nil {
		return errors.WithMessage(err, "检查桌台数量失败")
	}

	if countMap[deleteReq.Uuid] > 0 {
		return errors.New("该区域下存在桌台，无法删除")
	}

	err = regionRepo.DeleteDeskRegion(uint(deleteReq.Uuid))
	if err != nil {
		return errors.WithMessage(err, "删除区域失败")
	}

	return nil
}

// ==================== 类型管理 ====================

// GetDeskTypeList 获取类型列表
func (s *DeskManageSrvImpl) GetDeskTypeList(ctx context.Context) (*resp.DeskTypeListResp, error) {
	db := ctx.GetDB()
	typeRepo := repository.NewDeskTypeRepo(db)

	// 获取类型列表
	types, err := typeRepo.GetDeskTypeList()
	if err != nil {
		return nil, errors.WithMessage(err, "获取类型列表失败")
	}

	// 构建响应
	list := make([]resp.DeskTypeItem, 0, len(types))
	for _, t := range types {
		list = append(list, resp.DeskTypeItem{
			Uuid:     t.Uuid,
			Name:     t.Name,
			Sort:     t.Sort,
			RangeMin: t.RangeMin,
			RangeMax: t.RangeMax,
		})
	}

	return &resp.DeskTypeListResp{List: list}, nil
}

// AddDeskType 添加类型
func (s *DeskManageSrvImpl) AddDeskType(ctx context.Context, addReq req.DeskTypeAddReq) error {
	db := ctx.GetDB()
	typeRepo := repository.NewDeskTypeRepo(db)

	deskType := model.DeskType{
		Name:     addReq.Name,
		Sort:     addReq.Sort,
		RangeMin: addReq.RangeMin,
		RangeMax: addReq.RangeMax,
	}

	_, err := typeRepo.CreateDeskType(deskType)
	if err != nil {
		return errors.WithMessage(err, "创建类型失败")
	}

	return nil
}

// EditDeskType 编辑类型
func (s *DeskManageSrvImpl) EditDeskType(ctx context.Context, editReq req.DeskTypeEditReq) error {
	db := ctx.GetDB()
	typeRepo := repository.NewDeskTypeRepo(db)

	deskType := model.DeskType{
		Name:     editReq.Name,
		Sort:     editReq.Sort,
		RangeMin: editReq.RangeMin,
		RangeMax: editReq.RangeMax,
	}

	err := typeRepo.UpdateDeskType(uint(editReq.Uuid), deskType)
	if err != nil {
		return errors.WithMessage(err, "更新类型失败")
	}

	return nil
}

// DeleteDeskType 删除类型
func (s *DeskManageSrvImpl) DeleteDeskType(ctx context.Context, deleteReq req.DeskTypeDeleteReq) error {
	db := ctx.GetDB()
	typeRepo := repository.NewDeskTypeRepo(db)

	err := typeRepo.DeleteDeskType(uint(deleteReq.Uuid))
	if err != nil {
		return errors.WithMessage(err, "删除类型失败")
	}

	return nil
}

// ==================== 桌台管理 ====================

// GetDeskManageList 获取桌台管理列表
func (s *DeskManageSrvImpl) GetDeskManageList(ctx context.Context, listReq req.DeskManageListReq) (*resp.DeskManageListResp, error) {
	db := ctx.GetDB()
	deskRepo := repository.NewDeskRepo(db)
	regionRepo := repository.NewDeskRegionRepo(db)
	typeRepo := repository.NewDeskTypeRepo(db)

	// 获取桌台列表
	desks, total, err := deskRepo.GetDeskList(listReq.PageNo, listReq.PageSize)
	if err != nil {
		return nil, errors.WithMessage(err, "获取桌台列表失败")
	}

	// 获取区域信息
	regions, _ := regionRepo.GetDeskRegionList()
	regionMap := make(map[uint64]string)
	for _, r := range regions {
		regionMap[r.Uuid] = r.Name
	}

	// 获取类型信息
	types, _ := typeRepo.GetDeskTypeList()
	typeMap := make(map[uint64]string)
	for _, t := range types {
		typeMap[t.Uuid] = t.Name
	}

	// 构建响应
	list := make([]resp.DeskManageItem, 0, len(desks))
	for _, desk := range desks {
		list = append(list, resp.DeskManageItem{
			Uuid:                   desk.Uuid,
			DeskNo:                 desk.DeskNo,
			RegionUuid:             desk.RegionUuid,
			RegionName:             regionMap[desk.RegionUuid],
			TypeUuid:               desk.TypeUuid,
			TypeName:               typeMap[desk.TypeUuid],
			Sort:                   desk.Sort,
			Status:                 desk.Status,
			IsDisable:              desk.IsDisable,
			DefaultPeopleNum:       desk.DefaultPeopleNum,
			IsOpenDefaultPeopleNum: desk.IsOpenDefaultPeopleNum,
		})
	}

	return &resp.DeskManageListResp{List: list, Total: total}, nil
}

// AddDesk 添加桌台
func (s *DeskManageSrvImpl) AddDesk(ctx context.Context, addReq req.DeskAddReq) error {
	db := ctx.GetDB()
	deskRepo := repository.NewDeskRepo(db)

	desk := model.Desk{
		DeskNo:                 addReq.DeskNo,
		RegionUuid:             addReq.RegionUuid,
		TypeUuid:               addReq.TypeUuid,
		Sort:                   addReq.Sort,
		DefaultPeopleNum:       addReq.DefaultPeopleNum,
		IsOpenDefaultPeopleNum: addReq.IsOpenDefaultPeopleNum,
	}

	_, err := deskRepo.CreateDesk(desk)
	if err != nil {
		return errors.WithMessage(err, "创建桌台失败")
	}

	return nil
}

// EditDesk 编辑桌台
func (s *DeskManageSrvImpl) EditDesk(ctx context.Context, editReq req.DeskEditReq) error {
	db := ctx.GetDB()
	deskRepo := repository.NewDeskRepo(db)

	desk := model.Desk{
		DeskNo:                 editReq.DeskNo,
		RegionUuid:             editReq.RegionUuid,
		TypeUuid:               editReq.TypeUuid,
		Sort:                   editReq.Sort,
		IsDisable:              editReq.IsDisable,
		DefaultPeopleNum:       editReq.DefaultPeopleNum,
		IsOpenDefaultPeopleNum: editReq.IsOpenDefaultPeopleNum,
	}

	err := deskRepo.UpdateDesk(editReq.Uuid, desk)
	if err != nil {
		return errors.WithMessage(err, "更新桌台失败")
	}

	return nil
}

// DeleteDesk 删除桌台
func (s *DeskManageSrvImpl) DeleteDesk(ctx context.Context, deleteReq req.DeskDeleteReq) error {
	db := ctx.GetDB()
	deskRepo := repository.NewDeskRepo(db)

	// 检查桌台是否正在使用
	desk, err := deskRepo.GetDeskRecord(deleteReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "获取桌台信息失败")
	}

	if desk.Status != 0 {
		return errors.New("桌台正在使用中，无法删除")
	}

	err = deskRepo.DeleteDesk(deleteReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "删除桌台失败")
	}

	return nil
}
