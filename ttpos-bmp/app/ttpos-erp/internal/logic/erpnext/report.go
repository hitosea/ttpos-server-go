package erpnext

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

var (
	reportApiUrl = "/api/method/frappe.desk.query_report.run"
	Report       = new(sReport)
)

type sReport struct {
}

func init() {
	service.RegisterReport(Report)
}

func (s *sReport) Run(ctx context.Context, params *dto.ReportParams) (rst *g.Var, err error) {
	rst = GetClient(ctx).GetVar(ctx, reportApiUrl, params)
	err = detectError(rst)
	return
}
