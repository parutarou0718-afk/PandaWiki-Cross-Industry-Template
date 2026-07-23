package pg

import (
	"context"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/store/pg"
)

type WechatRepository struct {
	db     *pg.DB
	logger *log.Logger
}

func NewWechatRepository(db *pg.DB, logger *log.Logger) *WechatRepository {
	return &WechatRepository{db: db, logger: logger.WithModule("repo.pg.wechat")}
}

func (r *WechatRepository) GetWechatStatic(ctx context.Context, kbID string, appType domain.AppType) (*domain.WechatStatic, error) {
	var wechatStatic domain.WechatStatic
	if err := r.db.WithContext(ctx).Model(&domain.App{}).
		Where("kb_id = ? AND type = ?", kbID, appType).
		Joins("join knowledge_bases kb on kb.id = kb_id ").
		Select("apps.settings ->>'icon' as image_path", "kb.access_settings ->>'base_url' as base_url").
		Find(&wechatStatic).Error; err != nil {
		return nil, err
	}
	return &wechatStatic, nil
}

func (r *WechatRepository) GetWechatBaseURL(ctx context.Context, kbID string) (string, error) {
	var baseUrl string
	if err := r.db.WithContext(ctx).Model(&domain.KnowledgeBase{}).
		Where("id = ?", kbID).
		Select("access_settings ->>'base_url'").
		First(&baseUrl).Error; err != nil {
		return "", err
	}
	return baseUrl, nil
}
