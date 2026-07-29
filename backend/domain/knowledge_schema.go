package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type KnowledgeFieldTarget string

const (
	KnowledgeFieldTargetEntity KnowledgeFieldTarget = "entity"
)

type KnowledgeFieldValueType string

const (
	KnowledgeFieldValueTypeText    KnowledgeFieldValueType = "text"
	KnowledgeFieldValueTypeNumber  KnowledgeFieldValueType = "number"
	KnowledgeFieldValueTypeDate    KnowledgeFieldValueType = "date"
	KnowledgeFieldValueTypeBoolean KnowledgeFieldValueType = "boolean"
	KnowledgeFieldValueTypeSelect  KnowledgeFieldValueType = "select"
)

const (
	MaxKnowledgeSchemaFields        = 40
	MaxKnowledgeFieldInstruction    = 600
	MaxKnowledgeAttributeValues     = 24
	MaxKnowledgeAttributeTextLength = 1200
)

var knowledgeFieldKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

// GraphAttributes is dynamic only at the value level. Its keys and value
// types are validated by the active, administrator-owned KnowledgeSchema.
type GraphAttributes map[string][]any

func (a GraphAttributes) Value() (driver.Value, error) {
	if a == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(a)
}

func (a *GraphAttributes) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid graph attributes type: %T", value)
	}
	var attributes GraphAttributes
	if err := json.Unmarshal(bytes, &attributes); err != nil {
		return fmt.Errorf("decode graph attributes: %w", err)
	}
	if attributes == nil {
		attributes = GraphAttributes{}
	}
	*a = attributes
	return nil
}

type KnowledgeField struct {
	Key                string                  `json:"key"`
	Label              string                  `json:"label"`
	Target             KnowledgeFieldTarget    `json:"target"`
	EntityTypes        []GraphEntityType       `json:"entity_types,omitempty"`
	ValueType          KnowledgeFieldValueType `json:"value_type"`
	Multiple           bool                    `json:"multiple"`
	Filterable         bool                    `json:"filterable"`
	Enabled            bool                    `json:"enabled"`
	Options            []string                `json:"options,omitempty"`
	ExtractInstruction string                  `json:"extract_instruction"`
}

type KnowledgeNavigationSection struct {
	ID          string            `json:"id"`
	Label       string            `json:"label"`
	EntityTypes []GraphEntityType `json:"entity_types"`
	FieldKeys   []string          `json:"field_keys,omitempty"`
	Order       int               `json:"order"`
	Enabled     bool              `json:"enabled"`
}

type KnowledgeSchema struct {
	Version    int                          `json:"version"`
	Fields     []KnowledgeField             `json:"fields"`
	Navigation []KnowledgeNavigationSection `json:"navigation"`
}

// DefaultKnowledgeSchema is a server-side fallback, so a knowledge base that
// predates configurable schemas still has a stable, safe presentation.
func DefaultKnowledgeSchema() KnowledgeSchema {
	return KnowledgeSchema{
		Version: 1,
		Fields:  []KnowledgeField{},
		Navigation: []KnowledgeNavigationSection{
			{ID: "overview", Label: "Overview", EntityTypes: []GraphEntityType{GraphEntityTypeDocument, GraphEntityTypeEvent}, Order: 1, Enabled: true},
			{ID: "entities", Label: "Entities", EntityTypes: []GraphEntityType{GraphEntityTypePerson, GraphEntityTypeOrganization}, Order: 2, Enabled: true},
			{ID: "concepts", Label: "Concepts", EntityTypes: []GraphEntityType{GraphEntityTypeConcept, GraphEntityTypeMethod}, Order: 3, Enabled: true},
		},
	}
}

type KnowledgeGraphSchemaRecord struct {
	KBID      string          `json:"kb_id" gorm:"primaryKey;column:kb_id"`
	Schema    KnowledgeSchema `json:"schema" gorm:"type:jsonb;not null"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (KnowledgeGraphSchemaRecord) TableName() string { return "knowledge_graph_schemas" }

func (s KnowledgeSchema) Value() (driver.Value, error) { return json.Marshal(s) }

func (s *KnowledgeSchema) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid knowledge schema type: %T", value)
	}
	if err := json.Unmarshal(bytes, s); err != nil {
		return fmt.Errorf("decode knowledge schema: %w", err)
	}
	return s.Validate()
}

func (s KnowledgeSchema) Validate() error {
	if s.Version < 1 || len(s.Fields) > MaxKnowledgeSchemaFields {
		return fmt.Errorf("invalid knowledge schema version or field count")
	}
	keys := make(map[string]struct{}, len(s.Fields))
	for _, field := range s.Fields {
		if err := field.Validate(); err != nil {
			return err
		}
		if _, exists := keys[field.Key]; exists {
			return fmt.Errorf("duplicate knowledge field key %q", field.Key)
		}
		keys[field.Key] = struct{}{}
	}
	sectionIDs := make(map[string]struct{}, len(s.Navigation))
	for _, section := range s.Navigation {
		if !knowledgeFieldKeyPattern.MatchString(section.ID) || strings.TrimSpace(section.Label) == "" {
			return fmt.Errorf("invalid navigation section")
		}
		if _, exists := sectionIDs[section.ID]; exists {
			return fmt.Errorf("duplicate navigation section %q", section.ID)
		}
		sectionIDs[section.ID] = struct{}{}
		for _, key := range section.FieldKeys {
			if _, exists := keys[key]; !exists {
				return fmt.Errorf("navigation references unknown field %q", key)
			}
		}
	}
	return nil
}

func (f KnowledgeField) Validate() error {
	if !knowledgeFieldKeyPattern.MatchString(f.Key) || strings.TrimSpace(f.Label) == "" {
		return fmt.Errorf("invalid knowledge field identity")
	}
	if f.Target != KnowledgeFieldTargetEntity || !validKnowledgeValueType(f.ValueType) {
		return fmt.Errorf("invalid knowledge field target or value type")
	}
	if f.Enabled && (strings.TrimSpace(f.ExtractInstruction) == "" || utf8.RuneCountInString(f.ExtractInstruction) > MaxKnowledgeFieldInstruction) {
		return fmt.Errorf("invalid extraction instruction for field %q", f.Key)
	}
	if f.ValueType == KnowledgeFieldValueTypeSelect {
		if len(f.Options) == 0 {
			return fmt.Errorf("select field %q requires options", f.Key)
		}
		options := make(map[string]struct{}, len(f.Options))
		for _, option := range f.Options {
			option = strings.TrimSpace(option)
			if option == "" {
				return fmt.Errorf("select field %q has empty option", f.Key)
			}
			if _, exists := options[option]; exists {
				return fmt.Errorf("select field %q has duplicate option", f.Key)
			}
			options[option] = struct{}{}
		}
	} else if len(f.Options) > 0 {
		return fmt.Errorf("only select field %q may define options", f.Key)
	}
	for _, entityType := range f.EntityTypes {
		if !validGraphEntityType(entityType) {
			return fmt.Errorf("field %q has invalid entity type", f.Key)
		}
	}
	return nil
}

func (s KnowledgeSchema) ValidateEntityAttributes(entityType GraphEntityType, attributes GraphAttributes) error {
	if err := s.Validate(); err != nil {
		return err
	}
	fields := make(map[string]KnowledgeField, len(s.Fields))
	for _, field := range s.Fields {
		if field.Enabled && (len(field.EntityTypes) == 0 || containsEntityType(field.EntityTypes, entityType)) {
			fields[field.Key] = field
		}
	}
	for key, values := range attributes {
		field, exists := fields[key]
		if !exists {
			return fmt.Errorf("unknown or disabled entity attribute %q", key)
		}
		if len(values) == 0 || len(values) > MaxKnowledgeAttributeValues || (!field.Multiple && len(values) > 1) {
			return fmt.Errorf("invalid value count for attribute %q", key)
		}
		for _, value := range values {
			if err := validateKnowledgeAttributeValue(field, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func validKnowledgeValueType(value KnowledgeFieldValueType) bool {
	switch value {
	case KnowledgeFieldValueTypeText, KnowledgeFieldValueTypeNumber, KnowledgeFieldValueTypeDate, KnowledgeFieldValueTypeBoolean, KnowledgeFieldValueTypeSelect:
		return true
	default:
		return false
	}
}

func validateKnowledgeAttributeValue(field KnowledgeField, value any) error {
	switch field.ValueType {
	case KnowledgeFieldValueTypeText:
		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" || utf8.RuneCountInString(text) > MaxKnowledgeAttributeTextLength {
			return fmt.Errorf("attribute %q requires bounded text", field.Key)
		}
	case KnowledgeFieldValueTypeNumber:
		switch value.(type) {
		case float64, float32, int, int64, int32, json.Number:
		default:
			return fmt.Errorf("attribute %q requires number", field.Key)
		}
	case KnowledgeFieldValueTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("attribute %q requires boolean", field.Key)
		}
	case KnowledgeFieldValueTypeDate:
		text, ok := value.(string)
		if !ok {
			return fmt.Errorf("attribute %q requires date", field.Key)
		}
		if _, err := time.Parse("2006-01-02", text); err != nil {
			return fmt.Errorf("attribute %q requires ISO date", field.Key)
		}
	case KnowledgeFieldValueTypeSelect:
		text, ok := value.(string)
		if !ok || !containsString(field.Options, text) {
			return fmt.Errorf("attribute %q has invalid option", field.Key)
		}
	}
	return nil
}

func containsEntityType(types []GraphEntityType, value GraphEntityType) bool {
	for _, item := range types {
		if item == value {
			return true
		}
	}
	return false
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
