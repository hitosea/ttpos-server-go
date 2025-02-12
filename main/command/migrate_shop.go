package command

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"ttpos-server-go/config"

	"github.com/golang-migrate/migrate/v4"
	miMysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

var (
	op string
)

func init() {
	rootCommand.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVar(&op, "op", "up", "migration operation")
}

var migrateCmd = &cobra.Command{
	Use:   "migrate-shop",
	Short: "run migration",
	Long:  `run migration`,
	PreRun: func(cmd *cobra.Command, args []string) {

	},
	Run: func(cmd *cobra.Command, args []string) {
		if op == "up" {
			runMigrate(true)
		} else {
			runMigrate(false)
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
	},
}

func runMigrate(up bool) {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	migrateList := []*migrate.Migrate{}
	databaseConf := config.Database
	dbList := createDatabase("root", databaseConf.RootPassword, databaseConf.Host, strconv.Itoa(databaseConf.Port))
	for _, db := range dbList {
		instance, _ := miMysql.WithInstance(db, &miMysql.Config{})
		m, err := migrate.NewWithDatabaseInstance(
			"file://"+migrations,
			"",
			instance,
		)
		if err != nil {
			log.Fatal(err)
		}
		m.Log = NewLogger(true)
		migrateList = append(migrateList, m)
	}
	for _, m := range migrateList {
		if up {
			err := m.Up() //or m.Down()
			if err != nil {
				log.Println(err)
			}
		} else {
			err := m.Down() //or m.Down()
			if err != nil {
				log.Println(err)
			}
		}
	}
}

type Logger struct {
	verbose bool
}

func (l *Logger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func (l *Logger) Verbose() bool {
	va := []string{"", ""}
	l.Printf("verbose:%v", va)
	return l.verbose
}

func NewLogger(verbose bool) *Logger {
	return &Logger{
		verbose: verbose,
	}
}

func createDatabase(user string, password string, host string, port string) (dbList []*sql.DB) {
	// 连接到MySQL数据库
	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, password, host, port)
	fmt.Println(dns)
	db, err := sql.Open("mysql", dns)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 创建数据库的上下文
	ctx := context.Background()
	// 查询database列表
	rows, err := db.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var database string
		if err := rows.Scan(&database); err != nil {
			log.Fatal(err)
		}
		if matched, _ := regexp.MatchString(`^shop\d+$`, database); matched {
			databases = append(databases, database)
		}
	}

	// 打印数据库列表
	for _, database := range databases {
		log.Println(database)
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true&charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai", user, password, host, port, database))
		if err != nil {
			log.Fatal(err)
		}
		dbList = append(dbList, db)
	}
	return dbList
}
