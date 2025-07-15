package skootar

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"
	v1 "takeout/api/callback/v1"
	"takeout/internal/consts"
	"takeout/internal/dao"
	"takeout/internal/model/do"
	"takeout/internal/model/entity"
)

func (s *sSkootar) JobStatusChange(ctx context.Context, req *v1.SkootarStatusReq) (res *v1.SkootarStatusRes, err error) {
	//先做入库，后面再说
	err = dao.Job.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err = dao.Job.Ctx(ctx).Data(&do.CallbackMsg{
			Uuid:         guid.S(),
			TakeoutRefNo: req.JobID,
			Content:      gjson.MustEncodeString(req),
		}).Insert(); err != nil {
			return err
		}

		//状态变更
		var job *entity.Job
		jobModel := dao.Job.Ctx(ctx).Where(do.Job{
			TakeoutRefNo: req.JobID,
			ProviderName: consts.ProviderSkootar,
		})
		//状态变更
		if _, err = jobModel.Data(do.Job{
			JobStatus: req.StatusAfter,
		}).Update(); err != nil {
			return err
		}
		jobModel.Scan(&job)

		if _, err = dao.JobStatusLog.Ctx(ctx).Data(&do.JobStatusLog{
			Uuid:         guid.S(),
			JobUuid:      job.Uuid,
			StatusBefore: gconv.String(req.StatusBefore),
			StatusAfter:  gconv.String(req.StatusAfter),
		}).Insert(); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, gerror.Wrap(err, "更新任务状态失败")
	}

	//回调ttpos
	return &v1.SkootarStatusRes{
		Code: "200",
	}, nil
}
