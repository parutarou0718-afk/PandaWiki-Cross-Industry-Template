package pg

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/store/pg"
)

type SystemSettingRepo struct {
	db     *pg.DB
	logger *log.Logger
}

func (r *SystemSettingRepo) GetStoredEditionConfig(ctx context.Context) (*domain.StoredEditionConfig, error) {
	setting, err := r.GetSystemSetting(ctx, consts.SystemSettingEditionConfig)
	if err != nil {
		return nil, err
	}
	var stored domain.StoredEditionConfig
	if err := json.Unmarshal(setting.Value, &stored); err != nil {
		return nil, err
	}
	return &stored, nil
}

func (r *SystemSettingRepo) UpdateStoredEditionConfig(ctx context.Context, stored *domain.StoredEditionConfig) error {
	value, err := json.Marshal(stored)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&domain.SystemSetting{}).Where("key = ?", consts.SystemSettingEditionConfig).Update("value", value)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			return nil
		}
		return tx.Create(&domain.SystemSetting{Key: consts.SystemSettingEditionConfig, Value: value, Description: "Deployment edition configuration"}).Error
	})
}

func NewSystemSettingRepo(db *pg.DB, logger *log.Logger) *SystemSettingRepo {
	return &SystemSettingRepo{
		db:     db,
		logger: logger.WithModule("repo.pg.system_setting"),
	}
}

func (r *SystemSettingRepo) GetSystemSetting(ctx context.Context, key consts.SystemSettingKey) (*domain.SystemSetting, error) {
	var setting domain.SystemSetting
	result := r.db.WithContext(ctx).Where("key = ?", key).First(&setting)
	if result.Error != nil {
		return nil, result.Error
	}

	return &setting, nil
}

func (r *SystemSettingRepo) UpdateSystemSetting(ctx context.Context, key, value string) error {
	return r.db.WithContext(ctx).Model(&domain.SystemSetting{}).Where("key = ?", key).Update("value", value).Error
}
