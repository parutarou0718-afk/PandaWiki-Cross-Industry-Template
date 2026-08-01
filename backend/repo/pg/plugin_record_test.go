package pg

import (
	"strings"
	"testing"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestBuildVisiblePluginRecordQueryFiltersByOwnerKnowledgeBaseOrAuthorizedGroup(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=test dbname=test sslmode=disable"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	require.NoError(t, err)

	query := buildVisiblePluginRecordQuery(db, "kb-1", "official.submission-management", "submission", "user-42", []int{7, 11})
	result := query.Find(&[]domain.PluginRecord{})

	require.NoError(t, result.Error)
	sql := result.Statement.SQL.String()
	require.Contains(t, sql, "knowledge_base_plugin_records")
	require.Contains(t, sql, "owner_user_id")
	require.Contains(t, sql, "visibility")
	require.Contains(t, sql, "shared_group_ids &&")
	require.Contains(t, result.Statement.Vars, "kb-1")
	require.Contains(t, result.Statement.Vars, "official.submission-management")
	require.Contains(t, result.Statement.Vars, "submission")
	require.Contains(t, result.Statement.Vars, "user-42")
	require.True(t, strings.Contains(sql, "deleted_at"))
}
