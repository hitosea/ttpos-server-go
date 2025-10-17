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

// getRunningCompanyUuids 获取当前正在执行同步任务的所有companyUuid
func (m *TranslateTaskManager) getRunningCompanyUuids() []uint64 {
	var companyUuids []uint64
	m.runningTasks.Range(func(key, value any) bool {
		if companyUuid, ok := key.(uint64); ok {
			companyUuids = append(companyUuids, companyUuid)
		}
		return true
	})
	return companyUuids
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

	logger.Logger.Info("开始翻译任务", zap.Uint64("companyUuid", companyUuid))

	db := s.dbm.GetDB(companyUuid)
	translateClient := utils.NewTranslateClient()
	translatedMultiLanguageMap := make(map[uint64]dto.LocaleResponse)
	var multiLanguageNames []model.MultiLanguageName
	// 根据Uuid分批获取多语言数据，每次获取100条
	db.Model(&model.MultiLanguageName{}).Scopes(repository.NotDeleted).Where("uuid in (?) AND en_name != ''", multiLanguageNameUuids).FindInBatches(&multiLanguageNames, 100, func(tx *gorm.DB, batch int) error {
		var translateItems []utils.TranslateItem

		var uuidList []uint64
		for _, multiLanguageName := range multiLanguageNames {
			translateItems = append(translateItems, utils.TranslateItem{
				Lang:    "en",
				Content: multiLanguageName.EnName,
			})
			uuidList = append(uuidList, multiLanguageName.Uuid)
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
		for _, multiLanguageName := range multiLanguageNames {
			translated, ok := multiLanguageMap[multiLanguageName.EnName]
			// 未翻译成功 或者 已从集合中移除，则跳过修改
			if !ok || !slices.Contains(multiLanguageNameUuids, strconv.FormatUint(multiLanguageName.Uuid, 10)) {
				continue
			}
			// 更新翻译
			tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", multiLanguageName.Uuid).Updates(map[string]any{
				"zh_name":    translated.ZH,
				"th_name":    translated.TH,
				"en_name":    translated.EN,
				"zh_tw_name": translated.ZHTW,
				"ja_name":    translated.JA,
				"ko_name":    translated.KO,
				"my_name":    translated.MY,
				"tr_name":    translated.TR,
				"sv_name":    translated.SV,
			})
			// 翻译后 - 从集合中移除
			if err := s.RemoveMultiLanguageNameUuidFromSet(companyUuid, multiLanguageName.Uuid); err != nil {
				logger.Logger.Error("Translate-Redis-RemoveMultiLanguageNameUuidFromSet", zap.Uint64("multiLanguageNameUuid", multiLanguageName.Uuid), zap.Any("err", err))
			} else {
				translatedMultiLanguageMap[multiLanguageName.Uuid] = translated
			}
		}
		return nil
	})

	// 修改多语言uuid对应的模型的name
	if len(translatedMultiLanguageMap) > 0 {
		for uuid, translated := range translatedMultiLanguageMap {
			translatedText := translated.ToJson()
			db.Model(&model.ProductUnit{}).Where("multi_language_name_uuid = ?", uuid).Update("name", translatedText)      // 单位
			db.Model(&model.Warehouse{}).Where("multi_language_name_uuid = ?", uuid).Update("name", translatedText)        // 仓库
			db.Model(&model.Material{}).Where("multi_language_name_uuid = ?", uuid).Update("name", translatedText)         // 物品
			db.Model(&model.ProductFlavor{}).Where("multi_language_name_uuid = ?", uuid).Update("name", translatedText)    // 规格
			db.Model(&model.ProductAttribute{}).Where("multi_language_name_uuid = ?", uuid).Update("name", translatedText) // 属性
			db.Model(&model.ProductSauce{}).Where("multi_language_name_uuid = ?", uuid).Update("name", translatedText)     // 加料
			db.Model(&model.ProductPackage{}).Where("multi_language_name_uuid = ?", uuid).Update("name", translatedText)   // 商品
			db.Model(&model.ProductBomCard{}).Where("multi_language_name_uuid = ?", uuid).Update("name", translatedText)   // 成本卡
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
	if err := s.Translate(companyUuid); err != nil {
		logger.Logger.Error("Translate-Redis-Translate", zap.Any("err", err))
	}
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
	// redis获取所有 s.cachePrefix 开头的key，并查询set是否为空，不为空则翻译
	keys, err := s.getRedisClient().Keys(context.Background(), s.cacheKeyPrefix+"*").Result()
	if err != nil {
		logger.Logger.Error("Translate-Redis-Keys", zap.Any("err", err))
		return err
	}

	for _, key := range keys {
		companyUuid, err := strconv.ParseUint(strings.TrimPrefix(key, s.cacheKeyPrefix+":"), 10, 64)
		if err != nil {
			logger.Logger.Error("Translate-Redis-SMembers", zap.Any("err", err))
			continue
		}
		multiLanguageNameUuids, err := s.getRedisClient().SMembers(context.Background(), key).Result()
		if err != nil {
			logger.Logger.Error("Translate-Redis-Translate", zap.Any("err", err))
			continue
		}
		if len(multiLanguageNameUuids) > 0 {
			go func() {
				if err := s.Translate(companyUuid); err != nil {
					logger.Logger.Error("Translate-Redis-Translate-go", zap.Any("err", err))
				}
			}()
		} else {
			logger.Logger.Info("Translate-Redis-Translate-no-data", zap.Uint64("companyUuid", companyUuid))
		}
	}
	return nil
}
