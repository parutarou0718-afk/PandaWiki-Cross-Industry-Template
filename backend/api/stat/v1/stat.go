package v1

import (
	"github.com/chaitin/panda-wiki/consts"
	"github.com/chaitin/panda-wiki/domain"
)

type StatInstantCountReq struct {
	KbID string `json:"kb_id" query:"kb_id" validate:"required"`
}

type StatInstantPagesReq struct {
	KbID string `json:"kb_id" query:"kb_id" validate:"required"`
}

type StatHotPagesReq struct {
	KbID string         `json:"kb_id" query:"kb_id" validate:"required"`
	Day  consts.StatDay `json:"day" query:"day" validate:"omitempty,oneof=1 7 30 90"`
}

type StatCountReq struct {
	Day  consts.StatDay `json:"day" query:"day" validate:"omitempty,oneof=1 7 30 90"`
	KbID string         `json:"kb_id" query:"kb_id" validate:"required"`
}

type StatCountResp struct {
	IPCount           int64 `json:"ip_count"`
	SessionCount      int64 `json:"session_count"`
	PageVisitCount    int64 `json:"page_visit_count"`
	ConversationCount int64 `json:"conversation_count"`
}

type StatRefererHostsReq struct {
	KbID string         `json:"kb_id" query:"kb_id" validate:"required"`
	Day  consts.StatDay `json:"day" query:"day" validate:"omitempty,oneof=1 7 30 90"`
}

type StatBrowsersReq struct {
	KbID string         `json:"kb_id" query:"kb_id" validate:"required"`
	Day  consts.StatDay `json:"day" query:"day" validate:"omitempty,oneof=1 7 30 90"`
}

type StatGeoCountReq struct {
	KbID string         `json:"kb_id" query:"kb_id" validate:"required"`
	Day  consts.StatDay `json:"day" query:"day" validate:"omitempty,oneof=1 7 30 90"`
}

type StatConversationDistributionReq struct {
	KbID string         `json:"kb_id" query:"kb_id" validate:"required"`
	Day  consts.StatDay `json:"day" query:"day" validate:"omitempty,oneof=1 7 30 90"`
}

type StatConversationDistributionResp struct {
	AppType domain.AppType `json:"app_type"`
	Count   int64          `json:"count"`
}
