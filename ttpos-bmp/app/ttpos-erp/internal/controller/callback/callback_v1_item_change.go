package callback

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/internal/pkg/queue"

	"github.com/gogf/gf/v2/frame/g"

	"ttpos-bmp/app/ttpos-erp/api/callback/v1"
)

func (c *ControllerV1) ItemChange(ctx context.Context, req *v1.ItemChangeReq) (res *v1.ItemChangeRes, err error) {
	g.Log().Infof(ctx, "获取erpnext doc 变更： %s", req)
	err = queue.Push(string(consts.TopicItemChange), req)
	if err != nil {
		return nil, err
	}
	return &v1.ItemChangeRes{
		Code: "0",
		Msg:  "success",
	}, nil
}
