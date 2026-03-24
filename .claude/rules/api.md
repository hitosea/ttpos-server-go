---
description: API 设计规范 — 编写 API 路由、请求/响应结构时自动加载
globs:
  - "main/app/api/**/*.go"
  - "main/app/dto/**/*.go"
---

# API 规范

## URL 命名

snake_case，名词单数 + 操作后缀：`/api/v1/order_info`

## HTTP 方法

| 方法 | 用途 | 参数解析 | Req Tag |
|------|------|----------|---------|
| GET | 查询、列表、详情 | `ShouldBindQuery` | `form` |
| POST | 创建/修改数据 | `ShouldBindJSON` | `json` |
| DELETE | 删除数据 | `ShouldBindJSON` | `json` |
| PUT | **禁止使用** | - | - |

## 响应格式

```json
{"code": 0, "message": "success", "data": {}}
```

- `code=0` 表示成功（`constant.CodeSuccess = 0`）
- `data` 必须是对象，不能是 null 或数组
- 响应体中的切片必须用 `make` 初始化，避免返回 null
