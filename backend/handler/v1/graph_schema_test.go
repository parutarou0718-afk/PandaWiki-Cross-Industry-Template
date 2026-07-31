package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeKnowledgeSchemaRequestRejectsUnknownFields(t *testing.T) {
	_, err := decodeKnowledgeSchemaRequest(strings.NewReader(`{
    "kb_id": "kb-1",
    "schema": {"version": 1, "fields": [], "navigation": []},
    "group_ids": [1]
  }`))

	require.Error(t, err)
}

func TestDecodeGraphRebuildRequestReadsKBIDFromJSONBody(t *testing.T) {
	req, err := decodeGraphRebuildRequest(strings.NewReader(`{"kb_id":"kb-1"}`))

	require.NoError(t, err)
	require.Equal(t, "kb-1", req.KBID)
}
