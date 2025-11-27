# Bug-251127-005 修复任务清单

> **当前状态**：🟡 规划中  
> **Bug ID**：bug-251127-005  
> **模块**：material（物料管理）  
> **开始时间**：2025-11-27  
> **预计完成**：2025-11-28

---

## 📋 任务列表

### 1. 代码修复

- [x] **修改响应 DTO 结构体** `main/app/dto/resp/material_resp/material.go`
  - 需求：在 `MaterialDetailResp` 结构体中添加 `AllowSubstoreVisible` 字段
  - 位置：第 54-79 行，在 `Status` 字段后添加
  - 预计时间：0.5 小时
  - 负责人：weifashi
  - 详细说明：
    ```go
    AllowSubstoreVisible int `json:"allow_substore_visible"` // 允许子店可见：1-允许，0-不允许（仅总店可用）
    ```
  - ✅ 已完成

- [x] **修改 Service 层响应构建** `main/app/service/material.go`
  - 需求：在 `GetMaterialDetail` 方法中填充 `AllowSubstoreVisible` 字段
  - 位置：第 428-455 行，在返回结构体中添加字段赋值
  - 预计时间：0.5 小时
  - 负责人：weifashi
  - 详细说明：
    ```go
    AllowSubstoreVisible: material.AllowSubstoreVisible,
    ```
  - ✅ 已完成

### 2. 测试验证

- [x] **代码语法检查**
  - 需求：验证代码修改是否有语法错误
  - 操作：使用 `go vet` 检查
  - 预计时间：0.25 小时
  - 负责人：weifashi
  - ✅ 已完成（无错误）

- [ ] **手动验证 - 测试环境**
  - 需求：在测试环境验证接口返回
  - 环境：ttpos-test1.ttpos.com
  - 验证步骤：
    1. 准备测试数据（不同 allow_substore_visible 值的物料）
    2. 调用接口 `/api/v1/shop/material/detail?uuid={uuid}`
    3. 验证返回中包含 `allow_substore_visible` 字段
    4. 验证字段值与数据库一致
  - 预计时间：0.5 小时
  - 负责人：
  - 测试脚本：`./test_bug_251127_005.sh`

- [ ] **前端联调验证**（如果前端需要此字段）
  - 需求：确认前端能正确读取和使用该字段
  - 验证点：
    - ✅ 前端接口调用正常
    - ✅ 字段值正确显示
    - ✅ 门店可见性逻辑正常
  - 预计时间：1 小时
  - 负责人：

### 3. 文档更新

- [x] **更新 Swagger 文档**
  - 需求：重新生成 Swagger 文档
  - 操作：运行 `swag init` 命令
  - 验证：访问 `/swagger/index.html` 确认字段已显示
  - 预计时间：0.25 小时
  - 负责人：weifashi
  - ✅ 已完成

- [x] **记录到 Graphiti**
  - 需求：记录此 Bug 的修复经验
  - 内容：
    - 问题类型：响应 DTO 字段遗漏
    - 根本原因：Service 层构建响应时未填充字段
    - 修复方案：同步更新 DTO 结构体和 Service 层
    - 预防措施：Code Review 检查 DTO 完整性
  - 预计时间：0.5 小时
  - 负责人：weifashi
  - ✅ 已完成

### 4. 代码审查与部署

- [ ] **代码审查**
  - 需求：通过 Code Review
  - 审查点：
    - ✅ 字段位置合理
    - ✅ 注释清晰
    - ✅ 测试充分
  - 预计时间：0.5 小时
  - 负责人：

- [ ] **发布到测试环境**
  - 需求：部署并验证
  - 操作：
    1. 合并代码到 develop 分支
    2. 触发测试环境部署
    3. 烟雾测试
  - 预计时间：0.5 小时
  - 负责人：

- [ ] **发布到生产环境**
  - 需求：生产发布并监控
  - 操作：
    1. 合并代码到 main 分支
    2. 创建 release tag
    3. 触发生产环境部署
    4. 监控接口响应和错误率
  - 预计时间：1 小时
  - 负责人：

---

## 📊 任务统计

- **总任务数**：10
- **已完成**：5
- **进行中**：0
- **未开始**：5
- **完成率**：50%

---

## 🔗 相关链接

- **Bug 报告**：`bug.md`
- **修复方案**：`solution.md`
- **代码文件**：
  - `main/app/dto/resp/material_resp/material.go` (第 54-79 行)
  - `main/app/service/material.go` (第 428-455 行)
  - `main/app/model/material.go` (第 31 行 - 参考)
- **API 文档**：`/swagger/index.html#/商家端.物品管理/get_shop_material_detail`

---

## 📝 实施注意事项

### 字段位置

建议在 `Status` 字段之后添加 `AllowSubstoreVisible` 字段，保持字段顺序的逻辑性：

```go
Status               int                  `json:"status"`
AllowSubstoreVisible int                  `json:"allow_substore_visible"`  // 新增
Valuation            float64              `json:"valuation"`
```

### 测试数据准备

确保测试数据库中有不同可见性设置的物料：
```sql
-- 查询测试数据
SELECT uuid, code, name, allow_substore_visible 
FROM material 
WHERE uuid IN (3699861597323265, ...);

-- 如需修改测试数据
UPDATE material SET allow_substore_visible = 1 WHERE uuid = 3699861597323265;
UPDATE material SET allow_substore_visible = 0 WHERE uuid = 3699861597323266;
```

### 验证命令

```bash
# 测试环境验证
curl -X GET \
  'https://ttpos-test1.ttpos.com/api/v1/shop/material/detail?uuid=3699861597323265' \
  -H 'Authorization: Bearer {token}' \
  | jq '.data.allow_substore_visible'

# 预期输出：1 或 0
```

---

**创建时间**：2025-11-27  
**最后更新**：2025-11-27  
**状态**：🟡 规划中

