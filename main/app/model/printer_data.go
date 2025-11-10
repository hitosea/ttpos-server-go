package model

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/google"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

// PrinterLogData 打印日志数据表 ttpos_printer_log_data
type PrinterLogData struct {
	BaseModel
	LogUuid uint64 `gorm:"column:log_uuid;type:bigint(20) unsigned;default:0;comment:打印日志UUID;NOT NULL" json:"log_uuid"`
	Data    string `gorm:"column:data;type:text;comment:打印数据" json:"data"`
}

func (model *PrinterLogData) SetData(companyUuid uint64, data string) string {
	data = utils.GzipCompressData(data)

	// 如果谷歌云未配置
	if !config.GoogleBucket.Verification() {
		model.Data = data
		return data
	}

	// 获取文件名
	fileName := fmt.Sprintf("%d/print_data/%d.txt", companyUuid, model.LogUuid)
	md5Key := strings.ReplaceAll(fileName, "/", ":")

	// 存到redis
	err := cache.Global.Set(md5Key, data, 3*time.Minute)
	if err != nil {
		logger.Logger.Error("存到redis失败", zap.Error(err))
		fmt.Println("存到redis失败", err)
		model.Data = data
		return data
	}

	// 上传到谷歌云
	utils.Go(func() {
		// 设置10秒超时
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel() // 确保资源释放

		// 初始化存储桶
		ctx, err := google.InitBucket(
			ctx,
			config.GoogleBucket.GooglePrintBucketName,
			config.GoogleBucket.GetGoogleApplicationCredentialsFileName(),
		)
		if err != nil {
			logger.Logger.Error("初始化存储桶失败", zap.Error(err))
			fmt.Println("初始化存储桶失败", err)
			return
		}
		// 创建文件
		tmpfile, err := utils.CreateTempFile(fmt.Sprintf("print_data_%d.txt", model.LogUuid), data)
		if err != nil {
			logger.Logger.Error("创建测试文件失败", zap.Error(err))
			fmt.Println("创建测试文件失败", err)
			return
		}
		defer os.Remove(tmpfile.Name())

		// 重新打开文件用于读取
		file, err := os.Open(tmpfile.Name())
		if err != nil {
			logger.Logger.Error("重新打开文件用于读取失败", zap.Error(err))
			fmt.Println("重新打开文件用于读取失败", err)
			return
		}
		defer file.Close()

		// 测试上传（设置1分钟过期）
		_, err = google.UploadFile(ctx, fileName, file, 1)
		if err != nil {
			logger.Logger.Error("测试上传失败", zap.Error(err))
			fmt.Println("测试上传失败", err)
			return
		}
	})

	model.Data = fileName
	return fileName
}

func (model *PrinterLogData) GetData(isDelCache bool) string {
	if model.Data == "" || !strings.HasSuffix(model.Data, ".txt") {
		return model.Data
	}

	// 添加缓存预热
	md5Key := strings.ReplaceAll(model.Data, "/", ":")
	if data, isExist := cache.Global.Get(md5Key); isExist {
		model.Data = data.(string)
		if isDelCache {
			cache.Global.Del(md5Key)
		}
		return model.Data
	}

	// 下载文件
	if !config.GoogleBucket.Verification() {
		return ""
	}

	// 设置10秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // 确保资源释放

	ctx, err := google.InitBucket(
		ctx,
		config.GoogleBucket.GooglePrintBucketName,
		config.GoogleBucket.GetGoogleApplicationCredentialsFileName(),
	)
	if err != nil {
		logger.Logger.Error("初始化存储桶失败", zap.Error(err))
		fmt.Println("初始化存储桶失败", err)
		return ""
	}

	data, err := google.DownloadFileContent(ctx, model.Data)
	if err != nil {
		logger.Logger.Error("下载文件失败", zap.Error(err))
		fmt.Println("下载文件失败", err)
		return ""
	}

	return data
}

// 压缩数据
func (model *PrinterLogData) CompressData() string {
	return utils.GzipCompressData(model.Data)
}

// 还原压缩的数据
func (model *PrinterLogData) DecompressData() string {
	return utils.GzipDecompressData(model.GetData(false))
}
