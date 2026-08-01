package v1

import (
	"strings"
	"testing"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/stretchr/testify/require"
)

func TestDecodePluginRecordWriteRejectsUnknownFieldsAndNonObjectPayload(t *testing.T) {
	valid := `{"kb_id":"kb-1","plugin_id":"official.submission-management","record_type":"submission","payload":{"paper_title":"Draft"},"access":{"visibility":"private","shared_group_ids":[],"allow_collaborative_edit":false}}`

	decoded, err := decodePluginRecordWrite(strings.NewReader(valid))
	require.NoError(t, err)
	require.Equal(t, domain.PluginRecordVisibilityPrivate, decoded.Access.Visibility)

	_, err = decodePluginRecordWrite(strings.NewReader(`{"kb_id":"kb-1","plugin_id":"official.submission-management","record_type":"submission","payload":[],"access":{"visibility":"private","shared_group_ids":[],"allow_collaborative_edit":false}}`))
	require.Error(t, err)

	_, err = decodePluginRecordWrite(strings.NewReader(valid[:len(valid)-1] + `,"owner_user_id":"attacker"}`))
	require.Error(t, err)
}

func TestDecodePluginRecordIdentityForRestoreDoesNotRequirePayload(t *testing.T) {
	identity, err := decodePluginRecordIdentity(strings.NewReader(`{"kb_id":"kb-1","plugin_id":"official.submission-management","record_type":"submission"}`))
	require.NoError(t, err)
	require.Equal(t, "kb-1", identity.KBID)

	_, err = decodePluginRecordIdentity(strings.NewReader(`{"kb_id":"kb-1","plugin_id":"official.submission-management","record_type":"submission","payload":{}}`))
	require.Error(t, err)
}
