package mq

import (
	"context"
	"encoding/json"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/mq"
)

type RAGRepository struct {
	producer mq.MQProducer
}

func NewRAGRepository(producer mq.MQProducer) *RAGRepository {
	return &RAGRepository{producer: producer}
}

func (r *RAGRepository) AsyncUpdateNodeReleaseVector(ctx context.Context, request []*domain.NodeReleaseVectorRequest) error {
	for _, req := range request {
		requestBytes, err := json.Marshal(req)
		if err != nil {
			return err
		}
		if err := r.producer.Produce(ctx, domain.VectorTaskTopic, "", requestBytes); err != nil {
			return err
		}
	}
	return nil
}

func (r *RAGRepository) AsyncExtractNodeGraph(ctx context.Context, request []*domain.NodeGraphExtractionRequest) error {
	for _, item := range request {
		requestBytes, err := json.Marshal(item)
		if err != nil {
			return err
		}
		if err := r.producer.Produce(ctx, domain.GraphTaskTopic, "", requestBytes); err != nil {
			return err
		}
	}
	return nil
}
