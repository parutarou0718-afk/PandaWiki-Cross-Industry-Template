package v1

import (
	"encoding/json"

	"github.com/labstack/echo/v4"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/handler"
	"github.com/chaitin/panda-wiki/middleware"
	"github.com/chaitin/panda-wiki/usecase"
)

type ReportProfileAdminHandler struct {
	*handler.BaseHandler
	usecase *usecase.ReportProfileAdminUsecase
}

func NewReportProfileAdminHandler(e *echo.Echo, base *handler.BaseHandler, auth middleware.AuthMiddleware, uc *usecase.ReportProfileAdminUsecase) *ReportProfileAdminHandler {
	h := &ReportProfileAdminHandler{BaseHandler: base, usecase: uc}
	group := e.Group("/api/v1/admin/report-profiles", auth.Authorize, auth.ValidateUserRole(consts.UserRoleAdmin))
	group.GET("", h.List)
	group.GET("/:id", h.Get)
	group.POST("", h.Create)
	group.PUT("/:id", h.Update)
	group.PATCH("/:id/enabled", h.SetEnabled)
	group.DELETE("/:id", h.Delete)
	group.POST("/:id/restore-default", h.RestoreBuiltinDefault)
	return h
}

func (h *ReportProfileAdminHandler) List(c echo.Context) error {
	value, err := h.usecase.List(c.Request().Context())
	if err != nil {
		return h.NewResponseWithError(c, "list report profiles failed", err)
	}
	return h.NewResponseWithData(c, value)
}
func (h *ReportProfileAdminHandler) Get(c echo.Context) error {
	value, err := h.usecase.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return h.NewResponseWithError(c, "get report profile failed", err)
	}
	return h.NewResponseWithData(c, value)
}
func (h *ReportProfileAdminHandler) Create(c echo.Context) error {
	req, err := decodeAdminReportProfileRequest(c)
	if err != nil {
		return h.NewResponseWithError(c, "invalid report profile request", err)
	}
	value, err := h.usecase.Create(c.Request().Context(), req)
	if err != nil {
		return h.NewResponseWithError(c, "create report profile failed", err)
	}
	return h.NewResponseWithData(c, value)
}
func (h *ReportProfileAdminHandler) Update(c echo.Context) error {
	req, err := decodeAdminReportProfileRequest(c)
	if err != nil {
		return h.NewResponseWithError(c, "invalid report profile request", err)
	}
	value, err := h.usecase.Update(c.Request().Context(), c.Param("id"), req)
	if err != nil {
		return h.NewResponseWithError(c, "update report profile failed", err)
	}
	return h.NewResponseWithData(c, value)
}
func (h *ReportProfileAdminHandler) SetEnabled(c echo.Context) error {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return h.NewResponseWithError(c, "invalid enabled request", err)
	}
	if err := h.usecase.SetEnabled(c.Request().Context(), c.Param("id"), req.Enabled); err != nil {
		return h.NewResponseWithError(c, "update report profile status failed", err)
	}
	return h.NewResponseWithData(c, map[string]bool{"enabled": req.Enabled})
}
func (h *ReportProfileAdminHandler) Delete(c echo.Context) error {
	if err := h.usecase.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return h.NewResponseWithError(c, "delete report profile failed", err)
	}
	return h.NewResponseWithData(c, true)
}
func (h *ReportProfileAdminHandler) RestoreBuiltinDefault(c echo.Context) error {
	value, err := h.usecase.RestoreBuiltinDefault(c.Request().Context(), c.Param("id"))
	if err != nil {
		return h.NewResponseWithError(c, "restore report profile default failed", err)
	}
	return h.NewResponseWithData(c, value)
}
func decodeAdminReportProfileRequest(c echo.Context) (*usecase.ReportProfileAdminRequest, error) {
	var req usecase.ReportProfileAdminRequest
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}
