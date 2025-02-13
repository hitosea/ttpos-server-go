package command

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
	"ttpos-server-go/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	miMysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func init() {
	rootCommand.AddCommand(saasCmd)
	saasCmd.Flags().StringVar(&op, "op", "up", "migration operation")
}

var saasCmd = &cobra.Command{
	Use:   "migrate-saas",
	Short: "run saas migration",
	Long:  `run saas migration`,
	PreRun: func(cmd *cobra.Command, args []string) {

	},
	Run: func(cmd *cobra.Command, args []string) {
		runSaasMigrate(op)
	},
	PostRun: func(cmd *cobra.Command, args []string) {
	},
}

func NewMySQLConnection(conf config.DatabaseConf, dbName string) (*gorm.DB, error) {
	ignoreRecordNotFoundError := true // 忽略ErrRecordNotFound（记录未找到）错误
	logLevel := logger.Warn
	if config.Server.Mode == gin.DebugMode { // 调试模式
		ignoreRecordNotFoundError = false // 不忽略
		logLevel = logger.Info            // 详细日志
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?multiStatements=true&charset=utf8mb4&parseTime=True&loc=Local",
		"root",
		conf.RootPassword,
		conf.Host,
		conf.Port,
		dbName,
	)
	// 初始化会话
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   conf.TablePrefix, // 表名前缀
			SingularTable: true,             // 使用单一表名, eg. `User` => `user`
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用自动创建外键约束
		SkipDefaultTransaction:                   true, // 禁用默认事务
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的地方）
			logger.Config{
				SlowThreshold:             time.Duration(conf.SlowQueryTime) * time.Second, // 慢查询阈值
				LogLevel:                  logLevel,                                        // 日志级别
				IgnoreRecordNotFoundError: ignoreRecordNotFoundError,                       // 忽略ErrRecordNotFound（记录未找到）错误
				Colorful:                  true,                                            // 彩色打印
			},
		),
	})
}

func runSaasMigrate(op string) {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	databaseConf := config.Database
	fmt.Println(fmt.Sprintf("dbconf:%+v", databaseConf))

	dns := fmt.Sprintf("%s:%s@tcp(%s:%d)/", "root", databaseConf.RootPassword, databaseConf.Host, databaseConf.Port)
	fmt.Println(dns)
	db, err := sql.Open("mysql", dns)
	if err != nil {
		log.Fatal(err)
	}
	// 查询是否存在数据库
	err = createDatabaseIfNotExists(db, databaseConf.Database)
	if err != nil {
		log.Fatal(err)
	}
	saasDb, errSql := NewMySQLConnection(databaseConf, databaseConf.Database)
	if errSql != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	dbInstance, _ := saasDb.DB()
	instance, _ := miMysql.WithInstance(dbInstance, &miMysql.Config{
		MigrationsTable:  "",
		DatabaseName:     databaseConf.Database,
		NoLock:           false,
		StatementTimeout: 0,
	})
	m, err := migrate.NewWithDatabaseInstance(
		"file://migration/saas",
		databaseConf.Database,
		instance,
	)
	if err != nil {
		log.Fatal(err)
	}
	m.Log = NewLogger(true)

	if op == "up" {
		err = m.Up()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = m.Down()
		if err != nil {
			log.Fatal(err)
		}
	}

}

func createDatabaseIfNotExists(db *sql.DB, targetDbName string) error {
	rows, err := db.QueryContext(context.Background(), "SHOW DATABASES")
	if err != nil {
		return err
	}
	defer rows.Close()

	exists := false
	for rows.Next() {
		var dbName string
		err = rows.Scan(&dbName)
		if err != nil {
			return err
		}
		if dbName == targetDbName {
			exists = true
			break
		}
	}
	if !exists {
		_, err = db.ExecContext(context.Background(), "CREATE DATABASE "+targetDbName)
		if err != nil {
			return err
		}
	}
	return nil
}
