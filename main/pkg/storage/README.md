# 文件存储模块

基于工厂模式的文件存储模块，支持本地存储和Google云存储，可通过环境变量 `STORAGE_DRIVER` 动态切换存储引擎。

## 特性

- 🏭 **工厂模式设计**：支持多种存储引擎，易于扩展
- 🔄 **动态切换**：通过环境变量或配置动态选择存储引擎
- 📁 **本地存储**：支持本地文件系统存储
- ☁️ **Google云存储**：支持Google Cloud Storage
- 🖼️ **缩略图生成**：支持图片缩略图生成
- 🗃️ **数据库集成**：完整的文件记录管理
- 🔒 **类型安全**：完整的接口定义和类型检查

## 目录结构

```
pkg/storage/
├── interface.go        # 存储接口定义
├── server.go          # 基础服务器类
├── engines.go         # 存储引擎实现
├── factory.go         # 存储工厂
├── example_config.go  # 配置示例
└── README.md          # 说明文档

app/
├── model/upload_file.go              # 上传文件模型
├── repository/upload_file_repository.go  # 数据访问层
├── service/upload_file_service.go        # 业务逻辑层
└── dto/
    ├── req/upload_file.go             # 请求DTO
    └── resp/upload_file.go            # 响应DTO
```

## 使用方法

### 1. 配置存储引擎

```go
package main

import "ttpos-server-go/pkg/storage"

// 本地存储配置
config := &storage.StorageConfig{
    Default: "local",
    Engine: map[string]map[string]interface{}{
        "local": {
            "upload_path": "public/uploads",
        },
    },
}

// Google云存储配置
config := &storage.StorageConfig{
    Default: "google",
    Engine: map[string]map[string]interface{}{
        "google": {
            "credentials_file":   "service-account-key.json",
            "bucket":            "your-bucket-name", 
            "uploads_catalogue": "uploads",
            "domain":            "https://storage.googleapis.com",
        },
    },
}
```

### 2. 创建存储工厂

```go
// 创建工厂
factory := storage.NewFactory(config)

// 验证配置
if err := factory.ValidateConfig(); err != nil {
    log.Fatal("配置验证失败:", err)
}

// 创建存储引擎（使用默认配置）
engine, err := factory.CreateEngine()
if err != nil {
    log.Fatal("创建存储引擎失败:", err)
}

// 或者指定特定引擎
engine, err := factory.CreateEngine("local")
```

### 3. 上传文件

```go
import (
    "os"
    "ttpos-server-go/pkg/storage"
)

// 打开文件
file, err := os.Open("test.jpg")
if err != nil {
    return err
}
defer file.Close()

// 获取文件信息
fileInfo, _ := file.Stat()

// 设置文件信息
err = engine.SetUploadFile(file, "test.jpg", "image/jpeg", fileInfo.Size())
if err != nil {
    return err
}

// 上传文件（生成500px缩略图）
saveName, err := engine.Upload(500)
if err != nil {
    return err
}

// 获取文件名和URL参数
fileName := engine.GetFileName()
urlParam := engine.GetUrlParam()
```

### 4. 使用服务层

```go
import (
    "ttpos-server-go/app/service"
    "ttpos-server-go/pkg/database"
    "ttpos-server-go/pkg/storage"
)

// 创建服务
dbm := database.NewDBManager() // 数据库管理器
factory := storage.NewFactory(config)
uploadSrv := service.NewUploadFileSrv(dbm, factory)

// 上传图片
resp, err := uploadSrv.UploadImage(
    ctx,           // 上下文
    fileReader,    // 文件读取器
    "test.jpg",    // 文件名
    "image/jpeg",  // MIME类型
    12345,         // 文件大小
    100,           // 分组ID
    "admin"        // 来源标识
)

// 上传视频
resp, err := uploadSrv.UploadVideo(
    ctx,
    fileReader,
    "test.mp4",
    "video/mp4", 
    5242880, // 5MB
    200,     // 分组ID
    30       // 最大文件大小(MB)
)

// 获取文件列表
req := &req.FileListReq{
    GroupUuid: 100,
    FileType:  "image",
    PageNo:    1,
    PageSize:  20,
}
listResp, err := uploadSrv.GetFileList(ctx, req)

// 删除文件
err = uploadSrv.DeleteFile(ctx, 12345) // 文件UUID
```

### 5. 环境变量配置

```bash
# 设置存储驱动
export STORAGE_DRIVER=local    # 使用本地存储
export STORAGE_DRIVER=google   # 使用Google云存储
```

当设置了环境变量后，工厂会优先使用环境变量指定的存储引擎。

## 接口说明

### StorageEngine 接口

```go
type StorageEngine interface {
    // 上传文件，thumb为缩略图尺寸（0为不生成）
    Upload(thumb int) (string, error)
    
    // 删除文件
    Delete(fileName string) error
    
    // 设置上传文件信息
    SetUploadFile(fileReader io.Reader, fileName string, contentType string, fileSize int64) error
    
    // 设置本地文件路径
    SetLocalFile(filePath string) error
    
    // 获取生成的文件名
    GetFileName() string
    
    // 获取错误信息
    GetError() string
    
    // 获取URL参数（如签名参数）
    GetUrlParam() string
}
```

### 文件模型

```go
type UploadFile struct {
    Uuid          uint64 `json:"uuid"`            // 文件UUID
    GroupUuid     uint64 `json:"group_uuid"`      // 分组UUID
    Storage       string `json:"storage"`         // 存储引擎
    FileUrl       string `json:"file_url"`        // 访问域名
    FileName      string `json:"file_name"`       // 生成的文件名
    SaveName      string `json:"save_name"`       // 保存路径
    FileSize      int64  `json:"file_size"`       // 文件大小
    FileType      string `json:"file_type"`       // 文件类型
    Extension     string `json:"extension"`       // 扩展名
    RealName      string `json:"real_name"`       // 原始文件名
    IndexFileName string `json:"index_file_name"` // 索引文件名
    UrlParam      string `json:"url_param"`       // URL参数
    CreateTime    int    `json:"create_time"`     // 创建时间
}
```

## 支持的文件类型

### 图片
- JPG/JPEG
- PNG
- WEBP

### 视频
- AVI
- MPEG
- MOV
- MP4

## 特性说明

### 缩略图生成
- 支持图片缩略图生成
- 可指定缩略图尺寸
- 自动保持宽高比

### 文件命名
- 自动生成唯一文件名
- 格式：`时间戳_随机字符串.扩展名`
- 避免文件名冲突

### 路径组织
- 按日期自动创建目录
- 支持多应用隔离
- 格式：`应用ID/YYYYMMDD/文件名`

### 错误处理
- 完善的错误信息
- 自动回滚机制
- 文件上传失败时自动清理

## 注意事项

1. **Google云存储配置**：需要有效的服务账号密钥文件
2. **文件权限**：确保上传目录有写权限
3. **文件大小**：注意文件大小限制
4. **并发安全**：存储引擎是并发安全的
5. **资源清理**：及时关闭文件句柄

## 扩展存储引擎

要添加新的存储引擎，需要：

1. 实现 `StorageEngine` 接口
2. 在工厂中添加创建逻辑
3. 添加相应的配置结构体

```go
// 实现新的存储引擎
type S3Engine struct {
    *Server
    config *S3Config
}

func (s *S3Engine) Upload(thumb int) (string, error) {
    // 实现S3上传逻辑
}

// 在工厂中添加
func (f *Factory) createS3Engine(config map[string]interface{}) (StorageEngine, error) {
    // 创建S3引擎
}
```
