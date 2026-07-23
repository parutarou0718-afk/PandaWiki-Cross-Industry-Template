package fns

import (
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MigrationSeedReportProfiles struct {
	Name   string
	logger *log.Logger
}

func NewMigrationSeedReportProfiles(logger *log.Logger) *MigrationSeedReportProfiles {
	return &MigrationSeedReportProfiles{Name: "0006_seed_report_profiles", logger: logger}
}
func (m *MigrationSeedReportProfiles) Execute(tx *gorm.DB) error {
	for _, profile := range domain.BuiltinReportProfiles() {
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).Create(&profile).Error; err != nil {
			return err
		}
	}
	return nil
}
