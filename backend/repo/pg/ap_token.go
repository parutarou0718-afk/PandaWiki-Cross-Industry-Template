package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

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
