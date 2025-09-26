package callboard

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"time"

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

type ICallBoardService interface {
	// 展示设备端接口
	GetBindCode(ctx context.Context, req req.GetBindCodeReq) (*resp.BindCodeResp, error)
	GetBindInfo(ctx context.Context, req req.GetBindInfoReq) (*resp.BindInfoResp, error)
	GetQueueData(ctx context.Context, companyUuid uint64, req req.GetQueueDataReq) (*resp.QueueDataResp, error)

	// 商家管理端接口
	BindDevice(ctx context.Context, companyUuid uint64, req req.BindDeviceReq) error
	GetDeviceList(ctx context.Context, companyUuid uint64, req req.GetDeviceListReq) (*resp.DeviceListResp, error)
	UnbindDevice(ctx context.Context, companyUuid uint64, uuid uint64) error
	BindedDeviceValidate(deviceId string) (companyUuid uint64, err error)
	UpdateBindInfo(ctx context.Context, companyUuid uint64, req req.UpdateBindInfoReq) error
}

func NewCallBoardService(dbm *database.DBManager, cache cache.Cache) ICallBoardService {
	srv := &callBoardService{
		dbm:                dbm,
		cache:              cache,
		setPendingScript:   redis.NewScript(LuaScriptSetDeviceBindCode),
		clearPendingScript: redis.NewScript(LuaScriptClearPendingDevice),
		bus:                event.NewSystemBus(),
	}
	srv.bus.SubscribeCallBoardChangeEvent(srv.handleCallBoardChangeEvent)
	srv.bus.SubscribeCallBoardLanguageChangeEvent(srv.handleCallBoardLanguageChangeEvent)
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

func (s *callBoardService) UpdateBindInfo(ctx context.Context, companyUuid uint64, req req.UpdateBindInfoReq) error {
	repo := repository.NewDeviceRepo(s.dbm.GetDB(companyUuid))
	device, err := repo.GetDevice(repo.WhereUuid(req.Uuid))
	if err != nil {
		return err
	}
	if device.Uuid == 0 {
		return errors.New("设备不存在")
	}
	err = s.getRedisClient().HMSet(ctx, cachekey.GetBindedDeviceKey(device.DeviceId), "lang1", req.Lang1, "lang2", req.Lang2).Err()
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
	bindInfo, err := s.mustGetBindInfoFromCache(req.DeviceId)
	if err != nil {
		return nil, err
	}
	updateTime, _ := s.getRedisClient().Get(ctx, cachekey.GetCallBoardChangeKey(companyUuid)).Int64()
	if updateTime <= req.UpdateTime && updateTime != 0 {
		// 数据没有变化，直接返回空数据
		return &resp.QueueDataResp{
			Lang1:          bindInfo.Lang1,
			Lang2:          bindInfo.Lang2,
			UpdateTime:     updateTime,
			PreparingQueue: make([]string, 0),
			PreparedQueue:  make([]string, 0),
		}, nil
	}

	productRepo := repository.NewProductRepo(s.dbm.GetDB(companyUuid))
	// 获取制作中的
	preparingList, err := productRepo.GetLatestProductsByStatus(req.Limit, constant.ProductionOrderProductStatusCooking)
	if err != nil {
		return nil, err
	}
	// 获取制作完成的
	preparedList, err := productRepo.GetLatestProductsByStatus(req.Limit, constant.ProductionOrderProductStatusFinished)
	if err != nil {
		return nil, err
	}

	// 查询所有涉及的销售账单
	saleBillUuids := make([]uint64, 0, len(preparingList)+len(preparedList))
	for _, prd := range preparingList {
		saleBillUuids = append(saleBillUuids, prd.SaleBillUuid)
	}
	for _, prd := range preparedList {
		saleBillUuids = append(saleBillUuids, prd.SaleBillUuid)
	}
	saleBillRepo := repository.NewSaleBillRepo(s.dbm.GetDB(companyUuid))
	saleBills, err := saleBillRepo.GetSaleBillList(func(db *gorm.DB) *gorm.DB {
		return db.Select("uuid,serial_no").Where("uuid IN (?)", saleBillUuids)
	})
	if err != nil {
		return nil, err
	}

	// 获取销售账单的流水号
	serialNoMap := make(map[uint64]string)
	for _, saleBill := range saleBills {
		serialNoMap[saleBill.Uuid] = saleBill.SerialNo
	}

	resp := resp.QueueDataResp{
		Lang1:          bindInfo.Lang1,
		Lang2:          bindInfo.Lang2,
		UpdateTime:     time.Now().Unix(),
		PreparingQueue: make([]string, 0, len(preparingList)),
		PreparedQueue:  make([]string, 0, len(preparedList)),
	}
	s.cache.Set(cachekey.GetCallBoardChangeKey(companyUuid), time.Now().Unix(), time.Minute*1)

	// 按照最初的逆序返回
	for i := len(preparingList) - 1; i >= 0; i-- {
		preparing := preparingList[i]
		resp.PreparingQueue = append(resp.PreparingQueue, serialNoMap[preparing.SaleBillUuid])
	}
	for i := len(preparedList) - 1; i >= 0; i-- {
		prepared := preparedList[i]
		resp.PreparedQueue = append(resp.PreparedQueue, serialNoMap[prepared.SaleBillUuid])
	}

	return &resp, nil
}

// BindDevice 绑定设备
func (s *callBoardService) BindDevice(ctx context.Context, companyUuid uint64, req req.BindDeviceReq) error {
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
	device, err := repo.GetDevice(repo.WhereSn(deviceId))
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
		bindInfo, _ := s.mustGetBindInfoFromCache(device.DeviceId)
		list = append(list, resp.DeviceItem{
			Uuid:     device.Uuid,
			DeviceId: device.DeviceId,
			BindTime: device.CreateTime,
			Lang1:    bindInfo.Lang1,
			Lang2:    bindInfo.Lang2,
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

func (s *callBoardService) handleCallBoardChangeEvent(msg event.CallBoardChangeEvent) {
	// 更新叫号展示变化信息
	s.cache.Set(cachekey.GetCallBoardChangeKey(msg.CompanyUuid), time.Now().Unix(), time.Minute*5)
}

type DeviceBindInfo struct {
	CompanyUuid  uint64 `redis:"company_uuid"`
	DeviceSecret string `redis:"device_secret"`
	Lang1        string `redis:"lang1"`
	Lang2        string `redis:"lang2"`
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

func (s *callBoardService) handleCallBoardLanguageChangeEvent(msg event.CallBoardLanguageChangeEvent) {
	// 从数据库读取
	settingRepo := repository.NewSettingRepo(s.dbm.GetDB(msg.CompanyUuid))
	settingData := settingRepo.GetByKey(constant.SettingStore)
	if settingData.Key == "" {
		logger.Logger.Error("商家设置不存在", zap.Uint64("companyUuid", msg.CompanyUuid))
		return
	}
	storeSetting := dtosetting.Store{}
	_ = json.Unmarshal([]byte(settingData.Values), &storeSetting)
	if len(storeSetting.Language) == 0 {
		logger.Logger.Error("商家语言信息为空", zap.Uint64("companyUuid", msg.CompanyUuid))
		return
	}
	devRepo := repository.NewDeviceRepo(s.dbm.GetDB(msg.CompanyUuid))
	devices, err := devRepo.GetDeviceList(devRepo.WhereSource("call_board"))
	if err != nil {
		logger.Logger.Error("获取设备列表失败", zap.Uint64("companyUuid", msg.CompanyUuid))
		return
	}
	for _, dev := range devices {
		bindInfo, err := s.mustGetBindInfoFromCache(dev.DeviceId)
		if err != nil {
			logger.Logger.Error("获取绑定信息失败", zap.Uint64("companyUuid", msg.CompanyUuid), zap.String("deviceId", dev.DeviceId))
			continue
		}
		lang1, lang2 := findToUpdateLangs(storeSetting.Language, bindInfo.Lang1, bindInfo.Lang2)
		if lang1 == bindInfo.Lang1 && lang2 == bindInfo.Lang2 {
			continue
		}
		err = s.getRedisClient().HMSet(
			context.Background(),
			cachekey.GetBindedDeviceKey(dev.DeviceId),
			"lang1", lang1,
			"lang2", lang2,
		).Err()
		if err != nil {
			logger.Logger.Error("更新叫号语言信息失败", zap.Uint64("companyUuid", msg.CompanyUuid), zap.String("deviceId", dev.DeviceId))
		}
	}
}

func findToUpdateLangs(langList []dto.LanguageItem, targetLang1 string, targetLang2 string) (lang1 string, lang2 string) {
	toUpdateLangs := make([]string, 0, 2)
	for _, lang := range langList {
		if len(toUpdateLangs) == 2 {
			break
		}
		if lang.Name == targetLang1 {
			toUpdateLangs = append(toUpdateLangs, lang.Name)
			continue
		}
		if lang.Name == targetLang2 {
			toUpdateLangs = append(toUpdateLangs, lang.Name)
			continue
		}
	}
	if len(toUpdateLangs) == 0 {
		toUpdateLangs = append(toUpdateLangs, langList[0].Name, "")
	}
	if len(toUpdateLangs) == 1 {
		toUpdateLangs = append(toUpdateLangs, "")
	}
	return toUpdateLangs[0], toUpdateLangs[1]
}
