package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/chaitin/panda-wiki/config"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
)

type promptStoreFake struct {
	chat, summary string
	err           error
}

func (f *promptStoreFake) GetPromptContent(context.Context, string) (string, error) {
	return f.chat, f.err
}
func (f *promptStoreFake) GetSummaryPrompt(context.Context, string) (string, error) {
	return f.summary, f.err
}

type editionReaderFake struct {
	config *domain.EditionConfig
	err    error
}

func (f *editionReaderFake) Get(context.Context) (*domain.EditionConfig, error) {
	return f.config, f.err
}

func TestResolvePrompts(t *testing.T) {
	tests := []struct {
		name, chat, summary   string
		promptErr, editionErr error
		edition               *domain.EditionConfig
		wantChat, wantSummary string
	}{
		{name: "kb prompts win", chat: "kb chat", summary: "kb summary", edition: &domain.EditionConfig{DefaultPrompts: domain.EditionPrompts{Chat: "edition chat", Summary: "edition summary"}}, wantChat: "kb chat", wantSummary: "kb summary"},
		{name: "edition prompts fill empty kb prompts", edition: &domain.EditionConfig{DefaultPrompts: domain.EditionPrompts{Chat: "edition chat", Summary: "edition summary"}}, wantChat: "edition chat", wantSummary: "edition summary"},
		{name: "empty edition uses system defaults", edition: &domain.EditionConfig{}, wantChat: domain.SystemDefaultPrompt, wantSummary: domain.SystemDefaultSummaryPrompt},
		{name: "edition failure uses system defaults", editionErr: errors.New("broken edition"), wantChat: domain.SystemDefaultPrompt, wantSummary: domain.SystemDefaultSummaryPrompt},
		{name: "prompt read failure does not block", promptErr: errors.New("database unavailable"), edition: &domain.EditionConfig{DefaultPrompts: domain.EditionPrompts{Chat: "edition chat", Summary: "edition summary"}}, wantChat: domain.SystemDefaultPrompt, wantSummary: domain.SystemDefaultSummaryPrompt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &LLMUsecase{promptRepo: &promptStoreFake{chat: tt.chat, summary: tt.summary, err: tt.promptErr}, edition: &editionReaderFake{config: tt.edition, err: tt.editionErr}, logger: log.NewLogger(&config.Config{})}
			if got := u.ResolveChatPrompt(context.Background(), "kb"); got != tt.wantChat {
				t.Fatalf("chat = %q", got)
			}
			if got := u.ResolveSummaryPrompt(context.Background(), "kb"); got != tt.wantSummary {
				t.Fatalf("summary = %q", got)
			}
		})
	}
}
