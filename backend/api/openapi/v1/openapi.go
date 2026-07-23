package v1

type GitHubCallbackReq struct {
	Code  string `json:"code" query:"code"`
	State string `json:"state" query:"state"`
}

type GitHubCallbackResp struct {
}
