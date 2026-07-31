package v1

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/handler"
)

type LicenseHandler struct {
	*handler.BaseHandler
}

type licenseResponse struct {
	Edition   consts.LicenseEdition `json:"edition"`
	StartedAt int64                 `json:"started_at"`
	ExpiredAt int64                 `json:"expired_at"`
	State     int32                 `json:"state"`
}

func NewLicenseHandler(e *echo.Echo, base *handler.BaseHandler) *LicenseHandler {
	h := &LicenseHandler{BaseHandler: base}
	e.GET("/api/v1/license", h.Get)
	// The self-hosted edition is permanently enabled. These two compatibility
	// endpoints intentionally do not change deployment capability state.
	e.POST("/api/v1/license", h.Get)
	e.DELETE("/api/v1/license", h.Get)
	return h
}

func (h *LicenseHandler) Get(c echo.Context) error {
	return h.NewResponseWithData(c, selfHostedLicenseResponse())
}

func selfHostedLicenseResponse() licenseResponse {
	return licenseResponse{
		Edition:   consts.LicenseEditionEnterprise,
		StartedAt: 0,
		ExpiredAt: time.Date(2100, time.January, 1, 0, 0, 0, 0, time.UTC).Unix(),
		State:     1,
	}
}
