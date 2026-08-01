package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// PluginRecordVisibility controls who can discover a server-backed plugin
// record after they have already passed normal knowledge-base authorization.
// It never replaces node or knowledge-base permissions.
type PluginRecordVisibility string

const (
	PluginRecordVisibilityPrivate       PluginRecordVisibility = "private"
	PluginRecordVisibilityKnowledgeBase PluginRecordVisibility = "knowledge_base"
	PluginRecordVisibilityGroups        PluginRecordVisibility = "groups"

	MaxPluginRecordPayloadBytes = 256 * 1024
)

var pluginRecordIdentifier = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{0,95}$`)

// PluginRecordPayload deliberately accepts an opaque, JSON-object payload.
// The generic record service owns persistence and access control; each plugin
// owns its own strong payload validation at its API boundary.
type PluginRecordPayload json.RawMessage

func (p PluginRecordPayload) Validate() error {
	trimmed := strings.TrimSpace(string(p))
	if trimmed == "" || len(trimmed) > MaxPluginRecordPayloadBytes || !json.Valid([]byte(trimmed)) {
		return fmt.Errorf("invalid plugin record payload")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &object); err != nil || object == nil {
		return fmt.Errorf("plugin record payload must be a JSON object")
	}
	return nil
}

func (p PluginRecordPayload) Value() (driver.Value, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return []byte(p), nil
}

func (p *PluginRecordPayload) Scan(value any) error {
	var bytes []byte
	switch stored := value.(type) {
	case []byte:
		bytes = stored
	case string:
		bytes = []byte(stored)
	default:
		return fmt.Errorf("invalid plugin record payload type: %T", value)
	}
	decoded := PluginRecordPayload(append([]byte(nil), bytes...))
	if err := decoded.Validate(); err != nil {
		return fmt.Errorf("decode plugin record payload: %w", err)
	}
	*p = decoded
	return nil
}

type PluginRecordAccess struct {
	Visibility             PluginRecordVisibility `json:"visibility" gorm:"column:visibility;not null"`
	SharedGroupIDs         pq.Int64Array          `json:"shared_group_ids" gorm:"column:shared_group_ids;type:bigint[];not null;default:'{}'"`
	AllowCollaborativeEdit bool                   `json:"allow_collaborative_edit" gorm:"column:allow_collaborative_edit;not null;default:false"`
}

func (a PluginRecordAccess) Validate() error {
	switch a.Visibility {
	case PluginRecordVisibilityPrivate, PluginRecordVisibilityKnowledgeBase:
		if len(a.SharedGroupIDs) != 0 {
			return fmt.Errorf("%s plugin record cannot declare shared groups", a.Visibility)
		}
	case PluginRecordVisibilityGroups:
		if len(a.SharedGroupIDs) == 0 {
			return fmt.Errorf("group shared plugin record requires at least one group")
		}
		seen := make(map[int64]struct{}, len(a.SharedGroupIDs))
		for _, groupID := range a.SharedGroupIDs {
			if groupID <= 0 {
				return fmt.Errorf("invalid shared plugin group")
			}
			if _, exists := seen[groupID]; exists {
				return fmt.Errorf("duplicate shared plugin group")
			}
			seen[groupID] = struct{}{}
		}
	default:
		return fmt.Errorf("invalid plugin record visibility")
	}
	return nil
}

type PluginRecord struct {
	ID          string              `json:"id" gorm:"primaryKey;type:text"`
	KBID        string              `json:"kb_id" gorm:"column:kb_id;not null;index"`
	PluginID    string              `json:"plugin_id" gorm:"column:plugin_id;not null;index"`
	RecordType  string              `json:"record_type" gorm:"column:record_type;not null;index"`
	OwnerUserID string              `json:"owner_user_id" gorm:"column:owner_user_id;not null;index"`
	Payload     PluginRecordPayload `json:"payload" gorm:"column:payload;type:jsonb;not null"`
	Access      PluginRecordAccess  `json:"access" gorm:"embedded"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   gorm.DeletedAt      `json:"-" gorm:"index"`
}

func (PluginRecord) TableName() string { return "knowledge_base_plugin_records" }

func (r PluginRecord) Validate() error {
	if strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.KBID) == "" || strings.TrimSpace(r.OwnerUserID) == "" || !pluginRecordIdentifier.MatchString(r.PluginID) || !pluginRecordIdentifier.MatchString(r.RecordType) {
		return fmt.Errorf("invalid plugin record identity")
	}
	if err := r.Payload.Validate(); err != nil {
		return err
	}
	return r.Access.Validate()
}

func (r PluginRecord) IsVisibleTo(authUserID string, authGroupIDs []int) bool {
	if r.OwnerUserID == authUserID {
		return true
	}
	switch r.Access.Visibility {
	case PluginRecordVisibilityKnowledgeBase:
		return true
	case PluginRecordVisibilityGroups:
		return slices.ContainsFunc(r.Access.SharedGroupIDs, func(sharedID int64) bool {
			return slices.Contains(authGroupIDs, int(sharedID))
		})
	default:
		return false
	}
}

func (r PluginRecord) CanEdit(authUserID string, isKnowledgeBaseController bool, authGroupIDs ...[]int) bool {
	if isKnowledgeBaseController || r.OwnerUserID == authUserID {
		return true
	}
	if !r.Access.AllowCollaborativeEdit || len(authGroupIDs) == 0 {
		return false
	}
	return r.IsVisibleTo(authUserID, authGroupIDs[0])
}
