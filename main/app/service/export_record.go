package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"
)

// IExportRecordSrv 导出记录服务接口
type IExportRecordSrv interface {
	GetExportRecordList(ctx context.Context, listReq req.ExportRecordListReq) (resp.ExportRecordListPaginationResp, error) // 获取导出记录列表
	DeleteExportRecords(ctx context.Context, deleteReq req.ExportRecordDeleteReq) error                                    // 批量删除导出记录
}

// exportRecordSrv 导出记录服务实现
type exportRecordSrv struct {
	dbm *database.DBManager
}

// NewExportRecordSrv 创建导出记录服务
func NewExportRecordSrv(dbm *database.DBManager) IExportRecordSrv {
	return &exportRecordSrv{
		dbm: dbm,
	}
}

// GetExportRecordList 获取导出记录列表
func (s *exportRecordSrv) GetExportRecordList(ctx context.Context, listReq req.ExportRecordListReq) (resp.ExportRecordListPaginationResp, error) {
	var result resp.ExportRecordListPaginationResp

	// 获取数据库连接
	db := s.dbm.GetDB(ctx.GetDbId())
	if db == nil {
		return result, errors.New("数据库连接失败")
	}

	// 创建Repository
	exportRecordRepo := repository.NewExportRecordRepo(db)

	// 构建查询选项
	opts := []repository.DBOption{
		exportRecordRepo.WhereNotDeleted(),                   // 未删除
		exportRecordRepo.WhereExportType(listReq.ExportType), // 按类型筛选（如果传了类型）
		exportRecordRepo.OrderByCreateTime(true),             // 按创建时间倒序
		exportRecordRepo.PreloadFile(),                       // 预加载文件信息
	}

	// 分页查询
	records, total, err := exportRecordRepo.GetListWithPagination(listReq.PageNo, listReq.PageSize, opts...)
	if err != nil {
		return result, errors.WithMessage(err, "查询导出记录列表失败")
	}

	// 转换为响应格式
	list := make([]resp.ExportRecordListResp, 0, len(records))
	for _, record := range records {
		list = append(list, s.convertToListResp(record, utils.GetBaseURL(ctx.GetGin().Request)))
	}

	// 构建响应
	result.List = list
	result.Meta.PageNo = listReq.PageNo
	result.Meta.PageSize = listReq.PageSize
	result.Meta.Total = total

	return result, nil
}

// DeleteExportRecords 批量删除导出记录
func (s *exportRecordSrv) DeleteExportRecords(ctx context.Context, deleteReq req.ExportRecordDeleteReq) error {
	// 获取数据库连接
	db := s.dbm.GetDB(ctx.GetDbId())
	if db == nil {
		return errors.New("数据库连接失败")
	}

	// 创建Repository
	exportRecordRepo := repository.NewExportRecordRepo(db)

	// 批量查询要删除的记录
	records, err := exportRecordRepo.GetByUuids(deleteReq.Uuids, exportRecordRepo.WhereNotDeleted())
	if err != nil {
		return errors.WithMessage(err, "查询导出记录失败")
	}

	// 检查是否有导出中的记录
	for _, record := range records {
		if record.Status == constant.ExportRecordStatusExporting {
			return errors.New("导出中的记录不能删除")
		}
	}

	// 批量软删除
	err = exportRecordRepo.BatchDelete(deleteReq.Uuids)
	if err != nil {
		return errors.WithMessage(err, "删除导出记录失败")
	}

	return nil
}

// convertToListResp 转换为列表响应格式
func (s *exportRecordSrv) convertToListResp(record model.ExportRecord, baseUrl string) resp.ExportRecordListResp {
	result := resp.ExportRecordListResp{
		Uuid:       record.Uuid,
		ExportType: record.ExportType,
		ExportName: record.ExportName,
		FileUuid:   record.FileUuid,
		Status:     record.Status,
		ErrorMsg:   record.ErrorMsg,
		CreateTime: record.CreateTime,
		UpdateTime: record.UpdateTime,
	}

	// 从关联的File对象中获取文件信息
	if record.File != nil {
		result.FileName = record.File.RealName
		result.FileSize = record.File.FileSize
		// 获取文件下载URL（使用空baseUrl，实际使用时可能需要配置）
		result.FileUrl = record.File.GetUrl(baseUrl)
	}

	return result
}
