# story-shop-home-store-switch 任务清单

## 📊 进度总览

| 项目     | 数值 |
| -------- | ---- |
| 总 SP    | 2    |
| 总任务数 | 4    |
| 已完成   | 3    |
| 完成率   | 75%  |

---

## Phase 1: 核心实现

### 1.1 扩展请求 DTO

| 项目         | 内容                                         |
| ------------ | -------------------------------------------- |
| File         | `main/app/dto/req/statistics.go`             |
| Purpose      | 在请求 DTO 中增加 CompanyUuid 字段           |
| Requirements | R1, R2, R3 - 支持传入门店参数                |
| Leverage     | 现有 DTO 结构                                |

**修改内容**:
- `BusinessDataCountReq` 增加 `CompanyUuid uint64` 字段
- `BusinessDataRankProductReq` 增加 `CompanyUuid uint64` 字段

- [x] 完成

---

### 1.2 实现辅助方法 switchCompanyContext

| 项目         | 内容                                           |
| ------------ | ---------------------------------------------- |
| File         | `main/app/api/v1/shop/shop_statistics.go`      |
| Purpose      | 实现门店切换逻辑，包含权限校验和数据库切换     |
| Requirements | R1-AC3, R2-AC3, R3-AC3 - 权限校验              |
| Leverage     | `GetCompanyList` (business.go:3792)            |

**实现要点**:
1. 检查 company_uuid 是否传入
2. 调用 GetCompanyList 获取用户可访问的门店列表
3. 校验目标门店是否在权限范围内
4. 使用 DBManager 获取目标门店数据库连接
5. 创建新的 context 并设置数据库连接

- [x] 完成

---

### 1.3 修改 3 个统计 Handler

| 项目         | 内容                                       |
| ------------ | ------------------------------------------ |
| File         | `main/app/api/v1/shop/shop_statistics.go`  |
| Purpose      | 在 3 个 Handler 中调用 switchCompanyContext |
| Requirements | R1, R2, R3 - 全部验收标准                  |
| Leverage     | 1.2 实现的辅助方法                         |

**修改的方法**:
- `CountBusiness` (第 63-98 行)
- `CountArea` (第 183-218 行)
- `CountProductRank` (第 220-255 行)

**修改模式**（每个方法相同）:
```go
// 1. 切换门店上下文（如果指定了 company_uuid）
targetCtx, err := h.switchCompanyContext(ctx, countReq.CompanyUuid)
if err != nil {
    helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
    return
}

// 2. 使用新上下文调用 Service
result, err := h.businessSrv.CountXxx(targetCtx, req)
```

- [x] 完成

---

## Phase 2: 测试验证

### 2.1 功能测试

| 项目         | 内容                              |
| ------------ | --------------------------------- |
| Purpose      | 验证所有验收标准                  |
| Requirements | R1-R3 全部 AC                     |

**测试场景**:
- [ ] 不传 company_uuid，返回当前门店数据
- [ ] 传入有权限的 company_uuid，返回指定门店数据
- [ ] 传入无权限的 company_uuid，返回权限错误
- [ ] 3 个接口都正常工作

- [ ] 完成

---

## 提交清单

### 代码质量
- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过
- [ ] 测试通过: `go test ./...`

### 功能完整性
- [x] 不传 company_uuid 时保持原有行为
- [x] 传入 company_uuid 时返回对应门店数据
- [x] 无权限时返回明确错误信息

### 文档更新
- [ ] API 文档更新（如有 Swagger）

---

## 验收标准对照

| AC ID   | 描述                                   | 对应任务 | 状态 |
| ------- | -------------------------------------- | -------- | ---- |
| R1-AC1  | 传入 company_uuid 返回对应门店营业数据 | 1.3      | ✅   |
| R1-AC2  | 不传 company_uuid 保持原有逻辑         | 1.3      | ✅   |
| R1-AC3  | 无权限门店返回权限错误                 | 1.2, 1.3 | ✅   |
| R2-AC1  | 传入 company_uuid 返回对应门店区域数据 | 1.3      | ✅   |
| R2-AC2  | 不传 company_uuid 保持原有逻辑         | 1.3      | ✅   |
| R2-AC3  | 无权限门店返回权限错误                 | 1.2, 1.3 | ✅   |
| R3-AC1  | 传入 company_uuid 返回对应门店商品排行 | 1.3      | ✅   |
| R3-AC2  | 不传 company_uuid 保持原有逻辑         | 1.3      | ✅   |
| R3-AC3  | 无权限门店返回权限错误                 | 1.2, 1.3 | ✅   |

---

**版本**: v1.0.0
**创建日期**: 2026-02-25
**最后更新**: 2026-02-25
