package domain

type EpubReq struct {
	KbID string `json:"kb_id" binding:"required" validate:"required"`
}

type EpubResp struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Title   string `json:"title"`
}
