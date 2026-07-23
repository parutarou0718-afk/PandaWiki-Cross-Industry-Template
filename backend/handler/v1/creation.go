package v1

import (
	"context"

	"github.com/labstack/echo/v4"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/handler"
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/usecase"
)

type CreationHandler struct {
	*handler.BaseHandler
	logger  *log.Logger
	usecase *usecase.CreationUsecase
}

func NewCreationHandler(echo *echo.Echo, baseHandler *handler.BaseHandler, logger *log.Logger, usecase *usecase.CreationUsecase) *CreationHandler {
	h := &CreationHandler{
		BaseHandler: baseHandler,
		logger:      logger.WithModule("handler.v1.creation"),
		usecase:     usecase,
	}

	api := echo.Group("/api/v1/creation", h.V1Auth.Authorize)
	api.POST("/text", h.Text)
	api.POST("/tab-complete", h.TabComplete)

	return h
}

// Text text creation
//
//	@Summary		Text creation
//	@Description	Text creation
//	@Tags			creation
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.TextReq	true	"text creation request"
//	@Success		200		{string}	string			"success"
//	@Router			/api/v1/creation/text [post]
func (h *CreationHandler) Text(c echo.Context) error {
	var req domain.TextReq
	if err := c.Bind(&req); err != nil {
		return h.NewResponseWithError(c, "request body is invalid", err)
	}
	if err := c.Validate(req); err != nil {
		return h.NewResponseWithError(c, "validate request body failed", err)
	}

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Transfer-Encoding", "chunked")

	onChunk := func(ctx context.Context, dataType, chunk string) error {
		if _, err := c.Response().Write([]byte(chunk)); err != nil {
			return err
		}
		c.Response().Flush()
		return nil
	}
	err := h.usecase.TextCreation(c.Request().Context(), &req, onChunk)
	if err != nil {
		h.logger.Error("text creation failed", log.Error(err))
		return h.NewResponseWithError(c, "text creation failed", err)
	}
	return nil
}

// TabComplete handles tab-based document completion similar to AI coding's FIM (Fill in Middle)
//
//	@Summary		Tab-based document completion
//	@Description	Tab-based document completion similar to AI coding's FIM (Fill in Middle)
//	@Tags			creation
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CompleteReq	true	"tab completion request"
//	@Success		200		{string}	string				"success"
//	@Router			/api/v1/creation/tab-complete [post]
func (h *CreationHandler) TabComplete(c echo.Context) error {
	var req domain.CompleteReq
	if err := c.Bind(&req); err != nil {
		return h.NewResponseWithError(c, "request body is invalid", err)
	}
	if err := c.Validate(req); err != nil {
		return h.NewResponseWithError(c, "validate request body failed", err)
	}

	// For FIM-style completion, we don't need streaming
	result, err := h.usecase.TabComplete(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("tab completion failed", log.Error(err))
		return h.NewResponseWithError(c, "tab completion failed", err)
	}

	return c.JSON(200, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}
