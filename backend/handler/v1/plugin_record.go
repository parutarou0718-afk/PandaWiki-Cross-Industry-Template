package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/handler"
	"github.com/chaitin/panda-wiki/middleware"
	"github.com/chaitin/panda-wiki/usecase"
)

// PluginRecordHandler exposes generic, server-backed plugin records. It is a
// storage and access-control boundary: the server never interprets a plugin's
// payload beyond requiring one bounded JSON object.
type PluginRecordHandler struct {
	*handler.BaseHandler
	usecase *usecase.PluginRecordUsecase
}

type pluginRecordQuery struct {
	KBID       string `query:"kb_id"`
	PluginID   string `query:"plugin_id"`
	RecordType string `query:"record_type"`
}

type pluginRecordWriteRequest struct {
	KBID       string                    `json:"kb_id"`
	PluginID   string                    `json:"plugin_id"`
	RecordType string                    `json:"record_type"`
	Payload    json.RawMessage           `json:"payload"`
	Access     domain.PluginRecordAccess `json:"access"`
}

type pluginRecordResponse struct {
	ID         string                     `json:"id"`
	KBID       string                     `json:"kb_id"`
	PluginID   string                     `json:"plugin_id"`
	RecordType string                     `json:"record_type"`
	Payload    domain.PluginRecordPayload `json:"payload"`
	Access     domain.PluginRecordAccess  `json:"access"`
	CreatedAt  string                     `json:"created_at"`
	UpdatedAt  string                     `json:"updated_at"`
}

func NewPluginRecordHandler(e *echo.Echo, base *handler.BaseHandler, auth middleware.AuthMiddleware, records *usecase.PluginRecordUsecase) *PluginRecordHandler {
	h := &PluginRecordHandler{BaseHandler: base, usecase: records}
	group := e.Group("/api/v1/knowledge_base/plugin-records", auth.Authorize)
	group.GET("", h.List, auth.ValidateKBUserPerm(consts.UserKBPermissionNotNull))
	group.POST("", h.Create, auth.ValidateKBUserPerm(consts.UserKBPermissionNotNull))
	group.PUT("/:id", h.Update, auth.ValidateKBUserPerm(consts.UserKBPermissionNotNull))
	group.DELETE("/:id", h.Delete, auth.ValidateKBUserPerm(consts.UserKBPermissionNotNull))
	group.POST("/:id/restore", h.Restore, auth.ValidateKBUserPerm(consts.UserKBPermissionNotNull))
	return h
}

func (h *PluginRecordHandler) List(c echo.Context) error {
	query, err := bindPluginRecordQuery(c)
	if err != nil {
		return h.NewResponseWithError(c, "invalid plugin record query", err)
	}
	records, err := h.usecase.List(c.Request().Context(), query.KBID, query.PluginID, query.RecordType)
	if err != nil {
		return h.NewResponseWithError(c, "list plugin records failed", err)
	}
	return h.NewResponseWithData(c, mapPluginRecordResponses(records))
}

func (h *PluginRecordHandler) Create(c echo.Context) error {
	input, err := decodePluginRecordWrite(c.Request().Body)
	if err != nil {
		return h.NewResponseWithError(c, "invalid plugin record request", err)
	}
	record, err := h.usecase.Create(c.Request().Context(), input)
	if err != nil {
		return h.NewResponseWithError(c, "create plugin record failed", err)
	}
	return h.NewResponseWithData(c, mapPluginRecordResponse(*record))
}

func (h *PluginRecordHandler) Update(c echo.Context) error {
	input, err := decodePluginRecordWrite(c.Request().Body)
	if err != nil {
		return h.NewResponseWithError(c, "invalid plugin record request", err)
	}
	record, err := h.usecase.Update(c.Request().Context(), c.Param("id"), input)
	if err != nil {
		return h.NewResponseWithError(c, "update plugin record failed", err)
	}
	return h.NewResponseWithData(c, mapPluginRecordResponse(*record))
}

func (h *PluginRecordHandler) Delete(c echo.Context) error {
	query, err := bindPluginRecordQuery(c)
	if err != nil {
		return h.NewResponseWithError(c, "invalid plugin record query", err)
	}
	if err := h.usecase.SoftDelete(c.Request().Context(), query.KBID, query.PluginID, query.RecordType, c.Param("id")); err != nil {
		return h.NewResponseWithError(c, "delete plugin record failed", err)
	}
	return h.NewResponseWithData(c, map[string]bool{"deleted": true})
}

func (h *PluginRecordHandler) Restore(c echo.Context) error {
	input, err := decodePluginRecordIdentity(c.Request().Body)
	if err != nil {
		return h.NewResponseWithError(c, "invalid plugin record restore request", err)
	}
	if err := h.usecase.Restore(c.Request().Context(), input.KBID, input.PluginID, input.RecordType, c.Param("id")); err != nil {
		return h.NewResponseWithError(c, "restore plugin record failed", err)
	}
	return h.NewResponseWithData(c, map[string]bool{"restored": true})
}

func decodePluginRecordIdentity(reader io.Reader) (pluginRecordQuery, error) {
	var request struct {
		KBID       string `json:"kb_id"`
		PluginID   string `json:"plugin_id"`
		RecordType string `json:"record_type"`
	}
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return pluginRecordQuery{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return pluginRecordQuery{}, io.ErrUnexpectedEOF
	}
	query := pluginRecordQuery{KBID: strings.TrimSpace(request.KBID), PluginID: strings.TrimSpace(request.PluginID), RecordType: strings.TrimSpace(request.RecordType)}
	if query.KBID == "" || query.PluginID == "" || query.RecordType == "" {
		return pluginRecordQuery{}, fmt.Errorf("kb_id, plugin_id and record_type are required")
	}
	return query, nil
}

func bindPluginRecordQuery(c echo.Context) (pluginRecordQuery, error) {
	query := pluginRecordQuery{
		KBID:       strings.TrimSpace(c.QueryParam("kb_id")),
		PluginID:   strings.TrimSpace(c.QueryParam("plugin_id")),
		RecordType: strings.TrimSpace(c.QueryParam("record_type")),
	}
	if query.KBID == "" || query.PluginID == "" || query.RecordType == "" {
		return pluginRecordQuery{}, fmt.Errorf("kb_id, plugin_id and record_type are required")
	}
	return query, nil
}

func decodePluginRecordWrite(reader io.Reader) (usecase.PluginRecordWrite, error) {
	var request pluginRecordWriteRequest
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return usecase.PluginRecordWrite{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return usecase.PluginRecordWrite{}, io.ErrUnexpectedEOF
	}
	input := usecase.PluginRecordWrite{
		KBID:       strings.TrimSpace(request.KBID),
		PluginID:   strings.TrimSpace(request.PluginID),
		RecordType: strings.TrimSpace(request.RecordType),
		Payload:    domain.PluginRecordPayload(request.Payload),
		Access:     request.Access,
	}
	record := domain.PluginRecord{ID: "validation", KBID: input.KBID, PluginID: input.PluginID, RecordType: input.RecordType, OwnerUserID: "validation", Payload: input.Payload, Access: input.Access}
	if err := record.Validate(); err != nil {
		return usecase.PluginRecordWrite{}, err
	}
	return input, nil
}

func mapPluginRecordResponses(records []domain.PluginRecord) []pluginRecordResponse {
	response := make([]pluginRecordResponse, 0, len(records))
	for _, record := range records {
		response = append(response, mapPluginRecordResponse(record))
	}
	return response
}

func mapPluginRecordResponse(record domain.PluginRecord) pluginRecordResponse {
	return pluginRecordResponse{
		ID:         record.ID,
		KBID:       record.KBID,
		PluginID:   record.PluginID,
		RecordType: record.RecordType,
		Payload:    record.Payload,
		Access:     record.Access,
		CreatedAt:  record.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:  record.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}
