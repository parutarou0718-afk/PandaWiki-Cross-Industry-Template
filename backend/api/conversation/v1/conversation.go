package v1

type GetConversationDetailReq struct {
	KbId string `query:"kb_id" json:"kb_id" validate:"required"`
	ID   string `query:"id" json:"id" validate:"required"`
}

type GetConversationDetailResp struct {
}

type GetMessageDetailReq struct {
	KbId string `query:"kb_id" json:"kb_id" validate:"required"`
	ID   string `query:"id" json:"id" validate:"required"`
}

type GetMessageDetailResp struct {
}
