# 文件上传服务 (Upload File Service)

## 概述

`upload_file.go` 实现了文件上传管理服务，负责处理餐饮系统中各类文件的上传、存储和管理。该服务支持多种文件类型（图片、视频、文档），提供多种存储引擎（本地存储、Google Cloud Storage），并实现了文件验证、缩略图生成、记录管理等功能，是系统资源管理的核心模块。

**文件路径**: `ttpos-server-go/main/app/service/upload_file.go`

## 核心功能

### 1. 多类型文件上传
- 图片上传（JPG、JPEG、PNG、WEBP）
- 视频上传（AVI、MPEG、MOV、MP4）
- 文档上传（PDF、Word、Excel、图片）

### 2. 多存储引擎支持
- 本地文件系统存储
- Google Cloud Storage
- 存储引擎可配置切换

### 3. 文件处理
- 文件类型验证
- 文件大小限制
- 缩略图自动生成
- 文件记录持久化

### 4. 文件管理
- 文件信息查询
- URL 生成
- 分组管理

## 接口定义

### IUploadFileSrv 接口

```go
type IUploadFileSrv interface {
    UploadImage(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64, source string) (*resp.UploadFileResp, error)
    UploadVideo(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64, maxSize int) (*resp.UploadFileResp, error)
    UploadDocument(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64) (*resp.UploadFileResp, error)
    GetUploadFile(ctx context.Context, uuid uint64) (*resp.UploadFileResp, error)
}
```

### UploadFileSrvImpl 结构体

```go
type UploadFileSrvImpl struct {
    dbm            *database.DBManager  // 数据库管理器
    storageFactory *storage.Factory     // 存储引擎工厂
}
```

## 依赖项

### 内部依赖
- **storage.Factory**: 存储引擎工厂，创建和管理存储引擎
- **storage.Engine**: 存储引擎接口，实现具体的文件存储逻辑

### 外部依赖
- **database.DBManager**: 数据库管理器
- **viper**: 配置管理
- **io.Reader**: 文件读取接口

## 支持的文件类型

### 1. 图片文件

| 格式 | 扩展名 | 最大尺寸 | 缩略图 |
|------|--------|---------|--------|
| JPG | .jpg | 无限制 | 500px / 5000px |
| JPEG | .jpeg | 无限制 | 500px / 5000px |
| PNG | .png | 无限制 | 500px / 5000px |
| WEBP | .webp | 无限制 | 500px / 5000px |

**缩略图尺寸规则**:
- 有 `source` 参数：500px
- 无 `source` 参数：5000px（不缩放或轻微缩放）

### 2. 视频文件

| 格式 | 扩展名 | 默认限制 | 最大限制 |
|------|--------|---------|---------|
| AVI | .avi | 10MB | 30MB |
| MPEG | .mpeg | 10MB | 30MB |
| MOV | .mov | 10MB | 30MB |
| MP4 | .mp4 | 10MB | 30MB |

**大小限制**:
- 默认：10MB
- 可配置最大：30MB

### 3. 文档文件

| 格式 | 扩展名 | 最大尺寸 |
|------|--------|---------|
| PDF | .pdf | 20MB |
| Word | .doc, .docx | 20MB |
| Excel | .xls, .xlsx | 20MB |
| 图片 | .jpg, .jpeg, .png, .gif | 20MB |

**特殊处理**:
- 文档中的图片文件类型标记为 `image`
- 其他文档类型标记为 `document`

## 核心方法详解

### 1. UploadImage - 上传图片

**方法签名**:
```go
func (s *UploadFileSrvImpl) UploadImage(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64, source string) (*resp.UploadFileResp, error)
```

**功能**: 上传图片文件并生成缩略图。

**参数说明**:
- `ctx`: 上下文，包含公司信息等
- `fileReader`: 文件读取器（文件流）
- `fileName`: 原始文件名
- `fileSize`: 文件大小（字节）
- `groupId`: 文件分组ID
- `source`: 来源标识（影响缩略图尺寸）

**返回值**:
```go
type UploadFileResp struct {
    Uuid          uint64 // 文件UUID
    GroupUuid     uint64 // 分组UUID
    Storage       string // 存储引擎
    FileUrl       string // 文件域名
    FileName      string // 生成的文件名
    SaveName      string // 保存路径
    FileSize      int64  // 文件大小
    FileType      string // 文件类型
    Extension     string // 扩展名
    RealName      string // 原始文件名
    IndexFileName string // 索引文件名（不含扩展名）
    UrlParam      string // URL参数
    FilePath      string // 完整访问路径
    CreateTime    int    // 创建时间
}
```

**实现流程**:

```67:87:ttpos-server-go/main/app/service/upload_file.go
func (s *UploadFileSrvImpl) UploadImage(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64, source string) (*resp.UploadFileResp, error) {
	// 验证文件类型
	extension := strings.ToLower(filepath.Ext(fileName))
	if extension != "" && extension[0] == '.' {
		extension = extension[1:]
	}

	allowedExts := []string{"jpg", "jpeg", "png", "webp"}
	if !s.isAllowedExtension(extension, allowedExts) {
		return nil, fmt.Errorf("仅支持JPG、JPEG、PNG、WEBP格式")
	}

	// 确定缩略图尺寸
	thumbSize := 500
	if source == "" {
		thumbSize = 5000
	}

	return s.uploadFile(ctx, fileReader, fileName, fileSize, groupId, "image", thumbSize)
}
```

**处理流程**:
1. 提取文件扩展名并转小写
2. 验证扩展名是否在允许列表中
3. 根据 `source` 参数确定缩略图尺寸
4. 调用通用上传方法

**缩略图策略**:
- `source` 非空：生成 500px 缩略图（移动端、列表展示）
- `source` 为空：生成 5000px 缩略图（高清展示、打印）

**使用场景**:
- 商品图片上传
- 员工头像上传
- 分类图片上传
- 广告图片上传

---

### 2. UploadVideo - 上传视频

**方法签名**:
```go
func (s *UploadFileSrvImpl) UploadVideo(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64, maxSize int) (*resp.UploadFileResp, error)
```

**功能**: 上传视频文件，支持可配置的大小限制。

**参数说明**:
- `maxSize`: 最大文件大小（MB），影响错误提示

**实现流程**:

```89:112:ttpos-server-go/main/app/service/upload_file.go
func (s *UploadFileSrvImpl) UploadVideo(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64, maxSize int) (*resp.UploadFileResp, error) {
	// 验证文件类型
	extension := strings.ToLower(filepath.Ext(fileName))
	if extension != "" && extension[0] == '.' {
		extension = extension[1:]
	}

	allowedExts := []string{"avi", "mpeg", "mov", "mp4"}
	if !s.isAllowedExtension(extension, allowedExts) {
		return nil, fmt.Errorf("仅支持AVI、MPEG、MOV、MP4格式")
	}

	// 检查文件大小
	maxSizeBytes := int64(maxSize * 1024 * 1024)
	if fileSize > maxSizeBytes {
		if maxSize > 30 {
			return nil, fmt.Errorf("文件大小不能超过30MB")
		}
		return nil, fmt.Errorf("文件大小不能超过10MB")
	}

	return s.uploadFile(ctx, fileReader, fileName, fileSize, groupId, "video", 0)
}
```

**大小限制逻辑**:
```go
if fileSize > maxSize * 1024 * 1024 {
    if maxSize > 30 {
        return "文件大小不能超过30MB"
    } else {
        return "文件大小不能超过10MB"
    }
}
```

**错误提示策略**:
- `maxSize > 30`: 提示"不能超过30MB"
- `maxSize <= 30`: 提示"不能超过10MB"

**使用场景**:
- 营销活动视频
- 商品展示视频
- 培训视频上传

---

### 3. UploadDocument - 上传文档

**方法签名**:
```go
func (s *UploadFileSrvImpl) UploadDocument(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64) (*resp.UploadFileResp, error)
```

**功能**: 上传文档文件，支持办公文档和图片。

**实现流程**:

```255:282:ttpos-server-go/main/app/service/upload_file.go
func (s *UploadFileSrvImpl) UploadDocument(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64) (*resp.UploadFileResp, error) {
	// 验证文件类型
	extension := strings.ToLower(filepath.Ext(fileName))
	if extension != "" && extension[0] == '.' {
		extension = extension[1:]
	}

	allowedExts := []string{"pdf", "doc", "docx", "xls", "xlsx", "jpg", "jpeg", "png", "gif"}
	if !s.isAllowedExtension(extension, allowedExts) {
		return nil, fmt.Errorf("仅支持PDF、Word、Excel、JPG、PNG、GIF格式")
	}

	// 检查文件大小（20MB = 20 * 1024 * 1024 bytes）
	maxSizeBytes := int64(20 * 1024 * 1024)
	if fileSize > maxSizeBytes {
		return nil, fmt.Errorf("文件大小不能超过20MB")
	}

	// 确定文件类型
	fileType := "document"
	imageExts := []string{"jpg", "jpeg", "png", "gif"}
	if s.isAllowedExtension(extension, imageExts) {
		fileType = "image"
	}

	return s.uploadFile(ctx, fileReader, fileName, fileSize, groupId, fileType, 0)
}
```

**文件类型判断**:
```go
if extension in ["jpg", "jpeg", "png", "gif"] {
    fileType = "image"
} else {
    fileType = "document"
}
```

**特点**:
1. 支持办公文档和图片混合上传
2. 统一 20MB 大小限制
3. 图片文件特殊标记（便于后续处理）

**使用场景**:
- 采购单附件上传
- 合同文档上传
- 报表附件上传
- 说明文档上传

---

### 4. GetUploadFile - 获取文件信息

**方法签名**:
```go
func (s *UploadFileSrvImpl) GetUploadFile(ctx context.Context, uuid uint64) (*resp.UploadFileResp, error)
```

**功能**: 根据文件UUID获取文件详细信息。

**实现流程**:

```114:141:ttpos-server-go/main/app/service/upload_file.go
func (s *UploadFileSrvImpl) GetUploadFile(ctx context.Context, uuid uint64) (*resp.UploadFileResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	var uploadFile model.File
	err := db.Model(&model.File{}).Where("uuid = ? AND delete_time = 0", uuid).First(&uploadFile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文件不存在")
		}
		return nil, fmt.Errorf("获取文件失败: %v", err)
	}
	return &resp.UploadFileResp{
		Uuid:          uploadFile.Uuid,
		GroupUuid:     uploadFile.GroupUuid,
		Storage:       uploadFile.Storage,
		FileUrl:       uploadFile.FileUrl,
		FileName:      uploadFile.FileName,
		SaveName:      uploadFile.SaveName,
		FileSize:      int64(uploadFile.FileSize),
		FileType:      uploadFile.FileType,
		Extension:     uploadFile.Extension,
		RealName:      uploadFile.RealName,
		IndexFileName: uploadFile.IndexFileName,
		UrlParam:      uploadFile.UrlParam,
		FilePath:      uploadFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request)),
		CreateTime:    int(uploadFile.CreateTime),
	}, nil
}
```

**查询条件**:
```sql
WHERE uuid = ? AND delete_time = 0
```

**关键点**:
1. 只查询未删除的文件（`delete_time = 0`）
2. 使用 `GetUrl` 方法生成完整访问路径
3. 文件不存在返回友好错误

**使用场景**:
- 显示文件详情
- 生成下载链接
- 文件预览

---

### 5. uploadFile - 通用上传方法（私有）

**方法签名**:
```go
func (s *UploadFileSrvImpl) uploadFile(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64, fileType string, thumbSize int) (*resp.UploadFileResp, error)
```

**功能**: 所有上传方法的底层实现，处理文件存储和记录保存。

**实现流程**:

```144:229:ttpos-server-go/main/app/service/upload_file.go
func (s *UploadFileSrvImpl) uploadFile(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64, fileType string, thumbSize int) (*resp.UploadFileResp, error) {
	// 创建存储引擎
	engine, err := s.storageFactory.CreateEngine()
	if err != nil {
		return nil, fmt.Errorf("创建存储引擎失败: %v", err)
	}

	// 设置文件信息
	if err := engine.SetUploadFile(fileReader, fileName, fileType, fileSize); err != nil {
		return nil, fmt.Errorf("设置文件信息失败: %v", err)
	}

	// 设置公司ID
	engine.SetCompanyUuid(ctx.GetCompanyUuid())

	// 上传文件
	saveName, err := engine.Upload(thumbSize)
	if err != nil {
		return nil, fmt.Errorf("文件上传失败: %v", err)
	}

	if saveName == "" {
		errMsg := engine.GetError()
		if errMsg == "" {
			errMsg = "未知错误"
		}
		return nil, fmt.Errorf("上传失败: %s", errMsg)
	}

	// 标准化路径分隔符
	saveName = strings.ReplaceAll(saveName, "\\", "/")

	// 获取文件名和URL参数
	generatedFileName := engine.GetFileName()
	urlParam := engine.GetUrlParam()

	// 获取存储配置
	storageEngine := s.storageFactory.GetDefaultEngine()

	var fileUrl string
	if domain, ok := s.storageFactory.GetDefaultEngineConfig()["domain"]; ok {
		fileUrl = domain.(string)
	}

	// 创建文件记录
	uploadFile := model.File{
		BaseModel: model.BaseModel{
			Uuid: s.generateUuid(),
		},
		GroupUuid:     groupId,
		Storage:       storageEngine,
		FileUrl:       fileUrl,
		FileName:      generatedFileName,
		SaveName:      saveName,
		FileSize:      int(fileSize),
		FileType:      fileType,
		Extension:     s.getFileExtension(fileName),
		RealName:      fileName,
		IndexFileName: s.getFileNameWithoutExt(fileName),
		UrlParam:      urlParam,
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	err = db.Create(&uploadFile).Error
	if err != nil {
		return nil, fmt.Errorf("保存文件记录失败: %v", err)
	}

	// 返回结果
	return &resp.UploadFileResp{
		Uuid:          uploadFile.Uuid,
		GroupUuid:     uploadFile.GroupUuid,
		Storage:       uploadFile.Storage,
		FileUrl:       uploadFile.FileUrl,
		FileName:      uploadFile.FileName,
		SaveName:      uploadFile.SaveName,
		FileSize:      int64(uploadFile.FileSize),
		FileType:      uploadFile.FileType,
		Extension:     uploadFile.Extension,
		RealName:      uploadFile.RealName,
		IndexFileName: uploadFile.IndexFileName,
		UrlParam:      uploadFile.UrlParam,
		FilePath:      uploadFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request)),
		CreateTime:    int(uploadFile.CreateTime),
	}, nil
}
```

**处理步骤**:

#### 1. 创建存储引擎
```go
engine, err := s.storageFactory.CreateEngine()
```
- 根据配置创建对应的存储引擎
- 支持本地存储、Google Cloud Storage

#### 2. 设置文件信息
```go
engine.SetUploadFile(fileReader, fileName, fileType, fileSize)
engine.SetCompanyUuid(ctx.GetCompanyUuid())
```
- 传递文件流和元信息
- 设置公司UUID（用于文件隔离）

#### 3. 执行上传
```go
saveName, err := engine.Upload(thumbSize)
```
- 调用引擎的上传方法
- 传递缩略图尺寸参数
- 返回保存的文件路径

#### 4. 标准化路径
```go
saveName = strings.ReplaceAll(saveName, "\\", "/")
```
- 统一使用 `/` 作为路径分隔符
- 避免 Windows 和 Linux 路径差异

#### 5. 保存文件记录
```go
uploadFile := model.File{
    Uuid:      generateUuid(),
    Storage:   storageEngine,
    FileName:  generatedFileName,
    SaveName:  saveName,
    // ...
}
db.Create(&uploadFile)
```

#### 6. 返回结果
- 包含完整的文件信息
- 生成可访问的 URL

---

### 6. 辅助方法

#### isAllowedExtension - 检查扩展名
```go
func (s *UploadFileSrvImpl) isAllowedExtension(ext string, allowedExts []string) bool
```

**功能**: 验证文件扩展名是否在允许列表中

#### getFileExtension - 获取扩展名
```go
func (s *UploadFileSrvImpl) getFileExtension(fileName string) string
```

**功能**: 提取文件扩展名（不含点）

#### getFileNameWithoutExt - 获取文件名（不含扩展名）
```go
func (s *UploadFileSrvImpl) getFileNameWithoutExt(fileName string) string
```

**功能**: 提取文件名，去除扩展名部分

#### generateUuid - 生成UUID
```go
func (s *UploadFileSrvImpl) generateUuid() uint64
```

**功能**: 生成唯一的文件UUID

**实现**:
```285:291:ttpos-server-go/main/app/service/upload_file.go
func (s *UploadFileSrvImpl) generateUuid() uint64 {
	if uuid, err := utils.GetID(); err == nil {
		return uuid
	}
	// 如果生成失败，使用时间戳作为后备方案
	return uint64(time.Now().UnixNano())
}
```

**生成策略**:
1. 优先使用 `utils.GetID()`（雪花算法或其他分布式ID）
2. 失败时使用纳秒时间戳作为备用

---

## 存储引擎

### 1. 存储引擎配置

```go
storageFactory := storage.NewFactory(&storage.StorageConfig{
    Default: viper.GetString("STORAGE_DRIVER"),
    Engine: map[string]map[string]any{
        "local": {
            "upload_path": "public/uploads",
        },
        "google": {
            "credentials_file":  config.GoogleBucket.GoogleApplicationCredentialsFileName,
            "bucket":            config.GoogleBucket.GoogleApplicationUploadsBucketName,
            "uploads_catalogue": config.GoogleBucket.GoogleApplicationUploadsCatalogueName,
            "domain":            fmt.Sprintf("https://storage.googleapis.com/%s/%s", ...),
        },
    },
})
```

### 2. 本地存储引擎

**配置**:
```go
"local": {
    "upload_path": "public/uploads"
}
```

**特点**:
- 文件保存在服务器本地
- 路径: `public/uploads/`
- 适合单机部署或开发环境

**目录结构**:
```
public/
└── uploads/
    ├── 20231216/           # 按日期分目录
    │   ├── image/          # 按类型分类
    │   │   ├── original/   # 原图
    │   │   └── thumb/      # 缩略图
    │   └── document/
    └── 20231217/
```

### 3. Google Cloud Storage 引擎

**配置**:
```go
"google": {
    "credentials_file":  "credentials.json",
    "bucket":            "my-bucket",
    "uploads_catalogue": "uploads",
    "domain":            "https://storage.googleapis.com/my-bucket/uploads"
}
```

**特点**:
- 文件保存在 Google Cloud
- 分布式高可用
- 适合生产环境和大规模部署
- 支持 CDN 加速

**访问URL格式**:
```
https://storage.googleapis.com/{bucket}/{catalogue}/{path}
```

### 4. 存储引擎接口

```go
type Engine interface {
    SetUploadFile(reader io.Reader, fileName string, fileType string, fileSize int64) error
    SetCompanyUuid(companyUuid uint64)
    Upload(thumbSize int) (string, error)
    GetFileName() string
    GetUrlParam() string
    GetError() string
}
```

---

## 数据模型

### File - 文件表

```go
type File struct {
    BaseModel
    Uuid          uint64 `gorm:"primary_key"` // 文件UUID
    GroupUuid     uint64                      // 分组UUID
    Storage       string                      // 存储引擎（local/google）
    FileUrl       string                      // 文件域名
    FileName      string                      // 生成的文件名
    SaveName      string                      // 保存路径
    FileSize      int                         // 文件大小（字节）
    FileType      string                      // 文件类型（image/video/document）
    Extension     string                      // 扩展名
    RealName      string                      // 原始文件名
    IndexFileName string                      // 索引文件名（不含扩展名）
    UrlParam      string                      // URL参数
    CreateTime    int64                       // 创建时间
    UpdateTime    int64                       // 更新时间
    DeleteTime    int64                       // 删除时间（软删除）
}
```

### 字段说明

#### Storage - 存储引擎
- `local`: 本地文件系统
- `google`: Google Cloud Storage

#### FileType - 文件类型
- `image`: 图片
- `video`: 视频
- `document`: 文档

#### FileName vs SaveName
- `FileName`: 生成的唯一文件名（如 `20231216_abc123.jpg`）
- `SaveName`: 完整保存路径（如 `20231216/image/20231216_abc123.jpg`）

#### RealName vs IndexFileName
- `RealName`: 用户上传的原始文件名（如 `我的照片.jpg`）
- `IndexFileName`: 不含扩展名的文件名（如 `我的照片`）

#### UrlParam
- 存储引擎特定的 URL 参数
- 如 Google Storage 的访问 token

---

## URL 生成机制

### 1. 完整 URL 组成

```
{FileUrl}/{SaveName}{UrlParam}
```

**示例**:

**本地存储**:
```
https://example.com/public/uploads/20231216/image/20231216_abc123.jpg
```

**Google Storage**:
```
https://storage.googleapis.com/my-bucket/uploads/20231216/image/20231216_abc123.jpg?token=xxx
```

### 2. GetUrl 方法

```go
func (f *File) GetUrl(baseURL string) string {
    if f.Storage == "local" {
        return fmt.Sprintf("%s/%s", baseURL, f.SaveName)
    } else {
        return fmt.Sprintf("%s/%s%s", f.FileUrl, f.SaveName, f.UrlParam)
    }
}
```

---

## 使用场景

### 场景1: 商品图片上传

```go
// 前端上传商品图片
file, _ := ctx.FormFile("image")
fileReader, _ := file.Open()
defer fileReader.Close()

// 调用上传服务
result, err := uploadFileSrv.UploadImage(
    ctx,
    fileReader,
    file.Filename,
    file.Size,
    0,           // 不分组
    "product",   // 来源：商品
)

// 结果：
// - 生成 500px 缩略图
// - 返回文件 UUID 和访问 URL
// - 商品表保存文件 UUID
```

### 场景2: 营销视频上传

```go
// 上传营销活动视频
file, _ := ctx.FormFile("video")
fileReader, _ := file.Open()
defer fileReader.Close()

result, err := uploadFileSrv.UploadVideo(
    ctx,
    fileReader,
    file.Filename,
    file.Size,
    marketingGroupId, // 营销分组
    30,               // 最大 30MB
)

// 验证：
// - 文件格式（MP4、AVI等）
// - 文件大小不超过 30MB
```

### 场景3: 采购单附件上传

```go
// 上传采购单附件（PDF、图片等）
file, _ := ctx.FormFile("attachment")
fileReader, _ := file.Open()
defer fileReader.Close()

result, err := uploadFileSrv.UploadDocument(
    ctx,
    fileReader,
    file.Filename,
    file.Size,
    purchaseGroupId, // 采购分组
)

// 支持：
// - PDF、Word、Excel
// - JPG、PNG等图片
// - 统一 20MB 限制
```

### 场景4: 获取文件信息生成下载链接

```go
// 根据文件 UUID 获取信息
fileInfo, err := uploadFileSrv.GetUploadFile(ctx, fileUuid)
if err != nil {
    return err
}

// 生成下载链接
downloadUrl := fileInfo.FilePath

// 或者在前端使用
response := gin.H{
    "file_id":   fileInfo.Uuid,
    "file_name": fileInfo.RealName,
    "file_url":  fileInfo.FilePath,
    "file_size": fileInfo.FileSize,
}
```

### 场景5: 批量上传头像

```go
// 批量上传员工头像
for _, file := range files {
    fileReader, _ := file.Open()
    defer fileReader.Close()
    
    result, _ := uploadFileSrv.UploadImage(
        ctx,
        fileReader,
        file.Filename,
        file.Size,
        avatarGroupId,
        "avatar", // 生成 500px 缩略图
    )
    
    // 更新员工表
    staffRepo.UpdateAvatar(staffUuid, result.Uuid)
}
```

---

## 最佳实践

### 1. 文件上传前端处理

```javascript
// 前端上传示例
async function uploadImage(file) {
    // 1. 客户端验证
    if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) {
        alert('只支持 JPG、PNG、WEBP 格式');
        return;
    }
    
    // 2. 大小预检查
    if (file.size > 10 * 1024 * 1024) {
        alert('文件不能超过 10MB');
        return;
    }
    
    // 3. 构建 FormData
    const formData = new FormData();
    formData.append('image', file);
    formData.append('group_id', groupId);
    formData.append('source', 'product');
    
    // 4. 上传
    const response = await fetch('/api/v1/upload/image', {
        method: 'POST',
        body: formData,
    });
    
    const result = await response.json();
    return result.data;
}
```

### 2. 文件分组管理

```go
// 建议的分组策略
const (
    FileGroupProduct   = 1  // 商品图片
    FileGroupAvatar    = 2  // 头像
    FileGroupMarketing = 3  // 营销素材
    FileGroupDocument  = 4  // 文档附件
    FileGroupBanner    = 5  // 广告横幅
)

// 按分组查询文件
func GetFilesByGroup(groupId uint64) []model.File {
    var files []model.File
    db.Where("group_uuid = ? AND delete_time = 0", groupId).Find(&files)
    return files
}
```

### 3. 缩略图尺寸选择

```go
// 根据用途选择缩略图尺寸
type UploadContext struct {
    UsageType string
}

func getThumbSize(usageType string) int {
    switch usageType {
    case "avatar":
        return 200    // 头像小图
    case "product_list":
        return 500    // 列表缩略图
    case "product_detail":
        return 1500   // 详情大图
    case "banner":
        return 5000   // 广告横幅（不缩放）
    default:
        return 500
    }
}
```

### 4. 文件清理策略

```go
// 定期清理未使用的文件
func CleanupUnusedFiles() {
    // 1. 查找创建超过 30 天且未被引用的文件
    var files []model.File
    db.Where("create_time < ? AND delete_time = 0", 
        time.Now().AddDate(0, 0, -30).Unix()).Find(&files)
    
    // 2. 检查是否被引用
    for _, file := range files {
        if !isFileReferenced(file.Uuid) {
            // 3. 软删除
            db.Model(&file).Update("delete_time", time.Now().Unix())
            
            // 4. 从存储中删除（可选）
            // deleteFromStorage(file)
        }
    }
}
```

### 5. 错误处理

```go
// 统一的错误处理
result, err := uploadFileSrv.UploadImage(ctx, ...)
if err != nil {
    // 根据错误类型处理
    if strings.Contains(err.Error(), "仅支持") {
        // 文件类型错误
        return errors.New("文件格式不正确")
    } else if strings.Contains(err.Error(), "不能超过") {
        // 文件大小错误
        return errors.New("文件太大")
    } else {
        // 其他错误
        logger.Error("文件上传失败", zap.Error(err))
        return errors.New("上传失败，请稍后重试")
    }
}
```

---

## 性能优化

### 1. 异步上传

```go
// 大文件异步上传
func AsyncUploadFile(ctx context.Context, file *multipart.FileHeader) (uint64, error) {
    // 1. 立即返回任务ID
    taskId := generateTaskId()
    
    // 2. 异步上传
    go func() {
        fileReader, _ := file.Open()
        defer fileReader.Close()
        
        result, err := uploadFileSrv.UploadVideo(ctx, fileReader, ...)
        
        // 3. 更新任务状态
        updateTaskStatus(taskId, result, err)
    }()
    
    return taskId, nil
}
```

### 2. 缓存文件信息

```go
// 缓存文件访问URL
type FileCache struct {
    cache cache.Cache
}

func (c *FileCache) GetFileUrl(fileUuid uint64) (string, error) {
    // 1. 尝试从缓存获取
    cacheKey := fmt.Sprintf("file:url:%d", fileUuid)
    if url, err := c.cache.Get(cacheKey); err == nil {
        return url.(string), nil
    }
    
    // 2. 从数据库获取
    file, err := uploadFileSrv.GetUploadFile(ctx, fileUuid)
    if err != nil {
        return "", err
    }
    
    // 3. 缓存结果（1小时）
    c.cache.Set(cacheKey, file.FilePath, time.Hour)
    
    return file.FilePath, nil
}
```

### 3. CDN 加速

```go
// 为静态文件配置 CDN
type CDNConfig struct {
    Enabled bool
    Domain  string
}

func (f *File) GetCDNUrl(cdnConfig CDNConfig) string {
    if cdnConfig.Enabled {
        return fmt.Sprintf("%s/%s", cdnConfig.Domain, f.SaveName)
    }
    return f.GetUrl(baseURL)
}
```

---

## 安全考虑

### 1. 文件类型验证

```go
// 双重验证：扩展名 + MIME类型
func validateFileType(file *multipart.FileHeader) error {
    // 1. 验证扩展名
    ext := filepath.Ext(file.Filename)
    if !isAllowedExt(ext) {
        return errors.New("不支持的文件格式")
    }
    
    // 2. 验证 MIME 类型
    fileReader, _ := file.Open()
    defer fileReader.Close()
    
    buffer := make([]byte, 512)
    fileReader.Read(buffer)
    mimeType := http.DetectContentType(buffer)
    
    if !isAllowedMimeType(mimeType) {
        return errors.New("文件内容不符合格式要求")
    }
    
    return nil
}
```

### 2. 文件名安全处理

```go
// 防止路径遍历攻击
func sanitizeFileName(fileName string) string {
    // 1. 移除路径分隔符
    fileName = filepath.Base(fileName)
    
    // 2. 移除特殊字符
    fileName = regexp.MustCompile(`[^a-zA-Z0-9._-]`).ReplaceAllString(fileName, "_")
    
    // 3. 限制长度
    if len(fileName) > 100 {
        fileName = fileName[:100]
    }
    
    return fileName
}
```

### 3. 访问权限控制

```go
// 验证文件访问权限
func CanAccessFile(ctx context.Context, fileUuid uint64) bool {
    file, _ := uploadFileSrv.GetUploadFile(ctx, fileUuid)
    
    // 检查文件是否属于当前公司
    if file.CompanyUuid != ctx.GetCompanyUuid() {
        return false
    }
    
    // 检查文件分组权限
    if !hasGroupPermission(ctx, file.GroupUuid) {
        return false
    }
    
    return true
}
```

### 4. 防止恶意上传

```go
// 限制上传频率
type RateLimiter struct {
    cache cache.Cache
}

func (r *RateLimiter) CheckUploadLimit(ctx context.Context) error {
    key := fmt.Sprintf("upload:limit:%d", ctx.GetStaffUuid())
    
    // 1分钟内最多上传10个文件
    count, _ := r.cache.Incr(key)
    if count == 1 {
        r.cache.Expire(key, time.Minute)
    }
    
    if count > 10 {
        return errors.New("上传过于频繁，请稍后再试")
    }
    
    return nil
}
```

---

## 错误处理

### 1. 常见错误

| 错误场景 | 错误消息 | 处理方式 |
|---------|---------|---------|
| 文件格式不支持 | "仅支持XXX格式" | 提示用户选择正确格式 |
| 文件过大 | "文件大小不能超过XXX" | 提示用户压缩或选择小文件 |
| 文件不存在 | "文件不存在" | 检查文件UUID |
| 存储引擎失败 | "创建存储引擎失败" | 检查配置和网络 |
| 上传失败 | "文件上传失败" | 重试或联系管理员 |
| 保存记录失败 | "保存文件记录失败" | 检查数据库连接 |

### 2. 错误处理示例

```go
result, err := uploadFileSrv.UploadImage(ctx, fileReader, fileName, fileSize, groupId, source)
if err != nil {
    // 记录详细错误日志
    logger.Error("文件上传失败",
        zap.Error(err),
        zap.String("file_name", fileName),
        zap.Int64("file_size", fileSize),
        zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
    )
    
    // 返回友好错误
    switch {
    case strings.Contains(err.Error(), "仅支持"):
        return gin.H{"error": "文件格式不支持，请上传 JPG、PNG 或 WEBP 格式的图片"}
    case strings.Contains(err.Error(), "不能超过"):
        return gin.H{"error": "文件太大，请压缩后重试"}
    default:
        return gin.H{"error": "上传失败，请稍后重试"}
    }
}
```

---

## 潜在改进点

### 1. 支持更多存储引擎

```go
// 扩展支持
type StorageDriver string

const (
    DriverLocal   StorageDriver = "local"
    DriverGoogle  StorageDriver = "google"
    DriverAWS     StorageDriver = "aws"      // Amazon S3
    DriverAliyun  StorageDriver = "aliyun"   // 阿里云 OSS
    DriverQiniu   StorageDriver = "qiniu"    // 七牛云
)
```

### 2. 图片处理增强

```go
// 支持更多图片处理
type ImageProcessOptions struct {
    Width      int    // 宽度
    Height     int    // 高度
    Quality    int    // 质量
    Format     string // 格式转换
    Watermark  bool   // 水印
    Crop       bool   // 裁剪
}

func (s *UploadFileSrvImpl) UploadImageWithProcess(
    ctx context.Context, 
    fileReader io.Reader, 
    options ImageProcessOptions,
) (*resp.UploadFileResp, error)
```

### 3. 断点续传

```go
// 支持大文件断点续传
type ChunkUpload struct {
    TaskId    string
    ChunkNo   int
    TotalChunks int
    Data      []byte
}

func (s *UploadFileSrvImpl) UploadChunk(ctx context.Context, chunk ChunkUpload) error
func (s *UploadFileSrvImpl) MergeChunks(ctx context.Context, taskId string) (*resp.UploadFileResp, error)
```

### 4. 文件预处理

```go
// 上传前预处理
type PreprocessOptions struct {
    AutoRotate    bool // 自动旋转
    RemoveExif    bool // 移除EXIF信息
    Compress      bool // 自动压缩
    MaxDimension  int  // 最大尺寸
}

func (s *UploadFileSrvImpl) PreprocessAndUpload(
    ctx context.Context,
    fileReader io.Reader,
    options PreprocessOptions,
) (*resp.UploadFileResp, error)
```

### 5. 智能压缩

```go
// 根据文件大小自动压缩
func smartCompress(fileSize int64, quality int) int {
    if fileSize > 5*1024*1024 { // > 5MB
        return 60 // 低质量
    } else if fileSize > 2*1024*1024 { // > 2MB
        return 75 // 中质量
    } else {
        return quality // 保持原质量
    }
}
```

### 6. 文件版本管理

```go
// 支持文件版本
type FileVersion struct {
    FileUuid    uint64
    Version     int
    SaveName    string
    UploadTime  time.Time
    UploadBy    uint64
}

func (s *UploadFileSrvImpl) UploadNewVersion(
    ctx context.Context,
    originalFileUuid uint64,
    fileReader io.Reader,
) (*resp.UploadFileResp, error)
```

---

## 相关文件

### DTO 定义
- `ttpos-server-go/app/dto/resp/upload_file.go` - 上传响应数据

### 数据模型
- `ttpos-server-go/app/model/file.go` - 文件模型

### 存储引擎
- `ttpos-server-go/pkg/storage/factory.go` - 存储引擎工厂
- `ttpos-server-go/pkg/storage/local.go` - 本地存储引擎
- `ttpos-server-go/pkg/storage/google.go` - Google Cloud 存储引擎

### 配置
- `config/google_bucket.go` - Google Cloud Storage 配置

---

## 总结

文件上传服务是系统资源管理的核心模块，具有以下特点：

1. **多类型支持**: 图片、视频、文档三大类文件
2. **多引擎支持**: 本地存储、云存储灵活切换
3. **智能处理**: 自动生成缩略图、格式验证
4. **完善的验证**: 文件类型、大小双重验证
5. **灵活配置**: 缩略图尺寸、大小限制可配置
6. **记录管理**: 完整的文件元信息持久化
7. **URL 生成**: 自动生成可访问的文件路径
8. **分组管理**: 支持文件按用途分组
9. **软删除**: 文件删除不物理删除，可恢复
10. **扩展性好**: 易于添加新的存储引擎和文件类型

该服务为整个系统提供了统一的文件管理能力，支持各种业务场景的文件上传需求。
