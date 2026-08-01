package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
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

func TestGraphExtractionPromptDoesNotRequestSourceSummary(t *testing.T) {
	prompt := buildGraphExtractionPrompt(domain.DefaultKnowledgeSchema())

	require.NotContains(t, prompt, `"summary"`)
	require.NotContains(t, prompt, "entity summary")
}

func TestParseGraphExtractionResponseDiscardsSummaryFromInitialCall(t *testing.T) {
	extraction, err := parseGraphExtractionResponse(`{
		"entities":[{"name":"PandaWiki","type":"organization","summary":"must not survive the raw-document call"}],
		"relations":[]
	}`)

	require.NoError(t, err)
	require.Len(t, extraction.Entities, 1)
	require.Empty(t, extraction.Entities[0].Summary)
}

func TestRequestGraphEntitySummariesSendsOnlySanitizedProjection(t *testing.T) {
	chatModel := &recordingGraphSummaryModel{response: `{"summaries":[{"name":"PandaWiki","summary":"A server-side wiki."}]}`}
	extraction := domain.GraphExtraction{
		Entities: []domain.GraphExtractedEntity{{
			Name:       "PandaWiki",
			Type:       domain.GraphEntityTypeOrganization,
			Summary:    "raw document secret must not be forwarded",
			Attributes: domain.GraphAttributes{"argument": {"schema-approved fact"}},
		}},
		Relations: []domain.GraphExtractedRelation{{
			Source: "PandaWiki", Target: "Knowledge graph", Type: domain.GraphRelationTypeRelatedTo, Evidence: "bounded evidence",
		}},
	}

	summaries, err := (&LLMUsecase{}).requestGraphEntitySummaries(context.Background(), chatModel, extraction)

	require.NoError(t, err)
	require.Equal(t, map[string]string{"pandawiki": "A server-side wiki."}, summaries)
	require.Len(t, chatModel.messages, 2)
	require.Equal(t, schema.System, chatModel.messages[0].Role)
	require.Equal(t, schema.User, chatModel.messages[1].Role)
	require.JSONEq(t, `{
		"entities":[{"name":"PandaWiki","type":"organization","attributes":{"argument":["schema-approved fact"]}}],
		"relations":[{"source":"PandaWiki","target":"Knowledge graph","type":"related_to","evidence":"bounded evidence"}]
	}`, chatModel.messages[1].Content)
	require.NotContains(t, chatModel.messages[1].Content, "raw document secret")
}

func TestAttachGraphEntitySummariesKeepsFactsAndBoundsOutput(t *testing.T) {
	extraction := domain.GraphExtraction{
		Entities:  []domain.GraphExtractedEntity{{Name: "PandaWiki", Type: domain.GraphEntityTypeOrganization}},
		Relations: []domain.GraphExtractedRelation{{Source: "PandaWiki", Target: "Graph", Type: domain.GraphRelationTypeRelatedTo, Evidence: "fact"}},
	}
	tooLong := strings.Repeat("界", domain.MaxGraphEntitySummaryLength+20)

	result := attachGraphEntitySummaries(extraction, map[string]string{"pandawiki": tooLong})

	require.Equal(t, extraction.Relations, result.Relations)
	require.Equal(t, extraction.Entities[0].Name, result.Entities[0].Name)
	require.Len(t, []rune(result.Entities[0].Summary), domain.MaxGraphEntitySummaryLength)
}

func TestAttachGraphEntitySummariesWithEmptyResultLeavesFactsIntact(t *testing.T) {
	extraction := domain.GraphExtraction{
		Entities:  []domain.GraphExtractedEntity{{Name: "PandaWiki", Type: domain.GraphEntityTypeOrganization}},
		Relations: []domain.GraphExtractedRelation{{Source: "PandaWiki", Target: "Graph", Type: domain.GraphRelationTypeRelatedTo, Evidence: "fact"}},
	}

	result := attachGraphEntitySummaries(extraction, nil)

	require.Equal(t, extraction, result)
}

type recordingGraphSummaryModel struct {
	response string
	err      error
	messages []*schema.Message
}

func (m *recordingGraphSummaryModel) Generate(_ context.Context, input []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	m.messages = input
	if m.err != nil {
		return nil, m.err
	}
	return schema.AssistantMessage(m.response, nil), nil
}

func (m *recordingGraphSummaryModel) Stream(_ context.Context, _ []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, errors.New("stream not supported")
}
