package v1

import "github.com/chaitin/panda-wiki/consts"

type AuthLoginSimpleReq struct {
	Password string `json:"password" validate:"required"`
}

type AuthLoginSimpleResp struct {
}

type AuthGetReq struct {
}
type AuthGetResp struct {
	AuthType       consts.AuthType       `json:"auth_type"`
	SourceType     consts.SourceType     `json:"source_type"`
	LicenseEdition consts.LicenseEdition `json:"license_edition"`
}

type AuthGitHubReq struct {
	KbID        string `json:"kb_id"`
	RedirectUrl string `json:"redirect_url"`
}

type AuthGitHubResp struct {
	Url string `json:"url"`
}

type GitHubCallbackReq struct {
	Code  string `json:"code" query:"code"`
	State string `json:"state" query:"state"`
}

type GitHubCallbackResp struct {
}
