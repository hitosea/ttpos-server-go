package storage

import (
	"fmt"
	"os"
	"strings"
)

// Factory 存储工厂
type Factory struct {
	config *StorageConfig
}

// NewFactory 创建存储工厂实例
func NewFactory(config *StorageConfig) *Factory {
	return &Factory{
		config: config,
	}
}

// CreateEngine 创建存储引擎
func (f *Factory) CreateEngine(engineName ...string) (StorageEngine, error) {
	// 确定使用哪个存储引擎
	var targetEngine string
	if len(engineName) > 0 && engineName[0] != "" {
		targetEngine = engineName[0]
	} else {
		// 优先使用环境变量
		if env := os.Getenv("STORAGE_DRIVER"); env != "" {
			targetEngine = strings.ToLower(env)
		} else {
			// 使用配置中的默认值
			targetEngine = f.config.Default
		}
	}

	// 确保引擎名称为小写
	targetEngine = strings.ToLower(targetEngine)

	// 检查引擎配置是否存在
	engineConfig, exists := f.config.Engine[targetEngine]
	if !exists {
		return nil, fmt.Errorf("未找到存储引擎配置: %s", targetEngine)
	}

	// 根据引擎名称创建对应的实例
	switch targetEngine {
	case "local":
		return f.createLocalEngine(engineConfig)
	case "google":
		return f.createGoogleEngine(engineConfig)
	default:
		return nil, fmt.Errorf("不支持的存储引擎: %s", targetEngine)
	}
}

// createLocalEngine 创建本地存储引擎
func (f *Factory) createLocalEngine(config map[string]any) (StorageEngine, error) {
	localConfig := &LocalConfig{}

	// 解析配置
	if uploadPath, ok := config["upload_path"].(string); ok {
		localConfig.UploadPath = uploadPath
	} else {
		localConfig.UploadPath = "public/uploads" // 默认值
	}

	// 直接创建本地存储引擎，避免循环导入
	return NewLocalEngine(localConfig), nil
}

// createGoogleEngine 创建Google云存储引擎
func (f *Factory) createGoogleEngine(config map[string]any) (StorageEngine, error) {
	googleConfig := &GoogleConfig{}

	// 解析配置
	if credentialsFile, ok := config["credentials_file"].(string); ok {
		googleConfig.CredentialsFile = credentialsFile
	} else {
		return nil, fmt.Errorf("Google云存储缺少credentials_file配置")
	}

	if bucket, ok := config["bucket"].(string); ok {
		googleConfig.Bucket = bucket
	} else {
		return nil, fmt.Errorf("Google云存储缺少bucket配置")
	}

	if uploadsCatalogue, ok := config["uploads_catalogue"].(string); ok {
		googleConfig.UploadsCatalogue = uploadsCatalogue
	} else {
		return nil, fmt.Errorf("Google云存储缺少uploads_catalogue配置")
	}

	if domain, ok := config["domain"].(string); ok {
		googleConfig.Domain = domain
	}

	// 直接创建Google存储引擎，避免循环导入
	return NewGoogleEngine(googleConfig)
}

// ValidateConfig 验证存储配置
func (f *Factory) ValidateConfig() error {
	if f.config == nil {
		return fmt.Errorf("存储配置不能为空")
	}

	if f.config.Default == "" {
		return fmt.Errorf("必须指定默认存储引擎")
	}

	if f.config.Engine == nil || len(f.config.Engine) == 0 {
		return fmt.Errorf("必须配置至少一个存储引擎")
	}

	// 检查默认引擎是否在配置中
	if _, exists := f.config.Engine[f.config.Default]; !exists {
		return fmt.Errorf("默认存储引擎 %s 没有相应的配置", f.config.Default)
	}

	// 验证Google云存储配置
	if googleConfig, exists := f.config.Engine["google"]; exists {
		if credentialsFile, ok := googleConfig["credentials_file"].(string); !ok || credentialsFile == "" {
			return fmt.Errorf("Google云存储缺少credentials_file配置")
		}
		if bucket, ok := googleConfig["bucket"].(string); !ok || bucket == "" {
			return fmt.Errorf("Google云存储缺少bucket配置")
		}
		if uploadsCatalogue, ok := googleConfig["uploads_catalogue"].(string); !ok || uploadsCatalogue == "" {
			return fmt.Errorf("Google云存储缺少uploads_catalogue配置")
		}
	}

	return nil
}

// GetAvailableEngines 获取可用的存储引擎列表
func (f *Factory) GetAvailableEngines() []string {
	engines := make([]string, 0, len(f.config.Engine))
	for engine := range f.config.Engine {
		engines = append(engines, engine)
	}
	return engines
}

// GetDefaultEngine 获取默认存储引擎名称
func (f *Factory) GetDefaultEngine() string {
	// 优先使用环境变量
	if env := os.Getenv("STORAGE_DRIVER"); env != "" {
		return strings.ToLower(env)
	}
	return f.config.Default
}

func (f *Factory) GetDefaultEngineConfig() map[string]any {
	return f.config.Engine[f.GetDefaultEngine()]
}
