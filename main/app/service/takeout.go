package service

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/takeout/application"
	"ttpos-server-go/app/modules/takeout/domain/menu/valueobject"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	domainService "ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/modules/takeout/types/request"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ITakeoutSrv interface {
	// 导入菜单到TTPOS
	ImportMenuToTTPOS(ctx context.Context) (*resp.GrabMenuImportResp, error)

	// PushMenuToPlatform 推送菜单到外卖平台
	PushMenuToPlatform(ctx context.Context, platform string) error

	// ReimportMenuToTTPOS 重新导入菜单到TTPOS（基于失败日志重试）
	ReimportMenuToTTPOS(ctx context.Context, logUuid uint64) (*resp.GrabMenuImportResp, error)
}

type takeoutSrv struct {
	dbm               *database.DBManager
	cache             cache.Cache
	takeoutAppSrv     application.ITakeoutAppService
	productSrv        IProductSrv
	translateSrv      ITranslateSrv
	settingSrv        setting.ISrv
	uploadFileSrv     IUploadFileSrv
	productTakeoutSrv IProductTakeoutSrv
}

func NewTakeoutSrv(
	dbm *database.DBManager,
	cache cache.Cache,
	productSrv IProductSrv,
	productTakeoutSrv IProductTakeoutSrv,
	translateSrv ITranslateSrv,
	settingSrv setting.ISrv,
) ITakeoutSrv {
	return &takeoutSrv{
		dbm:               dbm,
		cache:             cache,
		takeoutAppSrv:     application.NewTakeoutAppService(dbm),
		productSrv:        productSrv,
		translateSrv:      translateSrv,
		settingSrv:        settingSrv,
		uploadFileSrv:     NewUploadFileSrv(dbm),
		productTakeoutSrv: productTakeoutSrv,
	}
}

// PushMenuToPlatform 推送菜单到外卖平台
func (s *takeoutSrv) PushMenuToPlatform(ctx context.Context, platform string) error {
	platform = strings.ToLower(platform)

	// 初始化领域服务
	db := ctx.GetDB()
	progressService := domainService.NewImportProgressService(db)

	// 1. 检查推送状态
	canImport, inProgressLog, err := progressService.CheckImportStatus(ctx, platform)
	if err != nil {
		return errors.WithMessage(err, "检查推送状态失败")
	}
	if !canImport {
		errMsg := fmt.Sprintf("已有推送任务正在进行中 (UUID: %d)", inProgressLog.UUID)
		return errors.New(errMsg)
	}

	// 2. 开始推送，创建推送日志
	// importType: 1 = TTPOS推送到平台
	// importDirection: 根据平台名称设置
	// 首字母大写
	platformName := platform
	if len(platform) > 0 {
		platformName = strings.ToUpper(string(platform[0])) + strings.ToLower(platform[1:])
	}
	importDirection := fmt.Sprintf("TTPOS推送到%s", platformName)
	importLog, err := progressService.StartImport(ctx, platform, 1, importDirection)
	if err != nil {
		return errors.WithMessage(err, "开始推送失败")
	}

	// 3. 获取货币设置 (进度 0-20%)
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		logger.Logger.Error("获取货币设置失败",
			zap.String("platform", platform),
			zap.Error(err),
		)
		// 标记推送失败
		_ = progressService.CompleteImport(ctx, importLog.UUID, false, "获取货币设置失败: "+err.Error())
		return errors.WithMessage(err, "获取货币设置失败")
	}

	// 更新进度到 20%
	_ = progressService.UpdateProgress(ctx, importLog.UUID, 20, 100)

	// 4. 推送菜单到平台 (进度 20-100%)
	err = s.takeoutAppSrv.PushMenu(ctx, platform, currencySetting.Unit)
	if err != nil {
		logger.Logger.Error("推送菜单到平台失败",
			zap.String("platform", platform),
			zap.Error(err),
		)
		// 标记推送失败
		_ = progressService.CompleteImport(ctx, importLog.UUID, false, "推送菜单失败: "+err.Error())
		return errors.WithMessage(err, "推送菜单失败")
	}

	// 更新进度到 100%
	_ = progressService.UpdateProgress(ctx, importLog.UUID, 100, 100)

	// 5. 标记推送成功
	err = progressService.CompleteImport(ctx, importLog.UUID, true, "")
	if err != nil {
		logger.Logger.Warn("标记推送成功失败",
			zap.String("platform", platform),
			zap.Uint64("import_log_uuid", importLog.UUID),
			zap.Error(err),
		)
	}

	return nil
}

// ImportMenuToTTPOS 导入菜单到TTPOS
func (s *takeoutSrv) ImportMenuToTTPOS(ctx context.Context) (*resp.GrabMenuImportResp, error) {
	// TODO: 传产品说 "导入菜单到TTPOS的功能不需要了，我认为不合理 就先保留注释"

	// takeoutMenu, err := s.takeoutAppSrv.GetGrabMenu(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// // 1. 导入菜单
	// resp, err := s.ImportMenu(ctx, request.ImportMenuRequest{
	// 	Platform: "grab",
	// 	MenuData: takeoutMenu.Menu,
	// })
	// if err != nil {
	// 	return nil, err
	// }

	err := s.PushMenuToPlatform(ctx, "grab")
	if err != nil {
		return nil, errors.WithMessage(err, "推送菜单到Grab失败")
	}

	return nil, nil
}

// ReimportMenuToTTPOS 重新导入菜单到TTPOS（基于失败日志重试）
func (s *takeoutSrv) ReimportMenuToTTPOS(ctx context.Context, logUuid uint64) (*resp.GrabMenuImportResp, error) {
	db := ctx.GetDB()
	progressService := domainService.NewImportProgressService(db)

	// 1. 查询日志
	log, err := persistence.NewTakeoutImportLogRepository(db).FindByUUID(ctx, logUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询导入日志失败")
	}
	if log == nil {
		return nil, errors.New("导入日志不存在")
	}

	// 2. 检查是否可以重新导入
	// 判断条件：导入失败 && 是从平台导入到TTPOS && 平台是 grab
	if !log.IsFailed() {
		return nil, errors.New("只能重新导入失败的记录")
	}
	if log.ImportType != takeoutModel.ImportTypePlatformToTTPOS {
		return nil, errors.New("只能重新导入从平台到TTPOS的记录")
	}
	if log.Platform != "grab" {
		return nil, errors.New("当前仅支持重新导入 Grab 平台")
	}

	// 3. 检查是否有正在进行的导入
	canImport, inProgressLog, err := progressService.CheckImportStatus(ctx, log.Platform)
	if err != nil {
		return nil, errors.WithMessage(err, "检查导入状态失败")
	}
	if !canImport {
		errMsg := fmt.Sprintf("已有导入任务正在进行中 (UUID: %d)", inProgressLog.UUID)
		return nil, errors.New(errMsg)
	}

	// 4. 获取菜单数据
	takeoutMenu, err := s.takeoutAppSrv.GetGrabMenu(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "获取菜单数据失败")
	}

	// 5. 调用 ImportMenu，传入日志 UUID 进行重新导入
	return s.importMenuWithLog(ctx, request.ImportMenuRequest{
		Platform: log.Platform,
		MenuData: takeoutMenu.Menu,
	}, &logUuid)
}

// ImportMenu 导入菜单
func (s *takeoutSrv) ImportMenu(ctx context.Context, reqs request.ImportMenuRequest) (*resp.GrabMenuImportResp, error) {
	return s.importMenuWithLog(ctx, reqs, nil)
}

// importMenuWithLog 导入菜单（支持复用日志）
func (s *takeoutSrv) importMenuWithLog(ctx context.Context, reqs request.ImportMenuRequest, reimportLogUuid *uint64) (*resp.GrabMenuImportResp, error) {
	takeoutMenu, err := s.takeoutAppSrv.ConvertMenuData(ctx, reqs)
	if err != nil {
		return nil, err
	}

	platform := strings.ToLower(reqs.Platform)

	// 初始化领域服务
	db := ctx.GetDB()
	progressService := domainService.NewImportProgressService(db)
	takeoutDomainSrv := domainService.NewTakeoutDomainService(db)

	// 检查Takeout表的导入状态
	takeoutStatus, err := takeoutDomainSrv.GetImportStatus(ctx, platform)
	if err != nil {
		return nil, errors.WithMessage(err, "获取Takeout表失败")
	}
	if takeoutStatus == takeoutModel.ImportStatusSuccess {
		return nil, nil
	}

	var importLog *takeoutModel.TakeoutImportLog

	// 如果是重新导入，复用原日志
	if reimportLogUuid != nil {
		// 查询并重置日志
		log, err := persistence.NewTakeoutImportLogRepository(db).FindByUUID(ctx, *reimportLogUuid)
		if err != nil {
			return nil, errors.WithMessage(err, "查询导入日志失败")
		}
		if log == nil {
			return nil, errors.New("导入日志不存在")
		}

		// 重置日志状态
		log.Status = takeoutModel.ImportLogStatusInProgress
		log.Progress = 0
		log.ErrorMessage = ""
		log.SuccessCount = 0
		log.FailureCount = 0
		log.StartTime = time.Now().Unix()
		log.EndTime = 0
		log.Duration = 0
		log.UpdateTime = time.Now().Unix()

		if err := persistence.NewTakeoutImportLogRepository(db).Update(ctx, log); err != nil {
			return nil, errors.WithMessage(err, "重置日志状态失败")
		}
		importLog = log
	} else {
		// 1. 检查导入状态
		canImport, inProgressLog, err := progressService.CheckImportStatus(ctx, platform)
		if err != nil {
			return nil, errors.WithMessage(err, "检查导入状态失败")
		}
		if !canImport {
			errMsg := fmt.Sprintf("已有导入任务正在进行中 (UUID: %d)", inProgressLog.UUID)
			return nil, errors.New(errMsg)
		}

		// 2. 开始导入，创建导入日志
		// importType: 2 = 平台推送到TTPOS
		// importDirection: 根据平台名称设置
		importDirection := fmt.Sprintf("%s推送到TTPOS", strings.ToUpper(platform))
		log, err := progressService.StartImport(ctx, platform, 2, importDirection)
		if err != nil {
			return nil, errors.WithMessage(err, "开始导入失败")
		}
		importLog = log
	}

	// 2.1 更新 ttpos_takeout 表的导入状态为"导入中"
	if err := takeoutDomainSrv.UpdateImportStatus(ctx, platform, takeoutModel.ImportStatusInProgress, takeoutMenu); err != nil {
		logger.Logger.Warn("更新导入状态为导入中失败",
			zap.String("platform", platform),
			zap.Error(err),
		)
	}

	var successCount int
	var failureCount int
	var errorList []ProductSyncError

	// 3. 执行导入流程
	err = ctx.GetDB().Transaction(func(tx *gorm.DB) error {
		copyCtx := ctx.Copy()
		copyCtx.SetDB(tx)

		// 3.1 提取分类并创建或更新 (进度 0-10%)
		categoryMap, err := s.syncCategories(copyCtx, platform, takeoutMenu.Categories)
		if err != nil {
			return err
		}

		// 更新进度到 10%
		_ = progressService.UpdateProgress(ctx, importLog.UUID, 10, 100)

		// 3.2 同步商品规格（标准规格）(进度 10-15%)
		productFlavorUuid, err := s.syncProductFlavor(copyCtx, platform)
		if err != nil {
			return err
		}

		// 更新进度到 15%
		_ = progressService.UpdateProgress(ctx, importLog.UUID, 15, 100)

		// 3.3 同步商品单位（标准单位）(进度 15-20%)
		unitUuid, err := s.syncProductUnit(copyCtx, platform)
		if err != nil {
			return err
		}

		// 更新进度到 20%
		_ = progressService.UpdateProgress(ctx, importLog.UUID, 20, 100)

		// 3.4 导入商品 (进度 20-100%，根据实际商品数量更新)
		success, failure, errors, err := s.syncProducts(
			copyCtx,
			platform,
			takeoutMenu.Categories,
			categoryMap,
			productFlavorUuid,
			unitUuid,
			progressService,
			importLog.UUID,
		)
		if err != nil {
			return err
		}
		successCount = success
		failureCount = failure
		errorList = errors

		return nil
	})

	// 4. 根据结果完成导入
	if err != nil {
		logger.Logger.Error("导入菜单失败",
			zap.String("platform", platform),
			zap.Uint64("import_log_uuid", importLog.UUID),
			zap.Error(err),
		)

		// 标记导入失败
		_ = progressService.CompleteImport(ctx, importLog.UUID, false, err.Error())

		// 4.1 更新 ttpos_takeout 表的导入状态为"导入失败"
		if updateErr := takeoutDomainSrv.UpdateImportStatus(ctx, platform, takeoutModel.ImportStatusFailed, nil); updateErr != nil {
			logger.Logger.Warn("更新导入状态为导入失败失败",
				zap.String("platform", platform),
				zap.Error(updateErr),
			)
		}

		return nil, errors.WithMessage(err, "导入菜单失败")
	}

	// 标记导入成功
	var completeErr error
	var importStatus int8
	var importError string

	if failureCount > 0 {
		// 部分成功，标记为失败
		importStatus = 3
		importError = utils.ToJson(errorList)
		completeErr = progressService.CompleteImport(ctx, importLog.UUID, false, importError)
	} else {
		// 全部成功
		importStatus = 2
		importError = ""
		completeErr = progressService.CompleteImport(ctx, importLog.UUID, true, "")
	}

	if completeErr != nil {
		logger.Logger.Error("完成导入状态更新失败",
			zap.String("platform", platform),
			zap.Uint64("import_log_uuid", importLog.UUID),
			zap.Error(completeErr),
		)
	}

	// 4.2 更新 ttpos_takeout 表的导入状态
	if updateErr := takeoutDomainSrv.UpdateImportStatus(ctx, platform, importStatus, nil); updateErr != nil {
		logger.Logger.Warn("更新导入完成状态失败",
			zap.String("platform", platform),
			zap.Int8("status", importStatus),
			zap.Error(updateErr),
		)
	}

	// 转换错误列表为响应 DTO
	errorDTOList := make([]resp.ProductImportError, 0, len(errorList))
	for _, err := range errorList {
		errorDTOList = append(errorDTOList, resp.ProductImportError{
			ProductID:   err.ProductID,
			ProductName: err.ProductName,
			CategoryID:  err.CategoryID,
			Error:       err.Error,
		})
	}

	return &resp.GrabMenuImportResp{
		SuccessCount: successCount,
		FailureCount: failureCount,
		ErrorList:    errorDTOList,
	}, nil
}

// syncCategories 同步外卖平台分类到本地
// 返回：map[platformCategoryID]localCategoryUuid
func (s *takeoutSrv) syncCategories(ctx context.Context, platform string, categories []*valueobject.Category) (map[string]uint64, error) {
	categoryMap := make(map[string]uint64)
	db := ctx.GetDB()
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	// 获取所有现有的平台分类
	existingCategories, err := productRepo.GetProductCategoryList(
		commonRepo.Preload(repository.WithPreload{Query: "MultiLanguageName"}),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "获取现有分类失败")
	}

	// 建立 source_id 到分类的映射
	existingMap := make(map[string]model.ProductCategory)
	existingMapTtposCat := make(map[string]model.ProductCategory)
	for i := range existingCategories {
		cat := &existingCategories[i]
		existingMapTtposCat[strconv.FormatUint(cat.Uuid, 10)] = *cat
		if cat.SourceId != "" {
			existingMap[cat.SourceId] = *cat
		}
	}

	// 遍历平台分类，判断创建或更新
	for _, category := range categories {
		if category == nil {
			continue
		}

		platformCategoryID := category.ID
		if platformCategoryID == "" {
			return nil, errors.WithMessage(errors.New("分类ID不能为空"), fmt.Sprintf("分类ID不能为空: %s", category.Name))
		}

		// 去掉 TTPOS-CAT- 前缀
		platformCategoryID = strings.TrimPrefix(platformCategoryID, "TTPOS-CAT-")

		// 检查分类是否已存在
		if existingCat, exists := existingMapTtposCat[platformCategoryID]; exists {
			// 当前不做双向同步，暂时忽略更新
			// if err := s.updateCategory(ctx, existingCat, category, platform); err != nil {
			// 	return nil, errors.WithMessage(err, fmt.Sprintf("更新分类失败: %s", category.Name))
			// }
			categoryMap[platformCategoryID] = existingCat.Uuid
		} else if existingCat, exists := existingMap[platformCategoryID]; exists {
			// 当前不做双向同步，暂时忽略更新
			// if err := s.updateCategory(ctx, existingCat, category, platform); err != nil {
			// 	return nil, errors.WithMessage(err, fmt.Sprintf("更新分类失败: %s", category.Name))
			// }
			categoryMap[platformCategoryID] = existingCat.Uuid
		} else {
			// 创建新分类
			productCategory, err := s.createCategory(ctx, category, platform)
			if err != nil {
				return nil, errors.WithMessage(err, fmt.Sprintf("创建分类失败: %s", category.Name))
			}
			categoryMap[platformCategoryID] = productCategory.Uuid
		}
	}

	return categoryMap, nil
}

// createCategory 创建新分类
func (s *takeoutSrv) createCategory(ctx context.Context, category *valueobject.Category, platform string) (model.ProductCategory, error) {
	// 准备多语言名称
	localeName := language.MapToLocaleResponse(category.NameTranslation, category.Name)

	// 通过 ProductService 创建分类
	displayInStore := 1
	displayInTakeout := 1
	categoryReq := req.ProductShopCategoryAddReq{
		ParentUuid:         0, // 外卖分类作为一级分类
		LocaleName:         localeName,
		IsSpecial:          false,
		Status:             category.GetCategoryStatusToInt(),
		IsDisplayInStore:   &displayInStore,   // 外卖分类在店内显示
		IsDisplayInTakeout: &displayInTakeout, // 外卖分类在外卖平台显示
		Source:             platform,
		SourceId:           category.ID,
	}

	productCategory, err := s.productSrv.AddProductShopCategory(ctx, categoryReq)
	if err != nil {
		return model.ProductCategory{}, errors.WithMessage(err, "创建分类失败")
	}

	if productCategory.MultiLanguageNameUuid > 0 {
		if err := s.translateSrv.AddMultiLanguageNameUuidToSet(ctx.GetCompanyUuid(), productCategory.MultiLanguageNameUuid); err != nil {
			return model.ProductCategory{}, errors.WithMessage(err, "同步基础规格添加多语言uuid到待翻译集合中失败")
		}
	}

	// 清除缓存
	s.clearCategoryCache(ctx)

	return productCategory, nil
}

// updateCategory 更新现有分类
func (s *takeoutSrv) updateCategory(ctx context.Context, existingCat model.ProductCategory, category *valueobject.Category, platform string) error {
	if !existingCat.IsEditable() {
		return nil
	}
	// 准备多语言名称
	localeName := existingCat.MultiLanguageName.GetNames()
	if len(category.NameTranslation) > 0 {
		if _, ok := category.NameTranslation["en"]; ok {
			localeName.EN = category.NameTranslation["en"]
		}
		if _, ok := category.NameTranslation["zh"]; ok {
			localeName.ZH = category.NameTranslation["zh"]
		}
		if _, ok := category.NameTranslation["th"]; ok {
			localeName.TH = category.NameTranslation["th"]
		}
		if _, ok := category.NameTranslation["my"]; ok {
			localeName.MY = category.NameTranslation["my"]
		}
	}
	if localeName.EN == "" && category.Name != "" {
		localeName.EN = category.Name
	}

	// 通过 ProductService 更新分类
	displayInStore := 1
	displayInTakeout := 1
	categoryReq := req.ProductShopCategoryEditReq{
		Uuid:               existingCat.Uuid,
		LocaleName:         localeName,
		Status:             category.GetCategoryStatusToInt(),
		IsDisplayInStore:   &displayInStore,
		IsDisplayInTakeout: &displayInTakeout,
	}

	productCategory, err := s.productSrv.EditProductShopCategory(ctx, categoryReq)
	if err != nil {
		return errors.WithMessage(err, "更新分类失败")
	}

	if productCategory.MultiLanguageNameUuid > 0 {
		// 不要覆盖多语言名称
		err = repository.NewMultiLanguageNameRepo(ctx.GetDB()).NotOverwriteMultiLanguageName(productCategory.MultiLanguageNameUuid)
		if err != nil {
			return errors.WithMessage(err, "更新多语言名称失败")
		}
		// 添加多语言名称到待翻译集合
		if err := s.translateSrv.AddMultiLanguageNameUuidToSet(ctx.GetCompanyUuid(), productCategory.MultiLanguageNameUuid); err != nil {
			return errors.WithMessage(err, "同步基础规格添加多语言uuid到待翻译集合中失败")
		}
	}

	// 清除缓存
	s.clearCategoryCache(ctx)

	return nil
}

// clearCategoryCache 清除分类缓存
func (s *takeoutSrv) clearCategoryCache(ctx context.Context) {
	// 删除分类缓存
	// 参考 product.go 的缓存清除逻辑
	companyUuid := ctx.GetCompanyUuid()
	tag := fmt.Sprintf("category%d11", companyUuid)
	_ = cache.NewTaggedCache(s.cache).TagClear(tag)
}

// syncProductFlavor 同步商品规格（标准规格）
// 检查是否已存在标准规格，不存在则创建
func (s *takeoutSrv) syncProductFlavor(ctx context.Context, platform string) (uint64, error) {
	db := ctx.GetDB()
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	// 查询是否已存在该平台的标准规格
	existingFlavor, err := productRepo.GetProductFlavor(
		commonRepo.WhereBySoftDelete(),
		func(db *gorm.DB) *gorm.DB {
			return db.Where("source = ? AND source_id = ?", platform, "standard")
		},
	)

	if err == nil && existingFlavor.Uuid > 0 {
		// 规格已存在，直接返回
		return existingFlavor.Uuid, nil
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		// 查询出错
		return 0, errors.WithMessage(err, "查询商品规格失败")
	}

	// 规格不存在，创建新规格
	productFlavorUuid, err := s.productSrv.AddProductFlavor(ctx, req.ProductFlavorAddReq{
		LocaleName: dto.LocaleResponse{
			EN:   "Standard",
			ZH:   "标准",
			TH:   "มาตรฐาน",
			JA:   "標準",
			KO:   "표준",
			MY:   "Standard",
			TR:   "Standart",
			SV:   "Standard",
			ZHTW: "標準",
		},
		Source:   platform,
		SourceId: "standard",
	})
	if err != nil {
		return 0, errors.WithMessage(err, "创建商品规格失败")
	}

	return productFlavorUuid, nil
}

// syncProductUnit 同步商品单位（标准单位）
// 检查是否已存在标准单位，不存在则创建
func (s *takeoutSrv) syncProductUnit(ctx context.Context, platform string) (uint64, error) {
	db := ctx.GetDB()
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	// 查询是否已存在该平台的标准单位
	existingUnit, err := productRepo.GetProductUnit(
		commonRepo.WhereBySoftDelete(),
		func(db *gorm.DB) *gorm.DB {
			return db.Where("source = ? AND source_id = ?", platform, "standard")
		},
	)

	if err == nil && existingUnit.Uuid > 0 {
		// 单位已存在，直接返回
		return existingUnit.Uuid, nil
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		// 查询出错
		return 0, errors.WithMessage(err, "查询商品单位失败")
	}

	// 单位不存在，创建新单位
	err = s.productSrv.AddProductUnit(ctx, req.ProductUnitAddReq{
		LocaleName: dto.LocaleResponse{
			EN:   platform,
			ZH:   platform,
			TH:   platform,
			JA:   platform,
			KO:   platform,
			MY:   platform,
			TR:   platform,
			SV:   platform,
			ZHTW: platform,
		},
		Source:   platform,
		SourceId: "standard",
	})
	if err != nil {
		return 0, errors.WithMessage(err, "创建商品单位失败")
	}

	// 重新查询获取 UUID
	existingUnit, err = productRepo.GetProductUnit(
		commonRepo.WhereBySoftDelete(),
		func(db *gorm.DB) *gorm.DB {
			return db.Where("source = ? AND source_id = ?", platform, "standard")
		},
	)
	if err != nil {
		return 0, errors.WithMessage(err, "获取商品单位UUID失败")
	}

	return existingUnit.Uuid, nil
}

// syncModifierGroups 同步商品的 ModifierGroups 为属性组
// 将外卖平台的修饰符组转换为 TTPOS 的属性组
func (s *takeoutSrv) syncModifierGroups(ctx context.Context, platform string, modifierGroups []*valueobject.ModifierGroup) ([]req.ProductShopAddAttributeGroupReq, error) {
	if len(modifierGroups) == 0 {
		return []req.ProductShopAddAttributeGroupReq{}, nil
	}

	attributeGroups := make([]req.ProductShopAddAttributeGroupReq, 0, len(modifierGroups))

	for _, modifierGroup := range modifierGroups {
		if modifierGroup == nil || len(modifierGroup.Modifiers) == 0 {
			continue
		}

		// 1. 同步属性组
		attributeGroupUuid, err := s.syncAttributeGroup(ctx, platform, modifierGroup)
		if err != nil {
			return nil, err
		}
		if attributeGroupUuid == 0 {
			continue
		}

		// 2. 批量同步属性值
		attributeUuids, err := s.syncAttributesBatch(ctx, platform, attributeGroupUuid, modifierGroup.Modifiers)
		if err != nil {
			return nil, err
		}

		if len(attributeUuids) == 0 {
			continue
		}

		// 3. 构建属性列表
		attributes := make([]req.ProductShopAddGroupAttributeReq, 0, len(attributeUuids))
		for _, attributeUuid := range attributeUuids {
			attributes = append(attributes, req.ProductShopAddGroupAttributeReq{
				Uuid:              attributeUuid,
				IsDefaultSelected: 0, // 默认不选中
			})
		}

		// 4. 构建商品属性组请求
		isMust := 0
		if modifierGroup.SelectionRangeMin > 0 {
			isMust = 1 // 最小选择数量大于0，则为必选
		}

		isOpenInput := false
		maxSelection := modifierGroup.SelectionRangeMax
		if maxSelection > 1 {
			isOpenInput = true // 最大选择数量大于1，开启选择数量
		}

		attributeGroups = append(attributeGroups, req.ProductShopAddAttributeGroupReq{
			Uuid:         attributeGroupUuid,
			IsMust:       isMust,
			IsOpenInput:  isOpenInput,
			MaxSelection: maxSelection,
			Attributes:   attributes,
		})
	}

	return attributeGroups, nil
}

// syncAttributeGroup 同步属性组
func (s *takeoutSrv) syncAttributeGroup(ctx context.Context, platform string, modifierGroup *valueobject.ModifierGroup) (uint64, error) {
	db := ctx.GetDB()
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	modifierGroupID := strings.TrimPrefix(modifierGroup.ID, "TTPOS-ATTR-GROUP-")

	// 查询是否已存在该属性组
	existingGroup, err := productRepo.GetProductAttributeGroup(
		commonRepo.WhereBySoftDelete(),
		commonRepo.Preload(repository.WithPreload{Query: "MultiLanguageName"}),
		func(db *gorm.DB) *gorm.DB {
			return db.Where("((source = ? AND source_id = ?) OR uuid = ?)", platform, modifierGroup.ID, modifierGroupID)
		},
	)
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	if err == nil && existingGroup.Uuid > 0 {
		// 属性组已存在，直接返回
		// 当前不做双向同步，暂时忽略更新
		return existingGroup.Uuid, nil
	}

	// 如果是 TTPOS 开头的 ID，说明是本地创建的，不应该由外卖平台同步创建
	if strings.Contains(modifierGroup.ID, "TTPOS-ATTR-GROUP-") {
		return 0, nil
	}

	// 属性组不存在，创建新属性组
	// 构建多语言名称
	localeName := language.MapToLocaleResponse(modifierGroup.NameTranslation)
	localeName.SetAllEmptyLocale(modifierGroup.Name)
	attributeGroupUuid, err := s.productSrv.AddProductAttributeGroup(ctx, req.ProductAttributeGroupAddReq{
		LocaleName:        localeName,
		Source:            platform,
		SourceId:          modifierGroup.ID,
		ProductAttributes: []req.ProductAttributeGroupAddProductAttributeReq{},
	})
	if err != nil {
		return 0, err
	}

	return attributeGroupUuid, nil
}

// syncAttributesBatch 批量同步属性值
// 将一个属性组的所有属性值一次性同步，避免循环调用数据库
func (s *takeoutSrv) syncAttributesBatch(ctx context.Context, platform string, attributeGroupUuid uint64, modifiers []*valueobject.Modifier) ([]uint64, error) {
	if len(modifiers) == 0 {
		return []uint64{}, nil
	}

	db := ctx.GetDB()
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	// 1. 获取属性组信息（包含已有属性）
	attrGroup, err := productRepo.GetProductAttributeGroup(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereUuid(attributeGroupUuid),
		commonRepo.Preload(repository.WithPreload{Query: "MultiLanguageName"}),
		commonRepo.Preload(repository.WithPreload{Query: "ProductAttributes"}),
		commonRepo.Preload(repository.WithPreload{Query: "ProductAttributes.MultiLanguageName"}),
	)
	if err != nil {
		return nil, err
	}

	// 2. 构建 source_id -> modifier 的映射
	modifierMap := make(map[string]*valueobject.Modifier)
	ttposAttrIds := make([]string, 0, len(modifiers))
	for _, modifier := range modifiers {
		if modifier == nil {
			continue
		}
		// 如果是 TTPOS 开头的 ID（非属性组），说明是本地创建的，跳过
		if strings.Contains(modifier.ID, "TTPOS") && !strings.Contains(modifier.ID, "TTPOS-ATTR-") {
			continue
		}
		if strings.Contains(modifier.ID, "TTPOS-ATTR-") {
			attrId := strings.TrimPrefix(modifier.ID, "TTPOS-ATTR-")
			ttposAttrIds = append(ttposAttrIds, attrId)
		}
		modifierMap[modifier.ID] = modifier
	}

	// 3. 查询已存在的属性（通过 source 和 source_id）
	sourceIds := make([]string, 0, len(modifierMap))
	for sourceId := range modifierMap {
		sourceIds = append(sourceIds, sourceId)
	}

	existingAttrs, err := productRepo.GetProductAttributes(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereAttributeGroupUuid(attributeGroupUuid),
		func(db *gorm.DB) *gorm.DB {
			return db.Where("((source = ? AND source_id IN ?) OR uuid IN ?)", platform, sourceIds, ttposAttrIds)
		},
		commonRepo.Preload(repository.WithPreload{Query: "MultiLanguageName"}),
	)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 4. 构建已存在属性的映射：source_id -> ProductAttribute
	existingAttrMap := make(map[string]model.ProductAttribute)
	for _, attr := range existingAttrs {
		existingAttrMap[attr.SourceId] = attr
	}

	// 5. 准备编辑请求：收集所有需要的属性（已有的 + 新增的）
	editReq := req.ProductAttributeGroupEditReq{
		Uuid:              attributeGroupUuid,
		LocaleName:        dto.LocaleResponse{},
		ProductAttributes: []req.ProductAttributeGroupEditProductAttributeReq{},
	}

	// 填充属性组名称
	if attrGroup.MultiLanguageName.Uuid > 0 {
		names := attrGroup.MultiLanguageName.GetNames()
		editReq.LocaleName = names
	}

	// 6. 添加已有的属性（过滤已软删除的）
	hasNewAttribute := false
	for _, attr := range attrGroup.ProductAttributes {
		// 跳过已软删除的属性
		if attr.DeleteTime > 0 {
			continue
		}
		// 检查是否需要更新此属性
		if attr.SourceId == "" {
			attr.SourceId = fmt.Sprintf("TTPOS-ATTR-%d", attr.Uuid)
		}
		if modifier, ok := modifierMap[attr.SourceId]; ok {
			// 当前不做双向同步，暂时忽略更新
			_ = modifier
			// // 属性存在且需要更新多语言
			// localeName := attr.MultiLanguageName.GetNames()
			// if _, ok := modifier.NameTranslation["en"]; ok {
			// 	localeName.SetLocale("en", modifier.NameTranslation["en"])
			// }
			// if _, ok := modifier.NameTranslation["zh"]; ok {
			// 	localeName.SetLocale("zh", modifier.NameTranslation["zh"])
			// }
			// if _, ok := modifier.NameTranslation["th"]; ok {
			// 	localeName.SetLocale("th", modifier.NameTranslation["th"])
			// }
			// if _, ok := modifier.NameTranslation["my"]; ok {
			// 	localeName.SetLocale("my", modifier.NameTranslation["my"])
			// }
			// if localeName.EN == "" && modifier.Name != "" {
			// 	localeName.EN = modifier.Name
			// }
			// // 将价格从分转换为元
			// editReq.ProductAttributes = append(editReq.ProductAttributes, req.ProductAttributeGroupEditProductAttributeReq{
			// 	Uuid:       attr.Uuid,
			// 	LocaleName: localeName,
			// 	Sort:       modifier.Sequence,
			// 	Price:      attr.Price,
			// 	Source:     attr.Source,
			// 	SourceId:   attr.SourceId,
			// 	ProductPackageUuids: func() []uint64 {
			// 		productPackageUuids := make([]uint64, 0, len(attr.ProductPackageAttributes))
			// 		for _, productPackageAttribute := range attr.ProductPackageAttributes {
			// 			productPackageUuids = append(productPackageUuids, productPackageAttribute.ProductPackageAttributeGroup.ProductPackageUuid)
			// 		}
			// 		return productPackageUuids
			// 	}(),
			// })
			// hasNewAttribute = true
			// 从 modifierMap 中移除已处理的
			delete(modifierMap, attr.SourceId)
		} else if attr.MultiLanguageName.Uuid > 0 {
			// 属性存在但不在本次同步列表中，保持原样
			attrNames := attr.MultiLanguageName.GetNames()
			editReq.ProductAttributes = append(editReq.ProductAttributes, req.ProductAttributeGroupEditProductAttributeReq{
				Uuid:       attr.Uuid,
				LocaleName: attrNames,
				Sort:       attr.Sort,
				Price:      attr.Price, // 保持原有价格
				Source:     attr.Source,
				SourceId:   attr.SourceId,
				ProductPackageUuids: func() []uint64 {
					productPackageUuids := make([]uint64, 0, len(attr.ProductPackageAttributes))
					for _, productPackageAttribute := range attr.ProductPackageAttributes {
						productPackageUuids = append(productPackageUuids, productPackageAttribute.ProductPackageAttributeGroup.ProductPackageUuid)
					}
					return productPackageUuids
				}(),
			})
		}
	}

	// 7. 添加新属性（modifierMap 中剩余的就是新属性）
	for _, modifier := range modifierMap {
		hasNewAttribute = true
		localeName := language.MapToLocaleResponse(modifier.NameTranslation)
		localeName.SetAllEmptyLocale(modifier.Name)

		// 将价格从分转换为元
		price := float64(modifier.Price) / 100.0

		editReq.ProductAttributes = append(editReq.ProductAttributes, req.ProductAttributeGroupEditProductAttributeReq{
			Uuid:                0, // 新属性
			LocaleName:          localeName,
			Sort:                modifier.Sequence,
			Price:               price,
			ProductPackageUuids: []uint64{},
			Source:              platform,
			SourceId:            modifier.ID,
		})
	}

	// 8. 只有在有新属性或需要更新时才调用编辑接口
	if hasNewAttribute || len(editReq.ProductAttributes) != len(attrGroup.ProductAttributes) {
		err = s.productSrv.EditProductAttributeGroup(ctx, editReq)
		if err != nil {
			return nil, err
		}
	}

	// 9. 重新查询所有属性获取最新的 UUID
	finalAttrs, err := productRepo.GetProductAttributes(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereAttributeGroupUuid(attributeGroupUuid),
		func(db *gorm.DB) *gorm.DB {
			return db.Where("source = ? AND source_id IN ?", platform, sourceIds)
		},
	)
	if err != nil {
		return nil, err
	}

	// 获取最新的属性UUID
	finalAttributeUuids := make([]uint64, 0, len(finalAttrs))
	for _, attr := range finalAttrs {
		finalAttributeUuids = append(finalAttributeUuids, attr.Uuid)
	}

	return finalAttributeUuids, nil
}

// ProductSyncError 商品同步错误
type ProductSyncError struct {
	ProductID   string `json:"product_id"`   // 商品ID
	ProductName string `json:"product_name"` // 商品名称
	CategoryID  string `json:"category_id"`  // 分类ID
	Error       string `json:"error"`        // 错误信息
}

// syncProducts 同步外卖平台商品到本地
// 返回：成功数量、失败数量、错误列表、错误
func (s *takeoutSrv) syncProducts(
	ctx context.Context,
	platform string,
	categories []*valueobject.Category,
	categoryMap map[string]uint64,
	productFlavorUuid uint64,
	unitUuid uint64,
	progressService *domainService.ImportProgressService,
	importLogUUID uint64,
) (int, int, []ProductSyncError, error) {
	db := ctx.GetDB()
	takeoutRepo := repository.NewProductPackageTakeoutRepo(db)

	successCount := 0
	failureCount := 0
	errorList := make([]ProductSyncError, 0)

	// 计算总商品数量
	totalCount := 0
	for _, category := range categories {
		if category != nil {
			totalCount += len(category.Items)
		}
	}

	// 更新总数量
	if progressService != nil && importLogUUID > 0 {
		_ = progressService.UpdateProgress(ctx, importLogUUID, 0, totalCount)
	}

	processedCount := 0

	// 遍历所有分类及其商品
	for _, category := range categories {
		if category == nil || len(category.Items) == 0 {
			continue
		}
		category.ID = strings.TrimPrefix(category.ID, "TTPOS-CAT-")

		// 获取本地分类 UUID
		localCategoryUuid, ok := categoryMap[category.ID]
		if !ok {
			// 分类映射失败，记录该分类下的所有商品为失败
			for _, item := range category.Items {
				if item != nil {
					errorList = append(errorList, ProductSyncError{
						ProductID:   item.ID,
						ProductName: item.Name,
						CategoryID:  category.ID,
						Error:       fmt.Sprintf("分类映射失败: %s", category.Name),
					})
				}
			}
			failureCount += len(category.Items)
			continue
		}

		// 遍历分类下的所有商品
		for _, item := range category.Items {
			if item == nil {
				continue
			}

			attributeGroups, err := s.syncModifierGroups(ctx, platform, item.ModifierGroups)
			if err != nil {
				errorMsg := fmt.Sprintf("同步商品属性组失败: %v", err)
				logger.Logger.Error("同步商品属性组失败",
					zap.String("product_id", item.ID),
					zap.String("product_name", item.Name),
					zap.Error(err),
				)
				errorList = append(errorList, ProductSyncError{
					ProductID:   item.ID,
					ProductName: item.Name,
					CategoryID:  category.ID,
					Error:       errorMsg,
				})
				failureCount++
				processedCount++

				// 更新进度
				if progressService != nil && importLogUUID > 0 {
					_ = progressService.UpdateProgress(ctx, importLogUUID, processedCount, totalCount)
				}
				continue
			}

			// 检查商品是否已导入（通过 source 和 source_product_id）
			existingTakeout, err := takeoutRepo.GetProductPackageTakeoutBySourceProductId(platform, item.ID)
			if err != nil && err != gorm.ErrRecordNotFound {
				errorMsg := fmt.Sprintf("查询外卖商品失败: %v", err)
				logger.Logger.Error("查询外卖商品失败",
					zap.String("platform", platform),
					zap.String("item_id", item.ID),
					zap.String("product_name", item.Name),
					zap.Error(err),
				)
				errorList = append(errorList, ProductSyncError{
					ProductID:   item.ID,
					ProductName: item.Name,
					CategoryID:  category.ID,
					Error:       errorMsg,
				})
				failureCount++
				processedCount++

				// 更新进度
				if progressService != nil && importLogUUID > 0 {
					_ = progressService.UpdateProgress(ctx, importLogUUID, processedCount, totalCount)
				}
				continue
			}

			if existingTakeout != nil && existingTakeout.Uuid > 0 {
				// 商品已存在，更新状态和分类
				err = takeoutRepo.UpdateProductPackageTakeout(
					map[string]any{
						"status":        item.AvailableStatus.ToInt(),
						"category_uuid": localCategoryUuid,
					},
					repository.CommonRepo.WhereByUuid(existingTakeout.Uuid),
				)
				if err != nil {
					logger.Logger.Error("更新外卖商品失败",
						zap.String("platform", platform),
						zap.String("item_id", item.ID),
						zap.String("product_name", item.Name),
						zap.Uint64("uuid", existingTakeout.Uuid),
						zap.Error(err),
					)
					errorList = append(errorList, ProductSyncError{
						ProductID:   item.ID,
						ProductName: item.Name,
						CategoryID:  category.ID,
						Error:       fmt.Sprintf("更新外卖商品失败: %v", err.Error()),
					})
					failureCount++
					processedCount++

					// 更新进度
					if progressService != nil && importLogUUID > 0 {
						_ = progressService.UpdateProgress(ctx, importLogUUID, processedCount, totalCount)
					}
					continue
				}
				successCount++
				processedCount++

				// 更新进度
				if progressService != nil && importLogUUID > 0 {
					_ = progressService.UpdateProgress(ctx, importLogUUID, processedCount, totalCount)
				}
			} else {
				// 2. 创建新商品
				productUuid, err := s.createProduct(ctx, platform, item, localCategoryUuid, productFlavorUuid, unitUuid, attributeGroups)
				if err != nil {
					errorMsg := fmt.Sprintf("创建商品失败: %v", err.Error())
					errorList = append(errorList, ProductSyncError{
						ProductID:   item.ID,
						ProductName: item.Name,
						CategoryID:  category.ID,
						Error:       errorMsg,
					})
					failureCount++
					processedCount++

					// 更新进度
					if progressService != nil && importLogUUID > 0 {
						_ = progressService.UpdateProgress(ctx, importLogUUID, processedCount, totalCount)
					}
					// 因为erp商品不能回退，所以直接返回
					return successCount, failureCount, errorList, nil
				}

				// createProduct 方法中已经创建了 ProductPackageTakeout 记录，这里不需要再创建
				logger.Logger.Info("成功创建外卖商品",
					zap.String("platform", platform),
					zap.String("item_id", item.ID),
					zap.String("product_name", item.Name),
					zap.Uint64("product_uuid", productUuid),
				)
				successCount++
				processedCount++

				// 更新进度
				if progressService != nil && importLogUUID > 0 {
					_ = progressService.UpdateProgress(ctx, importLogUUID, processedCount, totalCount)
				}
			}
		}
	}

	return successCount, failureCount, errorList, nil
}

// createProduct 创建新商品
func (s *takeoutSrv) createProduct(
	ctx context.Context,
	platform string,
	item *valueobject.MenuItem,
	categoryUuid uint64,
	productFlavorUuid uint64,
	unitUuid uint64,
	attributeGroups []req.ProductShopAddAttributeGroupReq,
) (uint64, error) {

	// 获取 Grab 税类 UUID
	taxUuid := s.getTaxUuid(ctx, platform)

	// 构建商品创建请求
	productReq := s.buildProductAddReq(ctx, item, categoryUuid, productFlavorUuid, unitUuid, taxUuid, attributeGroups)

	// 调用 ProductSrv 创建商品
	productUuid, err := s.productSrv.AddProductShop(ctx, productReq)
	if err != nil {
		return 0, err
	}

	// 如果有图片 URL，异步下载并上传
	imageURL := ""
	if len(item.Photos) > 0 && item.Photos[0] != "" {
		imageURL = item.Photos[0]
	}
	// 更新商品包图片URL
	if imageURL != "" {
		productPackageRepo := repository.NewProductPackageRepo(ctx.GetDB())
		err = productPackageRepo.UpdateProductPackage(map[string]any{
			"image_url": imageURL,
		}, repository.CommonRepo.WhereByUuid(productUuid))
		if err != nil {
			return 0, err
		}
	}

	// 创建外卖商品记录到 ttpos_product_package_takeout 表
	err = s.createProductPackageTakeout(ctx, platform, item, productUuid, categoryUuid)
	if err != nil {
		return 0, err
	}

	// 如果有图片 URL，异步下载并上传
	if imageURL != "" {
		go s.asyncUploadProductImage(ctx, ctx.GetCompanyUuid(), productUuid, imageURL, item.ID, item.Name)
	}

	return productUuid, nil
}

// buildProductAddReq 构建商品创建请求
func (s *takeoutSrv) buildProductAddReq(
	_ context.Context,
	item *valueobject.MenuItem,
	categoryUuid uint64,
	productFlavorUuid uint64,
	unitUuid uint64,
	taxUuid uint64,
	attributeGroups []req.ProductShopAddAttributeGroupReq,
) req.ProductShopAddReq {
	// 构建多语言名称
	localeName := language.MapToLocaleResponse(item.NameTranslation)
	localeName.SetAllEmptyLocale(item.Name)

	// 构建卖点描述（使用商品描述）
	sellingPoint := item.GetSellingPoint()

	// 构建商品规格（单规格）
	flavors := []req.ProductShopAddFlavorReq{
		{
			Uuid:  productFlavorUuid,           // 使用同步的标准规格 UUID
			Price: float64(item.Price) / 100.0, // 转换为元
		},
	}

	// 构建商品请求
	return req.ProductShopAddReq{
		Type:         0, // 0-商品
		LocaleName:   localeName,
		SellingPoint: sellingPoint,
		CategoryUuid: categoryUuid,
		UnitUuid:     unitUuid, // 使用同步的标准单位 UUID
		Flavors:      flavors,
		Tax: req.ProductShopAddTaxReq{
			DineUuid:    taxUuid, // 使用 Grab 税类（如门店未开启税率则为 0）
			TakeoutUuid: taxUuid, // 使用 Grab 税类（如门店未开启税率则为 0）
		},
		Status:          item.AvailableStatus.ToInt(),
		ImageFileUuid:   0,               // 不上传文件，通过 image_name 存储 URL
		NumType:         0,               // 整数
		Attributes:      attributeGroups, // 使用同步的属性组
		Sauce:           req.ProductShopAddSauceReq{},
		DeductStockType: 0, // 付款减库存
		Show: req.ProductShopAddShowReq{
			IsShowCashier:   1,
			IsShowTablet:    1,
			IsShowKitchen:   1,
			IsShowAssistant: 1,
			IsShowH5:        1,
			IsShowDelivery:  1,
			IsShowKiosk:     1,
		},
		Discount: req.ProductShopAddDiscountReq{
			IsEnableMemberDiscount:  0,
			IsEnableOverallDiscount: 0,
		},
		Package:             req.ProductShopAddPackageReq{},
		ProductPrinterUuids: []uint64{},
		IsImport:            true,
	}
}

// getGrabTaxUuid 获取或创建 Grab 税类（0% 税率）
// 返回税类 UUID，如果门店未开启税率则返回 0
func (s *takeoutSrv) getTaxUuid(ctx context.Context, taxName string) uint64 {
	const grabTaxRate = 0.0

	// 1. 检查门店是否开启税率
	taxRateSetting, err := s.settingSrv.GetTaxRateSetting(ctx)
	if err != nil {
		logger.Logger.Warn("获取税率设置失败", zap.Error(err))
		return 0
	}

	// 如果门店未开启税率，返回 0
	if taxRateSetting.IsOpen != "1" {
		return 0
	}

	// 2. 查询是否已存在名为 "Grab" 的税类
	db := ctx.GetDB()
	taxRepo := repository.NewTaxRepo(db)

	// 使用 GetTaxCategoryUuidByNameOptimized 方法查询
	existingTaxUuid, err := taxRepo.GetTaxCategoryUuidByNameOptimized(taxName)
	if err != nil {
		logger.Logger.Warn("查询 Grab 税类失败", zap.Error(err))
		return 0
	}

	// 如果找到了，直接返回
	if existingTaxUuid > 0 {
		return existingTaxUuid
	}

	// 3. 不存在则创建新的 Grab 税类
	newTax := model.Tax{
		Name:    taxName,
		TaxRate: grabTaxRate,
	}

	err = taxRepo.CreateTax(newTax)
	if err != nil {
		logger.Logger.Warn("创建 Grab 税类失败", zap.Error(err))
		return 0
	}

	logger.Logger.Info("成功创建 Grab 税类",
		zap.String("name", taxName),
		zap.Float64("tax_rate", grabTaxRate),
		zap.Uint64("uuid", newTax.Uuid),
	)

	return newTax.Uuid
}

// asyncUploadProductImage 异步下载并上传商品图片
func (s *takeoutSrv) asyncUploadProductImage(
	ctx context.Context,
	companyUuid uint64,
	productUuid uint64,
	imageURL string,
	productID string,
	productName string,
) error {
	db := s.dbm.GetDB(companyUuid)

	logger.Logger.Info("开始异步下载商品图片",
		zap.String("product_id", productID),
		zap.String("product_name", productName),
		zap.Uint64("product_uuid", productUuid),
		zap.String("image_url", imageURL),
	)

	// 1. 下载图片
	resp, err := http.Get(imageURL)
	if err != nil {
		logger.Logger.Warn("异步下载商品图片失败",
			zap.String("url", imageURL),
			zap.String("product_id", productID),
			zap.Uint64("product_uuid", productUuid),
			zap.Error(err),
		)
		return errors.WithMessage(err, "异步下载商品图片失败")
	}
	defer resp.Body.Close()

	// 2. 检查 HTTP 响应状态
	if resp.StatusCode != http.StatusOK {
		logger.Logger.Warn("异步下载商品图片失败，HTTP 状态码异常",
			zap.String("url", imageURL),
			zap.String("product_id", productID),
			zap.Uint64("product_uuid", productUuid),
			zap.Int("status_code", resp.StatusCode),
		)
		return errors.WithMessage(err, "异步下载商品图片失败，HTTP 状态码异常")
	}

	// 3. 从 URL 或 Content-Type 推断文件名
	fileName := s.getFileNameFromURL(imageURL, resp.Header.Get("Content-Type"))

	// 4. 上传图片到系统
	uploadResp, err := s.uploadFileSrv.UploadImage(
		ctx,
		resp.Body,
		fileName,
		resp.ContentLength,
		0,         // groupId 为 0
		"takeout", // source 标识为 takeout
	)
	if err != nil {
		logger.Logger.Warn("异步上传商品图片失败",
			zap.String("url", imageURL),
			zap.String("product_id", productID),
			zap.Uint64("product_uuid", productUuid),
			zap.Error(err),
		)
		return errors.WithMessage(err, "异步上传商品图片失败")
	}

	// 5. 更新商品的 image_file_uuid
	err = db.Model(&model.ProductPackage{}).
		Where("uuid = ?", productUuid).
		Update("image_file_uuid", uploadResp.Uuid).Error
	if err != nil {
		logger.Logger.Warn("更新商品图片UUID失败",
			zap.String("product_id", productID),
			zap.Uint64("product_uuid", productUuid),
			zap.Uint64("file_uuid", uploadResp.Uuid),
			zap.Error(err),
		)
		return errors.WithMessage(err, "更新商品图片UUID失败")
	}

	// 更新外卖商品记录的图片UUID
	takeoutRepo := repository.NewProductPackageTakeoutRepo(db)
	err = takeoutRepo.UpdateProductPackageTakeout(map[string]any{
		"image_file_uuid": uploadResp.Uuid,
	}, repository.CommonRepo.WhereByUuid(productUuid))
	if err != nil {
		logger.Logger.Warn("更新外卖商品记录的图片UUID失败",
			zap.String("product_id", productID),
			zap.Uint64("product_uuid", productUuid),
			zap.Error(err),
		)
		return errors.WithMessage(err, "更新外卖商品记录的图片UUID失败")
	}

	logger.Logger.Info("成功上传商品图片",
		zap.String("product_id", productID),
		zap.Uint64("product_uuid", productUuid),
		zap.Uint64("file_uuid", uploadResp.Uuid),
	)
	return nil
}

// createProductPackageTakeout 创建外卖商品记录
func (s *takeoutSrv) createProductPackageTakeout(
	ctx context.Context,
	platform string,
	item *valueobject.MenuItem,
	productPackageUuid uint64,
	categoryUuid uint64,
) error {
	db := ctx.GetDB()

	// 查询商品包信息获取多语言名称UUID
	productPackageRepo := repository.NewProductPackageRepo(db)
	productPackage, err := productPackageRepo.GetProductPackage(
		repository.CommonRepo.WhereByUuid(productPackageUuid),
		productPackageRepo.WithProductBoms(),
		productPackageRepo.WithProductPackageAttributeGroupAttributes(),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return errors.WithMessage(err, "获取商品包信息失败")
	}

	// 确定外卖类型
	takeoutType := s.getTakeoutTypeByPlatform(platform)

	// 检查是否已存在相同的外卖商品记录（通过 source 和 source_product_id）
	takeoutRepo := repository.NewProductPackageTakeoutRepo(db)
	existingTakeout, err := takeoutRepo.GetProductPackageTakeoutBySourceProductId(platform, item.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.WithMessage(err, "获取外卖商品记录失败")
	}

	// 如果已存在，更新记录
	if existingTakeout != nil && existingTakeout.Uuid > 0 {
		return nil
	}

	// 创建外卖商品记录
	_, err = s.productTakeoutSrv.AddProductTakeoutShop(ctx, req.ProductTakeoutShopAddReq{
		ProductPackageUuid:  productPackageUuid,
		TakeoutType:         takeoutType,
		CategoryUuid:        categoryUuid,
		SpecialCategoryUuid: 0,
		Flavors: func() []req.ProductTakeoutShopAddFlavorReq {
			flavors := make([]req.ProductTakeoutShopAddFlavorReq, 0)
			for _, flavor := range productPackage.ProductBoms {
				flavors = append(flavors, req.ProductTakeoutShopAddFlavorReq{
					BomUuid:        flavor.Uuid,
					Price:          flavor.Price,
					GrabModifierId: "1001",
				})
			}
			return flavors
		}(),
		Attributes: func() []req.ProductTakeoutShopAddAttributeReq {
			attributes := make([]req.ProductTakeoutShopAddAttributeReq, 0)
			for _, attributeGroup := range productPackage.ProductPackageAttributeGroups {
				for _, attribute := range attributeGroup.ProductPackageAttributes {
					attributes = append(attributes, req.ProductTakeoutShopAddAttributeReq{
						ProductPackageAttributeUuid: attribute.Uuid,
						Price:                       attribute.Price,
					})
				}
			}
			return attributes
		}(),
		Status:          item.AvailableStatus.ToInt(),
		Source:          platform,
		SourceProductId: item.ID,
	})
	if err != nil {
		return errors.WithMessage(err, "创建外卖商品记录失败")
	}

	return nil
}

// getTakeoutTypeByPlatform 根据平台名称获取外卖类型
func (s *takeoutSrv) getTakeoutTypeByPlatform(platform string) int {
	switch platform {
	case "grab":
		return 1 // Grab
	case "foodpanda":
		return 2 // FoodPanda
	default:
		return 3 // 其他
	}
}

// getFileNameFromURL 从 URL 或 Content-Type 获取文件名
func (s *takeoutSrv) getFileNameFromURL(url string, contentType string) string {
	// 尝试从 URL 路径获取文件名
	urlPath := url
	if idx := strings.Index(url, "?"); idx != -1 {
		urlPath = url[:idx] // 移除查询参数
	}

	fileName := filepath.Base(urlPath)

	// 如果从 URL 无法获取有效文件名，使用默认名称
	if fileName == "" || fileName == "." || fileName == "/" {
		fileName = "product_image"
	}

	// 如果文件名没有扩展名，从 Content-Type 推断
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = s.getExtensionFromContentType(contentType)
		fileName = fileName + ext
	}

	return fileName
}

// getExtensionFromContentType 从 Content-Type 获取文件扩展名
func (s *takeoutSrv) getExtensionFromContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "image/jpeg"):
		return ".jpg"
	case strings.Contains(contentType, "image/png"):
		return ".png"
	case strings.Contains(contentType, "image/webp"):
		return ".webp"
	case strings.Contains(contentType, "image/gif"):
		return ".gif"
	default:
		return ".jpg" // 默认使用 jpg
	}
}
