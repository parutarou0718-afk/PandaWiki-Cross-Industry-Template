package v1

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/handler"
	"github.com/chaitin/panda-wiki/middleware"
	"github.com/chaitin/panda-wiki/usecase"
)

// GraphHandler exposes a read-only, permission-filtered graph projection. The
// client supplies a KB ID but never permission group IDs or source content.
type GraphHandler struct {
	*handler.BaseHandler
	graph *usecase.GraphUsecase
}

type graphQuery struct {
	KBID string `query:"kb_id" validate:"required"`
}

type graphRebuildRequest struct {
	KBID string `json:"kb_id"`
}

type knowledgeSchemaRequest struct {
	KBID   string                 `json:"kb_id"`
	Schema domain.KnowledgeSchema `json:"schema"`
}

func NewGraphHandler(e *echo.Echo, base *handler.BaseHandler, auth middleware.AuthMiddleware, graph *usecase.GraphUsecase) *GraphHandler {
	h := &GraphHandler{BaseHandler: base, graph: graph}
	group := e.Group("/api/v1/knowledge_base", auth.Authorize)
	group.GET("/graph", h.GetGraph, auth.ValidateKBUserPerm(consts.UserKBPermissionNotNull))
	group.GET("/graph/schema", h.GetSchema, auth.ValidateKBUserPerm(consts.UserKBPermissionNotNull))
	// Schema changes control what the server extracts and exposes for the
	// whole knowledge base, so they require the same full-control boundary as
	// KB membership and settings changes.
	group.PUT("/graph/schema", h.SaveSchema, auth.ValidateKBUserPerm(consts.UserKBPermissionFullControl))
	group.POST("/graph/rebuild", h.RebuildGraph, auth.ValidateKBUserPerm(consts.UserKBPermissionDocManage))
	return h
}

func (h *GraphHandler) GetSchema(c echo.Context) error {
	var query graphQuery
	if err := c.Bind(&query); err != nil || strings.TrimSpace(query.KBID) == "" {
		return h.NewResponseWithError(c, "kb_id is required", err)
	}
	schema, err := h.graph.GetSchema(c.Request().Context(), query.KBID)
	if err != nil {
		return h.NewResponseWithError(c, "get knowledge graph schema failed", err)
	}
	return h.NewResponseWithData(c, schema)
}

func (h *GraphHandler) SaveSchema(c echo.Context) error {
	req, err := decodeKnowledgeSchemaRequest(c.Request().Body)
	if err != nil || strings.TrimSpace(req.KBID) == "" {
		return h.NewResponseWithError(c, "invalid knowledge graph schema request", err)
	}
	if err := h.graph.SaveSchema(c.Request().Context(), req.KBID, req.Schema); err != nil {
		return h.NewResponseWithError(c, "save knowledge graph schema failed", err)
	}
	return h.NewResponseWithData(c, req.Schema)
}

func decodeKnowledgeSchemaRequest(reader io.Reader) (knowledgeSchemaRequest, error) {
	var req knowledgeSchemaRequest
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return knowledgeSchemaRequest{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return knowledgeSchemaRequest{}, io.ErrUnexpectedEOF
	}
	if err := req.Schema.Validate(); err != nil {
		return knowledgeSchemaRequest{}, err
	}
	return req, nil
}

func (h *GraphHandler) RebuildGraph(c echo.Context) error {
	req, err := decodeGraphRebuildRequest(c.Request().Body)
	if err != nil || req.KBID == "" {
		return h.NewResponseWithError(c, "kb_id is required", err)
	}
	count, err := h.graph.EnqueueKnowledgeBase(c.Request().Context(), req.KBID)
	if err != nil {
		return h.NewResponseWithError(c, "queue knowledge graph rebuild failed", err)
	}
	return h.NewResponseWithData(c, map[string]int{"queued": count})
}

func decodeGraphRebuildRequest(reader io.Reader) (graphRebuildRequest, error) {
	var req graphRebuildRequest
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return graphRebuildRequest{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return graphRebuildRequest{}, io.ErrUnexpectedEOF
	}
	req.KBID = strings.TrimSpace(req.KBID)
	return req, nil
}

func (h *GraphHandler) GetGraph(c echo.Context) error {
	var query graphQuery
	if err := c.Bind(&query); err != nil || strings.TrimSpace(query.KBID) == "" {
		return h.NewResponseWithError(c, "kb_id is required", err)
	}
	graph, err := h.graph.GetVisibleGraph(c.Request().Context(), query.KBID, domain.GetAuthID(c))
	if err != nil {
		return h.NewResponseWithError(c, "get knowledge graph failed", err)
	}
	return h.NewResponseWithData(c, graph)
}
