package domain

import (
	"encoding/json"
	"strings"
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

func TestSanitizeGraphExtractionKeepsValidFactsAndDropsSchemaViolations(t *testing.T) {
	schema := KnowledgeSchema{Version: 1, Fields: []KnowledgeField{{
		Key:                "argument",
		Label:              "Argument",
		Target:             KnowledgeFieldTargetEntity,
		ValueType:          KnowledgeFieldValueTypeText,
		Enabled:            true,
		ExtractInstruction: "Extract explicit arguments only.",
	}}}

	cleaned, discarded, err := SanitizeGraphExtraction(GraphExtraction{
		Entities: []GraphExtractedEntity{
			{Name: "Valid concept", Type: GraphEntityTypeConcept, Attributes: GraphAttributes{"argument": {"supported claim"}, "research": {"invented"}}},
			{Name: "Bad type", Type: "research"},
		},
		Relations: []GraphExtractedRelation{{
			Source: "Valid concept", Target: "Source document", Type: GraphRelationTypeMentions, Confidence: 0.9, Evidence: "supported claim",
		}},
	}, schema)

	require.NoError(t, err)
	require.Equal(t, 2, discarded)
	require.Len(t, cleaned.Entities, 1)
	require.Equal(t, GraphAttributes{"argument": {"supported claim"}}, cleaned.Entities[0].Attributes)
	require.Len(t, cleaned.Relations, 1)
	require.NoError(t, cleaned.Validate())
}

func TestSanitizeGraphExtractionRetainsBoundedEntitySummaries(t *testing.T) {
	extraction, err := DecodeGraphExtraction(`{
		"entities": [
			{"name": "Short summary", "type": "concept", "summary": "A concise source-scoped description."},
			{"name": "Empty summary", "type": "concept", "summary": "   "},
			{"name": "Long summary", "type": "concept", "summary": "` + strings.Repeat("x", MaxGraphEntitySummaryLength+1) + `"}
		],
		"relations": []
	}`)
	require.NoError(t, err)

	cleaned, discarded, err := SanitizeGraphExtraction(extraction, DefaultKnowledgeSchema())
	require.NoError(t, err)
	require.Equal(t, 1, discarded)
	require.Len(t, cleaned.Entities, 3)

	encoded, err := json.Marshal(cleaned.Entities)
	require.NoError(t, err)
	require.JSONEq(t, `[
		{"name":"Short summary","type":"concept","summary":"A concise source-scoped description."},
		{"name":"Empty summary","type":"concept"},
		{"name":"Long summary","type":"concept"}
	]`, string(encoded))
	require.NoError(t, cleaned.Validate())
}

func TestSanitizeGraphExtractionRejectsModelOutputWithNoUsableFacts(t *testing.T) {
	_, _, err := SanitizeGraphExtraction(GraphExtraction{
		Entities: []GraphExtractedEntity{{Name: "Bad type", Type: "research"}},
	}, DefaultKnowledgeSchema())

	require.ErrorIs(t, err, ErrInvalidGraphExtraction)
}
