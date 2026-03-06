# story-erp-takeout-exception-fallback 任务清单

## 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 2 |
| 总任务数 | 5 |
| 已完成 | 0 |
| 完成率 | 0% |

---

## Phase 1: ERPNext 配置

### 1.1 添加 BY001 备用商品到各环境

| 项目 | 内容 |
|------|------|
| File | ERPNext 配置 / BMP 初始化脚本 |
| Purpose | 确保所有环境和 site 包含 BY001 |
| Requirements | 测试/UAT/生产/初始化模板 |

- [ ] 完成

### 1.2 更新 site 初始化模板

| 项目 | 内容 |
|------|------|
| File | BMP site 初始化相关代码 |
| Purpose | 新建门店时自动包含 BY001 |

- [ ] 完成

---

## Phase 2: 降级逻辑

### 2.1 SI 创建时 item 映射降级

| 项目 | 内容 |
|------|------|
| File | BMP SI 创建逻辑 或 Main 外卖 ERP 同步 |
| Purpose | 外卖订单商品无法映射时使用 BY001 |
| Requirements | 保留原金额，description 记录原商品名 |

- [ ] 完成

### 2.2 POS Invoice 兼容（旧流程）

| 项目 | 内容 |
|------|------|
| File | 现有 POS Invoice 创建逻辑 |
| Purpose | 旧流程同样支持 BY001 降级 |
| Requirements | 改造方案上线前也能用 |

- [ ] 完成

---

## Phase 3: 测试

### 3.1 编写单元测试

| 项目 | 内容 |
|------|------|
| File | 相关测试文件 |
| Purpose | 验证降级逻辑正确 |
| Requirements | 映射成功/失败两种场景 |

- [ ] 完成

---

## 提交清单

### 代码质量
- [ ] `go mod tidy` 执行
- [ ] `go fmt ./...` 执行
- [ ] `go vet ./...` 通过
- [ ] 测试通过

### 功能完整性
- [ ] 外卖异常订单使用 BY001 落单
- [ ] SI/POS Invoice 均正常生成
- [ ] 各环境 BY001 已配置
- [ ] 新门店初始化包含 BY001
