package service

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ITranslateSrv翻译服务接口
type ITranslateSrv interface {
	Translate(companyUuid uint64) error                                                           // 翻译
	TranslateAll() error                                                                          // 翻译所有
	AddMultiLanguageNameUuidToSet(companyUuid uint64, multiLanguageNameUuids ...uint64) error     // 添加多语言uuid到set
	RemoveMultiLanguageNameUuidFromSet(companyUuid uint64, multiLanguageNameUuid ...uint64) error // 从set中删除多语言uuid
}

// TranslateSrv翻译服务结构体
type TranslateSrv struct {
	dbm            *database.DBManager
	cache          cache.Cache
	cacheKeyPrefix string
}

// 全局同步任务管理器
var (
	translateTaskManager = &TranslateTaskManager{
		runningTasks: sync.Map{},
	}
)

// TranslateTaskManager 翻译任务管理器
type TranslateTaskManager struct {
	runningTasks sync.Map // key: companyUuid, value: bool
}

// tryStartTask 尝试启动任务，如果已有任务在运行则返回false
func (m *TranslateTaskManager) tryStartTask(companyUuid uint64) bool {
	_, loaded := m.runningTasks.LoadOrStore(companyUuid, true)
	return !loaded // 如果之前没有值则返回true，表示成功启动任务
}

// finishTask 完成任务
func (m *TranslateTaskManager) finishTask(companyUuid uint64) {
	m.runningTasks.Delete(companyUuid)
}

// NewTranslateSrv 创建新翻译服务
func NewTranslateSrv(dbm *database.DBManager, cache cache.Cache) ITranslateSrv {
	return NewTranslateSrvImpl(dbm, cache)
}

// NewTranslateSrvImpl 创建新翻译服务实现
func NewTranslateSrvImpl(dbm *database.DBManager, cache cache.Cache) ITranslateSrv {
	return &TranslateSrv{
		dbm:            dbm,
		cache:          cache,
		cacheKeyPrefix: "translate:company_uuid",
	}
}

// Translate 翻译
func (s *TranslateSrv) Translate(companyUuid uint64) error {
	cacheKey := fmt.Sprintf("%s:%d", s.cacheKeyPrefix, companyUuid)
	// 检查是否已有翻译任务在运行
	if !translateTaskManager.tryStartTask(companyUuid) {
		return errors.New("翻译中，请稍后再试")
	}

	// 获取要翻译的多语言Uuid
	multiLanguageNameUuids, err := s.getRedisClient().SMembers(context.Background(), cacheKey).Result()
	if err != nil {
		return err
	}
	if len(multiLanguageNameUuids) == 0 {
		return nil
	}

	// 确保任务完成时清理状态
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Error("翻译任务发生panic", zap.Uint64("companyUuid", companyUuid), zap.Any("panic", r))
		}
		translateTaskManager.finishTask(companyUuid)
		logger.Logger.Info("翻译完成", zap.Uint64("companyUuid", companyUuid))
	}()

	logger.Logger.Info("开始翻译任务", zap.Uint64("companyUuid", companyUuid), zap.Any("multiLanguageNameUuids", multiLanguageNameUuids))

	// redis中的多语言Uuid
	var uuidsInRedis []uint64
	for _, multiLanguageNameUuid := range multiLanguageNameUuids {
		uuid, _ := strconv.ParseUint(multiLanguageNameUuid, 10, 64)
		if uuid != 0 && !slices.Contains(uuidsInRedis, uuid) {
			uuidsInRedis = append(uuidsInRedis, uuid)
		}
	}
	// db中的多语言Uuid
	var uuidsInDB []uint64

	db := s.dbm.GetDB(companyUuid)
	translateClient := utils.NewTranslateClient()
	translatedMultiLanguageMap := make(map[uint64]dto.LocaleResponse)
	// 根据Uuid分批获取多语言数据，每次获取100条
	multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(db)
	multiLanguageNameRepo.FindInBatches(100, func(batch []model.MultiLanguageName, batchRepo repository.IMultiLanguageNameRepo) error {
		var translateItems []utils.TranslateItem

		var uuidList []uint64
		for _, multiLanguageName := range batch {
			translateItems = append(translateItems, utils.TranslateItem{
				Lang:    "en",
				Content: multiLanguageName.EnName,
			})
			uuidList = append(uuidList, multiLanguageName.Uuid)
			uuidsInDB = append(uuidsInDB, multiLanguageName.Uuid)
		}
		logger.Logger.Info("Translate-TranslateWithRetry", zap.Any("translateItems", translateItems), zap.Any("uuidList", uuidList))
		// 每次翻译20条
		multiLanguageMap := translateClient.TranslateWithRetry(context.Background(), translateItems, 20)
		logger.Logger.Info("Translate-TranslateWithRetry-multiLanguageMap", zap.Any("multiLanguageMap", multiLanguageMap))
		// 重新获取集合，用于判断是否已经【手动】翻译过
		multiLanguageNameUuids, err = s.getRedisClient().SMembers(context.Background(), cacheKey).Result()
		if err != nil {
			logger.Logger.Error("Translate-Redis-GetMultiLanguageNameUuids", zap.Any("err", err))
			return nil
		}
		for _, multiLanguageName := range batch {
			translated, ok := multiLanguageMap[multiLanguageName.EnName]
			// 未翻译成功 或者 已从集合中移除，则跳过修改
			if !ok || !slices.Contains(multiLanguageNameUuids, strconv.FormatUint(multiLanguageName.Uuid, 10)) {
				continue
			}
			if multiLanguageName.NotOverwrite == 1 {
				if multiLanguageName.ZhName != "" {
					translated.ZH = multiLanguageName.ZhName
				}
				if multiLanguageName.ThName != "" {
					translated.TH = multiLanguageName.ThName
				}
				if multiLanguageName.EnName != "" {
					translated.EN = multiLanguageName.EnName
				}
				if multiLanguageName.ZhTwName != "" {
					translated.ZHTW = multiLanguageName.ZhTwName
				}
				if multiLanguageName.KoName != "" {
					translated.KO = multiLanguageName.KoName
				}
				if multiLanguageName.MyName != "" {
					translated.MY = multiLanguageName.MyName
				}
				if multiLanguageName.TrName != "" {
					translated.TR = multiLanguageName.TrName
				}
				if multiLanguageName.SvName != "" {
					translated.SV = multiLanguageName.SvName
				}
			}
			// 更新翻译
			batchRepo.UpdateMultiLanguageNameData(map[string]any{
				"zh_name":       translated.ZH,
				"th_name":       translated.TH,
				"en_name":       translated.EN,
				"zh_tw_name":    translated.ZHTW,
				"ja_name":       translated.JA,
				"ko_name":       translated.KO,
				"my_name":       translated.MY,
				"tr_name":       translated.TR,
				"sv_name":       translated.SV,
				"not_overwrite": 0,
			}, repository.CommonRepo.WhereByUuid(multiLanguageName.Uuid))
			// 翻译后 - 从集合中移除
			if err := s.RemoveMultiLanguageNameUuidFromSet(companyUuid, multiLanguageName.Uuid); err != nil {
				logger.Logger.Error("Translate-Redis-RemoveMultiLanguageNameUuidFromSet", zap.Uint64("multiLanguageNameUuid", multiLanguageName.Uuid), zap.Any("err", err))
			} else {
				translatedMultiLanguageMap[multiLanguageName.Uuid] = translated
			}
		}
		return nil
	}, func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid IN (?) AND en_name != ''", multiLanguageNameUuids)
	})

	// 不在DB中的多语言Uuid，需要从集合中移除
	uuidsNotInDB := slice.Difference(uuidsInRedis, uuidsInDB)
	if len(uuidsNotInDB) > 0 {
		logger.Logger.Info("无效的多语言uuid", zap.Any("uuidsNotInDB", uuidsNotInDB))
		s.RemoveMultiLanguageNameUuidFromSet(companyUuid, uuidsNotInDB...)
	}

	// 修改多语言uuid对应的模型的name
	if len(translatedMultiLanguageMap) > 0 {
		productUnitRepo := repository.NewProductUnitRepo(db)
		warehouseRepo := repository.NewWarehouseRepo(db)
		materialRepo := repository.NewMaterialRepo(db)
		productFlavorRepo := repository.NewProductFlavorRepo(db)
		productAttributeRepo := repository.NewProductAttributeRepo(db)
		productSauceRepo := repository.NewProductSauceRepo(db)
		productRepo := repository.NewProductRepo(db)
		productBomCardRepo := repository.NewProductBomCardRepo(db)
		productCategoryRepo := repository.NewProductCategoryRepo(db)
		productPackageTakeoutRepo := repository.NewProductPackageTakeoutRepo(db)

		for uuid, translated := range translatedMultiLanguageMap {
			translatedText := translated.ToJson()
			productUnitRepo.UpdateNameByMultiLanguageNameUuid(uuid, translatedText)            // 单位
			warehouseRepo.UpdateNameByMultiLanguageNameUuid(uuid, translatedText)               // 仓库
			materialRepo.UpdateNameByMultiLanguageNameUuid(uuid, translatedText)                // 物品
			productFlavorRepo.UpdateNameByMultiLanguageNameUuid(uuid, translatedText)           // 规格
			productAttributeRepo.UpdateNameByMultiLanguageNameUuid(uuid, translatedText)        // 属性
			productSauceRepo.UpdateNameByMultiLanguageNameUuid(uuid, translatedText)            // 加料
			productRepo.UpdateNameByMultiLanguageNameUuid(uuid, translatedText)                 // 商品
			productRepo.UpdateDescribeByDescribeMultiLanguageNameUuid(uuid, translatedText)     // 商品描述
			productBomCardRepo.UpdateNameByMultiLanguageNameUuid(uuid, translatedText)          // 成本卡
			productCategoryRepo.UpdateNameByMultiLanguageNameUuid(uuid, translatedText)         // 分类
			productPackageTakeoutRepo.UpdateNameByMultiLanguageNameUuid(uuid, translatedText)   // 外卖商品
		}
	}

	return nil
}

// AddMultiLanguageNameUuidToSet 添加多语言uuid到set
func (s *TranslateSrv) AddMultiLanguageNameUuidToSet(companyUuid uint64, multiLanguageNameUuids ...uint64) error {
	// 如果没有传入uuid，直接返回
	if len(multiLanguageNameUuids) == 0 {
		return nil
	}

	cacheKey := fmt.Sprintf("%s:%d", s.cacheKeyPrefix, companyUuid)

	// 将 []uint64 转换为 []interface{}
	members := make([]any, len(multiLanguageNameUuids))
	for i, uuid := range multiLanguageNameUuids {
		members[i] = uuid
	}

	err := s.getRedisClient().SAdd(context.Background(), cacheKey, members...).Err()
	if err != nil {
		logger.Logger.Error("添加多语言uuid到待翻译集合失败", zap.Error(err))
		return errors.WithMessage(errors.New("添加多语言uuid到待翻译集合失败"), err.Error())
	}
	utils.Go(func() {
		func(companyUuid uint64) {
			if err := s.Translate(companyUuid); err != nil {
				logger.Logger.Error("Translate-Redis-Translate", zap.Any("err", err))
			}
		}(companyUuid)
	})
	return nil
}

// RemoveMultiLanguageNameUuidFromSet 从set中删除多语言uuid
func (s *TranslateSrv) RemoveMultiLanguageNameUuidFromSet(companyUuid uint64, multiLanguageNameUuids ...uint64) error {
	cacheKey := fmt.Sprintf("%s:%d", s.cacheKeyPrefix, companyUuid)
	// 将 []uint64 转换为 []interface{}
	members := make([]any, len(multiLanguageNameUuids))
	for i, uuid := range multiLanguageNameUuids {
		members[i] = uuid
	}
	return s.getRedisClient().SRem(context.Background(), cacheKey, members...).Err()
}

// 获取redis client
func (s *TranslateSrv) getRedisClient() redis.UniversalClient {
	clusterClient := s.cache.GetClusterClient()
	if clusterClient != nil {
		return clusterClient
	}
	return s.cache.GetClient()
}

func (s *TranslateSrv) TranslateAll() error {
	// 使用封装方法，支持集群和单机模式
	keys, err := cache.ScanRedisKeysDefault(context.Background(), s.getRedisClient(), s.cacheKeyPrefix+"*")
	if err != nil {
		logger.Logger.Error("Translate-Redis-Keys", zap.Any("err", err))
		return err
	}

	var uniqueKeys []string
	for _, key := range keys {
		if slices.Contains(uniqueKeys, key) {
			continue
		} else {
			uniqueKeys = append(uniqueKeys, key)
		}
		companyUuid, err := strconv.ParseUint(strings.TrimPrefix(key, s.cacheKeyPrefix+":"), 10, 64)
		if err != nil {
			logger.Logger.Error("Translate-Redis-SMembers", zap.Any("err", err), zap.String("key", key))
			continue
		}
		multiLanguageNameUuids, err := s.getRedisClient().SMembers(context.Background(), key).Result()
		if err != nil {
			logger.Logger.Error("Translate-Redis-Translate", zap.Any("err", err), zap.String("key", key))
			continue
		}
		if len(multiLanguageNameUuids) > 0 {
			utils.Go(func() {
				if err := s.Translate(companyUuid); err != nil {
					logger.Logger.Info("Translate-Redis-Translate-go", zap.Uint64("companyUuid", companyUuid), zap.Any("err", err), zap.String("key", key))
				}
			})
		} else {
			logger.Logger.Info("Translate-Redis-Translate-no-data", zap.Uint64("companyUuid", companyUuid), zap.String("key", key))
		}
	}
	return nil
}
