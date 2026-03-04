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

	"gorm.io/gorm"
)

type IOtherSrv interface {
	GetReturnFoodReasonList(ctx context.Context) (*resp.ReturnFoodReasonResp, error)
	GetGiftOrFreeReasonList(ctx context.Context) (*resp.GiftOrFreeOrderReasonResp, error)
	GetOrderRemarkList(ctx context.Context) (*resp.OrderRemarkResp, error)
	AddOrderRemark(ctx context.Context, addOrderRemark req.AddOrderRemarkReq) error
	EditOrderRemark(ctx context.Context, editOrderRemark req.EditOrderRemarkReq) error
	DeleteOrderRemark(ctx context.Context, deleteOrderRemark req.DeleteOrderRemarkReq) error
	GetOrderItemRemarkList(ctx context.Context) (*resp.OrderItemRemarkResp, error)
	AddOrderItemRemark(ctx context.Context, addOrderItemRemark req.AddOrderItemRemarkReq) error
	EditOrderItemRemark(ctx context.Context, editOrderItemRemark req.EditOrderItemRemarkReq) error
	DeleteOrderItemRemark(ctx context.Context, deleteOrderItemRemark req.DeleteOrderItemRemarkReq) error
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

// GetOrderRemarkList 获取整单备注列表
func (s *otherSrv) GetOrderRemarkList(ctx context.Context) (*resp.OrderRemarkResp, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	productRepo := base.NewOrderRemarkRepo(db)
	list, err := productRepo.GetOrderRemarkList()
	if err != nil {
		return nil, errors.WithMessage(err, "获取整单备注列表失败")
	}

	result := make([]resp.OrderRemark, 0, len(list))
	for _, remark := range list {
		result = append(result, resp.OrderRemark{
			Uuid:       remark.Uuid,
			LocaleName: remark.MultiLanguageName.GetNames(),
		})
	}

	return &resp.OrderRemarkResp{
		List: result,
	}, nil
}

// AddOrderRemark 新增整单备注
func (s *otherSrv) AddOrderRemark(ctx context.Context, addOrderRemark req.AddOrderRemarkReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !addOrderRemark.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("多语言名称不完整")
	}

	if !addOrderRemark.LocaleName.CheckLen(100) {
		return errors.New("名称长度不能超过100个字符")
	}

	// 限制整单备注数量,不能超过100个
	count, err := base.NewOrderRemarkRepo(s.dbm.GetDB(ctx.GetCompanyUuid())).CountOrderRemark()
	if err != nil {
		return errors.WithMessage(err, "获取整单备注数量失败")
	}
	if count >= 100 {
		return errors.New("整单备注数量不能超过100个")
	}

	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	if err := db.Transaction(func(tx *gorm.DB) error {
		repo := base.NewOrderRemarkRepo(tx)
		_, err := repo.CreateOrderRemark(model.OrderRemark{
			Name: addOrderRemark.LocaleName.GetLocale("zh"),
			MultiLanguageName: model.MultiLanguageName{
				ZhName:   addOrderRemark.LocaleName.ZH,
				ThName:   addOrderRemark.LocaleName.TH,
				EnName:   addOrderRemark.LocaleName.EN,
				ZhTwName: addOrderRemark.LocaleName.ZHTW,
				JaName:   addOrderRemark.LocaleName.JA,
				KoName:   addOrderRemark.LocaleName.KO,
				MyName:   addOrderRemark.LocaleName.MY,
				TrName:   addOrderRemark.LocaleName.TR,
				SvName:   addOrderRemark.LocaleName.SV,
			},
		})
		return err
	}); err != nil {
		return errors.WithMessage(err, "保存整单备注失败")
	}
	return nil
}

// EditOrderRemark 编辑整单备注
func (s *otherSrv) EditOrderRemark(ctx context.Context, editOrderRemark req.EditOrderRemarkReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !editOrderRemark.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("多语言名称不完整")
	}

	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	err := db.Transaction(func(tx *gorm.DB) error {
		repo := base.NewOrderRemarkRepo(tx)
		remark, err := repo.GetOrderRemarkByUuid(editOrderRemark.Uuid)
		if err != nil {
			return errors.WithMessage(err, "整单备注不存在")
		}

		remark.Name = editOrderRemark.LocaleName.GetLocale("zh")
		remark.MultiLanguageName.ZhName = editOrderRemark.LocaleName.ZH
		remark.MultiLanguageName.ThName = editOrderRemark.LocaleName.TH
		remark.MultiLanguageName.EnName = editOrderRemark.LocaleName.EN
		remark.MultiLanguageName.ZhTwName = editOrderRemark.LocaleName.ZHTW
		remark.MultiLanguageName.JaName = editOrderRemark.LocaleName.JA
		remark.MultiLanguageName.KoName = editOrderRemark.LocaleName.KO
		remark.MultiLanguageName.MyName = editOrderRemark.LocaleName.MY
		remark.MultiLanguageName.TrName = editOrderRemark.LocaleName.TR
		remark.MultiLanguageName.SvName = editOrderRemark.LocaleName.SV

		return repo.UpdateOrderRemark(editOrderRemark.Uuid, *remark)
	})
	if err != nil {
		return errors.WithMessage(err, "更新整单备注失败")
	}
	return nil
}

// DeleteOrderRemark 删除整单备注
func (s *otherSrv) DeleteOrderRemark(ctx context.Context, deleteOrderRemark req.DeleteOrderRemarkReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	err := db.Transaction(func(tx *gorm.DB) error {
		repo := base.NewOrderRemarkRepo(tx)
		return repo.DeleteOrderRemark(deleteOrderRemark.Uuid)
	})
	if err != nil {
		return errors.WithMessage(err, "删除整单备注失败")
	}
	return nil
}

// GetOrderItemRemarkList 获取单品备注列表
func (s *otherSrv) GetOrderItemRemarkList(ctx context.Context) (*resp.OrderItemRemarkResp, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	repo := base.NewOrderItemRemarkRepo(db)
	list, err := repo.GetOrderItemRemarkList()
	if err != nil {
		return nil, errors.WithMessage(err, "获取单品备注列表失败")
	}

	result := make([]resp.OrderItemRemark, 0, len(list))
	for _, remark := range list {
		result = append(result, resp.OrderItemRemark{
			Uuid:       remark.Uuid,
			LocaleName: remark.MultiLanguageName.GetNames(),
		})
	}

	return &resp.OrderItemRemarkResp{
		List: result,
	}, nil
}

// AddOrderItemRemark 新增单品备注
func (s *otherSrv) AddOrderItemRemark(ctx context.Context, addOrderItemRemark req.AddOrderItemRemarkReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !addOrderItemRemark.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("多语言名称不完整")
	}

	if !addOrderItemRemark.LocaleName.CheckLen(100) {
		return errors.New("字数不能超过100个字")
	}

	// 限制单品备注数量,不能超过100个
	count, err := base.NewOrderItemRemarkRepo(s.dbm.GetDB(ctx.GetCompanyUuid())).CountOrderItemRemark()
	if err != nil {
		return errors.WithMessage(err, "获取单品备注数量失败")
	}
	if count >= 100 {
		return errors.New("单品备注数量不能超过100个")
	}

	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	if err := db.Transaction(func(tx *gorm.DB) error {
		repo := base.NewOrderItemRemarkRepo(tx)
		_, err := repo.CreateOrderItemRemark(model.OrderItemRemark{
			Name: addOrderItemRemark.LocaleName.GetLocale("zh"),
			MultiLanguageName: model.MultiLanguageName{
				ZhName:   addOrderItemRemark.LocaleName.ZH,
				ThName:   addOrderItemRemark.LocaleName.TH,
				EnName:   addOrderItemRemark.LocaleName.EN,
				ZhTwName: addOrderItemRemark.LocaleName.ZHTW,
				JaName:   addOrderItemRemark.LocaleName.JA,
				KoName:   addOrderItemRemark.LocaleName.KO,
				MyName:   addOrderItemRemark.LocaleName.MY,
				TrName:   addOrderItemRemark.LocaleName.TR,
				SvName:   addOrderItemRemark.LocaleName.SV,
			},
		})
		return err
	}); err != nil {
		return errors.WithMessage(err, "保存单品备注失败")
	}
	return nil
}

// EditOrderItemRemark 编辑单品备注
func (s *otherSrv) EditOrderItemRemark(ctx context.Context, editOrderItemRemark req.EditOrderItemRemarkReq) error {
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !editOrderItemRemark.LocaleName.CheckRequiredLocale(storeLanguages) {
		return errors.New("多语言名称不完整")
	}

	if !editOrderItemRemark.LocaleName.CheckLen(100) {
		return errors.New("字数不能超过100个字")
	}

	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	err := db.Transaction(func(tx *gorm.DB) error {
		repo := base.NewOrderItemRemarkRepo(tx)
		remark, err := repo.GetOrderItemRemarkByUuid(editOrderItemRemark.Uuid)
		if err != nil {
			return errors.WithMessage(err, "单品备注不存在")
		}

		remark.Name = editOrderItemRemark.LocaleName.GetLocale("zh")
		remark.MultiLanguageName.ZhName = editOrderItemRemark.LocaleName.ZH
		remark.MultiLanguageName.ThName = editOrderItemRemark.LocaleName.TH
		remark.MultiLanguageName.EnName = editOrderItemRemark.LocaleName.EN
		remark.MultiLanguageName.ZhTwName = editOrderItemRemark.LocaleName.ZHTW
		remark.MultiLanguageName.JaName = editOrderItemRemark.LocaleName.JA
		remark.MultiLanguageName.KoName = editOrderItemRemark.LocaleName.KO
		remark.MultiLanguageName.MyName = editOrderItemRemark.LocaleName.MY
		remark.MultiLanguageName.TrName = editOrderItemRemark.LocaleName.TR
		remark.MultiLanguageName.SvName = editOrderItemRemark.LocaleName.SV

		return repo.UpdateOrderItemRemark(editOrderItemRemark.Uuid, *remark)
	})
	if err != nil {
		return errors.WithMessage(err, "更新单品备注失败")
	}
	return nil
}

// DeleteOrderItemRemark 删除单品备注
func (s *otherSrv) DeleteOrderItemRemark(ctx context.Context, deleteOrderItemRemark req.DeleteOrderItemRemarkReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	err := db.Transaction(func(tx *gorm.DB) error {
		repo := base.NewOrderItemRemarkRepo(tx)
		return repo.DeleteOrderItemRemark(deleteOrderItemRemark.Uuid)
	})
	if err != nil {
		return errors.WithMessage(err, "删除单品备注失败")
	}
	return nil
}
