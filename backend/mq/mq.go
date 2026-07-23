package mq

import (
	"context"
	"fmt"

	"github.com/google/wire"

	"github.com/chaitin/panda-wiki/config"
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/mq/nats"
	"github.com/chaitin/panda-wiki/mq/types"
)

// Message represents a generic message that can be from either Kafka or NATS
type Message interface {
	GetData() []byte
	GetTopic() string
}

type MQConsumer interface {
	StartConsumerHandlers(ctx context.Context) error
	RegisterHandler(topic string, handler func(ctx context.Context, msg types.Message) error) error
	Close() error
}

type MQProducer interface {
	Produce(ctx context.Context, topic string, key string, value []byte) error
}

func NewMQConsumer(config *config.Config, logger *log.Logger) (MQConsumer, error) {
	if config.MQ.Type == "nats" {
		return nats.NewMQConsumer(logger, config)
	}
	return nil, fmt.Errorf("invalid mq type: %s", config.MQ.Type)
}

func NewMQProducer(config *config.Config, logger *log.Logger) (MQProducer, error) {
	if config.MQ.Type == "nats" {
		return nats.NewMQProducer(config, logger)
	}
	return nil, fmt.Errorf("invalid mq type: %s", config.MQ.Type)
}

var ProviderSet = wire.NewSet(NewMQConsumer, NewMQProducer)
