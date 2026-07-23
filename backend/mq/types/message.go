package types

// Message represents a generic message that can be from either Kafka or NATS
type Message interface {
	GetData() []byte
	GetTopic() string
}
