package inventory

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/inventory/domain/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// MultiLanguageNameRepositoryImpl 多语言名称仓储实现
type MultiLanguageNameRepositoryImpl struct {
	dbm *database.DBManager
}

// NewMultiLanguageNameRepository 创建多语言名称仓储
func NewMultiLanguageNameRepository(dbm *database.DBManager) repository.IMultiLanguageNameRepository {
	return &MultiLanguageNameRepositoryImpl{
		dbm: dbm,
	}
}

// Save 保存多语言名称
func (r *MultiLanguageNameRepositoryImpl) Save(ctx context.Context, data *repository.MultiLanguageNameData) error {
	db := r.getDB(ctx)
	po := r.toPO(data)

	if data.Uuid == 0 {
		// 生成UUID
		uuid, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err, "生成UUID失败")
		}
		po.Uuid = uuid
		data.Uuid = uuid

		// 创建
		if err := db.Create(po).Error; err != nil {
			return errors.WithMessage(err, "创建多语言名称失败")
		}
	} else {
		// 更新
		if err := db.Model(&model.MultiLanguageName{}).Where("uuid = ?", po.Uuid).Updates(map[string]any{
			"zh_name":    po.ZhName,
			"th_name":    po.ThName,
			"en_name":    po.EnName,
			"zh_tw_name": po.ZhTwName,
			"ja_name":    po.JaName,
			"ko_name":    po.KoName,
			"my_name":    po.MyName,
			"tr_name":    po.TrName,
			"sv_name":    po.SvName,
		}).Error; err != nil {
			return errors.WithMessage(err, "更新多语言名称失败")
		}
	}

	return nil
}

// FindByUuid 根据UUID查找多语言名称
func (r *MultiLanguageNameRepositoryImpl) FindByUuid(ctx context.Context, uuid uint64) (*repository.MultiLanguageNameData, error) {
	db := r.getDB(ctx)

	var po model.MultiLanguageName
	if err := db.Where("uuid = ?", uuid).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "查询多语言名称失败")
	}

	return r.toDomain(&po), nil
}

// Remove 删除多语言名称
func (r *MultiLanguageNameRepositoryImpl) Remove(ctx context.Context, uuid uint64) error {
	db := r.getDB(ctx)

	if err := db.Where("uuid = ?", uuid).Delete(&model.MultiLanguageName{}).Error; err != nil {
		return errors.WithMessage(err, "删除多语言名称失败")
	}

	return nil
}

// toPO 转换为持久化对象
func (r *MultiLanguageNameRepositoryImpl) toPO(data *repository.MultiLanguageNameData) *model.MultiLanguageName {
	return &model.MultiLanguageName{
		BaseModel: model.BaseModel{
			Uuid: data.Uuid,
		},
		ZhName:   data.ZH,
		ThName:   data.TH,
		EnName:   data.EN,
		ZhTwName: data.ZHTW,
		JaName:   data.JA,
		KoName:   data.KO,
		MyName:   data.MY,
		TrName:   data.TR,
		SvName:   data.SV,
	}
}

// toDomain 转换为领域数据
func (r *MultiLanguageNameRepositoryImpl) toDomain(po *model.MultiLanguageName) *repository.MultiLanguageNameData {
	return &repository.MultiLanguageNameData{
		Uuid: po.Uuid,
		ZH:   po.ZhName,
		TH:   po.ThName,
		EN:   po.EnName,
		ZHTW: po.ZhTwName,
		JA:   po.JaName,
		KO:   po.KoName,
		MY:   po.MyName,
		TR:   po.TrName,
		SV:   po.SvName,
	}
}

// getDB 获取数据库连接
func (r *MultiLanguageNameRepositoryImpl) getDB(ctx context.Context) *gorm.DB {
	return ctx.GetDB()
}
