package v1

import (
	"encoding/json"

	"github.com/labstack/echo/v4"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/handler"
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/middleware"
	"github.com/chaitin/panda-wiki/usecase"
)

type ReportHandler struct {
	*handler.BaseHandler
	usecase *usecase.ReportUsecase
}

func NewReportHandler(e *echo.Echo, base *handler.BaseHandler, logger *log.Logger, auth middleware.AuthMiddleware, uc *usecase.ReportUsecase) *ReportHandler {
	h := &ReportHandler{BaseHandler: base, usecase: uc}
	group := e.Group("/api/v1", auth.Authorize)
	group.GET("/report-profiles", h.ListProfiles, auth.ValidateKBUserPerm(consts.UserKBPermissionNotNull))
	group.POST("/reports", h.Create, auth.ValidateKBUserPerm(consts.UserKBPermissionNotNull))
	group.GET("/reports", h.List)
	group.GET("/reports/:id", h.Get)
	_ = logger // constructor parity with other handlers; logging is performed in usecase without prompt content.
	return h
}

func (h *ReportHandler) ListProfiles(c echo.Context) error {
	items, err := h.usecase.ListProfiles(c.Request().Context())
	if err != nil {
		return h.NewResponseWithError(c, "list report profiles failed", err)
	}
	return h.NewResponseWithData(c, items)
}

func (h *ReportHandler) Create(c echo.Context) error {
	var req usecase.CreateReportRequest
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return h.NewResponseWithError(c, "invalid report request", err)
	}
	authID := domain.GetAuthID(c)
	detail, err := h.usecase.Create(c.Request().Context(), authID, req)
	if err != nil {
		return h.NewResponseWithError(c, "create report failed", err)
	}
	return h.NewResponseWithData(c, detail)
}

func (h *ReportHandler) List(c echo.Context) error {
	reports, err := h.usecase.ListMine(c.Request().Context(), domain.GetAuthID(c))
	if err != nil {
		return h.NewResponseWithError(c, "list reports failed", err)
	}
	return h.NewResponseWithData(c, reports)
}

func (h *ReportHandler) Get(c echo.Context) error {
	detail, err := h.usecase.GetMine(c.Request().Context(), domain.GetAuthID(c), c.Param("id"))
	if err != nil {
		return h.NewResponseWithError(c, "get report failed", err)
	}
	return h.NewResponseWithData(c, detail)
}
