package usecase

import (
	"context"
	"fmt"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	mqRepo "github.com/chaitin/panda-wiki/repo/mq"
	"github.com/chaitin/panda-wiki/repo/pg"
)

// GraphUsecase owns graph extraction and read access.  It deliberately keeps
// model output and persistence on the server side; clients only receive the
// permission-filtered projection exposed by GraphRepository.
type GraphUsecase struct {
	graphRepo  *pg.GraphRepository
	schemaRepo *pg.KnowledgeSchemaRepository
	nodeRepo   *pg.NodeRepository
	authRepo   *pg.AuthRepo
	llm        *LLMUsecase
	models     *ModelUsecase
	ragRepo    *mqRepo.RAGRepository
	logger     *log.Logger
}

func NewGraphUsecase(graphRepo *pg.GraphRepository, schemaRepo *pg.KnowledgeSchemaRepository, nodeRepo *pg.NodeRepository, authRepo *pg.AuthRepo, llm *LLMUsecase, models *ModelUsecase, ragRepo *mqRepo.RAGRepository, logger *log.Logger) *GraphUsecase {
	return &GraphUsecase{
		graphRepo:  graphRepo,
		schemaRepo: schemaRepo,
		nodeRepo:   nodeRepo,
		authRepo:   authRepo,
		llm:        llm,
		models:     models,
		ragRepo:    ragRepo,
		logger:     logger.WithModule("usecase.graph"),
	}
}

// EnqueueKnowledgeBase rebuilds graph facts through the dedicated graph queue.
// It does not send content through the API, rebuild vectors, or alter nodes.
func (u *GraphUsecase) EnqueueKnowledgeBase(ctx context.Context, kbID string) (int, error) {
	nodes, err := u.nodeRepo.GetList(ctx, &domain.GetNodeListReq{KBID: kbID})
	if err != nil {
		return 0, err
	}
	nodeIDs := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if node.Type != domain.NodeTypeFolder {
			nodeIDs = append(nodeIDs, node.ID)
		}
	}
	if len(nodeIDs) == 0 {
		return 0, nil
	}
	releases, err := u.nodeRepo.GetLatestNodeReleaseByNodeIDs(ctx, kbID, nodeIDs)
	if err != nil {
		return 0, err
	}
	requests := make([]*domain.NodeGraphExtractionRequest, 0, len(releases))
	for _, release := range releases {
		requests = append(requests, &domain.NodeGraphExtractionRequest{KBID: kbID, NodeReleaseID: release.ID})
	}
	if err := u.ragRepo.AsyncExtractNodeGraph(ctx, requests); err != nil {
		return 0, err
	}
	return len(requests), nil
}

func (u *GraphUsecase) RefreshNode(ctx context.Context, kbID, nodeReleaseID string) error {
	nodeRelease, err := u.nodeRepo.GetNodeReleaseWithDirPathByID(ctx, nodeReleaseID)
	if err != nil {
		return fmt.Errorf("get node release: %w", err)
	}
	if nodeRelease.KBID != kbID {
		return domain.ErrPermissionDenied
	}
	if nodeRelease.Type == domain.NodeTypeFolder {
		return nil
	}
	chatModel, err := u.models.GetChatModel(ctx)
	if err != nil {
		return fmt.Errorf("get graph extraction model: %w", err)
	}
	schema, err := u.schemaRepo.GetEffectiveSchema(ctx, kbID)
	if err != nil {
		return fmt.Errorf("get knowledge schema: %w", err)
	}
	extraction, err := u.llm.ExtractGraphFacts(ctx, chatModel, nodeRelease.Name, nodeRelease.Content, schema)
	if err != nil {
		return fmt.Errorf("extract graph facts: %w", err)
	}
	summaries, summaryErr := u.llm.GenerateGraphEntitySummaries(ctx, chatModel, extraction, schema)
	summaryRefreshSucceeded := summaryErr == nil && len(summaries) > 0
	if summaryErr != nil {
		u.logger.Warn("graph entity summary generation failed", log.String("kb_id", kbID), log.String("node_id", nodeRelease.NodeID), log.Int("entity_count", len(extraction.Entities)))
	} else if summaryRefreshSucceeded {
		extraction = attachGraphEntitySummaries(extraction, summaries)
	}
	if err := u.graphRepo.ReplaceNodeExtraction(ctx, kbID, nodeRelease.NodeID, nodeRelease.ID, extraction, pg.ReplaceNodeExtractionOptions{SummaryRefreshSucceeded: summaryRefreshSucceeded}); err != nil {
		return fmt.Errorf("store graph facts: %w", err)
	}
	u.logger.Info("graph extraction completed", log.String("kb_id", kbID), log.String("node_id", nodeRelease.NodeID), log.Int("relations", len(extraction.Relations)))
	return nil
}

func (u *GraphUsecase) GetVisibleGraph(ctx context.Context, kbID string, authUserID uint) (*pg.VisibleGraph, error) {
	groupIDs, err := u.authRepo.GetAuthGroupIdsWithParentsByAuthId(ctx, authUserID)
	if err != nil {
		return nil, err
	}
	graph, err := u.graphRepo.GetVisibleGraph(ctx, kbID, groupIDs)
	if err != nil {
		return nil, err
	}
	schema, err := u.schemaRepo.GetEffectiveSchema(ctx, kbID)
	if err != nil {
		return nil, err
	}
	graph.Schema = schema
	return graph, nil
}

func (u *GraphUsecase) GetSchema(ctx context.Context, kbID string) (domain.KnowledgeSchema, error) {
	return u.schemaRepo.GetEffectiveSchema(ctx, kbID)
}

func (u *GraphUsecase) SaveSchema(ctx context.Context, kbID string, schema domain.KnowledgeSchema) error {
	return u.schemaRepo.SaveSchema(ctx, kbID, schema)
}
