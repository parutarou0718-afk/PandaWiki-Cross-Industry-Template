package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	storepg "github.com/chaitin/panda-wiki/store/pg"
)

var ErrPluginRecordNotFound = errors.New("plugin record not found")

// PluginRecordRepository persists plugin-owned records without knowing the
// plugin payload shape. Access filtering is deliberately performed in SQL so
// callers never receive records they must later hide in memory.
type PluginRecordRepository struct {
	db     *storepg.DB
	logger *log.Logger
}

func NewPluginRecordRepository(db *storepg.DB, logger *log.Logger) *PluginRecordRepository {
	return &PluginRecordRepository{db: db, logger: logger.WithModule("repo.pg.plugin_record")}
}

func buildVisiblePluginRecordQuery(db *gorm.DB, kbID, pluginID, recordType, authUserID string, authGroupIDs []int) *gorm.DB {
	query := db.Model(&domain.PluginRecord{}).
		Where("kb_id = ? AND plugin_id = ? AND record_type = ?", kbID, pluginID, recordType)
	if len(authGroupIDs) == 0 {
		return query.Where("owner_user_id = ? OR visibility = ?", authUserID, domain.PluginRecordVisibilityKnowledgeBase)
	}
	groupIDs := make([]int64, 0, len(authGroupIDs))
	for _, groupID := range authGroupIDs {
		groupIDs = append(groupIDs, int64(groupID))
	}
	return query.Where(`owner_user_id = ?
		OR visibility = ?
		OR (visibility = ? AND shared_auth_group_ids && ?::bigint[])`,
		authUserID,
		domain.PluginRecordVisibilityKnowledgeBase,
		domain.PluginRecordVisibilityGroups,
		pq.Array(groupIDs),
	)
}

func (r *PluginRecordRepository) ListVisible(ctx context.Context, kbID, pluginID, recordType, authUserID string, authGroupIDs []int) ([]domain.PluginRecord, error) {
	records := make([]domain.PluginRecord, 0)
	if err := buildVisiblePluginRecordQuery(r.db.WithContext(ctx), kbID, pluginID, recordType, authUserID, authGroupIDs).
		Order("updated_at DESC, id DESC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *PluginRecordRepository) GetVisibleByID(ctx context.Context, kbID, pluginID, recordType, recordID, authUserID string, authGroupIDs []int) (*domain.PluginRecord, error) {
	var record domain.PluginRecord
	if err := buildVisiblePluginRecordQuery(r.db.WithContext(ctx), kbID, pluginID, recordType, authUserID, authGroupIDs).
		Where("id = ?", recordID).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPluginRecordNotFound
		}
		return nil, err
	}
	return &record, nil
}

func (r *PluginRecordRepository) GetByID(ctx context.Context, kbID, pluginID, recordType, recordID string, includeDeleted bool) (*domain.PluginRecord, error) {
	var record domain.PluginRecord
	query := r.db.WithContext(ctx).Model(&domain.PluginRecord{})
	if includeDeleted {
		query = query.Unscoped()
	}
	if err := query.Where("id = ? AND kb_id = ? AND plugin_id = ? AND record_type = ?", recordID, kbID, pluginID, recordType).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPluginRecordNotFound
		}
		return nil, err
	}
	return &record, nil
}

func (r *PluginRecordRepository) Create(ctx context.Context, record *domain.PluginRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *PluginRecordRepository) Update(ctx context.Context, record *domain.PluginRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Model(&domain.PluginRecord{}).
		Where("id = ? AND kb_id = ? AND plugin_id = ? AND record_type = ?", record.ID, record.KBID, record.PluginID, record.RecordType).
		Updates(map[string]any{
			"payload":                  record.Payload,
			"visibility":               record.Access.Visibility,
			"shared_auth_group_ids":    record.Access.SharedAuthGroupIDs,
			"allow_collaborative_edit": record.Access.AllowCollaborativeEdit,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrPluginRecordNotFound
	}
	return nil
}

func (r *PluginRecordRepository) SoftDelete(ctx context.Context, kbID, pluginID, recordType, recordID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND kb_id = ? AND plugin_id = ? AND record_type = ?", recordID, kbID, pluginID, recordType).
		Delete(&domain.PluginRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrPluginRecordNotFound
	}
	return nil
}

func (r *PluginRecordRepository) Restore(ctx context.Context, kbID, pluginID, recordType, recordID string) error {
	result := r.db.WithContext(ctx).Unscoped().Model(&domain.PluginRecord{}).
		Where("id = ? AND kb_id = ? AND plugin_id = ? AND record_type = ? AND deleted_at IS NOT NULL", recordID, kbID, pluginID, recordType).
		Update("deleted_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("%w or not deleted", ErrPluginRecordNotFound)
	}
	return nil
}
