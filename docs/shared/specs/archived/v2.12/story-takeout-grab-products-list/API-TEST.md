# 外卖商品统计接口 - 测试脚本

## 接口信息

- **接口路径**: `GET /shop/takeout/products/count`
- **认证方式**: JWT Token (Bearer)
- **参数**:
  - `platform` (可选): 外卖平台标识,如 `grab`, `lineman`等,不传则统计所有平台
  - `force_refresh` (可选): 强制刷新缓存,值为 `1` 时强制刷新,默认 `0`

## 测试用例

### 测试1: 查询Grab平台商品数

```bash
curl -X GET "http://localhost:8080/shop/takeout/products/count?platform=grab" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json"
```

**预期响应**:
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "total": 150
  }
}
```

### 测试2: 查询所有平台商品数

```bash
curl -X GET "http://localhost:8080/shop/takeout/products/count" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type": "application/json"
```

**预期响应**:
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "total": 200
  }
}
```

### 测试3: 强制刷新缓存

```bash
curl -X GET "http://localhost:8080/shop/takeout/products/count?platform=grab&force_refresh=1" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json"
```

**预期行为**:
- 跳过缓存读取
- 直接查询数据库
- 更新缓存数据

### 测试4: 未授权访问(应返回401)

```bash
curl -X GET "http://localhost:8080/shop/takeout/products/count?platform=grab"
```

**预期响应**:
```json
{
  "code": 0,
  "message": "未授权",
  "data": {}
}
```

### 测试5: 缓存命中验证

```bash
# 第一次请求
time curl -X GET "http://localhost:8080/shop/takeout/products/count?platform=grab" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 第二次请求(应该更快,缓存命中)
time curl -X GET "http://localhost:8080/shop/takeout/products/count?platform=grab" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**预期行为**:
- 第一次请求: ~50ms (数据库查询)
- 第二次请求: ~5ms (缓存命中)

## 验证清单

- [x] Service层实现完成
  - [x] 接口方法添加
  - [x] 数据库查询逻辑
  - [x] 缓存读取逻辑
  - [x] 缓存写入逻辑
  - [x] 缓存清除方法
- [x] Handler层实现完成
  - [x] GetProductCount方法
  - [x] Swagger注释
  - [x] 路由注册
- [x] 代码质量
  - [x] 无linter错误
  - [x] 错误处理完整
  - [x] 日志记录规范

## 测试步骤

### 1. 启动服务

```bash
cd /home/coder/workspaces/ttpos-server-go/main
go run main.go
```

### 2. 获取测试Token

```bash
# 登录获取Token
curl -X POST "http://localhost:8080/shop/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test_user",
    "password": "test_password"
  }'
```

### 3. 执行测试用例

依次执行上述测试用例1-5,验证功能正确性。

### 4. 性能验证

- 查看日志确认缓存命中
- 使用 `time` 命令对比响应时间
- 验证缓存过期时间(5分钟)

## 预期结果

✅ 所有测试用例通过
✅ 响应格式正确
✅ 缓存机制正常工作
✅ 性能指标符合要求
✅ 错误处理正确

---

**创建日期**: 2025-12-18
**实现状态**: ✅ 已完成

