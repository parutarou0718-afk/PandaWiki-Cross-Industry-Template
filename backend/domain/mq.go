package domain

const (
	VectorTaskTopic       = "apps.panda-wiki.vector.task"
	GraphTaskTopic        = "apps.panda-wiki.graph.task"
	AnydocTaskExportTopic = "anydoc.persistence.doc.task.export"
	RagDocUpdateTopic     = "raglite.events.doc.update"
)

var TopicConsumerName = map[string]string{
	VectorTaskTopic:       "panda-wiki-vector-consumer",
	GraphTaskTopic:        "panda-wiki-graph-consumer",
	AnydocTaskExportTopic: "anydoc-task-export-consumer",
	RagDocUpdateTopic:     "raglite-doc-update-consumer",
}

type NodeReleaseVectorRequest struct {
	KBID          string `json:"kb_id"`
	NodeReleaseID string `json:"node_release_id"`
	NodeID        string `json:"node_id"`
	DocID         string `json:"doc_id"` // for delete
	Action        string `json:"action"` // upsert, delete, summary
	GroupIds      []int  `json:"group_ids"`
}

// NodeGraphExtractionRequest is intentionally small: the worker reloads the
// current release server-side and never places the node body on the queue.
type NodeGraphExtractionRequest struct {
	KBID          string `json:"kb_id"`
	NodeReleaseID string `json:"node_release_id"`
}

// AnydocTaskExportEvent represents the task completion event from anydoc service
type AnydocTaskExportEvent struct {
	TaskID     string `json:"task_id"`
	PlatformID string `json:"platform_id"`
	DocID      string `json:"doc_id"`
	Status     string `json:"status"`
	Err        string `json:"err"`
	Markdown   string `json:"markdown"`
	JSON       string `json:"json"`
}

type RagDocInfoUpdateEvent struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
