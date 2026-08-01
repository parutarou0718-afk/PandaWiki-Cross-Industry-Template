package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPluginRecordAccessValidation(t *testing.T) {
	tests := []struct {
		name    string
		access  PluginRecordAccess
		wantErr bool
	}{
		{name: "private record has no groups", access: PluginRecordAccess{Visibility: PluginRecordVisibilityPrivate}},
		{name: "knowledge base record has no groups", access: PluginRecordAccess{Visibility: PluginRecordVisibilityKnowledgeBase}},
		{name: "group record requires groups", access: PluginRecordAccess{Visibility: PluginRecordVisibilityGroups}, wantErr: true},
		{name: "private record cannot include groups", access: PluginRecordAccess{Visibility: PluginRecordVisibilityPrivate, SharedGroupIDs: []int64{7}}, wantErr: true},
		{name: "group record permits groups", access: PluginRecordAccess{Visibility: PluginRecordVisibilityGroups, SharedGroupIDs: []int64{7, 9}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.access.Validate()
			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPluginRecordPayloadRejectsNonObjectJSON(t *testing.T) {
	record := PluginRecord{
		ID:          "record-1",
		KBID:        "kb-1",
		PluginID:    "official.submission-management",
		RecordType:  "submission",
		OwnerUserID: "user-1",
		Payload:     PluginRecordPayload(`[]`),
		Access:      PluginRecordAccess{Visibility: PluginRecordVisibilityPrivate},
	}

	require.Error(t, record.Validate())
}

func TestPluginRecordVisibilityAndEditPolicy(t *testing.T) {
	private := PluginRecord{OwnerUserID: "user-1", Access: PluginRecordAccess{Visibility: PluginRecordVisibilityPrivate}}
	groups := PluginRecord{OwnerUserID: "user-1", Access: PluginRecordAccess{Visibility: PluginRecordVisibilityGroups, SharedGroupIDs: []int64{9}, AllowCollaborativeEdit: true}}
	kb := PluginRecord{OwnerUserID: "user-1", Access: PluginRecordAccess{Visibility: PluginRecordVisibilityKnowledgeBase}}

	require.True(t, private.IsVisibleTo("user-1", nil))
	require.False(t, private.IsVisibleTo("user-2", []int{9}))
	require.True(t, groups.IsVisibleTo("user-2", []int{9}))
	require.False(t, groups.IsVisibleTo("user-2", []int{10}))
	require.True(t, kb.IsVisibleTo("user-2", nil))
	require.False(t, private.CanEdit("user-2", false))
	require.True(t, groups.CanEdit("user-2", false, []int{9}))
	require.True(t, private.CanEdit("user-2", true))
}
