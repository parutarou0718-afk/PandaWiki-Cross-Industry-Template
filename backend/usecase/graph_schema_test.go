package usecase

import (
	"testing"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/stretchr/testify/require"
)

func TestBuildGraphExtractionPromptIncludesEnabledFieldsOnly(t *testing.T) {
	prompt := buildGraphExtractionPrompt(domain.KnowledgeSchema{
		Version: 1,
		Fields: []domain.KnowledgeField{
			{Key: "argument", Label: "Argument", Target: domain.KnowledgeFieldTargetEntity, ValueType: domain.KnowledgeFieldValueTypeText, Enabled: true, ExtractInstruction: "Extract explicit arguments."},
			{Key: "private_note", Label: "Private note", Target: domain.KnowledgeFieldTargetEntity, ValueType: domain.KnowledgeFieldValueTypeText, Enabled: false},
		},
	})

	require.Contains(t, prompt, `"argument"`)
	require.Contains(t, prompt, "Extract explicit arguments.")
	require.NotContains(t, prompt, "private_note")
}
