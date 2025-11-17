# API 规范速查

> 📚 **受众**: 所有开发者  
> 📖 **用途**: API 设计规范快速参考

---

## URL 命名

### ✅ 使用 snake_case

```
/api/v1/user_profiles
/api/v1/order_items
/api/v1/payment_methods
```

### ❌ 不使用其他命名

```
/api/v1/userProfiles      # camelCase
/api/v1/user-profiles     # kebab-case
```

---

## HTTP 方法

| 方法 | 用途 | 示例 |
|------|------|------|
| GET | 获取资源 | `GET /api/v1/order/list` |
| POST | 创建/执行 | `POST /api/v1/order/create` |
| PUT | 完整更新 | `PUT /api/v1/order/update` |
| PATCH | 部分更新 | `PATCH /api/v1/order/update` |
| DELETE | 删除 | `DELETE /api/v1/order/delete` |

---

## 响应格式

### 统一格式

```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

### data 规则

- ✅ **必须是对象** `{}`
- ❌ **不能是 null**
- ❌ **不能是数组** `[]`

---

## 分页格式

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [...],
    "meta": {
      "page_no": 1,
      "page_size": 20,
      "total": 100
    }
  }
}
```

---

## 错误码

| code | 说明 |
|------|------|
| 1 | 成功 |
| 0 | 失败 |
| 400 | 参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 不存在 |
| 500 | 服务器错误 |

---

## 认证

```http
Authorization: Bearer {token}
```

---

## 请求头

```http
Content-Type: application/json
Authorization: Bearer xxx
Accept-Language: zh-CN
```

---

## 查询参数

```
GET /api/v1/order/list?page=1&page_size=20&status=1
```

---

**详细文档**: [API 设计指南](../../human/guides/api-design-guide.md)

---

**最后更新**: 2025-11-17  
**维护者**: TTPOS Team

