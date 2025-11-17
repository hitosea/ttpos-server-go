# Check Name Service 名称检查服务说明文档

## 📋 概述

`service/check_name.go` 是 TTPOS 系统的名称检查服务，负责验证多语言名称的唯一性。该服务支持对商品、分类、原料、仓库等各类业务实体进行多语言名称重复性检查，确保在同一类型下不存在重复的名称。

**文件路径**: `/home/coder/workspaces/ttpos-server-go/main/app/service/check_name.go`  
**代码行数**: 385 行  
**接口定义**: `ICheckNameSrv`  
**实现结构**: `checkNameSrv`

---

## 🏗️ 架构设计

### 接口定义 (ICheckNameSrv)

```go
type ICheckNameSrv interface {
    CheckNameExists(ctx context.Context, param req.CheckNameRequest) (resp.CheckNameResp, error)
    MakeCheckNameList(ctx context.Context, param dto.LocaleResponse) []req.CheckingName
    InnerCheckNameExists(ctx context.Context, param req.CheckNameRequest) bool
    CheckNameLength(ctx context.Context, name string, maxLength int) bool
}
```

### 依赖服务

```go
type checkNameSrv struct {
    dbm *database.DBManager  // 数据库管理器
}
```

**依赖说明**:
- 仅依赖数据库管理器
- 通过多语言名称表 (`ttpos_multi_language_name`) 关联查询

---

## 🌐 多语言支持

### 支持的语言

| 语言代码 | 语言名称 | 数据库字段 |
|---------|---------|-----------|
| `zh` | 中文简体 | `zh_name` |
| `zhtw` | 中文繁体 | `zh_tw_name` |
| `en` | 英文 | `en_name` |
| `ja` | 日语 | `ja_name` |
| `ko` | 韩语 | `ko_name` |
| `my` | 缅甸语 | `my_name` |
| `sv` | 瑞典语 | `sv_name` |
| `th` | 泰语 | `th_name` |
| `tr` | 土耳其语 | `tr_name` |

### 语言映射表

```go
keyMap := map[string]string{
    "zh":   "zh_name",
    "zhtw": "zh_tw_name",
    "en":   "en_name",
    "ja":   "ja_name",
    "ko":   "ko_name",
    "my":   "my_name",
    "sv":   "sv_name",
    "th":   "th_name",
    "tr":   "tr_name",
}
```

---

## 🎯 核心功能

### 1. 检查名称是否存在 (CheckNameExists)

**功能描述**: 检查指定来源和语言的名称是否已存在于系统中，支持编辑时排除当前记录。

#### 应用场景
- 创建商品时检查名称是否重复
- 编辑分类时验证新名称是否可用
- 添加原料时确保名称唯一性

#### 请求参数

```go
type CheckNameRequest struct {
    Source uint8           // 数据源类型
    Uuid   uint64          // 当前记录UUID（编辑时传入，创建时为0）
    Names  []CheckingName  // 需要检查的名称列表
}

type CheckingName struct {
    Lang string  // 语言代码
    Text string  // 名称文本
}
```

#### 响应数据

```go
type CheckNameResp struct {
    List []CheckNameItem  // 检查结果列表
}

type CheckNameItem struct {
    Lang      string  // 语言代码
    TextExist bool    // 名称是否已存在
}
```

#### 支持的数据源

| 常量 | 值 | 说明 | 关联表 |
|-----|---|------|-------|
| `CheckNameSourceUnit` | 1 | 商品单位 | `ttpos_product_unit` |
| `CheckNameSourceProductPackageGroup` | 2 | 商品套餐组 | 暂未实现 |
| `CheckNameSourceProduct` | 3 | 商品 | `ttpos_product_package` |
| `CheckNameSourceCategory` | 4 | 商品分类 | `ttpos_product_category` |
| `CheckNameSourceSauce` | 5 | 加料 | `ttpos_product_sauce` |
| `CheckNameSourceAttribute` | 6 | 商品属性 | `ttpos_product_attribute` |
| `CheckNameSourceAttributeGroup` | 7 | 属性组 | `ttpos_product_attribute_group` |
| `CheckNameSourceFlavor` | 8 | 规格 | `ttpos_product_flavor` |
| `CheckNameSourceMaterial` | 9 | 原料 | `ttpos_material` |
| `CheckNameSourceMaterialCategory` | 10 | 原料分类 | `ttpos_material_category` |
| `CheckNameSourceMaterialUnit` | 11 | 原料单位 | `ttpos_material_unit` |
| `CheckNameSourceWarehouse` | 12 | 仓库 | `ttpos_warehouse` |
| `CheckNameSourceBatchTag` | 13 | 批次标签 | `ttpos_batch_tag` |

#### 查询逻辑

```go
// 1. 关联多语言名称表
query := db.Model(&model.ProductUnit{}).
    Joins("JOIN ttpos_multi_language_name ON ttpos_product_unit.multi_language_name_uuid = ttpos_multi_language_name.uuid").
    
// 2. 查询指定语言的名称
    Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text).
    
// 3. 排除已删除的记录
    Where("ttpos_multi_language_name.delete_time = 0")

// 4. 编辑时排除当前记录
if checkNameReq.Uuid != 0 {
    query = query.Where("ttpos_product_unit.uuid != ?", checkNameReq.Uuid)
}

// 5. 统计数量
query.Count(&count)

// 6. 返回结果
result = append(result, resp.CheckNameItem{
    Lang:      name.Lang,
    TextExist: count > 0,
})
```

#### 使用示例

##### 示例1: 创建商品时检查名称

```go
// 请求
checkReq := req.CheckNameRequest{
    Source: constant.CheckNameSourceProduct,
    Uuid:   0,  // 创建时为0
    Names: []req.CheckingName{
        {Lang: "zh", Text: "宫保鸡丁"},
        {Lang: "en", Text: "Kung Pao Chicken"},
        {Lang: "th", Text: "ไก่กังเปา"},
    },
}

// 调用
resp, err := checkNameSrv.CheckNameExists(ctx, checkReq)

// 响应
{
    "List": [
        {"Lang": "zh", "TextExist": false},   // 中文名称不存在，可以使用
        {"Lang": "en", "TextExist": true},    // 英文名称已存在，不可使用
        {"Lang": "th", "TextExist": false}    // 泰文名称不存在，可以使用
    ]
}
```

##### 示例2: 编辑分类时检查名称

```go
// 请求
checkReq := req.CheckNameRequest{
    Source: constant.CheckNameSourceCategory,
    Uuid:   12345,  // 当前编辑的分类UUID
    Names: []req.CheckingName{
        {Lang: "zh", Text: "热菜"},
        {Lang: "en", Text: "Hot Dishes"},
    },
}

// 调用
resp, err := checkNameSrv.CheckNameExists(ctx, checkReq)

// 响应会排除UUID为12345的记录
{
    "List": [
        {"Lang": "zh", "TextExist": false},
        {"Lang": "en", "TextExist": false}
    ]
}
```

#### 数据源查询详解

##### 1. 商品单位 (Unit)

```go
case constant.CheckNameSourceUnit:
    var count int64
    query := db.Model(&model.ProductUnit{}).
        Joins("JOIN ttpos_multi_language_name ON ttpos_product_unit.multi_language_name_uuid = ttpos_multi_language_name.uuid").
        Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text).
        Where("ttpos_multi_language_name.delete_time = 0")
    if checkNameReq.Uuid != 0 {
        query = query.Where("ttpos_product_unit.uuid != ?", checkNameReq.Uuid)
    }
    query.Count(&count)
```

##### 2. 商品套餐组 (ProductPackageGroup)

```go
case constant.CheckNameSourceProductPackageGroup:
    // 暂未实现，直接返回false
    result = append(result, resp.CheckNameItem{
        Lang:      name.Lang,
        TextExist: false,
    })
```

##### 3. 商品 (Product)

```go
case constant.CheckNameSourceProduct:
    var count int64
    query := db.Model(&model.ProductPackage{}).
        Joins("JOIN ttpos_multi_language_name ON ttpos_product_package.multi_language_name_uuid = ttpos_multi_language_name.uuid").
        Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text).
        Where("ttpos_multi_language_name.delete_time = 0")
    if checkNameReq.Uuid != 0 {
        query = query.Where("ttpos_product_package.uuid != ?", checkNameReq.Uuid)
    }
    query.Count(&count)
```

##### 4. 原料 (Material)

```go
case constant.CheckNameSourceMaterial:
    var count int64
    query := db.Model(&model.Material{}).
        Joins("JOIN ttpos_multi_language_name ON ttpos_material.multi_language_name_uuid = ttpos_multi_language_name.uuid").
        Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text).
        Where("ttpos_multi_language_name.delete_time = 0")
    if checkNameReq.Uuid != 0 {
        query = query.Where("ttpos_material.uuid != ?", checkNameReq.Uuid)
    }
    query.Count(&count)
```

##### 5. 仓库 (Warehouse)

```go
case constant.CheckNameSourceWarehouse:
    var count int64
    query := db.Model(&model.Warehouse{}).
        Joins("JOIN ttpos_multi_language_name ON ttpos_warehouse.multi_language_name_uuid = ttpos_multi_language_name.uuid").
        Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text).
        Where("ttpos_multi_language_name.delete_time = 0")
    if checkNameReq.Uuid != 0 {
        query = query.Where("ttpos_warehouse.uuid != ?", checkNameReq.Uuid)
    }
    query.Count(&count)
```

---

### 2. 生成检查名称列表 (MakeCheckNameList)

**功能描述**: 将多语言对象转换为名称检查列表格式，过滤空值。

#### 应用场景
- 创建/编辑商品前准备检查数据
- 批量检查多语言名称

#### 输入参数

```go
type LocaleResponse struct {
    ZH   string  // 中文简体
    ZHTW string  // 中文繁体
    EN   string  // 英文
    JA   string  // 日语
    KO   string  // 韩语
    MY   string  // 缅甸语
    SV   string  // 瑞典语
    TH   string  // 泰语
    TR   string  // 土耳其语
}
```

#### 返回数据

```go
[]CheckingName{
    {Lang: "zh", Text: "宫保鸡丁"},
    {Lang: "en", Text: "Kung Pao Chicken"},
    {Lang: "th", Text: "ไก่กังเปา"},
}
```

#### 实现逻辑

```go
func (s *checkNameSrv) MakeCheckNameList(ctx context.Context, param dto.LocaleResponse) []req.CheckingName {
    var names []req.CheckingName
    
    // 仅添加非空的语言名称
    if param.ZH != "" {
        names = append(names, req.CheckingName{Lang: "zh", Text: param.ZH})
    }
    
    if param.ZHTW != "" {
        names = append(names, req.CheckingName{Lang: "zh-TW", Text: param.ZHTW})
    }
    
    if param.EN != "" {
        names = append(names, req.CheckingName{Lang: "en", Text: param.EN})
    }
    
    // ... 其他语言同理
    
    return names
}
```

#### 使用示例

```go
// 输入
locale := dto.LocaleResponse{
    ZH:   "宫保鸡丁",
    EN:   "Kung Pao Chicken",
    TH:   "ไก่กังเปา",
    JA:   "",  // 空值会被过滤
    KO:   "",  // 空值会被过滤
}

// 转换
names := checkNameSrv.MakeCheckNameList(ctx, locale)

// 输出
[]req.CheckingName{
    {Lang: "zh", Text: "宫保鸡丁"},
    {Lang: "en", Text: "Kung Pao Chicken"},
    {Lang: "th", Text: "ไก่กังเปา"},
}

// 使用检查
checkReq := req.CheckNameRequest{
    Source: constant.CheckNameSourceProduct,
    Uuid:   0,
    Names:  names,
}
resp, err := checkNameSrv.CheckNameExists(ctx, checkReq)
```

---

### 3. 内部检查名称是否存在 (InnerCheckNameExists)

**功能描述**: 简化版的名称检查，直接返回布尔值，不返回详细信息和错误。

#### 应用场景
- 内部业务逻辑判断
- 不需要返回详细错误信息的场景
- 快速验证名称可用性

#### 实现逻辑

```go
func (s *checkNameSrv) InnerCheckNameExists(ctx context.Context, param req.CheckNameRequest) bool {
    // 调用标准检查方法
    checkNameResp, err := s.CheckNameExists(ctx, param)
    
    // 出错返回false（不存在）
    if err != nil {
        return false
    }
    
    // 遍历结果，只要有一个语言的名称已存在，就返回true
    for _, item := range checkNameResp.List {
        if item.TextExist {
            return true
        }
    }
    
    return false
}
```

#### 使用示例

```go
// 场景：创建商品前快速检查
checkReq := req.CheckNameRequest{
    Source: constant.CheckNameSourceProduct,
    Uuid:   0,
    Names: []req.CheckingName{
        {Lang: "zh", Text: "宫保鸡丁"},
        {Lang: "en", Text: "Kung Pao Chicken"},
    },
}

// 快速检查
exists := checkNameSrv.InnerCheckNameExists(ctx, checkReq)

if exists {
    // 名称已存在，不能创建
    return errors.New("商品名称已存在")
} else {
    // 名称可用，继续创建流程
    createProduct(...)
}
```

#### 与 CheckNameExists 的区别

| 特性 | CheckNameExists | InnerCheckNameExists |
|-----|----------------|---------------------|
| 返回类型 | `(resp.CheckNameResp, error)` | `bool` |
| 返回详情 | 每个语言的检查结果 | 仅返回是否存在 |
| 错误处理 | 返回错误信息 | 忽略错误，返回false |
| 使用场景 | API接口、需要详细信息 | 内部逻辑、快速判断 |

---

### 4. 检查名称长度 (CheckNameLength)

**功能描述**: 检查名称长度是否超过限制，使用 UTF-8 字符数计算（而非字节数）。

#### 应用场景
- 创建/编辑时验证名称长度
- 表单验证
- 数据导入验证

#### 参数说明

```go
func CheckNameLength(ctx context.Context, name string, maxLength int) bool
```

| 参数 | 类型 | 说明 |
|-----|------|-----|
| `ctx` | `context.Context` | 上下文 |
| `name` | `string` | 待检查的名称 |
| `maxLength` | `int` | 最大字符数限制 |

#### 实现逻辑

```go
func (s *checkNameSrv) CheckNameLength(ctx context.Context, name string, maxLength int) bool {
    // 使用 utf8.RuneCountInString 计算字符数
    if utf8.RuneCountInString(name) > maxLength {
        return false  // 超过长度限制
    }
    return true  // 长度符合要求
}
```

#### 字符数计算说明

**为什么使用 `utf8.RuneCountInString`？**

```go
// ❌ 错误：使用 len() 计算字节数
len("宫保鸡丁")  // 返回 12（每个中文字符3字节）

// ✅ 正确：使用 utf8.RuneCountInString 计算字符数
utf8.RuneCountInString("宫保鸡丁")  // 返回 4（4个字符）
```

#### 使用示例

##### 示例1: 验证商品名称长度

```go
productName := "泰式咖喱炒蟹"
maxLength := 50

if !checkNameSrv.CheckNameLength(ctx, productName, maxLength) {
    return errors.New("商品名称不能超过50个字符")
}
```

##### 示例2: 验证不同语言的名称长度

```go
locale := dto.LocaleResponse{
    ZH: "泰式咖喱炒蟹配香米饭和泰式酸辣汤套餐",  // 18个字符
    EN: "Thai Curry Crab with Jasmine Rice and Tom Yum Soup Set",  // 55个字符
    TH: "ปูผัดผงกะหรี่พร้อมข้าวหอมมะลิและต้มยำกุ้งชุด",  // 42个字符
}

// 验证中文名称（限制50字符）
if !checkNameSrv.CheckNameLength(ctx, locale.ZH, 50) {
    return errors.New("中文名称不能超过50个字符")
}

// 验证英文名称（限制100字符）
if !checkNameSrv.CheckNameLength(ctx, locale.EN, 100) {
    return errors.New("英文名称不能超过100个字符")
}

// 验证泰文名称（限制50字符）
if !checkNameSrv.CheckNameLength(ctx, locale.TH, 50) {
    return errors.New("泰文名称不能超过50个字符")
}
```

##### 示例3: 表单验证

```go
// 验证所有语言的名称长度
func validateProductNameLength(locale dto.LocaleResponse) error {
    languages := map[string]struct{
        text      string
        maxLength int
        label     string
    }{
        "zh":   {locale.ZH, 50, "中文名称"},
        "zhtw": {locale.ZHTW, 50, "繁体中文名称"},
        "en":   {locale.EN, 100, "英文名称"},
        "th":   {locale.TH, 50, "泰文名称"},
    }
    
    for _, lang := range languages {
        if lang.text != "" && !checkNameSrv.CheckNameLength(ctx, lang.text, lang.maxLength) {
            return errors.Newf("%s不能超过%d个字符", lang.label, lang.maxLength)
        }
    }
    
    return nil
}
```

---

## 🔄 业务流程

### 创建商品流程

```
1. 前端提交商品信息（包含多语言名称）
   ↓
2. MakeCheckNameList() - 转换为检查列表
   ↓
3. CheckNameExists() - 检查名称是否重复
   ↓
4. 判断检查结果
   ├── 有重复 → 返回错误提示
   └── 无重复 → 继续
   ↓
5. CheckNameLength() - 检查名称长度
   ↓
6. 判断长度是否符合要求
   ├── 超长 → 返回错误提示
   └── 符合 → 继续
   ↓
7. 创建多语言名称记录
   ↓
8. 创建商品记录
   ↓
9. 返回成功
```

### 编辑商品流程

```
1. 前端提交修改后的商品信息
   ↓
2. MakeCheckNameList() - 转换为检查列表
   ↓
3. CheckNameExists() - 检查名称是否重复
   - 传入当前商品UUID（排除自己）
   ↓
4. 判断检查结果
   ├── 有重复 → 返回错误提示
   └── 无重复 → 继续
   ↓
5. CheckNameLength() - 检查名称长度
   ↓
6. 更新多语言名称记录
   ↓
7. 更新商品记录
   ↓
8. 返回成功
```

### 快速验证流程

```
1. 业务逻辑需要快速判断名称是否可用
   ↓
2. 组装 CheckNameRequest
   ↓
3. InnerCheckNameExists() - 快速检查
   ↓
4. 返回布尔值
   ├── true → 名称已存在，不可用
   └── false → 名称可用，继续业务逻辑
```

---

## 📊 数据库设计

### 多语言名称表 (ttpos_multi_language_name)

```sql
CREATE TABLE `ttpos_multi_language_name` (
  `uuid` bigint(20) NOT NULL,
  `zh_name` varchar(255) DEFAULT '' COMMENT '中文简体名称',
  `zh_tw_name` varchar(255) DEFAULT '' COMMENT '中文繁体名称',
  `en_name` varchar(255) DEFAULT '' COMMENT '英文名称',
  `ja_name` varchar(255) DEFAULT '' COMMENT '日语名称',
  `ko_name` varchar(255) DEFAULT '' COMMENT '韩语名称',
  `my_name` varchar(255) DEFAULT '' COMMENT '缅甸语名称',
  `sv_name` varchar(255) DEFAULT '' COMMENT '瑞典语名称',
  `th_name` varchar(255) DEFAULT '' COMMENT '泰语名称',
  `tr_name` varchar(255) DEFAULT '' COMMENT '土耳其语名称',
  `create_time` int(11) DEFAULT 0,
  `update_time` int(11) DEFAULT 0,
  `delete_time` int(11) DEFAULT 0,
  PRIMARY KEY (`uuid`),
  KEY `idx_delete_time` (`delete_time`)
) COMMENT='多语言名称表';
```

### 关联表示例

#### 商品表 (ttpos_product_package)

```sql
CREATE TABLE `ttpos_product_package` (
  `uuid` bigint(20) NOT NULL,
  `multi_language_name_uuid` bigint(20) DEFAULT 0 COMMENT '多语言名称UUID',
  -- 其他字段
  PRIMARY KEY (`uuid`),
  KEY `idx_multi_language_name_uuid` (`multi_language_name_uuid`)
);
```

#### 分类表 (ttpos_product_category)

```sql
CREATE TABLE `ttpos_product_category` (
  `uuid` bigint(20) NOT NULL,
  `multi_language_name_uuid` bigint(20) DEFAULT 0 COMMENT '多语言名称UUID',
  -- 其他字段
  PRIMARY KEY (`uuid`),
  KEY `idx_multi_language_name_uuid` (`multi_language_name_uuid`)
);
```

#### 原料表 (ttpos_material)

```sql
CREATE TABLE `ttpos_material` (
  `uuid` bigint(20) NOT NULL,
  `multi_language_name_uuid` bigint(20) DEFAULT 0 COMMENT '多语言名称UUID',
  -- 其他字段
  PRIMARY KEY (`uuid`),
  KEY `idx_multi_language_name_uuid` (`multi_language_name_uuid`)
);
```

---

## 🚨 错误处理

### 错误类型

| 错误 | 说明 | 处理方式 |
|-----|------|---------|
| 类型不支持 | 传入了未实现的数据源类型 | 返回 "类型不支持" |
| 数据库查询失败 | 查询时发生数据库错误 | 记录日志，返回错误 |
| 语言不支持 | 传入了不支持的语言代码 | 跳过该语言，继续处理其他语言 |

### 错误处理示例

```go
// ✅ 不支持的数据源
default:
    return resp.CheckNameResp{}, errors.New("类型不支持")

// ✅ 不支持的语言（跳过）
if _, ok := keyMap[name.Lang]; !ok {
    continue
}

// ✅ 内部检查忽略错误
checkNameResp, err := s.CheckNameExists(ctx, param)
if err != nil {
    return false  // 出错视为不存在
}
```

---

## 🎯 最佳实践

### 1. 创建/编辑时的完整验证

```go
// ✅ 完整的名称验证流程
func (s *productSrv) CreateProduct(ctx context.Context, req req.CreateProductReq) error {
    // 1. 转换为检查列表
    names := s.checkNameSrv.MakeCheckNameList(ctx, req.Name)
    
    // 2. 检查名称重复
    checkReq := req.CheckNameRequest{
        Source: constant.CheckNameSourceProduct,
        Uuid:   0,  // 创建时为0
        Names:  names,
    }
    checkResp, err := s.checkNameSrv.CheckNameExists(ctx, checkReq)
    if err != nil {
        return errors.WithMessage(err, "检查名称失败")
    }
    
    // 3. 判断是否有重复
    for _, item := range checkResp.List {
        if item.TextExist {
            return errors.Newf("%s名称已存在", item.Lang)
        }
    }
    
    // 4. 检查名称长度
    if req.Name.ZH != "" && !s.checkNameSrv.CheckNameLength(ctx, req.Name.ZH, 50) {
        return errors.New("中文名称不能超过50个字符")
    }
    if req.Name.EN != "" && !s.checkNameSrv.CheckNameLength(ctx, req.Name.EN, 100) {
        return errors.New("英文名称不能超过100个字符")
    }
    
    // 5. 创建商品
    return s.createProductWithName(ctx, req)
}
```

### 2. 编辑时排除当前记录

```go
// ✅ 编辑时传入UUID
func (s *productSrv) UpdateProduct(ctx context.Context, uuid uint64, req req.UpdateProductReq) error {
    names := s.checkNameSrv.MakeCheckNameList(ctx, req.Name)
    
    checkReq := req.CheckNameRequest{
        Source: constant.CheckNameSourceProduct,
        Uuid:   uuid,  // 传入当前商品UUID
        Names:  names,
    }
    
    checkResp, err := s.checkNameSrv.CheckNameExists(ctx, checkReq)
    if err != nil {
        return errors.WithMessage(err)
    }
    
    // 检查重复...
}
```

### 3. 内部逻辑快速判断

```go
// ✅ 使用 InnerCheckNameExists 简化代码
func (s *productSrv) QuickValidate(ctx context.Context, name string) bool {
    checkReq := req.CheckNameRequest{
        Source: constant.CheckNameSourceProduct,
        Uuid:   0,
        Names:  []req.CheckingName{{Lang: "zh", Text: name}},
    }
    
    // 直接返回布尔值
    return !s.checkNameSrv.InnerCheckNameExists(ctx, checkReq)
}
```

### 4. 批量检查多种语言

```go
// ✅ 一次检查所有语言
func (s *productSrv) ValidateAllLanguages(ctx context.Context, locale dto.LocaleResponse) error {
    // 转换为检查列表（自动过滤空值）
    names := s.checkNameSrv.MakeCheckNameList(ctx, locale)
    
    if len(names) == 0 {
        return errors.New("至少需要填写一个语言的名称")
    }
    
    checkReq := req.CheckNameRequest{
        Source: constant.CheckNameSourceProduct,
        Uuid:   0,
        Names:  names,
    }
    
    checkResp, err := s.checkNameSrv.CheckNameExists(ctx, checkReq)
    if err != nil {
        return errors.WithMessage(err)
    }
    
    // 收集所有重复的语言
    var duplicates []string
    for _, item := range checkResp.List {
        if item.TextExist {
            duplicates = append(duplicates, item.Lang)
        }
    }
    
    if len(duplicates) > 0 {
        return errors.Newf("以下语言的名称已存在: %v", duplicates)
    }
    
    return nil
}
```

---

## 🔧 性能优化

### 1. 单个数据库连接

```go
// ✅ 每次查询获取新的数据库连接实例
for _, name := range checkNameReq.Names {
    db := s.dbm.GetDB(ctx.GetCompanyUuid())
    // 执行查询...
}
```

**原因**: 避免连接复用导致的查询条件污染。

### 2. 仅统计数量

```go
// ✅ 使用 Count 而不是查询所有数据
query.Count(&count)

// ❌ 不要这样做
var items []model.Product
query.Find(&items)
count = len(items)  // 性能差
```

### 3. 索引优化

确保以下字段有索引：
- `multi_language_name_uuid` - 关联查询
- `delete_time` - 过滤已删除记录
- 各语言名称字段 - 名称查询

---

## 📈 扩展建议

### 1. 添加新语言支持

**步骤**:

1. 在 `keyMap` 中添加新语言映射
2. 在 `MakeCheckNameList` 中添加新语言处理
3. 数据库添加新的语言字段

```go
// 示例：添加法语支持
keyMap := map[string]string{
    // ... 现有语言
    "fr": "fr_name",  // 添加法语
}

// MakeCheckNameList 中添加
if param.FR != "" {
    names = append(names, req.CheckingName{
        Lang: "fr",
        Text: param.FR,
    })
}
```

### 2. 添加新数据源

**步骤**:

1. 在 `constant` 包中定义新的数据源常量
2. 在 `CheckNameExists` 的 `switch` 中添加新的 `case`
3. 编写对应的查询逻辑

```go
// 示例：添加菜单数据源
case constant.CheckNameSourceMenu:
    var count int64
    query := db.Model(&model.Menu{}).
        Joins("JOIN ttpos_multi_language_name ON ttpos_menu.multi_language_name_uuid = ttpos_multi_language_name.uuid").
        Where(fmt.Sprintf("ttpos_multi_language_name.%s = ?", keyMap[name.Lang]), name.Text).
        Where("ttpos_multi_language_name.delete_time = 0")
    if checkNameReq.Uuid != 0 {
        query = query.Where("ttpos_menu.uuid != ?", checkNameReq.Uuid)
    }
    query.Count(&count)
    
    result = append(result, resp.CheckNameItem{
        Lang:      name.Lang,
        TextExist: count > 0,
    })
```

### 3. 优化查询性能

**建议**:
- 添加缓存机制（常用名称查询）
- 异步批量检查
- 数据库读写分离

---

## 🧪 测试建议

### 单元测试覆盖

1. **CheckNameExists 测试**
   - 名称不存在的情况
   - 名称已存在的情况
   - 编辑时排除当前记录
   - 不支持的语言
   - 不支持的数据源
   - 多语言混合检查

2. **MakeCheckNameList 测试**
   - 所有语言都填写
   - 部分语言填写
   - 所有语言都为空
   - 特殊字符处理

3. **InnerCheckNameExists 测试**
   - 名称存在返回true
   - 名称不存在返回false
   - 查询出错返回false

4. **CheckNameLength 测试**
   - 中文字符计数
   - 英文字符计数
   - 混合字符计数
   - emoji字符计数
   - 边界值测试

### 测试示例

```go
func TestCheckNameExists(t *testing.T) {
    // 测试商品名称不存在
    t.Run("Product name not exists", func(t *testing.T) {
        req := req.CheckNameRequest{
            Source: constant.CheckNameSourceProduct,
            Uuid:   0,
            Names: []req.CheckingName{
                {Lang: "zh", Text: "测试商品"},
            },
        }
        
        resp, err := checkNameSrv.CheckNameExists(ctx, req)
        assert.NoError(t, err)
        assert.False(t, resp.List[0].TextExist)
    })
    
    // 测试商品名称已存在
    t.Run("Product name exists", func(t *testing.T) {
        // 先创建一个商品
        createProduct(ctx, "测试商品")
        
        req := req.CheckNameRequest{
            Source: constant.CheckNameSourceProduct,
            Uuid:   0,
            Names: []req.CheckingName{
                {Lang: "zh", Text: "测试商品"},
            },
        }
        
        resp, err := checkNameSrv.CheckNameExists(ctx, req)
        assert.NoError(t, err)
        assert.True(t, resp.List[0].TextExist)
    })
}

func TestCheckNameLength(t *testing.T) {
    tests := []struct{
        name      string
        input     string
        maxLength int
        expected  bool
    }{
        {"中文字符", "宫保鸡丁", 50, true},
        {"中文超长", "这是一个非常非常非常非常非常非常非常非常非常非常非常非常长的名称", 20, false},
        {"英文字符", "Kung Pao Chicken", 100, true},
        {"emoji字符", "🍗🍗🍗", 5, true},
        {"混合字符", "宫保鸡丁Kung Pao🍗", 50, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := checkNameSrv.CheckNameLength(ctx, tt.input, tt.maxLength)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

---

## 📚 相关文档

- [多语言管理](./multi_language.md)
- [商品管理](./product_service.md)
- [分类管理](./category_service.md)
- [原料管理](./material_service.md)

---

## 📄 更新日志

| 日期 | 版本 | 说明 |
|-----|------|-----|
| 2025-11-12 | 1.0 | 初始文档创建 |

---

## 👥 维护者

- 开发团队：Backend Team
- 文档维护：AI Assistant

---

## ⚠️ 注意事项

1. **语言代码一致性**: 确保前端和后端使用相同的语言代码
2. **商品套餐组**: `CheckNameSourceProductPackageGroup` 暂未实现，直接返回不存在
3. **字符数计算**: 使用 `utf8.RuneCountInString` 而非 `len()`
4. **编辑时必须传UUID**: 编辑操作必须传入当前记录的UUID，否则会误判为重复
5. **软删除**: 查询时必须排除 `delete_time != 0` 的记录
6. **数据库表前缀**: JOIN查询时注意使用完整表名（含 `ttpos_` 前缀）

---

**注意**: 本文档基于代码自动生成，如有代码变更，请及时更新文档。如需添加新语言或新数据源，请参考扩展建议章节。

