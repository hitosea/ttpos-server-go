package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 全局配置
var (
	Server   ServerConf
	LLM      LLMConf
	Log      LogConf
	Database DatabaseConf
)

// ServerConf 服务配置
type ServerConf struct {
	Port string // 端口
	Mode string // 模式：debug/release
}

// LLMConf LLM配置
type LLMConf struct {
	Provider    string  // 提供商：openai/step/claude
	APIKey      string  // API密钥
	BaseURL     string  // API基础URL（OpenAI兼容接口）
	Model       string  // 模型名称
	MaxTokens   int     // 最大token数
	Temperature float32 // 温度参数（0-2）
	Enabled     bool    // 是否启用
}

// LogConf 日志配置
type LogConf struct {
	Dir       string // 日志路径
	Level     string // 日志级别
	MaxSize   int    // 每个日志文件最大尺寸(MB)
	MaxBackup int    // 日志文件最大数量
}

// DatabaseConf 数据库配置
type DatabaseConf struct {
	Host        string   // 主机
	Port        int      // 端口
	User        string   // 用户名
	Password    string   // 密码
	Database    string   // 数据库名
	TablePrefix string   // 表前缀
	MaxRows     int      // 查询最大行数限制
	AllowTables []string // 允许查询的表（白名单）
	DenyTables  []string // 禁止查询的表（黑名单）
}

// Init 初始化配置
func Init() error {
	// 尝试加载 .env 文件
	if err := godotenv.Load("../.env"); err != nil {
		// 尝试当前目录
		if err := godotenv.Load(".env"); err != nil {
			log.Printf("warning: .env file not found: %v\n", err)
		}
	}

	viper.AutomaticEnv()

	initServerConf()
	initLLMConf()
	initLogConf()
	initDatabaseConf()

	return nil
}

func initServerConf() {
	Server = ServerConf{
		Port: "8090",
		Mode: "release",
	}

	if port := viper.GetString("AI_SERVER_PORT"); port != "" {
		Server.Port = port
	}
	if mode := viper.GetString("AI_SERVER_MODE"); mode != "" {
		Server.Mode = mode
	}
}

func initLLMConf() {
	LLM = LLMConf{
		Provider:    "step",
		APIKey:      "",
		BaseURL:     "https://api.stepfun.com/v1",
		Model:       "step-3.5-flash",
		MaxTokens:   4096,
		Temperature: 0.7,
		Enabled:     false,
	}

	if provider := viper.GetString("LLM_PROVIDER"); provider != "" {
		LLM.Provider = provider
	}
	if apiKey := viper.GetString("LLM_API_KEY"); apiKey != "" {
		LLM.APIKey = apiKey
	}
	if baseURL := viper.GetString("LLM_BASE_URL"); baseURL != "" {
		LLM.BaseURL = baseURL
	}
	if model := viper.GetString("LLM_MODEL"); model != "" {
		LLM.Model = model
	}
	if maxTokens := viper.GetInt("LLM_MAX_TOKENS"); maxTokens > 0 {
		LLM.MaxTokens = maxTokens
	}
	if temp := viper.GetFloat64("LLM_TEMPERATURE"); temp > 0 {
		LLM.Temperature = float32(temp)
	}
	LLM.Enabled = viper.GetBool("LLM_ENABLED")

	// 验证配置
	if LLM.Enabled && LLM.APIKey == "" {
		log.Fatal("错误: LLM_ENABLED=true 但未设置 LLM_API_KEY")
	}
}

func initLogConf() {
	Log = LogConf{
		Dir:       "log",
		Level:     "info",
		MaxSize:   100,
		MaxBackup: 7,
	}

	if dir := viper.GetString("AI_LOG_DIR"); dir != "" {
		Log.Dir = dir
	}
	if level := viper.GetString("AI_LOG_LEVEL"); level != "" {
		Log.Level = level
	}
	if maxSize := viper.GetInt("AI_LOG_MAX_SIZE"); maxSize > 0 {
		Log.MaxSize = maxSize
	}
	if maxBackup := viper.GetInt("AI_LOG_MAX_BACKUP"); maxBackup > 0 {
		Log.MaxBackup = maxBackup
	}
}

func initDatabaseConf() {
	Database = DatabaseConf{
		Host:        "127.0.0.1",
		Port:        3306,
		User:        "",
		Password:    "",
		Database:    "",
		TablePrefix: "ttpos_",
		MaxRows:     100,
		AllowTables: []string{}, // 空表示允许所有（受黑名单限制）
		DenyTables: []string{ // 默认禁止查询的敏感表
			"staff", "staff_token", "company_setting",
			"payment_credential", "api_key",
		},
	}

	if host := viper.GetString("DB_HOST"); host != "" {
		Database.Host = host
	}
	if port := viper.GetInt("DB_PORT"); port > 0 {
		Database.Port = port
	}
	if user := viper.GetString("DB_USERNAME"); user != "" {
		Database.User = user
	}
	if password := viper.GetString("DB_PASSWORD"); password != "" {
		Database.Password = password
	}
	if database := viper.GetString("DB_DATABASE"); database != "" {
		Database.Database = database
	}
	if prefix := viper.GetString("DB_PREFIX"); prefix != "" {
		Database.TablePrefix = prefix
	}
	if maxRows := viper.GetInt("AI_SQL_MAX_ROWS"); maxRows > 0 {
		Database.MaxRows = maxRows
	}
}

// GetDSN 获取数据库连接字符串
func GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		Database.User,
		Database.Password,
		Database.Host,
		Database.Port,
		Database.Database,
	)
}
