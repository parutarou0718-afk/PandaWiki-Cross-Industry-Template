package consts

type CrawlerStatus string

const (
	CrawlerStatusPending   CrawlerStatus = "pending"
	CrawlerStatusInProcess CrawlerStatus = "in_process"
	CrawlerStatusCompleted CrawlerStatus = "completed"
	CrawlerStatusFailed    CrawlerStatus = "failed"
)
