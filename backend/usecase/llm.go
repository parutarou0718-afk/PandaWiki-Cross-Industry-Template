package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	modelkit "github.com/chaitin/ModelKit/v2/usecase"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/pkoukk/tiktoken-go"
	"github.com/samber/lo"

	"github.com/chaitin/panda-wiki/config"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/repo/pg"
	"github.com/chaitin/panda-wiki/store/rag"
	"github.com/chaitin/panda-wiki/utils"
)

type LLMUsecase struct {
	rag              rag.RAGService
	conversationRepo *pg.ConversationRepository
	kbRepo           *pg.KnowledgeBaseRepository
	nodeRepo         *pg.NodeRepository
	modelRepo        *pg.ModelRepository
	promptRepo       *pg.PromptRepo
	config           *config.Config
	logger           *log.Logger
	modelkit         *modelkit.ModelKit
}

const (
	summaryChunkTokenLimit  = 30720 // 30KB tokens per chunk
	summaryMaxChunks        = 4     // max chunks to process for summary
	graphExtractionMaxRunes = 60000
)

const graphEntitySummaryPrompt = `Generate optional source-scoped summaries for entities in the sanitized knowledge-graph projection supplied by the user. Return JSON only, with exactly this shape:
{"summaries":[{"name":"entity name","summary":"plain text"}]}
Use only the supplied entity names, types, attributes, relation labels, adjacent entity names, and bounded evidence. Do not add facts, Markdown, explanations, or extra fields. Each summary must be at most 600 Unicode characters. Omit entities when no useful summary can be produced.`

func NewLLMUsecase(config *config.Config, rag rag.RAGService, conversationRepo *pg.ConversationRepository, kbRepo *pg.KnowledgeBaseRepository, nodeRepo *pg.NodeRepository, modelRepo *pg.ModelRepository, promptRepo *pg.PromptRepo, logger *log.Logger) *LLMUsecase {
	tiktoken.SetBpeLoader(&utils.Localloader{})
	modelkit := modelkit.NewModelKit(logger.Logger)
	return &LLMUsecase{
		config:           config,
		rag:              rag,
		conversationRepo: conversationRepo,
		kbRepo:           kbRepo,
		nodeRepo:         nodeRepo,
		modelRepo:        modelRepo,
		promptRepo:       promptRepo,
		logger:           logger.WithModule("usecase.llm"),
		modelkit:         modelkit,
	}
}

func (u *LLMUsecase) BuildConversationMessageWithRAG(
	ctx context.Context,
	conversationID string,
	kbID string,
	groupIDs []int,
	systemPrompt string,
) ([]*schema.Message, []*domain.RankedNodeChunks, error) {
	messages := make([]*schema.Message, 0)
	rankedNodes := make([]*domain.RankedNodeChunks, 0)

	msgs, err := u.conversationRepo.GetConversationMessagesByID(ctx, conversationID)
	if err != nil {
		u.logger.Error("get conversation messages failed", log.Error(err))
		return nil, nil, errors.New("get conversation messages failed")
	}
	if len(msgs) > 0 {
		historyMessages := make([]*schema.Message, 0)
		for _, msg := range msgs {
			switch msg.Role {
			case schema.Assistant:
				historyMessages = append(historyMessages, schema.AssistantMessage(msg.Content, nil))
			case schema.User:
				content := u.formatMessageWithImages(msg.Content, msg.ImagePaths)
				historyMessages = append(historyMessages, schema.UserMessage(content))
			default:
				continue
			}
		}
		if len(historyMessages) > 0 {
			question := historyMessages[len(historyMessages)-1].Content
			var rewrittenQuery string
			if systemPrompt == "" {
				if settingPrompt, err := u.promptRepo.GetPromptContent(ctx, kbID); err != nil {
					u.logger.Error("get prompt from settings failed", log.Error(err))
				} else {
					if settingPrompt != "" {
						systemPrompt = settingPrompt
					} else {
						systemPrompt = domain.SystemDefaultPrompt
					}
				}
			}

			template := prompt.FromMessages(schema.GoTemplate,
				schema.SystemMessage(systemPrompt),
				schema.UserMessage(domain.UserQuestionFormatter),
			)
			kb, err := u.kbRepo.GetKnowledgeBaseByID(ctx, kbID)
			if err != nil {
				u.logger.Error("get kb failed", log.Error(err))
				return nil, nil, errors.New("get kb failed")
			}
			rewrittenQuery, rankedNodes, err = u.GetRankNodes(ctx, GetRankNodesRequest{
				DatasetID:           kb.DatasetID,
				Question:            question,
				GroupIDs:            groupIDs,
				SimilarityThreshold: 0.2,
				HistoryMessages:     historyMessages[:len(historyMessages)-1],
			})
			if err != nil {
				u.logger.Error("get rank nodes failed", log.Error(err))
				return nil, nil, errors.New("get rank nodes failed")
			}
			documents := domain.FormatNodeChunks(rankedNodes, kb.AccessSettings.BaseURL)
			u.logger.Debug("documents", log.String("documents", documents))

			formattedMessages, err := template.Format(ctx, map[string]any{
				"CurrentDate": time.Now().Format("2006-01-02"),
				"Question":    rewrittenQuery,
				"Documents":   documents,
			})
			if err != nil {
				u.logger.Error("format messages failed", log.Error(err))
				return nil, nil, errors.New("format messages failed")
			}
			messages = slices.Insert(formattedMessages, 1, historyMessages[:len(historyMessages)-1]...)
		}
	}
	return messages, rankedNodes, nil
}

func (u *LLMUsecase) ChatWithAgent(
	ctx context.Context,
	chatModel model.BaseChatModel,
	messages []*schema.Message,
	usage *schema.TokenUsage,
	onChunk func(ctx context.Context, dataType, chunk string) error,
) error {
	resp, err := chatModel.Stream(ctx, messages)
	if err != nil {
		return fmt.Errorf("stream failed: %w", err)
	}
	firstReasoning := false
	firstData := false

	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("recv failed: %w", err)
		}
		reasoning, ok := deepseek.GetReasoningContent(msg)
		if ok {
			if !firstReasoning {
				firstReasoning = true
				reasoning = "<think>" + reasoning
			}
			if err := onChunk(ctx, "data", reasoning); err != nil {
				return fmt.Errorf("on chunk reasoning: %w", err)
			}
			continue
		}
		if firstReasoning && !firstData {
			firstData = true
			msg.Content = "</think>\n" + msg.Content
			if err := onChunk(ctx, "data", msg.Content); err != nil {
				return fmt.Errorf("on chunk data: %w", err)
			}
			continue
		}
		if err := onChunk(ctx, "data", msg.Content); err != nil {
			return fmt.Errorf("on chunk data: %w", err)
		}

		// set to usage
		if msg.ResponseMeta.Usage != nil {
			*usage = *msg.ResponseMeta.Usage
		}
	}

	return nil
}

func (u *LLMUsecase) Generate(
	ctx context.Context,
	chatModel model.BaseChatModel,
	messages []*schema.Message,
) (string, error) {
	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("generate failed: %w", err)
	}
	return resp.Content, nil
}

// ExtractGraphFacts asks the configured server-side model for a strictly
// bounded JSON projection.  The source document is not persisted by this
// method and is never returned to a client.
func (u *LLMUsecase) ExtractGraphFacts(ctx context.Context, configuredModel *domain.Model, name, content string, knowledgeSchema domain.KnowledgeSchema) (domain.GraphExtraction, error) {
	modelkitModel, err := configuredModel.ToModelkitModel()
	if err != nil {
		return domain.GraphExtraction{}, err
	}
	chatModel, err := u.modelkit.GetChatModel(ctx, modelkitModel)
	if err != nil {
		return domain.GraphExtraction{}, err
	}
	content = truncateRunes(content, graphExtractionMaxRunes)
	prompt := buildGraphExtractionPrompt(knowledgeSchema)
	result, err := u.Generate(ctx, chatModel, []*schema.Message{
		schema.SystemMessage(prompt),
		schema.UserMessage("Document title: " + name + "\n\nDocument:\n" + content),
	})
	if err != nil {
		return domain.GraphExtraction{}, err
	}
	extraction, err := parseGraphExtractionResponse(result)
	if err != nil {
		return domain.GraphExtraction{}, err
	}
	extraction, discarded, err := domain.SanitizeGraphExtraction(extraction, knowledgeSchema)
	if err != nil {
		return domain.GraphExtraction{}, err
	}
	if discarded > 0 {
		u.logger.Warn("discarded invalid graph extraction facts", log.Int("discarded_count", discarded))
	}
	return extraction, nil
}

// GenerateGraphEntitySummaries performs a separate model call after graph
// extraction has been sanitized. Raw node content and titles are never inputs
// to this call.
func (u *LLMUsecase) GenerateGraphEntitySummaries(ctx context.Context, configuredModel *domain.Model, extraction domain.GraphExtraction, knowledgeSchema domain.KnowledgeSchema) (map[string]string, error) {
	if err := extraction.Validate(); err != nil {
		return nil, err
	}
	for _, entity := range extraction.Entities {
		if err := knowledgeSchema.ValidateEntityAttributes(entity.Type, entity.Attributes); err != nil {
			return nil, err
		}
	}
	modelkitModel, err := configuredModel.ToModelkitModel()
	if err != nil {
		return nil, err
	}
	chatModel, err := u.modelkit.GetChatModel(ctx, modelkitModel)
	if err != nil {
		return nil, err
	}
	return u.requestGraphEntitySummaries(ctx, chatModel, extraction)
}

type graphEntitySummaryProjection struct {
	Entities  []graphEntitySummaryProjectionEntity   `json:"entities"`
	Relations []graphEntitySummaryProjectionRelation `json:"relations"`
}

type graphEntitySummaryProjectionEntity struct {
	Name       string                 `json:"name"`
	Type       domain.GraphEntityType `json:"type"`
	Attributes domain.GraphAttributes `json:"attributes"`
}

type graphEntitySummaryProjectionRelation struct {
	Source   string                   `json:"source"`
	Target   string                   `json:"target"`
	Type     domain.GraphRelationType `json:"type"`
	Evidence string                   `json:"evidence"`
}

type graphEntitySummaryResponse struct {
	Summaries []struct {
		Name    string `json:"name"`
		Summary string `json:"summary"`
	} `json:"summaries"`
}

func (u *LLMUsecase) requestGraphEntitySummaries(ctx context.Context, chatModel model.BaseChatModel, extraction domain.GraphExtraction) (map[string]string, error) {
	projection := graphEntitySummaryProjection{
		Entities:  make([]graphEntitySummaryProjectionEntity, 0, len(extraction.Entities)),
		Relations: make([]graphEntitySummaryProjectionRelation, 0, len(extraction.Relations)),
	}
	for _, entity := range extraction.Entities {
		attributes := entity.Attributes
		if attributes == nil {
			attributes = domain.GraphAttributes{}
		}
		projection.Entities = append(projection.Entities, graphEntitySummaryProjectionEntity{
			Name: entity.Name, Type: entity.Type, Attributes: attributes,
		})
	}
	for _, relation := range extraction.Relations {
		projection.Relations = append(projection.Relations, graphEntitySummaryProjectionRelation{
			Source: relation.Source, Target: relation.Target, Type: relation.Type, Evidence: truncateRunes(strings.TrimSpace(relation.Evidence), domain.MaxGraphEvidenceLength),
		})
	}
	input, err := json.Marshal(projection)
	if err != nil {
		return nil, err
	}
	result, err := u.Generate(ctx, chatModel, []*schema.Message{
		schema.SystemMessage(graphEntitySummaryPrompt),
		schema.UserMessage(string(input)),
	})
	if err != nil {
		return nil, err
	}
	return parseGraphEntitySummaryResponse(result, extraction)
}

func parseGraphEntitySummaryResponse(value string, extraction domain.GraphExtraction) (map[string]string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, errors.New("empty graph entity summary response")
	}
	if strings.HasPrefix(value, "```") {
		value = strings.TrimPrefix(value, "```json")
		value = strings.TrimPrefix(value, "```")
		if end := strings.LastIndex(value, "```"); end >= 0 {
			value = value[:end]
		}
	}
	var response graphEntitySummaryResponse
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(value)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("decode graph entity summaries: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, errors.New("decode graph entity summaries: multiple JSON values")
	}
	knownNames := make(map[string]struct{}, len(extraction.Entities))
	for _, entity := range extraction.Entities {
		knownNames[domain.NormalizeGraphName(entity.Name)] = struct{}{}
	}
	summaries := make(map[string]string, len(response.Summaries))
	for _, item := range response.Summaries {
		nameKey := domain.NormalizeGraphName(item.Name)
		if _, ok := knownNames[nameKey]; !ok {
			continue
		}
		summary := strings.TrimSpace(item.Summary)
		if summary == "" {
			continue
		}
		if _, exists := summaries[nameKey]; exists {
			continue
		}
		summaries[nameKey] = truncateRunes(summary, domain.MaxGraphEntitySummaryLength)
	}
	return summaries, nil
}

func attachGraphEntitySummaries(extraction domain.GraphExtraction, summaries map[string]string) domain.GraphExtraction {
	if len(summaries) == 0 {
		return extraction
	}
	result := extraction
	result.Entities = append([]domain.GraphExtractedEntity(nil), extraction.Entities...)
	for i := range result.Entities {
		summary := strings.TrimSpace(summaries[domain.NormalizeGraphName(result.Entities[i].Name)])
		if summary != "" {
			result.Entities[i].Summary = truncateRunes(summary, domain.MaxGraphEntitySummaryLength)
		}
	}
	return result
}

func buildGraphExtractionPrompt(schema domain.KnowledgeSchema) string {
	type fieldProjection struct {
		Key         string                         `json:"key"`
		EntityTypes []domain.GraphEntityType       `json:"entity_types,omitempty"`
		ValueType   domain.KnowledgeFieldValueType `json:"value_type"`
		Multiple    bool                           `json:"multiple"`
		Options     []string                       `json:"options,omitempty"`
		Instruction string                         `json:"instruction"`
	}
	fields := make([]fieldProjection, 0, len(schema.Fields))
	for _, field := range schema.Fields {
		if !field.Enabled || field.Target != domain.KnowledgeFieldTargetEntity {
			continue
		}
		fields = append(fields, fieldProjection{Key: field.Key, EntityTypes: field.EntityTypes, ValueType: field.ValueType, Multiple: field.Multiple, Options: field.Options, Instruction: field.ExtractInstruction})
	}
	fieldJSON, _ := json.Marshal(fields)
	return `Extract a compact knowledge graph from the document below. Return JSON only, with exactly this shape:
{"entities":[{"name":"string","type":"person|organization|concept|method|event|document|other","attributes":{"field_key":["value"]}}],"relations":[{"source":"string","target":"string","type":"mentions|related_to|part_of|causes|contradicts|cites","confidence":0.0,"evidence":"short source excerpt"}]}
Use only facts supported by the document. Keep evidence at most 512 characters. The enabled field list below is exhaustive: attributes may contain only those exact keys. If it is empty or a value is unavailable, return an empty attributes object. Never invent semantic keys such as "research". Every attribute value must be an array; omit unavailable fields. Do not include summaries, Markdown, explanations, document content outside evidence, or extra fields.
Enabled field definitions: ` + string(fieldJSON)
}

func parseGraphExtractionResponse(value string) (domain.GraphExtraction, error) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "```") {
		value = strings.TrimPrefix(value, "```json")
		value = strings.TrimPrefix(value, "```")
		if end := strings.LastIndex(value, "```"); end >= 0 {
			value = value[:end]
		}
	}
	extraction, err := domain.DecodeGraphExtraction(strings.TrimSpace(value))
	if err != nil {
		return domain.GraphExtraction{}, err
	}
	// Summaries from the raw-document call are never trusted or persisted. Only
	// the separate sanitized-projection call may populate this field.
	for i := range extraction.Entities {
		extraction.Entities[i].Summary = ""
	}
	return extraction, nil
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func (u *LLMUsecase) SummaryNode(ctx context.Context, kbID string, model *domain.Model, name, content string) (string, error) {
	modelkitModel, err := model.ToModelkitModel()
	if err != nil {
		return "", err
	}
	chatModel, err := u.modelkit.GetChatModel(ctx, modelkitModel)
	if err != nil {
		return "", err
	}

	chunks, err := u.SplitByTokenLimit(content, summaryChunkTokenLimit)
	if err != nil {
		return "", err
	}
	if len(chunks) > summaryMaxChunks {
		u.logger.Debug("trim summary chunks for large document", log.String("node", name), log.Int("original_chunks", len(chunks)), log.Int("used_chunks", summaryMaxChunks))
		chunks = chunks[:summaryMaxChunks]
	}

	summaries := make([]string, 0, len(chunks))
	for idx, chunk := range chunks {
		summary, err := u.requestSummary(ctx, kbID, chatModel, name, chunk)
		if err != nil {
			u.logger.Error("Failed to generate summary for chunk", log.Int("chunk_index", idx), log.Error(err))
			continue
		}
		if summary == "" {
			u.logger.Warn("Empty summary returned for chunk", log.Int("chunk_index", idx))
			continue
		}
		summaries = append(summaries, summary)
	}

	if len(summaries) == 0 {
		return "", fmt.Errorf("failed to generate summary for document %s", name)
	}
	if len(summaries) == 1 {
		return summaries[0], nil
	}

	// Join all summaries and generate final summary
	joined := strings.Join(summaries, "\n\n")
	finalSummary, err := u.requestSummary(ctx, kbID, chatModel, name, joined)
	if err != nil {
		u.logger.Error("Failed to generate final summary, using aggregated summaries", log.Error(err))
		// Fallback: return the joined summaries directly
		if len(joined) > 500 {
			return joined[:500] + "...", nil
		}
		return joined, nil
	}
	return finalSummary, nil
}

func (u *LLMUsecase) StreamSummaryNode(
	ctx context.Context,
	kbID string,
	model *domain.Model,
	name, content string,
	onChunk func(ctx context.Context, dataType, chunk string) error,
) error {
	modelkitModel, err := model.ToModelkitModel()
	if err != nil {
		return err
	}
	chatModel, err := u.modelkit.GetChatModel(ctx, modelkitModel)
	if err != nil {
		return err
	}

	chunks, err := u.SplitByTokenLimit(content, summaryChunkTokenLimit)
	if err != nil {
		return err
	}
	if len(chunks) > summaryMaxChunks {
		u.logger.Debug("trim summary chunks for large document", log.String("node", name), log.Int("original_chunks", len(chunks)), log.Int("used_chunks", summaryMaxChunks))
		chunks = chunks[:summaryMaxChunks]
	}

	if len(chunks) == 1 {
		return u.streamSummary(ctx, kbID, chatModel, name, chunks[0], onChunk)
	}

	summaries := make([]string, 0, len(chunks))
	for idx, chunk := range chunks {
		summary, summaryErr := u.requestSummary(ctx, kbID, chatModel, name, chunk)
		if summaryErr != nil {
			u.logger.Error("Failed to generate summary for chunk", log.Int("chunk_index", idx), log.Error(summaryErr))
			continue
		}
		if summary == "" {
			u.logger.Warn("Empty summary returned for chunk", log.Int("chunk_index", idx))
			continue
		}
		summaries = append(summaries, summary)
	}

	if len(summaries) == 0 {
		return fmt.Errorf("failed to generate summary for document %s", name)
	}
	if len(summaries) == 1 {
		if err := onChunk(ctx, "data", summaries[0]); err != nil {
			return fmt.Errorf("on chunk data: %w", err)
		}
		return nil
	}

	joined := strings.Join(summaries, "\n\n")
	if err := u.streamSummary(ctx, kbID, chatModel, name, joined, onChunk); err != nil {
		u.logger.Error("Failed to generate final summary, using aggregated summaries", log.Error(err))
		if len(joined) > 500 {
			joined = joined[:500] + "..."
		}
		if chunkErr := onChunk(ctx, "data", joined); chunkErr != nil {
			return fmt.Errorf("on chunk data: %w", chunkErr)
		}
	}
	return nil
}

func (u *LLMUsecase) trimThinking(summary string) string {
	if !strings.HasPrefix(summary, "<think>") {
		return summary
	}
	endIndex := strings.Index(summary, "</think>")
	if endIndex == -1 {
		return summary
	}
	return strings.TrimSpace(summary[endIndex+len("</think>"):])
}

func (u *LLMUsecase) requestSummary(ctx context.Context, kbID string, chatModel model.BaseChatModel, name, content string) (string, error) {
	summaryPrompt, err := u.promptRepo.GetSummaryPrompt(ctx, kbID)
	if err != nil {
		return "", err
	}

	summary, err := u.Generate(ctx, chatModel, []*schema.Message{
		{
			Role:    "system",
			Content: summaryPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("文档名称：%s\n文档内容：%s", name, content),
		},
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(u.trimThinking(summary)), nil
}

func (u *LLMUsecase) streamSummary(
	ctx context.Context,
	kbID string,
	chatModel model.BaseChatModel,
	name, content string,
	onChunk func(ctx context.Context, dataType, chunk string) error,
) error {
	summaryPrompt, err := u.promptRepo.GetSummaryPrompt(ctx, kbID)
	if err != nil {
		return err
	}

	usage := schema.TokenUsage{}
	filter := newThinkingStreamFilter()
	return u.ChatWithAgent(ctx, chatModel, []*schema.Message{
		{
			Role:    "system",
			Content: summaryPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("文档名称：%s\n文档内容：%s", name, content),
		},
	}, &usage, func(ctx context.Context, dataType, chunk string) error {
		if dataType != "data" {
			return onChunk(ctx, dataType, chunk)
		}
		cleaned := filter.Append(chunk)
		if cleaned == "" {
			return nil
		}
		return onChunk(ctx, dataType, cleaned)
	})
}

type thinkingStreamFilter struct {
	buffer strings.Builder
	done   bool
}

func newThinkingStreamFilter() *thinkingStreamFilter {
	return &thinkingStreamFilter{}
}

func (f *thinkingStreamFilter) Append(chunk string) string {
	if f.done {
		return chunk
	}
	f.buffer.WriteString(chunk)
	content := f.buffer.String()
	if !strings.HasPrefix(content, "<think>") {
		f.done = true
		f.buffer.Reset()
		return content
	}
	endIndex := strings.Index(content, "</think>")
	if endIndex == -1 {
		return ""
	}
	cleaned := strings.TrimSpace(content[endIndex+len("</think>"):])
	f.done = true
	f.buffer.Reset()
	return cleaned
}

func (u *LLMUsecase) SplitByTokenLimit(text string, maxTokens int) ([]string, error) {
	if maxTokens <= 0 {
		return nil, fmt.Errorf("maxTokens must be greater than 0")
	}
	encoding, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("failed to get encoding: %w", err)
	}
	tokens := encoding.Encode(text, nil, nil)
	if len(tokens) <= maxTokens {
		return []string{text}, nil
	}

	// 预先计算需要的片段数量并分配空间
	numChunks := (len(tokens) + maxTokens - 1) / maxTokens // 向上取整
	result := make([]string, 0, numChunks)

	for i := 0; i < len(tokens); i += maxTokens {
		end := i + maxTokens
		if end > len(tokens) {
			end = len(tokens)
		}

		chunk := tokens[i:end]
		decodedChunk := encoding.Decode(chunk)
		result = append(result, decodedChunk)
	}

	return result, nil
}

type GetRankNodesRequest struct {
	DatasetID           string
	Question            string
	GroupIDs            []int
	SimilarityThreshold float64
	HistoryMessages     []*schema.Message
	MaxChunksPerDoc     int
}

func (u *LLMUsecase) GetRankNodes(ctx context.Context, req GetRankNodesRequest) (string, []*domain.RankedNodeChunks, error) {
	var rankedNodes []*domain.RankedNodeChunks
	// get related documents from raglite
	rewrittenQuery, records, err := u.rag.QueryRecords(ctx, &rag.QueryRecordsRequest{
		DatasetID:           req.DatasetID,
		Query:               req.Question,
		GroupIDs:            req.GroupIDs,
		SimilarityThreshold: req.SimilarityThreshold,
		HistoryMsgs:         req.HistoryMessages,
		MaxChunksPerDoc:     req.MaxChunksPerDoc,
	})
	if err != nil {
		return "", nil, fmt.Errorf("get records from raglite failed: %w", err)
	}
	u.logger.Info("get related documents from raglite", log.Any("record_count", len(records)))
	rankedNodesMap := make(map[string]*domain.RankedNodeChunks)
	// get raw node by doc_id
	if len(records) > 0 {
		docIDs := lo.Uniq(lo.Map(records, func(item *domain.NodeContentChunk, _ int) string {
			return item.DocID
		}))
		u.logger.Info("node chunk doc ids", log.Any("docIDs", docIDs))
		docIDNode, err := u.nodeRepo.GetNodeReleasesWithPathsByDocIDs(ctx, docIDs)
		if err != nil {
			return "", nil, fmt.Errorf("get nodes by ids failed: %w", err)
		}
		u.logger.Info("get node release by doc ids", log.Any("docIDNode", lo.Keys(docIDNode)))
		for _, record := range records {
			if nodeChunk, ok := rankedNodesMap[record.DocID]; !ok {
				if docNode, ok := docIDNode[record.DocID]; ok {
					rankNodeChunk := &domain.RankedNodeChunks{
						NodeID:        docNode.NodeID,
						NodeName:      docNode.Name,
						NodeSummary:   docNode.Meta.Summary,
						NodeEmoji:     docNode.Meta.Emoji,
						NodePathNames: docNode.PathNames,
						Chunks:        []*domain.NodeContentChunk{record},
					}
					rankedNodes = append(rankedNodes, rankNodeChunk)
					rankedNodesMap[record.DocID] = rankNodeChunk
				}
			} else {
				nodeChunk.Chunks = append(nodeChunk.Chunks, record)
			}
		}
	}
	return rewrittenQuery, rankedNodes, nil
}

// formatMessageWithImages converts image paths to markdown format and appends to message
func (u *LLMUsecase) formatMessageWithImages(message string, imagePaths []string) string {
	if len(imagePaths) == 0 {
		return message
	}
	var builder strings.Builder
	builder.WriteString(message)
	for _, path := range imagePaths {
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("![](%s)", path))
	}
	return builder.String()
}
