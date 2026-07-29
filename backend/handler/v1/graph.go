package v1

import (
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

func NewGraphHandler(e *echo.Echo, base *handler.BaseHandler, auth middleware.AuthMiddleware, graph *usecase.GraphUsecase) *GraphHandler {
	h := &GraphHandler{BaseHandler: base, graph: graph}
	group := e.Group("/api/v1/knowledge_base", auth.Authorize)
	group.GET("/graph", h.GetGraph, auth.ValidateKBUserPerm(consts.UserKBPermissionNotNull))
	group.POST("/graph/rebuild", h.RebuildGraph, auth.ValidateKBUserPerm(consts.UserKBPermissionDocManage))
	return h
}

func (h *GraphHandler) RebuildGraph(c echo.Context) error {
	var query graphQuery
	if err := c.Bind(&query); err != nil || strings.TrimSpace(query.KBID) == "" {
		return h.NewResponseWithError(c, "kb_id is required", err)
	}
	count, err := h.graph.EnqueueKnowledgeBase(c.Request().Context(), query.KBID)
	if err != nil {
		return h.NewResponseWithError(c, "queue knowledge graph rebuild failed", err)
	}
	return h.NewResponseWithData(c, map[string]int{"queued": count})
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
