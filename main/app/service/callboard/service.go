package callboard

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	dtosetting "ttpos-server-go/app/dto/resp/setting"
	tterrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/cachekey"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	_ActionPushToPreparingQueue = iota
	_ActionPushToPreparedQueue
	_ActionRemoveFromQueue
)

type ICallBoardService interface {
	// 展示设备端接口
	GetBindCode(ctx context.Context, req req.GetBindCodeReq) (*resp.BindCodeResp, error)
	GetBindInfo(ctx context.Context, req req.GetBindInfoReq) (*resp.BindInfoResp, error)
	GetQueueData(ctx context.Context, companyUuid uint64, req req.GetQueueDataReq) (*resp.QueueDataResp, error)

	// 商家管理端接口
	BindDevice(ctx context.Context, companyUuid uint64, req req.BindDeviceReq, clientVersion string) error
	GetDeviceList(ctx context.Context, companyUuid uint64, req req.GetDeviceListReq) (*resp.DeviceListResp, error)
	UnbindDevice(ctx context.Context, companyUuid uint64, uuid uint64) error
	BindedDeviceValidate(deviceId string) (companyUuid uint64, err error)
	UpdateBindInfo(ctx context.Context, companyUuid uint64, req req.UpdateBindInfoReq, clientVersion string) error
}

func NewCallBoardService(dbm *database.DBManager, cache cache.Cache) ICallBoardService {
	srv := &callBoardService{
		dbm:                dbm,
		cache:              cache,
		setPendingScript:   redis.NewScript(LuaScriptSetDeviceBindCode),
		clearPendingScript: redis.NewScript(LuaScriptClearPendingDevice),
		bus:                event.NewSystemBus(),
	}

	// 订阅处理结账事件
	srv.bus.SubscribeCheckoutSaleOrderEvent(func(payload event.CheckoutSaleOrderPayload) {
		logger.Logger.Info("callboard srv received checkout sale order event",
			zap.Uint64("companyUuid", payload.CompanyUuid),
			zap.Uint64("saleBillUuid", payload.SaleBillUuid),
		)
		err := srv.handleSaleBillEvent(payload.CompanyUuid, payload.SaleBillUuid, _ActionPushToPreparingQueue)
		if err != nil {
			logger.Logger.Error("callboard srv handleSaleBillEvent failed", zap.Error(err))
		}
	})

	// 订阅处理整单取消事件
	srv.bus.SubscribeCancelOrderEvent(func(payload event.CancelOrderPayload) {
		err := srv.handleSaleBillEvent(payload.CompanyUuid, payload.SaleBillUuid, _ActionRemoveFromQueue)
		if err != nil {
			logger.Logger.Error("callboard srv handleSaleBillEvent failed", zap.Error(err))
		}
	})

	// 订阅处理整单退款事件
	srv.bus.SubscribeReturnOrderEvent(func(payload event.ReturnOrderPayload) {
		err := srv.handleSaleBillEvent(payload.CompanyUuid, payload.SaleBillUuid, _ActionRemoveFromQueue)
		if err != nil {
			logger.Logger.Error("callboard srv handleSaleBillEvent failed", zap.Error(err))
		}
	})

	// 订阅处理整单反结账事件
	srv.bus.SubscribeOrderReverseSettleEvent(func(payload event.OrderReverseSettlePayload) {
		err := srv.handleSaleBillEvent(payload.CompanyUuid, payload.SaleBillUuid, _ActionRemoveFromQueue)
		if err != nil {
			logger.Logger.Error("callboard srv handleSaleBillEvent failed", zap.Error(err))
		}
	})

	// 订阅处理整单免单事件
	srv.bus.SubscribeFreeSaleOrderEvent(func(payload event.FreeSaleOrderPayload) {
		err := srv.handleSaleBillEvent(payload.CompanyUuid, payload.SaleBillUuid, _ActionPushToPreparingQueue)
		if err != nil {
			logger.Logger.Error("callboard srv handleSaleBillEvent failed", zap.Error(err))
		}
	})

	// 订阅处理整单完成制作事件
	srv.bus.SubscribeFinishMenuEvent(func(payload event.FinishMenuPayload) {
		err := srv.handleProductionOrderCookingEvent(payload.CompanyUuid, payload.SaleBillUuid)
		if err != nil {
			logger.Logger.Error("callboard srv handleProductionOrderCookingEvent failed", zap.Error(err))
		}
	})
	return srv
}

type callBoardService struct {
	dbm                *database.DBManager
	cache              cache.Cache
	setPendingScript   *redis.Script
	clearPendingScript *redis.Script
	bus                *event.SystemEventBus
}

func (s *callBoardService) getRedisClient() redis.UniversalClient {
	clusterClient := s.cache.GetClusterClient()
	if clusterClient != nil {
		return clusterClient
	}
	return s.cache.GetClient()
}

func (s *callBoardService) UpdateBindInfo(ctx context.Context, companyUuid uint64, req req.UpdateBindInfoReq, clientVersion string) error {
	repo := repository.NewDeviceRepo(s.dbm.GetDB(companyUuid))

	// 版本 >= 2.11.0 时才进行参数校验
	if utils.CompareVersion(clientVersion, utils.VersionGTE, "2.11.0") {
		if req.Name == "" {
			return errors.New("请输入名称")
		}

		if utf8.RuneCountInString(req.Name) > 20 {
			return errors.New("名称不能超过20个字")
		}

		if req.CallCount != 0 && (req.CallCount < 1 || req.CallCount > 3) {
			return errors.New("叫号次数必须在1-3之间")
		}
	}

	// 设置默认值
	if req.TimeoutLimit == nil {
		req.TimeoutLimit = &[]int{0}[0]
	}
	if *req.TimeoutLimit < 0 || *req.TimeoutLimit > 120 {
		return errors.New("超时限制必须在0-120分钟之间")
	}
	if req.VoiceCallEnabled == nil {
		req.VoiceCallEnabled = &[]bool{false}[0]
	}
	if req.CallCount == 0 {
		req.CallCount = 1
	}
	// 设置默认值后，CallCount 必须在有效范围内
	if req.CallCount < 1 || req.CallCount > 3 {
		return errors.New("叫号次数必须在1-3之间")
	}

	device, err := repo.GetDevice(repo.WhereUuid(req.Uuid))
	if err != nil {
		return err
	}
	if device.Uuid == 0 {
		return errors.New("设备不存在")
	}

	// 更新设备绑定信息
	err = s.getRedisClient().HMSet(ctx, cachekey.GetBindedDeviceKey(device.DeviceId),
		"lang1", req.Lang1,
		"lang2", req.Lang2,
		"name", req.Name,
		"background_image_url", req.BackgroundImageUrl,
		"timeout_limit", req.TimeoutLimit,
		"voice_call_enabled", req.VoiceCallEnabled,
		"call_count", req.CallCount,
	).Err()
	if err != nil {
		return err
	}

	return nil
}

// GetBindCode 获取绑定码
func (s *callBoardService) GetBindCode(ctx context.Context, req req.GetBindCodeReq) (res *resp.BindCodeResp, err error) {
	// 生成6位数字绑定码
	ttl := int64(BindCodeTTL)
	for i := 0; i < 3; i++ {
		bindCode := s.generateBindCode()
		code, expireTime, err := s.setBindCodeByLua(ctx, req.DeviceId, bindCode, ttl)
		if err != nil {
			parsedErr := ParseScriptError(err)
			if parsedErr == ErrBindCodeConflict {
				continue
			}
			return nil, err
		}
		return &resp.BindCodeResp{
			BindCode:   code,
			ExpireTime: expireTime,
		}, nil
	}
	return nil, tterrors.New("生成绑定码失败")
}

// GetBindInfo 获取绑定信息
func (s *callBoardService) GetBindInfo(ctx context.Context, req req.GetBindInfoReq) (*resp.BindInfoResp, error) {
	// 只从缓存中查询
	cacheKey := cachekey.GetBindedDeviceKey(req.DeviceId)
	mapCmd := s.getRedisClient().HGetAll(ctx, cacheKey)
	if err := mapCmd.Err(); err != nil {
		return nil, err
	}
	bindInfo := DeviceBindInfo{}
	_ = mapCmd.Scan(&bindInfo)

	return &resp.BindInfoResp{
		DeviceSecret: bindInfo.DeviceSecret,
		Lang1:        bindInfo.Lang1,
		Lang2:        bindInfo.Lang2,
	}, nil
}

// GetQueueData 获取队列数据
func (s *callBoardService) GetQueueData(ctx context.Context, companyUuid uint64, req req.GetQueueDataReq) (*resp.QueueDataResp, error) {
	bindInfo, err := s.mustGetCompanyDeviceBindInfo(companyUuid, req.DeviceId)
	if err != nil {
		return nil, err
	}

	// 获取 timeout_limit，nil 或 0 表示不过滤
	timeoutLimitMinutes := 0
	if bindInfo.TimeoutLimit != nil && *bindInfo.TimeoutLimit > 0 {
		timeoutLimitMinutes = *bindInfo.TimeoutLimit
	}

	// PreparingQueue 不进行过滤（传递 0）
	preparingQueue, err := s.getCallBoardQueue(cachekey.GetPreparingQueueKey(companyUuid), req.Limit, bindInfo.CreateTime, 0)
	if err != nil {
		return nil, err
	}

	// PreparedQueue 根据 timeout_limit 过滤
	preparedQueue, err := s.getCallBoardQueue(cachekey.GetPreparedQueueKey(companyUuid), req.Limit, bindInfo.CreateTime, timeoutLimitMinutes)
	if err != nil {
		return nil, err
	}

	utils.Go(func() {
		maxScore := time.Now().Add(-7 * 24 * time.Hour).Unix() // 删除7天前的数据
		s.removeExpireMemberFromQueues(cachekey.GetPreparingQueueKey(companyUuid), maxScore)
		s.removeExpireMemberFromQueues(cachekey.GetPreparedQueueKey(companyUuid), maxScore)
	})

	// 设置默认值
	name := bindInfo.Name
	if name == "" {
		name = "WALLACE"
	}

	timeoutLimit := bindInfo.TimeoutLimit
	if timeoutLimit == nil {
		timeoutLimit = &[]int{0}[0]
	}

	voiceCallEnabled := bindInfo.VoiceCallEnabled
	if voiceCallEnabled == nil {
		voiceCallEnabled = &[]bool{false}[0]
	}

	callCount := bindInfo.CallCount
	if callCount == 0 {
		callCount = 1
	}

	return &resp.QueueDataResp{
		Lang1:          bindInfo.Lang1,
		Lang2:          bindInfo.Lang2,
		UpdateTime:     time.Now().Unix(),
		PreparingQueue: preparingQueue,
		PreparedQueue:  preparedQueue,
		// 新增配置字段
		Name:               name,
		BackgroundImageUrl: bindInfo.BackgroundImageUrl,
		TimeoutLimit:       timeoutLimit,
		VoiceCallEnabled:   voiceCallEnabled,
		CallCount:          callCount,
	}, nil
}

// BindDevice 绑定设备
func (s *callBoardService) BindDevice(ctx context.Context, companyUuid uint64, req req.BindDeviceReq, clientVersion string) error {
	// 版本 >= 2.11.0 时才进行参数校验
	if utils.CompareVersion(clientVersion, utils.VersionGTE, "2.11.0") {
		if req.Name == "" {
			return errors.New("请输入名称")
		}
		if utf8.RuneCountInString(req.Name) > 20 {
			return errors.New("名称不能超过20个字")
		}
		if req.CallCount != 0 && (req.CallCount < 1 || req.CallCount > 3) {
			return errors.New("叫号次数必须在1-3之间")
		}
	}

	// 设置默认值
	if req.TimeoutLimit == nil {
		req.TimeoutLimit = &[]int{0}[0]
	}
	if *req.TimeoutLimit < 0 || *req.TimeoutLimit > 120 {
		return errors.New("超时限制必须在0-120分钟之间")
	}
	if req.VoiceCallEnabled == nil {
		req.VoiceCallEnabled = &[]bool{false}[0]
	}
	if req.CallCount == 0 {
		req.CallCount = 1
	}
	// 设置默认值后，CallCount 必须在有效范围内
	if req.CallCount < 1 || req.CallCount > 3 {
		return errors.New("叫号次数必须在1-3之间")
	}
	// 从Redis获取绑定码对应的设备ID
	bindCodeKey := cachekey.GetBindCodeKey(req.BindCode)
	deviceId, err := s.getRedisClient().Get(ctx, bindCodeKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return tterrors.New("绑定码无效或已过期")
		}
		return tterrors.New("redis error")
	}

	// 生成设备密钥
	deviceSecret := s.generateDeviceSecret()

	repo := repository.NewDeviceRepo(s.dbm.GetDB(companyUuid))
	device, err := repo.GetDevice(
		repo.WhereSn(deviceId),
		repo.WhereSource("call_board"),
	)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if device.Uuid != 0 && device.DeleteTime == 0 {
		bindInfo, _ := s.mustGetBindInfoFromCache(deviceId)
		if bindInfo.CompanyUuid != 0 {
			return tterrors.New("设备已绑定")
		}
	} else {
		// 写入数据库用于前端逻辑
		device.DeviceId = deviceId
		device.Source = "call_board"
		device.CreateTime = time.Now().Unix()
		device.UpdateTime = time.Now().Unix()
		device.DeleteTime = 0
		device, err = repo.CreateDevice(device)
		if err != nil {
			return tterrors.New("创建设备失败")
		}
	}

	// 展示端不走db，直接走缓存
	deviceKey := cachekey.GetBindedDeviceKey(deviceId)
	err = s.getRedisClient().HMSet(
		ctx,
		deviceKey,
		"company_uuid", companyUuid,
		"device_secret", deviceSecret,
		"lang1", req.Lang1,
		"lang2", req.Lang2,
		"name", req.Name,
		"background_image_url", req.BackgroundImageUrl,
		"timeout_limit", req.TimeoutLimit,
		"voice_call_enabled", req.VoiceCallEnabled,
		"call_count", req.CallCount,
		"create_time", time.Now().Unix(),
	).Err()
	if err != nil {
		return err
	}

	s.clearPendingDevice(ctx, deviceId) // 从等待绑定设备列表中移除
	return err
}

// GetDeviceList 获取设备列表
func (s *callBoardService) GetDeviceList(ctx context.Context, companyUuid uint64, req req.GetDeviceListReq) (*resp.DeviceListResp, error) {

	repo := repository.NewDeviceRepo(s.dbm.GetDB(companyUuid))
	devices, err := repo.GetDeviceList(repo.WhereSource("call_board"))
	if err != nil {
		return nil, err
	}

	list := make([]resp.DeviceItem, 0, len(devices))
	for _, device := range devices {
		bindInfo, _ := s.mustGetCompanyDeviceBindInfo(companyUuid, device.DeviceId)
		// 如果设备名称为空，返回默认值 "WALLACE"
		name := bindInfo.Name
		if name == "" {
			name = "WALLACE"
		}
		callCount := bindInfo.CallCount
		if callCount == 0 {
			callCount = 1
		}
		timeoutLimit := bindInfo.TimeoutLimit
		if timeoutLimit == nil {
			timeoutLimit = &[]int{0}[0]
		}
		voiceCallEnabled := bindInfo.VoiceCallEnabled
		if voiceCallEnabled == nil {
			voiceCallEnabled = &[]bool{false}[0]
		}
		backgroundImageUrl := bindInfo.BackgroundImageUrl
		if backgroundImageUrl == "" {
			backgroundImageUrl = "/"
		}
		list = append(list, resp.DeviceItem{
			Uuid:               device.Uuid,
			DeviceId:           device.DeviceId,
			BindTime:           device.CreateTime,
			Lang1:              bindInfo.Lang1,
			Lang2:              bindInfo.Lang2,
			Name:               name,
			BackgroundImageUrl: backgroundImageUrl,
			TimeoutLimit:       timeoutLimit,
			VoiceCallEnabled:   voiceCallEnabled,
			CallCount:          callCount,
		})
	}

	return &resp.DeviceListResp{
		List: list,
	}, nil
}

// UnbindDevice 解绑设备
func (s *callBoardService) UnbindDevice(ctx context.Context, companyUuid uint64, uuid uint64) error {
	repo := repository.NewDeviceRepo(s.dbm.GetDB(companyUuid))
	device, err := repo.GetDevice(repo.WhereUuid(uuid))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if device.Uuid == 0 {
		return errors.New("设备不存在")
	}
	err = repo.UpdateDevice(device.Uuid, map[string]any{"delete_time": time.Now().Unix()})
	if err != nil {
		return err
	}
	s.cache.Del(cachekey.GetBindedDeviceKey(device.DeviceId))
	return nil

}

// BindedDeviceValidate 验证设备绑定
func (s *callBoardService) BindedDeviceValidate(deviceId string) (companyUuid uint64, err error) {
	deviceKey := cachekey.GetBindedDeviceKey(deviceId)
	mapCmd := s.getRedisClient().HGetAll(context.Background(), deviceKey)
	if err := mapCmd.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, tterrors.NewWithCode(constant.CodeUnauthorized, "设备未绑定")
		}
		return 0, err
	}
	bindInfo := DeviceBindInfo{}
	_ = mapCmd.Scan(&bindInfo)
	if bindInfo.CompanyUuid == 0 {
		return 0, tterrors.NewWithCode(constant.CodeUnauthorized, "设备未绑定")
	}
	return bindInfo.CompanyUuid, nil
}

// generateBindCode 生成6位数字绑定码
func (s *callBoardService) generateBindCode() string {
	code := ""
	for i := 0; i < 6; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += n.String()
	}
	return code
}

// generateDeviceSecret 生成设备密钥
func (s *callBoardService) generateDeviceSecret() string {
	return utils.GetRandomString(32)
}

func (s *callBoardService) setBindCodeByLua(ctx context.Context, deviceId string, bindCode string, ttl int64) (code string, expireTime int64, err error) {
	keys := []string{
		cachekey.PendingDeviceListKey,
		cachekey.GetPendingDeviceKey(deviceId),
		cachekey.GetBindCodeKey(bindCode),
	}
	args := []any{
		deviceId,
		bindCode,
		ttl,
		MaxPendingDeviceCount,
	}
	client := s.getRedisClient()
	result, err := s.setPendingScript.Run(ctx, client, keys, args...).Result()
	if err != nil {
		return "", 0, err
	}
	code, expireTime, err = ParseBindCodeResult(result)
	if err != nil {
		return "", 0, err
	}
	return code, expireTime, nil
}

type DeviceBindInfo struct {
	CreateTime         int64  `redis:"create_time"`
	CompanyUuid        uint64 `redis:"company_uuid"`
	DeviceSecret       string `redis:"device_secret"`
	Lang1              string `redis:"lang1"`
	Lang2              string `redis:"lang2"`
	Name               string `redis:"name"`                 // 设备名称
	BackgroundImageUrl string `redis:"background_image_url"` // 背景图片 URL
	TimeoutLimit       *int   `redis:"timeout_limit"`        // 超时限制（分钟）
	VoiceCallEnabled   *bool  `redis:"voice_call_enabled"`   // 语音叫号开关
	CallCount          int    `redis:"call_count"`           // 叫号次数
}

func (s *callBoardService) clearPendingDevice(ctx context.Context, deviceId string) error {
	bindCode, err := s.getRedisClient().Get(ctx, cachekey.GetPendingDeviceKey(deviceId)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}

	keys := []string{
		cachekey.PendingDeviceListKey,
		cachekey.GetPendingDeviceKey(deviceId),
		cachekey.GetBindCodeKey(bindCode),
	}
	args := []any{deviceId}
	client := s.getRedisClient()
	_, err = s.clearPendingScript.Run(ctx, client, keys, args...).Result()
	if err != nil {
		return err
	}
	return nil
}

type _Setting struct {
	Key    string `json:"key"`
	Values string `json:"values"`
}

func (s *callBoardService) mustGetCompanyDeviceBindInfo(companyUuid uint64, deviceId string) (bindInfo DeviceBindInfo, err error) {
	bindInfo, err = s.mustGetBindInfoFromCache(deviceId)
	if err != nil {
		return bindInfo, err
	}

	{
		// 从缓存中获取最新的语言设置
		compSettKey := cachekey.GetCompanySettingKey(companyUuid)
		res, _ := s.cache.Get(compSettKey)
		jsonStr, _ := res.(string)
		latestLangs := make([]_Setting, 0)
		_ = json.Unmarshal([]byte(jsonStr), &latestLangs)
		for _, lang := range latestLangs {
			if lang.Key != constant.SettingStore {
				continue
			}
			storeSetting := dtosetting.Store{}
			_ = json.Unmarshal([]byte(lang.Values), &storeSetting)
			if len(storeSetting.Language) > 0 {
				bindInfo.Lang1, bindInfo.Lang2 = checkAndFixLangs(storeSetting.Language, bindInfo.Lang1, bindInfo.Lang2)
				return bindInfo, nil
			}
		}
	}

	settingRepo := repository.NewSettingRepo(s.dbm.GetDB(companyUuid))
	settingData := settingRepo.GetByKey(constant.SettingStore)
	if settingData.Key == "" {
		logger.Logger.Error("商家设置不存在或获取失败", zap.Uint64("companyUuid", companyUuid))
		return bindInfo, nil
	}
	storeSetting := dtosetting.Store{}
	_ = json.Unmarshal([]byte(settingData.Values), &storeSetting)
	bindInfo.Lang1, bindInfo.Lang2 = checkAndFixLangs(storeSetting.Language, bindInfo.Lang1, bindInfo.Lang2)
	return bindInfo, nil
}

func (s *callBoardService) mustGetBindInfoFromCache(deviceId string) (bindInfo DeviceBindInfo, err error) {
	cmd := s.getRedisClient().HGetAll(context.Background(), cachekey.GetBindedDeviceKey(deviceId))
	if err := cmd.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return bindInfo, tterrors.NewWithCode(constant.CodeUnauthorized, "设备未绑定")
		}
		return bindInfo, err
	}
	bindInfo = DeviceBindInfo{}
	_ = cmd.Scan(&bindInfo)
	if bindInfo.CompanyUuid == 0 || bindInfo.DeviceSecret == "" {
		return bindInfo, tterrors.NewWithCode(constant.CodeUnauthorized, "设备未绑定")
	}
	return bindInfo, nil
}

func checkAndFixLangs(langList []dto.LanguageItem, targetLang1 string, targetLang2 string) (lang1 string, lang2 string) {
	// 先查找 targetLang1，再查找 targetLang2，保证返回顺序与参数顺序一致
	var foundLang1, foundLang2 string
	for _, lang := range langList {
		if lang.Name == targetLang1 && foundLang1 == "" {
			foundLang1 = lang.Name
		}
		if lang.Name == targetLang2 && foundLang2 == "" {
			foundLang2 = lang.Name
		}
	}
	// 如果都没找到，使用 langList 的第一个作为 lang1
	if foundLang1 == "" && foundLang2 == "" {
		if len(langList) > 0 {
			foundLang1 = langList[0].Name
		}
	}
	return foundLang1, foundLang2
}

func (s *callBoardService) handleSaleBillEvent(companyUuid uint64, saleBillUuid uint64, action int) error {
	saleBillRepo := repository.NewSaleBillRepo(s.dbm.GetDB(companyUuid))
	saleBill, err := saleBillRepo.GetSaleBillByUuid(saleBillUuid)
	if err != nil {
		return err
	}
	member := formatQueueMember(saleBill.Uuid, saleBill.SerialNo, saleBill.CreateTime)

	// 不是点餐账单，不处理
	if saleBill.BillType != constant.SaleBillTypeInstant {
		return nil
	}
	if action == _ActionRemoveFromQueue {
		return s.removeMemberFromQueues(member, cachekey.GetPreparingQueueKey(companyUuid), cachekey.GetPreparedQueueKey(companyUuid))
	}
	if action == _ActionPushToPreparingQueue {
		return s.pushToPreparingQueue(companyUuid, member, saleBill.CreateTime)
	}
	if action == _ActionPushToPreparedQueue {
		return s.pushToPreparedQueue(companyUuid, member, time.Now().Unix())
	}
	return nil
}

func (s *callBoardService) handleProductionOrderCookingEvent(companyUuid uint64, saleBillUuid uint64) error {
	db := s.dbm.GetDB(companyUuid)
	// 获取销售账单信息
	saleBillRepo := repository.NewSaleBillRepo(db)
	saleBill, err := saleBillRepo.GetSaleBillByUuid(saleBillUuid)
	if err != nil {
		return err
	}
	// 不是点餐账单，不处理
	if saleBill.BillType != constant.SaleBillTypeInstant {
		return nil
	}
	member := formatQueueMember(saleBillUuid, saleBill.SerialNo, saleBill.CreateTime)
	// 是否已完成制作
	finished, err := repository.NewProductionRepo(db).IsProductionFinishedBySaleBillUuid(saleBillUuid)
	if err != nil {
		return err
	}
	if !finished {
		return nil
	}
	return s.pushToPreparedQueue(companyUuid, member, time.Now().Unix())
}

// 获取叫号队列
// timeoutLimit: 超时限制（单位：分钟），0 表示不过滤
func (s *callBoardService) getCallBoardQueue(queueKey string, limit int64, minScore int64, timeoutLimit int) ([]string, error) {
	opt := &redis.ZRangeBy{
		Min:    strconv.FormatInt(minScore, 10),
		Max:    "+inf",
		Offset: 0,
		Count:  limit,
	}
	results, err := s.getRedisClient().ZRevRangeByScoreWithScores(context.Background(), queueKey, opt).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []string{}, nil // 队列不存在，返回空列表
		}
		return nil, err
	}

	// 当前时间（Unix 时间戳，单位：秒）
	now := time.Now().Unix()
	// 超时时间阈值（秒）
	timeoutThreshold := int64(timeoutLimit * 60)

	// 收集所有有效的队列成员
	queueMembers := make([]queueMember, 0, len(results))
	for _, result := range results {
		str, _ := result.Member.(string)
		mem, ok := parseQueueMember(str)
		if !ok {
			continue
		}

		// 如果 timeoutLimit > 0，进行超时过滤
		if timeoutLimit > 0 {
			// score 是订单完成时间（Unix 时间戳，单位：秒）
			completedTime := int64(result.Score)
			// 如果订单完成时间超过超时阈值，跳过该订单
			if now-completedTime > timeoutThreshold {
				continue
			}
		}

		queueMembers = append(queueMembers, mem)
	}

	// 按照 CreateTime 从小到大排序
	sort.Slice(queueMembers, func(i, j int) bool {
		return queueMembers[i].CreateTime < queueMembers[j].CreateTime
	})

	// 提取排序后的 SerialNo
	serialNoList := make([]string, 0, len(queueMembers))
	for _, member := range queueMembers {
		serialNoList = append(serialNoList, member.SerialNo)
	}
	return serialNoList, nil
}

// 添加到制作中队列
func (s *callBoardService) pushToPreparingQueue(companyUuid uint64, member string, score int64) error {
	ctx := context.Background()
	_, err := s.getRedisClient().TxPipelined(ctx, func(p redis.Pipeliner) error {
		p.ZAdd(ctx, cachekey.GetPreparingQueueKey(companyUuid), redis.Z{
			Score:  float64(score),
			Member: member,
		})
		p.ZRem(ctx, cachekey.GetPreparedQueueKey(companyUuid), member)
		return nil
	})
	return err
}

// 添加到制作完成队列
func (s *callBoardService) pushToPreparedQueue(companyUuid uint64, member string, score int64) error {
	ctx := context.Background()
	_, err := s.getRedisClient().TxPipelined(ctx, func(p redis.Pipeliner) error {
		p.ZRem(ctx, cachekey.GetPreparingQueueKey(companyUuid), member)
		p.ZAdd(ctx, cachekey.GetPreparedQueueKey(companyUuid), redis.Z{
			Score:  float64(score),
			Member: member,
		})
		return nil
	})
	return err
}

// 从多个队列中移除成员
func (s *callBoardService) removeMemberFromQueues(member string, queueKeyList ...string) error {
	ctx := context.Background()
	if len(queueKeyList) == 0 {
		return nil
	}
	if len(queueKeyList) == 1 {
		return s.getRedisClient().ZRem(ctx, queueKeyList[0], member).Err()
	}
	_, err := s.getRedisClient().TxPipelined(ctx, func(p redis.Pipeliner) error {
		for _, key := range queueKeyList {
			p.ZRem(ctx, key, member)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func formatQueueMember(uuid uint64, serialNo string, createTime int64) string {
	return fmt.Sprintf("%d:%s:%d", uuid, serialNo, createTime)
}

type queueMember struct {
	Uuid       uint64
	SerialNo   string
	CreateTime int64
}

func parseQueueMember(member string) (queueMember, bool) {
	list := strings.Split(member, ":")
	if len(list) != 3 {
		return queueMember{}, false
	}
	uuid, _ := strconv.ParseUint(list[0], 10, 64)
	serialNo := list[1]
	createTime, _ := strconv.ParseInt(list[2], 10, 64)
	return queueMember{
		Uuid:       uuid,
		SerialNo:   serialNo,
		CreateTime: createTime,
	}, true
}

func (s *callBoardService) removeExpireMemberFromQueues(queueKey string, maxScore int64) error {
	return s.getRedisClient().ZRemRangeByScore(context.Background(), queueKey, "-inf", strconv.FormatInt(maxScore, 10)).Err()
}
