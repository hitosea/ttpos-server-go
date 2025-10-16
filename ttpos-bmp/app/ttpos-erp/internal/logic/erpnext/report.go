package erpnext

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
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

func (s *sReport) Run(ctx context.Context, params *erp.ReportParams) (rst *gjson.Json, err error) {
	resp := GetClient(ctx).GetVar(ctx, reportApiUrl, params)
	rst, err = detectError(resp)
	return
}
