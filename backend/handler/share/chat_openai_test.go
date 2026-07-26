package share

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/chaitin/panda-wiki/domain"
)

func TestOpenAIAPIResolutionErrorStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"invalid token", domain.ErrOpenAIAPITokenNotFound, http.StatusUnauthorized},
		{"unauthorized knowledge base", domain.ErrOpenAIAPIKnowledgeBaseNotAuthorized, http.StatusForbidden},
		{"multiple knowledge bases need selection", domain.ErrOpenAIAPIKnowledgeBaseSelectionRequired, http.StatusBadRequest},
		{"unexpected failure", errors.New("database unavailable"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, openAIAPIResolutionErrorStatus(tt.err))
		})
	}
}

func TestHandleOpenAIStreamResponseWritesSSEAndDoneSentinel(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/share/v1/chat/completions", nil)
	c := e.NewContext(req, rec)
	events := make(chan domain.SSEEvent, 2)
	events <- domain.SSEEvent{Type: "data", Content: "hello"}
	events <- domain.SSEEvent{Type: "done"}
	close(events)

	err := (&ShareChatHandler{}).handleOpenAIStreamResponse(c, events, "knowledge-base")

	require.NoError(t, err)
	require.Contains(t, rec.Body.String(), `"object":"chat.completion.chunk"`)
	require.Contains(t, rec.Body.String(), "data: [DONE]\n\n")
}
