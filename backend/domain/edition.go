package domain

import "time"

type EditionID string

const (
	EditionCommon   EditionID = "common"
	EditionResearch EditionID = "research"
	EditionLegal    EditionID = "legal"
	EditionFinance  EditionID = "finance"
)

type EditionBranding struct {
	Logo string `json:"logo"`
}

type EditionPrompts struct {
	Chat    string `json:"chat"`
	Summary string `json:"summary"`
}

type EditionConfig struct {
	SchemaVersion   int               `json:"schema_version"`
	EditionID       EditionID         `json:"edition_id"`
	EditionVersion  string            `json:"edition_version"`
	ProductName     string            `json:"product_name"`
	ShortName       string            `json:"short_name"`
	Branding        EditionBranding   `json:"branding"`
	HomeDescription string            `json:"home_description"`
	Terminology     map[string]string `json:"terminology"`
	EnabledFeatures []string          `json:"enabled_features"`
	DocumentTypes   []string          `json:"document_types"`
	RelationTypes   []string          `json:"relation_types"`
	DefaultPrompts  EditionPrompts    `json:"default_prompts"`
}

// PublicEditionConfig is the allowlisted subset safe for unauthenticated Wiki pages.
type PublicEditionConfig struct {
	EditionID       EditionID         `json:"edition_id"`
	EditionVersion  string            `json:"edition_version"`
	ProductName     string            `json:"product_name"`
	ShortName       string            `json:"short_name"`
	Branding        EditionBranding   `json:"branding"`
	HomeDescription string            `json:"home_description"`
	Terminology     map[string]string `json:"terminology"`
	EnabledFeatures []string          `json:"enabled_features"`
	DocumentTypes   []string          `json:"document_types"`
	RelationTypes   []string          `json:"relation_types"`
}

func (c *EditionConfig) Public() *PublicEditionConfig {
	return &PublicEditionConfig{EditionID: c.EditionID, EditionVersion: c.EditionVersion, ProductName: c.ProductName, ShortName: c.ShortName, Branding: c.Branding, HomeDescription: c.HomeDescription, Terminology: c.Terminology, EnabledFeatures: c.EnabledFeatures, DocumentTypes: c.DocumentTypes, RelationTypes: c.RelationTypes}
}

type EditionOverrides struct {
	ProductName     *string           `json:"product_name,omitempty"`
	ShortName       *string           `json:"short_name,omitempty"`
	Branding        *EditionBranding  `json:"branding,omitempty"`
	HomeDescription *string           `json:"home_description,omitempty"`
	Terminology     map[string]string `json:"terminology,omitempty"`
	EnabledFeatures []string          `json:"enabled_features,omitempty"`
	DocumentTypes   []string          `json:"document_types,omitempty"`
	RelationTypes   []string          `json:"relation_types,omitempty"`
	DefaultPrompts  *EditionPrompts   `json:"default_prompts,omitempty"`
}

type StoredEditionConfig struct {
	SchemaVersion  int              `json:"schema_version"`
	EditionID      EditionID        `json:"edition_id"`
	EditionVersion string           `json:"edition_version"`
	Overrides      EditionOverrides `json:"overrides"`
	UpdatedAt      time.Time        `json:"updated_at"`
	UpdatedBy      uint             `json:"updated_by"`
}

type UpdateEditionReq struct {
	EditionID EditionID        `json:"edition_id" validate:"required"`
	Overrides EditionOverrides `json:"overrides"`
}
