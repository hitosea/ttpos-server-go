# 渠道营业统计 API 任务分解

> 仅涉及 Go Main 后端，专注渠道营业统计查询与导出。

## 📋 任务分解原则

- 单任务 1-4 小时，可独立交付
- 任务与 requirements.md 需求编号互相映射

## 📊 进度总览

**总任务数**: 10  
**已完成**: 0  
**完成率**: 0%

---

## Phase 1: 需求确认 & 代码复用

- [ ] 1.1 复查 `CountSale` 逻辑可复用范围  
  - File: `main/app/repository/statistics.go`  
  - Requirements: R1.1-R1.4  
  - Purpose: 找出渠道统计所需字段/条件

---

## Phase 2: Repository & DTO

- [ ] 2.1 抽象 `CountChannelSale` 方法  
  - File: `main/app/repository/statistics.go`  
  - Requirements: R1.1-R1.4  
  - Leverage: `CountSale`  
  - Output: map[string]*ChannelSaleRepoResult

- [ ] 2.2 编写 Repository 单元测试  
  - File: `main/app/repository/statistics_channel_test.go`  
  - Requirements: R1.*  
  - 覆盖：无数据、单渠道、多渠道

- [ ] 2.3 新增 DTO (req/resp)  
  - Files: `main/app/dto/req/statistics_channel_req.go`, `main/app/dto/resp/statistics_channel_resp.go`  
  - Requirements: R1.1-R1.4, R2.1  
  - Purpose: 统一 API 输入输出

---

## Phase 3: Service

- [ ] 3.1 扩展 `IBusinessSrv` 接口  
  - File: `main/app/service/i_business_srv.go`  
  - Requirements: R1.*, R2.*  
  - Methods: `CountChannelSales`, `ExportChannelSales`

- [ ] 3.2 在 `business.go` 中实现  
  - File: `main/app/service/business.go`  
  - Requirements: R1.*, R2.*  
  - 任务：时间默认值、渠道补齐、导出调用

- [ ] 3.3 Service 单元测试  
  - File: `main/app/service/business_channel_test.go`  
  - 覆盖：默认时间、渠道补齐、导出异常

---

## Phase 4: API 层

- [ ] 4.1 新增 API Handler  
  - File: `main/app/api/v1/shop/shop_statistics.go`  
  - Methods: `ChannelSales`, `ChannelSalesExport`  
  - Requirements: R1.*, R2.*

- [ ] 4.2 注册路由  
  - File: `main/router/router.go`  
  - Ensure: 权限中间件覆盖

- [ ] 4.3 API 集成测试  
  - File: `main/app/api/v1/shop/shop_statistics_test.go`  
  - 场景：默认时间、参数错误、导出成功

---

## Phase 5: 导出能力 & 文档

- [ ] 5.1 Excel 模板实现  
  - File: `pkg/excel/channel_sales_template.go`（或现有导出工具）  
  - Requirements: R2.1-R2.6  
  - 输出：与截图一致的表头/行，参考 `ExportBusinessPaymentMethod` 的导出写法（表头、文件命名、导出日志）

- [ ] 5.2 文档 / API 描述同步  
  - Files: `docs/shared/api/statistics.md`, `CHANGELOG.md`  
  - Purpose: 记录新接口、导出说明

---

## Graphiti

- 若在实现中过滤渠道口径有经验，请记录 Episode 并更新此章节链接。


