package main

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/chaitin/panda-wiki/config"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/migration"
	"github.com/chaitin/panda-wiki/store/pg"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}
	db, err := pg.NewDB(cfg)
	if err != nil {
		panic(err)
	}
	manager := migration.NewDatabaseManager(db, log.NewLogger(cfg), func(tx *gorm.DB) error {
		for _, profile := range domain.BuiltinReportProfiles() {
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).Create(&profile).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err := manager.Execute(); err != nil {
		panic(fmt.Errorf("database-only migration failed: %w", err))
	}
}
