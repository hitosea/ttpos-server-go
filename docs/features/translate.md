# Translate Service (多语言翻译服务) 详细说明

## 概述

`translate.go` 文件实现了多语言自动翻译服务，负责将系统中的英文名称自动翻译成其他 8 种语言（简体中文、繁体中文、泰语、日语、韩语、缅甸语、土耳其语、瑞典语）。该服务使用 Redis 集合管理待翻译队列，支持批量翻译、重试机制和并发控制，并在翻译完成后自动同步到相关业务模型。

## 文件位置
```
ttpos-server-go/main/app/service/translate.go
```

## 核心功能

### 1. 服务接口定义

#### ITranslateSrv 接口
定义了翻译服务的所有方法：

```go
type ITranslateSrv interface {
    Translate(companyUuid uint64) error                         // 翻译指定公司的待翻译内容
    TranslateAll() error                                        // 翻译所有公司的待翻译内容
    AddMultiLanguageNameUuidToSet(companyUuid, uuids...) error  // 添加多语言UUID到待翻译集合
    RemoveMultiLanguageNameUuidFromSet(companyUuid, uuids...) error // 从待翻译集合删除UUID
}
```

### 2. 支持的语言

系统支持从英文（源语言）翻译到以下 8 种目标语言：

| 语言代码 | 语言名称 | 字段名 |
|---------|---------|--------|
| `zh` | 简体中文 | `zh_name` |
| `zh-tw` | 繁体中文 | `zh_tw_name` |
| `th` | 泰语 | `th_name` |
| `ja` | 日语 | `ja_name` |
| `ko` | 韩语 | `ko_name` |
| `my` | 缅甸语 | `my_name` |
| `tr` | 土耳其语 | `tr_name` |
| `sv` | 瑞典语 | `sv_name` |

**源语言**：
- `en` (英语): `en_name`

### 3. 涉及的业务模型

翻译完成后会自动更新以下模型的 `name` 字段（JSON格式多语言字段）：

| 模型 | 表名 | 业务含义 |
|------|------|---------|
| `ProductUnit` | `product_unit` | 商品单位 |
| `Warehouse` | `warehouse` | 仓库 |
| `Material` | `material` | 物品（原材料）|
| `ProductFlavor` | `product_flavor` | 商品规格/口味 |
| `ProductAttribute` | `product_attribute` | 商品属性 |
| `ProductSauce` | `product_sauce` | 加料/酱料 |
| `ProductPackage` | `product_package` | 商品套餐 |
| `ProductBomCard` | `product_bom_card` | 成本卡（BOM卡）|

所有模型通过 `multi_language_name_uuid` 字段关联到 `multi_language_name` 表。

## 核心组件

### 1. TranslateTaskManager (翻译任务管理器)

#### 功能描述
全局单例管理器，防止同一公司同时执行多个翻译任务。

#### 数据结构
```go
type TranslateTaskManager struct {
    runningTasks sync.Map  // key: companyUuid (uint64), value: bool
}
```

#### 核心方法

**tryStartTask - 尝试启动任务**
```go
func (m *TranslateTaskManager) tryStartTask(companyUuid uint64) bool {
    _, loaded := m.runningTasks.LoadOrStore(companyUuid, true)
    return !loaded  // 如果之前没有值，返回true（成功启动）
}
```

**逻辑说明**：
- 使用 `LoadOrStore` 原子操作
- 如果 `companyUuid` 不存在，则存储并返回 `true`（启动成功）
- 如果 `companyUuid` 已存在，则返回 `false`（已有任务在运行）

**finishTask - 完成任务**
```go
func (m *TranslateTaskManager) finishTask(companyUuid uint64) {
    m.runningTasks.Delete(companyUuid)
}
```

**getRunningCompanyUuids - 获取运行中的公司列表**
```go
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
```

### 2. TranslateSrv (翻译服务)

#### 数据结构
```go
type TranslateSrv struct {
    dbm            *database.DBManager  // 数据库管理器
    cache          cache.Cache          // 缓存客户端（Redis）
    cacheKeyPrefix string               // 缓存键前缀："translate:company_uuid"
}
```

#### Redis 缓存键格式
```
translate:company_uuid:{companyUuid}
```

**示例**：
- `translate:company_uuid:1001` - 公司1001的待翻译集合
- `translate:company_uuid:1002` - 公司1002的待翻译集合

**数据类型**：Redis Set（集合），存储多语言名称的UUID列表

## 核心流程

### 1. 翻译流程 (`Translate`)

#### 功能描述
执行指定公司的翻译任务，从 Redis 获取待翻译的多语言UUID列表，批量翻译并更新数据库。

#### 核心流程

**步骤 1：并发控制检查**
```go
if !translateTaskManager.tryStartTask(companyUuid) {
    return errors.New("翻译中，请稍后再试")
}
```

**步骤 2：获取待翻译的UUID列表**
```go
cacheKey := fmt.Sprintf("%s:%d", s.cacheKeyPrefix, companyUuid)
// 从 Redis Set 获取所有成员
multiLanguageNameUuids, err := s.getRedisClient().SMembers(context.Background(), cacheKey).Result()
if len(multiLanguageNameUuids) == 0 {
    return nil  // 没有待翻译内容，直接返回
}
```

**步骤 3：异常恢复与状态清理**
```go
defer func() {
    if r := recover(); r != nil {
        logger.Logger.Error("翻译任务发生panic", 
            zap.Uint64("companyUuid", companyUuid), 
            zap.Any("panic", r))
    }
    translateTaskManager.finishTask(companyUuid)
    logger.Logger.Info("翻译完成", zap.Uint64("companyUuid", companyUuid))
}()
```

**步骤 4：解析并去重UUID**
```go
var uuidsInRedis []uint64
for _, multiLanguageNameUuid := range multiLanguageNameUuids {
    uuid, _ := strconv.ParseUint(multiLanguageNameUuid, 10, 64)
    if uuid != 0 && !slices.Contains(uuidsInRedis, uuid) {
        uuidsInRedis = append(uuidsInRedis, uuid)
    }
}
```

**步骤 5：批量查询并翻译**

使用 GORM 的 `FindInBatches` 分批处理，每批100条：

```go
db := s.dbm.GetDB(companyUuid)
translateClient := utils.NewTranslateClient()
translatedMultiLanguageMap := make(map[uint64]dto.LocaleResponse)
var multiLanguageNames []model.MultiLanguageName
var uuidsInDB []uint64

db.Model(&model.MultiLanguageName{}).
    Scopes(repository.NotDeleted).
    Where("uuid in (?) AND en_name != ''", multiLanguageNameUuids).
    FindInBatches(&multiLanguageNames, 100, func(tx *gorm.DB, batch int) error {
        // 5.1 准备翻译项
        var translateItems []utils.TranslateItem
        var uuidList []uint64
        
        for _, multiLanguageName := range multiLanguageNames {
            translateItems = append(translateItems, utils.TranslateItem{
                Lang:    "en",
                Content: multiLanguageName.EnName,
            })
            uuidList = append(uuidList, multiLanguageName.Uuid)
            uuidsInDB = append(uuidsInDB, multiLanguageName.Uuid)
        }
        
        logger.Logger.Info("Translate-TranslateWithRetry", 
            zap.Any("translateItems", translateItems), 
            zap.Any("uuidList", uuidList))
        
        // 5.2 调用翻译服务，每次翻译20条
        multiLanguageMap := translateClient.TranslateWithRetry(
            context.Background(), 
            translateItems, 
            20)
        
        logger.Logger.Info("Translate-TranslateWithRetry-multiLanguageMap", 
            zap.Any("multiLanguageMap", multiLanguageMap))
        
        // 5.3 重新获取集合，判断是否已被【手动】翻译
        multiLanguageNameUuids, err = s.getRedisClient().
            SMembers(context.Background(), cacheKey).Result()
        if err != nil {
            logger.Logger.Error("Translate-Redis-GetMultiLanguageNameUuids", 
                zap.Any("err", err))
            return nil
        }
        
        // 5.4 更新翻译结果
        for _, multiLanguageName := range multiLanguageNames {
            translated, ok := multiLanguageMap[multiLanguageName.EnName]
            // 未翻译成功 或者 已从集合中移除（手动翻译），则跳过
            if !ok || !slices.Contains(multiLanguageNameUuids, 
                strconv.FormatUint(multiLanguageName.Uuid, 10)) {
                continue
            }
            
            // 更新多语言名称表
            tx.Model(&model.MultiLanguageName{}).
                Where("uuid = ?", multiLanguageName.Uuid).
                Updates(map[string]any{
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
            
            // 5.5 从待翻译集合中移除
            if err := s.RemoveMultiLanguageNameUuidFromSet(
                companyUuid, multiLanguageName.Uuid); err != nil {
                logger.Logger.Error("Translate-Redis-RemoveMultiLanguageNameUuidFromSet", 
                    zap.Uint64("multiLanguageNameUuid", multiLanguageName.Uuid), 
                    zap.Any("err", err))
            } else {
                translatedMultiLanguageMap[multiLanguageName.Uuid] = translated
            }
        }
        return nil
    })
```

**步骤 6：清理无效的UUID**
```go
// 不在DB中的多语言UUID，需要从集合中移除
uuidsNotInDB := slice.Difference(uuidsInRedis, uuidsInDB)
if len(uuidsNotInDB) > 0 {
    logger.Logger.Info("无效的多语言uuid", zap.Any("uuidsNotInDB", uuidsNotInDB))
    s.RemoveMultiLanguageNameUuidFromSet(companyUuid, uuidsNotInDB...)
}
```

**步骤 7：更新业务模型的name字段**
```go
if len(translatedMultiLanguageMap) > 0 {
    for uuid, translated := range translatedMultiLanguageMap {
        translatedText := translated.ToJson()  // 转换为JSON格式
        
        // 更新各业务表
        db.Model(&model.ProductUnit{}).
            Where("multi_language_name_uuid = ?", uuid).
            Update("name", translatedText)
        
        db.Model(&model.Warehouse{}).
            Where("multi_language_name_uuid = ?", uuid).
            Update("name", translatedText)
        
        db.Model(&model.Material{}).
            Where("multi_language_name_uuid = ?", uuid).
            Update("name", translatedText)
        
        db.Model(&model.ProductFlavor{}).
            Where("multi_language_name_uuid = ?", uuid).
            Update("name", translatedText)
        
        db.Model(&model.ProductAttribute{}).
            Where("multi_language_name_uuid = ?", uuid).
            Update("name", translatedText)
        
        db.Model(&model.ProductSauce{}).
            Where("multi_language_name_uuid = ?", uuid).
            Update("name", translatedText)
        
        db.Model(&model.ProductPackage{}).
            Where("multi_language_name_uuid = ?", uuid).
            Update("name", translatedText)
        
        db.Model(&model.ProductBomCard{}).
            Where("multi_language_name_uuid = ?", uuid).
            Update("name", translatedText)
    }
}
```

#### 批量处理策略

**分批查询**：
- 每批100条记录（`FindInBatches` 参数）
- 避免一次性加载大量数据

**分批翻译**：
- 每次翻译20条（`TranslateWithRetry` 参数）
- 降低单次请求的翻译量
- 提高成功率

### 2. 添加待翻译UUID (`AddMultiLanguageNameUuidToSet`)

#### 功能描述
将多语言UUID添加到待翻译集合，并自动触发翻译任务。

#### 核心流程

**步骤 1：参数校验**
```go
if len(multiLanguageNameUuids) == 0 {
    return nil
}
```

**步骤 2：构建缓存键**
```go
cacheKey := fmt.Sprintf("%s:%d", s.cacheKeyPrefix, companyUuid)
```

**步骤 3：转换数据类型并添加到Redis Set**
```go
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
```

**步骤 4：异步触发翻译**
```go
utils.Go(func() {
    func(companyUuid uint64) {
        if err := s.Translate(companyUuid); err != nil {
            logger.Logger.Error("Translate-Redis-Translate", zap.Any("err", err))
        }
    }(companyUuid)
})
```

**特点**：
- 使用 Redis Set 自动去重
- 添加后立即异步触发翻译
- 不阻塞调用方

### 3. 移除待翻译UUID (`RemoveMultiLanguageNameUuidFromSet`)

#### 功能描述
从待翻译集合中移除指定的UUID，用于以下场景：
1. 翻译完成后移除
2. 清理无效的UUID
3. 用户手动翻译后移除

#### 核心流程

**步骤 1：构建缓存键**
```go
cacheKey := fmt.Sprintf("%s:%d", s.cacheKeyPrefix, companyUuid)
```

**步骤 2：转换数据类型并从Redis Set移除**
```go
// 将 []uint64 转换为 []interface{}
members := make([]any, len(multiLanguageNameUuids))
for i, uuid := range multiLanguageNameUuids {
    members[i] = uuid
}

return s.getRedisClient().SRem(context.Background(), cacheKey, members...).Err()
```

### 4. 翻译所有公司 (`TranslateAll`)

#### 功能描述
扫描 Redis 中所有的待翻译集合，为每个有待翻译内容的公司异步触发翻译任务。

#### 核心流程

**步骤 1：扫描所有相关的Redis键**
```go
// 使用封装方法，支持集群和单机模式
keys, err := cache.ScanRedisKeysDefault(
    context.Background(), 
    s.getRedisClient(), 
    s.cacheKeyPrefix+"*")  // "translate:company_uuid*"
if err != nil {
    logger.Logger.Error("Translate-Redis-Keys", zap.Any("err", err))
    return err
}
```

**步骤 2：去重并解析公司UUID**
```go
var uniqueKeys []string
for _, key := range keys {
    if slices.Contains(uniqueKeys, key) {
        continue
    } else {
        uniqueKeys = append(uniqueKeys, key)
    }
    
    // 从 "translate:company_uuid:1001" 中解析出 1001
    companyUuid, err := strconv.ParseUint(
        strings.TrimPrefix(key, s.cacheKeyPrefix+":"), 
        10, 64)
    if err != nil {
        logger.Logger.Error("Translate-Redis-SMembers", 
            zap.Any("err", err), 
            zap.String("key", key))
        continue
    }
```

**步骤 3：检查并触发翻译**
```go
    // 获取待翻译UUID列表
    multiLanguageNameUuids, err := s.getRedisClient().
        SMembers(context.Background(), key).Result()
    if err != nil {
        logger.Logger.Error("Translate-Redis-Translate", 
            zap.Any("err", err), 
            zap.String("key", key))
        continue
    }
    
    if len(multiLanguageNameUuids) > 0 {
        // 异步执行翻译
        utils.Go(func() {
            if err := s.Translate(companyUuid); err != nil {
                logger.Logger.Info("Translate-Redis-Translate-go", 
                    zap.Uint64("companyUuid", companyUuid), 
                    zap.Any("err", err), 
                    zap.String("key", key))
            }
        })
    } else {
        logger.Logger.Info("Translate-Redis-Translate-no-data", 
            zap.Uint64("companyUuid", companyUuid), 
            zap.String("key", key))
    }
}
```

**特点**：
- 支持 Redis 集群和单机模式
- 并发处理多个公司的翻译任务
- 跳过空集合

### 5. 获取Redis客户端 (`getRedisClient`)

#### 功能描述
兼容 Redis 集群和单机模式，返回对应的客户端。

```go
func (s *TranslateSrv) getRedisClient() redis.UniversalClient {
    clusterClient := s.cache.GetClusterClient()
    if clusterClient != nil {
        return clusterClient
    }
    return s.cache.GetClient()
}
```

**逻辑**：
- 优先使用集群客户端
- 集群客户端不存在时使用单机客户端

## 数据模型

### 1. model.MultiLanguageName (多语言名称表)

主表，存储所有多语言翻译结果：

```go
type MultiLanguageName struct {
    BaseModel            // Uuid, CreateTime, UpdateTime, DeleteTime
    EnName    string    // 英文名称（源语言）
    ZhName    string    // 简体中文
    ThName    string    // 泰语
    ZhTwName  string    // 繁体中文
    JaName    string    // 日语
    KoName    string    // 韩语
    MyName    string    // 缅甸语
    TrName    string    // 土耳其语
    SvName    string    // 瑞典语
}
```

### 2. 业务模型关联

所有业务模型通过 `multi_language_name_uuid` 字段关联：

```go
type ProductUnit struct {
    // ... 其他字段
    MultiLanguageNameUuid uint64  // 多语言UUID
    Name                  string  // JSON格式的多语言名称
}

type Warehouse struct {
    // ... 其他字段
    MultiLanguageNameUuid uint64
    Name                  string
}

type Material struct {
    // ... 其他字段
    MultiLanguageNameUuid uint64
    Name                  string
}

// ... 其他模型类似
```

### 3. 多语言JSON格式

业务模型的 `name` 字段存储JSON格式的多语言数据：

```json
{
    "en": "Apple",
    "zh": "苹果",
    "th": "แอปเปิล",
    "zh-tw": "蘋果",
    "ja": "りんご",
    "ko": "사과",
    "my": "ပန်းသီး",
    "tr": "Elma",
    "sv": "Äpple"
}
```

**转换方法**：
```go
translatedText := translated.ToJson()
```

## 翻译客户端

### utils.TranslateClient

#### 数据结构

**TranslateItem - 翻译项**：
```go
type TranslateItem struct {
    Lang    string  // 源语言代码（"en"）
    Content string  // 待翻译内容
}
```

#### 核心方法

**TranslateWithRetry - 带重试的翻译**：
```go
func (c *TranslateClient) TranslateWithRetry(
    ctx context.Context, 
    items []TranslateItem, 
    batchSize int) map[string]dto.LocaleResponse
```

**参数说明**：
- `items`: 待翻译项列表
- `batchSize`: 每批翻译数量（建议20）

**返回值**：
- `map[string]dto.LocaleResponse`: 键为英文原文，值为翻译结果

**dto.LocaleResponse 结构**：
```go
type LocaleResponse struct {
    EN   string  // 英文
    ZH   string  // 简体中文
    TH   string  // 泰语
    ZHTW string  // 繁体中文
    JA   string  // 日语
    KO   string  // 韩语
    MY   string  // 缅甸语
    TR   string  // 土耳其语
    SV   string  // 瑞典语
}

func (l *LocaleResponse) ToJson() string {
    // 返回JSON字符串
}
```

## 业务规则

### 1. 并发控制

**同一公司只能同时运行一个翻译任务**：
```go
if !translateTaskManager.tryStartTask(companyUuid) {
    return errors.New("翻译中，请稍后再试")
}
```

**原因**：
- 防止重复翻译
- 避免资源浪费
- 确保数据一致性

### 2. 翻译条件

**仅翻译满足以下条件的记录**：
```sql
WHERE uuid IN (?) 
  AND en_name != ''           -- 英文名称不为空
  AND delete_time = 0         -- 未删除
```

**原因**：
- `en_name` 为空无法翻译
- 已删除的记录无需翻译

### 3. 手动翻译优先

在批处理过程中，每批翻译前都会重新检查 Redis 集合：

```go
// 重新获取集合，用于判断是否已经【手动】翻译过
multiLanguageNameUuids, err = s.getRedisClient().
    SMembers(context.Background(), cacheKey).Result()

// 如果不在集合中，说明已被手动翻译，跳过
if !slices.Contains(multiLanguageNameUuids, 
    strconv.FormatUint(multiLanguageName.Uuid, 10)) {
    continue
}
```

**场景**：
- 自动翻译进行中
- 用户手动修改了翻译
- 手动修改后从集合中移除UUID
- 自动翻译检测到并跳过该条

### 4. 失败处理策略

**翻译失败的项不会被移除**：
```go
translated, ok := multiLanguageMap[multiLanguageName.EnName]
// 未翻译成功，跳过
if !ok {
    continue
}
```

**优点**：
- 失败的项保留在集合中
- 下次翻译时会重试
- 不会丢失待翻译内容

### 5. 数据同步规则

**翻译完成后同步两个地方**：
1. **多语言名称表** (`multi_language_name`)
   - 更新所有语言字段
   
2. **业务模型表** (8个业务表)
   - 更新 `name` 字段（JSON格式）

**同步时机**：
- 批处理完成后统一同步
- 避免频繁更新业务表

## 性能优化

### 1. 批量处理

**查询批量化**：
```go
db.Model(&model.MultiLanguageName{}).
    // ...
    FindInBatches(&multiLanguageNames, 100, func(tx *gorm.DB, batch int) error {
        // 每批100条
    })
```

**翻译批量化**：
```go
multiLanguageMap := translateClient.TranslateWithRetry(
    context.Background(), 
    translateItems, 
    20)  // 每次翻译20条
```

**原因**：
- 减少数据库查询次数
- 降低翻译API压力
- 提高整体效率

### 2. 异步执行

**添加后异步翻译**：
```go
utils.Go(func() {
    s.Translate(companyUuid)
})
```

**全量翻译异步执行**：
```go
utils.Go(func() {
    if err := s.Translate(companyUuid); err != nil {
        // 错误处理
    }
})
```

**优点**：
- 不阻塞主流程
- 提升用户体验
- 并发处理多个公司

### 3. Redis Set去重

使用 Redis Set 存储待翻译UUID：
```go
s.getRedisClient().SAdd(context.Background(), cacheKey, members...)
```

**优点**：
- 自动去重
- O(1)时间复杂度
- 支持并发添加

### 4. 仅更新已翻译的业务模型

```go
if len(translatedMultiLanguageMap) > 0 {
    for uuid, translated := range translatedMultiLanguageMap {
        // 仅更新成功翻译的
        db.Model(&model.ProductUnit{}).
            Where("multi_language_name_uuid = ?", uuid).
            Update("name", translatedText)
        // ...
    }
}
```

**优点**：
- 避免无效更新
- 减少数据库写入
- 提高性能

## 异常处理

### 1. Panic 恢复

```go
defer func() {
    if r := recover(); r != nil {
        logger.Logger.Error("翻译任务发生panic", 
            zap.Uint64("companyUuid", companyUuid), 
            zap.Any("panic", r))
    }
    translateTaskManager.finishTask(companyUuid)
    logger.Logger.Info("翻译完成", zap.Uint64("companyUuid", companyUuid))
}()
```

**特点**：
- 捕获所有 panic
- 记录错误日志
- 确保释放并发锁

### 2. 错误日志

**关键节点日志**：
```go
logger.Logger.Info("开始翻译任务", 
    zap.Uint64("companyUuid", companyUuid), 
    zap.Any("multiLanguageNameUuids", multiLanguageNameUuids))

logger.Logger.Info("Translate-TranslateWithRetry", 
    zap.Any("translateItems", translateItems), 
    zap.Any("uuidList", uuidList))

logger.Logger.Info("Translate-TranslateWithRetry-multiLanguageMap", 
    zap.Any("multiLanguageMap", multiLanguageMap))

logger.Logger.Error("添加多语言uuid到待翻译集合失败", zap.Error(err))
```

**日志命名规范**：
- 使用 `Translate-` 前缀
- 分层命名：`Translate-Redis-Action`
- 包含关键参数

### 3. 无效数据清理

```go
// 不在DB中的多语言UUID，从集合中移除
uuidsNotInDB := slice.Difference(uuidsInRedis, uuidsInDB)
if len(uuidsNotInDB) > 0 {
    logger.Logger.Info("无效的多语言uuid", zap.Any("uuidsNotInDB", uuidsNotInDB))
    s.RemoveMultiLanguageNameUuidFromSet(companyUuid, uuidsNotInDB...)
}
```

**原因**：
- UUID在Redis中但数据库中不存在
- 可能是数据已被删除
- 清理避免无效翻译

## 使用场景

### 场景 1：新增商品单位

```go
// 1. 创建多语言名称记录
multiLanguageName := &model.MultiLanguageName{
    EnName: "Kilogram",
}
db.Create(multiLanguageName)

// 2. 创建商品单位
productUnit := &model.ProductUnit{
    MultiLanguageNameUuid: multiLanguageName.Uuid,
    // ... 其他字段
}
db.Create(productUnit)

// 3. 添加到待翻译集合（自动触发翻译）
translateSrv.AddMultiLanguageNameUuidToSet(companyUuid, multiLanguageName.Uuid)
```

**结果**：
- 自动翻译 "Kilogram" 到8种语言
- 更新 `multi_language_name` 表
- 更新 `product_unit` 表的 `name` 字段

### 场景 2：批量导入商品

```go
var multiLanguageUuids []uint64

// 批量创建商品和多语言记录
for _, product := range importProducts {
    multiLanguageName := &model.MultiLanguageName{
        EnName: product.EnName,
    }
    db.Create(multiLanguageName)
    multiLanguageUuids = append(multiLanguageUuids, multiLanguageName.Uuid)
    
    // 创建商品...
}

// 批量添加到待翻译集合
translateSrv.AddMultiLanguageNameUuidToSet(companyUuid, multiLanguageUuids...)
```

**优点**：
- 批量添加到 Redis Set
- 一次翻译处理所有
- 效率高

### 场景 3：定时任务触发全量翻译

```go
// 定时任务（如每小时执行一次）
func cronTranslateAll() {
    translateSrv := NewTranslateSrv(dbm, cache)
    if err := translateSrv.TranslateAll(); err != nil {
        logger.Logger.Error("定时翻译失败", zap.Error(err))
    }
}
```

**作用**：
- 处理积压的翻译任务
- 重试之前失败的翻译
- 保证数据完整性

### 场景 4：手动翻译后清理

```go
// 用户手动修改了翻译
multiLanguageName := &model.MultiLanguageName{
    Uuid:     1001,
    EnName:   "Apple",
    ZhName:   "苹果（自定义）",
    ThName:   "แอปเปิล（自定义）",
    // ... 其他语言
}
db.Model(&model.MultiLanguageName{}).
    Where("uuid = ?", multiLanguageName.Uuid).
    Updates(multiLanguageName)

// 从待翻译集合中移除，避免被自动翻译覆盖
translateSrv.RemoveMultiLanguageNameUuidFromSet(companyUuid, multiLanguageName.Uuid)
```

## 数据流图

### 完整翻译流程
```
业务操作（创建/更新）
    ↓
创建多语言记录（en_name）
    ↓
添加UUID到Redis Set
    ↓
自动触发翻译任务
    ↓
【翻译任务开始】
    ↓
并发控制检查
    ↓
从Redis获取待翻译UUID列表
    ↓
分批查询数据库（100条/批）
    ↓
【每批处理】
    ├─ 准备翻译项（英文内容）
    ├─ 调用翻译API（20条/次）
    ├─ 重新检查Redis（防止手动翻译冲突）
    ├─ 更新多语言表（9种语言）
    ├─ 从Redis移除已翻译UUID
    └─ 记录翻译结果
【批处理结束】
    ↓
清理无效UUID
    ↓
更新业务表name字段（8个表）
    ↓
释放并发锁
    ↓
【翻译任务结束】
```

### Redis Set操作流程
```
【添加操作】
业务创建 → 生成UUID → SAdd到Redis → 触发翻译
                              ↓
                         Set自动去重

【移除操作】
翻译成功 → SRem从Redis
手动翻译 → SRem从Redis
无效UUID → SRem从Redis

【查询操作】
翻译任务 → SMembers获取列表 → 处理 → 逐个SRem
```

## 监控与优化建议

### 1. 监控指标

**待翻译队列长度**：
```go
count := s.getRedisClient().SCard(context.Background(), cacheKey).Val()
```

**正在执行的翻译任务数**：
```go
runningCount := len(translateTaskManager.getRunningCompanyUuids())
```

**翻译成功率**：
```go
successRate := len(translatedMultiLanguageMap) / len(multiLanguageNames)
```

### 2. 告警建议

- 待翻译队列长度 > 1000（积压过多）
- 翻译任务执行时间 > 5分钟（可能卡住）
- 翻译成功率 < 80%（翻译服务异常）
- 同一公司频繁触发翻译（可能存在问题）

### 3. 优化建议

**减少翻译频率**：
```go
// 不要每次修改都立即翻译
// 可以批量收集后统一翻译
var uuidsToTranslate []uint64
for _, item := range items {
    uuidsToTranslate = append(uuidsToTranslate, item.MultiLanguageUuid)
}
translateSrv.AddMultiLanguageNameUuidToSet(companyUuid, uuidsToTranslate...)
```

**增加缓存**：
- 缓存翻译结果
- 相同内容不重复翻译

**监控翻译质量**：
- 记录翻译前后对比
- 人工抽查翻译质量
- 收集用户反馈

## 依赖关系

### 外部依赖

1. **utils.TranslateClient** (翻译客户端):
   - `NewTranslateClient()`: 创建翻译客户端
   - `TranslateWithRetry()`: 带重试的批量翻译

2. **cache.Cache** (Redis缓存):
   - `GetClient()`: 获取单机客户端
   - `GetClusterClient()`: 获取集群客户端
   - `ScanRedisKeysDefault()`: 扫描Redis键

### Repository 依赖

使用原生 GORM 操作，未使用 Repository 层。

### 数据库依赖

- 使用 `dbm.GetDB(companyUuid)` 获取公司数据库
- 操作 `multi_language_name` 表和8个业务表

## 总结

`translate.go` 实现了一个智能的多语言自动翻译服务，主要特点包括：

1. **自动翻译**：从英文自动翻译到8种语言
2. **队列管理**：使用Redis Set管理待翻译队列，支持去重
3. **并发控制**：全局管理器防止重复执行
4. **批量处理**：分批查询、分批翻译，提高效率
5. **手动优先**：支持手动翻译，自动翻译会避让
6. **异步执行**：不阻塞主流程，提升用户体验
7. **失败重试**：失败的项保留在队列中，支持重试
8. **数据同步**：翻译完成后自动同步到业务表
9. **健壮性强**：Panic恢复、错误日志、无效数据清理

该服务是系统国际化的关键组件，确保所有用户都能使用母语查看商品、物品、单位等核心数据，极大提升了系统的易用性和用户体验。

