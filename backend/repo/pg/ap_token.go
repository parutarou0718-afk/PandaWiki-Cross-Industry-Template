package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/store/cache"
	"github.com/chaitin/panda-wiki/store/pg"
)

type APITokenRepo struct {
	db     *pg.DB
	logger *log.Logger
	cache  *cache.Cache
}

func NewAPITokenRepo(db *pg.DB, logger *log.Logger, cache *cache.Cache) *APITokenRepo {
	return &APITokenRepo{
		db:     db,
		logger: logger,
		cache:  cache,
	}
}

func (r *APITokenRepo) GetByTokenWithCache(ctx context.Context, token string) (*domain.APIToken, error) {
	cacheKey := fmt.Sprintf("api_token:%s", token)

	cachedData, err := r.cache.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		var apiToken domain.APIToken
		if err := json.Unmarshal([]byte(cachedData), &apiToken); err == nil {
			return &apiToken, nil
		}
	}

	// 缓存未命中，从数据库查询
	var apiToken domain.APIToken
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&apiToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get api token by token failed: %w", err)
	}

	if tokenData, err := json.Marshal(&apiToken); err == nil {
		if err := r.cache.Set(ctx, cacheKey, tokenData, 30*time.Minute).Err(); err != nil {
			r.logger.Warn("failed to cache API token", log.Error(err))
		}
	}

	return &apiToken, nil
}

func (r *APITokenRepo) ListByKBID(ctx context.Context, kbID string) ([]domain.APIToken, error) {
	var tokens []domain.APIToken
	if err := r.db.WithContext(ctx).Where("kb_id = ?", kbID).Order("created_at DESC").Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("list API tokens failed: %w", err)
	}
	return tokens, nil
}

func (r *APITokenRepo) Create(ctx context.Context, token *domain.APIToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return fmt.Errorf("create API token failed: %w", err)
	}
	return nil
}

func (r *APITokenRepo) Update(ctx context.Context, kbID, id string, name *string, permission *consts.UserKBPermission) error {
	var token domain.APIToken
	if err := r.db.WithContext(ctx).Where("id = ? AND kb_id = ?", id, kbID).First(&token).Error; err != nil {
		return fmt.Errorf("get API token failed: %w", err)
	}
	updates := map[string]any{}
	if name != nil {
		updates["name"] = *name
	}
	if permission != nil {
		updates["permission"] = *permission
	}
	if err := r.db.WithContext(ctx).Model(&domain.APIToken{}).Where("id = ? AND kb_id = ?", id, kbID).Updates(updates).Error; err != nil {
		return fmt.Errorf("update API token failed: %w", err)
	}
	return r.invalidateTokenCache(ctx, token.Token)
}

func (r *APITokenRepo) Delete(ctx context.Context, kbID, id string) error {
	var token domain.APIToken
	if err := r.db.WithContext(ctx).Where("id = ? AND kb_id = ?", id, kbID).First(&token).Error; err != nil {
		return fmt.Errorf("get API token failed: %w", err)
	}
	if err := r.db.WithContext(ctx).Delete(&domain.APIToken{}, "id = ? AND kb_id = ?", id, kbID).Error; err != nil {
		return fmt.Errorf("delete API token failed: %w", err)
	}
	return r.invalidateTokenCache(ctx, token.Token)
}

func (r *APITokenRepo) invalidateTokenCache(ctx context.Context, token string) error {
	if r.cache == nil {
		return nil
	}
	if err := r.cache.Del(ctx, fmt.Sprintf("api_token:%s", token)).Err(); err != nil {
		return fmt.Errorf("invalidate API token cache failed: %w", err)
	}
	return nil
}
