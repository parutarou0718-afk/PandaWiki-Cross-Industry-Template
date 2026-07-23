package anydoc

type FeishuSetting struct {
	UserAccessToken string `json:"user_access_token"`
	AppID           string `json:"app_id"`
	AppSecret       string `json:"app_secret"`
	SpaceId         string `json:"space_id"`
}

type DingtalkSetting struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
	SpaceID   string `json:"space_id"`
	UnionID   string `json:"unionid"`
	Phone     string `json:"phone"`
}
