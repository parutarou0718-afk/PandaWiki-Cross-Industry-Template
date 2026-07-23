package migration

import (
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/store/pg"
	"gorm.io/gorm"
	"time"
)

type DatabaseMigration struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"uniqueIndex"`
	ExecutedAt time.Time
}
type DatabaseManager struct {
	db     *pg.DB
	logger *log.Logger
	Seed   func(*gorm.DB) error
}

func NewDatabaseManager(db *pg.DB, logger *log.Logger, seed func(*gorm.DB) error) *DatabaseManager {
	return &DatabaseManager{db: db, logger: logger, Seed: seed}
}
func (m *DatabaseManager) Execute() error {
	if err := m.db.AutoMigrate(&DatabaseMigration{}); err != nil {
		return err
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		var r DatabaseMigration
		if err := tx.Where("name=?", "0006_seed_report_profiles").First(&r).Error; err == nil {
			return nil
		}
		if err := m.Seed(tx); err != nil {
			return err
		}
		return tx.Create(&DatabaseMigration{Name: "0006_seed_report_profiles", ExecutedAt: time.Now()}).Error
	})
}
