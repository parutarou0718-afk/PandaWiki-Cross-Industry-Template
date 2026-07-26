package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveOpenAIAPIBotKnowledgeBaseUsesSingleAuthorizedKnowledgeBase(t *testing.T) {
	apps := []*App{{ID: "app-legal", KBID: "kb-legal", Type: AppTypeOpenAIAPI}}

	app, err := ResolveOpenAIAPIBotKnowledgeBase(apps, "")

	require.NoError(t, err)
	require.Equal(t, "kb-legal", app.KBID)
}

func TestResolveOpenAIAPIBotKnowledgeBaseRejectsUnauthorizedRequestedKnowledgeBase(t *testing.T) {
	apps := []*App{{ID: "app-legal", KBID: "kb-legal", Type: AppTypeOpenAIAPI}}

	_, err := ResolveOpenAIAPIBotKnowledgeBase(apps, "kb-finance")

	require.ErrorIs(t, err, ErrOpenAIAPIKnowledgeBaseNotAuthorized)
}

func TestResolveOpenAIAPIBotKnowledgeBaseRequiresSelectionForMultipleKnowledgeBases(t *testing.T) {
	apps := []*App{
		{ID: "app-legal", KBID: "kb-legal", Type: AppTypeOpenAIAPI},
		{ID: "app-finance", KBID: "kb-finance", Type: AppTypeOpenAIAPI},
	}

	_, err := ResolveOpenAIAPIBotKnowledgeBase(apps, "")

	require.True(t, errors.Is(err, ErrOpenAIAPIKnowledgeBaseSelectionRequired))
}

func TestResolveOpenAIAPIBotKnowledgeBaseUsesAuthorizedRequestedKnowledgeBase(t *testing.T) {
	apps := []*App{
		{ID: "app-legal", KBID: "kb-legal", Type: AppTypeOpenAIAPI},
		{ID: "app-finance", KBID: "kb-finance", Type: AppTypeOpenAIAPI},
	}

	app, err := ResolveOpenAIAPIBotKnowledgeBase(apps, "kb-finance")

	require.NoError(t, err)
	require.Equal(t, "kb-finance", app.KBID)
}

func TestResolveOpenAIAPIBotKnowledgeBaseRejectsUnknownToken(t *testing.T) {
	_, err := ResolveOpenAIAPIBotKnowledgeBase(nil, "")

	require.ErrorIs(t, err, ErrOpenAIAPITokenNotFound)
}
