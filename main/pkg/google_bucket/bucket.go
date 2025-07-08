package google_bucket

import (
	"context"
	"fmt"
	"io"
	"time"
	"ttpos-server-go/pkg/utils"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var (
	bucketClient *storage.Client
	bucketName   string
)

// InitBucket 初始化存储桶
func InitBucket(ctx context.Context, bucketNameStr string, credentialsFile string) (context.Context, error) {
	if bucketNameStr == "" || credentialsFile == "" {
		return nil, fmt.Errorf("创建存储客户端失败")
	}

	var err error
	bucketName = bucketNameStr

	// 创建存储客户端
	bucketClient, err = storage.NewClient(
		ctx,
		option.WithCredentialsFile(credentialsFile),
	)
	if err != nil {
		return nil, fmt.Errorf("创建存储客户端失败: %v", err)
	}

	// 检查存储桶是否存在
	_, err = bucketClient.Bucket(bucketName).Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("检查存储桶失败: %v", err)
	}

	return ctx, nil
}

// GetBucket 获取存储桶句柄
func GetBucket(ctx context.Context) *storage.BucketHandle {
	return bucketClient.Bucket(bucketName)
}

// ListAllFiles 列出存储桶中的所有文件
func ListAllFiles(ctx context.Context) ([]*storage.ObjectAttrs, error) {
	bucket := GetBucket(ctx)
	var files []*storage.ObjectAttrs

	// 创建查询对象
	query := &storage.Query{
		// 不设置Prefix，获取所有文件
	}

	// 获取迭代器
	it := bucket.Objects(ctx, query)

	// 遍历所有文件
	for {
		attrs, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("列出文件失败: %v", err)
		}
		files = append(files, attrs)
	}

	return files, nil
}

// ListFilesWithPrefix 列出存储桶中指定前缀的文件
func ListFilesWithPrefix(ctx context.Context, prefix string) ([]*storage.ObjectAttrs, error) {
	bucket := GetBucket(ctx)
	var files []*storage.ObjectAttrs

	// 创建查询对象
	query := &storage.Query{
		Prefix: prefix,
	}

	// 获取迭代器
	it := bucket.Objects(ctx, query)

	// 遍历所有文件
	for {
		attrs, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("列出文件失败: %v", err)
		}
		files = append(files, attrs)
	}

	return files, nil
}

// UploadFile 上传文件到存储桶并返回签名URL
func UploadFile(ctx context.Context, objectName string, file io.Reader, expirationDays int) (string, error) {
	bucket := GetBucket(ctx)
	obj := bucket.Object(objectName)
	writer := obj.NewWriter(ctx)

	// 设置对象的元数据
	writer.ObjectAttrs.ContentType = "application/octet-stream"
	writer.ObjectAttrs.CacheControl = "no-cache"

	// 复制文件内容到存储桶
	if _, err := io.Copy(writer, file); err != nil {
		return "", fmt.Errorf("上传文件失败: %v", err)
	}

	// 关闭写入器
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("关闭写入器失败: %v", err)
	}

	// // 生成签名URL，默认有效期永久（请注意：签名URL是临时的，最大有效期取决于服务帐号和设置）
	// url, err := GetSignedURL(ctx, objectName, time.Duration(expirationDays)*24*time.Hour)
	// if err != nil {
	// 	return "", fmt.Errorf("生成签名URL失败: %v", err)
	// }

	return objectName, nil
}

// GetFileExpiration 获取文件的过期时间
func GetFileExpiration(ctx context.Context, objectName string) (time.Time, error) {
	bucket := GetBucket(ctx)
	obj := bucket.Object(objectName)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("获取文件属性失败: %v", err)
	}

	// 从元数据中获取过期时间
	if expirationTime, ok := attrs.Metadata["expiration_time"]; ok {
		return time.Parse(time.RFC3339, expirationTime)
	}

	return time.Time{}, fmt.Errorf("文件没有设置过期时间")
}

// DownloadFile 从存储桶下载文件
func DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	bucket := GetBucket(ctx)
	obj := bucket.Object(objectName)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

// DownloadFile 从存储桶下载文件
func DownloadFileContent(ctx context.Context, objectName string) (string, error) {
	bucket := GetBucket(ctx)
	obj := bucket.Object(objectName)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// DeleteFile 从存储桶删除文件
func DeleteFile(ctx context.Context, objectName string) error {
	bucket := GetBucket(ctx)
	obj := bucket.Object(objectName)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}
	return nil
}

// GetSignedURL 获取文件的签名URL
func GetSignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
	bucket := GetBucket(ctx)
	opts := &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(expiration),
	}

	url, err := bucket.SignedURL(objectName, opts)
	if err != nil {
		return "", fmt.Errorf("生成签名URL失败: %v", err)
	}

	return url, nil
}

// ListFiles 列出存储桶中的文件（仅返回文件名）
// 提示：如果需要更详细的文件属性，请使用 ListFilesWithPrefix 函数。
func ListFiles(ctx context.Context, prefix string) ([]string, error) {
	bucket := GetBucket(ctx)
	query := &storage.Query{
		Prefix: prefix,
	}

	var files []string
	it := bucket.Objects(ctx, query)

	fmt.Println(utils.ToJsonString(it))

	for {
		attrs, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("列出文件失败: %v", err)
		}
		files = append(files, attrs.Name)
	}

	return files, nil
}

// GetFileMetadata 获取文件元数据
func GetFileMetadata(ctx context.Context, objectName string) (*storage.ObjectAttrs, error) {
	bucket := GetBucket(ctx)
	obj := bucket.Object(objectName)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取文件元数据失败: %v", err)
	}
	return attrs, nil
}

// UpdateFileMetadata 更新文件元数据
func UpdateFileMetadata(ctx context.Context, objectName string, metadata map[string]string) error {
	bucket := GetBucket(ctx)
	obj := bucket.Object(objectName)

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("获取文件属性失败: %v", err)
	}

	attrs.Metadata = metadata
	_, err = obj.Update(ctx, storage.ObjectAttrsToUpdate{
		Metadata: metadata,
	})
	if err != nil {
		return fmt.Errorf("更新文件元数据失败: %v", err)
	}

	return nil
}

// Close 关闭存储客户端
func Close() error {
	if bucketClient != nil {
		return bucketClient.Close()
	}
	return nil
}

// DeleteFilesWithPrefix 删除指定前缀下的所有文件
func DeleteFilesWithPrefix(ctx context.Context, prefix string) error {
	bucket := GetBucket(ctx)

	// 创建查询对象，指定前缀
	query := &storage.Query{
		Prefix: prefix,
	}

	it := bucket.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == io.EOF {
			break // 遍历结束
		}
		if err != nil {
			return fmt.Errorf("列出文件失败: %v", err)
		}

		// 删除文件
		obj := bucket.Object(attrs.Name)
		if err := obj.Delete(ctx); err != nil {
			return fmt.Errorf("删除文件 %s 失败: %v", attrs.Name, err)
		}
		fmt.Printf("已删除文件: %s\n", attrs.Name)
	}
	return nil
}

// GetBucketLifecycle 获取存储桶生命周期规则
func GetBucketLifecycle(ctx context.Context) (storage.Lifecycle, error) {
	bucket := GetBucket(ctx)
	attrs, err := bucket.Attrs(ctx)
	if err != nil {
		return storage.Lifecycle{}, fmt.Errorf("获取存储桶属性失败: %v", err)
	}
	return attrs.Lifecycle, nil
}

// CancelBucketLifecycle 取消存储桶生命周期规则
func CancelBucketLifecycle(ctx context.Context) error {
	bucket := GetBucket(ctx)
	// 设置空生命周期规则
	lifecycle := storage.Lifecycle{
		Rules: []storage.LifecycleRule{},
	}
	_, err := bucket.Update(ctx, storage.BucketAttrsToUpdate{
		Lifecycle: &lifecycle,
	})
	if err != nil {
		return fmt.Errorf("取消存储桶生命周期失败: %v", err)
	}
	return nil
}
