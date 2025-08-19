package service

import (
	"log"
	"time"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/captcha"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type IOtherSrv interface {
	Generate() (*resp.Captcha, error)
	GetReturnFoodReasonList(ctx context.Context) (*resp.ReturnFoodReasonResp, error)
	GetGiftOrFreeReasonList(ctx context.Context) (*resp.GiftOrFreeOrderReasonResp, error)
	AddFreeOrGiftReason(ctx context.Context, addFreeOrGiftReasonReq req.AddFreeOrGiftReasonReq) error
	EditFreeOrGiftReason(ctx context.Context, editFreeReason req.EditFreeOrGiftReasonReq) error
	DeleteFreeOrGiftReason(ctx context.Context, deleteFreeReason req.DeleteFreeOrGiftReasonReq) error
	AddReturnFoodReason(ctx context.Context, addReturnFoodReason req.AddReturnFoodReasonReq) error
	EditReturnFoodReason(ctx context.Context, editReturnFoodReason req.EditReturnFoodReasonReq) error
	DeleteReturnFoodReason(ctx context.Context, deleteReturnFoodReason req.DeleteReturnFoodReasonReq) error
}

func NewOtherSrv(dbm *database.DBManager, cache cache.Cache, settingSrv setting.ISrv) IOtherSrv {
	return NewOtherSrvImpl(dbm, cache, settingSrv)
}

type otherSrv struct {
	captcha     *captcha.Captcha
	cachePrefix string
	dbm         *database.DBManager
	settingSrv  setting.ISrv
}

// Generate implements IOtherSrv.
func (s *otherSrv) Generate() (*resp.Captcha, error) {
	panic("unimplemented")
}

func NewOtherSrvImpl(dbm *database.DBManager, cache cache.Cache, settingSrv setting.ISrv) IOtherSrv {
	srv := &otherSrv{
		cachePrefix: "captcha:",
		dbm:         dbm,
		settingSrv:  settingSrv,
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

func (s *otherSrv) AddFreeOrGiftReason(ctx context.Context, addFreeOrGiftReasonReq req.AddFreeOrGiftReasonReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !addFreeOrGiftReasonReq.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("名称不能为空")
	}
	companySetting := ctx.GetCompanySetting()
	defaultLang := companySetting.GetDefaultLanguage()
	_, err := base.NewGiftOrFreeOrderReasonRepo(s.dbm.GetDB(ctx.GetCompanyUuid())).CreateGiftOrFreeOrderReason(model.FreeReason{
		Name: addFreeOrGiftReasonReq.LocaleName.GetLocale(defaultLang),
		MultiLanguageName: model.MultiLanguageName{
			ZhName:   addFreeOrGiftReasonReq.LocaleName.ZH,
			ThName:   addFreeOrGiftReasonReq.LocaleName.TH,
			EnName:   addFreeOrGiftReasonReq.LocaleName.EN,
			ZhTwName: addFreeOrGiftReasonReq.LocaleName.ZHTW,
			JaName:   addFreeOrGiftReasonReq.LocaleName.JA,
			KoName:   addFreeOrGiftReasonReq.LocaleName.KO,
			MyName:   addFreeOrGiftReasonReq.LocaleName.MY,
			TrName:   addFreeOrGiftReasonReq.LocaleName.TR,
			SvName:   addFreeOrGiftReasonReq.LocaleName.SV,
		},
	})
	if err != nil {
		return errors.WithMessage(err, "新增免单原因失败")
	}
	return nil
}

func (s *otherSrv) EditFreeOrGiftReason(ctx context.Context, editFreeOrGiftReasonReq req.EditFreeOrGiftReasonReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !editFreeOrGiftReasonReq.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("名称不能为空")
	}
	db := s.dbm.GetDB(ctx.GetCompanyUuid())

	companySetting := ctx.GetCompanySetting()
	defaultLang := companySetting.GetDefaultLanguage()

	productRepo := base.NewGiftOrFreeOrderReasonRepo(db)

	reason, err := productRepo.GetFreeOrderReasonByUuid(editFreeOrGiftReasonReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "免单/赠菜原因不存在")
	}

	reason.Name = editFreeOrGiftReasonReq.LocaleName.GetLocale(defaultLang)
	reason.MultiLanguageName.ZhName = editFreeOrGiftReasonReq.LocaleName.ZH
	reason.MultiLanguageName.ThName = editFreeOrGiftReasonReq.LocaleName.TH
	reason.MultiLanguageName.EnName = editFreeOrGiftReasonReq.LocaleName.EN
	reason.MultiLanguageName.ZhTwName = editFreeOrGiftReasonReq.LocaleName.ZHTW
	reason.MultiLanguageName.JaName = editFreeOrGiftReasonReq.LocaleName.JA
	reason.MultiLanguageName.KoName = editFreeOrGiftReasonReq.LocaleName.KO
	reason.MultiLanguageName.MyName = editFreeOrGiftReasonReq.LocaleName.MY
	reason.MultiLanguageName.TrName = editFreeOrGiftReasonReq.LocaleName.TR
	reason.MultiLanguageName.SvName = editFreeOrGiftReasonReq.LocaleName.SV

	err = base.NewGiftOrFreeOrderReasonRepo(db).UpdateGiftOrFreeOrderReason(editFreeOrGiftReasonReq.Uuid, *reason)
	if err != nil {
		return errors.WithMessage(err, "保存免单/赠菜原因失败")
	}
	return nil
}

func (s *otherSrv) AddReturnFoodReason(ctx context.Context, addReturnFoodReason req.AddReturnFoodReasonReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !addReturnFoodReason.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("名称不能为空")
	}

	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	companySetting := ctx.GetCompanySetting()
	defaultLang := companySetting.GetDefaultLanguage()

	_, err := base.NewReturnFoodReasonRepo(db).CreateReturnFoodReason(model.ReturnFoodReason{
		Name: addReturnFoodReason.LocaleName.GetLocale(defaultLang),
		MultiLanguageName: model.MultiLanguageName{
			ZhName:   addReturnFoodReason.LocaleName.ZH,
			ThName:   addReturnFoodReason.LocaleName.TH,
			EnName:   addReturnFoodReason.LocaleName.EN,
			ZhTwName: addReturnFoodReason.LocaleName.ZHTW,
			JaName:   addReturnFoodReason.LocaleName.JA,
			KoName:   addReturnFoodReason.LocaleName.KO,
			MyName:   addReturnFoodReason.LocaleName.MY,
			TrName:   addReturnFoodReason.LocaleName.TR,
			SvName:   addReturnFoodReason.LocaleName.SV,
		},
	})
	if err != nil {
		return errors.WithMessage(err, "新增退菜原因失败")
	}
	return nil
}

func (s *otherSrv) EditReturnFoodReason(ctx context.Context, editReturnFoodReason req.EditReturnFoodReasonReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !editReturnFoodReason.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("名称不能为空")
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	companySetting := ctx.GetCompanySetting()
	defaultLang := companySetting.GetDefaultLanguage()

	productRepo := base.NewReturnFoodReasonRepo(db)
	reason, err := productRepo.GetReturnFoodReasonByUuid(editReturnFoodReason.Uuid)
	if err != nil {
		return errors.WithMessage(err, "退菜原因不存在")
	}

	reason.Name = editReturnFoodReason.LocaleName.GetLocale(defaultLang)
	reason.MultiLanguageName.ZhName = editReturnFoodReason.LocaleName.ZH
	reason.MultiLanguageName.ThName = editReturnFoodReason.LocaleName.TH
	reason.MultiLanguageName.EnName = editReturnFoodReason.LocaleName.EN
	reason.MultiLanguageName.ZhTwName = editReturnFoodReason.LocaleName.ZHTW
	reason.MultiLanguageName.JaName = editReturnFoodReason.LocaleName.JA
	reason.MultiLanguageName.KoName = editReturnFoodReason.LocaleName.KO
	reason.MultiLanguageName.MyName = editReturnFoodReason.LocaleName.MY
	reason.MultiLanguageName.TrName = editReturnFoodReason.LocaleName.TR
	reason.MultiLanguageName.SvName = editReturnFoodReason.LocaleName.SV

	err = base.NewReturnFoodReasonRepo(db).UpdateReturnFoodReason(editReturnFoodReason.Uuid, *reason)
	if err != nil {
		return errors.WithMessage(err, "保存退菜原因失败")
	}
	return nil
}

func (s *otherSrv) DeleteFreeOrGiftReason(ctx context.Context, deleteFreeReason req.DeleteFreeOrGiftReasonReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	err := base.NewGiftOrFreeOrderReasonRepo(db).DeleteGiftOrFreeOrderReason(deleteFreeReason.Uuid)
	if err != nil {
		return errors.WithMessage(err, "删除免单原因失败")
	}
	return nil
}

func (s *otherSrv) DeleteReturnFoodReason(ctx context.Context, deleteReturnFoodReason req.DeleteReturnFoodReasonReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	err := base.NewReturnFoodReasonRepo(db).DeleteReturnFoodReason(deleteReturnFoodReason.Uuid)
	if err != nil {
		return errors.WithMessage(err, "删除退菜原因失败")
	}
	return nil
}
