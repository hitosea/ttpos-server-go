# Sync Service (数据同步服务) 详细说明

## 概述

`sync.go` 文件实现了数据同步服务，负责从 ERP 系统批量同步各类基础数据到本地数据库。该服务支持 15 种不同类型的数据同步，包括商品、物品、供应商、仓库、库存等核心业务数据。提供完整的任务管理、并发控制、错误处理和重试机制。

## 文件位置
```
ttpos-server-go/main/app/service/sync.go
```

## 核心功能

### 1. 服务接口定义

#### ISyncSrv 接口
定义了同步服务的所有方法：

```go
type ISyncSrv interface {
    Sync(ctx, syncReq)          // 同步（创建并执行同步任务）
    GetTaskList(ctx, listReq)   // 获取同步任务列表
    GetTaskDetail(ctx, detailReq) // 获取同步任务详情
}
```

### 2. 同步任务类型

系统支持 15 种同步任务类型，按执行顺序依次为：

| 序号 | 任务类型 | 常量名 | 任务名称 | 执行方法 |
|------|---------|--------|---------|---------|
| 1 | 商品分类 | `SyncTaskTypeProductCategory` | 商品分类同步 | `productSrv.SyncProductShopCategory` |
| 2 | 物品分类 | `SyncTaskTypeMaterialCategory` | 物品分类同步 | `materialSrv.SyncMaterialCategory` |
| 3 | 税率 | `SyncTaskTypeTax` | 税率同步 | `productSrv.SyncProductTax` |
| 4 | 单位 | `SyncTaskTypeUnit` | 单位同步 | `productSrv.SyncUnit` |
| 5 | 物品 | `SyncTaskTypeMaterial` | 物品同步 | `materialSrv.SyncMaterial` |
| 6 | 仓库 | `SyncTaskTypeWarehouse` | 仓库同步 | `warehouseSrv.SyncWarehouse` |
| 7 | 规格 | `SyncTaskTypeFlavor` | 规格同步 | `productSrv.SyncProductFlavor` |
| 8 | 属性 | `SyncTaskTypeAttribute` | 属性同步 | `productSrv.SyncAttributeGroup` |
| 9 | 加料 | `SyncTaskTypeSauce` | 加料同步 | `productSrv.SyncSauce` |
| 10 | 商品 | `SyncTaskTypeProduct` | 商品同步 | `productSrv.SyncProduct` |
| 11 | BOM卡 | `SyncTaskTypeBomCard` | BOM卡同步 | `materialSrv.SyncProductBomCard` |
| 12 | 供应商 | `SyncTaskTypeSupplier` | 供应商同步 | `supplierSrv.SyncSupplier` |
| 13 | 仓库库存 | `SyncTaskTypeWarehouseStock` | 仓库库存同步 | `warehouseSrv.SyncWarehouseItemStock` |
| 14 | 商品库存 | `SyncTaskTypeProductStock` | 商品库存同步 | `productSrv.SyncProductStockByBomCard` |
| 15 | 套餐图片 | `SyncTaskTypePackageImage` | 套餐图片同步 | `productSrv.SyncProductPackageImage` |

**执行顺序说明**：
- 基础数据优先（分类、税率、单位）
- 主数据其次（物品、仓库、商品）
- 关联数据最后（BOM卡、供应商、库存）
- 顺序执行，确保依赖关系正确

### 3. 同步任务配置 (`syncTaskConfig`)

每个同步任务的配置结构：

```go
type syncTaskConfig struct {
    TaskType string                      // 任务类型（常量）
    TaskName string                      // 任务名称（显示用）
    Executor func(context.Context) error // 执行函数
}
```

**配置示例**：
```go
{
    TaskType: constant.SyncTaskTypeProduct,
    TaskName: constant.SyncTaskTypeNames[constant.SyncTaskTypeProduct],
    Executor: s.productSrv.SyncProduct,
}
```

## 核心组件

### 1. SyncTaskManager (同步任务管理器)

#### 功能描述
全局单例管理器，防止同一公司同时执行多个同步任务。

#### 数据结构
```go
type SyncTaskManager struct {
    runningTasks sync.Map  // key: companyUuid (uint64), value: bool
}
```

#### 核心方法

**tryStartTask - 尝试启动任务**
```go
func (m *SyncTaskManager) tryStartTask(companyUuid uint64) bool {
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
func (m *SyncTaskManager) finishTask(companyUuid uint64) {
    m.runningTasks.Delete(companyUuid)
}
```

**getRunningCompanyUuids - 获取运行中的公司列表**
```go
func (m *SyncTaskManager) getRunningCompanyUuids() []uint64 {
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

### 2. SyncSrv (同步服务)

#### 数据结构
```go
type SyncSrv struct {
    dbm          *database.DBManager  // 数据库管理器
    warehouseSrv IWarehouseSrv        // 仓库服务
    materialSrv  IMaterialSrv         // 物品服务
    supplierSrv  ISupplierSrv         // 供应商服务
    productSrv   IProductSrv          // 商品服务
}
```

#### 依赖服务
- **WarehouseSrv**: 仓库、库存同步
- **MaterialSrv**: 物品、物品分类、BOM卡同步
- **SupplierSrv**: 供应商同步
- **ProductSrv**: 商品、商品分类、税率、单位、规格、属性、加料、套餐图片同步

## 核心流程

### 1. 同步流程 (`Sync`)

#### 功能描述
创建并执行同步任务，支持全量同步和失败重试两种模式。

#### 请求参数 (`req.SyncReq`)
- `TaskUuid`: 任务UUID（重试模式时传入）
- `IsSyncExecute`: 是否同步执行（true-同步，false-异步）

#### 执行流程

**步骤 1：并发控制检查**
```go
if !syncTaskManager.tryStartTask(companyUuid) {
    return resp.SyncResp{}, errors.New("数据同步中，请稍后再试")
}
```

**步骤 2：判断执行模式**

**模式A：重试模式（TaskUuid > 0）**
```go
if syncReq.TaskUuid > 0 {
    retryMode = true
    // 查询原任务
    existTask, err := syncTaskRepo.GetByUuid(syncReq.TaskUuid, 
        syncTaskRepo.PreloadItems())
    
    // 获取失败的任务类型
    for _, item := range existTask.Items {
        if item.Status == constant.SyncTaskItemStatusFailed {
            retryTaskTypes = append(retryTaskTypes, item.TaskType)
        }
    }
    
    if len(retryTaskTypes) == 0 {
        return resp.SyncResp{}, errors.New("没有需要重试的任务")
    }
    
    syncTask = existTask
}
```

**模式B：全量同步（TaskUuid = 0）**
```go
else {
    // 创建新的同步任务
    syncTask = &model.SyncTask{
        Status:       constant.SyncTaskStatusRunning,
        TotalCount:   uint32(len(allTasks)),  // 15个任务
        SuccessCount: 0,
        FailCount:    0,
        StartTime:    time.Now().Unix(),
    }
    
    syncTaskRepo.Create(syncTask)
}
```

**步骤 3：启动同步任务**

**异步执行（默认）**：
```go
if !syncReq.IsSyncExecute {
    utils.Go(func() {
        s.executeSync(ctx, syncTask, allTasks, retryMode, retryTaskTypes)
    })
}
```

**同步执行**：
```go
else {
    s.executeSync(ctx, syncTask, allTasks, retryMode, retryTaskTypes)
}
```

**步骤 4：返回响应**
```go
return resp.SyncResp{
    TaskUuid: syncTask.Uuid,
    Message:  message,  // "数据同步已启动" 或 "重试同步任务已启动"
}, nil
```

### 2. 执行同步 (`executeSync`)

#### 功能描述
实际执行同步任务的核心方法，包含完整的错误处理和状态管理。

#### 核心流程

**步骤 1：异常恢复机制**
```go
defer func() {
    var isPanicOccurred bool
    if r := recover(); r != nil {
        // 获取堆栈
        stack := string(debug.Stack())
        logger.Logger.Error("同步任务发生panic", 
            zap.Uint64("companyUuid", companyUuid), 
            zap.Any("panic", r), 
            zap.String("stack", stack))
        
        // 更新任务状态为失败
        syncTaskRepo.Update(syncTask.Uuid, map[string]any{
            "status":   constant.SyncTaskStatusFailed,
            "panic":    fmt.Sprintf("%v: %s", r, stack),
            "end_time": time.Now().Unix(),
        })
        isPanicOccurred = true
    }
    
    // 清理任务状态
    syncTaskManager.finishTask(companyUuid)
    
    // 记录最后同步时间（仅成功时）
    isExceptionOccurred := failCount > 0 || isPanicOccurred
    if !isExceptionOccurred {
        lastSyncTime := time.Now().Unix()
        // 更新公司表的最后同步时间
        s.dbm.GetDB(companyUuid).Model(&model.Company{}).
            Where("uuid = ?", companyUuid).
            Update("last_sync_time", lastSyncTime)
        s.dbm.GetDB(constant.DefaultDB).Model(&model.Company{}).
            Where("uuid = ?", companyUuid).
            Update("last_sync_time", lastSyncTime)
    }
    
    // 推送 WebSocket 通知
    utils.Go(func() {
        websocket.PushClient(company.Uuid, 
            websocket.SourceShop, 
            websocket.SourceAll, 
            websocket.SYNC_DATA, 
            map[string]any{
                "task_uuid":             syncTask.Uuid,
                "is_exception_occurred": isExceptionOccurred,
                "sync_time":             time.Now().Unix(),
            })
    })
}()
```

**步骤 2：筛选待执行任务**

**重试模式**：
```go
if retryMode {
    tasksToExecute = []syncTaskConfig{}
    for _, task := range allTasks {
        if slices.Contains(retryTaskTypes, task.TaskType) {
            tasksToExecute = append(tasksToExecute, task)
        }
    }
}
```

**全量模式**：
```go
else {
    tasksToExecute = allTasks  // 所有15个任务
}
```

**步骤 3：顺序执行子任务**
```go
for _, taskCfg := range tasksToExecute {
    s.executeSyncTask(ctx, syncTask.Uuid, taskCfg, 
        &successCount, &failCount, retryMode)
}
```

**步骤 4：更新主任务状态**
```go
endTime := time.Now().Unix()
finalStatus := constant.SyncTaskStatusSuccess
if failCount > 0 {
    finalStatus = constant.SyncTaskStatusFailed
}

syncTaskRepo.Update(syncTask.Uuid, map[string]any{
    "status":        finalStatus,
    "success_count": successCount,
    "fail_count":    failCount,
    "end_time":      endTime,
})
```

**步骤 5：更新公司最后同步时间**
```go
if finalStatus == constant.SyncTaskStatusSuccess {
    s.dbm.GetDB(companyUuid).Model(&model.Company{}).
        Where("uuid = ?", companyUuid).
        Update("last_sync_time", endTime)
    s.dbm.GetDB(constant.DefaultDB).Model(&model.Company{}).
        Where("uuid = ?", companyUuid).
        Update("last_sync_time", endTime)
}
```

### 3. 执行单个同步任务 (`executeSyncTask`)

#### 功能描述
执行单个子任务，处理任务项的创建、更新、状态追踪。

#### 核心流程

**步骤 1：处理任务项（重试模式）**
```go
var taskItem *model.SyncTaskItem

if retryMode {
    // 查找已存在的任务项
    items, err := syncTaskItemRepo.GetList(
        syncTaskItemRepo.WhereSyncTaskUuid(syncTaskUuid),
        syncTaskItemRepo.WhereTaskType(taskCfg.TaskType),
    )
    if err == nil && len(items) > 0 {
        taskItem = &items[0]
        // 重置任务项状态
        syncTaskItemRepo.Update(taskItem.Uuid, map[string]any{
            "status":        constant.SyncTaskItemStatusRunning,
            "error_message": "",
            "start_time":    time.Now().Unix(),
            "end_time":      0,
        })
    }
}
```

**步骤 2：创建新任务项（非重试或找不到已有项）**
```go
if taskItem == nil {
    taskItem = &model.SyncTaskItem{
        SyncTaskUuid: syncTaskUuid,
        TaskType:     taskCfg.TaskType,
        TaskName:     taskCfg.TaskName,
        Status:       constant.SyncTaskItemStatusRunning,
        StartTime:    time.Now().Unix(),
    }
    
    syncTaskItemRepo.Create(taskItem)
}
```

**步骤 3：执行同步任务**
```go
logger.Logger.Info("开始同步", zap.String("taskName", taskCfg.TaskName))

err := taskCfg.Executor(ctx)  // 调用实际的同步方法
endTime := time.Now().Unix()
```

**步骤 4：更新任务项状态**

**失败处理**：
```go
if err != nil {
    logger.Logger.Error("同步失败", 
        zap.String("taskName", taskCfg.TaskName), 
        zap.Error(err))
    *failCount++
    
    syncTaskItemRepo.Update(taskItem.Uuid, map[string]any{
        "status":        constant.SyncTaskItemStatusFailed,
        "error_message": err.Error(),
        "end_time":      endTime,
    })
}
```

**成功处理**：
```go
else {
    logger.Logger.Info("同步成功", 
        zap.String("taskName", taskCfg.TaskName))
    *successCount++
    
    syncTaskItemRepo.Update(taskItem.Uuid, map[string]any{
        "status":        constant.SyncTaskItemStatusSuccess,
        "error_message": "",
        "end_time":      endTime,
    })
}
```

### 4. 获取同步任务列表 (`GetTaskList`)

#### 功能描述
分页查询同步任务列表，支持按状态筛选。

#### 请求参数 (`req.SyncTaskListReq`)
- `PageNo`: 页码
- `PageSize`: 每页数量
- `Status`: 任务状态（可选，1-运行中，2-成功，3-失败）

#### 核心流程

**步骤 1：构建查询选项**
```go
opts := []repository.DBOption{
    syncTaskRepo.OrderByCreateTime(true),  // 按创建时间倒序
}

if listReq.Status != nil {
    opts = append(opts, syncTaskRepo.WhereStatus(*listReq.Status))
}
```

**步骤 2：分页查询**
```go
tasks, total, err := syncTaskRepo.GetListWithPagination(
    listReq.PageReq.PageNo, 
    listReq.PageReq.PageSize, 
    opts...)
```

**步骤 3：转换响应格式**
```go
list := make([]resp.SyncTaskListResp, 0, len(tasks))
for _, task := range tasks {
    list = append(list, convertToSyncTaskListResp(task))
}
```

#### 返回结构 (`resp.SyncTaskListPaginationResp`)
```go
{
    List: []SyncTaskListResp{
        Uuid,          // 任务UUID
        Status,        // 状态（1-运行中，2-成功，3-失败）
        TotalCount,    // 总任务数
        SuccessCount,  // 成功数
        FailCount,     // 失败数
        StartTime,     // 开始时间
        EndTime,       // 结束时间
        Duration,      // 耗时（秒）
        CreateTime,    // 创建时间
    },
    Meta: {PageNo, PageSize, Total}
}
```

### 5. 获取同步任务详情 (`GetTaskDetail`)

#### 功能描述
获取单个同步任务的详细信息，包括所有子任务的执行情况。

#### 请求参数 (`req.SyncTaskDetailReq`)
- `TaskUuid`: 任务UUID

#### 核心流程

**步骤 1：查询任务（预加载子任务）**
```go
task, err := syncTaskRepo.GetByUuid(detailReq.TaskUuid, 
    syncTaskRepo.PreloadItems())
```

**步骤 2：转换响应格式**
```go
return convertToSyncTaskDetailResp(*task), nil
```

#### 返回结构 (`resp.SyncTaskDetailResp`)
```go
{
    Uuid,          // 任务UUID
    Status,        // 状态
    TotalCount,    // 总任务数
    SuccessCount,  // 成功数
    FailCount,     // 失败数
    StartTime,     // 开始时间
    EndTime,       // 结束时间
    Duration,      // 耗时（秒）
    CreateTime,    // 创建时间
    Items: []SyncTaskItemResp{  // 子任务列表
        Uuid,          // 子任务UUID
        TaskType,      // 任务类型
        TaskName,      // 任务名称
        Status,        // 状态
        ErrorMessage,  // 错误信息
        StartTime,     // 开始时间
        EndTime,       // 结束时间
        Duration,      // 耗时（秒）
    }
}
```

## 辅助函数

### 1. convertToSyncTaskItemResp
转换子任务为响应格式，计算耗时。

```go
func convertToSyncTaskItemResp(item model.SyncTaskItem) resp.SyncTaskItemResp {
    duration := int64(0)
    if item.EndTime > 0 && item.StartTime > 0 {
        duration = item.EndTime - item.StartTime
    }
    
    return resp.SyncTaskItemResp{
        Uuid, TaskType, TaskName, Status, ErrorMessage,
        StartTime, EndTime, Duration,
    }
}
```

### 2. convertToSyncTaskDetailResp
转换主任务为详情响应格式，包含所有子任务。

```go
func convertToSyncTaskDetailResp(task model.SyncTask) resp.SyncTaskDetailResp {
    duration := int64(0)
    if task.EndTime > 0 && task.StartTime > 0 {
        duration = task.EndTime - task.StartTime
    }
    
    items := make([]resp.SyncTaskItemResp, 0, len(task.Items))
    for _, item := range task.Items {
        items = append(items, convertToSyncTaskItemResp(item))
    }
    
    return resp.SyncTaskDetailResp{
        Uuid, Status, TotalCount, SuccessCount, FailCount,
        StartTime, EndTime, Duration, CreateTime,
        Items: items,
    }
}
```

### 3. convertToSyncTaskListResp
转换主任务为列表响应格式（不包含子任务）。

```go
func convertToSyncTaskListResp(task model.SyncTask) resp.SyncTaskListResp {
    duration := int64(0)
    if task.EndTime > 0 && task.StartTime > 0 {
        duration = task.EndTime - task.StartTime
    }
    
    return resp.SyncTaskListResp{
        Uuid, Status, TotalCount, SuccessCount, FailCount,
        StartTime, EndTime, Duration, CreateTime,
    }
}
```

## 数据模型

### 1. model.SyncTask (同步任务)
主任务模型：

```go
type SyncTask struct {
    BaseModel                // Uuid, CreateTime, UpdateTime, DeleteTime
    Status       uint8       // 状态（1-运行中，2-成功，3-失败）
    TotalCount   uint32      // 总任务数
    SuccessCount uint32      // 成功数
    FailCount    uint32      // 失败数
    StartTime    int64       // 开始时间（Unix时间戳）
    EndTime      int64       // 结束时间（Unix时间戳）
    Panic        string      // Panic信息（发生panic时记录）
    
    Items        []SyncTaskItem `gorm:"foreignKey:SyncTaskUuid"` // 子任务
}
```

### 2. model.SyncTaskItem (同步任务子项)
子任务模型：

```go
type SyncTaskItem struct {
    BaseModel                // Uuid, CreateTime, UpdateTime, DeleteTime
    SyncTaskUuid uint64      // 主任务UUID
    TaskType     string      // 任务类型
    TaskName     string      // 任务名称
    Status       uint8       // 状态（1-运行中，2-成功，3-失败）
    ErrorMessage string      // 错误信息
    StartTime    int64       // 开始时间
    EndTime      int64       // 结束时间
}
```

## 状态常量

### 任务状态 (SyncTaskStatus)
- `SyncTaskStatusRunning = 1`: 运行中
- `SyncTaskStatusSuccess = 2`: 成功
- `SyncTaskStatusFailed = 3`: 失败

### 子任务状态 (SyncTaskItemStatus)
- `SyncTaskItemStatusRunning = 1`: 运行中
- `SyncTaskItemStatusSuccess = 2`: 成功
- `SyncTaskItemStatusFailed = 3`: 失败

## 业务规则

### 1. 并发控制规则

**同一公司只能同时运行一个同步任务**：
```go
if !syncTaskManager.tryStartTask(companyUuid) {
    return errors.New("数据同步中，请稍后再试")
}
```

**原因**：
- 防止数据冲突
- 避免资源竞争
- 保证数据一致性

### 2. 任务执行顺序

**顺序执行，不可并发**：
```go
for _, taskCfg := range tasksToExecute {
    s.executeSyncTask(ctx, syncTask.Uuid, taskCfg, 
        &successCount, &failCount, retryMode)
}
```

**原因**：
- 存在依赖关系（如商品依赖分类）
- 简化错误处理
- 便于追踪执行进度

### 3. 错误处理策略

**单个任务失败不影响其他任务**：
- 失败的任务记录错误信息
- 继续执行后续任务
- 最终统计成功和失败数量

**支持失败重试**：
- 记录失败的任务类型
- 允许针对失败任务重试
- 重试时复用原任务记录

### 4. 状态更新规则

**主任务状态**：
- 开始时：`SyncTaskStatusRunning`
- 完成时：
  - 全部成功：`SyncTaskStatusSuccess`
  - 有失败：`SyncTaskStatusFailed`

**子任务状态**：
- 开始时：`SyncTaskItemStatusRunning`
- 完成时：
  - 成功：`SyncTaskItemStatusSuccess`
  - 失败：`SyncTaskItemStatusFailed`

### 5. 最后同步时间更新规则

**仅在全部成功时更新**：
```go
if !isExceptionOccurred {
    lastSyncTime := time.Now().Unix()
    // 更新公司表的 last_sync_time
    s.dbm.GetDB(companyUuid).Model(&model.Company{}).
        Where("uuid = ?", companyUuid).
        Update("last_sync_time", lastSyncTime)
    // 同时更新默认库中的公司表
    s.dbm.GetDB(constant.DefaultDB).Model(&model.Company{}).
        Where("uuid = ?", companyUuid).
        Update("last_sync_time", lastSyncTime)
}
```

**更新两个地方**：
1. 公司数据库中的 `company` 表
2. 默认数据库中的 `company` 表

## 异常处理

### 1. Panic 恢复机制

使用 `defer recover()` 捕获 panic：
```go
defer func() {
    if r := recover(); r != nil {
        stack := string(debug.Stack())
        logger.Logger.Error("同步任务发生panic", 
            zap.Uint64("companyUuid", companyUuid), 
            zap.Any("panic", r), 
            zap.String("stack", stack))
        
        // 更新任务状态，记录 panic 信息
        syncTaskRepo.Update(syncTask.Uuid, map[string]any{
            "status":   constant.SyncTaskStatusFailed,
            "panic":    fmt.Sprintf("%v: %s", r, stack),
            "end_time": time.Now().Unix(),
        })
        isPanicOccurred = true
    }
    
    // 清理任务状态
    syncTaskManager.finishTask(companyUuid)
}()
```

**特点**：
- 捕获所有 panic
- 记录完整堆栈信息
- 更新任务状态为失败
- 确保释放并发锁

### 2. 错误日志

**任务级别日志**：
```go
logger.Logger.Info("开始执行同步任务", 
    zap.Uint64("companyUuid", companyUuid), 
    zap.Uint64("taskUuid", syncTask.Uuid))

logger.Logger.Info("同步任务完成", 
    zap.Uint64("companyUuid", companyUuid),
    zap.Uint32("successCount", successCount),
    zap.Uint32("failCount", failCount))
```

**子任务级别日志**：
```go
logger.Logger.Info("开始同步", zap.String("taskName", taskCfg.TaskName))
logger.Logger.Info("同步成功", zap.String("taskName", taskCfg.TaskName))
logger.Logger.Error("同步失败", 
    zap.String("taskName", taskCfg.TaskName), 
    zap.Error(err))
```

### 3. 错误信息存储

**子任务错误**：
```go
syncTaskItemRepo.Update(taskItem.Uuid, map[string]any{
    "status":        constant.SyncTaskItemStatusFailed,
    "error_message": err.Error(),  // 存储错误信息
    "end_time":      endTime,
})
```

**主任务 Panic**：
```go
syncTaskRepo.Update(syncTask.Uuid, map[string]any{
    "status":   constant.SyncTaskStatusFailed,
    "panic":    fmt.Sprintf("%v: %s", r, stack),  // 存储 panic 信息
    "end_time": time.Now().Unix(),
})
```

## WebSocket 通知

### 推送时机
同步任务完成时（无论成功、失败还是 panic）。

### 推送内容
```go
websocket.PushClient(
    company.Uuid,           // 公司UUID
    websocket.SourceShop,   // 源：Shop端
    websocket.SourceAll,    // 目标：所有端
    websocket.SYNC_DATA,    // 消息类型：同步数据
    map[string]any{
        "task_uuid":             syncTask.Uuid,
        "is_exception_occurred": isExceptionOccurred,  // 是否发生异常
        "sync_time":             time.Now().Unix(),
    })
```

### 客户端处理
- 前端接收通知后，可以：
  - 刷新同步状态显示
  - 提示用户同步完成
  - 如果异常，提示用户查看详情或重试

## 重试机制

### 1. 重试触发条件
- 主任务状态为失败（`SyncTaskStatusFailed`）
- 至少有一个子任务状态为失败（`SyncTaskItemStatusFailed`）

### 2. 重试流程

**步骤 1：识别失败任务**
```go
var retryTaskTypes []string
for _, item := range existTask.Items {
    if item.Status == constant.SyncTaskItemStatusFailed {
        retryTaskTypes = append(retryTaskTypes, item.TaskType)
    }
}
```

**步骤 2：筛选待重试任务**
```go
tasksToExecute := []syncTaskConfig{}
for _, task := range allTasks {
    if slices.Contains(retryTaskTypes, task.TaskType) {
        tasksToExecute = append(tasksToExecute, task)
    }
}
```

**步骤 3：重置任务项状态**
```go
syncTaskItemRepo.Update(taskItem.Uuid, map[string]any{
    "status":        constant.SyncTaskItemStatusRunning,
    "error_message": "",
    "start_time":    time.Now().Unix(),
    "end_time":      0,
})
```

**步骤 4：执行重试**
按原顺序执行失败的任务。

### 3. 重试特点
- **仅重试失败的任务**：不会重新执行成功的任务
- **复用原任务记录**：不创建新任务，更新原任务状态
- **保留历史记录**：可以看到重试前的错误信息（被覆盖前）

## 性能考虑

### 1. 异步执行
**默认异步执行**：
```go
utils.Go(func() {
    s.executeSync(ctx, syncTask, allTasks, retryMode, retryTaskTypes)
})
```

**优点**：
- 不阻塞API响应
- 用户体验好（立即返回）
- 适合大数据量同步

**同步执行选项**：
- 测试场景
- 需要立即知道结果的场景

### 2. 顺序执行子任务
虽然是顺序执行，但：
- **简化依赖管理**：不需要复杂的依赖图
- **避免资源竞争**：不会同时访问相同数据
- **便于调试**：执行顺序清晰

### 3. 数据库访问优化
- 使用批量操作（在各子任务的实现中）
- 适当使用事务（在各子任务的实现中）
- 减少不必要的查询

### 4. 并发锁
使用 `sync.Map` 实现轻量级并发控制：
- 无锁设计（原子操作）
- 性能高
- 适合读多写少的场景

## 依赖关系

### 外部服务依赖
1. **WarehouseSrv** (仓库服务):
   - `SyncWarehouse`: 同步仓库
   - `SyncWarehouseItemStock`: 同步仓库库存

2. **MaterialSrv** (物品服务):
   - `SyncMaterialCategory`: 同步物品分类
   - `SyncMaterial`: 同步物品
   - `SyncProductBomCard`: 同步BOM卡

3. **SupplierSrv** (供应商服务):
   - `SyncSupplier`: 同步供应商

4. **ProductSrv** (商品服务):
   - `SyncProductShopCategory`: 同步商品分类
   - `SyncProductTax`: 同步税率
   - `SyncUnit`: 同步单位
   - `SyncProductFlavor`: 同步规格
   - `SyncAttributeGroup`: 同步属性组
   - `SyncSauce`: 同步加料
   - `SyncProduct`: 同步商品
   - `SyncProductStockByBomCard`: 同步商品库存
   - `SyncProductPackageImage`: 同步套餐图片

### Repository 依赖
1. **SyncTaskRepo** (`repository.NewSyncTaskRepo`):
   - `Create`: 创建任务
   - `Update`: 更新任务
   - `GetByUuid`: 根据UUID查询
   - `GetListWithPagination`: 分页查询
   - `PreloadItems`: 预加载子任务

2. **SyncTaskItemRepo** (`repository.NewSyncTaskItemRepo`):
   - `Create`: 创建子任务
   - `Update`: 更新子任务
   - `GetList`: 查询子任务列表
   - `WhereSyncTaskUuid`: 按主任务UUID过滤
   - `WhereTaskType`: 按任务类型过滤

### 数据库依赖
- 使用 `dbm.GetDB(companyUuid)` 获取公司数据库（同步任务表在公司库）
- 使用 `dbm.GetDB(constant.DefaultDB)` 获取默认数据库（更新公司表）

## 使用示例

### 1. 启动全量同步（异步）
```go
syncSrv := NewSyncSrv(dbm, warehouseSrv, supplierSrv, productSrv, materialSrv)

resp, err := syncSrv.Sync(ctx, req.SyncReq{
    IsSyncExecute: false,  // 异步执行
})

// 立即返回
// resp.TaskUuid: 任务UUID
// resp.Message: "数据同步已启动"
```

### 2. 启动全量同步（同步）
```go
resp, err := syncSrv.Sync(ctx, req.SyncReq{
    IsSyncExecute: true,  // 同步执行
})

// 等待所有任务执行完成后返回
```

### 3. 重试失败的任务
```go
resp, err := syncSrv.Sync(ctx, req.SyncReq{
    TaskUuid:      existingTaskUuid,  // 原任务UUID
    IsSyncExecute: false,
})

// 仅重试失败的子任务
// resp.Message: "重试同步任务已启动"
```

### 4. 查询同步任务列表
```go
resp, err := syncSrv.GetTaskList(ctx, req.SyncTaskListReq{
    PageReq: req.PageReq{PageNo: 1, PageSize: 20},
    Status:  &status,  // 可选，1-运行中，2-成功，3-失败
})
```

### 5. 查询同步任务详情
```go
resp, err := syncSrv.GetTaskDetail(ctx, req.SyncTaskDetailReq{
    TaskUuid: taskUuid,
})

// 返回主任务和所有子任务的详细信息
```

## 数据流图

### 全量同步流程
```
用户触发同步
    ↓
检查并发（tryStartTask）
    ↓
创建主任务记录（SyncTask）
    ↓
启动异步执行
    ↓
【开始执行】
    ├─ 商品分类同步 → 创建子任务 → 执行 → 更新状态
    ├─ 物品分类同步 → 创建子任务 → 执行 → 更新状态
    ├─ 税率同步     → 创建子任务 → 执行 → 更新状态
    ├─ 单位同步     → 创建子任务 → 执行 → 更新状态
    ├─ 物品同步     → 创建子任务 → 执行 → 更新状态
    ├─ 仓库同步     → 创建子任务 → 执行 → 更新状态
    ├─ 规格同步     → 创建子任务 → 执行 → 更新状态
    ├─ 属性同步     → 创建子任务 → 执行 → 更新状态
    ├─ 加料同步     → 创建子任务 → 执行 → 更新状态
    ├─ 商品同步     → 创建子任务 → 执行 → 更新状态
    ├─ BOM卡同步    → 创建子任务 → 执行 → 更新状态
    ├─ 供应商同步   → 创建子任务 → 执行 → 更新状态
    ├─ 仓库库存同步 → 创建子任务 → 执行 → 更新状态
    ├─ 商品库存同步 → 创建子任务 → 执行 → 更新状态
    └─ 套餐图片同步 → 创建子任务 → 执行 → 更新状态
【执行完成】
    ↓
更新主任务状态（成功/失败）
    ↓
更新公司最后同步时间（仅成功时）
    ↓
推送 WebSocket 通知
    ↓
释放并发锁（finishTask）
```

### 重试流程
```
用户触发重试
    ↓
检查并发（tryStartTask）
    ↓
查询原任务及子任务
    ↓
识别失败的子任务类型
    ↓
筛选待重试任务
    ↓
【开始执行】
    ├─ 任务A（失败） → 查找子任务记录 → 重置状态 → 执行 → 更新状态
    ├─ 任务C（失败） → 查找子任务记录 → 重置状态 → 执行 → 更新状态
    └─ 任务E（失败） → 查找子任务记录 → 重置状态 → 执行 → 更新状态
【执行完成】
    ↓
更新主任务状态
    ↓
更新公司最后同步时间（仅全部成功时）
    ↓
推送 WebSocket 通知
    ↓
释放并发锁（finishTask）
```

## 监控与日志

### 1. 关键日志点
- 任务开始：`开始执行同步任务`
- 子任务开始：`开始同步 [任务名称]`
- 子任务成功：`同步成功 [任务名称]`
- 子任务失败：`同步失败 [任务名称]`
- 任务完成：`同步任务完成`
- Panic：`同步任务发生panic`

### 2. 监控指标
- 成功率：`successCount / totalCount`
- 失败率：`failCount / totalCount`
- 执行时间：`endTime - startTime`
- 并发任务数：`len(syncTaskManager.getRunningCompanyUuids())`

### 3. 告警建议
- 同步失败次数过多（如 3 次以上）
- 单次同步耗时过长（如 > 10 分钟）
- 发生 panic
- 同一公司频繁触发同步

## 总结

`sync.go` 实现了一个功能完整、设计合理的数据同步服务，主要特点包括：

1. **完整的任务管理**：主任务、子任务分离，状态清晰
2. **并发控制**：使用全局管理器防止重复执行
3. **灵活的执行模式**：支持同步/异步、全量/重试
4. **健壮的错误处理**：Panic 恢复、错误记录、失败重试
5. **实时通知**：WebSocket 推送同步结果
6. **性能优化**：异步执行、合理的执行顺序
7. **完善的日志**：关键节点都有日志记录
8. **清晰的状态追踪**：可查询历史任务和详细执行情况

该服务是系统与 ERP 集成的关键桥梁，确保本地数据与 ERP 数据的一致性，为采购、库存、销售等业务提供基础数据支持。

