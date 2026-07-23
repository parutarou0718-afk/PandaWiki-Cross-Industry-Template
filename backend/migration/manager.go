package migration

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/store/pg"
)

type GoMigrationFunc interface {
	Execute(tx *gorm.DB) error
}

// MigrationFunc represents a migration function
type MigrationFunc struct {
	Name string
	Fn   func(*gorm.DB) error
}

// Migration represents a migration record in database
type Migration struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"uniqueIndex"`
	ExecutedAt time.Time
}

// Manager handles database migrations
type Manager struct {
	db            *pg.DB
	logger        *log.Logger
	MigrationFunc *MigrationFuncs
}

// NewManager creates a new migration manager
func NewManager(db *pg.DB, logger *log.Logger, migrationFuncs *MigrationFuncs) (*Manager, error) {
	return &Manager{
		db:            db,
		logger:        logger.WithModule("migration"),
		MigrationFunc: migrationFuncs,
	}, nil
}

// Execute executes all pending migrations
func (m *Manager) Execute() error {
	// Execute pending migrations
	for _, migration := range m.MigrationFunc.GetMigrationFuncs() {
		m.logger.Info("find migration", log.String("name", migration.Name))
		err := m.db.Transaction(func(tx *gorm.DB) error {
			// Double check if migration was executed
			var record Migration
			if err := tx.Where("name = ?", migration.Name).First(&record).Error; err == nil {
				// Migration was executed by another instance
				m.logger.Info("skip migration", log.String("name", migration.Name))
				return nil
			}

			// Create migration record
			if err := tx.Create(&Migration{Name: migration.Name, ExecutedAt: time.Now()}).Error; err != nil {
				return fmt.Errorf("create migration record failed: %w", err)
			}

			m.logger.Info("starting migration", log.String("name", migration.Name))
			// Execute the migration
			if err := migration.Fn(tx); err != nil {
				return fmt.Errorf("execute migration %s failed: %w", migration.Name, err)
			}
			m.logger.Info("finished migration", log.String("name", migration.Name))

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
