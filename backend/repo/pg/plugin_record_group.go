package pg

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	storepg "github.com/chaitin/panda-wiki/store/pg"
)

var ErrPluginRecordGroupNotFound = errors.New("plugin record group not found")

type PluginRecordGroupRepository struct {
	db     *storepg.DB
	logger *log.Logger
}

func NewPluginRecordGroupRepository(db *storepg.DB, logger *log.Logger) *PluginRecordGroupRepository {
	return &PluginRecordGroupRepository{db: db, logger: logger.WithModule("repo.pg.plugin_record_group")}
}

func (r *PluginRecordGroupRepository) ListForUser(ctx context.Context, kbID, userID string, controller bool) ([]domain.PluginRecordGroup, error) {
	groups := make([]domain.PluginRecordGroup, 0)
	query := r.db.WithContext(ctx).Model(&domain.PluginRecordGroup{}).Where("kb_id = ?", kbID)
	if !controller {
		query = query.Joins("JOIN knowledge_base_plugin_group_members members ON members.group_id = knowledge_base_plugin_groups.id").Where("members.user_id = ?", userID)
	}
	if err := query.Order("name ASC, id ASC").Find(&groups).Error; err != nil {
		return nil, err
	}
	for index := range groups {
		if err := r.loadMembers(ctx, &groups[index]); err != nil {
			return nil, err
		}
	}
	return groups, nil
}

func (r *PluginRecordGroupRepository) ListIDsForUser(ctx context.Context, userID string) ([]int, error) {
	ids := make([]int, 0)
	err := r.db.WithContext(ctx).Table("knowledge_base_plugin_group_members members").
		Select("members.group_id").
		Joins("JOIN knowledge_base_plugin_groups groups ON groups.id = members.group_id").
		Where("members.user_id = ?", userID).
		Pluck("members.group_id", &ids).Error
	return ids, err
}

func (r *PluginRecordGroupRepository) ValidateIDsBelongToKnowledgeBase(ctx context.Context, kbID string, groupIDs []int64) error {
	if len(groupIDs) == 0 {
		return nil
	}
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.PluginRecordGroup{}).Where("kb_id = ? AND id IN ?", kbID, groupIDs).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(groupIDs)) {
		return errors.New("plugin record group does not belong to knowledge base")
	}
	return nil
}

func (r *PluginRecordGroupRepository) Create(ctx context.Context, group *domain.PluginRecordGroup) error {
	if err := group.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("MemberUserIDs").Create(group).Error; err != nil {
			return err
		}
		return replacePluginRecordGroupMembers(tx, group.ID, group.MemberUserIDs)
	})
}

func (r *PluginRecordGroupRepository) Update(ctx context.Context, group *domain.PluginRecordGroup) error {
	if err := group.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&domain.PluginRecordGroup{}).Where("id = ? AND kb_id = ?", group.ID, group.KBID).Update("name", strings.TrimSpace(group.Name))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrPluginRecordGroupNotFound
		}
		if err := tx.Where("group_id = ?", group.ID).Delete(&pluginRecordGroupMember{}).Error; err != nil {
			return err
		}
		return replacePluginRecordGroupMembers(tx, group.ID, group.MemberUserIDs)
	})
}

func (r *PluginRecordGroupRepository) Delete(ctx context.Context, kbID string, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ? AND kb_id = ?", id, kbID).Delete(&domain.PluginRecordGroup{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrPluginRecordGroupNotFound
	}
	return nil
}

type pluginRecordGroupMember struct {
	GroupID int64  `gorm:"column:group_id;primaryKey"`
	UserID  string `gorm:"column:user_id;primaryKey"`
}

func (pluginRecordGroupMember) TableName() string { return "knowledge_base_plugin_group_members" }
func replacePluginRecordGroupMembers(tx *gorm.DB, groupID int64, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}
	members := make([]pluginRecordGroupMember, 0, len(userIDs))
	for _, userID := range userIDs {
		members = append(members, pluginRecordGroupMember{GroupID: groupID, UserID: strings.TrimSpace(userID)})
	}
	return tx.Create(&members).Error
}
func (r *PluginRecordGroupRepository) loadMembers(ctx context.Context, group *domain.PluginRecordGroup) error {
	return r.db.WithContext(ctx).Model(&pluginRecordGroupMember{}).Where("group_id = ?", group.ID).Order("user_id ASC").Pluck("user_id", &group.MemberUserIDs).Error
}
