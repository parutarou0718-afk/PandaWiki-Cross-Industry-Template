package v1

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/handler"
	"github.com/chaitin/panda-wiki/middleware"
	"github.com/chaitin/panda-wiki/usecase"
)

type PluginRecordGroupHandler struct {
	*handler.BaseHandler
	usecase *usecase.PluginRecordUsecase
}
type pluginRecordGroupRequest struct {
	KBID          string   `json:"kb_id"`
	Name          string   `json:"name"`
	MemberUserIDs []string `json:"member_user_ids"`
}

func NewPluginRecordGroupHandler(e *echo.Echo, base *handler.BaseHandler, auth middleware.AuthMiddleware, records *usecase.PluginRecordUsecase) *PluginRecordGroupHandler {
	h := &PluginRecordGroupHandler{BaseHandler: base, usecase: records}
	group := e.Group("/api/v1/knowledge_base/plugin-record-groups", auth.Authorize)
	group.GET("", h.List, auth.ValidateKBUserPerm(consts.UserKBPermissionNotNull))
	group.POST("", h.Create, auth.ValidateKBUserPerm(consts.UserKBPermissionFullControl))
	return h
}
func (h *PluginRecordGroupHandler) List(c echo.Context) error {
	kbID := strings.TrimSpace(c.QueryParam("kb_id"))
	if kbID == "" {
		return h.NewResponseWithError(c, "invalid plugin record group query", fmt.Errorf("kb_id is required"))
	}
	groups, err := h.usecase.ListGroups(c.Request().Context(), kbID)
	if err != nil {
		return h.NewResponseWithError(c, "list plugin record groups failed", err)
	}
	return h.NewResponseWithData(c, groups)
}
func (h *PluginRecordGroupHandler) Create(c echo.Context) error {
	var request pluginRecordGroupRequest
	if err := c.Bind(&request); err != nil {
		return h.NewResponseWithError(c, "invalid plugin record group request", err)
	}
	group, err := h.usecase.CreateGroup(c.Request().Context(), usecase.PluginRecordGroupWrite{KBID: strings.TrimSpace(request.KBID), Name: strings.TrimSpace(request.Name), MemberUserIDs: request.MemberUserIDs})
	if err != nil {
		return h.NewResponseWithError(c, "create plugin record group failed", err)
	}
	return h.NewResponseWithData(c, group)
}
