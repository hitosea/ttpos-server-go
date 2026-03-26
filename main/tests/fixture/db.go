// Package fixture provides test fixtures and helpers for integration tests.
package fixture

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const templateDBName = "shop_template"

var (
	templateOnce      sync.Once
	templateErr       error
	cachedTableList   []string
	cachedTableListMu sync.Mutex
)

const (
	defaultTestDBMaxOpenConns       = 4
	defaultTestDBMaxIdleConns       = 2
	defaultTestDBConnMaxLifetimeSec = 30
	defaultAdminDBMaxOpenConns      = 1
	defaultAdminDBMaxIdleConns      = 1
	defaultAdminDBConnLifetimeSec   = 15
)

// templateLockName is the MySQL advisory lock name used for cross-process synchronization.
const templateLockName = "ttpos_test_template_setup"

// ensureTemplateDB creates the shop_template database exactly once.
// Uses sync.Once for in-process deduplication and MySQL GET_LOCK for cross-process
// synchronization when parallel test packages run as separate binaries.
func ensureTemplateDB(tb testing.TB) {
	tb.Helper()
	templateOnce.Do(func() {
		rootConfig := RootDBConfig()

		rootDB, err := sql.Open("mysql", rootConfig.DSNWithoutDB())
		if err != nil {
			templateErr = fmt.Errorf("open root connection for template: %w", err)
			return
		}
		configureAdminDBPool(rootDB)
		defer rootDB.Close()

		// Acquire MySQL advisory lock — blocks until the lock holder (first process) finishes.
		var lockResult int
		if err := rootDB.QueryRow("SELECT GET_LOCK(?, 120)", templateLockName).Scan(&lockResult); err != nil || lockResult != 1 {
			templateErr = fmt.Errorf("failed to acquire MySQL advisory lock %q: err=%v result=%d", templateLockName, err, lockResult)
			return
		}
		defer rootDB.Exec("SELECT RELEASE_LOCK(?)", templateLockName)

		_, err = rootDB.Exec(fmt.Sprintf(
			"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
			templateDBName,
		))
		if err != nil {
			templateErr = fmt.Errorf("create template database: %w", err)
			return
		}

		// Check if tables already exist (another process already imported under the lock).
		var count int
		row := rootDB.QueryRow(
			"SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ?",
			templateDBName,
		)
		if err := row.Scan(&count); err != nil || count > 0 {
			// Template fully populated — just cache the table list.
			templateErr = cacheTableList(rootDB)
			return
		}

		// Import the full schema SQL into the template database.
		sqlPath := getEnv("SHOP_SQL_PATH", "../../../admin/database/seeds/shop_01.sql")
		sqlBytes, err := os.ReadFile(sqlPath)
		if err != nil {
			templateErr = fmt.Errorf("read shop SQL file %s: %w", sqlPath, err)
			return
		}

		multiDSN := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
			rootConfig.Username, rootConfig.Password, rootConfig.Host, rootConfig.Port, templateDBName,
		)
		schemaDB, err := sql.Open("mysql", multiDSN)
		if err != nil {
			templateErr = fmt.Errorf("open multiStatements connection for template: %w", err)
			return
		}
		configureAdminDBPool(schemaDB)
		sqlContent := strings.ReplaceAll(string(sqlBytes), "CREATE TABLE `", "CREATE TABLE IF NOT EXISTS `")
		if _, err = schemaDB.Exec(sqlContent); err != nil {
			schemaDB.Close()
			templateErr = fmt.Errorf("apply schema to template database: %w", err)
			return
		}
		schemaDB.Close()

		templateErr = cacheTableList(rootDB)
	})

	if templateErr != nil {
		tb.Fatalf("template database setup failed: %v", templateErr)
	}
}

// cacheTableList queries INFORMATION_SCHEMA for all base tables in the template DB
// and stores them for use by cloneFromTemplate.
func cacheTableList(rootDB *sql.DB) error {
	rows, err := rootDB.Query(
		"SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'",
		templateDBName,
	)
	if err != nil {
		return fmt.Errorf("query template table list: %w", err)
	}
	defer rows.Close()

	cachedTableListMu.Lock()
	defer cachedTableListMu.Unlock()
	cachedTableList = cachedTableList[:0]
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("scan table name: %w", err)
		}
		cachedTableList = append(cachedTableList, name)
	}
	return rows.Err()
}

// cloneFromTemplate creates targetDB and copies all table structures from shop_template
// using a single batched CREATE TABLE LIKE execution via multiStatements=true.
// Safe for parallel packages: ensureTemplateDB uses sync.Once within each process,
// and CREATE TABLE IF NOT EXISTS in the template import is idempotent across processes.
func cloneFromTemplate(tb testing.TB, rootDB *sql.DB, targetDB string) {
	tb.Helper()

	_, err := rootDB.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		targetDB,
	))
	if err != nil {
		tb.Fatalf("failed to create tenant database %s: %v", targetDB, err)
	}

	cachedTableListMu.Lock()
	tables := make([]string, len(cachedTableList))
	copy(tables, cachedTableList)
	cachedTableListMu.Unlock()

	if len(tables) == 0 {
		tb.Fatalf("template table list is empty — ensureTemplateDB may have failed")
	}

	// Build a single batched DDL string and execute in one round-trip.
	// MySQL executes all statements before sending any response, so all tables
	// are created even though database/sql only reads the first OK packet.
	var batch strings.Builder
	for _, table := range tables {
		fmt.Fprintf(&batch, "CREATE TABLE `%s`.`%s` LIKE `%s`.`%s`;\n",
			targetDB, table, templateDBName, table)
	}

	rootConfig := RootDBConfig()
	multiDSN := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		rootConfig.Username, rootConfig.Password, rootConfig.Host, rootConfig.Port,
	)
	multiDB, err := sql.Open("mysql", multiDSN)
	if err != nil {
		tb.Fatalf("failed to open multiStatements connection for clone: %v", err)
	}
	configureAdminDBPool(multiDB)
	defer multiDB.Close()

	if _, err = multiDB.Exec(batch.String()); err != nil {
		tb.Fatalf("failed to clone template tables into %s: %v", targetDB, err)
	}
}

// DBConfig holds database connection configuration.
type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

// DefaultDBConfig returns the default test database configuration from environment.
func DefaultDBConfig() DBConfig {
	return DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		Username: getEnv("DB_USERNAME", "test"),
		Password: getEnv("DB_PASSWORD", "test"),
		Database: getEnv("DB_DATABASE", "saas"),
	}
}

// RootDBConfig returns the root database configuration for creating databases.
// This uses MYSQL_ROOT_PASSWORD which has all privileges.
func RootDBConfig() DBConfig {
	return DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		Username: getEnv("DB_ROOT_USERNAME", "root"),
		Password: getEnv("DB_ROOT_PASSWORD", "testroot"),
		Database: "",
	}
}

// DSN returns the MySQL data source name for the configuration.
func (c DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

// DSNWithoutDB returns the DSN without specifying a database (for creating databases).
func (c DBConfig) DSNWithoutDB() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port)
}

// NewSaasDB returns a connection to the saas (default) database.
// Useful for tests that need to seed data in global tables (e.g., auto_receipt_rule, company).
func NewSaasDB(tb testing.TB) *sql.DB {
	tb.Helper()
	return NewDB(tb, DefaultDBConfig())
}

// NewDB creates a new database connection for testing.
func NewDB(tb testing.TB, config DBConfig) *sql.DB {
	tb.Helper()

	db, err := sql.Open("mysql", config.DSN())
	if err != nil {
		tb.Fatalf("failed to open database connection: %v", err)
	}
	configureTestDBPool(db)

	if err := db.Ping(); err != nil {
		tb.Fatalf("failed to ping database: %v", err)
	}

	tb.Cleanup(func() {
		db.Close()
	})

	return db
}

// NewTestTenant creates a new test tenant database with the given UUID.
// The database name will be "shop{uuid}".
// It returns a connection to the new database.
func NewTestTenant(tb testing.TB, tenantUUID string) *sql.DB {
	tb.Helper()

	config := DefaultDBConfig()
	rootConfig := RootDBConfig()

	// Connect as root without database to create it
	rootDB, err := sql.Open("mysql", rootConfig.DSNWithoutDB())
	if err != nil {
		tb.Fatalf("failed to open root database connection: %v", err)
	}
	configureAdminDBPool(rootDB)
	defer rootDB.Close()

	// Create the tenant database
	dbName := fmt.Sprintf("shop%s", tenantUUID)
	_, err = rootDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName))
	if err != nil {
		tb.Fatalf("failed to create tenant database %s: %v", dbName, err)
	}

	// Grant permissions to test user
	_, err = rootDB.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%%'", dbName, config.Username))
	if err != nil {
		tb.Fatalf("failed to grant privileges: %v", err)
	}

	// Also grant CREATE DATABASE permissions so test can create tenant databases
	_, err = rootDB.Exec(fmt.Sprintf("GRANT CREATE ON *.* TO '%s'@'%%'", config.Username))
	if err != nil {
		tb.Logf("warning: failed to grant CREATE privilege: %v", err)
	}

	// Connect to the tenant database
	tenantConfig := config
	tenantConfig.Database = dbName
	db := NewDB(tb, tenantConfig)

	// Create essential tables
	createEssentialTables(tb, db)

	// Cleanup: drop the database when test is done
	tb.Cleanup(func() {
		_ = db.Close()

		if !shouldDropTenantDBImmediately() {
			return
		}

		rootConfig := RootDBConfig()
		dropDB, err := sql.Open("mysql", rootConfig.DSNWithoutDB())
		if err != nil {
			tb.Logf("warning: failed to open connection for cleanup: %v", err)
			return
		}
		configureAdminDBPool(dropDB)
		defer dropDB.Close()

		_, err = dropDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
		if err != nil {
			tb.Logf("warning: failed to drop tenant database %s: %v", dbName, err)
		}
	})

	return db
}

// NewTestTenantFull creates a new test tenant database with the full production schema.
// On the first call it creates a shared shop_template database by importing shop_01.sql once.
// Subsequent calls clone the template via CREATE TABLE LIKE (one round-trip), making each
// call roughly 10-20× faster than re-importing the full SQL file.
func NewTestTenantFull(tb testing.TB, tenantUUID string) *sql.DB {
	tb.Helper()

	// Ensure the shared template DB exists (imports shop_01.sql exactly once per process).
	ensureTemplateDB(tb)

	config := DefaultDBConfig()
	rootConfig := RootDBConfig()

	rootDB, err := sql.Open("mysql", rootConfig.DSNWithoutDB())
	if err != nil {
		tb.Fatalf("failed to open root database connection: %v", err)
	}
	configureAdminDBPool(rootDB)
	defer rootDB.Close()

	dbName := fmt.Sprintf("shop%s", tenantUUID)

	// Clone table structures from the template (fast metadata-only operation).
	cloneFromTemplate(tb, rootDB, dbName)

	_, err = rootDB.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%%'", dbName, config.Username))
	if err != nil {
		tb.Fatalf("failed to grant privileges: %v", err)
	}
	_, _ = rootDB.Exec(fmt.Sprintf("GRANT CREATE ON *.* TO '%s'@'%%'", config.Username))

	tenantConfig := config
	tenantConfig.Database = dbName
	db := NewDB(tb, tenantConfig)

	// Cleanup: drop the test database after the test completes.
	// With MySQL data on tmpfs, DROP DATABASE (195 tables) takes ~0.3s instead of ~25s.
	tb.Cleanup(func() {
		_ = db.Close()

		if !shouldDropTenantDBImmediately() {
			return
		}

		dropConfig := RootDBConfig()
		dropDB, err := sql.Open("mysql", dropConfig.DSNWithoutDB())
		if err != nil {
			tb.Logf("warning: failed to open connection for cleanup: %v", err)
			return
		}
		configureAdminDBPool(dropDB)
		defer dropDB.Close()
		_, err = dropDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
		if err != nil {
			tb.Logf("warning: failed to drop tenant database %s: %v", dbName, err)
		}
	})

	return db
}

// createEssentialTables creates the essential tables required for tests.
func createEssentialTables(tb testing.TB, db *sql.DB) {
	tb.Helper()

	// Create ttpos_staff table (schema matches model/staff.go)
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_staff (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			uuid BIGINT UNSIGNED NOT NULL DEFAULT 0,
			company_uuid BIGINT UNSIGNED NOT NULL DEFAULT 0,
			username VARCHAR(255) NOT NULL DEFAULT '',
			password VARCHAR(255) NOT NULL DEFAULT '',
			permission_password VARCHAR(255) NOT NULL DEFAULT '',
			phone VARCHAR(20) DEFAULT '',
			password_change_count INT DEFAULT 0,
			password_change_time INT UNSIGNED NOT NULL DEFAULT 0,
			real_name VARCHAR(255) NOT NULL DEFAULT '',
			is_super INT NOT NULL DEFAULT 0,
			has_data_permission TINYINT NOT NULL DEFAULT 0,
			user_type INT NOT NULL DEFAULT 0,
			is_disable INT NOT NULL DEFAULT 0,
			bind_key VARCHAR(255) DEFAULT '',
			cashier_online INT NOT NULL DEFAULT 0,
			cashier_login_time INT UNSIGNED NOT NULL DEFAULT 0,
			duty_no VARCHAR(64) DEFAULT '',
			status TINYINT DEFAULT 1,
			create_time INT UNSIGNED DEFAULT 0,
			update_time INT UNSIGNED DEFAULT 0,
			delete_time INT UNSIGNED DEFAULT 0,
			UNIQUE KEY idx_uuid (uuid),
			KEY idx_company_uuid (company_uuid)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		tb.Fatalf("failed to create ttpos_staff table: %v", err)
	}

	// Create ttpos_desk table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_desk (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			uuid BIGINT UNSIGNED NOT NULL,
			company_uuid BIGINT UNSIGNED NOT NULL,
			name VARCHAR(100) NOT NULL,
			area_uuid BIGINT UNSIGNED,
			status TINYINT DEFAULT 1,
			create_time INT UNSIGNED DEFAULT 0,
			update_time INT UNSIGNED DEFAULT 0,
			delete_time INT UNSIGNED DEFAULT 0,
			UNIQUE KEY idx_uuid (uuid),
			KEY idx_company_uuid (company_uuid)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		tb.Fatalf("failed to create ttpos_desk table: %v", err)
	}

	// Create ttpos_device table (schema must match model/device.go)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_device (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			uuid BIGINT UNSIGNED NOT NULL,
			company_uuid BIGINT UNSIGNED NOT NULL DEFAULT 0,
			name VARCHAR(100) NOT NULL DEFAULT '',
			source VARCHAR(255) NOT NULL DEFAULT '',
			device_id VARCHAR(255) NOT NULL DEFAULT '',
			status TINYINT DEFAULT 1,
			finally_login_uuid BIGINT UNSIGNED DEFAULT 0,
			finally_login_time INT DEFAULT 0,
			version VARCHAR(50) NOT NULL DEFAULT '',
			create_time INT DEFAULT 0,
			update_time INT DEFAULT 0,
			delete_time INT DEFAULT 0,
			UNIQUE KEY idx_uuid (uuid),
			KEY idx_company_uuid (company_uuid),
			KEY idx_source_device_id (source, device_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		tb.Fatalf("failed to create ttpos_device table: %v", err)
	}

	// Create ttpos_company table (schema must match model/company.go)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_company (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			uuid BIGINT UNSIGNED NOT NULL DEFAULT 0,
			name VARCHAR(255) NOT NULL DEFAULT '',
			logo VARCHAR(255) NOT NULL DEFAULT '',
			expire_time INT DEFAULT 0,
			auth_day INT DEFAULT 0,
			status TINYINT DEFAULT 1,
			auth_start_time INT DEFAULT 0,
			old_company_id INT DEFAULT 0,
			is_enable_erp INT DEFAULT 0,
			last_sync_time INT DEFAULT 0,
			create_time INT DEFAULT 0,
			update_time INT DEFAULT 0,
			delete_time INT DEFAULT 0,
			UNIQUE KEY idx_uuid (uuid)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		tb.Fatalf("failed to create ttpos_company table: %v", err)
	}

	// Create ttpos_company_setting table (schema must match model/company.go CompanySetting)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_company_setting (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			uuid BIGINT UNSIGNED NOT NULL DEFAULT 0,
			company_uuid BIGINT UNSIGNED NOT NULL DEFAULT 0,
			real_name VARCHAR(50) NOT NULL DEFAULT '',
			link_name VARCHAR(50) NOT NULL DEFAULT '',
			link_phone VARCHAR(25) NOT NULL DEFAULT '',
			sale_stock INT NOT NULL DEFAULT 0,
			is_open_coupon INT NOT NULL DEFAULT 0,
			is_open_marketing INT NOT NULL DEFAULT 0,
			is_open_member INT NOT NULL DEFAULT 0,
			is_open_tablet INT NOT NULL DEFAULT 0,
			is_open_h5 INT NOT NULL DEFAULT 0,
			is_open_assistant INT NOT NULL DEFAULT 0,
			is_open_kitchen_kds INT NOT NULL DEFAULT 0,
			is_open_buffet INT NOT NULL DEFAULT 0,
			enable_sms INT NOT NULL DEFAULT 0,
			sms_quota INT NOT NULL DEFAULT 0,
			is_open_h5_order INT NOT NULL DEFAULT 0,
			is_open_local_print INT NOT NULL DEFAULT 0,
			is_open_advanced_ticket_print INT NOT NULL DEFAULT 0,
			cash_limit INT NOT NULL DEFAULT 0,
			kitchen_limit INT NOT NULL DEFAULT 0,
			tablet_limit INT NOT NULL DEFAULT 0,
			assistant_limit INT NOT NULL DEFAULT 0,
			table_limit INT NOT NULL DEFAULT 0,
			printer_limit INT NOT NULL DEFAULT 0,
			timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Shanghai',
			languages VARCHAR(255) NOT NULL DEFAULT '',
			address VARCHAR(255) NOT NULL DEFAULT '',
			coordinates VARCHAR(255) NOT NULL DEFAULT '',
			delivery_config TEXT,
			delivery_status INT NOT NULL DEFAULT 0,
			erpnext_site_code VARCHAR(255) NOT NULL DEFAULT '',
			erpnext_company_abbr VARCHAR(255) NOT NULL DEFAULT '',
			erpnext_headquarter_abbr VARCHAR(255) NOT NULL DEFAULT '',
			headquarter_uuid BIGINT NOT NULL DEFAULT 0,
			erpnext_branch_name VARCHAR(255) NOT NULL DEFAULT '',
			erpnext_pos_profile_name VARCHAR(255) NOT NULL DEFAULT '',
			erpnext_admin_email VARCHAR(255) NOT NULL DEFAULT '',
			parent_company_uuids VARCHAR(255) NOT NULL DEFAULT '',
			has_children INT NOT NULL DEFAULT 0,
			enable_table_map INT NOT NULL DEFAULT 0,
			enable_data_management INT NOT NULL DEFAULT 0,
			enable_kiosk INT NOT NULL DEFAULT 0,
			enable_grab_delivery INT NOT NULL DEFAULT 0,
			enable_lineman_delivery INT NOT NULL DEFAULT 0,
			create_time INT DEFAULT 0,
			update_time INT DEFAULT 0,
			delete_time INT DEFAULT 0,
			UNIQUE KEY idx_uuid (uuid),
			KEY idx_company_uuid (company_uuid)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		tb.Fatalf("failed to create ttpos_company_setting table: %v", err)
	}

	// Create ttpos_product table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ttpos_product (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			uuid BIGINT UNSIGNED NOT NULL,
			company_uuid BIGINT UNSIGNED NOT NULL,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(10,2) DEFAULT 0.00,
			category_uuid BIGINT UNSIGNED,
			status TINYINT DEFAULT 1,
			create_time INT UNSIGNED DEFAULT 0,
			update_time INT UNSIGNED DEFAULT 0,
			delete_time INT UNSIGNED DEFAULT 0,
			UNIQUE KEY idx_uuid (uuid),
			KEY idx_company_uuid (company_uuid)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		tb.Fatalf("failed to create ttpos_product table: %v", err)
	}
}

// TruncateTable truncates a table in the database.
func TruncateTable(tb testing.TB, db *sql.DB, tableName string) {
	tb.Helper()
	_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE `%s`", tableName))
	if err != nil {
		tb.Fatalf("failed to truncate table %s: %v", tableName, err)
	}
}

func shouldDropTenantDBImmediately() bool {
	return getEnv("TEST_DROP_TENANT_DB", "1") != "0"
}

func configureTestDBPool(db *sql.DB) {
	db.SetMaxOpenConns(getEnvInt("TEST_DB_MAX_OPEN_CONNS", defaultTestDBMaxOpenConns))
	db.SetMaxIdleConns(getEnvInt("TEST_DB_MAX_IDLE_CONNS", defaultTestDBMaxIdleConns))
	db.SetConnMaxLifetime(time.Duration(getEnvInt("TEST_DB_CONN_MAX_LIFETIME_SEC", defaultTestDBConnMaxLifetimeSec)) * time.Second)
}

func configureAdminDBPool(db *sql.DB) {
	db.SetMaxOpenConns(getEnvInt("TEST_DB_ADMIN_MAX_OPEN_CONNS", defaultAdminDBMaxOpenConns))
	db.SetMaxIdleConns(getEnvInt("TEST_DB_ADMIN_MAX_IDLE_CONNS", defaultAdminDBMaxIdleConns))
	db.SetConnMaxLifetime(time.Duration(getEnvInt("TEST_DB_ADMIN_CONN_MAX_LIFETIME_SEC", defaultAdminDBConnLifetimeSec)) * time.Second)
}
