# 外卖状态管理 API 文档

> 本文档记录外卖状态管理相关的 API 接口。

## 概述

外卖状态管理 API 提供以下功能：

- 获取指定平台外卖状态
- 获取所有平台外卖状态
- 切换指定平台外卖状态
- 更新指定平台菜单数据

## 接口列表

### 1. 获取指定平台外卖状态

**接口地址**: `GET /api/v1/shop/takeout/status/{platform}`

**功能描述**: 获取指定外卖平台的状态信息

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| platform | string | 是 | 外卖平台标识 (grab/lineman/foodpanda/shopeefood) |

**请求头**:

```
Authorization: Bearer {token}
Content-Type: application/json
```

**响应示例**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "platform": "grab",
    "enabled": true,
    "menu": {
      "categories": [...],
      "items": [...]
    },
    "updatedAt": 1734268800
  }
}
```

**错误响应**:

```json
{
  "code": 0,
  "message": "平台不存在或未配置",
  "data": {}
}
```

### 2. 获取所有平台外卖状态

**接口地址**: `GET /api/v1/shop/takeout/status`

**功能描述**: 获取所有已配置外卖平台的状态信息

**请求参数**: 无

**请求头**:

```
Authorization: Bearer {token}
Content-Type: application/json
```

**响应示例**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "platform": "grab",
        "enabled": true,
        "menu": {...},
        "updatedAt": 1734268800
      },
      {
        "platform": "lineman",
        "enabled": false,
        "menu": null,
        "updatedAt": 1734268800
      }
    ]
  }
}
```

### 3. 切换指定平台外卖状态

**接口地址**: `PUT /api/v1/shop/takeout/status/{platform}`

**功能描述**: 开启或关闭指定外卖平台的功能

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| platform | string | 是 | 外卖平台标识 |

**请求体**:

```json
{
  "enabled": true,
  "menu": {
    "categories": [...],
    "items": [...]
  }
}
```

**请求参数说明**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| enabled | boolean | 是 | 是否开启外卖功能 |
| menu | object | 否 | 菜单数据（JSON格式） |

**请求头**:

```
Authorization: Bearer {token}
Content-Type: application/json
```

**响应示例**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "platform": "grab",
    "enabled": true,
    "menu": {...},
    "updatedAt": 1734268800
  }
}
```

### 4. 更新指定平台菜单数据

**接口地址**: `PUT /api/v1/shop/takeout/menu/{platform}`

**功能描述**: 更新指定外卖平台的菜单数据

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| platform | string | 是 | 外卖平台标识 |

**请求体**:

```json
{
  "categories": [
    {
      "id": "category_001",
      "name": "主菜",
      "items": [
        {
          "id": "item_001",
          "name": "宫保鸡丁",
          "price": 25.00,
          "description": "经典川菜"
        }
      ]
    }
  ]
}
```

**请求头**:

```
Authorization: Bearer {token}
Content-Type: application/json
```

**响应示例**:

```json
{
  "code": 1,
  "message": "菜单数据更新成功",
  "data": {}
}
```

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 0 | 请求失败 |
| 1 | 请求成功 |
| 1001 | 参数错误 |
| 1002 | 权限不足 |
| 1003 | 平台不支持 |
| 1004 | 数据验证失败 |

## 数据格式说明

### 平台标识

支持的平台标识：

- `grab`: Grab 外卖
- `lineman`: Lineman 外卖
- `foodpanda`: Foodpanda 外卖
- `shopeefood`: ShopeeFood 外卖

### 菜单数据格式

菜单数据为 JSON 格式，包含分类和商品信息：

```json
{
  "categories": [
    {
      "id": "category_id",
      "name": "分类名称",
      "description": "分类描述",
      "sort": 1,
      "items": [
        {
          "id": "item_id",
          "name": "商品名称",
          "price": 25.50,
          "description": "商品描述",
          "image": "image_url",
          "available": true
        }
      ]
    }
  ]
}
```

## 缓存策略

- 单个平台状态缓存：5分钟
- 所有平台状态缓存：5分钟
- 状态变更时自动清理相关缓存

## 权限要求

所有接口都需要有效的 JWT Token 认证，并且用户需要有相应的店铺管理权限。

## 更新历史

- v1.0.0 (2025-12-13): 初始版本，支持基本的平台状态管理和菜单数据更新
