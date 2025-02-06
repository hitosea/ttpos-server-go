package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/config"
	"ttpos-server-go/migration/util"
	"ttpos-server-go/pkg/database"

	"github.com/golang-migrate/migrate/v4"
	miMysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/spf13/cobra"
)

var (
	version uint64
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().Uint64Var(&version, "version", 0, "migration version")
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run migration",
	Long:  `run migration`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if version == 0 {
			util.PrintError("Version is required!")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		runMigrate(version)
	},
	PostRun: func(cmd *cobra.Command, args []string) {
	},
}

func runMigrate(version uint64) {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	migrateList := []*migrate.Migrate{}
	databaseConf := config.Database
	fmt.Println(fmt.Sprintf("dbconf:%+v", databaseConf))
	var dbm *database.DBManager = database.GetDBManager(databaseConf)

	dbNameList := dbm.GetDBNameList()
	for shopID, dbName := range dbNameList {
		if shopID == constant.DefaultDB {
			continue
		}
		fmt.Println(fmt.Sprintf("shopID:%+v", shopID))
		db := dbm.GetDB(shopID)
		dbInstance, _ := db.DB()

		instance, _ := miMysql.WithInstance(dbInstance, &miMysql.Config{
			MigrationsTable:  "",
			DatabaseName:     dbName,
			NoLock:           false,
			StatementTimeout: 0,
		})
		m, err := migrate.NewWithDatabaseInstance(
			"file://migration/files",
			dbName,
			instance,
		)
		if err != nil {
			log.Fatal(err)
		}
		m.Log = NewLogger(true)
		migrateList = append(migrateList, m)
	}

	db, err := sql.Open("mysql", "root:QWERASDFQWE23421@tcp(localhost:3306)/shop1003?charset=utf8mb4&parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	instance, _ := miMysql.WithInstance(db, &miMysql.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://migration/migration",
		"shop1003",
		instance,
	)
	if err != nil {
		log.Fatal(err)
	}
	m.Log = NewLogger(true)
	migrateList = append(migrateList, m)

	for _, m := range migrateList {
		//m.Steps(-2)
		//m.Migrate(uint(version))
		err = m.Up() //or m.Down()
		if err != nil {
			log.Println(err)
		}
		_ = m.Steps(1) //执行的文件数
	}
}

type Logger struct {
	verbose bool
}

func (l *Logger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func (l *Logger) Verbose() bool {
	return l.verbose
}

func NewLogger(verbose bool) *Logger {
	return &Logger{
		verbose: verbose,
	}
}
