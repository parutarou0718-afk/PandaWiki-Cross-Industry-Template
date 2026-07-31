package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/chaitin/panda-wiki/config"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
)

type MQProducer struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	logger *log.Logger
}

type managedStreamDefinition struct {
	name     string
	subjects []string
}

// managedStreamDefinitions is the single source of truth for subjects the
// application expects JetStream to retain. Keep new task subjects here so an
// existing deployment can be upgraded without manually recreating NATS data.
func managedStreamDefinitions() []managedStreamDefinition {
	return []managedStreamDefinition{
		{
			name: "task",
			subjects: []string{
				"apps.panda-wiki.summary.task",
				domain.VectorTaskTopic,
				domain.GraphTaskTopic,
			},
		},
		{
			name:     "scraper",
			subjects: []string{"apps.panda-wiki.scraper.>"},
		},
	}
}

func mergeStreamSubjects(current, required []string) []string {
	merged := append([]string(nil), current...)
	for _, subject := range required {
		found := false
		for _, existing := range merged {
			if existing == subject {
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, subject)
		}
	}
	return merged
}

func (p *MQProducer) EnsureStreams() error {
	streams := managedStreamDefinitions()

	for _, stream := range streams {
		info, err := p.js.StreamInfo(stream.name)
		if err == nil {
			subjects := mergeStreamSubjects(info.Config.Subjects, stream.subjects)
			if len(subjects) != len(info.Config.Subjects) {
				config := info.Config
				config.Subjects = subjects
				if _, err = p.js.UpdateStream(&config); err != nil {
					return fmt.Errorf("failed to update stream %s subjects: %w", stream.name, err)
				}
				p.logger.Info("updated stream subjects", log.String("stream", stream.name), log.Any("subjects", subjects))
			}
			p.logger.Debug("stream already exists",
				log.String("stream", stream.name))
			continue
		}

		// Stream doesn't exist, create it
		_, err = p.js.AddStream(&nats.StreamConfig{
			Name:       stream.name,
			Subjects:   stream.subjects,
			Storage:    nats.FileStorage,
			Retention:  nats.LimitsPolicy,
			Discard:    nats.DiscardOld,
			MaxAge:     7 * 24 * time.Hour,
			MaxBytes:   1 * 1024 * 1024 * 1024,
			MaxMsgs:    1000000,
			MaxMsgSize: 50 * 1024 * 1024,
			Replicas:   1,
			Duplicates: 120 * time.Second,
		})
		if err != nil {
			return fmt.Errorf("failed to create stream %s: %w", stream.name, err)
		}

		p.logger.Info("created stream",
			log.String("stream", stream.name),
			log.Any("subjects", stream.subjects))
	}

	return nil
}

func NewMQProducer(config *config.Config, logger *log.Logger) (*MQProducer, error) {
	opts := []nats.Option{
		nats.Name("panda-wiki"),
	}

	if user := config.MQ.NATS.User; user != "" {
		opts = append(opts, nats.UserInfo(user, config.MQ.NATS.Password))
	}

	server := config.MQ.NATS.Server

	conn, err := nats.Connect(server, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}

	producer := &MQProducer{
		conn:   conn,
		js:     js,
		logger: logger,
	}

	// Ensure streams exist
	if err := producer.EnsureStreams(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ensure streams: %w", err)
	}

	return producer, nil
}

func (p *MQProducer) Produce(ctx context.Context, topic string, key string, value []byte) error {
	p.logger.Debug("publishing message",
		log.String("topic", topic),
		log.String("key", key),
		log.Int("value_size", len(value)))

	_, err := p.js.Publish(topic, value)
	if err != nil {
		p.logger.Error("failed to publish message",
			log.String("topic", topic),
			log.Error(err))
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Debug("message published successfully",
		log.String("topic", topic))
	return nil
}

func (p *MQProducer) Close() error {
	p.conn.Close()
	return nil
}
