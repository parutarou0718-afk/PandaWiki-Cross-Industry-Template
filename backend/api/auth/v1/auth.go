package v1

import (
	"time"

	"github.com/chaitin/panda-wiki/consts"
)

type AuthGetReq struct {
	KBID       string            `json:"kb_id,omitempty"  query:"kb_id"`
	SourceType consts.SourceType `query:"source_type"  json:"source_type" validate:"required,oneof=github"`
}

type AuthGetResp struct {
	ClientID     string            `json:"client_id"`
	ClientSecret string            `json:"client_secret"`
	Proxy        string            `json:"proxy"`
	SourceType   consts.SourceType `json:"source_type"`
	Auths        []AuthItem        `json:"auths"`
}

type AuthItem struct {
	ID            uint              `gorm:"primaryKey;column:id" json:"id,omitempty"`
	Username      string            `gorm:"column:username;not null" json:"username,omitempty"`
	AvatarUrl     string            `json:"avatar_url"`
	IP            string            `gorm:"column:ip;not null" json:"ip,omitempty"`
	SourceType    consts.SourceType `gorm:"column:source_type;not null" json:"source_type,omitempty"`
	LastLoginTime time.Time         `gorm:"column:last_login_time" json:"last_login_time,omitempty"`
	CreatedAt     time.Time         `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

type AuthSetReq struct {
	KBID         string            `json:"kb_id,omitempty"`
	SourceType   consts.SourceType `query:"source_type"  json:"source_type" validate:"required,oneof=github"`
	ClientID     string            `json:"client_id"`
	ClientSecret string            `json:"client_secret"`
	Proxy        string            `json:"proxy"`
}

type AuthSetResp struct{}

type AuthDeleteReq struct {
	ID   int64  `query:"id" json:"id"`
	KbID string `query:"kb_id" json:"kb_id"`
}

type AuthDeleteResp struct {
}
