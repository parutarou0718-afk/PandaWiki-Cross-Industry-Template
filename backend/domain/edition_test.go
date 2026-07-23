package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPublicEditionConfigAllowlist(t *testing.T) {
	c := &EditionConfig{EditionID: EditionLegal, EditionVersion: "1.0.0", ProductName: "Legal", DefaultPrompts: EditionPrompts{Chat: "secret"}}
	b, err := json.Marshal(c.Public())
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, forbidden := range []string{"prompt", "overrides", "updated_by", "schema_version"} {
		if strings.Contains(strings.ToLower(s), forbidden) {
			t.Fatalf("public response contains %q: %s", forbidden, s)
		}
	}
	for _, required := range []string{"edition_id", "edition_version", "product_name", "branding", "enabled_features", "document_types", "relation_types"} {
		if !strings.Contains(s, required) {
			t.Fatalf("public response missing %q: %s", required, s)
		}
	}
}

func strptr(s string) *string { return &s }
