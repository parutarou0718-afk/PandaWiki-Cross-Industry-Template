package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodePluginRecordGroupWriteRejectsUnknownFields(t *testing.T) {
	decoded, err := decodePluginRecordGroupWrite(strings.NewReader(`{"kb_id":"kb-1","name":"Research team","member_user_ids":["user-a"]}`))
	require.NoError(t, err)
	require.Equal(t, "Research team", decoded.Name)

	_, err = decodePluginRecordGroupWrite(strings.NewReader(`{"kb_id":"kb-1","name":"Research team","member_user_ids":["user-a"],"created_by":"attacker"}`))
	require.Error(t, err)
}
