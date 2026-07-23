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

type EditionHandler struct {
	*handler.BaseHandler
	logger  *log.Logger
	auth    middleware.AuthMiddleware
	usecase *usecase.EditionUsecase
}

func NewEditionHandler(e *echo.Echo, base *handler.BaseHandler, logger *log.Logger, auth middleware.AuthMiddleware, uc *usecase.EditionUsecase) *EditionHandler {
	h := &EditionHandler{BaseHandler: base, logger: logger.WithModule("handler.v1.edition"), auth: auth, usecase: uc}
	group := e.Group("/api/v1/system", auth.Authorize)
	group.GET("/edition", h.Get)
	group.PUT("/edition", h.Update, auth.ValidateUserRole(consts.UserRoleAdmin))
	return h
}

func (h *EditionHandler) Get(c echo.Context) error {
	config, err := h.usecase.Get(c.Request().Context())
	if err != nil {
		return h.NewResponseWithError(c, "get edition config failed", err)
	}
	return h.NewResponseWithData(c, config)
}

func (h *EditionHandler) Update(c echo.Context) error {
	var req domain.UpdateEditionReq
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return h.NewResponseWithError(c, "invalid edition config request", err)
	}
	userID := domain.GetAuthID(c)
	if userID == 0 {
		return h.NewResponseWithError(c, "unauthorized", nil)
	}
	config, err := h.usecase.Update(c.Request().Context(), userID, &req)
	if err != nil {
		return h.NewResponseWithError(c, "update edition config failed", err)
	}
	return h.NewResponseWithData(c, config)
}
