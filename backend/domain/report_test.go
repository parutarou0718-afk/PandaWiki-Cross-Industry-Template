package domain

import "testing"

func TestBuiltinReportProfilesAreStableAndValid(t *testing.T) {
	profiles := BuiltinReportProfiles()
	if len(profiles) != 10 {
		t.Fatalf("got %d profiles, want 10", len(profiles))
	}
	seen := map[string]bool{}
	for _, profile := range profiles {
		if profile.ID == "" || seen[profile.ID] {
			t.Fatalf("unstable or duplicate id: %q", profile.ID)
		}
		seen[profile.ID] = true
		if !profile.IsBuiltin || profile.Validate() != nil {
			t.Fatalf("invalid builtin profile: %s", profile.ID)
		}
	}
}

func TestReportProfileValidation(t *testing.T) {
	profile := BuiltinReportProfiles()[0]
	profile.EditionIDs = []EditionID{"invalid"}
	if profile.Validate() == nil {
		t.Fatal("invalid edition accepted")
	}
	profile = BuiltinReportProfiles()[0]
	profile.InputFields = append(profile.InputFields, profile.InputFields[0])
	if profile.Validate() == nil {
		t.Fatal("duplicate input field accepted")
	}
	profile = BuiltinReportProfiles()[0]
	profile.Sections = append(profile.Sections, profile.Sections[0])
	if profile.Validate() == nil {
		t.Fatal("duplicate section accepted")
	}
	profile = BuiltinReportProfiles()[0]
	profile.InputFields = []ReportProfileInputField{{Key: "audience", Label: "Audience", Type: ReportInputTypeSelect}}
	if profile.Validate() == nil {
		t.Fatal("select without options accepted")
	}
}
