package v1

import (
	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/pkg/anydoc"
)

type CrawlerParseReq struct {
	Key             string                 `json:"key"`
	KbID            string                 `json:"kb_id" validate:"required"`
	CrawlerSource   consts.CrawlerSource   `json:"crawler_source" validate:"required"`
	Filename        string                 `json:"filename"`
	FeishuSetting   anydoc.FeishuSetting   `json:"feishu_setting"`
	DingtalkSetting anydoc.DingtalkSetting `json:"dingtalk_setting"`
}

type CrawlerParseResp struct {
	ID   string       `json:"id"`
	Docs anydoc.Child `json:"docs"`
}

type CrawlerExportReq struct {
	KbID     string `json:"kb_id" validate:"required"`
	ID       string `json:"id" validate:"required"`
	DocID    string `json:"doc_id" validate:"required"`
	SpaceId  string `json:"space_id"`
	FileType string `json:"file_type"`
}

type CrawlerExportResp struct {
	TaskId string `json:"task_id"`
}

type CrawlerResultReq struct {
	TaskId string `json:"task_id"  query:"task_id" validate:"required"`
}

type CrawlerResultResp struct {
	Status  consts.CrawlerStatus `json:"status" validate:"required"`
	Content string               `json:"content"`
}

type CrawlerResultsReq struct {
	TaskIds []string `json:"task_ids"  validate:"required"`
}

type CrawlerResultsResp struct {
	Status consts.CrawlerStatus `json:"status"`
	List   []CrawlerResultItem  `json:"list"`
}
type CrawlerResultItem struct {
	TaskId  string               `json:"task_id"`
	Status  consts.CrawlerStatus `json:"status"`
	Content string               `json:"content"`
}
