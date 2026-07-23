//go:build integration

package pg

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/chaitin/panda-wiki/domain"
	store "github.com/chaitin/panda-wiki/store/pg"
)

func integrationRepo(t *testing.T) *ReportProfileRepo {
	t.Helper()
	dsn := os.Getenv("REPORT_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("REPORT_TEST_PG_DSN is required")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return NewReportProfileRepo(&store.DB{DB: db})
}
func TestReportProfilesIntegration(t *testing.T) {
	r := integrationRepo(t)
	ctx := context.Background()
	if err := r.EnsureBuiltinProfiles(ctx); err != nil {
		t.Fatal(err)
	}
	profiles, err := r.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 10 {
		t.Fatalf("got %d builtin profiles", len(profiles))
	}
	if err := r.EnsureBuiltinProfiles(ctx); err != nil {
		t.Fatal(err)
	}
	profiles, _ = r.List(ctx)
	if len(profiles) != 10 {
		t.Fatal("seed is not idempotent")
	}
	custom := &domain.ReportProfile{ID: uuid.NewString(), Name: "custom", Description: "custom", EditionIDs: []domain.EditionID{domain.EditionCommon}, SystemPrompt: "prompt", InputFields: []domain.ReportProfileInputField{{Key: "topic", Label: "Topic", Type: domain.ReportInputTypeText}}, Sections: []domain.ReportProfileSection{{Key: "summary", Title: "Summary", Instruction: "summary", Order: 1}}, Version: "1.0.0"}
	if err := r.Create(ctx, custom); err != nil {
		t.Fatal(err)
	}
	if err := r.SoftDelete(ctx, custom.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := r.GetByID(ctx, custom.ID); err == nil {
		t.Fatal("soft deleted profile was visible")
	}
}

func TestReportsAndCitationsIntegration(t *testing.T) {
	r := integrationRepo(t)
	ctx := context.Background()
	_ = r.EnsureBuiltinProfiles(ctx)
	repo := NewReportRepo(r.db)
	report := domain.Report{ID: uuid.NewString(), KBID: "kb", ProfileID: "builtin.common.knowledge-report", ProfileVersion: "1.0.0", ProfileSnapshot: domain.BuiltinReportProfiles()[0], CreatedBy: 101, InputValues: domain.ReportInputValues{"topic": {Value: "test"}}}
	if err := repo.CreatePending(ctx, &report); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkRunning(ctx, report.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.CompleteWithCitations(ctx, report.ID, "result [1]", domain.ReportResultSuccess, nil); err != nil {
		t.Fatal(err)
	}
	var loaded domain.Report
	if err := r.db.First(&loaded, "id=?", report.ID).Error; err != nil {
		t.Fatal(err)
	}
	if loaded.InputValues["topic"].Value != "test" || loaded.ProfileSnapshot.ID != report.ProfileSnapshot.ID {
		t.Fatal("jsonb round trip failed")
	}
	if loaded.Status != domain.ReportStatusCompleted || loaded.ResultCode != domain.ReportResultSuccess {
		t.Fatal("completed result code was not persisted")
	}
	if err := repo.MarkFailed(ctx, report.ID, domain.ReportResultTimeout, "late timeout"); err != ErrInvalidReportTransition {
		t.Fatalf("terminal report accepted competing transition: %v", err)
	}
	c := domain.ReportCitation{ID: uuid.NewString(), ReportID: report.ID, CitationIndex: 1, NodeID: "node", DocumentName: "doc", Locator: "chunk", Excerpt: "excerpt"}
	if err := r.db.Create(&c).Error; err != nil {
		t.Fatal(err)
	}
	listed, err := repo.ListByCreator(ctx, report.CreatedBy)
	if err != nil || len(listed) != 1 || listed[0].ID != report.ID || listed[0].CitationCount != 1 {
		t.Fatalf("report list did not return citation count: reports=%#v err=%v", listed, err)
	}
	if err := r.db.Create(&domain.ReportCitation{ID: uuid.NewString(), ReportID: report.ID, CitationIndex: 1, NodeID: "node2", DocumentName: "doc", Locator: "chunk", Excerpt: "excerpt"}).Error; err == nil {
		t.Fatal("citation index unique constraint missing")
	}
	if err := r.db.Delete(&report).Error; err != nil {
		t.Fatal(err)
	}
	var count int64
	r.db.Model(&domain.ReportCitation{}).Where("report_id=?", report.ID).Count(&count)
	if count != 0 {
		t.Fatal("citation was not cascade deleted")
	}
	if err := r.RestoreBuiltinDefault(ctx, "builtin.common.knowledge-report"); err != nil {
		t.Fatal(err)
	}
}

func TestReportProfileSnapshotRejectsIncompatibleJSON(t *testing.T) {
	r := integrationRepo(t)
	ctx := context.Background()
	if err := r.EnsureBuiltinProfiles(ctx); err != nil {
		t.Fatal(err)
	}
	id := uuid.NewString()
	badSnapshot := `{"id":"builtin.common.knowledge-report","name":"bad","edition_ids":["common"],"sections":"not-an-array"}`
	if err := r.db.Exec(`INSERT INTO reports (id,kb_id,profile_id,profile_version,profile_snapshot,created_by,status,input_values) VALUES (?,?,?,?,?::jsonb,?,?,?::jsonb)`, id, "kb", "builtin.common.knowledge-report", "1.0.0", badSnapshot, 1, "pending", `{}`).Error; err != nil {
		t.Fatal(err)
	}
	var report domain.Report
	if err := r.db.First(&report, "id = ?", id).Error; err == nil {
		t.Fatal("incompatible JSON was silently decoded")
	}
}

func TestCompleteWithCitationsRollsBackOnCitationFailure(t *testing.T) {
	r := integrationRepo(t)
	ctx := context.Background()
	if err := r.EnsureBuiltinProfiles(ctx); err != nil {
		t.Fatal(err)
	}
	repo := NewReportRepo(r.db)
	report := domain.Report{ID: uuid.NewString(), KBID: "kb", ProfileID: "builtin.common.knowledge-report", ProfileVersion: "1.0.0", ProfileSnapshot: domain.BuiltinReportProfiles()[0], CreatedBy: 7, InputValues: domain.ReportInputValues{"topic": {Value: "test"}}}
	if err := repo.CreatePending(ctx, &report); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkRunning(ctx, report.ID); err != nil {
		t.Fatal(err)
	}
	citations := []domain.ReportCitation{
		{CitationIndex: 1, NodeID: "node-a", DocumentName: "a", Locator: "a", Excerpt: "a"},
		{CitationIndex: 1, NodeID: "node-b", DocumentName: "b", Locator: "b", Excerpt: "b"},
	}
	if err := repo.CompleteWithCitations(ctx, report.ID, "must roll back", domain.ReportResultSuccess, citations); err == nil {
		t.Fatal("duplicate citations were accepted")
	}
	var stored domain.Report
	if err := r.db.First(&stored, "id = ?", report.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Status != domain.ReportStatusRunning || stored.Content != "" {
		t.Fatalf("completion was partly persisted: %#v", stored)
	}
}

func TestFailStaleRunningReportIntegration(t *testing.T) {
	r := integrationRepo(t)
	ctx := context.Background()
	if err := r.EnsureBuiltinProfiles(ctx); err != nil {
		t.Fatal(err)
	}
	repo := NewReportRepo(r.db)
	report := domain.Report{ID: uuid.NewString(), KBID: "kb", ProfileID: "builtin.common.knowledge-report", ProfileVersion: "1.0.0", ProfileSnapshot: domain.BuiltinReportProfiles()[0], CreatedBy: 8, InputValues: domain.ReportInputValues{"topic": {Value: "test"}}}
	if err := repo.CreatePending(ctx, &report); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkRunning(ctx, report.ID); err != nil {
		t.Fatal(err)
	}
	if err := r.db.Model(&domain.Report{}).Where("id = ?", report.ID).Update("updated_at", time.Now().UTC().Add(-3*time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	updated, err := repo.FailStaleRunning(ctx, time.Now().UTC().Add(-2*time.Minute))
	if err != nil || updated < 1 {
		t.Fatalf("stale report was not recovered: updated=%d err=%v", updated, err)
	}
	var loaded domain.Report
	if err := r.db.First(&loaded, "id = ?", report.ID).Error; err != nil {
		t.Fatal(err)
	}
	if loaded.Status != domain.ReportStatusFailed || loaded.ResultCode != domain.ReportResultTimeout {
		t.Fatalf("unexpected stale recovery result: %#v", loaded)
	}
}
