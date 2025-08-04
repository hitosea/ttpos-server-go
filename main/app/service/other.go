package service

import (
	"log"
	"time"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/captcha"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type IOtherSrv interface {
	Generate() (*resp.Captcha, error)
	GetReturnFoodReasonList(ctx context.Context) (*resp.ReturnFoodReasonResp, error)
	GetGiftOrFreeReasonList(ctx context.Context) (*resp.GiftOrFreeOrderReasonResp, error)
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
func (s *otherSrv) GetReturnFoodReasonList(ctx context.Context) (*resp.ReturnFoodReasonResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := base.NewReturnFoodReasonRepo(db)
	list, err := productRepo.GetReturnFoodReasonList()
	if err != nil {
		return nil, errors.WithMessage(err, "获取退菜原因列表失败")
	}

	result := make([]resp.ReturnFoodReason, 0, len(list))
	for _, item := range list {
		result = append(result, resp.ReturnFoodReason{
			Uuid:       item.Uuid,
			LocaleName: item.MultiLanguageName.GetNames(),
		})
	}

	return &resp.ReturnFoodReasonResp{
		List: result,
	}, nil
}

// GetGiftOrFreeReasonList 获取免单原因列表
func (s *otherSrv) GetGiftOrFreeReasonList(ctx context.Context) (*resp.GiftOrFreeOrderReasonResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	productRepo := base.NewGiftOrFreeOrderReasonRepo(db)
	list, err := productRepo.GetGiftOrFreeOrderReasonList()
	if err != nil {
		return nil, errors.WithMessage(err, "获取免单原因列表失败")
	}

	result := make([]resp.GiftOrFreeOrderReason, 0, len(list))
	for _, item := range list {
		result = append(result, resp.GiftOrFreeOrderReason{
			Uuid:       item.Uuid,
			LocaleName: item.MultiLanguageName.GetNames(),
		})
	}

	return &resp.GiftOrFreeOrderReasonResp{
		List: result,
	}, nil
}
