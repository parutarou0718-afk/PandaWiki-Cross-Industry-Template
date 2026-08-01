package domain

import "testing"

func TestPluginRecordGroupValidate(t *testing.T) {
	group := PluginRecordGroup{ID: 1, KBID: "kb-1", Name: "Research team", MemberUserIDs: []string{"user-a", "user-b"}}
	if err := group.Validate(); err != nil {
		t.Fatalf("expected valid plugin group, got %v", err)
	}
}

func TestPluginRecordGroupRejectsDuplicateMembers(t *testing.T) {
	group := PluginRecordGroup{ID: 1, KBID: "kb-1", Name: "Research team", MemberUserIDs: []string{"user-a", "user-a"}}
	if err := group.Validate(); err == nil {
		t.Fatal("expected duplicate member validation failure")
	}
}
