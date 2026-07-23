package usecase

import (
	"testing"

	"github.com/chaitin/panda-wiki/domain"
)

func TestApplyAdminRequestKeepsBuiltinIdentity(t *testing.T) {
	profile := domain.BuiltinReportProfiles()[0]
	originalID, originalBuiltin := profile.ID, profile.IsBuiltin
	applyAdminRequest(&profile, &ReportProfileAdminRequest{
		Name: "Edited", EditionIDs: []domain.EditionID{domain.EditionLegal},
		SystemPrompt: "prompt", InputFields: []domain.ReportProfileInputField{{Key: "topic", Type: domain.ReportInputTypeText}},
		Sections: []domain.ReportProfileSection{{Key: "summary", Order: 1}}, Version: "2.0.0",
	})
	if profile.ID != originalID || profile.IsBuiltin != originalBuiltin {
		t.Fatal("admin editable fields changed server-owned builtin identity")
	}
}

func TestAdminRequestUsesProfileValidation(t *testing.T) {
	profile := profileFromAdminRequest(&ReportProfileAdminRequest{
		Name: "Template", EditionIDs: []domain.EditionID{domain.EditionCommon}, SystemPrompt: "prompt", Version: "1.0.0",
		InputFields: []domain.ReportProfileInputField{{Key: "topic", Type: domain.ReportInputTypeText}, {Key: "topic", Type: domain.ReportInputTypeText}},
		Sections:    []domain.ReportProfileSection{{Key: "summary", Order: 1}},
	})
	profile.ID = "new"
	if profile.Validate() == nil {
		t.Fatal("duplicate input field key was accepted")
	}
}
