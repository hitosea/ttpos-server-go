package selling

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	Selling = new(sSelling)
)

// sSelling 结构体定义
type sSelling struct{}

func init() {
	service.RegisterSelling(Selling)
}

// GetPosProfileList 查询Pos Profile列表
// 参数：ctx 上下文，req 查询请求
// 返回：erp.ResponseInfo，错误信息
func (s *sSelling) GetPosProfileList(ctx context.Context, req *selling.PosProfileReq) (res *selling.PosProfileListResp, err error) {
	// 构建过滤条件
	var filters = make([][]string, 0)
	if len(req.Name) > 0 {
		filters = append(filters, []string{"name", "like", req.Name})
	}
	if len(req.Company) > 0 {
		filters = append(filters, []string{"company", "like", req.Company})
	}

	// 查询Pos Profile列表
	list, err := service.Document().List(ctx, &dto.ErpReq{
		DocType: "Pos Profile",
	}, &dto.RequestParams{
		Fields:  []string{"name", "company", "warehouse", "branch"},
		Filters: filters,
	})
	if err != nil {
		g.Log().Error(ctx, "查询Pos Profile失败", err)
		return nil, gerror.Wrapf(err, "查询Pos Profile失败")
	}
	if j, err := gjson.DecodeToJson(list.Bytes()); err == nil {
		// 遍历j.Get("data") 返回的数组字段，设置到 DataList 中
		dataList := make([]*selling.PosProfile, 0)
		dataArray := j.GetJsons("data")
		for _, item := range dataArray {
			dataInfo := &selling.PosProfile{
				Name:      item.Get("name").String(),
				Company:   item.Get("company").String(),
				Branch:    item.Get("branch").String(),
				Warehouse: item.Get("warehouse").String(),
			}
			dataList = append(dataList, dataInfo)
		}
		res = &selling.PosProfileListResp{
			ProfileList: dataList,
		}

	}
	return
}
