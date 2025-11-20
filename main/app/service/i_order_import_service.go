package service

import (
	"mime/multipart"
	"ttpos-server-go/pkg/context"
)

// IOrderImportSrv 订单导入服务接口
type IOrderImportSrv interface {
	// Import 导入订单数据
	// file: Excel 文件
	// 返回: 成功数量、失败数量、失败列表
	Import(ctx context.Context, file *multipart.FileHeader) (successCount int, failCount int, failList []OrderImportFailItem, err error)
}

// OrderImportFailItem 导入失败项
type OrderImportFailItem struct {
	Row     int    // Excel 行号
	OrderNo string // 订单号
	Reason  string // 失败原因
}
