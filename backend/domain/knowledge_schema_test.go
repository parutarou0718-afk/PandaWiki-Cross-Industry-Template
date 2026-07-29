package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKnowledgeSchemaValidateAcceptsAdministratorDefinedEntityFields(t *testing.T) {
	schema := KnowledgeSchema{
		Version: 1,
		Fields: []KnowledgeField{
			{
				Key:                "argument",
				Label:              "Core argument",
				Target:             KnowledgeFieldTargetEntity,
				EntityTypes:        []GraphEntityType{GraphEntityTypePerson, GraphEntityTypeConcept},
				ValueType:          KnowledgeFieldValueTypeText,
				Multiple:           true,
				Filterable:         true,
				Enabled:            true,
				ExtractInstruction: "Extract only explicit arguments.",
			},
			{
				Key:                "risk_level",
				Label:              "Risk level",
				Target:             KnowledgeFieldTargetEntity,
				ValueType:          KnowledgeFieldValueTypeSelect,
				Filterable:         true,
				Enabled:            true,
				Options:            []string{"low", "medium", "high"},
				ExtractInstruction: "Classify only explicit risk statements.",
			},
		},
	}

	require.NoError(t, schema.Validate())
	require.NoError(t, schema.ValidateEntityAttributes(GraphEntityTypePerson, GraphAttributes{
		"argument":   {"a stated claim"},
		"risk_level": {"medium"},
	}))
}

func TestKnowledgeSchemaRejectsInvalidFieldDefinitionsAndAttributes(t *testing.T) {
	schema := KnowledgeSchema{Version: 1, Fields: []KnowledgeField{{
		Key:                "risk_level",
		Label:              "Risk level",
		Target:             KnowledgeFieldTargetEntity,
		ValueType:          KnowledgeFieldValueTypeSelect,
		Enabled:            true,
		Options:            []string{"low", "high"},
		ExtractInstruction: "Extract explicit risk only.",
	}}}

	require.NoError(t, schema.Validate())
	require.Error(t, schema.ValidateEntityAttributes(GraphEntityTypeOrganization, GraphAttributes{"unknown": {"value"}}))
	require.Error(t, schema.ValidateEntityAttributes(GraphEntityTypeOrganization, GraphAttributes{"risk_level": {"medium"}}))
	require.Error(t, KnowledgeSchema{Version: 1, Fields: []KnowledgeField{
		{Key: "same_key", Label: "One", Target: KnowledgeFieldTargetEntity, ValueType: KnowledgeFieldValueTypeText, Enabled: true, ExtractInstruction: "first"},
		{Key: "same_key", Label: "Two", Target: KnowledgeFieldTargetEntity, ValueType: KnowledgeFieldValueTypeText, Enabled: true, ExtractInstruction: "second"},
	}}.Validate())
}

func TestDefaultKnowledgeSchemaHasEnabledServerControlledNavigation(t *testing.T) {
	schema := DefaultKnowledgeSchema()

	require.NoError(t, schema.Validate())
	require.NotEmpty(t, schema.Navigation)
	require.Equal(t, "overview", schema.Navigation[0].ID)
}
