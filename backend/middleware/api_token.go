package middleware

import (
	"context"

	"github.com/chaitin/panda-wiki/domain"
)

type APITokenRepository interface {
	GetByTokenWithCache(ctx context.Context, token string) (*domain.APIToken, error)
}
