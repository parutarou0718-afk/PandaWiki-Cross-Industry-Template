package v1

type EpubParseReq struct {
	KbID     string `json:"kb_id" validate:"required"`
	Filename string `json:"filename" validate:"required"`
	Key      string `json:"key" validate:"required"`
}

type EpubParseResp struct {
	TaskID string `json:"task_id"`
}
