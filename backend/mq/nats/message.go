package nats

import (
	"github.com/nats-io/nats.go"

	"github.com/chaitin/panda-wiki/mq/types"
)

type Message struct {
	msg *nats.Msg
}

func (m *Message) GetData() []byte {
	return m.msg.Data
}

func (m *Message) GetTopic() string {
	return m.msg.Subject
}

var _ types.Message = (*Message)(nil)
