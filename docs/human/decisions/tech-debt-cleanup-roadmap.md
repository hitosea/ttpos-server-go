# TTPOS 技术债务治理路线图

> 生成日期：2026-03-05
> 状态：Draft — 待团队评审

---

## 总览

基于全量代码扫描，识别出 **4 类 P0 生产风险、5 类 P1 架构问题、9 类 P2 代码质量问题、5 类 P3 运维风险**。本路线图分 4 个阶段推进，遵循"先止血、再固本、后提质"原则。

### 阶段总览

| 阶段 | 主题 | 周期 | 核心目标 |
|------|------|------|----------|
| Phase 1 | 止血 — 生产风险消除 | 1-2 周 | 消除 panic、数据损坏风险 |
| Phase 2 | 固本 — 架构治理 | 4-8 周 | God Object 拆分、分层修正、双轨收敛 |
| Phase 3 | 提质 — 代码质量 | 持续 | 死代码清理、规范统一、测试补全 |
| Phase 4 | 防御 — 机制建设 | 持续 | CI 卡点、自动化检测、规范守护 |

---

## Phase 1：止血 — 生产风险消除（1-2 周）

### 1.1 修复分布式锁 panic（P0）

**文件**：`main/pkg/lock/lock_redsync.go:126,129,172,175`

**问题**：Redis unlock 失败时 `panic`，生产环境 Redis 抖动会导致请求协程崩溃。

**方案**：
```go
// Before: panic on unlock failure
// After: log error + return error, caller decides
func ClearUuidLock(ctx context.Context, lock *redsync.Mutex) error {
    ok, err := lock.UnlockContext(ctx)
    if err != nil || !ok {
        logger.Error(ctx, "redis unlock failed", zap.Error(err))
        return err // 不再 panic
    }
    return nil
}
```

**影响范围**：需排查所有 `ClearUuidLock` / `ClearUuidLockString` 调用点，补充 error 处理。

**验证**：模拟 Redis 连接断开场景，确认服务不崩溃。

---

### 1.2 修复业务模型 panic（P0）

**文件**：`main/app/model/sale_bill_ext_getset.go:728`

**问题**：`panic("saleOrderProduct.ProductPackage is nil")` — 订单产品套餐为空时直接崩溃。

**方案**：替换为 `return error`，由调用方处理（跳过或返回业务错误码）。

---

### 1.3 修复 BMP takeout panic（P0）

**文件**：
- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go:35`
- `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/skootar.go:35`
- `ttpos-bmp/app/ttpos-takeout/internal/client/grab/config.go:20`
- `ttpos-bmp/app/ttpos-takeout/internal/client/lineman/config.go:30`

**问题**：配置读取失败时 `panic(gerror.Newf(...))`，导致进程崩溃。

**方案**：改为返回 error，由上层决定是否降级或告警。

---

### 1.4 修复 Skootar 订单 shop_uuid 为空（P0）

**文件**：`ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/create_order.go:93`

**问题**：`ShopUuid: ""` 硬编码空字符串入库，商户查询匹配不到这些订单。

**方案**：从请求 context 或配置中获取 shop_uuid，入库前校验非空。

---

### 1.5 清理提交的敏感信息（P0）

**文件**：
- `ttpos-bmp/app/ttpos-*/manifest/config/config.yaml` — 明文 dev 数据库密码
- `ttpos-bmp/app/ttpos-erp/manifest/sql/20250806141830_site-conf.up.sql` — 硬编码 API Key + 内网 IP

**方案**：
1. config.yaml 改为 `config.yaml.example`，实际配置走环境变量/Nacos
2. migration seed 数据使用占位符，由部署脚本注入
3. `.gitignore` 排除 `config.yaml`（保留 `.example`）

---

## Phase 2：固本 — 架构治理（4-8 周）

### 2.1 God Object 拆分

按业务子域拆分，每次拆一个文件，逐步推进。

#### product.go（9757 行）→ 拆分方案

| 新文件 | 职责 | 预估行数 |
|--------|------|----------|
| `product.go` | 核心 CRUD、查询 | ~2000 |
| `product_sync.go` | ERP 同步逻辑 | ~1500 |
| `product_sku.go` | SKU/BOM 管理 | ~1500 |
| `product_attribute.go` | 属性组、规格 | ~1200 |
| `product_category.go` | 分类管理 | ~800 |
| `product_takeout.go` | 外卖平台映射 | ~1000 |
| `product_import.go` | 导入/批量操作 | ~800 |
| `product_utils.go` | 工具函数 | ~500 |

**原则**：
- 同一 Service 接口 `IProductSrv` 不变，实现拆分到多文件
- Go 允许同包多文件实现同一 struct 的方法，无需改接口
- 逐文件提交，每次 PR 只拆一个子域

#### order.go（6144 行）→ 拆分方案

| 新文件 | 职责 |
|--------|------|
| `order.go` | 核心下单流程 |
| `order_query.go` | 列表/详情/搜索 |
| `order_status.go` | 状态机流转 |
| `order_discount.go` | 折扣/优惠计算 |
| `order_batch.go` | 批次标签逻辑 |
| `order_export.go` | 导出/报表 |

#### business.go（6445 行）→ 拆分方案

| 新文件 | 职责 |
|--------|------|
| `business.go` | 营业管理核心 |
| `business_summary.go` | 营业汇总 |
| `business_shift.go` | 交接班 |
| `business_report.go` | 报表逻辑 |

---

### 2.2 API 层分层违规修正

**违规文件**（5 个）：

| 文件 | 违规方式 | 修复方案 |
|------|----------|----------|
| `api/v1/admin/admin_takeout.go:87` | 直接调用 `repository.NewCompanyRepo` | 迁移到 Service 层 |
| `api/v1/shop/shop_ai_agent.go:67` | 直接调用 `repository.NewWarehouseRepo` | 迁移到 Service 层 |
| `api/v1/h5/h5_handler.go` (6处) | 传递 `repository.WithXxx()` 选项 | Service 层封装为业务语义参数 |
| `api/v1/assistant/assistant_desk.go:391` | 传递 `repository.WithH5OrderUuid` | 同上 |
| `api/v1/cashier/cashier_desk.go:840` | 传递 `repository.WithH5OrderUuid` | 同上 |

**修复模式**：在 Service 层新增业务语义方法，API 层只传业务参数：
```go
// Before (API 层):
orderSrv.GetOrders(ctx, repository.WithH5OrderUuid(uuid))

// After (API 层):
orderSrv.GetH5Orders(ctx, uuid)

// After (Service 层内部):
func (s *OrderSrv) GetH5Orders(ctx context.Context, uuid string) {
    s.repo.GetOrders(repository.WithH5OrderUuid(uuid))
}
```

---

### 2.3 Go/PHP 双轨收敛策略

当前 PHP admin 模块仍在活跃开发（90 天 303 次提交），且是数据库迁移权威（88 个 migration）。

**分阶段收敛计划**：

| 阶段 | 行动 | 产出 |
|------|------|------|
| 盘点 | 列出 PHP 中所有活跃 API endpoint 及调用方 | 双轨 API 清单 |
| 冻结 | PHP 端功能冻结，仅允许 bugfix | 团队共识 |
| 迁移 | 按优先级将 PHP API 迁移至 Go | 每个 API 一个 PR |
| 切换 | Nginx 路由逐步切换到 Go | 灰度发布 |
| 清理 | 废弃 PHP 端点、移除 dead code | 代码瘦身 |

**数据库迁移权威转移**：
1. 冻结 PHP migration，新 migration 只在 Go 侧创建
2. Go 侧引入 migration 工具（如 golang-migrate）
3. 迁移历史记录对齐验证

---

### 2.4 DDD 仓库模块决策

**文件**：`main/app/modules/inventory/application/warehouse_adapter.go` — 6 个方法全注释。

**决策**：
- 若确认继续推进 DDD 仓库模块 → 补全实现
- 若决定放弃 → 删除整个 `modules/inventory/` 目录，避免误导

---

## Phase 3：提质 — 代码质量（持续）

### 3.1 死代码清理（1-2 天批量处理）

| 类型 | 数量 | 处理方式 |
|------|------|----------|
| 大块注释代码（>8 行） | 10+ 处 | 直接删除，git history 保留 |
| `fmt.Println` 调试输出 | 29 处（printer 模块） | 替换为 `logger.Debug` 或删除 |
| 未使用 DAO（Logistics） | 2 个 | 确认无引用后删除 |
| 空壳 stub 函数 | SMS/webhook/metrics | 标注 `// NOT_IMPLEMENTED` 或删除 |

### 3.2 规范统一（可脚本化）

| 项目 | 当前状态 | 目标 | 工具 |
|------|----------|------|------|
| `interface{}` → `any` | 多处未替换 | 全量替换 | `gofmt -r 'interface{} -> any'` |
| `context.TODO()` → 传递真实 ctx | 3 处 ERP gRPC | 传递请求 ctx | 手动 |
| `fmt.Errorf` → `gerror`（BMP） | 4 个文件 | 全量替换 | 手动 |
| TODO v2.12.0 估值率 | 5+ 处硬编码 0 | 接入真实数据或创建 Issue 跟踪 | 产品决策 |

### 3.3 Migration 规范修复

| 问题 | 文件 | 修复 |
|------|------|------|
| takeout_order 系列仍用 `datetime` | `20251205085121_*.sql` | 新增修正 migration，ALTER 为 `int` |
| `erp_logistics` 缺 `AUTO_INCREMENT` | `20250917110105_*.sql` | 新增修正 migration |
| `erp_logistics` 缺 `create_time` 等 | 同上 | 同上 |
| 占位列 `reserve1/reserve2` | 同上 | 删除或重命名为有意义字段名 |
| Migration 命名不规范 `_add` | `20250926140318_add.up.sql` | 无法重命名（已执行），记录规范避免再犯 |

### 3.4 测试覆盖率提升

当前 Service 层覆盖率 ~9%（120 个文件仅 11 个有测试）。

**策略：先守后攻**

| 阶段 | 目标 | 方式 |
|------|------|------|
| 守 | 新代码必须有测试 | PR 强制 coverage 检测 |
| 攻（核心） | order/product/payment 核心路径 | 集成测试 + 关键函数单测 |
| 攻（风险） | 金额计算、税费、折扣 | 表驱动测试，覆盖边界条件 |

**优先补测文件**（按业务风险排序）：

1. `service/order_pay.go` — 支付逻辑，金额敏感
2. `service/order.go` — 核心下单
3. `service/product.go` — 商品管理
4. `model/sale_order_buffet_customer_type_ext_calc.go` — 自助餐计算（已有 TODO 提示精度问题）
5. `model/sale_order_product_ext_calc.go` — 折扣计算（已有 TODO 提示精度问题）

---

## Phase 4：防御 — 机制建设（持续）

### 4.1 CI 卡点

| 检查项 | 工具 | 阶段 |
|--------|------|------|
| `panic` 使用检测 | `go vet` + custom linter | PR 检查 |
| `interface{}` 禁用 | `gofmt -r` check | PR 检查 |
| `fmt.Println` 禁止 | `grep` + CI script | PR 检查 |
| API 层 import repository 检测 | custom lint rule | PR 检查 |
| 新文件超过 800 行告警 | CI script | PR 检查 |
| 测试覆盖率不回退 | `go test -coverprofile` | PR 检查 |
| 敏感信息检测 | `trufflehog` / `gitleaks` | PR 检查 |

### 4.2 自动化脚本

```bash
# 建议添加到 Makefile
make lint          # 运行所有 lint 检查
make lint-panic    # 检测业务代码中的 panic
make lint-layer    # 检测分层违规
make lint-dead     # 检测 fmt.Println / 注释代码块
make lint-secret   # 敏感信息检测
```

### 4.3 架构守护

- 新增 `ARCHITECTURE.md` 记录分层规则和依赖方向
- Service 文件超过 1500 行时 CI 发出 warning
- 定期（季度）运行全量技术债务扫描，对比趋势

---

## 执行跟踪

| Phase | 任务 | Owner | 状态 | 备注 |
|-------|------|-------|------|------|
| 1.1 | 修复分布式锁 panic | - | TODO | |
| 1.2 | 修复 sale_bill panic | - | TODO | |
| 1.3 | 修复 BMP takeout panic | - | TODO | |
| 1.4 | 修复 Skootar shop_uuid | - | TODO | |
| 1.5 | 清理提交的敏感信息 | - | TODO | |
| 2.1 | product.go 拆分 | - | TODO | |
| 2.1 | order.go 拆分 | - | TODO | |
| 2.1 | business.go 拆分 | - | TODO | |
| 2.2 | API 层分层修正 | - | TODO | |
| 2.3 | Go/PHP 双轨盘点 | - | TODO | |
| 2.4 | DDD 仓库模块决策 | - | TODO | |
| 3.1 | 死代码清理 | - | TODO | |
| 3.2 | 规范统一 | - | TODO | |
| 3.3 | Migration 修复 | - | TODO | |
| 3.4 | 测试覆盖率提升 | - | TODO | |
| 4.1 | CI 卡点建设 | - | TODO | |
| 4.2 | Makefile lint 命令 | - | TODO | |
