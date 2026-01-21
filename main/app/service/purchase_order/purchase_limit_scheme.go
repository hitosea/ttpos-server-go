package purchase_order

import (
	"fmt"
	"strings"
	"time"

	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IPurchaseLimitSchemeSrv 限购方案服务接口
type IPurchaseLimitSchemeSrv interface {
	// Create 创建限购方案
	Create(ctx context.Context, req req.PurchaseLimitSchemeCreateReq) (uint64, error)

	// Update 更新限购方案
	Update(ctx context.Context, req req.PurchaseLimitSchemeUpdateReq) error

	// GetByUuid 根据UUID获取限购方案详情
	GetByUuid(ctx context.Context, uuid uint64) (*resp.PurchaseLimitSchemeResp, error)

	// GetList 获取限购方案列表
	GetList(ctx context.Context, req req.PurchaseLimitSchemeListReq) (*resp.PurchaseLimitSchemeListResp, error)

	// Delete 删除限购方案
	Delete(ctx context.Context, uuid uint64) error
}

// purchaseLimitSchemeSrv 限购方案服务实现
type purchaseLimitSchemeSrv struct {
	dbm *database.DBManager
}

// NewPurchaseLimitSchemeSrv 创建限购方案服务
func NewPurchaseLimitSchemeSrv(dbm *database.DBManager) IPurchaseLimitSchemeSrv {
	return &purchaseLimitSchemeSrv{
		dbm: dbm,
	}
}

// Create 创建限购方案
func (s *purchaseLimitSchemeSrv) Create(ctx context.Context, req req.PurchaseLimitSchemeCreateReq) (uint64, error) {
	db := ctx.GetDB()

	// // 1. 校验周期和物品配置
	// if len(req.Weekdays) == 0 {
	// 	return 0, errors.New(i18n.Translate(ctx.GetLanguage(), "至少需选择一个星期"))
	// }
	// if len(req.Items) == 0 {
	// 	return 0, errors.New(i18n.Translate(ctx.GetLanguage(), "至少需选择一个物品"))
	// }

	// // 2. 如果不是应用到全部门店，则必须选择门店
	// if req.ApplyToAllShops == 0 && len(req.Shops) == 0 {
	// 	return 0, errors.New(i18n.Translate(ctx.GetLanguage(), "请选择适用的门店"))
	// }

	// 3. 验证物品是否存在并获取 Code
	materialRepo := repository.NewMaterialRepo(db)

	itemCodes := make(map[uint64]string) // materialUuid -> materialCode

	for _, item := range req.Items {
		// 验证物品是否存在并获取 Code
		if _, exists := itemCodes[item.MaterialUuid]; !exists {
			material, err := materialRepo.GetMaterialByUuid(item.MaterialUuid)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return 0, errors.New(i18n.Translate(ctx.GetLanguage(), "物品不存在"))
				}
				return 0, errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "查询物品失败"))
			}
			itemCodes[item.MaterialUuid] = material.Code
		}
	}

	// 4. 开启事务
	var schemeUuid uint64
	err := db.Transaction(func(tx *gorm.DB) error {
		currentTime := time.Now().Unix()

		// 4.1 创建限购方案主记录
		scheme := &model.PurchaseLimitScheme{
			Name:            req.LocaleName.ToJson(),
			Status:          req.Status,
			ApplyToAllShops: req.ApplyToAllShops,
			DailyLimit:      req.DailyLimit,
			Weekdays:        s.weekdaysToString(req.Weekdays),
		}
		scheme.CreateTime = currentTime
		scheme.UpdateTime = currentTime

		if err := repository.NewPurchaseLimitSchemeRepo(tx).Create(scheme); err != nil {
			logger.Logger.Error("创建限购方案失败", zap.Error(err))
			return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "创建限购方案失败"))
		}
		schemeUuid = scheme.Uuid

		// 4.2 批量插入物品配置
		items := make([]*model.PurchaseLimitSchemeItem, 0, len(req.Items))
		for _, item := range req.Items {
			itemModel := &model.PurchaseLimitSchemeItem{
				SchemeUuid:   scheme.Uuid,
				MaterialCode: itemCodes[item.MaterialUuid],
				QuotaLimit:   item.QuotaLimit,
			}
			itemModel.CreateTime = currentTime
			itemModel.UpdateTime = currentTime
			items = append(items, itemModel)
		}
		if err := repository.NewPurchaseLimitSchemeItemRepo(tx).BatchCreate(items); err != nil {
			logger.Logger.Error("创建物品配置失败", zap.Error(err))
			return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "创建物品配置失败"))
		}

		// 4.3 插入门店配置（如果不是全部门店）
		if req.ApplyToAllShops == 0 && len(req.Shops) > 0 {
			shops := make([]*model.PurchaseLimitSchemeShop, 0, len(req.Shops))
			for _, shopUuid := range req.Shops {
				shopModel := &model.PurchaseLimitSchemeShop{
					SchemeUuid:  scheme.Uuid,
					CompanyUuid: shopUuid,
				}
				shopModel.CreateTime = currentTime
				shopModel.UpdateTime = currentTime
				shops = append(shops, shopModel)
			}
			if err := repository.NewPurchaseLimitSchemeShopRepo(tx).BatchCreate(shops); err != nil {
				logger.Logger.Error("创建门店配置失败", zap.Error(err))
				return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "创建门店配置失败"))
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return schemeUuid, nil
}

// Update 更新限购方案
func (s *purchaseLimitSchemeSrv) Update(ctx context.Context, req req.PurchaseLimitSchemeUpdateReq) error {
	db := ctx.GetDB()

	// 1. 检查方案是否存在
	schemeRepo := repository.NewPurchaseLimitSchemeRepo(db)
	scheme, err := schemeRepo.GetByUuid(req.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(i18n.Translate(ctx.GetLanguage(), "限购方案不存在"))
		}
		return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "查询限购方案失败"))
	}

	// 2. 校验周期和物品配置
	// if len(req.Weekdays) == 0 {
	// 	return errors.New(i18n.Translate(ctx.GetLanguage(), "至少需选择一个星期"))
	// }
	// if len(req.Items) == 0 {
	// 	return errors.New(i18n.Translate(ctx.GetLanguage(), "至少需选择一个物品"))
	// }

	// // 3. 如果不是应用到全部门店，则必须选择门店
	// if req.ApplyToAllShops == 0 && len(req.Shops) == 0 {
	// 	return errors.New(i18n.Translate(ctx.GetLanguage(), "请选择适用的门店"))
	// }

	// 4. 验证物品是否存在并获取 Code
	materialRepo := repository.NewMaterialRepo(db)

	itemCodes := make(map[uint64]string) // materialUuid -> materialCode

	for _, item := range req.Items {
		// 验证物品是否存在并获取 Code
		if _, exists := itemCodes[item.MaterialUuid]; !exists {
			material, err := materialRepo.GetMaterialByUuid(item.MaterialUuid)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return errors.New(i18n.Translate(ctx.GetLanguage(), "物品不存在"))
				}
				return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "查询物品失败"))
			}
			itemCodes[item.MaterialUuid] = material.Code
		}
	}

	// 5. 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		currentTime := time.Now().Unix()

		// 5.1 更新限购方案主记录
		scheme.Name = req.LocaleName.ToJson()
		scheme.Status = req.Status
		scheme.ApplyToAllShops = req.ApplyToAllShops
		scheme.DailyLimit = req.DailyLimit
		scheme.Weekdays = s.weekdaysToString(req.Weekdays)
		scheme.UpdateTime = currentTime

		if err := repository.NewPurchaseLimitSchemeRepo(tx).Update(scheme); err != nil {
			logger.Logger.Error("更新限购方案失败", zap.Error(err))
			return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "更新限购方案失败"))
		}

		// 5.2 删除旧的物品配置
		itemRepo := repository.NewPurchaseLimitSchemeItemRepo(tx)
		if err := itemRepo.DeleteBySchemeUuid(scheme.Uuid); err != nil {
			logger.Logger.Error("删除旧物品配置失败", zap.Error(err))
			return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "删除旧物品配置失败"))
		}

		// 5.3 插入新的物品配置
		items := make([]*model.PurchaseLimitSchemeItem, 0, len(req.Items))
		for _, item := range req.Items {
			itemModel := &model.PurchaseLimitSchemeItem{
				SchemeUuid:   scheme.Uuid,
				MaterialCode: itemCodes[item.MaterialUuid],
				QuotaLimit:   item.QuotaLimit,
			}
			itemModel.CreateTime = currentTime
			itemModel.UpdateTime = currentTime
			items = append(items, itemModel)
		}
		if err := itemRepo.BatchCreate(items); err != nil {
			logger.Logger.Error("创建物品配置失败", zap.Error(err))
			return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "创建物品配置失败"))
		}

		// 4.4 删除旧的门店配置
		shopRepo := repository.NewPurchaseLimitSchemeShopRepo(tx)
		if err := shopRepo.DeleteBySchemeUuid(scheme.Uuid); err != nil {
			logger.Logger.Error("删除旧门店配置失败", zap.Error(err))
			return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "删除旧门店配置失败"))
		}

		// 4.5 插入新的门店配置（如果不是全部门店）
		if req.ApplyToAllShops == 0 && len(req.Shops) > 0 {
			shops := make([]*model.PurchaseLimitSchemeShop, 0, len(req.Shops))
			for _, shopUuid := range req.Shops {
				shopModel := &model.PurchaseLimitSchemeShop{
					SchemeUuid:  scheme.Uuid,
					CompanyUuid: shopUuid,
				}
				shopModel.CreateTime = currentTime
				shopModel.UpdateTime = currentTime
				shops = append(shops, shopModel)
			}
			if err := shopRepo.BatchCreate(shops); err != nil {
				logger.Logger.Error("创建门店配置失败", zap.Error(err))
				return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "创建门店配置失败"))
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// GetByUuid 根据UUID获取限购方案详情
func (s *purchaseLimitSchemeSrv) GetByUuid(ctx context.Context, uuid uint64) (*resp.PurchaseLimitSchemeResp, error) {
	db := ctx.GetDB()

	// 1. 查询限购方案主记录
	schemeRepo := repository.NewPurchaseLimitSchemeRepo(db)
	scheme, err := schemeRepo.GetByUuid(uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(i18n.Translate(ctx.GetLanguage(), "限购方案不存在"))
		}
		return nil, errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "查询限购方案失败"))
	}

	// 2. 查询物品配置
	itemRepo := repository.NewPurchaseLimitSchemeItemRepo(db)
	items, err := itemRepo.GetBySchemeUuid(scheme.Uuid)
	if err != nil {
		logger.Logger.Error("查询物品配置失败", zap.Error(err))
		return nil, errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "查询物品配置失败"))
	}

	// 3. 查询门店配置（如果不是应用到全部门店）
	var shopUuids = make([]uint64, 0)
	if scheme.ApplyToAllShops == 0 {
		shopRepo := repository.NewPurchaseLimitSchemeShopRepo(db)
		shops, err := shopRepo.GetBySchemeUuid(scheme.Uuid)
		if err != nil {
			logger.Logger.Error("查询门店配置失败", zap.Error(err))
			return nil, errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "查询门店配置失败"))
		}
		for _, shop := range shops {
			shopUuids = append(shopUuids, shop.CompanyUuid)
		}
	}

	// 4. 组装响应数据 - 需要将 Code 转换为 Uuid
	result := &resp.PurchaseLimitSchemeResp{
		Uuid:            scheme.Uuid,
		LocaleName:      *language.JsonToLocaleResponse(scheme.Name),
		Status:          scheme.Status,
		ApplyToAllShops: scheme.ApplyToAllShops,
		DailyLimit:      scheme.DailyLimit,
		Weekdays:        s.stringToWeekdays(scheme.Weekdays),
		Items:           make([]resp.PurchaseLimitSchemeItemResp, 0, len(items)),
		Shops:           shopUuids,
		CreateTime:      scheme.CreateTime,
		UpdateTime:      scheme.UpdateTime,
	}

	// 填充物品配置 - 需要通过 MaterialCode 查询对应的 MaterialUuid
	materialRepo := repository.NewMaterialRepo(db)

	for _, item := range items {
		// 通过 MaterialCode 获取 MaterialUuid
		material, err := materialRepo.GetMaterialByCode(item.MaterialCode)
		if err != nil {
			logger.Logger.Error("查询物品失败", zap.Error(err), zap.String("material_code", item.MaterialCode))
			continue
		}

		result.Items = append(result.Items, resp.PurchaseLimitSchemeItemResp{
			MaterialUuid: material.Uuid,
			QuotaLimit:   item.QuotaLimit,
		})
	}

	return result, nil
}

// GetList 获取限购方案列表
func (s *purchaseLimitSchemeSrv) GetList(ctx context.Context, req req.PurchaseLimitSchemeListReq) (*resp.PurchaseLimitSchemeListResp, error) {
	db := ctx.GetDB()
	lang := ctx.GetLanguage()

	// 1. 构建查询条件
	schemeRepo := repository.NewPurchaseLimitSchemeRepo(db)
	options := []repository.DBOption{
		schemeRepo.Paginate(req.PageNo, req.PageSize),
	}

	if req.Status != nil {
		options = append(options, schemeRepo.WhereStatus(*req.Status))
	}

	if req.Name != "" {
		options = append(options, schemeRepo.WhereNameLike(req.Name))
	}

	// 2. 查询限购方案列表
	schemes, total, err := schemeRepo.GetList(options...)
	if err != nil {
		logger.Logger.Error("查询限购方案列表失败", zap.Error(err))
		return nil, errors.WithMessage(err, i18n.Translate(lang, "查询限购方案列表失败"))
	}

	// 3. 组装响应数据
	list := make([]resp.PurchaseLimitSchemeSummaryResp, 0, len(schemes))
	for _, scheme := range schemes {
		// 3.1 查询物品配置数量
		itemRepo := repository.NewPurchaseLimitSchemeItemRepo(db)
		items, _ := itemRepo.GetBySchemeUuid(scheme.Uuid)

		// 3.2 查询门店配置数量
		shopCount := 0
		if scheme.ApplyToAllShops == 0 {
			shopRepo := repository.NewPurchaseLimitSchemeShopRepo(db)
			shops, _ := shopRepo.GetBySchemeUuid(scheme.Uuid)
			shopCount = len(shops)
		}

		// 3.3 生成星期字符串
		weekdayStr := s.formatWeekdaysFromString(lang, scheme.Weekdays)

		list = append(list, resp.PurchaseLimitSchemeSummaryResp{
			Uuid:            scheme.Uuid,
			LocaleName:      *language.JsonToLocaleResponse(scheme.Name),
			Status:          scheme.Status,
			ApplyToAllShops: scheme.ApplyToAllShops,
			WeekdayStr:      weekdayStr,
			ShopCount:       shopCount,
			ItemCount:       len(items),
			DailyLimit:      scheme.DailyLimit,
			CreateTime:      scheme.CreateTime,
			UpdateTime:      scheme.UpdateTime,
		})
	}

	return &resp.PurchaseLimitSchemeListResp{
		List: list,
		Meta: resp.PageMeta{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// Delete 删除限购方案
func (s *purchaseLimitSchemeSrv) Delete(ctx context.Context, uuid uint64) error {
	db := ctx.GetDB()

	// 1. 检查方案是否存在
	schemeRepo := repository.NewPurchaseLimitSchemeRepo(db)
	scheme, err := schemeRepo.GetByUuid(uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(i18n.Translate(ctx.GetLanguage(), "限购方案不存在"))
		}
		return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "查询限购方案失败"))
	}

	// 2. 开启事务删除主表和关联表
	err = db.Transaction(func(tx *gorm.DB) error {
		// 2.1 软删除限购方案主记录
		if err := repository.NewPurchaseLimitSchemeRepo(tx).Delete(uuid); err != nil {
			logger.Logger.Error("删除限购方案失败", zap.Error(err))
			return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "删除限购方案失败"))
		}

		// 2.2 软删除物品配置
		if err := repository.NewPurchaseLimitSchemeItemRepo(tx).DeleteBySchemeUuid(scheme.Uuid); err != nil {
			logger.Logger.Error("删除物品配置失败", zap.Error(err))
			return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "删除物品配置失败"))
		}

		// 2.3 软删除门店配置
		if err := repository.NewPurchaseLimitSchemeShopRepo(tx).DeleteBySchemeUuid(scheme.Uuid); err != nil {
			logger.Logger.Error("删除门店配置失败", zap.Error(err))
			return errors.WithMessage(err, i18n.Translate(ctx.GetLanguage(), "删除门店配置失败"))
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// weekdaysToString 将星期数组转换为逗号分隔的字符串
func (s *purchaseLimitSchemeSrv) weekdaysToString(weekdays []int8) string {
	if len(weekdays) == 0 {
		return ""
	}

	strs := make([]string, 0, len(weekdays))
	for _, w := range weekdays {
		strs = append(strs, fmt.Sprintf("%d", w))
	}
	return strings.Join(strs, ",")
}

// stringToWeekdays 将逗号分隔的字符串转换为星期数组
func (s *purchaseLimitSchemeSrv) stringToWeekdays(weekdaysStr string) []int8 {
	if weekdaysStr == "" {
		return []int8{}
	}

	parts := strings.Split(weekdaysStr, ",")
	weekdays := make([]int8, 0, len(parts))
	for _, part := range parts {
		var weekday int8
		if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &weekday); err == nil {
			if weekday >= 1 && weekday <= 7 {
				weekdays = append(weekdays, weekday)
			}
		}
	}
	return weekdays
}

// formatWeekdaysFromString 格式化星期字符串为可读格式
func (s *purchaseLimitSchemeSrv) formatWeekdaysFromString(language string, weekdaysStr string) string {
	weekdays := s.stringToWeekdays(weekdaysStr)
	if len(weekdays) == 0 {
		return ""
	}

	weekdayMap := map[int8]string{
		1: i18n.Translate(language, "周一"),
		2: i18n.Translate(language, "周二"),
		3: i18n.Translate(language, "周三"),
		4: i18n.Translate(language, "周四"),
		5: i18n.Translate(language, "周五"),
		6: i18n.Translate(language, "周六"),
		7: i18n.Translate(language, "周日"),
	}

	weekdayStrs := make([]string, 0, len(weekdays))
	for _, w := range weekdays {
		if str, ok := weekdayMap[w]; ok {
			weekdayStrs = append(weekdayStrs, str)
		}
	}

	return strings.Join(weekdayStrs, "、")
}
