package v1

import (
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/middleware"
)

type recordingTokenAuth struct {
	permission consts.UserKBPermission
}

func (a *recordingTokenAuth) Authorize(next echo.HandlerFunc) echo.HandlerFunc { return next }
func (a *recordingTokenAuth) ValidateUserRole(consts.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc { return next }
}
func (a *recordingTokenAuth) ValidateKBUserPerm(permission consts.UserKBPermission) echo.MiddlewareFunc {
	a.permission = permission
	return func(next echo.HandlerFunc) echo.HandlerFunc { return next }
}
func (a *recordingTokenAuth) ValidateLicenseEdition(...consts.LicenseEdition) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc { return next }
}
func (a *recordingTokenAuth) MustGetUserID(echo.Context) (string, bool) { return "", false }

var _ middleware.AuthMiddleware = (*recordingTokenAuth)(nil)

func TestDecodeCreateAPITokenRequestRejectsUnknownFields(t *testing.T) {
	_, err := decodeCreateAPITokenRequest(strings.NewReader(`{
  "kb_id": "kb-1",
  "name": "desktop client",
  "permission": "doc_manage",
  "unexpected": true
}`))

	require.Error(t, err)
}

func TestSelfHostedLicenseResponseUsesEnterpriseCapabilities(t *testing.T) {
	response := selfHostedLicenseResponse()

	require.Equal(t, consts.LicenseEditionEnterprise, response.Edition)
	require.Equal(t, int32(1), response.State)
}

func TestAPITokenRoutesRequireKnowledgeBaseFullControl(t *testing.T) {
	e := echo.New()
	auth := &recordingTokenAuth{}
	NewAPITokenHandler(e, nil, auth, nil)

	require.Equal(t, consts.UserKBPermissionFullControl, auth.permission)
	paths := make(map[string]bool)
	for _, route := range e.Routes() {
		paths[route.Method+" "+route.Path] = true
	}
	require.True(t, paths["GET /api/pro/v1/token/list"])
	require.True(t, paths["POST /api/pro/v1/token/create"])
	require.True(t, paths["PATCH /api/pro/v1/token/update"])
	require.True(t, paths["DELETE /api/pro/v1/token/delete"])
}

func TestGenerateAPITokenSecretProducesOpaqueUniqueValues(t *testing.T) {
	first, err := generateAPITokenSecret()
	require.NoError(t, err)
	second, err := generateAPITokenSecret()
	require.NoError(t, err)

	require.True(t, strings.HasPrefix(first, "pw_"))
	require.NotEqual(t, first, second)
	require.Greater(t, len(first), 32)
}
