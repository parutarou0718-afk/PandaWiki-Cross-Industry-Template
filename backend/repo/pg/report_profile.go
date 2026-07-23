package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/chaitin/panda-wiki/domain"
	store "github.com/chaitin/panda-wiki/store/pg"
)

type ReportProfileRepo struct{ db *store.DB }

func NewReportProfileRepo(db *store.DB) *ReportProfileRepo { return &ReportProfileRepo{db: db} }
func (r *ReportProfileRepo) List(ctx context.Context) ([]domain.ReportProfile, error) {
	var p []domain.ReportProfile
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Find(&p).Error
	return p, err
}
func (r *ReportProfileRepo) ListEnabledForEdition(ctx context.Context, editionID domain.EditionID) ([]domain.ReportProfile, error) {
	var profiles []domain.ReportProfile
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL AND enabled = true AND edition_ids @> ?::jsonb", `[`+string(editionID)+`]`).Order("name ASC").Find(&profiles).Error
	return profiles, err
}
func (r *ReportProfileRepo) GetByID(ctx context.Context, id string) (*domain.ReportProfile, error) {
	var p domain.ReportProfile
	err := r.db.WithContext(ctx).Where("id=? AND deleted_at IS NULL", id).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}
func (r *ReportProfileRepo) Create(ctx context.Context, p *domain.ReportProfile) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	if err := p.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(p).Error
}
func (r *ReportProfileRepo) Update(ctx context.Context, p *domain.ReportProfile) error {
	if err := p.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var old domain.ReportProfile
		if err := tx.Where("id=? AND deleted_at IS NULL", p.ID).First(&old).Error; err != nil {
			return err
		}
		if old.IsBuiltin != p.IsBuiltin {
			return errors.New("builtin identity is immutable")
		}
		return tx.Model(&old).Select("name", "description", "edition_ids", "system_prompt", "input_fields", "sections", "citation_required", "enabled", "version", "updated_at").Updates(p).Error
	})
}
func (r *ReportProfileRepo) SetEnabled(ctx context.Context, id string, enabled bool) error {
	result := r.db.WithContext(ctx).Model(&domain.ReportProfile{}).Where("id=? AND deleted_at IS NULL", id).Update("enabled", enabled)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (r *ReportProfileRepo) SoftDelete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.ReportProfile{}).Where("id=? AND is_builtin=false AND deleted_at IS NULL", id).Update("deleted_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("report profile cannot be deleted")
	}
	return nil
}
func (r *ReportProfileRepo) RestoreBuiltinDefault(ctx context.Context, id string) error {
	var preset *domain.ReportProfile
	for _, p := range domain.BuiltinReportProfiles() {
		if p.ID == id {
			pCopy := p
			preset = &pCopy
			break
		}
	}
	if preset == nil {
		return errors.New("not a builtin report profile")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var old domain.ReportProfile
		if err := tx.Where("id=? AND is_builtin=true", id).First(&old).Error; err != nil {
			return err
		}
		// Restore template content without unexpectedly re-enabling a template
		// the administrator intentionally disabled.
		return tx.Model(&old).Select("name", "description", "edition_ids", "system_prompt", "input_fields", "sections", "citation_required", "version", "deleted_at", "updated_at").Updates(preset).Error
	})
}
func (r *ReportProfileRepo) EnsureBuiltinProfiles(ctx context.Context) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, p := range domain.BuiltinReportProfiles() {
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).Create(&p).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
