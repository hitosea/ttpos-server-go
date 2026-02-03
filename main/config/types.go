package config

import "fmt"

type GoogleConf struct {
	PlacesApiKey string // 谷歌地图搜索API密钥
}

type ServerConf struct {
	Port           string // 端口
	Mode           string // 模式：debug/release/test
	DeployMode     string // 部署模式：cloud云上，offline离线
	BrandName      string // 品牌名称
	Domain         string // 域名
	MemberBaseUrl  string // 会员端域名
	PaymentTimeout int64  // 支付超时时间，单位秒，默认24小时
}

type DatabaseConf struct {
	DBType        string
	Host          string
	Port          int
	User          string
	Password      string
	RootPassword  string
	Database      string // 主数据库，如果是sqlite3，则为文件路径
	TablePrefix   string // 表名前缀
	SlowQueryTime int    // 慢查询阈值，单位秒

	MaxIdleConns    int // 最大空闲连接数
	MaxOpenConns    int // 最大打开的连接数
	ConnMaxLifetime int // 连接的最大可存活时间，单位分钟
}

type RedisConf struct {
	Host     string
	Port     string
	Password string
	DB       int // 数据库(0~15)
}

type JWTConf struct {
	Secret        string // 密钥
	Expire        int    // 有效期，单位秒
	RefreshExpire int    // refresh token 有效期，单位秒
}

type CaptchaConf struct {
	CachePrefix string // 缓存前缀
}

type EncryptConf struct {
	CachePrefix   string // 缓存前缀
	EncryptHeader string // 请求头x-encrypt
	ClientID      string // 请求头x-encrypt字段client_id
	ClientKey     string // 请求头x-encrypt字段client_key
	AesSecretKey  string // 服务端公钥
}

type LogConf struct {
	Dir           string // 日志路径
	Level         string // 日志级别：debug/info/warn/error
	MaxSize       int    // 每个日志文件保存的最大尺寸 单位：M
	MaxBackup     int    // 日志文件最大数量
	CleanSchedule string // 定时清理日志
}

type MigrateDatabaseConf struct {
	MigrateOldDBHost     string
	MigrateOldDBPort     int
	MigrateOldDBUser     string
	MigrateOldDBPassword string
	MigrateOldDBDatabase string // 主数据库，如果是sqlite3，则为文件路径
	MigrateOldDBPrefix   string // 表名前缀
}

type TakeoutConf struct {
	TakeoutTtposSecret    string // 导出密钥
	TakeoutLinemanStoreId uint64 // Lineman 门店ID
}

type SMSConf struct {
	BaseURL     string // 基础URL
	APIKey      string // API密钥
	ProjectName string // 项目名称
}

type GoogleBucketConf struct {
	GoogleApplicationCredentialsFileName  string // 谷歌云证书文件名
	GoogleApplicationBucketName           string // 谷歌云bucket - 安装包
	GoogleApplicationBucketEnv            string // 谷歌云bucket - 安装包 - 环境
	GoogleApplicationUploadsBucketName    string // 谷歌云 - 上传图片的bucket目录名称
	GoogleApplicationUploadsCatalogueName string // 谷歌云 - 上传图片的目录名称
	GooglePrintBucketName                 string // 谷歌云 - 打印相关的bucket
}

// NacosConf Nacos配置结构体
type NacosConf struct {
	Addresses string // 多实例配置（格式：host1:port1,host2:port2,host3:port3），优先使用
	Host      string // nacos服务地址（兼容旧配置）
	Port      int    // nacos端口（兼容旧配置）
	Namespace string // 命名空间
	Username  string // 用户名
	Password  string // 密码
	DataId    string // 配置DataId
	Group     string // 配置Group
}

// BmpConf BMP服务直连地址配置（调试模式使用）
type BmpConf struct {
	ErpGrpcAddr       string // ttpos-erp gRPC直连地址（如：localhost:14022）
	TakeoutGrpcAddr   string // ttpos-takeout gRPC直连地址
	MessageGrpcAddr   string // ttpos-message gRPC直连地址
	WebsocketGrpcAddr string // ttpos-websocket gRPC直连地址
}

func (c *GoogleBucketConf) Verification() bool {
	if c.GoogleApplicationCredentialsFileName == "" || c.GooglePrintBucketName == "" {
		return false
	}
	return true
}

func (c *GoogleBucketConf) GetGoogleApplicationCredentialsFileName() string {
	if c.Verification() {
		if Server.Mode == "debug" {
			return fmt.Sprintf("../docker/certificate/%s", c.GoogleApplicationCredentialsFileName)
		}
		return fmt.Sprintf("/var/certificate/%s", c.GoogleApplicationCredentialsFileName)
	}
	return ""
}
