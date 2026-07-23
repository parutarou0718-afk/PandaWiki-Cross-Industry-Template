package v1

type FeishuSpaceListReq struct {
	UserAccessToken string `json:"user_access_token" validate:"required"`
	AppID           string `json:"app_id" validate:"required"`
	AppSecret       string `json:"app_secret" validate:"required"`
}
type FeishuSpaceListResp struct {
	Name    string `json:"name"`
	SpaceId string `json:"space_id"`
}

type FeishuSearchWikiReq struct {
	UserAccessToken string `json:"user_access_token" validate:"required"`
	AppID           string `json:"app_id" validate:"required"`
	AppSecret       string `json:"app_secret" validate:"required"`
	SpaceId         string `json:"space_id"`
}

type FeishuSearchWikiResp struct {
	ID       string `json:"id" validate:"required"`
	DocId    string `json:"doc_id" validate:"required"`
	Title    string `json:"title"`
	FileType string `json:"file_type"`
	SpaceId  string `json:"space_id"`
}

type FeishuListCloudDocReq struct {
	UserAccessToken string `json:"user_access_token" validate:"required"`
	AppID           string `json:"app_id" validate:"required"`
	AppSecret       string `json:"app_secret" validate:"required"`
}

type FeishuListCloudDocResp struct {
	ID       string `json:"id" validate:"required"`
	DocId    string `json:"doc_id" validate:"required"`
	Title    string `json:"title"`
	FileType string `json:"file_type"`
	SpaceId  string `json:"space_id"`
}

type FeishuGetDocReq struct {
	KbID     string `json:"kb_id" validate:"required"`
	ID       string `json:"id" validate:"required"`
	DocId    string `json:"doc_id" validate:"required"`
	FileType string `json:"file_type"`
	SpaceId  string `json:"space_id"`
}

type FeishuGetDocResp struct {
	Content string `json:"content"`
}
