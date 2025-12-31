# 外卖商品统计接口 任务分解

> 本文档定义外卖商品统计接口的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 8  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: Service 层实现

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 1.2）
- **Language**: 技术栈（Go/PHP/Vue）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [ ] 1.1 在 Service 接口中添加统计方法

  - **File**: `main/app/service/product_takeout.go`
  - **Purpose**: 为 IProductTakeoutSrv 接口添加 GetProductCount 方法签名
  - **Requirements**: 1.1, 1.4
  - **Language**: Go
  - **Leverage**: 现有接口定义: `main/app/service/product_takeout.go` (IProductTakeoutSrv)
  - **Prompt**: 
    ```
    Role: Go Developer specializing in Service Layer
    Task: 在 IProductTakeoutSrv 接口中添加 GetProductCount 方法
    Context: 方法签名为 GetProductCount(ctx context.Context, companyUuid uint64, platform string, forceRefresh bool) (int64, error)
    Restrictions: 遵循 .cursor/rules/go-main.mdc
    Success: 接口方法添加成功,签名正确
    ```

- [ ] 1.2 实现统计查询核心逻辑

  - **File**: `main/app/service/product_takeout.go`
  - **Purpose**: 实现 GetProductCount 方法,包含数据库查询逻辑
  - **Requirements**: 1.1, 1.3, 1.4
  - **Language**: Go
  - **Leverage**: 
    - 现有查询逻辑: `main/app/service/product_takeout.go` (其他查询方法)
    - Model: `main/app/model/product_takeout.go`
  - **Prompt**:
    ```
    Role: Go Developer with GORM expertise
    Task: 实现 GetProductCount 方法的数据库查询逻辑
    Context: 
    - 获取 db 实例: s.dbm.GetDB(ctx)
    - 查询条件: company_uuid = ? AND delete_time = 0
    - 如果 platform 非空,添加条件: takeout_platform = ?
    - 使用 db.Model(&model.ProductTakeout{}).Where(...).Count(&count)
    Restrictions: 遵循 .cursor/rules/go-main.mdc, 不使用 panic
    Success: 查询逻辑正确,返回商品总数
    ```

- [ ] 1.3 实现缓存读取逻辑

  - **File**: `main/app/service/product_takeout.go`
  - **Purpose**: 在查询前检查Redis缓存,命中则直接返回
  - **Requirements**: 1.5
  - **Language**: Go
  - **Leverage**: 
    - Cache 实例: `s.cache`
    - 现有缓存使用: 搜索 `cache.Get` 示例
  - **Prompt**:
    ```
    Role: Go Developer with Redis expertise
    Task: 实现缓存读取逻辑
    Context:
    - 构造缓存 Key: takeout:products:count:{companyUuid}:{platform}
    - 如果 forceRefresh=true, 跳过缓存读取
    - 使用 s.cache.Get(key) 读取缓存
    - 类型断言为 int64
    - 缓存命中则直接返回,未命中则继续数据库查询
    Restrictions: 遵循 .cursor/rules/go-main.mdc
    Success: 缓存读取正确,命中时直接返回
    ```

- [ ] 1.4 实现缓存写入逻辑

  - **File**: `main/app/service/product_takeout.go`
  - **Purpose**: 数据库查询后将结果写入Redis缓存
  - **Requirements**: 1.5
  - **Language**: Go
  - **Leverage**: Cache 实例: `s.cache`
  - **Prompt**:
    ```
    Role: Go Developer with Redis expertise
    Task: 实现缓存写入逻辑
    Context:
    - 数据库查询成功后写入缓存
    - 使用 s.cache.Set(key, count, 5*time.Minute)
    - 缓存写入失败只记录警告日志,不影响结果返回
    - 使用 logger.Logger.Warn 记录错误
    Restrictions: 遵循 .cursor/rules/go-main.mdc
    Success: 缓存写入正确,失败不影响返回
    ```

- [ ] 1.5 实现缓存清除方法

  - **File**: `main/app/service/product_takeout.go`
  - **Purpose**: 提供缓存清除方法,供商品导入/删除时调用
  - **Requirements**: 1.5
  - **Language**: Go
  - **Leverage**: Cache 实例: `s.cache`
  - **Prompt**:
    ```
    Role: Go Developer
    Task: 实现 ClearProductCountCache 方法
    Context:
    - 方法签名: ClearProductCountCache(ctx context.Context, companyUuid uint64, platform string)
    - 清除指定平台的缓存 Key
    - 清除所有平台的缓存 Key (platform="")
    - 使用 s.cache.Del(key)
    Restrictions: 遵循 .cursor/rules/go-main.mdc
    Success: 缓存清除方法实现完整
    ```

---

## Phase 2: Handler 层实现

- [ ] 2.1 创建 Handler 方法

  - **File**: `main/app/api/v1/shop/shop_takeout.go`
  - **Purpose**: 在 TakeoutHandler 中添加 GetProductCount 方法
  - **Requirements**: 1.1, 1.2, 1.6
  - **Language**: Go
  - **Leverage**: 
    - 现有 Handler: `main/app/api/v1/shop/shop_takeout.go`
    - 其他 Handler 方法作为参考
  - **Prompt**:
    ```
    Role: Go Developer with Gin framework expertise
    Task: 创建 GetProductCount Handler 方法
    Context:
    - 获取参数: platform := c.Query("platform"), forceRefresh := c.Query("force_refresh") == "1"
    - 获取 Context: ctx := helper.GetContext(c)
    - 获取商家UUID: companyUuid := ctx.GetCompanyUuid()
    - 调用 Service: h.productTakeoutSrv.GetProductCount(ctx, companyUuid, platform, forceRefresh)
    - 成功响应: helper.Success(c, gin.H{"total": total})
    - 错误响应: helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "查询商品统计失败"))
    Restrictions: 遵循 .cursor/rules/api.mdc
    Success: Handler 方法实现完整,参数处理正确
    ```

- [ ] 2.2 添加 Swagger 注释

  - **File**: `main/app/api/v1/shop/shop_takeout.go`
  - **Purpose**: 为 GetProductCount 方法添加完整的 Swagger API 文档注释
  - **Requirements**: 1.1
  - **Language**: Go
  - **Leverage**: 现有 Swagger 注释示例
  - **Prompt**:
    ```
    Role: API Documentation Specialist
    Task: 添加 Swagger 注释
    Context:
    - @Summary 获取外卖商品统计
    - @Description 获取指定平台或所有平台的外卖商品总数
    - @Tags 商家端.外卖管理
    - @Param platform query string false "外卖平台(grab/lineman等,不传则统计所有平台)"
    - @Param force_refresh query int false "强制刷新缓存(1=是,0=否)"
    - @Success 200 {object} response.ProductCountResponse
    - @Router /shop/takeout/products/count [get]
    Restrictions: 遵循 Swagger 注释规范
    Success: Swagger 注释完整准确
    ```

- [ ] 2.3 注册路由

  - **File**: `main/app/api/v1/shop/shop_takeout.go`
  - **Purpose**: 在 RegisterTakeoutHandlers 中注册统计接口路由
  - **Requirements**: 1.1
  - **Language**: Go
  - **Leverage**: 现有路由注册: `RegisterTakeoutHandlers` 函数
  - **Prompt**:
    ```
    Role: Go Developer
    Task: 注册统计接口路由
    Context:
    - 在 privateApi 组中添加路由
    - 代码: privateApi.GET("/takeout/products/count", takeoutHandler.GetProductCount)
    - 确保在 Auth 中间件保护下
    Restrictions: 遵循现有路由注册模式
    Success: 路由注册成功,可通过 /shop/takeout/products/count 访问
    ```

---

## Phase 3: 测试

- [ ] 3.1 编写 API 测试

  - **File**: 手动测试或使用 Postman/curl
  - **Purpose**: 验证 API 接口功能正确
  - **Requirements**: 所有需求
  - **Language**: Shell/Postman
  - **Leverage**: 现有 API 测试方法
  - **Test Cases**:
    ```bash
    # 测试1: 查询Grab平台商品数
    curl -H "Authorization: Bearer {token}" \
         "http://localhost/shop/takeout/products/count?platform=grab"
    
    # 测试2: 查询所有平台商品数
    curl -H "Authorization: Bearer {token}" \
         "http://localhost/shop/takeout/products/count"
    
    # 测试3: 强制刷新缓存
    curl -H "Authorization: Bearer {token}" \
         "http://localhost/shop/takeout/products/count?platform=grab&force_refresh=1"
    
    # 测试4: 未授权访问(应返回401)
    curl "http://localhost/shop/takeout/products/count"
    
    # 预期结果:
    # - 返回正确的商品统计数量
    # - 响应格式: {code: 1, message: "success", data: {total: N}}
    # - 第二次请求应该更快(缓存命中)
    # - 强制刷新后缓存被清除
    ```
  - **Success**: 所有测试用例通过,响应格式正确

---

## 执行建议

### 推荐执行顺序

1. **Phase 1**: Service 层实现(任务 1.1-1.5)
   - 先实现核心查询逻辑(1.1, 1.2)
   - 再实现缓存机制(1.3, 1.4, 1.5)

2. **Phase 2**: Handler 层实现(任务 2.1-2.3)
   - 实现 Handler 方法(2.1)
   - 添加文档注释(2.2)
   - 注册路由(2.3)

3. **Phase 3**: 测试(任务 3.1)
   - 执行所有测试用例
   - 验证功能正确性

### 技术栈确认

- **后端**: Go (main 模块)
- **框架**: Gin + GORM
- **缓存**: Redis
- **数据表**: ttpos_product_takeout (已存在)

### 关键检查点

- [ ] Service 接口方法签名正确
- [ ] 数据库查询逻辑正确(条件、软删除)
- [ ] 缓存机制完整(读、写、清除)
- [ ] Handler 参数处理正确
- [ ] 响应格式符合 API 规范
- [ ] Swagger 文档完整
- [ ] 路由注册成功
- [ ] 所有测试通过

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板:`docs/agent/templates/graphiti-episode.md`
- 活动日志:`docs/team/activities/2025-12/2025-12-18.md`
- 任务执行过程中遇到的技术难点、踩坑经验应及时记录到 Episode。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-18  
**作者**: weifashi  
**最后更新**: 2025-12-18

