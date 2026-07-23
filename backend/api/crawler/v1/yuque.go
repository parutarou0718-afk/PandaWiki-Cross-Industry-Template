package v1

type YuqueParseReq struct {
	KbID     string `json:"kb_id" validate:"required"`
	Filename string `json:"filename" validate:"required"`
	Key      string `json:"key" validate:"required"`
}

type YuqueParseResp struct {
	List []YuqueParseItem `json:"list"`
}

type YuqueParseItem struct {
	TaskID string `json:"task_id"`
	Title  string `json:"title"`
}
