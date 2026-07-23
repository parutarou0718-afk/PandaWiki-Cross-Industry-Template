package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageContent_UnmarshalJSON_String(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected string
	}{
		{"simple string", `"hello"`, "hello"},
		{"with quotes", `"say \"hello\""`, `say "hello"`},
		{"with newline", `"line1\nline2"`, "line1\nline2"},
		{"empty string", `""`, ""},
		{"unicode", `"你好 🌍"`, "你好 🌍"},
		{"special chars", `"Hello \"World\"\nNew Line\tTab"`, "Hello \"World\"\nNew Line\tTab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mc MessageContent
			err := json.Unmarshal([]byte(tt.json), &mc)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, mc.String())
			assert.True(t, mc.isString)
		})
	}
}

func TestMessageContent_UnmarshalJSON_Array(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected string
	}{
		{
			"single text part",
			`[{"type":"text","text":"Hello"}]`,
			"Hello",
		},
		{
			"multiple text parts",
			`[{"type":"text","text":"Hello"},{"type":"text","text":"World"}]`,
			"Hello World",
		},
		{
			"mixed types with image",
			`[{"type":"text","text":"Look at this"},{"type":"image_url","image_url":{"url":"https://example.com/img.png"}},{"type":"text","text":"image"}]`,
			"Look at this image",
		},
		{
			"empty array",
			`[]`,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mc MessageContent
			err := json.Unmarshal([]byte(tt.json), &mc)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, mc.String())
			assert.False(t, mc.isString)
		})
	}
}

func TestMessageContent_UnmarshalJSON_Invalid(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"number", `123`},
		{"boolean", `true`},
		{"object", `{"key":"value"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mc MessageContent
			err := json.Unmarshal([]byte(tt.json), &mc)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "content must be string or array")
		})
	}
}

func TestMessageContent_UnmarshalJSON_Null(t *testing.T) {
	var mc *MessageContent
	err := json.Unmarshal([]byte(`null`), &mc)
	assert.NoError(t, err)
	assert.Nil(t, mc)
}

func TestMessageContent_MarshalJSON_String(t *testing.T) {
	mc := NewStringContent("Hello World")
	data, err := json.Marshal(mc)
	require.NoError(t, err)
	assert.Equal(t, `"Hello World"`, string(data))
}

func TestMessageContent_MarshalJSON_Array(t *testing.T) {
	mc := NewArrayContent([]OpenAIContentPart{
		{Type: "text", Text: "Hello"},
		{Type: "text", Text: "World"},
	})
	data, err := json.Marshal(mc)
	require.NoError(t, err)
	assert.JSONEq(t, `[{"type":"text","text":"Hello"},{"type":"text","text":"World"}]`, string(data))
}

func TestMessageContent_Roundtrip_String(t *testing.T) {
	original := NewStringContent("Test message with \"quotes\" and \nnewlines")

	// Marshal
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal
	var decoded MessageContent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, original.String(), decoded.String())
	assert.Equal(t, original.isString, decoded.isString)
}

func TestMessageContent_Roundtrip_Array(t *testing.T) {
	parts := []OpenAIContentPart{
		{Type: "text", Text: "Part 1"},
		{Type: "text", Text: "Part 2"},
	}
	original := NewArrayContent(parts)

	// Marshal
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal
	var decoded MessageContent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, original.String(), decoded.String())
	assert.Equal(t, original.isString, decoded.isString)
}

func TestNewStringContent(t *testing.T) {
	mc := NewStringContent("test")
	assert.NotNil(t, mc)
	assert.True(t, mc.isString)
	assert.Equal(t, "test", mc.strValue)
	assert.Equal(t, "test", mc.String())
}

func TestNewArrayContent(t *testing.T) {
	parts := []OpenAIContentPart{
		{Type: "text", Text: "Hello"},
	}
	mc := NewArrayContent(parts)
	assert.NotNil(t, mc)
	assert.False(t, mc.isString)
	assert.Equal(t, parts, mc.arrValue)
	assert.Equal(t, "Hello", mc.String())
}

func TestMessageContent_String_EmptyArray(t *testing.T) {
	mc := NewArrayContent([]OpenAIContentPart{})
	assert.Equal(t, "", mc.String())
}

func TestMessageContent_String_NoTextParts(t *testing.T) {
	mc := NewArrayContent([]OpenAIContentPart{
		{Type: "image_url", Text: ""},
	})
	assert.Equal(t, "", mc.String())
}
