package v1

import (
	"github.com/labstack/echo/v4"

	v1 "github.com/chaitin/panda-wiki/api/auth/v1"
	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/handler"
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/usecase"
)

type AuthV1Handler struct {
	*handler.BaseHandler
	logger      *log.Logger
	authUseCase *usecase.AuthUsecase
}

func NewAuthV1Handler(
	e *echo.Echo,
	baseHandler *handler.BaseHandler,
	logger *log.Logger,
	authUseCase *usecase.AuthUsecase,
) *AuthV1Handler {
	h := &AuthV1Handler{
		BaseHandler: baseHandler,
		logger:      logger,
		authUseCase: authUseCase,
	}

	AuthGroup := e.Group(
		"/api/v1/auth",
		h.V1Auth.Authorize,
		h.V1Auth.ValidateKBUserPerm(consts.UserKBPermissionFullControl),
	)
	AuthGroup.GET("/get", h.OpenAuthGet)
	AuthGroup.POST("/set", h.OpenAuthSet)
	AuthGroup.DELETE("/delete", h.OpenAuthDelete)

	return h
}

// OpenAuthGet 获取授权信息
//
//	@Tags			Auth
//	@Summary		获取授权信息
//	@Description	获取授权信息
//	@ID				v1-OpenAuthGet
//	@Accept			json
//	@Produce		json
//	@Security		bearerAuth
//	@Param			param	query		v1.AuthGetReq	true	"para"
//	@Success		200		{object}	domain.PWResponse{data=v1.AuthGetResp}
//	@Router			/api/v1/auth/get [get]
func (h *AuthV1Handler) OpenAuthGet(c echo.Context) error {

	var req v1.AuthGetReq
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return h.NewResponseWithError(c, "validate request params failed", err)
	}

	resp, err := h.authUseCase.GetAuth(c.Request().Context(), req.KBID, req.SourceType)
	if err != nil {
		return h.NewResponseWithError(c, "failed to get Auth", err)
	}

	return h.NewResponseWithData(c, resp)
}

// OpenAuthSet 获取授权信息
//
//	@Tags			Auth
//	@Summary		设置授权信息
//	@Description	设置授权信息
//	@ID				v1-OpenAuthSet
//	@Accept			json
//	@Produce		json
//	@Security		bearerAuth
//	@Param			param	body		v1.AuthSetReq	true	"para"
//	@Success		200		{object}	domain.Response
//	@Router			/api/v1/auth/set [post]
func (h *AuthV1Handler) OpenAuthSet(c echo.Context) error {

	var req v1.AuthSetReq
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return h.NewResponseWithError(c, "validate request params failed", err)
	}

	if err := h.authUseCase.SetAuth(c.Request().Context(), req); err != nil {
		return h.NewResponseWithError(c, "failed to set Auth", err)
	}

	return h.NewResponseWithData(c, nil)
}

// OpenAuthDelete 删除授权信息
//
//	@Tags			Auth
//	@Summary		删除授权信息
//	@Description	删除授权信息
//	@ID				v1-OpenAuthDelete
//	@Accept			json
//	@Produce		json
//	@Security		bearerAuth
//	@Param			param	query		v1.AuthDeleteReq	true	"para"
//	@Success		200		{object}	domain.Response
//	@Router			/api/v1/auth/delete [delete]
func (h *AuthV1Handler) OpenAuthDelete(c echo.Context) error {

	var req v1.AuthDeleteReq
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return h.NewResponseWithError(c, "validate request params failed", err)
	}

	if err := h.authUseCase.DeleteAuth(c.Request().Context(), req); err != nil {
		return h.NewResponseWithError(c, "failed to delete Auth", err)
	}

	return h.NewResponseWithData(c, nil)
}
