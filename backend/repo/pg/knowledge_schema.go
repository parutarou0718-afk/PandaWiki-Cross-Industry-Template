package pg

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	storepg "github.com/chaitin/panda-wiki/store/pg"
)

type KnowledgeSchemaRepository struct {
	db     *storepg.DB
	logger *log.Logger
}

func NewKnowledgeSchemaRepository(db *storepg.DB, logger *log.Logger) *KnowledgeSchemaRepository {
	return &KnowledgeSchemaRepository{db: db, logger: logger.WithModule("repo.pg.knowledge_schema")}
}

func (r *KnowledgeSchemaRepository) GetEffectiveSchema(ctx context.Context, kbID string) (domain.KnowledgeSchema, error) {
	var record domain.KnowledgeGraphSchemaRecord
	err := r.db.WithContext(ctx).Where("kb_id = ?", kbID).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.DefaultKnowledgeSchema(), nil
	}
	if err != nil {
		return domain.KnowledgeSchema{}, err
	}
	if err := record.Schema.Validate(); err != nil {
		return domain.KnowledgeSchema{}, fmt.Errorf("stored knowledge schema is invalid: %w", err)
	}
	return record.Schema, nil
}

func (r *KnowledgeSchemaRepository) SaveSchema(ctx context.Context, kbID string, schema domain.KnowledgeSchema) error {
	if err := schema.Validate(); err != nil {
		return err
	}
	record := domain.KnowledgeGraphSchemaRecord{KBID: kbID, Schema: schema}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "kb_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"schema", "updated_at"}),
		}).Create(&record).Error
	})
}
