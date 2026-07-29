package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGraphExtractionValidateAcceptsBoundedStructuredFacts(t *testing.T) {
	extraction := GraphExtraction{
		Entities: []GraphExtractedEntity{{Name: "PandaWiki", Type: GraphEntityTypeOrganization}},
		Relations: []GraphExtractedRelation{{
			Source:     "PandaWiki",
			Target:     "Knowledge graph",
			Type:       GraphRelationTypeRelatedTo,
			Confidence: 0.85,
			Evidence:   "PandaWiki supports knowledge graph extraction.",
		}},
	}

	require.NoError(t, extraction.Validate())
}

func TestGraphExtractionValidateRejectsUnsafeOrUnsupportedFacts(t *testing.T) {
	tests := []struct {
		name       string
		extraction GraphExtraction
	}{
		{
			name:       "unknown entity type",
			extraction: GraphExtraction{Entities: []GraphExtractedEntity{{Name: "PandaWiki", Type: "sql"}}},
		},
		{
			name:       "unknown relation type",
			extraction: GraphExtraction{Relations: []GraphExtractedRelation{{Source: "A", Target: "B", Type: "executes", Confidence: 0.5}}},
		},
		{
			name:       "confidence out of range",
			extraction: GraphExtraction{Relations: []GraphExtractedRelation{{Source: "A", Target: "B", Type: GraphRelationTypeRelatedTo, Confidence: 1.1}}},
		},
		{
			name:       "evidence too long",
			extraction: GraphExtraction{Relations: []GraphExtractedRelation{{Source: "A", Target: "B", Type: GraphRelationTypeRelatedTo, Confidence: 0.5, Evidence: string(make([]byte, MaxGraphEvidenceLength+1))}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.extraction.Validate())
		})
	}
}

func TestParseGraphExtractionRejectsUnknownFields(t *testing.T) {
	_, err := ParseGraphExtraction(`{"entities": [], "relations": [], "raw_document": "must not persist"}`)
	require.ErrorIs(t, err, ErrInvalidGraphExtraction)
}

func TestParseGraphExtractionRejectsMultipleJSONValues(t *testing.T) {
	_, err := ParseGraphExtraction(`{"entities": [], "relations": []} {"entities": [], "relations": []}`)
	require.ErrorIs(t, err, ErrInvalidGraphExtraction)
}
