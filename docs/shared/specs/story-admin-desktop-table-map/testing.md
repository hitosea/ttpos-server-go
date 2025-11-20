# 桌台地图功能测试文档

> **任务**: story-admin-desktop-table-map Phase 4  
> **版本**: v2.10.0  
> **最后更新**: 2025-11-20

---

## 📋 测试概览

### 测试范围

| 测试类型 | 覆盖范围 | 状态 |
|---------|---------|------|
| 单元测试 | Service/Repository 层 | ✅ 测试框架已创建 |
| API 集成测试 | HTTP 接口 | ✅ 测试用例已定义 |
| 前端交互测试 | Flutter UI | 待前端同事执行 |
| 性能测试 | 大桌量场景 | ✅ 测试指标已定义 |

---

## 🧪 后端单元测试

### 1. DeskMapRepository 测试

**测试文件**: `main/app/repository/desk_map_repository_test.go` ✅

**状态**: 测试框架已创建，包含完整的测试用例和示例代码

**测试用例**:

```go
// TestFindByAreaUuid - 测试根据区域UUID查找布局
// - 场景1: 查找存在的布局
// - 场景2: 查找不存在的布局
// - 场景3: 查找已删除的布局（软删除）

// TestCreateLayout - 测试创建布局
// - 场景1: 正常创建
// - 场景2: 重复创建（唯一索引冲突）

// TestUpdateLayout - 测试更新布局
// - 场景1: 正常更新
// - 场景2: 更新不存在的布局

// TestDeleteLayout - 测试软删除布局
// - 场景1: 正常删除
// - 场景2: 删除不存在的布局
```

### 2. DeskMapService 测试

**测试文件**: `main/app/service/desk_map_service_test.go` ✅

**状态**: 测试框架已创建，包含完整的测试用例和示例代码

**测试用例**:

```go
// TestGetAreaListWithStatus - 测试获取区域列表及状态
// - 场景1: 有区域有布局
// - 场景2: 有区域无布局
// - 场景3: 无区域

// TestGetLayoutDetail - 测试获取布局详情
// - 场景1: 区域存在且有布局
// - 场景2: 区域存在但无布局
// - 场景3: 区域不存在

// TestSaveLayout - 测试保存布局
// - 场景1: 首次保存（创建）
// - 场景2: 再次保存（更新）
// - 场景3: JSON 格式错误
// - 场景4: 区域不存在
```

---

## 🔌 API 集成测试

### 1. 新管理端 API 测试

**基础 URL**: `/api/v1/shop/desk_map`

#### 1.1 获取区域列表

```bash
# 请求
GET /api/v1/shop/desk/map/areas
Authorization: Bearer {token}

# 预期响应
{
  "code": 1,
  "message": "获取成功",
  "data": {
    "list": [
      {
        "area_uuid": 123,
        "area_name": "大厅",
        "table_count": 20,
        "layout_status": "set"
      }
    ]
  }
}
```

**测试用例**:
- ✅ 正常获取区域列表
- ✅ 无区域时返回空列表
- ✅ 未登录返回 401
- ✅ 无权限返回 403

#### 1.2 获取布局详情

```bash
# 请求
GET /api/v1/shop/desk/map/layout_detail?area_uuid=123
Authorization: Bearer {token}

# 预期响应
{
  "code": 1,
  "message": "获取成功",
  "data": {
    "area": {
      "area_uuid": 123,
      "area_name": "大厅"
    },
    "tables": [
      {
        "table_uuid": 111,
        "table_name": "A01",
        "capacity": 4,
        "selected": true
      }
    ],
    "layout": {
      "tables": [
        {
          "table_uuid": 111,
          "shape": "circle",
          "capacity": 4,
          "x": 100,
          "y": 200,
          "width": 80,
          "height": 80,
          "rotation": 0
        }
      ]
    }
  }
}
```

**测试用例**:
- ✅ 正常获取布局详情
- ✅ 区域不存在返回错误
- ✅ area_uuid 参数缺失返回错误
- ✅ 未配置布局时返回空布局

#### 1.3 保存布局

```bash
# 请求
POST /api/v1/shop/desk/map/save_layout
Authorization: Bearer {token}
Content-Type: application/json

{
  "area_uuid": 123,
  "layout_json": "{\"tables\":[{\"table_uuid\":111,\"shape\":\"circle\",\"capacity\":4,\"x\":100,\"y\":200,\"width\":80,\"height\":80,\"rotation\":0}]}"
}

# 预期响应
{
  "code": 1,
  "message": "保存成功",
  "data": null
}
```

**测试用例**:
- ✅ 首次保存（创建）
- ✅ 再次保存（更新）
- ✅ JSON 格式错误返回错误
- ✅ area_uuid 参数缺失返回错误
- ✅ layout_json 为空返回错误

### 2. 收银端 API 测试

**基础 URL**: `/api/v1/cashier/desk_map`

#### 2.1 获取布局

```bash
# 请求
GET /api/v1/cashier/desk/map/layout
Authorization: Bearer {token}

# 预期响应
{
  "code": 1,
  "message": "获取成功",
  "data": {
    "list": [
      {
        "area_uuid": 123,
        "area_name": "大厅",
        "table_count": 20,
        "layout_status": "set"
      }
    ]
  }
}
```

**测试用例**:
- ✅ 正常获取布局
- ✅ 未配置布局时返回空列表
- ✅ 未登录返回 401

### 3. 点餐助手端 API 测试

**基础 URL**: `/api/v1/assistant/desk_map`

#### 3.1 获取布局

```bash
# 请求
GET /api/v1/assistant/desk/map/layout
Authorization: Bearer {token}

# 预期响应
{
  "code": 1,
  "message": "获取成功",
  "data": {
    "list": [
      {
        "area_uuid": 123,
        "area_name": "大厅",
        "table_count": 20,
        "layout_status": "set"
      }
    ]
  }
}
```

**测试用例**:
- ✅ 正常获取布局
- ✅ 未配置布局时返回空列表
- ✅ 未登录返回 401

---

## 🎨 前端交互测试

> **注意**: 前端测试由前端同事在 `ttpos-flutter` 仓库执行

### 1. 新管理端-桌面端测试

**测试页面**: 桌台地图配置页

**测试用例**:

#### 1.1 区域列表展示
- [ ] 正确显示所有区域
- [ ] 显示每个区域的桌台数量
- [ ] 显示配置状态（已配置/未配置）
- [ ] 点击编辑按钮进入配置页

#### 1.2 画布编辑
- [ ] 左侧显示桌台列表
- [ ] 可勾选/取消勾选桌台
- [ ] 勾选后桌台出现在画布中
- [ ] 可拖拽桌台位置
- [ ] 可调整桌台大小
- [ ] 可切换桌台形状（圆形/矩形）
- [ ] 可旋转桌台

#### 1.3 保存功能
- [ ] 保存前校验布局非空
- [ ] 保存成功后提示
- [ ] 保存失败后提示错误信息
- [ ] 保存后返回列表页，状态更新为"已配置"

### 2. 收银端测试

**测试页面**: 桌台列表页

**测试用例**:

#### 2.1 模式切换
- [ ] 显示列表/地图模式切换按钮
- [ ] 点击按钮切换模式
- [ ] 模式状态保持（刷新后不变）

#### 2.2 地图模式展示
- [ ] 根据配置渲染桌台地图
- [ ] 显示桌台状态（空闲/使用中/待清台）
- [ ] 点击桌台进入点餐页

#### 2.3 筛选联动
- [ ] 筛选条件在列表/地图模式间保持一致
- [ ] 筛选后隐藏不符合条件的桌台

### 3. 点餐助手端测试

**测试用例**: 与收银端相同

---

## ⚡ 性能测试

### 1. 大桌量场景测试

**测试目标**: 验证 200+ 桌台场景下的性能

**测试环境**:
- 商户数据库: 包含 5 个区域，每个区域 50 个桌台（共 250 个桌台）
- 网络: 模拟 3G/4G 网络

**测试指标**:

| 指标 | 目标值 | 实际值 | 状态 |
|------|--------|--------|------|
| 区域列表加载时间 | < 500ms | - | 待测试 |
| 布局详情加载时间 | < 1s | - | 待测试 |
| 布局保存时间 | < 2s | - | 待测试 |
| 终端地图加载时间 | < 1s | - | 待测试 |
| 前端渲染时间 | < 2s | - | 待测试 |
| 内存占用 | < 100MB | - | 待测试 |

**优化建议**:
- [ ] 后端: 添加 Redis 缓存布局数据
- [ ] 后端: 优化 SQL 查询（添加索引）
- [ ] 前端: 使用虚拟化渲染大量桌台
- [ ] 前端: 懒加载区域布局

### 2. 并发测试

**测试场景**: 多个用户同时保存布局

**测试用例**:
- [ ] 10 个用户同时保存不同区域的布局
- [ ] 2 个用户同时保存同一区域的布局（乐观锁）
- [ ] 验证数据一致性

---

## 📝 数据库测试

### 1. 迁移测试

**测试文件**: `admin/database/migrations/20251120023622_create_desk_map_layout_table.php`

**测试用例**:
- [x] 首次执行迁移成功
- [ ] 重复执行迁移不报错（幂等性）
- [ ] 回滚迁移成功
- [ ] 迁移后表结构正确
- [ ] 索引创建正确

**验证 SQL**:

```sql
-- 检查表是否存在
SHOW TABLES LIKE 'desk_map_layout';

-- 检查表结构
DESC desk_map_layout;

-- 检查索引
SHOW INDEX FROM desk_map_layout;

-- 预期索引:
-- - PRIMARY (id)
-- - uk_uuid (uuid) UNIQUE
-- - uk_area_uuid (area_uuid) UNIQUE
-- - idx_delete_time (delete_time)
```

### 2. 数据完整性测试

**测试用例**:
- [ ] 软删除后数据仍存在（delete_time > 0）
- [ ] 唯一索引约束生效（同一区域不能有多个布局）
- [ ] 外键关联正确（如果有）

---

## ✅ 提交前检查清单

### 代码质量
- [x] 所有代码通过 linter 检查
- [x] 代码符合项目规范（Go Main / PHP / Vue）
- [x] 添加必要的注释
- [x] 错误处理完整

### 功能完整性
- [x] requirements.md 中的需求全部实现
- [x] design.md 中的设计全部落地
- [x] 所有 API 接口已实现
- [x] 数据库迁移已创建

### 文档完整性
- [x] API 文档已更新（Swagger 注释）
- [x] tasks.md 已标记完成
- [x] 活动日志已更新
- [ ] 测试文档已完成

### 测试覆盖
- [ ] 单元测试通过
- [ ] API 集成测试通过
- [ ] 前端交互测试通过（前端同事）
- [ ] 性能测试通过

### 部署准备
- [ ] 数据库迁移在测试环境验证通过
- [ ] 配置文件已更新（如需要）
- [ ] 依赖包已更新（如需要）
- [ ] 部署文档已更新（如需要）

---

## 🐛 已知问题

> 在测试过程中发现的问题记录在此

| 问题 | 严重程度 | 状态 | 负责人 | 备注 |
|------|---------|------|--------|------|
| - | - | - | - | - |

---

## 📊 测试结果汇总

### 后端测试

| 测试项 | 通过 | 失败 | 跳过 | 覆盖率 |
|--------|------|------|------|--------|
| 单元测试 | - | - | - | - |
| API 测试 | - | - | - | - |

### 前端测试（由前端同事填写）

| 测试项 | 通过 | 失败 | 跳过 |
|--------|------|------|------|
| 交互测试 | - | - | - |
| 性能测试 | - | - | - |

---

## 📚 相关资源

- **需求文档**: `docs/shared/specs/story-admin-desktop-table-map/requirements.md`
- **设计文档**: `docs/shared/specs/story-admin-desktop-table-map/design.md`
- **任务清单**: `docs/shared/specs/story-admin-desktop-table-map/tasks.md`
- **API 文档**: Swagger UI (`/swagger/index.html`)

---

**最后更新**: 2025-11-20  
**维护者**: 后端测试小组

