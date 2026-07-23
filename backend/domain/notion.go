package domain

type Page struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	ParentId string `json:"parent_id"`
	Content  string `json:"content"`
}
type PageInfo struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}
