package manufacturing

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/manufacturing"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	Bom = new(sBom)
)

type sBom struct {
}

func (s *sBom) GetBomList(ctx context.Context, req *manufacturing.GetBomListReq) (res *manufacturing.GetBomListResp, err error) {
	var filters = make([][]string, 0)
	if len(req.CompanyAbbr) > 0 {
		filters = append(filters, []string{"company_abbr", "=", req.CompanyAbbr})
	}
	if len(req.Branch) > 0 {
		filters = append(filters, []string{"branch", "=", req.Branch})
	}
	if len(req.ItemCode) > 0 {
		filters = append(filters, []string{"item_code", "=", req.ItemCode})
	}
	if len(req.ItemName) > 0 {
		filters = append(filters, []string{"item_name", "like", req.ItemName})
	}
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: "BOM",
	}, &erp.RequestParams{
		Fields:  []string{"name", "item", "qty"},
		Filters: filters,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "获取BOM列表失败")
	}
	if j, err := gjson.DecodeToJson(resp.Bytes()); err == nil {
		// 遍历j.Get("data") 返回的数组字段，设置到 BomList 中
		bomList := make([]*manufacturing.BomInfo, 0)
		dataArray := j.GetJsons("data")
		for _, item := range dataArray {

			bomList = append(bomList, &manufacturing.BomInfo{
				ItemCode: item.Get("item").String(),
			})
		}
		res = &manufacturing.GetBomListResp{
			BomList: bomList,
		}
	} else {
		g.Log().Error(ctx, "解析BOM列表失败", err)
	}

	return
}
