package v1

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/handler"
	"github.com/chaitin/panda-wiki/middleware"
	"github.com/chaitin/panda-wiki/repo/pg"
)

const maxAPITokenNameLength = 100

type APITokenHandler struct {
	*handler.BaseHandler
	repo *pg.APITokenRepo
}

type createAPITokenRequest struct {
	KBID       string                  `json:"kb_id"`
	Name       string                  `json:"name"`
	Permission consts.UserKBPermission `json:"permission"`
}

type updateAPITokenRequest struct {
	ID         string                   `json:"id"`
	KBID       string                   `json:"kb_id"`
	Name       *string                  `json:"name,omitempty"`
	Permission *consts.UserKBPermission `json:"permission,omitempty"`
}

type deleteAPITokenRequest struct {
	ID   string `query:"id"`
	KBID string `query:"kb_id"`
}

type apiTokenListItem struct {
	ID         string                  `json:"id"`
	Name       string                  `json:"name"`
	Token      string                  `json:"token"`
	Permission consts.UserKBPermission `json:"permission"`
}

// NewAPITokenHandler restores the self-hosted implementation of the existing
// admin contract. The /api/pro/v1 path is retained only for compatibility with
// the shipped admin client; it is protected by normal self-hosted KB access
// control rather than a commercial-license gate.
func NewAPITokenHandler(e *echo.Echo, base *handler.BaseHandler, auth middleware.AuthMiddleware, repo *pg.APITokenRepo) *APITokenHandler {
	h := &APITokenHandler{BaseHandler: base, repo: repo}
	group := e.Group("/api/pro/v1/token", auth.Authorize, auth.ValidateKBUserPerm(consts.UserKBPermissionFullControl))
	group.GET("/list", h.List)
	group.POST("/create", h.Create)
	group.PATCH("/update", h.Update)
	group.DELETE("/delete", h.Delete)
	return h
}

func (h *APITokenHandler) List(c echo.Context) error {
	kbID := strings.TrimSpace(c.QueryParam("kb_id"))
	if kbID == "" {
		return h.NewResponseWithError(c, "kb_id is required", nil)
	}
	tokens, err := h.repo.ListByKBID(c.Request().Context(), kbID)
	if err != nil {
		return h.NewResponseWithError(c, "list API tokens failed", err)
	}
	items := make([]apiTokenListItem, 0, len(tokens))
	for _, token := range tokens {
		items = append(items, apiTokenListItem{ID: token.ID, Name: token.Name, Token: token.Token, Permission: token.Permission})
	}
	return h.NewResponseWithData(c, items)
}

func (h *APITokenHandler) Create(c echo.Context) error {
	req, err := decodeCreateAPITokenRequest(c.Request().Body)
	if err != nil {
		return h.NewResponseWithError(c, "invalid API token request", err)
	}
	authInfo := domain.GetAuthInfoFromCtx(c.Request().Context())
	if authInfo == nil || authInfo.IsToken {
		return h.NewResponseWithError(c, "user authentication is required", nil)
	}
	secret, err := generateAPITokenSecret()
	if err != nil {
		return h.NewResponseWithError(c, "generate API token failed", err)
	}
	token := &domain.APIToken{ID: uuid.NewString(), KbId: req.KBID, Name: req.Name, UserID: authInfo.UserId, Token: secret, Permission: req.Permission}
	if err := h.repo.Create(c.Request().Context(), token); err != nil {
		return h.NewResponseWithError(c, "create API token failed", err)
	}
	return h.NewResponseWithData(c, apiTokenListItem{ID: token.ID, Name: token.Name, Token: token.Token, Permission: token.Permission})
}

func (h *APITokenHandler) Update(c echo.Context) error {
	req, err := decodeUpdateAPITokenRequest(c.Request().Body)
	if err != nil {
		return h.NewResponseWithError(c, "invalid API token request", err)
	}
	if err := h.repo.Update(c.Request().Context(), req.KBID, req.ID, req.Name, req.Permission); err != nil {
		return h.NewResponseWithError(c, "update API token failed", err)
	}
	return h.NewResponseWithData(c, nil)
}

func (h *APITokenHandler) Delete(c echo.Context) error {
	var req deleteAPITokenRequest
	if err := c.Bind(&req); err != nil || strings.TrimSpace(req.ID) == "" || strings.TrimSpace(req.KBID) == "" {
		return h.NewResponseWithError(c, "id and kb_id are required", err)
	}
	if err := h.repo.Delete(c.Request().Context(), req.KBID, req.ID); err != nil {
		return h.NewResponseWithError(c, "delete API token failed", err)
	}
	return h.NewResponseWithData(c, nil)
}

func decodeCreateAPITokenRequest(reader io.Reader) (createAPITokenRequest, error) {
	var req createAPITokenRequest
	if err := decodeStrictJSON(reader, &req); err != nil {
		return createAPITokenRequest{}, err
	}
	req.KBID = strings.TrimSpace(req.KBID)
	req.Name = strings.TrimSpace(req.Name)
	if req.KBID == "" || req.Name == "" || len(req.Name) > maxAPITokenNameLength || !validAPITokenPermission(req.Permission) {
		return createAPITokenRequest{}, fmt.Errorf("invalid API token fields")
	}
	return req, nil
}

func decodeUpdateAPITokenRequest(reader io.Reader) (updateAPITokenRequest, error) {
	var req updateAPITokenRequest
	if err := decodeStrictJSON(reader, &req); err != nil {
		return updateAPITokenRequest{}, err
	}
	req.ID = strings.TrimSpace(req.ID)
	req.KBID = strings.TrimSpace(req.KBID)
	if req.ID == "" || req.KBID == "" || (req.Name == nil && req.Permission == nil) {
		return updateAPITokenRequest{}, fmt.Errorf("invalid API token fields")
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" || len(name) > maxAPITokenNameLength {
			return updateAPITokenRequest{}, fmt.Errorf("invalid API token name")
		}
		req.Name = &name
	}
	if req.Permission != nil && !validAPITokenPermission(*req.Permission) {
		return updateAPITokenRequest{}, fmt.Errorf("invalid API token permission")
	}
	return req, nil
}

func decodeStrictJSON(reader io.Reader, target any) error {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func validAPITokenPermission(permission consts.UserKBPermission) bool {
	return permission == consts.UserKBPermissionFullControl || permission == consts.UserKBPermissionDocManage || permission == consts.UserKBPermissionDataOperate
}

func generateAPITokenSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "pw_" + base64.RawURLEncoding.EncodeToString(bytes), nil
}
