package domain

import (
	"fmt"
	"strings"
	"time"
)

// PluginRecordGroup is a knowledge-base scoped group for server-backed
// plugin records. It deliberately uses ordinary PandaWiki user IDs rather
// than portal-only auth identities, so desktop JWT users can share records.
type PluginRecordGroup struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	KBID          string    `json:"kb_id" gorm:"column:kb_id;not null;index"`
	Name          string    `json:"name" gorm:"column:name;not null"`
	MemberUserIDs []string  `json:"member_user_ids" gorm:"-"`
	CreatedBy     string    `json:"created_by" gorm:"column:created_by;not null"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (PluginRecordGroup) TableName() string { return "knowledge_base_plugin_groups" }

func (g PluginRecordGroup) Validate() error {
	if g.ID < 0 || strings.TrimSpace(g.KBID) == "" || strings.TrimSpace(g.Name) == "" || len([]rune(strings.TrimSpace(g.Name))) > 100 {
		return fmt.Errorf("invalid plugin record group")
	}
	seen := make(map[string]struct{}, len(g.MemberUserIDs))
	for _, userID := range g.MemberUserIDs {
		userID = strings.TrimSpace(userID)
		if userID == "" {
			return fmt.Errorf("invalid plugin record group member")
		}
		if _, exists := seen[userID]; exists {
			return fmt.Errorf("duplicate plugin record group member")
		}
		seen[userID] = struct{}{}
	}
	return nil
}
