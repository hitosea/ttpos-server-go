package core

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var User = new(sUser)

type sUser struct {
}

func init() {
	service.RegisterUser(User)
}

// GetUserByUsername 根据用户名获取用户信息
// 参数：
//   - ctx: 上下文
//   - userEmail: 用户名
//
// 返回：
//   - *erp.User: 用户信息
//   - error: 错误信息
func (s *sUser) GetUserByUsername(ctx context.Context, userEmail string) (*erp.User, error) {
	// 参数验证
	if userEmail == "" {
		return nil, gerror.New("用户名不能为空")
	}

	// 调用 Document 服务获取用户信息
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: "User",
		Name:    userEmail,
	}, nil)
	if err != nil {
		g.Log().Error(ctx, "获取用户信息失败", err, g.Map{"userEmail": userEmail})
		return nil, gerror.Wrapf(err, "获取用户信息失败")
	}

	// 解析响应数据
	j := resp

	// 提取用户数据
	userData := j.GetJson("data")
	if userData == nil {
		g.Log().Error(ctx, "用户不存在", g.Map{"userEmail": userEmail})
		return nil, gerror.Newf("用户不存在: %s", userEmail)
	}

	// 将数据转换为 User 结构体
	var user = &erp.User{}
	if err := userData.Scan(&user); err != nil {
		g.Log().Error(ctx, "转换用户数据失败", err, g.Map{"userEmail": userEmail})
		return nil, gerror.Wrapf(err, "转换用户数据失败")
	}
	return user, nil
}

// MustGetUserTimeZone 获取用户时区
// 参数：
//   - ctx: 上下文
//   - userEmail: 用户邮箱
//
// 返回：
//   - string: 时区
func (s *sUser) MustGetUserTimeZone(ctx context.Context, userEmail string) string {
	user, err := s.GetUserByUsername(ctx, userEmail)
	if err != nil {
		//默认返回曼谷
		return "Asia/Bangkok"
	}
	return user.TimeZone
}
