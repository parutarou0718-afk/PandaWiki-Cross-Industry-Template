package pg

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratePG "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/chaitin/panda-wiki/config"
)

//go:embed migration/*.sql
var migrationFS embed.FS

type DB struct {
	*gorm.DB
}

func NewDB(config *config.Config) (*DB, error) {
	dsn := config.PG.DSN
	// same as gorm logger.Default, but without colorful output and ignore record not found error
	newLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         newLogger,
	})
	if err != nil {
		return nil, err
	}
	// create raglite database if not exists
	var exists bool
	if err := db.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'raglite')").Scan(&exists).Error; err != nil {
		return nil, err
	}
	if !exists {
		if err := db.Exec("CREATE DATABASE raglite").Error; err != nil {
			return nil, err
		}
	}
	if err := doMigrate(dsn); err != nil {
		return nil, err
	}

	return &DB{DB: db}, nil
}

func doMigrate(dsn string) error {
	path := os.Getenv("MIGRATION_PATH")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}
	driver, err := migratePG.WithInstance(db, &migratePG.Config{})
	if err != nil {
		return fmt.Errorf("with instance failed: %w", err)
	}
	var m *migrate.Migrate
	if path == "" {
		source, sourceErr := iofs.New(migrationFS, "migration")
		if sourceErr != nil {
			return sourceErr
		}
		m, err = migrate.NewWithInstance("iofs", source, "postgres", driver)
	} else {
		resolved, resolveErr := resolveMigrationPath(path)
		if resolveErr != nil {
			return resolveErr
		}
		m, err = migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(resolved), "postgres", driver)
	}
	if err != nil {
		return fmt.Errorf("new with database instance failed: %w", err)
	}
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return fmt.Errorf("migrate db failed: %w", err)
	}

	return nil
}

func resolveMigrationPath(explicit string) (string, error) {
	candidates := []string{"store/pg/migration", "backend/store/pg/migration", "migration"}
	if explicit != "" {
		candidates = []string{explicit}
	}
	for _, candidate := range candidates {
		path, err := filepath.Abs(filepath.Clean(candidate))
		if err != nil {
			continue
		}
		matches, _ := filepath.Glob(filepath.Join(path, "*.up.sql"))
		if len(matches) > 0 {
			return path, nil
		}
	}
	return "", fmt.Errorf("migration path not found; tried %v", candidates)
}
