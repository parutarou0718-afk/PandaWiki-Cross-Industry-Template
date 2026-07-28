package v1

import (
	"encoding/json"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/handler"
	"github.com/chaitin/panda-wiki/middleware"
	"github.com/chaitin/panda-wiki/usecase"
)

// KnowledgeSearchHandler exposes the existing RAG retrieval path to signed-in
// users. Authorization is performed by ValidateKBUserPerm before the query is
// evaluated; group IDs are always derived from the authenticated user in the
// ChatUsecase, never from this request.
type KnowledgeSearchHandler struct {
	*handler.BaseHandler
	chat *usecase.ChatUsecase
}

type knowledgeSearchRequest struct {
	KBID  string `json:"kb_id" validate:"required"`
	Query string `json:"query" validate:"required"`
}

func NewKnowledgeSearchHandler(e *echo.Echo, base *handler.BaseHandler, auth middleware.AuthMiddleware, chat *usecase.ChatUsecase) *KnowledgeSearchHandler {
	h := &KnowledgeSearchHandler{BaseHandler: base, chat: chat}
	group := e.Group("/api/v1/knowledge_base", auth.Authorize)
	group.POST("/search", h.Search, auth.ValidateKBUserPerm(consts.UserKBPermissionNotNull))
	return h
}

func (h *KnowledgeSearchHandler) Search(c echo.Context) error {
	var req knowledgeSearchRequest
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return h.NewResponseWithError(c, "invalid knowledge search request", err)
	}
	if strings.TrimSpace(req.KBID) == "" || strings.TrimSpace(req.Query) == "" {
		return h.NewResponseWithError(c, "kb_id and query are required", nil)
	}

	resp, err := h.chat.Search(c.Request().Context(), &domain.ChatSearchReq{
		KBID:       req.KBID,
		Message:    strings.TrimSpace(req.Query),
		AuthUserID: domain.GetAuthID(c),
	})
	if err != nil {
		return h.NewResponseWithError(c, "knowledge search failed", err)
	}
	return h.NewResponseWithData(c, resp)
}
