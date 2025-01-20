package config

import (
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var Server ServerConf
var Database DatabaseConf
var Redis RedisConf
var JWT JWTConf
var Captcha CaptchaConf
var Pgp PgpConf
var Log LogConf

func Init() error {
	// 加载 .env 文件
	err := godotenv.Load()
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

	// 验证码
	Captcha = CaptchaConf{CachePrefix: "captcha:"}
	// PGP
	Pgp = PgpConf{
		CachePrefix:     "keypair:",
		EncryptHeader:   "x-encrypt",
		ClientID:        "client_id",
		ClientPublicKey: "client_public_key",
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
		Secret: "your-secret-key-here",
		Expire: 3600 * 24,
	}
	copier.CopyWithOption(&JWT, JWTConf{
		Secret: viper.GetString("JWT_SECRET"),
		Expire: viper.GetInt("JWT_EXPIRE"),
	}, opt)
}

func redisConf(opt copier.Option) {
	Redis = RedisConf{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "password",
		DB:       0,
	}
	copier.CopyWithOption(&Redis, RedisConf{
		Host:     viper.GetString("REDIS_HOST"),
		Port:     viper.GetInt("REDIS_PORT"),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       viper.GetInt("REDIS_DB"),
	}, opt)
}

func databaseConf(opt copier.Option) {
	Database = DatabaseConf{
		Host:          "127.0.0.1",
		Port:          3306,
		User:          "user",
		Password:      "password",
		RootPassword:  "root-password",
		Database:      "db",
		TablePrefix:   "pre_",
		SlowQueryTime: 2,
	}
	copier.CopyWithOption(&Database, DatabaseConf{
		Host:          viper.GetString("DB_HOST"),
		Port:          viper.GetInt("DB_PORT"),
		User:          viper.GetString("DB_USER"),
		Password:      viper.GetString("DB_PASSWORD"),
		RootPassword:  viper.GetString("DB_ROOT_PASSWORD"),
		Database:      viper.GetString("DB_DATABASE"),
		TablePrefix:   viper.GetString("DB_TABLE_PREFIX"),
		SlowQueryTime: viper.GetInt("DB_SLOW_QUERY_TIME"),
	}, opt)
}

func serverConf(opt copier.Option) {
	Server = ServerConf{
		Port: "8080",
		Mode: "debug",
	}
	copier.CopyWithOption(&Server, ServerConf{
		Port: viper.GetString("SERVER_PORT"),
		Mode: viper.GetString("SERVER_MODE"),
	}, opt)
}
