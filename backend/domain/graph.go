package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"
)

// GraphEntityType is deliberately small in V0.1.  It is a data contract, not
// an open-ended LLM label, so clients can render and filter it safely.
type GraphEntityType string

const (
	GraphEntityTypePerson       GraphEntityType = "person"
	GraphEntityTypeOrganization GraphEntityType = "organization"
	GraphEntityTypeConcept      GraphEntityType = "concept"
	GraphEntityTypeMethod       GraphEntityType = "method"
	GraphEntityTypeEvent        GraphEntityType = "event"
	GraphEntityTypeDocument     GraphEntityType = "document"
	GraphEntityTypeOther        GraphEntityType = "other"
)

type GraphRelationType string

const (
	GraphRelationTypeMentions    GraphRelationType = "mentions"
	GraphRelationTypeRelatedTo   GraphRelationType = "related_to"
	GraphRelationTypePartOf      GraphRelationType = "part_of"
	GraphRelationTypeCauses      GraphRelationType = "causes"
	GraphRelationTypeContradicts GraphRelationType = "contradicts"
	GraphRelationTypeCites       GraphRelationType = "cites"
)

const (
	MaxGraphEntityNameLength = 160
	MaxGraphEvidenceLength   = 512
	MaxGraphFactsPerNode     = 80
)

var (
	ErrInvalidGraphExtraction = errors.New("invalid graph extraction")
	ErrGraphEntityNotFound    = errors.New("graph entity not found")
)

type GraphEntity struct {
	ID         string          `json:"id" gorm:"primaryKey;type:text"`
	KBID       string          `json:"kb_id" gorm:"column:kb_id;not null;index"`
	Name       string          `json:"name" gorm:"not null"`
	NameKey    string          `json:"-" gorm:"column:name_key;not null"`
	Type       GraphEntityType `json:"type" gorm:"not null"`
	Attributes GraphAttributes `json:"attributes" gorm:"type:jsonb;not null"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

func (GraphEntity) TableName() string { return "graph_entities" }

type GraphRelation struct {
	ID             string            `json:"id" gorm:"primaryKey;type:text"`
	KBID           string            `json:"kb_id" gorm:"column:kb_id;not null;index"`
	SourceEntityID string            `json:"source_entity_id" gorm:"not null;index"`
	TargetEntityID string            `json:"target_entity_id" gorm:"not null;index"`
	Type           GraphRelationType `json:"type" gorm:"not null"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

func (GraphRelation) TableName() string { return "graph_relations" }

// GraphEvidence is intentionally the only persisted link to source content.
// The excerpt is bounded and is never a replacement for the source node.
type GraphEvidence struct {
	ID            string    `json:"id" gorm:"primaryKey;type:text"`
	KBID          string    `json:"kb_id" gorm:"column:kb_id;not null;index"`
	RelationID    string    `json:"relation_id" gorm:"not null;index"`
	NodeID        string    `json:"node_id" gorm:"not null;index"`
	NodeReleaseID string    `json:"node_release_id" gorm:"not null"`
	Excerpt       string    `json:"excerpt" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at"`
}

func (GraphEvidence) TableName() string { return "graph_evidence" }

type GraphExtractedEntity struct {
	Name       string          `json:"name"`
	Type       GraphEntityType `json:"type"`
	Attributes GraphAttributes `json:"attributes,omitempty"`
}

type GraphExtractedRelation struct {
	Source     string            `json:"source"`
	Target     string            `json:"target"`
	Type       GraphRelationType `json:"type"`
	Confidence float64           `json:"confidence"`
	Evidence   string            `json:"evidence"`
}

type GraphExtraction struct {
	Entities  []GraphExtractedEntity   `json:"entities"`
	Relations []GraphExtractedRelation `json:"relations"`
}

func (e GraphExtraction) Validate() error {
	if len(e.Entities)+len(e.Relations) > MaxGraphFactsPerNode {
		return fmt.Errorf("%w: too many facts", ErrInvalidGraphExtraction)
	}
	for _, entity := range e.Entities {
		if !validGraphEntityType(entity.Type) || !validGraphLabel(entity.Name) {
			return fmt.Errorf("%w: invalid entity", ErrInvalidGraphExtraction)
		}
	}
	for _, relation := range e.Relations {
		if !validGraphRelationType(relation.Type) || !validGraphLabel(relation.Source) || !validGraphLabel(relation.Target) {
			return fmt.Errorf("%w: invalid relation", ErrInvalidGraphExtraction)
		}
		if relation.Confidence < 0 || relation.Confidence > 1 {
			return fmt.Errorf("%w: invalid confidence", ErrInvalidGraphExtraction)
		}
		if utf8.RuneCountInString(relation.Evidence) > MaxGraphEvidenceLength {
			return fmt.Errorf("%w: evidence exceeds limit", ErrInvalidGraphExtraction)
		}
	}
	return nil
}

func NormalizeGraphName(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}

// ParseGraphExtraction accepts only a single JSON object. Model-facing code may
// remove a Markdown fence first, but no permissive fallback is allowed here.
func ParseGraphExtraction(value string) (GraphExtraction, error) {
	var extraction GraphExtraction
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(value)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&extraction); err != nil {
		return GraphExtraction{}, fmt.Errorf("%w: decode result: %v", ErrInvalidGraphExtraction, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return GraphExtraction{}, fmt.Errorf("%w: multiple JSON values", ErrInvalidGraphExtraction)
	}
	if err := extraction.Validate(); err != nil {
		return GraphExtraction{}, err
	}
	return extraction, nil
}

func validGraphLabel(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && utf8.RuneCountInString(value) <= MaxGraphEntityNameLength
}

func validGraphEntityType(value GraphEntityType) bool {
	switch value {
	case GraphEntityTypePerson, GraphEntityTypeOrganization, GraphEntityTypeConcept, GraphEntityTypeMethod, GraphEntityTypeEvent, GraphEntityTypeDocument, GraphEntityTypeOther:
		return true
	default:
		return false
	}
}

func validGraphRelationType(value GraphRelationType) bool {
	switch value {
	case GraphRelationTypeMentions, GraphRelationTypeRelatedTo, GraphRelationTypePartOf, GraphRelationTypeCauses, GraphRelationTypeContradicts, GraphRelationTypeCites:
		return true
	default:
		return false
	}
}
