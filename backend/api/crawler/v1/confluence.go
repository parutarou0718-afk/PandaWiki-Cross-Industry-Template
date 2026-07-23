package v1

type ConfluenceParseReq struct {
	KbID string `json:"kb_id" validate:"required"`
}

type ConfluenceParseItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type ConfluenceParseResp struct {
	ID   string                `json:"id"`
	Docs []ConfluenceParseItem `json:"docs"`
}

type ConfluenceScrapeReq struct {
	KbID  string `json:"kb_id" validate:"required"`
	ID    string `json:"id" validate:"required"`
	DocID string `json:"doc_id" validate:"required"`
}

type ConfluenceScrapeResp struct {
	Content string `json:"content"`
}
