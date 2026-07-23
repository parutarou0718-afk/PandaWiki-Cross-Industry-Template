package v1

type WechatAppInfoResp struct {
	WeChatAppIsEnabled bool     `json:"wechat_app_is_enabled"`
	FeedbackEnable     bool     `json:"feedback_enable"`
	FeedbackType       []string `json:"feedback_type"`
	DisclaimerContent  string   `json:"disclaimer_content"`
}
