package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	"github.com/cloudwego/eino/schema"
)

const reportGenerationTimeout = 90 * time.Second
const reportMaxChunks = 12
const reportMaxPromptBytes = 48000
const reportStaleAfter = reportGenerationTimeout + 30*time.Second

var reportCitationPattern = regexp.MustCompile(`\[([0-9]+)\]`)

type ReportUsecase struct {
	profiles reportProfileStore
	reports  reportStore
	auth     reportAuthStore
	kbs      reportKnowledgeBaseStore
	llm      reportLLM
	models   reportModelProvider
	edition  reportEditionReader
	logger   *log.Logger
	active   sync.Map
}

type reportProfileStore interface {
	GetByID(context.Context, string) (*domain.ReportProfile, error)
	ListEnabledForEdition(context.Context, domain.EditionID) ([]domain.ReportProfile, error)
}
type reportStore interface {
	CreatePending(context.Context, *domain.Report) error
	MarkRunning(context.Context, string) error
	CompleteWithCitations(context.Context, string, string, domain.ReportResultCode, []domain.ReportCitation) error
	MarkFailed(context.Context, string, domain.ReportResultCode, string) error
	FailStaleRunning(context.Context, time.Time) (int64, error)
	ListByCreator(context.Context, uint) ([]domain.Report, error)
	GetByIDAndCreator(context.Context, string, uint) (*domain.Report, []domain.ReportCitation, error)
}
type reportAuthStore interface {
	GetAuthGroupIdsWithParentsByAuthId(context.Context, uint) ([]int, error)
}
type reportKnowledgeBaseStore interface {
	GetKnowledgeBaseByID(context.Context, string) (*domain.KnowledgeBase, error)
}
type reportLLM interface {
	GetRankNodes(context.Context, GetRankNodesRequest) (string, []*domain.RankedNodeChunks, error)
	GenerateWithConfiguredChatModel(context.Context, *domain.Model, []*schema.Message) (string, error)
}
type reportModelProvider interface {
	GetChatModel(context.Context) (*domain.Model, error)
}
type reportEditionReader interface {
	Get(context.Context) (*domain.EditionConfig, error)
}

func NewReportUsecase(profiles reportProfileStore, reports reportStore, auth reportAuthStore, kbs reportKnowledgeBaseStore, llm reportLLM, models reportModelProvider, edition reportEditionReader, logger *log.Logger) *ReportUsecase {
	return &ReportUsecase{profiles: profiles, reports: reports, auth: auth, kbs: kbs, llm: llm, models: models, edition: edition, logger: logger.WithModule("usecase.report")}
}

type ReportScope struct {
	NodeIDs []string `json:"node_ids"`
}
type CreateReportRequest struct {
	KBID        string                   `json:"kb_id" validate:"required"`
	ProfileID   string                   `json:"profile_id" validate:"required"`
	Title       string                   `json:"title"`
	InputValues domain.ReportInputValues `json:"input_values"`
	Scope       ReportScope              `json:"scope"`
}

type ReportProfileListItem struct {
	ID               string                           `json:"id"`
	Name             string                           `json:"name"`
	Description      string                           `json:"description"`
	InputFields      []domain.ReportProfileInputField `json:"input_fields"`
	Sections         []domain.ReportProfileSection    `json:"sections"`
	CitationRequired bool                             `json:"citation_required"`
	Version          string                           `json:"version"`
}

type ReportDetail struct {
	Report    ReportResponse          `json:"report"`
	Citations []domain.ReportCitation `json:"citations"`
}

// ReportResponse intentionally excludes ProfileSnapshot. Snapshots retain the
// internal system prompt for auditability and must not be returned to users.
type ReportResponse struct {
	ID             string                   `json:"id"`
	KBID           string                   `json:"kb_id"`
	ProfileID      string                   `json:"profile_id"`
	ProfileVersion string                   `json:"profile_version"`
	Status         domain.ReportStatus      `json:"status"`
	ResultCode     domain.ReportResultCode  `json:"result_code"`
	Title          string                   `json:"title"`
	InputValues    domain.ReportInputValues `json:"input_values"`
	Content        string                   `json:"content"`
	ErrorMessage   string                   `json:"error_message"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at"`
	CompletedAt    *time.Time               `json:"completed_at"`
	ProfileName    string                   `json:"profile_name"`
	CitationCount  int                      `json:"citation_count"`
}

func publicReport(report *domain.Report) ReportResponse {
	return ReportResponse{ID: report.ID, KBID: report.KBID, ProfileID: report.ProfileID, ProfileVersion: report.ProfileVersion, Status: report.Status, ResultCode: report.ResultCode, Title: report.Title, InputValues: report.InputValues, Content: report.Content, ErrorMessage: report.ErrorMessage, CreatedAt: report.CreatedAt, UpdatedAt: report.UpdatedAt, CompletedAt: report.CompletedAt, ProfileName: report.ProfileSnapshot.Name, CitationCount: report.CitationCount}
}

func (u *ReportUsecase) ListProfiles(ctx context.Context) ([]ReportProfileListItem, error) {
	edition, err := u.edition.Get(ctx)
	if err != nil {
		return nil, err
	}
	profiles, err := u.profiles.ListEnabledForEdition(ctx, edition.EditionID)
	if err != nil {
		return nil, err
	}
	items := make([]ReportProfileListItem, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, ReportProfileListItem{ID: profile.ID, Name: profile.Name, Description: profile.Description, InputFields: profile.InputFields, Sections: profile.Sections, CitationRequired: profile.CitationRequired, Version: profile.Version})
	}
	return items, nil
}

func (u *ReportUsecase) Create(ctx context.Context, authID uint, req CreateReportRequest) (*ReportDetail, error) {
	if authID == 0 {
		return nil, errors.New("authentication is required")
	}
	if strings.TrimSpace(req.KBID) == "" || strings.TrimSpace(req.ProfileID) == "" || len(req.Title) > 240 {
		return nil, errors.New("invalid report request")
	}
	if err := u.recoverStale(ctx); err != nil {
		return nil, err
	}
	profile, err := u.profiles.GetByID(ctx, req.ProfileID)
	if err != nil {
		return nil, err
	}
	edition, err := u.edition.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !profile.Enabled || !profileSupportsEdition(profile, edition.EditionID) {
		return nil, errors.New("report profile is unavailable")
	}
	if err := validateReportInputs(profile, req.InputValues); err != nil {
		return nil, err
	}
	kb, err := u.kbs.GetKnowledgeBaseByID(ctx, req.KBID)
	if err != nil {
		return nil, err
	}
	if !u.tryStartGeneration(authID) {
		return nil, errors.New("a report generation is already running for this user")
	}
	defer u.finishGeneration(authID)
	report := &domain.Report{KBID: req.KBID, ProfileID: profile.ID, ProfileVersion: profile.Version, ProfileSnapshot: *profile, CreatedBy: authID, Title: strings.TrimSpace(req.Title), InputValues: req.InputValues}
	if err := u.reports.CreatePending(ctx, report); err != nil {
		return nil, err
	}
	if err := u.reports.MarkRunning(ctx, report.ID); err != nil {
		return nil, err
	}

	// Generation deliberately detaches from HTTP cancellation after all request
	// validation and authentication have completed.
	generateCtx, cancel := context.WithTimeout(context.Background(), reportGenerationTimeout)
	defer cancel()
	startedAt := time.Now()
	groups, err := u.auth.GetAuthGroupIdsWithParentsByAuthId(generateCtx, authID)
	if err != nil {
		code, message := reportFailureForContext(generateCtx, err, domain.ReportResultRetrievalError, "failed to resolve authorized groups")
		return nil, u.fail(report, modelLogIdentity{}, code, message, 0, nil, time.Since(startedAt))
	}
	query := reportQuery(profile, req)
	_, ranked, err := u.llm.GetRankNodes(generateCtx, GetRankNodesRequest{DatasetID: kb.DatasetID, Question: query, GroupIDs: groups, SimilarityThreshold: 0.2, MaxChunksPerDoc: reportMaxChunks})
	if err != nil {
		code, message := reportFailureForContext(generateCtx, err, domain.ReportResultRetrievalError, "failed to retrieve authorized materials")
		return nil, u.fail(report, modelLogIdentity{}, code, message, 0, nil, time.Since(startedAt))
	}
	allowed, err := flattenReportChunks(ranked, req.Scope.NodeIDs)
	if err != nil {
		return nil, u.fail(report, modelLogIdentity{}, domain.ReportResultRetrievalError, "requested document scope is not authorized", 0, nil, time.Since(startedAt))
	}
	if len(allowed) == 0 && profile.CitationRequired {
		content := "Insufficient authorized materials were found to produce a cited report."
		if err := u.reports.CompleteWithCitations(generateCtx, report.ID, content, domain.ReportResultInsufficientEvidence, nil); err != nil {
			return nil, err
		}
		report.Status, report.ResultCode, report.Content = domain.ReportStatusCompleted, domain.ReportResultInsufficientEvidence, content
		u.logOutcome(report, modelLogIdentity{}, len(allowed), nil, time.Since(startedAt))
		return &ReportDetail{Report: publicReport(report)}, nil
	}
	model, err := u.models.GetChatModel(generateCtx)
	if err != nil {
		code, message := reportFailureForContext(generateCtx, err, domain.ReportResultModelError, "chat model is unavailable")
		return nil, u.fail(report, modelLogIdentity{}, code, message, len(allowed), nil, time.Since(startedAt))
	}
	answer, err := u.llm.GenerateWithConfiguredChatModel(generateCtx, model, []*schema.Message{schema.SystemMessage(profile.SystemPrompt), schema.UserMessage(buildReportPrompt(profile, req, allowed))})
	if err != nil {
		code, message := domain.ReportResultModelError, "report generation failed"
		if errors.Is(generateCtx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			code, message = domain.ReportResultTimeout, "report generation timed out"
		} else if isProviderRateLimit(err) {
			message = "model provider is temporarily rate limited"
		}
		return nil, u.fail(report, modelLogIdentity{provider: string(model.Provider), model: model.Model}, code, message, len(allowed), nil, time.Since(startedAt))
	}
	if strings.TrimSpace(answer) == "" {
		return nil, u.fail(report, modelLogIdentity{provider: string(model.Provider), model: model.Model}, domain.ReportResultModelError, "model provider returned an empty response", len(allowed), nil, time.Since(startedAt))
	}
	citations, err := citationsFromAnswer(answer, allowed)
	if err != nil {
		return nil, u.fail(report, modelLogIdentity{provider: string(model.Provider), model: model.Model}, domain.ReportResultInvalidCitation, "report contains an invalid citation", len(allowed), nil, time.Since(startedAt))
	}
	if profile.CitationRequired && len(citations) == 0 {
		return nil, u.fail(report, modelLogIdentity{provider: string(model.Provider), model: model.Model}, domain.ReportResultInvalidCitation, "report must cite authorized materials", len(allowed), nil, time.Since(startedAt))
	}
	if err := u.reports.CompleteWithCitations(generateCtx, report.ID, answer, domain.ReportResultSuccess, citations); err != nil {
		return nil, err
	}
	report.Status, report.ResultCode, report.Content = domain.ReportStatusCompleted, domain.ReportResultSuccess, answer
	u.logOutcome(report, modelLogIdentity{provider: string(model.Provider), model: model.Model}, len(allowed), citations, time.Since(startedAt))
	return &ReportDetail{Report: publicReport(report), Citations: citations}, nil
}

func (u *ReportUsecase) fail(report *domain.Report, model modelLogIdentity, code domain.ReportResultCode, publicMessage string, retrievedChunkCount int, citations []domain.ReportCitation, duration time.Duration) error {
	report.Status, report.ResultCode = domain.ReportStatusFailed, code
	u.logOutcome(report, model, retrievedChunkCount, citations, duration)
	if err := u.reports.MarkFailed(context.Background(), report.ID, code, publicMessage); err != nil {
		return err
	}
	// Keep cloud-provider errors out of handler logging. They can contain provider
	// response bodies, request metadata, or data-derived content.
	return errors.New(publicMessage)
}

func (u *ReportUsecase) ListMine(ctx context.Context, authID uint) ([]ReportResponse, error) {
	if err := u.recoverStale(ctx); err != nil {
		return nil, err
	}
	reports, err := u.reports.ListByCreator(ctx, authID)
	if err != nil {
		return nil, err
	}
	items := make([]ReportResponse, 0, len(reports))
	for i := range reports {
		items = append(items, publicReport(&reports[i]))
	}
	return items, nil
}
func (u *ReportUsecase) GetMine(ctx context.Context, authID uint, id string) (*ReportDetail, error) {
	if err := u.recoverStale(ctx); err != nil {
		return nil, err
	}
	report, citations, err := u.reports.GetByIDAndCreator(ctx, id, authID)
	if err != nil {
		return nil, err
	}
	return &ReportDetail{Report: publicReport(report), Citations: citations}, nil
}

type modelLogIdentity struct{ provider, model string }

func (u *ReportUsecase) logOutcome(report *domain.Report, model modelLogIdentity, retrievedChunkCount int, citations []domain.ReportCitation, duration time.Duration) {
	attrs := []any{
		log.String("report_id", report.ID),
		log.String("kb_id", report.KBID),
		log.String("profile_id", report.ProfileID),
		log.String("status", string(report.Status)),
		log.String("result_code", string(report.ResultCode)),
		log.Int64("duration_ms", duration.Milliseconds()),
		log.Int("retrieved_chunk_count", retrievedChunkCount),
		log.Int("citation_count", len(citations)),
	}
	if model.provider != "" {
		attrs = append(attrs, log.String("provider", model.provider), log.String("model", model.model))
	}
	if report.Status == domain.ReportStatusFailed {
		u.logger.Warn("report generation finished", attrs...)
		return
	}
	u.logger.Info("report generation finished", attrs...)
}

func (u *ReportUsecase) recoverStale(ctx context.Context) error {
	count, err := u.reports.FailStaleRunning(ctx, time.Now().UTC().Add(-reportStaleAfter))
	if err != nil {
		return err
	}
	if count > 0 {
		u.logger.Warn("stale reports recovered", log.Int64("count", count), log.String("result_code", string(domain.ReportResultTimeout)))
	}
	return nil
}

func isProviderRateLimit(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "429") || strings.Contains(message, "rate limit") || strings.Contains(message, "too many requests")
}

func reportFailureForContext(ctx context.Context, err error, defaultCode domain.ReportResultCode, defaultMessage string) (domain.ReportResultCode, string) {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return domain.ReportResultTimeout, "report generation timed out"
	}
	return defaultCode, defaultMessage
}

// tryStartGeneration is a deliberately small in-process concurrency guard for
// the synchronous report endpoint. Multi-instance deployments should also use
// their existing gateway or shared rate-limit policy.
func (u *ReportUsecase) tryStartGeneration(authID uint) bool {
	_, loaded := u.active.LoadOrStore(authID, struct{}{})
	return !loaded
}

func (u *ReportUsecase) finishGeneration(authID uint) {
	u.active.Delete(authID)
}

func profileSupportsEdition(profile *domain.ReportProfile, edition domain.EditionID) bool {
	for _, id := range profile.EditionIDs {
		if id == edition {
			return true
		}
	}
	return false
}
func validateReportInputs(profile *domain.ReportProfile, values domain.ReportInputValues) error {
	fields := make(map[string]domain.ReportProfileInputField, len(profile.InputFields))
	for _, field := range profile.InputFields {
		fields[field.Key] = field
	}
	for key, value := range values {
		field, ok := fields[key]
		if !ok {
			return fmt.Errorf("unknown report input field: %s", key)
		}
		if field.Type == domain.ReportInputTypeDateRange {
			if value.DateRange == nil || value.Value != "" {
				return fmt.Errorf("report input %s must be a date range", key)
			}
			continue
		}
		if value.DateRange != nil || len(value.Value) > 4000 {
			return fmt.Errorf("invalid report input %s", key)
		}
		if field.Type == domain.ReportInputTypeSelect && !contains(field.Options, value.Value) {
			return fmt.Errorf("invalid select value for %s", key)
		}
	}
	for _, field := range profile.InputFields {
		if field.Required {
			value, ok := values[field.Key]
			if !ok || (value.Value == "" && value.DateRange == nil) {
				return fmt.Errorf("required report input missing: %s", field.Key)
			}
		}
	}
	return nil
}
func contains(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
func reportQuery(profile *domain.ReportProfile, req CreateReportRequest) string {
	var b strings.Builder
	b.WriteString(profile.Name)
	b.WriteString("\n")
	for key, value := range req.InputValues {
		b.WriteString(key + ": ")
		if value.DateRange != nil {
			b.WriteString(value.DateRange.Start + " to " + value.DateRange.End)
		} else {
			b.WriteString(value.Value)
		}
		b.WriteString("\n")
	}
	return b.String()
}

type reportAllowedChunk struct{ NodeID, DocumentName, Locator, Excerpt string }

func flattenReportChunks(ranked []*domain.RankedNodeChunks, scope []string) ([]reportAllowedChunk, error) {
	scopeSet := make(map[string]bool, len(scope))
	for _, id := range scope {
		if strings.TrimSpace(id) == "" {
			return nil, errors.New("document scope contains an empty node id")
		}
		scopeSet[id] = true
	}
	result := make([]reportAllowedChunk, 0)
	authorizedNodes := make(map[string]bool, len(ranked))
	for _, node := range ranked {
		authorizedNodes[node.NodeID] = true
		if len(scopeSet) > 0 && !scopeSet[node.NodeID] {
			continue
		}
		for _, chunk := range node.Chunks {
			result = append(result, reportAllowedChunk{NodeID: node.NodeID, DocumentName: node.NodeName, Locator: chunk.ID, Excerpt: chunk.Content})
		}
	}
	for nodeID := range scopeSet {
		if !authorizedNodes[nodeID] {
			return nil, fmt.Errorf("requested node is outside authorized retrieval results")
		}
	}
	return result, nil
}
func buildReportPrompt(profile *domain.ReportProfile, req CreateReportRequest, chunks []reportAllowedChunk) string {
	var b strings.Builder
	b.WriteString("Create a structured report using only these authorized excerpts. Cite every factual claim using only [n] corresponding to the excerpt list. Do not invent citations.\n\n")
	for _, section := range profile.Sections {
		b.WriteString("Section: " + section.Title + "\n" + section.Instruction + "\n")
	}
	b.WriteString("\nInputs:\n" + reportQuery(profile, req) + "\nAuthorized excerpts:\n")
	for i, chunk := range chunks {
		if b.Len() > reportMaxPromptBytes {
			break
		}
		fmt.Fprintf(&b, "[%d] %s\n%s\n\n", i+1, chunk.DocumentName, chunk.Excerpt)
	}
	return b.String()
}
func citationsFromAnswer(answer string, allowed []reportAllowedChunk) ([]domain.ReportCitation, error) {
	matches := reportCitationPattern.FindAllStringSubmatch(answer, -1)
	seen := map[int]bool{}
	citations := make([]domain.ReportCitation, 0, len(matches))
	for _, match := range matches {
		var index int
		if _, err := fmt.Sscanf(match[1], "%d", &index); err != nil || index < 1 || index > len(allowed) {
			return nil, errors.New("citation index outside authorized evidence")
		}
		if seen[index] {
			continue
		}
		seen[index] = true
		chunk := allowed[index-1]
		citations = append(citations, domain.ReportCitation{CitationIndex: len(citations) + 1, NodeID: chunk.NodeID, DocumentName: chunk.DocumentName, Locator: chunk.Locator, Excerpt: chunk.Excerpt})
	}
	sort.SliceStable(citations, func(i, j int) bool { return citations[i].CitationIndex < citations[j].CitationIndex })
	return citations, nil
}
