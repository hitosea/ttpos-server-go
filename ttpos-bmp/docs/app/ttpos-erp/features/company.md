# Company 服务功能说明

## 概述
Company 服务提供公司信息管理功能，包括公司基本信息查询、公司层级关系管理等。

## 服务接口

### ICompany - 公司管理

#### 公司查询
- **GetCompanyList**: 获取公司列表
- **GetCompany**: 根据公司名称获取公司详细信息
- **GetCompanyWithAbbr**: 根据公司简称获取公司信息
- **GetCompanyNameWithAbbr**: 根据公司简称获取公司名称

#### 公司层级关系
- **HasSubCompany**: 判断公司是否有子公司
  - 通过 parent_company 字段关联查询
  - 返回布尔值表示是否存在子公司

- **GetAllSubCompanies**: 递归查询指定公司下的所有子公司
  - 通过 parent_company 字段递归查询
  - 返回所有层级的子公司列表（包括子公司的子公司）

## 业务场景

### 公司信息查询
- 根据公司名称或简称快速查询公司信息
- 支持批量查询公司列表

### 公司层级管理
- 查询公司的组织架构关系
- 获取完整的子公司树形结构
- 用于权限控制和数据隔离

## 使用说明

### 服务注册
```go
service.RegisterCompany(companyImpl)
```

### 服务调用
```go
// 获取公司服务实例
company := service.Company()

// 查询公司信息
companyInfo, err := company.GetCompany(ctx, "公司名称")

// 查询所有子公司
subCompanies, err := company.GetAllSubCompanies(ctx, "父公司名称")
```

## 数据结构

### 公司关系字段
- **parent_company**: 父公司字段，用于建立公司层级关系
- 支持多层级嵌套结构

## 注意事项
1. 公司简称（abbr）需要保证唯一性
2. 递归查询子公司时注意性能问题，避免层级过深
3. 公司层级关系变更需要考虑对现有业务的影响
