package v1

type WikijsParseReq struct {
	KbID string `json:"kb_id" validate:"required"`
}

type WikijsParseItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type WikijsParseResp struct {
	ID   string            `json:"id"`
	Docs []WikijsParseItem `json:"docs"`
}

type WikijsScrapeReq struct {
	KbID  string `json:"kb_id" validate:"required"`
	ID    string `json:"id" validate:"required"`
	DocID string `json:"doc_id" validate:"required"`
}

type WikijsScrapeResp struct {
	Content string `json:"content"`
}
