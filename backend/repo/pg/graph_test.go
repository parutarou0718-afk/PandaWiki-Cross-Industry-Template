package pg

import (
	"strings"
	"testing"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/stretchr/testify/require"
)

func TestBuildVisibleGraphEntitiesUsesOnlyVisibleSourceSummaries(t *testing.T) {
	entities := []domain.GraphEntity{{ID: "entity-1", Name: "PandaWiki", Type: domain.GraphEntityTypeOrganization}}
	// The restricted source row must be excluded by GetVisibleGraph's SQL before
	// source summaries reach this response-assembly function.
	restrictedSourceSummary := visibleGraphSummaryRow{EntityID: "entity-1", NodeID: "restricted-node", Summary: "Restricted source summary."}
	visibleSourceSummaries := []visibleGraphSummaryRow{
		{EntityID: "entity-1", NodeID: "public-node", Summary: "Public source summary."},
	}

	result := buildVisibleGraphEntities(entities, visibleSourceSummaries)

	require.Len(t, result, 1)
	require.Equal(t, "Public source summary.", result[0].Summary)
	require.NotContains(t, result[0].Summary, restrictedSourceSummary.Summary)
}

func TestBuildVisibleGraphEntitiesDeduplicatesAndBoundsVisibleSourceSummaries(t *testing.T) {
	entities := []domain.GraphEntity{{ID: "entity-1", Name: "PandaWiki", Type: domain.GraphEntityTypeOrganization}}
	visibleSourceSummaries := []visibleGraphSummaryRow{
		{EntityID: "entity-1", NodeID: "node-3", Summary: "Third."},
		{EntityID: "entity-1", NodeID: "node-1", Summary: "First."},
		{EntityID: "entity-1", NodeID: "node-2", Summary: "Second."},
		{EntityID: "entity-1", NodeID: "node-4", Summary: "Fourth."},
		{EntityID: "entity-1", NodeID: "node-5", Summary: "First."},
	}

	result := buildVisibleGraphEntities(entities, visibleSourceSummaries)

	require.Len(t, result, 1)
	require.Equal(t, "First.\n\nSecond.\n\nThird.", result[0].Summary)
	require.LessOrEqual(t, len([]rune(result[0].Summary)), domain.MaxGraphEntitySummaryLength)
}

func TestBuildVisibleGraphEntitiesBoundsJoinedSourceSummaries(t *testing.T) {
	entities := []domain.GraphEntity{{ID: "entity-1", Name: "PandaWiki", Type: domain.GraphEntityTypeOrganization}}
	visibleSourceSummaries := []visibleGraphSummaryRow{
		{EntityID: "entity-1", NodeID: "node-1", Summary: strings.Repeat("a", 300)},
		{EntityID: "entity-1", NodeID: "node-2", Summary: strings.Repeat("b", 300)},
		{EntityID: "entity-1", NodeID: "node-3", Summary: strings.Repeat("c", 300)},
	}

	result := buildVisibleGraphEntities(entities, visibleSourceSummaries)

	require.Len(t, result, 1)
	require.Equal(t, strings.Repeat("a", 300)+"\n\n"+strings.Repeat("b", 298), result[0].Summary)
	require.Len(t, []rune(result[0].Summary), domain.MaxGraphEntitySummaryLength)
	require.NotContains(t, result[0].Summary, "c")
}
