package fns

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
)

type MigrationCreateBotAuth struct {
	Name   string
	logger *log.Logger
}

func NewMigrationCreateBotAuth(logger *log.Logger) *MigrationCreateBotAuth {
	return &MigrationCreateBotAuth{
		Name:   "0002_create_bot_auth",
		logger: logger,
	}
}

func (m *MigrationCreateBotAuth) Execute(tx *gorm.DB) error {
	ctx := context.Background()

	// 获取所有机器人类型的应用
	var apps []domain.App
	if err := tx.WithContext(ctx).Where("type IN ?", []domain.AppType{
		domain.AppTypeWidget,
		domain.AppTypeDingTalkBot,
		domain.AppTypeFeishuBot,
		domain.AppTypeWechatBot,
		domain.AppTypeWechatServiceBot,
		domain.AppTypeDisCordBot,
		domain.AppTypeWechatOfficialAccount,
	}).Find(&apps).Error; err != nil {
		return fmt.Errorf("failed to get apps: %w", err)
	}

	m.logger.Info("found apps for bot auth creation", log.Int("count", len(apps)))

	for _, app := range apps {
		sourceType := app.Type.ToSourceType()
		if sourceType == "" {
			m.logger.Warn("app type has no corresponding source type", log.String("app_id", app.ID), log.Any("app_type", uint8(app.Type)))
			continue
		}

		// 检查是否需要创建认证记录（检查应用是否启用）
		shouldCreateAuth := false

		switch app.Type {
		case domain.AppTypeWidget:
			shouldCreateAuth = app.Settings.WidgetBotSettings.IsOpen
		case domain.AppTypeDingTalkBot:
			shouldCreateAuth = app.Settings.DingTalkBotIsEnabled != nil && *app.Settings.DingTalkBotIsEnabled
		case domain.AppTypeFeishuBot:
			shouldCreateAuth = app.Settings.FeishuBotIsEnabled != nil && *app.Settings.FeishuBotIsEnabled
		case domain.AppTypeWechatBot:
			shouldCreateAuth = app.Settings.WeChatAppIsEnabled != nil && *app.Settings.WeChatAppIsEnabled
		case domain.AppTypeWechatServiceBot:
			shouldCreateAuth = app.Settings.WeChatServiceIsEnabled != nil && *app.Settings.WeChatServiceIsEnabled
		case domain.AppTypeDisCordBot:
			shouldCreateAuth = app.Settings.DiscordBotIsEnabled != nil && *app.Settings.DiscordBotIsEnabled
		case domain.AppTypeWechatOfficialAccount:
			shouldCreateAuth = app.Settings.WechatOfficialAccountIsEnabled != nil && *app.Settings.WechatOfficialAccountIsEnabled
		}

		if !shouldCreateAuth {
			m.logger.Debug("app is not enabled, skipping auth creation", log.String("app_id", app.ID), log.String("source_type", string(sourceType)))
			continue
		}

		// 检查是否已存在该类型的认证记录
		var existingAuthCount int64
		if err := tx.WithContext(ctx).Model(&domain.Auth{}).
			Where("kb_id = ? AND source_type = ?", app.KBID, string(sourceType)).
			Count(&existingAuthCount).Error; err != nil {
			return fmt.Errorf("failed to check existing auth for kb_id %s, source_type %s: %w", app.KBID, sourceType, err)
		}

		if existingAuthCount > 0 {
			m.logger.Debug("auth already exists, skipping", log.String("kb_id", app.KBID), log.String("source_type", string(sourceType)))
			continue
		}

		// 创建新的认证记录
		auth := &domain.Auth{
			KBID:          app.KBID,
			UnionID:       fmt.Sprintf("bot_%s_%s", app.ID, sourceType),
			SourceType:    sourceType,
			LastLoginTime: time.Now(),
			UserInfo: domain.AuthUserInfo{
				Username: sourceType.Name(),
			},
		}

		if err := tx.WithContext(ctx).Create(auth).Error; err != nil {
			return fmt.Errorf("failed to create auth for kb_id %s, source_type %s: %w", app.KBID, sourceType, err)
		}

		m.logger.Info("created bot auth",
			log.String("kb_id", app.KBID),
			log.String("app_id", app.ID),
			log.String("source_type", string(sourceType)),
			log.String("union_id", auth.UnionID),
			log.Any("auth_id", auth.ID))
	}

	m.logger.Info("bot auth migration completed successfully")
	return nil
}
