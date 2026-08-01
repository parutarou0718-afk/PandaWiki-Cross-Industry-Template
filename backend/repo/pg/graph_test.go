package pg

import (
	"strings"
	"testing"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestBuildVisibleGraphSummaryQueryUsesOpenNodePermissionFilter(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=test dbname=test sslmode=disable"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	require.NoError(t, err)

	query := buildVisibleGraphSummaryQuery(db, "kb-1", []string{"entity-public", "entity-restricted"}, nil)
	result := query.Find(&[]visibleGraphSummaryRow{})

	require.NoError(t, result.Error)
	sql := result.Statement.SQL.String()
	require.Contains(t, sql, "graph_entity_summaries s")
	require.Contains(t, sql, "JOIN nodes n ON n.id = s.node_id AND n.kb_id = s.kb_id")
	require.Contains(t, sql, "n.permissions->>'visitable'")
	require.NotContains(t, sql, "node_auth_groups")
	require.Contains(t, result.Statement.Vars, consts.NodeAccessPermOpen)
}

func TestBuildVisibleGraphSummaryQueryUsesOpenOrAuthorizedPartialNodeFilter(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=test dbname=test sslmode=disable"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	require.NoError(t, err)

	query := buildVisibleGraphSummaryQuery(db, "kb-1", []string{"entity-public", "entity-restricted"}, []int{7, 11})
	result := query.Find(&[]visibleGraphSummaryRow{})

	require.NoError(t, result.Error)
	sql := result.Statement.SQL.String()
	require.Contains(t, sql, "graph_entity_summaries s")
	require.Contains(t, sql, "n.permissions->>'visitable'")
	require.Contains(t, sql, "node_auth_groups nag")
	require.Contains(t, sql, "nag.node_id = n.id")
	require.Contains(t, sql, "nag.auth_group_id IN")
	require.Contains(t, result.Statement.Vars, consts.NodeAccessPermOpen)
	require.Contains(t, result.Statement.Vars, consts.NodeAccessPermPartial)
	require.Contains(t, result.Statement.Vars, 7)
	require.Contains(t, result.Statement.Vars, 11)
}

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

func TestShouldReplaceGraphEntitySummariesRequiresSuccessfulNonEmptyRefreshResult(t *testing.T) {
	require.False(t, shouldReplaceGraphEntitySummaries(false, domain.GraphExtraction{
		Entities: []domain.GraphExtractedEntity{{Name: "PandaWiki", Type: domain.GraphEntityTypeOrganization, Summary: "Untrusted stale value."}},
	}))
	require.False(t, shouldReplaceGraphEntitySummaries(true, domain.GraphExtraction{
		Entities: []domain.GraphExtractedEntity{{Name: "PandaWiki", Type: domain.GraphEntityTypeOrganization}},
	}))
	require.False(t, shouldReplaceGraphEntitySummaries(true, domain.GraphExtraction{
		Entities: []domain.GraphExtractedEntity{{Name: "PandaWiki", Type: domain.GraphEntityTypeOrganization, Summary: "  "}},
	}))
	require.True(t, shouldReplaceGraphEntitySummaries(true, domain.GraphExtraction{
		Entities: []domain.GraphExtractedEntity{{Name: "PandaWiki", Type: domain.GraphEntityTypeOrganization, Summary: "Fresh summary."}},
	}))
}
