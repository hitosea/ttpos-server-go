package service

import (
	"errors"
	"log"
	"time"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/captcha"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type IOtherSrv interface {
	Generate() (*resp.Captcha, error)
	GetReturnFoodReasonList(ctx context.Context) (*resp.ReturnFoodReasonResps, error) // CheckAnswer 检查答案是否正确 string, answer string) bool
}

func NewOtherSrv(dbm *database.DBManager, cache cache.Cache) IOtherSrv {
	return NewOtherSrvImpl(dbm, cache)
}

type otherSrv struct {
	captcha     *captcha.Captcha
	cachePrefix string
	dbm         *database.DBManager
}

// Generate implements IOtherSrv.
func (s *otherSrv) Generate() (*resp.Captcha, error) {
	panic("unimplemented")
}

func NewOtherSrvImpl(dbm *database.DBManager, cache cache.Cache) IOtherSrv {
	srv := &otherSrv{
		cachePrefix: "captcha:",
		dbm:         dbm,
	}
	captchaTool, err := captcha.New(cache, srv.cachePrefix, 5*time.Minute)
	if err != nil {
		log.Fatalln(err)
	}
	srv.captcha = captchaTool
	return srv
}

// GetReturnFoodReasonList 获取退菜原因列表
func (s *otherSrv) GetReturnFoodReasonList(ctx context.Context) (*resp.ReturnFoodReasonResps, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := base.NewReturnFoodReasonRepo(db)
	list, err := productRepo.GetReturnFoodReasonList()
	if err != nil {
		return nil, errors.New("获取退菜原因列表失败")
	}

	result := make([]resp.ReturnFoodReasonResp, 0, len(list))
	for _, item := range list {
		result = append(result, resp.ReturnFoodReasonResp{
			Uuid:       item.Uuid,
			LocaleName: item.MultiLanguageName.GetNames(),
		})
	}

	return &resp.ReturnFoodReasonResps{
		List: result,
	}, nil
}
