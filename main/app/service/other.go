package service

import (
	"log"
	"time"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/captcha"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/duke-git/lancet/v2/slice"
	"gorm.io/gorm"
)

type IOtherSrv interface {
	Generate() (*resp.Captcha, error)
	GetReturnFoodReasonList(ctx context.Context) (*resp.ReturnFoodReasonResp, error)
	GetGiftOrFreeReasonList(ctx context.Context) (*resp.GiftOrFreeOrderReasonResp, error)
	EditFreeOrGiftReason(ctx context.Context, editFreeReason req.EditFreeOrGiftReasonReq) error
	EditReturnFoodReason(ctx context.Context, editReturnFoodReason req.EditReturnFoodReasonReq) error
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

func (s *otherSrv) EditFreeOrGiftReason(ctx context.Context, editFreeOrGiftReason req.EditFreeOrGiftReasonReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())

	companySetting := ctx.GetCompanySetting()
	defaultLang := companySetting.GetDefaultLanguage()

	productRepo := base.NewGiftOrFreeOrderReasonRepo(db)
	oldList, err := productRepo.GetGiftOrFreeOrderReasonList()
	if err != nil {
		return errors.WithMessage(err, "获取免单原因列表失败")
	}

	oldUuids := make([]uint64, 0, len(oldList))
	for _, oldItem := range oldList {
		oldUuids = append(oldUuids, oldItem.Uuid)
	}

	newUuids := make([]uint64, 0, len(editFreeOrGiftReason.List))
	for _, item := range editFreeOrGiftReason.List {
		if item.Uuid != 0 {
			newUuids = append(newUuids, item.Uuid)
		}
	}

	diffUuids := slice.Difference(newUuids, oldUuids)

	err = db.Transaction(func(tx *gorm.DB) error {
		if len(diffUuids) > 0 {
			err := base.NewGiftOrFreeOrderReasonRepo(tx).DeleteGiftOrFreeOrderReasons(diffUuids)
			if err != nil {
				return errors.WithMessage(err, "删除免单原因失败")
			}
		}
		for _, item := range editFreeOrGiftReason.List {
			name := item.LocaleName.GetLocale(defaultLang)
			if item.Uuid != 0 { // 新增
				// 保存多语言名称
				multiLanguageName := model.MultiLanguageName{
					ZhName:   name,
					ThName:   name,
					EnName:   name,
					ZhTwName: name,
					JaName:   name,
					KoName:   name,
					MyName:   name,
					TrName:   name,
					SvName:   name,
				}
				err := tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
				if err != nil {
					return errors.WithMessage(err, "保存多语言名称失败")
				}
				base.NewGiftOrFreeOrderReasonRepo(tx).UpdateGiftOrFreeOrderReason(item.Uuid, model.FreeReason{
					Name:                  name,
					MultiLanguageNameUuid: multiLanguageName.Uuid,
				})
			} else { // 编辑
				err := tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", item.Uuid).Updates(map[string]any{
					"zh_name":    item.LocaleName.ZH,
					"th_name":    item.LocaleName.TH,
					"en_name":    item.LocaleName.EN,
					"zh_tw_name": item.LocaleName.ZHTW,
					"ja_name":    item.LocaleName.JA,
					"ko_name":    item.LocaleName.KO,
					"my_name":    item.LocaleName.MY,
					"tr_name":    item.LocaleName.TR,
					"sv_name":    item.LocaleName.SV,
				}).Error
				if err != nil {
					return errors.WithMessage(err, "修改多语言名称失败")
				}
				base.NewGiftOrFreeOrderReasonRepo(tx).UpdateGiftOrFreeOrderReason(item.Uuid, model.FreeReason{
					Name: name,
				})
			}
		}

		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "保存免单原因失败")
	}
	return nil
}

func (s *otherSrv) EditReturnFoodReason(ctx context.Context, editReturnFoodReason req.EditReturnFoodReasonReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())

	companySetting := ctx.GetCompanySetting()
	defaultLang := companySetting.GetDefaultLanguage()

	productRepo := base.NewReturnFoodReasonRepo(db)
	oldList, err := productRepo.GetReturnFoodReasonList()
	if err != nil {
		return errors.WithMessage(err, "获取退菜原因列表失败")
	}

	oldUuids := make([]uint64, 0, len(oldList))
	for _, oldItem := range oldList {
		oldUuids = append(oldUuids, oldItem.Uuid)
	}

	newUuids := make([]uint64, 0, len(editReturnFoodReason.List))
	for _, item := range editReturnFoodReason.List {
		if item.Uuid != 0 {
			newUuids = append(newUuids, item.Uuid)
		}
	}

	diffUuids := slice.Difference(newUuids, oldUuids)

	err = db.Transaction(func(tx *gorm.DB) error {
		if len(diffUuids) > 0 {
			err := base.NewReturnFoodReasonRepo(tx).DeleteReturnFoodReasons(diffUuids)
			if err != nil {
				return errors.WithMessage(err, "删除退菜原因失败")
			}
		}
		for _, item := range editReturnFoodReason.List {
			name := item.LocaleName.GetLocale(defaultLang)
			if item.Uuid != 0 { // 新增
				// 保存多语言名称
				multiLanguageName := model.MultiLanguageName{
					ZhName:   name,
					ThName:   name,
					EnName:   name,
					ZhTwName: name,
					JaName:   name,
					KoName:   name,
					MyName:   name,
					TrName:   name,
					SvName:   name,
				}
				err := tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
				if err != nil {
					return errors.WithMessage(err, "保存多语言名称失败")
				}
				_, err = base.NewReturnFoodReasonRepo(tx).CreateReturnFoodReason(model.ReturnFoodReason{
					Name:                  name,
					MultiLanguageNameUuid: multiLanguageName.Uuid,
				})
				if err != nil {
					return errors.WithMessage(err, "保存退菜原因失败")
				}
			} else { // 编辑
				err := tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", item.Uuid).Updates(map[string]any{
					"zh_name":    item.LocaleName.ZH,
					"th_name":    item.LocaleName.TH,
					"en_name":    item.LocaleName.EN,
					"zh_tw_name": item.LocaleName.ZHTW,
					"ja_name":    item.LocaleName.JA,
					"ko_name":    item.LocaleName.KO,
					"my_name":    item.LocaleName.MY,
					"tr_name":    item.LocaleName.TR,
					"sv_name":    item.LocaleName.SV,
				}).Error
				if err != nil {
					return errors.WithMessage(err, "修改多语言名称失败")
				}
				err = base.NewReturnFoodReasonRepo(tx).UpdateReturnFoodReason(item.Uuid, model.ReturnFoodReason{
					Name: name,
				})
				if err != nil {
					return errors.WithMessage(err, "修改退菜原因失败")
				}
			}
		}

		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "保存退菜原因失败")
	}
	return nil
}
