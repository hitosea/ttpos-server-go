package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type IPrinterSrv interface {
	GetProductPrinterList(ctx context.Context) (resp.ProductPrinterList, error) // 获取打印档口列表
}

type printerSrv struct {
	dbm *database.DBManager
}

func NewPrinterSrv(dbm *database.DBManager) IPrinterSrv {
	return NewPrinterSrvImpl(dbm)
}

func NewPrinterSrvImpl(dbm *database.DBManager) IPrinterSrv {
	return &printerSrv{
		dbm: dbm,
	}
}

func (s *printerSrv) GetProductPrinterList(ctx context.Context) (resp.ProductPrinterList, error) {
	productPrinterRepo := repository.NewProductPrinterRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	printers, err := productPrinterRepo.GetProductPrinters(productPrinterRepo.WhereStatus(constant.ProductPrinterStatusOpen))
	if err != nil {
		return resp.ProductPrinterList{List: make([]resp.ProductPrinter, 0)}, errors.ErrInternal
	}
	productPrinters := make([]resp.ProductPrinter, 0, len(printers))
	for _, printer := range printers {
		productPrinters = append(productPrinters, resp.ProductPrinter{
			Uuid: printer.Uuid,
			Name: printer.Name,
		})
	}
	return resp.ProductPrinterList{List: productPrinters}, nil
}
