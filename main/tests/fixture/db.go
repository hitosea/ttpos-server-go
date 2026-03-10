// Package fixture provides test fixtures and helpers for integration tests.
package fixture

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

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

// NewDB creates a new database connection for testing.
func NewDB(tb testing.TB, config DBConfig) *sql.DB {
	tb.Helper()

	db, err := sql.Open("mysql", config.DSN())
	if err != nil {
		tb.Fatalf("failed to open database connection: %v", err)
	}

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
		rootConfig := RootDBConfig()
		dropDB, err := sql.Open("mysql", rootConfig.DSNWithoutDB())
		if err != nil {
			tb.Logf("warning: failed to open connection for cleanup: %v", err)
			return
		}
		defer dropDB.Close()

		_, err = dropDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
		if err != nil {
			tb.Logf("warning: failed to drop tenant database %s: %v", dbName, err)
		}
	})

	return db
}

// NewTestTenantFull creates a new test tenant database with the full production schema.
// It reads shop_01.sql from SHOP_SQL_PATH env var (default: ../../admin/database/seeds/shop_01.sql)
// and executes it to create all tenant tables. This is required for tests that exercise
// business logic beyond auth (e.g. order creation).
func NewTestTenantFull(tb testing.TB, tenantUUID string) *sql.DB {
	tb.Helper()

	config := DefaultDBConfig()
	rootConfig := RootDBConfig()

	// Connect as root to create the database
	rootDB, err := sql.Open("mysql", rootConfig.DSNWithoutDB())
	if err != nil {
		tb.Fatalf("failed to open root database connection: %v", err)
	}
	defer rootDB.Close()

	dbName := fmt.Sprintf("shop%s", tenantUUID)
	_, err = rootDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName))
	if err != nil {
		tb.Fatalf("failed to create tenant database %s: %v", dbName, err)
	}

	_, err = rootDB.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%%'", dbName, config.Username))
	if err != nil {
		tb.Fatalf("failed to grant privileges: %v", err)
	}
	_, _ = rootDB.Exec(fmt.Sprintf("GRANT CREATE ON *.* TO '%s'@'%%'", config.Username))

	// Use multiStatements=true DSN to execute the full schema SQL file
	// Default path is relative from the test package directory (e.g. main/tests/order/).
	// In Docker, SHOP_SQL_PATH env var overrides this to the absolute container path.
	sqlPath := getEnv("SHOP_SQL_PATH", "../../../admin/database/seeds/shop_01.sql")
	sqlBytes, err := os.ReadFile(sqlPath)
	if err != nil {
		tb.Fatalf("failed to read shop SQL file %s: %v", sqlPath, err)
	}

	multiDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		rootConfig.Username, rootConfig.Password, rootConfig.Host, rootConfig.Port, dbName)
	schemaDB, err := sql.Open("mysql", multiDSN)
	if err != nil {
		tb.Fatalf("failed to open schema connection: %v", err)
	}
	if _, err := schemaDB.Exec(string(sqlBytes)); err != nil {
		schemaDB.Close()
		tb.Fatalf("failed to apply shop schema to %s: %v", dbName, err)
	}
	schemaDB.Close()

	// Return a regular single-statement connection
	tenantConfig := config
	tenantConfig.Database = dbName
	db := NewDB(tb, tenantConfig)

	// Cleanup: drop the database when test is done
	tb.Cleanup(func() {
		dropConfig := RootDBConfig()
		dropDB, err := sql.Open("mysql", dropConfig.DSNWithoutDB())
		if err != nil {
			tb.Logf("warning: failed to open connection for cleanup: %v", err)
			return
		}
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
