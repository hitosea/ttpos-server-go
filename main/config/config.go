package config

import (
	"fmt"

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

func Init() error {
	// 加载 .env 文件
	err := godotenv.Load("../.env")
	if err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}
	// 加载环境变量到 viper
	viper.AutomaticEnv()

	opt := copier.Option{IgnoreEmpty: true}

	serverConf(opt)   // 服务器
	databaseConf(opt) // 数据库
	redisConf(opt)    // Redis
	jwtConf(opt)      // JWT
	logConf(opt)      // 日志

	migrateDatabaseConf(opt) // 迁移数据库

	// 验证码
	Captcha = CaptchaConf{CachePrefix: "captcha:"}
	// 接口加密相关
	Encrypt = EncryptConf{
		CachePrefix:   "keypair:",
		EncryptHeader: "encrypt",
		ClientID:      "encrypt_id",
		ClientKey:     "client_key",
	}

	return nil
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
		Secret:        "your-secret-key-here",
		Expire:        3600 * 24,      // 默认24小时
		RefreshExpire: 3600 * 24 * 30, // 默认30天
	}
	copier.CopyWithOption(&JWT, JWTConf{
		Secret:        viper.GetString("JWT_SECRET"),
		Expire:        viper.GetInt("JWT_EXPIRE"),
		RefreshExpire: viper.GetInt("JWT_REFRESH_EXPIRE"),
	}, opt)
}

func redisConf(opt copier.Option) {
	Redis = RedisConf{
		Host:     "127.0.0.1",
		Port:     "6379",
		Password: "password",
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
		User:            "user",
		Password:        "password",
		RootPassword:    "root-password",
		Database:        "db",
		TablePrefix:     "ttpos_",
		SlowQueryTime:   2,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 5,
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
}

func serverConf(opt copier.Option) {
	Server = ServerConf{
		Port:       "8080",
		Mode:       "debug",
		DeployMode: "cloud",
		BrandName:  "TTPOS",
	}
	copier.CopyWithOption(&Server, ServerConf{
		Port:       viper.GetString("SERVER_PORT"),
		Mode:       viper.GetString("SERVER_MODE"),
		DeployMode: viper.GetString("DEPLOY_MODE"),
		BrandName:  viper.GetString("BRAND_NAME"),
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
