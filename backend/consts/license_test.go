package consts

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestGetLicenseEditionUsesSelfHostedCompatibilityEdition(t *testing.T) {
	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	c.Set("edition", LicenseEditionFree)

	require.Equal(t, LicenseEditionEnterprise, GetLicenseEdition(c))
}
