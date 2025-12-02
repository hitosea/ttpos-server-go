# Bug-251127-005: 物料详情接口缺失 allow_substore_visible 字段

> ✅ **已解决** - 此 Bug 已在 v2.10 中修复。
>
> - 解决时间: 2025-12-01
> - 解决者: weifashi
> - 验证状态: ✅ 已验证

## 基本信息

| 字段       | 值                                           |
| ---------- | -------------------------------------------- |
| Bug ID     | bug-251127-005                               |
| 模块       | material（物料管理）                         |
| 严重程度   | medium                                       |
| 发现版本   | v2.10.9                                      |
| 发现日期   | 2025-11-27                                   |
| 发现者     | weifashi                                     |
| 状态       | 🔵 已修复                                    |
| 解决版本   | v2.10                                        |
| 解决日期   | 2025-12-01                                   |
| 解决者     | weifashi                                     |
| 影响终端   | shop（店铺后台）                             |

## 问题描述

### 现象

调用物料详情接口 `/api/v1/shop/material/detail?uuid=3699861597323265` 时，返回的数据中缺失 `allow_substore_visible` 字段。

### 复现步骤

1. 访问测试环境：`https://ttpos-test1.ttpos.com`
2. 调用接口：`GET /api/v1/shop/material/detail?uuid=3699861597323265`
3. 观察返回结果

### 预期行为

接口应该返回 `allow_substore_visible` 字段，用于标识物料是否允许门店可见。

**预期返回示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "uuid": "3699861597323265",
    "name": "...",
    "allow_substore_visible": true,  // 应该包含此字段
    "..."
  }
}
```

### 实际行为

返回数据中没有 `allow_substore_visible` 字段。

**实际返回**：
```json
{
  "code": -102,
  "message": "token is required",
  "data": {}
}
```

**注意**：实际测试时返回了 token 错误，需要先解决认证问题后再验证字段是否缺失。

## 环境信息

- **环境**: 测试环境（ttpos-test1.ttpos.com）
- **接口**: `/api/v1/shop/material/detail`
- **请求方式**: GET
- **技术栈**: Go + Gin (Main 模块)

## 影响范围

### 影响终端
- **shop**（店铺后台）：物料管理相关功能

### 影响模块
- **Main 模块**: `main/app/api/v1/shop/shop_material.go`
- **Service 层**: `main/app/service/material.go`
- **DTO 响应**: `main/app/dto/resp/material_resp/material.go`

### 业务影响
- 门店无法正确判断物料的可见性配置
- 可能导致前端显示逻辑错误
- 影响物料权限控制功能

## 初步分析

### 代码位置
- **API 入口**: `main/app/api/v1/shop/shop_material.go` - `Detail` 方法
- **业务逻辑**: `main/app/service/material.go`
- **响应结构**: `main/app/dto/resp/material_resp/material.go`

### 可能原因
1. DTO 响应结构体中缺少 `AllowSubstoreVisible` 字段定义
2. Service 层查询数据时未包含该字段
3. 数据库表中可能缺少该字段（需确认）
4. 字段映射时未正确转换

### 修复方向
1. 检查 `material_resp.MaterialDetail` 结构体定义
2. 检查数据库表 `material` 是否包含 `allow_substore_visible` 字段
3. 确认 Service 层查询逻辑是否包含该字段
4. 添加或修复字段映射

## 相关链接

- **代码文件**:
  - `main/app/api/v1/shop/shop_material.go`
  - `main/app/service/material.go`
  - `main/app/dto/resp/material_resp/material.go`

- **修复方案**: `solution.md`
- **任务清单**: `tasks.md` (50% 完成)
- **测试报告**: `test_report.md`
- **测试脚本**: `test_bug_251127_005.sh`

- **Graphiti 参考**: 暂无相似问题记录

- **相关 Spec**: （待补充）

## 下一步

1. ✅ **技术分析**：
   - 查看 `material_resp.MaterialDetail` 结构体定义
   - 检查数据库表结构
   - 确认 Service 层实现

2. **创建修复方案**：
   ```bash
   /bug-spec bug-251127-005-material-missing-allow-substore-visible
   ```

3. **测试验证**：
   - 补充认证 token 后重新测试接口
   - 验证字段是否正确返回

---

**创建时间**: 2025-11-27  
**最后更新**: 2025-11-27 16:36  
**状态**: 🟢 待测试（代码修复已完成）


---

## 经验总结

**问题类型**: 响应字段遗漏

**根本原因**: 在构建响应 DTO 时遗漏了字段。数据库 Model 层有该字段，但响应结构体 `MaterialDetailResp` 中缺少字段定义，Service 层构建响应时也未填充该字段。

**解决方案**: 
1. 在响应 DTO (`MaterialDetailResp`) 中添加 `AllowSubstoreVisible` 字段
2. 在 Service 层 (`GetMaterialDetail` 方法) 中填充该字段值
3. 更新 Swagger 文档（自动生成）

**关键步骤**:
1. 修改 `main/app/dto/resp/material_resp/material.go`，添加字段定义
2. 修改 `main/app/service/material.go`，填充字段值
3. 运行 `swag init` 更新 Swagger 文档
4. 测试验证字段返回正确

**预防措施**:
1. **Code Review 检查清单**: 新增字段时需同步更新 Model、请求 DTO、响应 DTO 和 Service 层
2. **自动化检测**: 编写 lint 规则检测 Model 与 DTO 字段一致性
3. **测试覆盖**: 增加 API 响应字段完整性测试
4. **文档同步**: 在需求评审时明确列出所有字段

**相关知识**:
- DTO (Data Transfer Object) 设计模式
- Go 结构体标签（struct tags）
- RESTful API 响应规范
- 字段映射和数据转换

**适用场景**:
- API 接口返回字段缺失的问题
- DTO 与 Model 不一致的问题
- 响应数据完整性问题

**参考资料**:
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/go-main.mdc` - Go 开发规范
- `main/app/dto/resp/material_resp/material.go` (Line 54-79)
- `main/app/service/material.go` (Line 428-455)
