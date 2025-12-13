package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"ttpos-server-go/pkg/otlp"
	"ttpos-server-go/pkg/rocketmq"

	"github.com/jinzhu/copier"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var Server ServerConf
var Database DatabaseConf
var MigrateDatabase MigrateDatabaseConf
var Redis RedisConf
var JWT JWTConf
var Captcha CaptchaConf
var Encrypt EncryptConf
var Log LogConf
var SMS SMSConf
var GoogleBucket GoogleBucketConf
var Google GoogleConf
var Nacos NacosConf
var CORS CORSConf

var Rocketmq rocketmq.Config
var Otlp *otlp.OtlpConfig

func Init() error {
	// 加载 .env 文件
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Printf("error loading .env file: %v\n", err)
		return fmt.Errorf("error loading .env file: %v", err)
	}
	// 加载环境变量到 viper
	viper.AutomaticEnv()

	opt := copier.Option{IgnoreEmpty: true}

	serverConf(opt)          // 服务器
	databaseConf(opt)        // 数据库
	redisConf(opt)           // Redis
	jwtConf(opt)             // JWT
	logConf(opt)             // 日志
	smsConf(opt)             // 短信
	googleBucketConf(opt)    // 谷歌云
	googleConf(opt)          // 谷歌
	migrateDatabaseConf(opt) // 迁移数据库
	nacosConf(opt)           // nacos配置
	rocketmqConf(opt)        // rocketmq配置

	// 验证码
	Captcha = CaptchaConf{CachePrefix: "captcha:"}
	// 接口加密相关
	Encrypt = EncryptConf{
		CachePrefix:   "keypair:",
		EncryptHeader: "encrypt",
		ClientID:      "encrypt_id",
		ClientKey:     "client_key",
		AesSecretKey:  "",
	}

	Otlp = otlp.LoadOtlpConfig(opt)
	return nil
}

func rocketmqConf(opt copier.Option) {
	Rocketmq = rocketmq.Config{
		NameServers: nil,
		AccessKey:   "",
		SecretKey:   "",
		GroupName:   "ttpos-go-group",
		Retry:       3,
		LogLevel:    "warn",
		Enabled:     false,
	}
	copier.CopyWithOption(&Rocketmq, rocketmq.Config{
		NameServers: strings.Split(viper.GetString("ROCKETMQ_NAME_SRV_ADDR"), ","),
		AccessKey:   viper.GetString("ROCKETMQ_ACCESS_KEY"),
		SecretKey:   viper.GetString("ROCKETMQ_SECRET_KEY"),
		GroupName:   viper.GetString("ROCKETMQ_GROUP_NAME"),
		Retry:       viper.GetInt("ROCKETMQ_RETRY"),
		LogLevel:    viper.GetString("ROCKETMQ_LOG_LEVEL"),
		Enabled:     viper.GetBool("ROCKETMQ_ENABLED"),
	}, opt)
}

func nacosConf(opt copier.Option) {
	Nacos = NacosConf{
		Addresses: "",
		Host:      "localhost",
		Port:      8848,
		Namespace: "",
		Username:  "",
		Password:  "",
		DataId:    "",
		Group:     "",
	}
	copier.CopyWithOption(&Nacos, NacosConf{
		Addresses: viper.GetString("NACOS_SERVER_ADDRESSES"), // 多实例配置（优先）
		Host:      viper.GetString("NACOS_SERVER_IP"),        // 兼容旧配置
		Port:      viper.GetInt("NACOS_SERVER_PORT"),         // 兼容旧配置
		Namespace: viper.GetString("NACOS_NAMESPACE"),
		Username:  viper.GetString("NACOS_USERNAME"),
		Password:  viper.GetString("NACOS_PASSWORD"),
		DataId:    viper.GetString("NACOS_DATAID"),
		Group:     viper.GetString("NACOS_GROUP"),
	}, opt)
}

func logConf(opt copier.Option) {
	Log = LogConf{
		Dir:           "log",
		Level:         "debug",
		MaxSize:       500,
		MaxBackup:     14,
		CleanSchedule: "0 0 * * *",
	}
	copier.CopyWithOption(&Log, LogConf{
		Dir:           viper.GetString("LOG_DIR"),
		Level:         viper.GetString("LOG_LEVEL"),
		MaxSize:       viper.GetInt("LOG_MAX_SIZE"),
		MaxBackup:     viper.GetInt("LOG_MAX_BACKUP"),
		CleanSchedule: viper.GetString("LOG_CLEAN_SCHEDULE"),
	}, opt)
}

func jwtConf(opt copier.Option) {
	JWT = JWTConf{
		Secret:        "",
		Expire:        3600 * 24,      // 默认24小时
		RefreshExpire: 3600 * 24 * 30, // 默认30天
	}
	copier.CopyWithOption(&JWT, JWTConf{
		Secret:        viper.GetString("JWT_SECRET"),
		Expire:        viper.GetInt("JWT_EXPIRE"),
		RefreshExpire: viper.GetInt("JWT_REFRESH_EXPIRE"),
	}, opt)

	// 安全检查：JWT Secret 必须设置且不能为空
	if JWT.Secret == "" {
		log.Fatal("错误: 必须设置 JWT_SECRET 环境变量，且不能为为空值")
	}

	// 额外检查：防止使用明显的弱密钥
	if JWT.Secret == "your-secret-key-here" || JWT.Secret == "secret" {
		log.Fatal("错误: JWT_SECRET 不能使用弱密码，建议使用至少32位随机字符串")
	}
}

func redisConf(opt copier.Option) {
	Redis = RedisConf{
		Host:     "127.0.0.1",
		Port:     "6379",
		Password: "",
		DB:       0,
	}
	copier.CopyWithOption(&Redis, RedisConf{
		Host:     viper.GetString("REDIS_HOST"),
		Port:     viper.GetString("REDIS_PORT"),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       viper.GetInt("REDIS_DB"),
	}, opt)
}

func databaseConf(opt copier.Option) {
	Database = DatabaseConf{
		DBType:          "mysql",
		Host:            "127.0.0.1",
		Port:            3306,
		User:            "",
		Password:        "",
		RootPassword:    "",
		Database:        "db",
		TablePrefix:     "ttpos_",
		SlowQueryTime:   2,
		MaxIdleConns:    20,
		MaxOpenConns:    200,
		ConnMaxLifetime: 300,
	}

	copier.CopyWithOption(&Database, DatabaseConf{
		DBType:          viper.GetString("DB_TYPE"),
		Host:            viper.GetString("DB_HOST"),
		Port:            viper.GetInt("DB_PORT"),
		User:            viper.GetString("DB_USERNAME"),
		Password:        viper.GetString("DB_PASSWORD"),
		RootPassword:    viper.GetString("DB_ROOT_PASSWORD"),
		Database:        viper.GetString("DB_DATABASE"),
		TablePrefix:     viper.GetString("DB_PREFIX"),
		SlowQueryTime:   viper.GetInt("DB_SLOW_QUERY_TIME"),
		MaxIdleConns:    viper.GetInt("MAX_IDLE_CONNS"),
		MaxOpenConns:    viper.GetInt("MAX_OPEN_CONNS"),
		ConnMaxLifetime: viper.GetInt("CONN_MAX_LIFE_TIME"),
	}, opt)

	// 在生产环境中强制要求设置数据库密码
	if Server.Mode != "debug" {
		if Database.Password == "" {
			fmt.Printf("错误: 生产环境必须设置 DB_PASSWORD 环境变量\n")
			// 注意: 这里不返回错误，因为配置函数不应该失败
			// 而是在运行时检查
		}
	}
}

func serverConf(opt copier.Option) {
	Server = ServerConf{
		Port:           "8080",
		Mode:           "release",
		DeployMode:     "cloud",
		BrandName:      "TTPOS",
		Domain:         "http://127.0.0.1:8080",
		PaymentTimeout: 24 * 60 * 60, // 24小时
	}
	//
	serverPort := viper.GetString("SERVER_PORT")
	if debugServerPort := viper.GetString("DEBUG_SERVER_PORT"); debugServerPort != "" &&
		viper.GetString("SERVER_MODE") == "debug" {
		if execPath, err := os.Executable(); err == nil {
			execName := strings.ToLower(execPath)
			if strings.Contains(execName, "__debug_bin") || strings.Contains(execName, "dlv") {
				serverPort = debugServerPort
			}
		} else if _, err := os.Stat("/.dockerenv"); err != nil {
			serverPort = debugServerPort
		}
	}
	//
	copier.CopyWithOption(&Server, ServerConf{
		Port:           serverPort,
		Mode:           viper.GetString("SERVER_MODE"),
		DeployMode:     viper.GetString("DEPLOY_MODE"),
		BrandName:      viper.GetString("BRAND_NAME"),
		Domain:         viper.GetString("DOMAIN"),
		MemberBaseUrl:  viper.GetString("MEMBER_BASE_URL"),
		PaymentTimeout: viper.GetInt64("PAYMENT_TIMEOUT"),
	}, opt)
}

func migrateDatabaseConf(opt copier.Option) {
	MigrateDatabase = MigrateDatabaseConf{
		MigrateOldDBHost:     "",
		MigrateOldDBPort:     0,
		MigrateOldDBUser:     "",
		MigrateOldDBPassword: "",
		MigrateOldDBDatabase: "",
		MigrateOldDBPrefix:   "",
	}
	copier.CopyWithOption(&MigrateDatabase, MigrateDatabaseConf{
		MigrateOldDBHost:     viper.GetString("MIGRATE_OLD_DB_HOST"),
		MigrateOldDBPort:     viper.GetInt("MIGRATE_OLD_DB_PORT"),
		MigrateOldDBUser:     viper.GetString("MIGRATE_OLD_DB_USERNAME"),
		MigrateOldDBPassword: viper.GetString("MIGRATE_OLD_DB_PASSWORD"),
		MigrateOldDBDatabase: viper.GetString("MIGRATE_OLD_DB_DATABASE"),
		MigrateOldDBPrefix:   viper.GetString("MIGRATE_OLD_DB_PREFIX"),
	}, opt)
}

func smsConf(opt copier.Option) {
	SMS = SMSConf{
		BaseURL:     "",
		APIKey:      "",
		ProjectName: "",
	}
	copier.CopyWithOption(&SMS, SMSConf{
		BaseURL:     viper.GetString("SMS_BASE_URL"),
		APIKey:      viper.GetString("SMS_API_KEY"),
		ProjectName: viper.GetString("SMS_PROJECT_NAME"),
	}, opt)
}

func googleBucketConf(opt copier.Option) {
	GoogleBucket = GoogleBucketConf{
		GoogleApplicationCredentialsFileName:  "",
		GoogleApplicationBucketName:           "",
		GoogleApplicationBucketEnv:            "",
		GoogleApplicationUploadsBucketName:    "",
		GoogleApplicationUploadsCatalogueName: "",
		GooglePrintBucketName:                 "",
	}
	copier.CopyWithOption(&GoogleBucket, GoogleBucketConf{
		GoogleApplicationCredentialsFileName:  viper.GetString("GOOGLE_APPLICATION_CREDENTIALS_FILE_NAME"),
		GoogleApplicationBucketName:           viper.GetString("GOOGLE_APPLICATION_BUCKET_NAME"),
		GoogleApplicationBucketEnv:            viper.GetString("GOOGLE_APPLICATION_BUCKET_ENV"),
		GoogleApplicationUploadsBucketName:    viper.GetString("GOOGLE_APPLICATION_UPLOADS_BUCKET_NAME"),
		GoogleApplicationUploadsCatalogueName: viper.GetString("GOOGLE_APPLICATION_UPLOADS_CATALOGUE_NAME"),
		GooglePrintBucketName:                 viper.GetString("GOOGLE_PRINT_BUCKET_NAME"),
	}, opt)
}

func googleConf(opt copier.Option) {
	Google = GoogleConf{
		PlacesApiKey: "",
	}
	copier.CopyWithOption(&Google, GoogleConf{
		PlacesApiKey: viper.GetString("GOOGLE_PLACES_API_KEY"),
	}, opt)
}
