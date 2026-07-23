package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/repo/pg"
)

// ReportProfileAdminRequest contains only administrator-editable template
// fields. Stable IDs and persistence/audit fields are always server-owned.
type ReportProfileAdminRequest struct {
	Name             string                           `json:"name"`
	Description      string                           `json:"description"`
	EditionIDs       []domain.EditionID               `json:"edition_ids"`
	SystemPrompt     string                           `json:"system_prompt"`
	InputFields      []domain.ReportProfileInputField `json:"input_fields"`
	Sections         []domain.ReportProfileSection    `json:"sections"`
	CitationRequired bool                             `json:"citation_required"`
	Enabled          bool                             `json:"enabled"`
	Version          string                           `json:"version"`
}

type ReportProfileAdminUsecase struct{ repo *pg.ReportProfileRepo }

func NewReportProfileAdminUsecase(repo *pg.ReportProfileRepo) *ReportProfileAdminUsecase {
	return &ReportProfileAdminUsecase{repo: repo}
}

func (u *ReportProfileAdminUsecase) List(ctx context.Context) ([]domain.ReportProfile, error) {
	return u.repo.List(ctx)
}
func (u *ReportProfileAdminUsecase) Get(ctx context.Context, id string) (*domain.ReportProfile, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *ReportProfileAdminUsecase) Create(ctx context.Context, req *ReportProfileAdminRequest) (*domain.ReportProfile, error) {
	profile := profileFromAdminRequest(req)
	if err := profile.Validate(); err != nil {
		return nil, err
	}
	if err := u.repo.Create(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (u *ReportProfileAdminUsecase) Update(ctx context.Context, id string, req *ReportProfileAdminRequest) (*domain.ReportProfile, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("report profile id is required")
	}
	profile, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	applyAdminRequest(profile, req)
	if err := profile.Validate(); err != nil {
		return nil, err
	}
	if err := u.repo.Update(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (u *ReportProfileAdminUsecase) SetEnabled(ctx context.Context, id string, enabled bool) error {
	return u.repo.SetEnabled(ctx, id, enabled)
}
func (u *ReportProfileAdminUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.SoftDelete(ctx, id)
}
func (u *ReportProfileAdminUsecase) RestoreBuiltinDefault(ctx context.Context, id string) (*domain.ReportProfile, error) {
	if err := u.repo.RestoreBuiltinDefault(ctx, id); err != nil {
		return nil, err
	}
	return u.repo.GetByID(ctx, id)
}

func profileFromAdminRequest(req *ReportProfileAdminRequest) *domain.ReportProfile {
	p := &domain.ReportProfile{}
	applyAdminRequest(p, req)
	return p
}
func applyAdminRequest(profile *domain.ReportProfile, req *ReportProfileAdminRequest) {
	profile.Name, profile.Description = strings.TrimSpace(req.Name), strings.TrimSpace(req.Description)
	profile.EditionIDs, profile.SystemPrompt, profile.InputFields, profile.Sections = req.EditionIDs, req.SystemPrompt, req.InputFields, req.Sections
	profile.CitationRequired, profile.Enabled, profile.Version = req.CitationRequired, req.Enabled, strings.TrimSpace(req.Version)
}
