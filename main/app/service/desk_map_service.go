package service

import (
	"encoding/json"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// IDeskMapSrv 桌台地图服务接口
//
// 任务: story-admin-desktop-table-map Phase 2.5
// 需求: R1.1-R1.6, R3.2-R3.3
//
// @version v2.10.0
type IDeskMapSrv interface {
	GetAreaListWithStatus(ctx context.Context) (resp.DeskMapAreaListResp, error)
	GetLayoutDetail(ctx context.Context, areaUuid uint64) (resp.DeskMapLayoutResp, error)
	SaveLayout(ctx context.Context, req req.DeskMapSaveLayoutReq) error
}

// deskMapSrv 桌台地图服务实现
type deskMapSrv struct {
	dbm *database.DBManager // 数据库管理器
}

// NewDeskMapSrv 创建桌台地图服务
func NewDeskMapSrv(dbm *database.DBManager) IDeskMapSrv {
	return NewDeskMapSrvImpl(dbm)
}

// NewDeskMapSrvImpl 创建桌台地图服务实现
func NewDeskMapSrvImpl(dbm *database.DBManager) IDeskMapSrv {
	return &deskMapSrv{
		dbm: dbm,
	}
}

// GetAreaListWithStatus 获取区域列表及地图配置状态
func (s *deskMapSrv) GetAreaListWithStatus(ctx context.Context) (resp.DeskMapAreaListResp, error) {
	db := ctx.GetDB()

	// 获取所有区域
	deskRegionRepo := repository.NewDeskRegionRepo(db)
	regions, err := deskRegionRepo.GetDeskRegionList()
	if err != nil {
		return resp.DeskMapAreaListResp{}, errors.WithMessage(err)
	}

	// 获取所有布局
	layoutRepo := repository.NewDeskMapLayoutRepo(db)
	layouts, err := layoutRepo.FindAll()
	if err != nil {
		return resp.DeskMapAreaListResp{}, errors.WithMessage(err)
	}

	// 创建布局映射表（region_uuid -> 是否已配置）
	layoutMap := make(map[uint64]bool)
	for _, layout := range layouts {
		layoutMap[layout.RegionUuid] = true
	}

	// 获取桌台数量（按区域分组）
	deskRepo := repository.NewDeskRepo(db)
	deskCountMap, err := deskRepo.GetDeskCountsByRegion()
	if err != nil {
		return resp.DeskMapAreaListResp{}, errors.WithMessage(err)
	}

	// 构建响应数据
	list := make([]resp.DeskMapAreaItem, 0, len(regions))
	for _, region := range regions {
		layoutStatus := "unset"
		if layoutMap[region.Uuid] {
			layoutStatus = "set"
		}

		list = append(list, resp.DeskMapAreaItem{
			RegionUuid:   region.Uuid,
			RegionName:   region.Name,
			DeskCount:    uint(deskCountMap[region.Uuid]),
			LayoutStatus: layoutStatus,
		})
	}

	return resp.DeskMapAreaListResp{
		List: list,
	}, nil
}

// GetLayoutDetail 获取某区域布局详情
func (s *deskMapSrv) GetLayoutDetail(ctx context.Context, areaUuid uint64) (resp.DeskMapLayoutResp, error) {
	db := ctx.GetDB()

	// 获取区域信息
	deskRegionRepo := repository.NewDeskRegionRepo(db)
	regions, err := deskRegionRepo.GetDeskRegionList()
	if err != nil {
		return resp.DeskMapLayoutResp{}, errors.WithMessage(err)
	}

	var targetRegion *model.DeskRegion
	for _, region := range regions {
		if region.Uuid == areaUuid {
			targetRegion = &region
			break
		}
	}

	if targetRegion == nil {
		return resp.DeskMapLayoutResp{}, errors.New("区域不存在")
	}

	// 获取该区域的所有桌台
	deskRepo := repository.NewDeskRepo(db)
	desks, err := deskRepo.GetDesksByRegionUuid(areaUuid)
	if err != nil {
		return resp.DeskMapLayoutResp{}, errors.WithMessage(err)
	}

	// 获取桌台类型信息（用于获取容量）
	deskTypeRepo := repository.NewDeskTypeRepo(db)
	types, err := deskTypeRepo.GetDeskTypeList()
	if err != nil {
		return resp.DeskMapLayoutResp{}, errors.WithMessage(err)
	}

	// 创建类型映射表
	typeMap := make(map[uint64]model.DeskType)
	for _, t := range types {
		typeMap[t.Uuid] = t
	}

	// 获取布局数据
	layoutRepo := repository.NewDeskMapLayoutRepo(db)
	layout, err := layoutRepo.FindByRegionUuid(areaUuid)
	if err != nil {
		return resp.DeskMapLayoutResp{}, errors.WithMessage(err)
	}

	// 解析布局 JSON
	layoutData := resp.DeskMapLayoutData{
		Desks: []resp.DeskMapLayoutTable{}, // 初始化为空数组，避免返回 null
	}
	selectedTableMap := make(map[uint64]bool)

	if layout != nil && layout.LayoutJson != "" {
		if err := json.Unmarshal([]byte(layout.LayoutJson), &layoutData); err != nil {
			return resp.DeskMapLayoutResp{}, errors.WithMessage(err)
		}

		// 创建已选中桌台映射表
		for _, table := range layoutData.Desks {
			selectedTableMap[table.DeskUuid] = true
		}
	}

	// 构建桌台列表
	tables := make([]resp.DeskMapTableItem, 0, len(desks))
	for _, desk := range desks {
		rangeMin := uint(0)
		rangeMax := uint(0)
		if deskType, ok := typeMap[desk.TypeUuid]; ok {
			rangeMin = deskType.RangeMin
			rangeMax = deskType.RangeMax
		}

		tables = append(tables, resp.DeskMapTableItem{
			DeskUuid: desk.Uuid,
			DeskName: desk.DeskNo,
			RangeMin: rangeMin,
			RangeMax: rangeMax,
			Selected: selectedTableMap[desk.Uuid],
		})
	}

	return resp.DeskMapLayoutResp{
		Region: resp.DeskMapAreaInfo{
			RegionUuid: targetRegion.Uuid,
			RegionName: targetRegion.Name,
		},
		Desks:  tables,
		Layout: layoutData,
	}, nil
}

// SaveLayout 保存区域布局
func (s *deskMapSrv) SaveLayout(ctx context.Context, req req.DeskMapSaveLayoutReq) error {
	db := ctx.GetDB()

	// 验证请求参数
	if err := req.Validate(); err != nil {
		return err
	}

	// 转换请求数据为响应结构（用于 JSON 序列化）
	layoutData := resp.DeskMapLayoutData{
		Desks: make([]resp.DeskMapLayoutTable, len(req.Desks)),
	}
	for i, desk := range req.Desks {
		layoutData.Desks[i] = resp.DeskMapLayoutTable{
			DeskUuid: desk.DeskUuid,
			Shape:    desk.Shape,
			RangeMin: desk.RangeMin,
			RangeMax: desk.RangeMax,
			X:        desk.X,
			Y:        desk.Y,
			Width:    desk.Width,
			Height:   desk.Height,
			Rotation: desk.Rotation,
		}
	}

	// 序列化为 JSON 字符串
	layoutJSON, err := json.Marshal(layoutData)
	if err != nil {
		return errors.New("布局数据序列化失败")
	}

	layoutRepo := repository.NewDeskMapLayoutRepo(db)

	// 检查是否已存在布局
	existingLayout, err := layoutRepo.FindByRegionUuid(req.RegionUuid)
	if err != nil {
		return err
	}

	if existingLayout == nil {
		// 创建新布局
		newLayout := model.DeskMapLayout{
			RegionUuid: req.RegionUuid,
			LayoutJson: string(layoutJSON),
		}
		_, err = layoutRepo.CreateLayout(newLayout)
		return err
	}

	// 更新现有布局
	existingLayout.LayoutJson = string(layoutJSON)
	return layoutRepo.UpdateLayout(req.RegionUuid, *existingLayout)
}
