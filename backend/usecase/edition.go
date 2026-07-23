package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
)

const editionSchemaVersion = 1

type EditionUsecase struct {
	repo   EditionStore
	logger *log.Logger
}

type EditionStore interface {
	GetStoredEditionConfig(context.Context) (*domain.StoredEditionConfig, error)
	UpdateStoredEditionConfig(context.Context, *domain.StoredEditionConfig) error
}

func NewEditionUsecase(repo EditionStore, logger *log.Logger) *EditionUsecase {
	return &EditionUsecase{repo: repo, logger: logger.WithModule("usecase.edition")}
}

func (u *EditionUsecase) Get(ctx context.Context) (*domain.EditionConfig, error) {
	stored, err := u.repo.GetStoredEditionConfig(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CommonEditionConfig(), nil
		}
		u.logger.Warn("invalid edition config, use common", log.Error(err))
		return CommonEditionConfig(), nil
	}
	config, err := resolveEdition(*stored)
	if err != nil {
		u.logger.Warn("invalid stored edition config, use common", log.Error(err))
		return CommonEditionConfig(), nil
	}
	if stored.EditionVersion != config.EditionVersion {
		u.logger.Info("edition config migrated to current preset", log.String("edition", string(stored.EditionID)), log.String("from_version", stored.EditionVersion), log.String("to_version", config.EditionVersion))
	}
	return config, nil
}

func (u *EditionUsecase) Update(ctx context.Context, userID uint, req *domain.UpdateEditionReq) (*domain.EditionConfig, error) {
	if err := validateEditionRequest(req); err != nil {
		return nil, err
	}
	base := preset(req.EditionID)
	resolved, err := resolveEdition(domain.StoredEditionConfig{
		SchemaVersion:  editionSchemaVersion,
		EditionID:      req.EditionID,
		EditionVersion: base.EditionVersion,
		Overrides:      req.Overrides,
	})
	if err != nil {
		return nil, err
	}
	previous, _ := u.repo.GetStoredEditionConfig(ctx)
	stored := &domain.StoredEditionConfig{
		SchemaVersion:  editionSchemaVersion,
		EditionID:      req.EditionID,
		EditionVersion: resolved.EditionVersion,
		Overrides:      req.Overrides,
		UpdatedAt:      time.Now().UTC(),
		UpdatedBy:      userID,
	}
	if err := u.repo.UpdateStoredEditionConfig(ctx, stored); err != nil {
		u.logger.Error("update edition config failed", log.Error(err), log.Int64("updated_by", int64(userID)))
		return nil, err
	}
	if previous == nil {
		u.logger.Info("edition config updated", log.Int64("updated_by", int64(userID)), log.String("edition", string(req.EditionID)), log.Any("fields", overrideFields(req.Overrides)))
	} else {
		u.logger.Info("edition config updated", log.Int64("updated_by", int64(userID)), log.String("before", string(previous.EditionID)), log.String("after", string(req.EditionID)), log.String("before_version", previous.EditionVersion), log.String("after_version", resolved.EditionVersion), log.Any("fields", overrideFields(req.Overrides)))
	}
	return resolved, nil
}

func CommonEditionConfig() *domain.EditionConfig { return preset(domain.EditionCommon) }

func preset(id domain.EditionID) *domain.EditionConfig {
	base := &domain.EditionConfig{
		SchemaVersion:   editionSchemaVersion,
		EditionID:       id,
		EditionVersion:  "1.0.0",
		Branding:        domain.EditionBranding{},
		Terminology:     map[string]string{"knowledge_base": "Knowledge Base", "document": "Document", "folder": "Folder", "report": "Report"},
		EnabledFeatures: []string{"chat", "search"},
		RelationTypes:   []string{"reference", "related"},
		DefaultPrompts:  domain.EditionPrompts{},
	}
	switch id {
	case domain.EditionResearch:
		base.ProductName, base.ShortName, base.HomeDescription = "Research Knowledge Base", "Research Wiki", "Research documents and project materials"
		base.DocumentTypes = []string{"Literature", "Research Materials", "Research Projects", "Literature Reviews", "Research Reports", "Research Progress"}
	case domain.EditionLegal:
		base.ProductName, base.ShortName, base.HomeDescription = "Legal Knowledge Base", "Legal Wiki", "Internal legal materials and reports"
		base.DocumentTypes = []string{"Case Materials", "Case Files", "Contracts", "Evidence", "Statutes", "Precedents"}
		base.Terminology = map[string]string{"knowledge_base": "Case Materials", "document": "Case File", "folder": "Case Folder", "report": "Legal Report"}
	case domain.EditionFinance:
		base.ProductName, base.ShortName, base.HomeDescription = "Finance Knowledge Base", "Finance Wiki", "Company, financial and industry analysis materials"
		base.DocumentTypes = []string{"Company Materials", "Financial Reports", "Announcements", "Industry Materials", "Company Analysis", "Industry Analysis", "Risk Summary"}
		base.Terminology = map[string]string{"knowledge_base": "Materials Base", "document": "Material", "folder": "Materials Folder", "report": "Analysis Report"}
	default:
		base.EditionID, base.ProductName, base.ShortName, base.HomeDescription = domain.EditionCommon, "PandaWiki", "PandaWiki", "AI-powered knowledge base"
		base.DocumentTypes = []string{"Product Documents", "Technical Documents", "FAQ", "Blog"}
	}
	return base
}

func validateEditionRequest(req *domain.UpdateEditionReq) error {
	if req == nil {
		return errors.New("edition request is required")
	}
	if !isKnownEdition(req.EditionID) {
		return fmt.Errorf("invalid edition_id: %s", req.EditionID)
	}
	if req.Overrides.ProductName != nil && !validText(*req.Overrides.ProductName, 1, 120) {
		return errors.New("product_name must be 1-120 characters")
	}
	if req.Overrides.ShortName != nil && !validText(*req.Overrides.ShortName, 1, 80) {
		return errors.New("short_name must be 1-80 characters")
	}
	if req.Overrides.HomeDescription != nil && !validText(*req.Overrides.HomeDescription, 1, 500) {
		return errors.New("home_description must be 1-500 characters")
	}
	if req.Overrides.Branding != nil {
		logo := strings.TrimSpace(req.Overrides.Branding.Logo)
		if logo != "" && len(logo) > 500 {
			return errors.New("branding.logo must be at most 500 characters")
		}
	}
	for key, value := range req.Overrides.Terminology {
		if !allowedTerminology[key] || !validText(value, 1, 80) {
			return errors.New("terminology keys and values cannot be empty")
		}
	}
	for _, values := range [][]string{req.Overrides.EnabledFeatures, req.Overrides.DocumentTypes, req.Overrides.RelationTypes} {
		for _, value := range values {
			if !validText(value, 1, 80) {
				return errors.New("array values cannot be empty")
			}
		}
	}
	for _, feature := range req.Overrides.EnabledFeatures {
		if !allowedFeatures[feature] {
			return fmt.Errorf("unsupported feature: %s", feature)
		}
	}
	if req.Overrides.DefaultPrompts != nil {
		if len(req.Overrides.DefaultPrompts.Chat) > 16000 || len(req.Overrides.DefaultPrompts.Summary) > 16000 {
			return errors.New("default prompts must be at most 16000 characters")
		}
	}
	return nil
}

func resolveEdition(stored domain.StoredEditionConfig) (*domain.EditionConfig, error) {
	if stored.SchemaVersion != editionSchemaVersion || !isKnownEdition(stored.EditionID) {
		return nil, errors.New("unsupported stored edition config")
	}
	base := preset(stored.EditionID)
	o := stored.Overrides
	if o.ProductName != nil {
		base.ProductName = *o.ProductName
	}
	if o.ShortName != nil {
		base.ShortName = *o.ShortName
	}
	if o.Branding != nil {
		base.Branding = *o.Branding
	}
	if o.HomeDescription != nil {
		base.HomeDescription = *o.HomeDescription
	}
	for key, value := range o.Terminology {
		base.Terminology[key] = value
	}
	if o.EnabledFeatures != nil {
		base.EnabledFeatures = o.EnabledFeatures
	}
	if o.DocumentTypes != nil {
		base.DocumentTypes = o.DocumentTypes
	}
	if o.RelationTypes != nil {
		base.RelationTypes = o.RelationTypes
	}
	if o.DefaultPrompts != nil {
		base.DefaultPrompts = *o.DefaultPrompts
	}
	if err := validateResolved(base); err != nil {
		return nil, err
	}
	return base, nil
}

func validateResolved(config *domain.EditionConfig) error {
	if config.SchemaVersion != editionSchemaVersion || !isKnownEdition(config.EditionID) || strings.TrimSpace(config.EditionVersion) == "" || strings.TrimSpace(config.ProductName) == "" || strings.TrimSpace(config.ShortName) == "" || strings.TrimSpace(config.HomeDescription) == "" {
		return errors.New("invalid resolved edition config")
	}
	if len(config.DocumentTypes) == 0 || len(config.EnabledFeatures) == 0 {
		return errors.New("resolved edition config requires features and document types")
	}
	return nil
}

func isKnownEdition(id domain.EditionID) bool {
	return id == domain.EditionCommon || id == domain.EditionResearch || id == domain.EditionLegal || id == domain.EditionFinance
}

var allowedTerminology = map[string]bool{"knowledge_base": true, "document": true, "folder": true, "report": true}
var allowedFeatures = map[string]bool{"chat": true, "search": true}

func validText(value string, min, max int) bool {
	value = strings.TrimSpace(value)
	return len(value) >= min && len(value) <= max
}

func overrideFields(overrides domain.EditionOverrides) []string {
	fields := make([]string, 0, 8)
	if overrides.ProductName != nil {
		fields = append(fields, "product_name")
	}
	if overrides.ShortName != nil {
		fields = append(fields, "short_name")
	}
	if overrides.Branding != nil {
		fields = append(fields, "branding")
	}
	if overrides.HomeDescription != nil {
		fields = append(fields, "home_description")
	}
	if overrides.Terminology != nil {
		fields = append(fields, "terminology")
	}
	if overrides.EnabledFeatures != nil {
		fields = append(fields, "enabled_features")
	}
	if overrides.DocumentTypes != nil {
		fields = append(fields, "document_types")
	}
	if overrides.RelationTypes != nil {
		fields = append(fields, "relation_types")
	}
	if overrides.DefaultPrompts != nil {
		fields = append(fields, "default_prompts")
	}
	return fields
}
