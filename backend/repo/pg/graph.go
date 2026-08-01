package pg

import (
	"context"
	"fmt"
	"slices"
	"sort"
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

type ReplaceNodeExtractionOptions struct {
	// SummaryRefreshSucceeded means the separate summary model call completed
	// with at least one usable source summary. The repository also verifies the
	// extraction is non-empty before replacing prior rows.
	SummaryRefreshSucceeded bool
}

func NewGraphRepository(db *storepg.DB, logger *log.Logger) *GraphRepository {
	return &GraphRepository{db: db, logger: logger.WithModule("repo.pg.graph")}
}

// ReplaceNodeExtraction removes evidence derived from one node before inserting
// fresh facts. It never removes a relation still evidenced by another node.
// Source summaries are replaced only when the extraction contains at least one
// non-empty result from the separate summary refresh; an empty refresh keeps the
// previous source summaries while graph facts and evidence still advance.
func (r *GraphRepository) ReplaceNodeExtraction(ctx context.Context, kbID, nodeID, nodeReleaseID string, extraction domain.GraphExtraction, options ReplaceNodeExtractionOptions) error {
	if err := extraction.Validate(); err != nil {
		return err
	}
	replaceSummaries := shouldReplaceGraphEntitySummaries(options.SummaryRefreshSucceeded, extraction)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if replaceSummaries {
			if err := tx.Where("kb_id = ? AND node_id = ?", kbID, nodeID).Delete(&domain.GraphEntitySummary{}).Error; err != nil {
				return err
			}
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
			if replaceSummaries {
				if summary := strings.TrimSpace(entity.Summary); summary != "" {
					summaries[stored.ID] = summary
				}
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

func shouldReplaceGraphEntitySummaries(summaryRefreshSucceeded bool, extraction domain.GraphExtraction) bool {
	if !summaryRefreshSucceeded {
		return false
	}
	for _, entity := range extraction.Entities {
		if strings.TrimSpace(entity.Summary) != "" {
			return true
		}
	}
	return false
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
	Entities  []VisibleGraphEntity   `json:"entities"`
	Relations []VisibleGraphRelation `json:"relations"`
	Schema    domain.KnowledgeSchema `json:"schema"`
}

// VisibleGraphEntity is a response projection. Summary is deliberately not a
// GraphEntity field because it is assembled only from source nodes visible to
// the current caller.
type VisibleGraphEntity struct {
	domain.GraphEntity
	Summary string `json:"summary,omitempty"`
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

type visibleGraphSummaryRow struct {
	EntityID string `gorm:"column:entity_id"`
	NodeID   string `gorm:"column:node_id"`
	Summary  string `gorm:"column:summary"`
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
	query = applyVisibleGraphNodePermissionFilter(query, authGroupIDs)

	var rows []row
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := &VisibleGraph{Entities: make([]VisibleGraphEntity, 0), Relations: make([]VisibleGraphRelation, 0)}
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
	var entities []domain.GraphEntity
	if err := r.db.WithContext(ctx).Where("id IN ?", entityIDs).Find(&entities).Error; err != nil {
		return nil, err
	}
	summaryQuery := buildVisibleGraphSummaryQuery(r.db.WithContext(ctx), kbID, entityIDs, authGroupIDs)
	var summaryRows []visibleGraphSummaryRow
	if err := summaryQuery.Find(&summaryRows).Error; err != nil {
		return nil, err
	}
	result.Entities = buildVisibleGraphEntities(entities, summaryRows)
	return result, nil
}

func buildVisibleGraphSummaryQuery(db *gorm.DB, kbID string, entityIDs []string, authGroupIDs []int) *gorm.DB {
	query := db.Table("graph_entity_summaries s").
		Select("s.entity_id, s.node_id, s.summary").
		Joins("JOIN nodes n ON n.id = s.node_id AND n.kb_id = s.kb_id").
		Where("s.kb_id = ? AND s.entity_id IN ?", kbID, entityIDs).
		Order("s.entity_id ASC, s.node_id ASC")
	return applyVisibleGraphNodePermissionFilter(query, authGroupIDs)
}

// applyVisibleGraphNodePermissionFilter is shared by graph evidence and source
// summaries so a summary can only be returned when its own node is visitable.
func applyVisibleGraphNodePermissionFilter(query *gorm.DB, authGroupIDs []int) *gorm.DB {
	if len(authGroupIDs) == 0 {
		return query.Where("n.permissions->>'visitable' = ?", consts.NodeAccessPermOpen)
	}
	return query.Where(`n.permissions->>'visitable' = ? OR (n.permissions->>'visitable' = ? AND EXISTS (
		SELECT 1 FROM node_auth_groups nag
		WHERE nag.node_id = n.id AND nag.perm = ? AND nag.auth_group_id IN ?
	))`, consts.NodeAccessPermOpen, consts.NodeAccessPermPartial, consts.NodePermNameVisitable, authGroupIDs)
}

func buildVisibleGraphEntities(entities []domain.GraphEntity, rows []visibleGraphSummaryRow) []VisibleGraphEntity {
	orderedRows := append([]visibleGraphSummaryRow(nil), rows...)
	sort.Slice(orderedRows, func(i, j int) bool {
		if orderedRows[i].EntityID != orderedRows[j].EntityID {
			return orderedRows[i].EntityID < orderedRows[j].EntityID
		}
		if orderedRows[i].NodeID != orderedRows[j].NodeID {
			return orderedRows[i].NodeID < orderedRows[j].NodeID
		}
		return orderedRows[i].Summary < orderedRows[j].Summary
	})

	summariesByEntity := make(map[string][]string, len(entities))
	seenByEntity := make(map[string]map[string]struct{}, len(entities))
	for _, row := range orderedRows {
		summary := strings.TrimSpace(row.Summary)
		if summary == "" || len(summariesByEntity[row.EntityID]) == 3 {
			continue
		}
		seen := seenByEntity[row.EntityID]
		if seen == nil {
			seen = make(map[string]struct{})
			seenByEntity[row.EntityID] = seen
		}
		if _, exists := seen[summary]; exists {
			continue
		}
		seen[summary] = struct{}{}
		summariesByEntity[row.EntityID] = append(summariesByEntity[row.EntityID], summary)
	}

	result := make([]VisibleGraphEntity, 0, len(entities))
	for _, entity := range entities {
		result = append(result, VisibleGraphEntity{
			GraphEntity: entity,
			Summary:     truncateVisibleGraphSummary(strings.Join(summariesByEntity[entity.ID], "\n\n")),
		})
	}
	return result
}

func truncateVisibleGraphSummary(summary string) string {
	runes := []rune(summary)
	if len(runes) <= domain.MaxGraphEntitySummaryLength {
		return summary
	}
	return string(runes[:domain.MaxGraphEntitySummaryLength])
}

func newGraphID() string {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.NewString()
	}
	return id.String()
}
