package engines

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ttpos-server-go/pkg/storage"

	gcStorage "cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// Google Google云存储引擎
type Google struct {
	*storage.Server
	config          *storage.GoogleConfig
	credentialsPath string
	urlParam        string
	companyUuid     uint64
}

// NewGoogle 创建Google云存储引擎实例
func NewGoogle(config *storage.GoogleConfig) (*Google, error) {
	g := &Google{
		Server: storage.NewServer(),
		config: config,
	}

	// 设置凭证文件路径
	g.credentialsPath = filepath.Join("../certificate", config.CredentialsFile)

	// 检查凭证文件是否存在
	if _, err := os.Stat(g.credentialsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("未找到云存储配置文件: %s", g.credentialsPath)
	}

	return g, nil
}

// Upload 上传文件
func (g *Google) Upload(thumb int) (string, error) {
	ctx := context.Background()

	// 创建存储客户端
	client, err := gcStorage.NewClient(ctx, option.WithCredentialsFile(g.credentialsPath))
	if err != nil {
		g.SetError(fmt.Sprintf("创建存储客户端失败: %v", err))
		return "", err
	}
	defer client.Close()

	// 获取存储桶
	bucket := client.Bucket(g.config.Bucket)

	// 构建上传路径
	uploadPath := g.getUploadPath()
	saveName := filepath.Join(uploadPath, g.GetFileName())

	// 准备上传数据
	var uploadData io.Reader
	var uploadErr error

	if thumb > 0 && g.isImage() {
		// 生成缩略图
		uploadData, uploadErr = g.generateThumbnail(thumb)
		if uploadErr != nil {
			g.SetError(fmt.Sprintf("生成缩略图失败: %v", uploadErr))
			return "", uploadErr
		}
	} else {
		// 使用原始文件
		uploadData, uploadErr = g.getFileReader()
		if uploadErr != nil {
			g.SetError(fmt.Sprintf("获取文件内容失败: %v", uploadErr))
			return "", uploadErr
		}
	}

	// 上传文件到Google Cloud Storage
	objectName := filepath.Join(g.config.UploadsCatalogue, saveName)
	obj := bucket.Object(objectName)
	writer := obj.NewWriter(ctx)

	// 设置内容类型
	if g.GetContentType() != "" {
		writer.ContentType = g.GetContentType()
	}

	// 复制数据
	if _, err := io.Copy(writer, uploadData); err != nil {
		writer.Close()
		g.SetError(fmt.Sprintf("上传文件失败: %v", err))
		return "", err
	}

	// 关闭写入器
	if err := writer.Close(); err != nil {
		g.SetError(fmt.Sprintf("完成上传失败: %v", err))
		return "", err
	}

	// 生成签名URL参数（Google Cloud Storage v1 API方式）
	g.urlParam, err = bucket.SignedURL(objectName, &gcStorage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(time.Hour * 24 * 365 * 100), // 1百年
	})
	if err != nil {
		g.SetError(fmt.Sprintf("生成签名URL参数失败: %v", err))
		return "", err
	}

	// 获取urlParam中?后面的参数
	urlParam := strings.Split(g.urlParam, "?")[1]
	g.urlParam = urlParam

	return saveName, nil
}

// Delete 删除文件
func (g *Google) Delete(fileName string) error {
	ctx := context.Background()

	// 创建存储客户端
	client, err := gcStorage.NewClient(ctx, option.WithCredentialsFile(g.credentialsPath))
	if err != nil {
		return fmt.Errorf("创建存储客户端失败: %v", err)
	}
	defer client.Close()

	// 获取存储桶
	bucket := client.Bucket(g.config.Bucket)

	// 删除对象
	objectName := filepath.Join(g.config.UploadsCatalogue, fileName)
	obj := bucket.Object(objectName)

	return obj.Delete(ctx)
}

// GetUrlParam 获取URL参数
func (g *Google) GetUrlParam() string {
	return g.urlParam
}

// getUploadPath 获取上传路径
func (g *Google) getUploadPath() string {
	// 根据应用ID确定子目录
	var subDir string
	if g.companyUuid > 0 {
		subDir = "shop" + strconv.FormatUint(g.companyUuid, 10)
	} else {
		subDir = "saas"
	}

	// 添加日期子目录
	dateDir := time.Now().Format("20060102")

	return filepath.Join(subDir, dateDir)
}

// isImage 检查是否为图片文件
func (g *Google) isImage() bool {
	return strings.HasPrefix(g.GetContentType(), "image/")
}

// getFileReader 获取文件读取器
func (g *Google) getFileReader() (io.Reader, error) {
	if g.GetFile() != nil {
		return g.GetFile(), nil
	} else if g.GetFilePath() != "" {
		file, err := os.Open(g.GetFilePath())
		if err != nil {
			return nil, err
		}
		// 注意：调用者需要负责关闭文件
		return file, nil
	}

	return nil, fmt.Errorf("没有可用的文件源")
}

// generateThumbnail 生成缩略图
func (g *Google) generateThumbnail(size int) (io.Reader, error) {
	// 获取原始图片
	var src image.Image
	var err error

	if g.GetFile() != nil {
		src, _, err = image.Decode(g.GetFile())
	} else if g.GetFilePath() != "" {
		file, openErr := os.Open(g.GetFilePath())
		if openErr != nil {
			return nil, openErr
		}
		defer file.Close()

		src, _, err = image.Decode(file)
	} else {
		return nil, fmt.Errorf("没有可用的图片源")
	}

	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %v", err)
	}

	// 创建缩略图
	thumbnail := g.resizeImage(src, size, size)

	// 将图片编码为PNG格式
	var buf bytes.Buffer
	if err := png.Encode(&buf, thumbnail); err != nil {
		return nil, fmt.Errorf("编码缩略图失败: %v", err)
	}

	return &buf, nil
}

// resizeImage 调整图片尺寸（简单的缩放实现）
func (g *Google) resizeImage(src image.Image, width, height int) image.Image {
	// 这里可以使用更复杂的图片处理库，如 github.com/disintegration/imaging
	// 目前使用简单的实现
	bounds := src.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()

	// 计算缩放比例
	scaleX := float64(width) / float64(dx)
	scaleY := float64(height) / float64(dy)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	newWidth := int(float64(dx) * scale)
	newHeight := int(float64(dy) * scale)

	// 创建新图像
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// 简单的最近邻缩放
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) / scale)
			srcY := int(float64(y) / scale)
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}

	return dst
}

// SetCompanyUuid 设置公司ID
func (g *Google) SetCompanyUuid(companyUuid uint64) {
	g.companyUuid = companyUuid
}
