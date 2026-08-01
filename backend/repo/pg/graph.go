package pg

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	storepg "github.com/chaitin/panda-wiki/store/pg"
)

type GraphRepository struct {
	db     *storepg.DB
	logger *log.Logger
}

func NewGraphRepository(db *storepg.DB, logger *log.Logger) *GraphRepository {
	return &GraphRepository{db: db, logger: logger.WithModule("repo.pg.graph")}
}

// ReplaceNodeExtraction removes evidence derived from one node before inserting
// fresh facts. It never removes a relation still evidenced by another node.
func (r *GraphRepository) ReplaceNodeExtraction(ctx context.Context, kbID, nodeID, nodeReleaseID string, extraction domain.GraphExtraction) error {
	if err := extraction.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("kb_id = ? AND node_id = ?", kbID, nodeID).Delete(&domain.GraphEntitySummary{}).Error; err != nil {
			return err
		}
		if err := tx.Where("kb_id = ? AND node_id = ?", kbID, nodeID).Delete(&domain.GraphEvidence{}).Error; err != nil {
			return err
		}
		if err := tx.Exec(`DELETE FROM graph_relations r
			WHERE r.kb_id = ?
			AND NOT EXISTS (SELECT 1 FROM graph_evidence e WHERE e.relation_id = r.id)`, kbID).Error; err != nil {
			return err
		}

		entityTypes := make(map[string]domain.GraphEntityType, len(extraction.Entities)+len(extraction.Relations)*2)
		for _, entity := range extraction.Entities {
			entityTypes[domain.NormalizeGraphName(entity.Name)] = entity.Type
		}
		for _, relation := range extraction.Relations {
			for _, name := range []string{relation.Source, relation.Target} {
				key := domain.NormalizeGraphName(name)
				if _, ok := entityTypes[key]; !ok {
					entityTypes[key] = domain.GraphEntityTypeOther
				}
			}
		}

		entities := make(map[string]*domain.GraphEntity, len(entityTypes))
		summaries := make(map[string]string, len(extraction.Entities))
		for _, entity := range extraction.Entities {
			stored, err := r.ensureEntity(tx, kbID, entity.Name, entity.Type, entity.Attributes)
			if err != nil {
				return err
			}
			key := domain.NormalizeGraphName(entity.Name)
			entities[key] = stored
			if summary := strings.TrimSpace(entity.Summary); summary != "" {
				summaries[stored.ID] = summary
			}
		}
		for _, relation := range extraction.Relations {
			for _, name := range []string{relation.Source, relation.Target} {
				key := domain.NormalizeGraphName(name)
				if _, ok := entities[key]; ok {
					continue
				}
				stored, err := r.ensureEntity(tx, kbID, name, entityTypes[key], domain.GraphAttributes{})
				if err != nil {
					return err
				}
				entities[key] = stored
			}
		}
		for entityID, summary := range summaries {
			stored := &domain.GraphEntitySummary{
				ID:            newGraphID(),
				KBID:          kbID,
				EntityID:      entityID,
				NodeID:        nodeID,
				NodeReleaseID: nodeReleaseID,
				Summary:       summary,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "entity_id"}, {Name: "node_id"}},
				DoUpdates: clause.Assignments(map[string]any{
					"kb_id":           gorm.Expr("EXCLUDED.kb_id"),
					"node_release_id": gorm.Expr("EXCLUDED.node_release_id"),
					"summary":         gorm.Expr("EXCLUDED.summary"),
					"updated_at":      gorm.Expr("NOW()"),
				}),
			}).Create(stored).Error; err != nil {
				return err
			}
		}

		for _, relation := range extraction.Relations {
			source := entities[domain.NormalizeGraphName(relation.Source)]
			target := entities[domain.NormalizeGraphName(relation.Target)]
			if source == nil || target == nil {
				return fmt.Errorf("%w: relation entity missing", domain.ErrInvalidGraphExtraction)
			}
			stored, err := r.ensureRelation(tx, kbID, source.ID, target.ID, relation.Type)
			if err != nil {
				return err
			}
			evidence := &domain.GraphEvidence{
				ID:            newGraphID(),
				KBID:          kbID,
				RelationID:    stored.ID,
				NodeID:        nodeID,
				NodeReleaseID: nodeReleaseID,
				Excerpt:       strings.TrimSpace(relation.Evidence),
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "relation_id"}, {Name: "node_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"node_release_id", "excerpt"}),
			}).Create(evidence).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GraphRepository) ensureEntity(tx *gorm.DB, kbID, name string, entityType domain.GraphEntityType, attributes domain.GraphAttributes) (*domain.GraphEntity, error) {
	entity := &domain.GraphEntity{ID: newGraphID(), KBID: kbID, Name: strings.TrimSpace(name), NameKey: domain.NormalizeGraphName(name), Type: entityType, Attributes: attributes}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "kb_id"}, {Name: "name_key"}},
		DoUpdates: clause.Assignments(map[string]any{
			"attributes": gorm.Expr("graph_entities.attributes || EXCLUDED.attributes"),
			"updated_at": gorm.Expr("NOW()"),
		}),
	}).Create(entity).Error; err != nil {
		return nil, err
	}
	if err := tx.Where("kb_id = ? AND name_key = ?", kbID, entity.NameKey).First(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *GraphRepository) ensureRelation(tx *gorm.DB, kbID, sourceID, targetID string, relationType domain.GraphRelationType) (*domain.GraphRelation, error) {
	relation := &domain.GraphRelation{ID: newGraphID(), KBID: kbID, SourceEntityID: sourceID, TargetEntityID: targetID, Type: relationType}
	if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "kb_id"}, {Name: "source_entity_id"}, {Name: "target_entity_id"}, {Name: "type"}}, DoNothing: true}).Create(relation).Error; err != nil {
		return nil, err
	}
	if err := tx.Where("kb_id = ? AND source_entity_id = ? AND target_entity_id = ? AND type = ?", kbID, sourceID, targetID, relationType).First(relation).Error; err != nil {
		return nil, err
	}
	return relation, nil
}

type VisibleGraph struct {
	Entities  []domain.GraphEntity   `json:"entities"`
	Relations []VisibleGraphRelation `json:"relations"`
	Schema    domain.KnowledgeSchema `json:"schema"`
}

type VisibleGraphRelation struct {
	ID             string                   `json:"id"`
	SourceEntityID string                   `json:"source_entity_id"`
	TargetEntityID string                   `json:"target_entity_id"`
	Type           domain.GraphRelationType `json:"type"`
	Evidence       []VisibleGraphEvidence   `json:"evidence"`
}

type VisibleGraphEvidence struct {
	NodeID        string `json:"node_id"`
	NodeReleaseID string `json:"node_release_id"`
	Excerpt       string `json:"excerpt"`
}

// GetVisibleGraph applies current node visit permissions in PostgreSQL. Group
// IDs are server-derived by the caller; clients cannot supply them.
func (r *GraphRepository) GetVisibleGraph(ctx context.Context, kbID string, authGroupIDs []int) (*VisibleGraph, error) {
	type row struct {
		RelationID     string                   `gorm:"column:relation_id"`
		SourceEntityID string                   `gorm:"column:source_entity_id"`
		TargetEntityID string                   `gorm:"column:target_entity_id"`
		RelationType   domain.GraphRelationType `gorm:"column:relation_type"`
		NodeID         string                   `gorm:"column:node_id"`
		NodeReleaseID  string                   `gorm:"column:node_release_id"`
		Excerpt        string                   `gorm:"column:excerpt"`
	}
	query := r.db.WithContext(ctx).Table("graph_evidence e").
		Select("DISTINCT r.id AS relation_id, r.source_entity_id, r.target_entity_id, r.type AS relation_type, e.node_id, e.node_release_id, e.excerpt").
		Joins("JOIN graph_relations r ON r.id = e.relation_id").
		Joins("JOIN nodes n ON n.id = e.node_id AND n.kb_id = e.kb_id").
		Where("e.kb_id = ?", kbID)

	if len(authGroupIDs) == 0 {
		query = query.Where("n.permissions->>'visitable' = ?", consts.NodeAccessPermOpen)
	} else {
		query = query.Where(`n.permissions->>'visitable' = ? OR (n.permissions->>'visitable' = ? AND EXISTS (
			SELECT 1 FROM node_auth_groups nag
			WHERE nag.node_id = n.id AND nag.perm = ? AND nag.auth_group_id IN ?
		))`, consts.NodeAccessPermOpen, consts.NodeAccessPermPartial, consts.NodePermNameVisitable, authGroupIDs)
	}

	var rows []row
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := &VisibleGraph{Entities: make([]domain.GraphEntity, 0), Relations: make([]VisibleGraphRelation, 0)}
	if len(rows) == 0 {
		return result, nil
	}
	relationByID := make(map[string]*VisibleGraphRelation, len(rows))
	entityIDs := make([]string, 0, len(rows)*2)
	for _, item := range rows {
		relation := relationByID[item.RelationID]
		if relation == nil {
			relation = &VisibleGraphRelation{ID: item.RelationID, SourceEntityID: item.SourceEntityID, TargetEntityID: item.TargetEntityID, Type: item.RelationType, Evidence: make([]VisibleGraphEvidence, 0)}
			relationByID[item.RelationID] = relation
			entityIDs = append(entityIDs, item.SourceEntityID, item.TargetEntityID)
		}
		relation.Evidence = append(relation.Evidence, VisibleGraphEvidence{NodeID: item.NodeID, NodeReleaseID: item.NodeReleaseID, Excerpt: item.Excerpt})
	}
	for _, relation := range relationByID {
		result.Relations = append(result.Relations, *relation)
	}
	entityIDs = slices.Compact(entityIDs)
	if err := r.db.WithContext(ctx).Where("id IN ?", entityIDs).Find(&result.Entities).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func newGraphID() string {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.NewString()
	}
	return id.String()
}
