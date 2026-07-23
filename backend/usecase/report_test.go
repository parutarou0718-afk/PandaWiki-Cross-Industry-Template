package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	"github.com/cloudwego/eino/schema"
)

func TestReportInputValuesAcceptsScalarAndDateRange(t *testing.T) {
	var values domain.ReportInputValues
	if err := json.Unmarshal([]byte(`{"topic":"security","date_range":{"start":"2020-01-01","end":"2026-01-01"}}`), &values); err != nil {
		t.Fatal(err)
	}
	if values["topic"].Value != "security" || values["date_range"].DateRange == nil {
		t.Fatal("report request values were not decoded")
	}
}

func TestCitationsKeepFirstAppearanceAndRejectUnknownIndex(t *testing.T) {
	allowed := []reportAllowedChunk{{NodeID: "node-1", DocumentName: "one", Locator: "chunk-1", Excerpt: "a"}, {NodeID: "node-2", DocumentName: "two", Locator: "chunk-2", Excerpt: "b"}}
	citations, err := citationsFromAnswer("Evidence [2], then [1], repeat [2].", allowed)
	if err != nil {
		t.Fatal(err)
	}
	if len(citations) != 2 || citations[0].NodeID != "node-2" || citations[1].NodeID != "node-1" {
		t.Fatalf("unexpected citation ordering: %#v", citations)
	}
	if _, err := citationsFromAnswer("not authorized [3]", allowed); err == nil {
		t.Fatal("out-of-range citation was accepted")
	}
}

func TestReportScopeCannotIntroduceUnauthorizedNode(t *testing.T) {
	ranked := []*domain.RankedNodeChunks{{NodeID: "node-1", NodeName: "doc", Chunks: []*domain.NodeContentChunk{{ID: "chunk-1", Content: "allowed"}}}}
	if _, err := flattenReportChunks(ranked, []string{"node-2"}); err == nil {
		t.Fatal("out-of-scope node id was accepted")
	}
}

type fakeReportProfiles struct{ profile *domain.ReportProfile }

func (f fakeReportProfiles) GetByID(context.Context, string) (*domain.ReportProfile, error) {
	return f.profile, nil
}
func (f fakeReportProfiles) ListEnabledForEdition(context.Context, domain.EditionID) ([]domain.ReportProfile, error) {
	return nil, nil
}

type fakeReports struct {
	report    *domain.Report
	completed domain.ReportResultCode
	failed    domain.ReportResultCode
	citations []domain.ReportCitation
}

func (f *fakeReports) CreatePending(_ context.Context, report *domain.Report) error {
	report.ID = "report-1"
	report.Status = domain.ReportStatusPending
	f.report = report
	return nil
}
func (f *fakeReports) MarkRunning(_ context.Context, _ string) error {
	f.report.Status = domain.ReportStatusRunning
	return nil
}
func (f *fakeReports) CompleteWithCitations(_ context.Context, _ string, content string, code domain.ReportResultCode, citations []domain.ReportCitation) error {
	f.report.Status = domain.ReportStatusCompleted
	f.report.ResultCode = code
	f.report.Content = content
	f.completed = code
	f.citations = citations
	return nil
}
func (f *fakeReports) MarkFailed(_ context.Context, _ string, code domain.ReportResultCode, _ string) error {
	f.report.Status = domain.ReportStatusFailed
	f.report.ResultCode = code
	f.failed = code
	return nil
}
func (f *fakeReports) FailStaleRunning(context.Context, time.Time) (int64, error)   { return 0, nil }
func (f *fakeReports) ListByCreator(context.Context, uint) ([]domain.Report, error) { return nil, nil }
func (f *fakeReports) GetByIDAndCreator(context.Context, string, uint) (*domain.Report, []domain.ReportCitation, error) {
	return nil, nil, errors.New("not found")
}

type fakeReportAuth struct {
	groups []int
	err    error
}

func (f fakeReportAuth) GetAuthGroupIdsWithParentsByAuthId(context.Context, uint) ([]int, error) {
	return f.groups, f.err
}

type fakeReportKB struct{}

func (fakeReportKB) GetKnowledgeBaseByID(context.Context, string) (*domain.KnowledgeBase, error) {
	return &domain.KnowledgeBase{ID: "kb", DatasetID: "dataset"}, nil
}

type fakeReportLLM struct {
	ranked  []*domain.RankedNodeChunks
	rankReq GetRankNodesRequest
	answer  string
	err     error
}

func (f *fakeReportLLM) GetRankNodes(_ context.Context, req GetRankNodesRequest) (string, []*domain.RankedNodeChunks, error) {
	f.rankReq = req
	return "", f.ranked, nil
}
func (f *fakeReportLLM) GenerateWithConfiguredChatModel(context.Context, *domain.Model, []*schema.Message) (string, error) {
	return f.answer, f.err
}

type fakeReportModel struct{}

func (fakeReportModel) GetChatModel(context.Context) (*domain.Model, error) {
	return &domain.Model{}, nil
}

type fakeReportEdition struct{}

func (fakeReportEdition) Get(context.Context) (*domain.EditionConfig, error) {
	return &domain.EditionConfig{EditionID: domain.EditionCommon}, nil
}
func newReportUsecaseForTest(profile *domain.ReportProfile, reports *fakeReports, llm *fakeReportLLM) *ReportUsecase {
	return NewReportUsecase(fakeReportProfiles{profile}, reports, fakeReportAuth{groups: []int{11, 12}}, fakeReportKB{}, llm, fakeReportModel{}, fakeReportEdition{}, &log.Logger{Logger: slog.Default()})
}

func TestCreateReportUsesAuthorizedGroupIDsAndCompletesInsufficientEvidence(t *testing.T) {
	profile := domain.BuiltinReportProfiles()[0]
	reports, llm := &fakeReports{}, &fakeReportLLM{}
	result, err := newReportUsecaseForTest(&profile, reports, llm).Create(context.Background(), 42, CreateReportRequest{KBID: "kb", ProfileID: profile.ID, InputValues: domain.ReportInputValues{"topic": {Value: "test"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(llm.rankReq.GroupIDs) != 2 || llm.rankReq.GroupIDs[0] != 11 {
		t.Fatalf("authorized group IDs were not passed to RAG: %#v", llm.rankReq)
	}
	if reports.completed != domain.ReportResultInsufficientEvidence || result.Report.Status != domain.ReportStatusCompleted {
		t.Fatalf("insufficient evidence must complete the report: %#v", result)
	}
}

func TestCreateReportRejectsModelCitationOutsideAuthorizedEvidence(t *testing.T) {
	profile := domain.BuiltinReportProfiles()[0]
	reports := &fakeReports{}
	llm := &fakeReportLLM{ranked: []*domain.RankedNodeChunks{{NodeID: "node-1", NodeName: "doc", Chunks: []*domain.NodeContentChunk{{ID: "chunk-1", Content: "authorized"}}}}, answer: "fabricated [2]"}
	_, err := newReportUsecaseForTest(&profile, reports, llm).Create(context.Background(), 42, CreateReportRequest{KBID: "kb", ProfileID: profile.ID, InputValues: domain.ReportInputValues{"topic": {Value: "test"}}})
	if err == nil || reports.failed != domain.ReportResultInvalidCitation || reports.report.Status != domain.ReportStatusFailed {
		t.Fatalf("invalid citation did not fail safely: err=%v report=%#v", err, reports.report)
	}
}

func TestCreateReportMarksTimeout(t *testing.T) {
	profile := domain.BuiltinReportProfiles()[0]
	reports := &fakeReports{}
	llm := &fakeReportLLM{ranked: []*domain.RankedNodeChunks{{NodeID: "node-1", NodeName: "doc", Chunks: []*domain.NodeContentChunk{{ID: "chunk-1", Content: "authorized"}}}}, err: context.DeadlineExceeded}
	_, err := newReportUsecaseForTest(&profile, reports, llm).Create(context.Background(), 42, CreateReportRequest{KBID: "kb", ProfileID: profile.ID, InputValues: domain.ReportInputValues{"topic": {Value: "test"}}})
	if err == nil || reports.failed != domain.ReportResultTimeout {
		t.Fatalf("timeout did not mark report failed with timeout: %v", err)
	}
}

func TestCreateReportMarksTimeoutWhenAuthorizationResolutionExpires(t *testing.T) {
	profile := domain.BuiltinReportProfiles()[0]
	reports, llm := &fakeReports{}, &fakeReportLLM{}
	uc := NewReportUsecase(fakeReportProfiles{&profile}, reports, fakeReportAuth{err: context.DeadlineExceeded}, fakeReportKB{}, llm, fakeReportModel{}, fakeReportEdition{}, &log.Logger{Logger: slog.Default()})
	_, err := uc.Create(context.Background(), 42, CreateReportRequest{KBID: "kb", ProfileID: profile.ID, InputValues: domain.ReportInputValues{"topic": {Value: "test"}}})
	if err == nil || reports.failed != domain.ReportResultTimeout {
		t.Fatalf("authorization deadline did not mark report as timeout: %v", err)
	}
}

func TestCreateReportRejectsEmptyModelResponse(t *testing.T) {
	profile := domain.BuiltinReportProfiles()[0]
	profile.CitationRequired = false
	reports := &fakeReports{}
	llm := &fakeReportLLM{ranked: []*domain.RankedNodeChunks{{NodeID: "node-1", NodeName: "doc", Chunks: []*domain.NodeContentChunk{{ID: "chunk-1", Content: "authorized"}}}}, answer: "   "}
	_, err := newReportUsecaseForTest(&profile, reports, llm).Create(context.Background(), 42, CreateReportRequest{KBID: "kb", ProfileID: profile.ID, InputValues: domain.ReportInputValues{"topic": {Value: "test"}}})
	if err == nil || reports.failed != domain.ReportResultModelError {
		t.Fatalf("empty model response did not fail safely: err=%v report=%#v", err, reports.report)
	}
}

func TestCreateReportSanitizesProviderRateLimitError(t *testing.T) {
	profile := domain.BuiltinReportProfiles()[0]
	reports := &fakeReports{}
	providerError := errors.New("provider response 429: request included confidential excerpt")
	llm := &fakeReportLLM{ranked: []*domain.RankedNodeChunks{{NodeID: "node-1", NodeName: "doc", Chunks: []*domain.NodeContentChunk{{ID: "chunk-1", Content: "authorized"}}}}, err: providerError}
	_, err := newReportUsecaseForTest(&profile, reports, llm).Create(context.Background(), 42, CreateReportRequest{KBID: "kb", ProfileID: profile.ID, InputValues: domain.ReportInputValues{"topic": {Value: "test"}}})
	if err == nil || reports.failed != domain.ReportResultModelError {
		t.Fatalf("rate limit did not become a safe model failure: err=%v report=%#v", err, reports.report)
	}
	if strings.Contains(err.Error(), "confidential excerpt") || err.Error() != "model provider is temporarily rate limited" {
		t.Fatalf("provider error leaked through report API path: %v", err)
	}
}

func TestCreateReportLimitsOneSynchronousGenerationPerUser(t *testing.T) {
	profile := domain.BuiltinReportProfiles()[0]
	reports, llm := &fakeReports{}, &fakeReportLLM{}
	uc := newReportUsecaseForTest(&profile, reports, llm)
	if !uc.tryStartGeneration(42) {
		t.Fatal("failed to reserve the first generation slot")
	}
	defer uc.finishGeneration(42)
	_, err := uc.Create(context.Background(), 42, CreateReportRequest{KBID: "kb", ProfileID: profile.ID, InputValues: domain.ReportInputValues{"topic": {Value: "test"}}})
	if err == nil || err.Error() != "a report generation is already running for this user" {
		t.Fatalf("concurrent report request was not rejected: %v", err)
	}
}

func TestReportInputValidationRejectsUnknownAndInvalidSelect(t *testing.T) {
	profile := domain.BuiltinReportProfiles()[0]
	if err := validateReportInputs(&profile, domain.ReportInputValues{"unknown": {Value: "x"}}); err == nil {
		t.Fatal("unknown input field was accepted")
	}
	profile.InputFields = []domain.ReportProfileInputField{{Key: "kind", Type: domain.ReportInputTypeSelect, Required: true, Options: []string{"a"}}}
	if err := validateReportInputs(&profile, domain.ReportInputValues{"kind": {Value: "b"}}); err == nil {
		t.Fatal("invalid select value was accepted")
	}
}
