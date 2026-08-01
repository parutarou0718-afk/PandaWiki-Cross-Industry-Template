//go:build integration

package pg

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/chaitin/panda-wiki/domain"
	storepg "github.com/chaitin/panda-wiki/store/pg"
)

func TestGraphRepositoryReplaceNodeExtractionRetainsSummaryWhenRefreshIsEmpty(t *testing.T) {
	repo, db := newGraphRepositoryIntegrationTest(t)
	ctx := context.Background()

	initial := domain.GraphExtraction{
		Entities: []domain.GraphExtractedEntity{
			{Name: "PandaWiki", Type: domain.GraphEntityTypeOrganization, Summary: "Prior source summary."},
			{Name: "Knowledge graph", Type: domain.GraphEntityTypeConcept},
		},
		Relations: []domain.GraphExtractedRelation{{
			Source: "PandaWiki", Target: "Knowledge graph", Type: domain.GraphRelationTypeRelatedTo, Evidence: "old evidence",
		}},
	}
	require.NoError(t, repo.ReplaceNodeExtraction(ctx, "kb-retain", "node-retain", "release-1", initial, ReplaceNodeExtractionOptions{SummaryRefreshSucceeded: true}))

	withoutSummaries := domain.GraphExtraction{
		Entities: []domain.GraphExtractedEntity{
			{Name: "PandaWiki", Type: domain.GraphEntityTypeOrganization},
			{Name: "Knowledge graph", Type: domain.GraphEntityTypeConcept},
		},
		Relations: []domain.GraphExtractedRelation{{
			Source: "PandaWiki", Target: "Knowledge graph", Type: domain.GraphRelationTypeRelatedTo, Evidence: "fresh evidence",
		}},
	}
	require.NoError(t, repo.ReplaceNodeExtraction(ctx, "kb-retain", "node-retain", "release-2", withoutSummaries, ReplaceNodeExtractionOptions{SummaryRefreshSucceeded: false}))

	var summaries []domain.GraphEntitySummary
	require.NoError(t, db.Where("kb_id = ? AND node_id = ?", "kb-retain", "node-retain").Find(&summaries).Error)
	require.Len(t, summaries, 1)
	require.Equal(t, "Prior source summary.", summaries[0].Summary)
	require.Equal(t, "release-1", summaries[0].NodeReleaseID)

	var evidence []domain.GraphEvidence
	require.NoError(t, db.Where("kb_id = ? AND node_id = ?", "kb-retain", "node-retain").Find(&evidence).Error)
	require.Len(t, evidence, 1)
	require.Equal(t, "fresh evidence", evidence[0].Excerpt)
	require.Equal(t, "release-2", evidence[0].NodeReleaseID)
}

func TestGraphRepositoryGetVisibleGraphFiltersSourceSummariesByRealPostgresPermissions(t *testing.T) {
	repo, db := newGraphRepositoryIntegrationTest(t)
	ctx := context.Background()

	require.NoError(t, db.Exec(`INSERT INTO nodes (id, kb_id, permissions) VALUES
		('node-public', 'kb-visible', '{"visitable":"open"}'::jsonb),
		('node-matching', 'kb-visible', '{"visitable":"partial"}'::jsonb),
		('node-nonmatching', 'kb-visible', '{"visitable":"partial"}'::jsonb)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO node_auth_groups (node_id, auth_group_id, perm) VALUES
		('node-matching', 7, 'visitable'),
		('node-nonmatching', 11, 'visitable')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO graph_entities (id, kb_id, name, name_key, type, attributes) VALUES
		('entity-main', 'kb-visible', 'PandaWiki', 'pandawiki', 'organization', '{}'::jsonb),
		('entity-target', 'kb-visible', 'Knowledge graph', 'knowledge graph', 'concept', '{}'::jsonb)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO graph_relations (id, kb_id, source_entity_id, target_entity_id, type)
		VALUES ('relation-visible', 'kb-visible', 'entity-main', 'entity-target', 'related_to')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO graph_evidence (id, kb_id, relation_id, node_id, node_release_id, excerpt)
		VALUES ('evidence-public', 'kb-visible', 'relation-visible', 'node-public', 'release-public', 'public evidence')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO graph_entity_summaries (id, kb_id, entity_id, node_id, node_release_id, summary) VALUES
		('summary-public', 'kb-visible', 'entity-main', 'node-public', 'release-public', 'Public summary.'),
		('summary-matching', 'kb-visible', 'entity-main', 'node-matching', 'release-matching', 'Matching group summary.'),
		('summary-nonmatching', 'kb-visible', 'entity-main', 'node-nonmatching', 'release-nonmatching', 'Nonmatching group summary.')`).Error)

	publicGraph, err := repo.GetVisibleGraph(ctx, "kb-visible", nil)
	require.NoError(t, err)
	require.Equal(t, "Public summary.", visibleEntitySummary(t, publicGraph, "entity-main"))

	matchingGraph, err := repo.GetVisibleGraph(ctx, "kb-visible", []int{7})
	require.NoError(t, err)
	matchingSummary := visibleEntitySummary(t, matchingGraph, "entity-main")
	require.Contains(t, matchingSummary, "Public summary.")
	require.Contains(t, matchingSummary, "Matching group summary.")
	require.NotContains(t, matchingSummary, "Nonmatching group summary.")

	nonmatchingGraph, err := repo.GetVisibleGraph(ctx, "kb-visible", []int{99})
	require.NoError(t, err)
	require.Equal(t, "Public summary.", visibleEntitySummary(t, nonmatchingGraph, "entity-main"))
}

func newGraphRepositoryIntegrationTest(t *testing.T) (*GraphRepository, *gorm.DB) {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("POSTGRES_DSN"))
	}
	if dsn == "" {
		t.Skip("set DATABASE_URL or POSTGRES_DSN to run PostgreSQL repository integration tests")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	schemaName := "graph_repo_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, db.Exec(fmt.Sprintf(`CREATE SCHEMA "%s"`, schemaName)).Error)
	require.NoError(t, db.Exec(fmt.Sprintf(`SET search_path TO "%s"`, schemaName)).Error)
	t.Cleanup(func() {
		require.NoError(t, db.Exec("SET search_path TO public").Error)
		require.NoError(t, db.Exec(fmt.Sprintf(`DROP SCHEMA "%s" CASCADE`, schemaName)).Error)
	})

	for _, statement := range strings.Split(graphRepositoryIntegrationSchema, ";") {
		statement = strings.TrimSpace(statement)
		if statement != "" {
			require.NoError(t, db.Exec(statement).Error)
		}
	}
	return &GraphRepository{db: &storepg.DB{DB: db}}, db
}

func visibleEntitySummary(t *testing.T, graph *VisibleGraph, entityID string) string {
	t.Helper()
	for _, entity := range graph.Entities {
		if entity.ID == entityID {
			return entity.Summary
		}
	}
	t.Fatalf("entity %q was not returned", entityID)
	return ""
}

const graphRepositoryIntegrationSchema = `
CREATE TABLE nodes (
	id TEXT PRIMARY KEY,
	kb_id TEXT NOT NULL,
	permissions JSONB NOT NULL
);
CREATE TABLE node_auth_groups (
	id SERIAL PRIMARY KEY,
	node_id TEXT NOT NULL,
	auth_group_id INTEGER NOT NULL,
	perm TEXT NOT NULL
);
CREATE TABLE graph_entities (
	id TEXT PRIMARY KEY,
	kb_id TEXT NOT NULL,
	name TEXT NOT NULL,
	name_key TEXT NOT NULL,
	type TEXT NOT NULL,
	attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
	UNIQUE (kb_id, name_key)
);
CREATE TABLE graph_relations (
	id TEXT PRIMARY KEY,
	kb_id TEXT NOT NULL,
	source_entity_id TEXT NOT NULL REFERENCES graph_entities(id) ON DELETE CASCADE,
	target_entity_id TEXT NOT NULL REFERENCES graph_entities(id) ON DELETE CASCADE,
	type TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
	UNIQUE (kb_id, source_entity_id, target_entity_id, type)
);
CREATE TABLE graph_evidence (
	id TEXT PRIMARY KEY,
	kb_id TEXT NOT NULL,
	relation_id TEXT NOT NULL REFERENCES graph_relations(id) ON DELETE CASCADE,
	node_id TEXT NOT NULL,
	node_release_id TEXT NOT NULL,
	excerpt TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	UNIQUE (relation_id, node_id)
);
CREATE TABLE graph_entity_summaries (
	id TEXT PRIMARY KEY,
	kb_id TEXT NOT NULL,
	entity_id TEXT NOT NULL REFERENCES graph_entities(id) ON DELETE CASCADE,
	node_id TEXT NOT NULL,
	node_release_id TEXT NOT NULL,
	summary TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
	UNIQUE (entity_id, node_id)
);`
