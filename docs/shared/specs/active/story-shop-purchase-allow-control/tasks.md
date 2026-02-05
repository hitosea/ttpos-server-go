# story-shop-purchase-allow-control 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 2 |
| 总任务数 | 7 |
| 已完成 | 6 |
| 完成率 | 86% |

---

## Phase 1: Model + DTO + 迁移

### 1.1 添加 Model 字段

| 项目 | 内容 |
|------|------|
| File | `main/app/model/purchase_limit_scheme_item.go` |
| Purpose | 在 PurchaseLimitSchemeItem 结构体中添加 IsAllowPurchase 字段 |
| Requirements | R1: 字段扩展 |
| Leverage | 现有 Model 结构 |

**变更内容**:
```go
// 在 PurchaseLimitSchemeItem 结构体中添加
IsAllowPurchase string `gorm:"column:is_allow_purchase;type:varchar(10);not null;default:'yes';comment:'是否允许采购 yes/no'" json:"is_allow_purchase"`
```

- [x] 完成

### 1.2 添加 DTO Req 字段

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/req/purchase_limit_scheme.go` |
| Purpose | 在 PurchaseLimitSchemeItemReq 结构体中添加 IsAllowPurchase 字段 |
| Requirements | R2, R3: 接口适配 |
| Leverage | 现有 DTO 结构 |

**变更内容**:
```go
// 在 PurchaseLimitSchemeItemReq 结构体中添加
IsAllowPurchase string `json:"is_allow_purchase"` // 是否允许采购 yes/no，默认 yes
```

- [x] 完成

### 1.3 添加 DTO Resp 字段

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/resp/purchase_limit_scheme_resp.go` |
| Purpose | 在 PurchaseLimitSchemeItemResp 结构体中添加 IsAllowPurchase 字段 |
| Requirements | R2, R3: 接口适配 |
| Leverage | 现有 DTO 结构 |

**变更内容**:
```go
// 在 PurchaseLimitSchemeItemResp 结构体中添加
IsAllowPurchase string `json:"is_allow_purchase"`
```

- [x] 完成

### 1.4 创建数据库迁移文件

| 项目 | 内容 |
|------|------|
| File | `admin/database/migrations/20260203151135_add_is_allow_purchase_to_purchase_limit_scheme_item.php` |
| Purpose | 添加 is_allow_purchase 字段到数据库表 |
| Requirements | R1: 数据库字段扩展 |
| Leverage | 现有迁移脚本模式 |

- [x] 完成

### 1.5 更新种子文件

| 项目 | 内容 |
|------|------|
| File | `admin/database/seeds/shop_01.sql` |
| Purpose | 同步更新种子数据中的表结构 |
| Requirements | 数据库约束 |
| Leverage | 现有种子文件 |

- [x] 完成

---

## Phase 2: Service + API 适配

### 2.1 Service 层适配（如需要）

| 项目 | 内容 |
|------|------|
| File | `main/app/service/purchase_order/purchase_limit_scheme.go` |
| Purpose | 检查并适配 Service 层逻辑（字段映射、默认值处理） |
| Requirements | R2, R3: 接口适配 |
| Leverage | 现有 Service 实现 |

**检查点**:
- [x] Item 转换逻辑是否需要处理新字段
- [x] 默认值 "yes" 是否需要在 Service 层设置
- [ ] 参数校验（仅接受 yes/no）- 可选，当前未实现

- [x] 完成

---

## Phase 3: 测试与文档

### 3.1 功能测试

| 项目 | 内容 |
|------|------|
| File | - |
| Purpose | 验证功能正确性 |
| Requirements | 所有验收标准 |

**测试用例**:
- [ ] 创建限购方案时设置 is_allow_purchase = "yes"
- [ ] 创建限购方案时设置 is_allow_purchase = "no"
- [ ] 创建限购方案时不传 is_allow_purchase（应默认 "yes"）
- [ ] 更新限购方案时修改 is_allow_purchase
- [ ] 传入无效值（非 yes/no）应返回错误

- [ ] 完成（待手动测试）

---

## 提交清单

### 代码质量
- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过
- [ ] 测试通过: `go test ./...`

### 功能完整性
- [ ] 所有验收标准通过
- [x] API 响应格式正确（data 为对象）
- [x] 字段默认值正确（未传参数时为 "yes"）

### 迁移同步
- [x] 迁移文件已创建
- [x] shop_01.sql 已更新

### Git 提交
- [ ] 提交信息符合规范: `feat(shop): 采购限制方案增加是否允许采购字段`

---

**版本**: v1.0.0
**创建日期**: 2026-02-03
**更新日期**: 2026-02-03
