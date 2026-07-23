package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	shareV1 "github.com/chaitin/panda-wiki/api/share/v1"
	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/pkg/oauth"
)

func (u *AuthUsecase) getGitHubClient(ctx context.Context, kbId, redirectURI string) (*oauth.Client, error) {
	authConfig, err := u.AuthRepo.GetAuthConfig(ctx, kbId, consts.SourceTypeGitHub)
	if authConfig == nil || err != nil {
		return nil, err
	}

	authSetting := authConfig.AuthSetting

	return oauth.NewGithubClient(ctx, u.logger, authSetting.ClientID, authSetting.ClientSecret, redirectURI, authSetting.Proxy)
}

func (u *AuthUsecase) GenerateGitHubAuthUrl(ctx context.Context, req shareV1.AuthGitHubReq) (string, error) {
	state, err := u.genState(ctx, StateInfo{
		KbId:        req.KbID,
		RedirectUrl: req.RedirectUrl,
	})
	if err != nil {
		return "", fmt.Errorf("gen state failed: %w", err)
	}

	githubClient, err := u.getGitHubClient(ctx, req.KbID, req.RedirectUrl)
	if err != nil {
		return "", fmt.Errorf("get githubClient failed: %w", err)
	}

	url := githubClient.GetAuthorizeURL(state)
	return url, nil
}

func (u *AuthUsecase) GitHubCallback(ctx context.Context, req shareV1.GitHubCallbackReq) (*domain.Auth, string, error) {

	statInfo, err := u.getStateInfo(ctx, req.State)
	if err != nil {
		return nil, "", err
	}

	githubClient, err := u.getGitHubClient(ctx, statInfo.KbId, statInfo.RedirectUrl)
	if err != nil {
		return nil, "", err
	}

	userInfo, err := githubClient.GetUserInfo(req.Code)
	if err != nil {
		return nil, "", err
	}

	auth := &domain.Auth{
		UserInfo: domain.AuthUserInfo{
			Username:  userInfo.Name,
			AvatarUrl: userInfo.AvatarUrl,
			Email:     userInfo.Email,
		},
		KBID:       statInfo.KbId,
		UnionID:    userInfo.ID,
		SourceType: consts.SourceTypeGitHub,
	}

	auth, err = u.AuthRepo.GetOrCreateAuth(ctx, auth, consts.SourceTypeGitHub)
	if err != nil {
		return nil, "", fmt.Errorf("create auth failed: %w", err)
	}

	return auth, statInfo.RedirectUrl, err
}

func (u *AuthUsecase) getStateInfo(ctx context.Context, state string) (*StateInfo, error) {
	statInfoBytes, err := u.cache.Get(ctx, state).Result()
	if err != nil {
		return nil, err
	}
	if statInfoBytes == "" {
		return nil, fmt.Errorf("state info not found")
	}

	var statInfo StateInfo
	err = json.Unmarshal([]byte(statInfoBytes), &statInfo)
	if err != nil {
		return nil, err
	}

	return &statInfo, nil
}
