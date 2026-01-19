# 新管理端-调拨单-优化 设计文档

> 本文档定义调拨单管理界面优化功能的技术设计和实现方案。

## 📋 概述

本功能在现有调拨单管理界面基础上，增加"我发出/我接收/我审核"三种角色视角的筛选功能，让门店用户能够快速定位与自己相关的调拨单。

**核心改动**：
- 前端：新增类型筛选器组件，调整列表查询逻辑
- 后端：调整列表接口，增加 `role_type` 参数和审核节点判断逻辑

**技术栈**：Go (main/) + Vue 3 (admin/views/)

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

本设计遵循以下规范：

- ✅ Service 只依赖其他 Service 接口（如需要）
- ✅ Repository 只持有 db 实例
- ✅ URL 使用 snake_case
- ✅ data 字段必须是对象
- ✅ 不使用 panic，返回 error
- ✅ 使用 errors.WithMessage 包装错误

### API 设计规范 (api.mdc)

- ✅ URL: `/api/v1/shop/transfer_order/list`
- ✅ 响应格式: `{code, message, data{}, meta{}}`
- ✅ data 不能为 null 或数组
- ✅ 分页信息放在 meta 中

### Vue 规范 (vue.mdc)

- ✅ 使用 Vue 3 + TypeScript + Vite
- ✅ 使用 Element Plus 组件库
- ✅ 使用 Composition API
- ✅ 遵循命名规范

---

## 🔄 代码复用分析

### 可复用的现有代码

本功能主要是在现有调拨单功能基础上进行优化，可以直接复用以下代码：

1. **Transfer Order Repository**: `main/app/repository/transfer_order.go`
   - 复用现有的 `GetList()` 方法和选项模式
   - 新增选项方法：`WhereCurrentApprovalNodeShopUuid()`

2. **Transfer Order Service**: `main/app/service/transfer_order/transfer_order.go`
   - 复用现有的列表查询逻辑
   - 新增审核节点判断逻辑

3. **Transfer Order API**: `main/app/api/v1/shop/shop_transfer.go`
   - 复用现有的列表接口
   - 新增 `role_type` 参数处理

4. **Transfer Order DTO**: `main/app/dto/req/transfer_order.go`
   - 扩展 `TransferOrderListReq` 结构体
   - 新增 `RoleType` 字段

### 集成点

- **现有 API**: 在现有 `/api/v1/shop/transfer_order/list` 接口基础上扩展
- **数据库表**: 使用现有 `ttpos_transfer_order` 表，不需要新增表
- **前端组件**: 在现有调拨单列表页面基础上扩展

---

## 🏗️ 架构设计

### 分层设计原则

**Go Main 三层架构**:

```
API 层 (Controller/API)
  ↓ 接收请求参数 role_type
Service 层 (Service)
  ↓ 判断审核节点逻辑
Repository 层 (Repository)
  ↓ 根据 role_type 筛选数据
```

**依赖规则**:

- ✅ API → Service → Repository
- ❌ 禁止跨层调用
- ✅ Service 可以依赖其他 Service 接口（如需要）

### 模块划分

#### Go Main 模块

- **API 层**: `main/app/api/v1/shop/shop_transfer.go` - 扩展列表接口
- **Service 层**: `main/app/service/transfer_order/transfer_order.go` - 新增审核节点判断逻辑
- **Repository 层**: `main/app/repository/transfer_order.go` - 新增选项方法
- **DTO 层**: `main/app/dto/req/transfer_order.go` - 扩展请求参数

#### Vue 前端模块

- **Pages**: `admin/views/shop/pages/transfer-order/index.vue` - 扩展列表页面
- **Components**: `admin/views/shop/components/transfer-order/RoleTypeFilter.vue` - 新增筛选器组件
- **API**: `admin/views/shop/api/transfer-order.ts` - 扩展 API 调用

---

## 🗄️ 数据库设计

### 使用现有表

本功能使用现有 `ttpos_transfer_order` 表，不需要新增表。

**相关字段**：
- `transfer_out_shop_uuid`: 调出门店 UUID（用于"我发出"筛选）
- `transfer_in_shop_uuid`: 调入门店 UUID（用于"我接收"筛选）
- `current_approval_node_shop_uuid`: 当前审核节点门店 UUID（用于"我审核"筛选）
- `status`: 单据状态（1=待提交, 2=待审核, 3=已驳回, 4=待收货, 5=已完成）

### 索引优化

为提升查询性能，建议增加以下索引：

```sql
-- 索引1：优化"我发出"查询
ALTER TABLE `ttpos_transfer_order` 
ADD INDEX `idx_transfer_out_shop_uuid_status` (`transfer_out_shop_uuid`, `status`);

-- 索引2：优化"我接收"查询
ALTER TABLE `ttpos_transfer_order` 
ADD INDEX `idx_transfer_in_shop_uuid_status` (`transfer_in_shop_uuid`, `status`);

-- 索引3：优化"我审核"查询
ALTER TABLE `ttpos_transfer_order` 
ADD INDEX `idx_current_approval_node_shop_uuid_status` (`current_approval_node_shop_uuid`, `status`);
```

**迁移文件**: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_transfer_order_role_type_indexes.php`

---

## 📊 数据模型

### DTO 扩展

#### Request DTO

```go
// main/app/dto/req/transfer_order.go
type TransferOrderListReq struct {
    PageNo   int    `json:"page_no" binding:"required"`
    PageSize int    `json:"page_size" binding:"required"`
    Status   int    `json:"status"`                    // 状态筛选
    RoleType string `json:"role_type"`                 // 新增：角色类型（sender/receiver/approver）
    Search   string `json:"search"`                    // 搜索关键词
}
```

**RoleType 枚举值**：
- `sender`: 我发出（本店作为调出方）
- `receiver`: 我接收（本店作为调入方）
- `approver`: 我审核（本店作为审核方）
- 空值: 全部单据（不筛选）

---

## 🔌 API 设计

### RESTful API

#### API: 调拨单列表（扩展）

**请求**:

- **URL**: `/api/v1/shop/transfer_order/list`
- **Method**: `POST`
- **Headers**:
  ```json
  {
    "Authorization": "Bearer {token}",
    "Content-Type": "application/json"
  }
  ```
- **Body**:
  ```json
  {
    "page_no": 1,
    "page_size": 20,
    "status": 2,               // 可选，状态筛选
    "role_type": "receiver",   // 新增：角色类型筛选
    "search": "门店名称"        // 可选，搜索关键词
  }
  ```

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "uuid": 123456,
        "order_no": "TO202601060001",
        "transfer_out_shop_name": "总店",
        "transfer_in_shop_name": "分店A",
        "status": 2,
        "status_name": "待审核",
        "current_approval_node_name": "接收店审核",
        "product_count": 10,
        "total_amount": "1000.00",
        "create_time": 1704528000,
        "update_time": 1704528000
      }
    ]
  },
  "meta": {
    "page_no": 1,
    "page_size": 20,
    "total": 50
  }
}
```

---

## 🧩 组件和接口

### Repository 层

#### 新增选项方法

```go
// main/app/repository/transfer_order.go

// WhereCurrentApprovalNodeShopUuid 筛选当前审核节点为指定门店的单据
func (r *TransferOrderRepoImpl) WhereCurrentApprovalNodeShopUuid(shopUuid uint64) DBOption {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("current_approval_node_shop_uuid = ?", shopUuid)
    }
}
```

### Service 层

#### 新增审核节点判断逻辑

```go
// main/app/service/transfer_order/transfer_order.go

// GetListByRoleType 根据角色类型获取调拨单列表
func (s *transferOrderSrv) GetListByRoleType(ctx *gin.Context, req *dto_req.TransferOrderListReq) (*dto_resp.TransferOrderListResp, error) {
    shopUuid := ctx.GetShopUuid() // 获取当前门店 UUID
    
    // 构建查询选项
    options := []repository.DBOption{
        s.transferOrderRepo.PageLimit(req.PageNo, req.PageSize),
    }
    
    // 根据角色类型添加筛选条件
    switch req.RoleType {
    case "sender":
        // 我发出：本店作为调出方
        options = append(options, s.transferOrderRepo.WhereTransferOutShopUuid(shopUuid))
    case "receiver":
        // 我接收：本店作为调入方
        options = append(options, s.transferOrderRepo.WhereTransferInShopUuid(shopUuid))
    case "approver":
        // 我审核：当前审核节点为本店 且 状态为待审核
        options = append(options,
            s.transferOrderRepo.WhereCurrentApprovalNodeShopUuid(shopUuid),
            s.transferOrderRepo.WhereStatus(2), // 2=待审核
        )
    }
    
    // 状态筛选（我审核视图强制为待审核）
    if req.RoleType != "approver" && req.Status > 0 {
        options = append(options, s.transferOrderRepo.WhereStatus(req.Status))
    }
    
    // 搜索关键词
    if req.Search != "" {
        options = append(options, s.transferOrderRepo.Search(req.Search))
    }
    
    // 查询列表
    list, total, err := s.transferOrderRepo.GetList(options...)
    if err != nil {
        return nil, errors.WithMessage(err, "查询调拨单列表失败")
    }
    
    // 转换为响应 DTO
    respList := make([]*dto_resp.TransferOrderResp, 0, len(list))
    for _, item := range list {
        respList = append(respList, s.convertToResp(item))
    }
    
    return &dto_resp.TransferOrderListResp{
        List: respList,
        Meta: &dto_resp.PageMeta{
            PageNo:   req.PageNo,
            PageSize: req.PageSize,
            Total:    total,
        },
    }, nil
}
```

### API 层

#### 扩展列表接口

```go
// main/app/api/v1/shop/shop_transfer.go

// List 调拨单列表
func (api *ShopTransferAPI) List(c *gin.Context) {
    var req dto_req.TransferOrderListReq
    if err := c.ShouldBindJSON(&req); err != nil {
        helper.ErrorWithDetail(c, constant.CodeInvalidParam, err)
        return
    }
    
    // 调用 Service 层
    resp, err := api.transferOrderSrv.GetListByRoleType(c, &req)
    if err != nil {
        helper.ErrorWithDetail(c, constant.CodeFail, err)
        return
    }
    
    helper.SuccessWithData(c, resp)
}
```

### 前端组件

#### 类型筛选器组件

```vue
<!-- admin/views/shop/components/transfer-order/RoleTypeFilter.vue -->
<template>
  <div class="role-type-filter">
    <el-radio-group v-model="selectedType" @change="handleChange">
      <el-radio-button label="">全部</el-radio-button>
      <el-radio-button label="sender">我发出</el-radio-button>
      <el-radio-button label="receiver">我接收</el-radio-button>
      <el-radio-button label="approver">我审核</el-radio-button>
    </el-radio-group>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

interface Props {
  modelValue?: string
}

interface Emits {
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: 'receiver' // 默认选中"我接收"
})

const emit = defineEmits<Emits>()

const selectedType = ref(props.modelValue)

watch(() => props.modelValue, (val) => {
  selectedType.value = val
})

const handleChange = (value: string) => {
  emit('update:modelValue', value)
  emit('change', value)
}
</script>
```

---

## ⚡ 缓存设计

### Redis 缓存

**缓存策略**:

- **Key 命名**: `ttpos:transfer_order:list:{shop_uuid}:{role_type}:{status}:{page_no}`
- **过期时间**: 60 秒
- **更新策略**: Cache-Aside Pattern（缓存未命中时查询数据库）

**实现**:

```go
// 缓存读取
key := fmt.Sprintf("ttpos:transfer_order:list:%d:%s:%d:%d", 
    shopUuid, req.RoleType, req.Status, req.PageNo)
cached, err := redis.Get(key)
if err == nil {
    // 缓存命中
    return cached
}

// 缓存未命中，查询数据库
data, err := repo.GetList(options...)
if err != nil {
    return err
}

// 写入缓存
redis.Set(key, data, 60*time.Second)
return data
```

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 审核节点判断失败

- **处理方式**: 记录错误日志，返回空列表
- **用户影响**: 用户看到"当前无待审核单据"
- **代码示例**:
  ```go
  if err != nil {
      logger.Logger.Error("查询调拨单列表失败", zap.Error(err))
      return &dto_resp.TransferOrderListResp{
          List: []*dto_resp.TransferOrderResp{},
          Meta: &dto_resp.PageMeta{PageNo: req.PageNo, PageSize: req.PageSize, Total: 0},
      }, nil
  }
  ```

#### 场景 2: 参数验证失败

- **处理方式**: 返回参数错误提示
- **用户影响**: 用户看到具体的参数错误信息

---

## 🔒 安全设计

### 身份验证

- **JWT Token**: 所有 API 需要 Token 验证
- **门店权限**: 验证当前用户属于哪个门店，只能查看本店相关的调拨单

### 权限控制

- **门店权限**: 只能查看本店作为调出方、调入方或审核方的调拨单
- **角色权限**: 如需要，可以根据用户角色限制某些操作（如只有管理员才能查看"我审核"视图）

### 数据安全

- **SQL 注入防护**: 使用参数化查询（GORM）
- **XSS 防护**: 前端输入校验

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:

- Repository: 80%+
- Service: 70%+

**测试内容**:

1. **Repository 测试**:
   - 测试新增的选项方法 `WhereCurrentApprovalNodeShopUuid()`
   - 测试组合选项的正确性

2. **Service 测试**:
   - 测试"我发出"视图查询逻辑
   - 测试"我接收"视图查询逻辑
   - 测试"我审核"视图查询逻辑（四种审核场景）
   - 测试边界场景（如门店无上级）

### 集成测试

**测试流程**:

1. 创建测试数据（不同角色的调拨单）
2. 测试三种角色视角的列表查询
3. 验证审核节点判断逻辑
4. 验证状态筛选联动

---

## 📈 性能优化

### 优化策略

1. **数据库优化**:
   - 添加索引（见"数据库设计"章节）
   - 使用选项模式优化查询

2. **缓存优化**:
   - Redis 缓存列表查询结果（60秒）
   - 缓存 key 设计合理，避免缓存穿透

3. **查询优化**:
   - 减少不必要的 JOIN 查询
   - 使用索引字段筛选

### 性能指标

- 本地响应时间: < 200ms
- 数据库查询: < 50ms
- 缓存命中率: > 80%

---

## 📚 实现清单

### Phase 1: 后端实现

- [ ] 扩展 TransferOrderListReq DTO（新增 RoleType 字段）
- [ ] Repository 层新增选项方法
- [ ] Service 层新增审核节点判断逻辑
- [ ] API 层扩展列表接口
- [ ] 编写单元测试

### Phase 2: 前端实现

- [ ] 创建类型筛选器组件
- [ ] 扩展列表页面（集成筛选器）
- [ ] 扩展 API 封装（新增 role_type 参数）
- [ ] 实现状态筛选器联动逻辑

### Phase 3: 数据库优化

- [ ] 创建索引迁移文件
- [ ] 执行索引迁移
- [ ] 性能测试验证

### Phase 4: 测试和优化

- [ ] 单元测试
- [ ] 集成测试
- [ ] 性能测试
- [ ] 缓存优化

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/weifashi/2026-01/2026-01-06.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-06  
**作者**: weifashi  
**审核者**: 待指定

