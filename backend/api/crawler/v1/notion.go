package v1

type NotionParseReq struct {
	Integration string `json:"integration" validate:"required"`
}
type NotionParseResp struct {
	ID   string            `json:"id"`
	Docs []NotionParseItem `json:"docs"`
}

type NotionParseItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type NotionScrapeReq struct {
	KbID  string `json:"kb_id" validate:"required"`
	ID    string `json:"id" validate:"required"`
	DocId string `json:"doc_id" validate:"required"`
}

type NotionScrapeResp struct {
	Content string `json:"content"`
}
