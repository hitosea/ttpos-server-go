package storage

// ExampleConfig 示例配置
func ExampleConfig() *StorageConfig {
	return &StorageConfig{
		Default: "local", // 默认使用本地存储
		Engine: map[string]map[string]interface{}{
			"local": {
				"upload_path": "public/uploads",
			},
			"google": {
				"credentials_file":  "service-account-key.json",
				"bucket":            "your-bucket-name",
				"uploads_catalogue": "uploads",
				"domain":            "https://storage.googleapis.com",
			},
		},
	}
}

// LocalStorageConfig 本地存储配置示例
func LocalStorageConfig() *StorageConfig {
	return &StorageConfig{
		Default: "local",
		Engine: map[string]map[string]interface{}{
			"local": {
				"upload_path": "public/uploads",
			},
		},
	}
}

// GoogleStorageConfig Google云存储配置示例
func GoogleStorageConfig() *StorageConfig {
	return &StorageConfig{
		Default: "google",
		Engine: map[string]map[string]interface{}{
			"google": {
				"credentials_file":  "google-service-account.json",
				"bucket":            "ttpos-uploads",
				"uploads_catalogue": "files",
				"domain":            "https://storage.googleapis.com",
			},
		},
	}
}
