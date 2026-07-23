package usecase

import (
	"context"
	"fmt"
	"strings"

	modelkit "github.com/chaitin/ModelKit/v2/usecase"
	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

type CreationUsecase struct {
	llm      *LLMUsecase
	model    *ModelUsecase
	logger   *log.Logger
	modelkit *modelkit.ModelKit
}

func NewCreationUsecase(logger *log.Logger, llm *LLMUsecase, model *ModelUsecase) *CreationUsecase {
	modelkit := modelkit.NewModelKit(logger.Logger)
	return &CreationUsecase{
		llm:      llm,
		model:    model,
		logger:   logger.WithModule("usecase.creation"),
		modelkit: modelkit,
	}
}

func (u *CreationUsecase) TextCreation(ctx context.Context, req *domain.TextReq, onChunk func(ctx context.Context, dataType, chunk string) error) error {
	model, err := u.model.GetChatModel(ctx)
	if err != nil {
		u.logger.Error("get chat model failed", log.Error(err))
		return domain.ErrModelNotConfigured
	}

	modelkitModel, err := model.ToModelkitModel()
	if err != nil {
		return fmt.Errorf("failed to convert model to modelkit model: %w", err)
	}
	chatModel, err := u.modelkit.GetChatModel(ctx, modelkitModel)
	if err != nil {
		return fmt.Errorf("get chat model failed: %w", err)
	}

	messages := []*schema.Message{
		{
			Role: "system",
			Content: "你是一位专业的文本编辑。你的任务是对输入的文本进行润色和优化。\n\n" +
				"规则：\n" +
				"1. 保持输入文本的原始语言\n" +
				"2. 禁止将文本翻译成其他语言\n" +
				"3. 保持原文的语言风格和表达方式\n\n" +
				"优化方向：\n" +
				"1. 内容优化：\n" +
				"   - 提高文本的清晰度和可读性\n" +
				"   - 确保逻辑流畅和连贯性\n" +
				"   - 保持原文的核心信息和重点\n" +
				"2. 语言优化：\n" +
				"   - 改进语法和句子结构\n" +
				"   - 使语言更加简洁有力\n" +
				"   - 优化用词和表达方式\n\n" +
				"输出要求：\n" +
				"1. 只返回优化后的文本\n" +
				"2. 不要添加任何解释或额外评论\n" +
				"3. 不要改变文本的语言\n" +
				"4. 保持原文的段落结构",
		},
		{
			Role:    "user",
			Content: req.Text,
		},
	}
	usage := &schema.TokenUsage{}
	err = u.llm.ChatWithAgent(ctx, chatModel, messages, usage, onChunk)
	if err != nil {
		return fmt.Errorf("chat with llm failed: %w", err)
	}
	return nil
}

func (u *CreationUsecase) TabComplete(ctx context.Context, req *domain.CompleteReq) (string, error) {
	// For FIM (Fill in Middle) style completion, we need to handle prefix and suffix
	if req.Prefix != "" || req.Suffix != "" {
		model, err := u.model.GetChatModel(ctx)
		if err != nil {
			u.logger.Error("get chat model failed", log.Error(err))
			return "", domain.ErrModelNotConfigured
		}

		modelkitModel, err := model.ToModelkitModel()
		if err != nil {
			return "", fmt.Errorf("failed to convert model to modelkit model: %w", err)
		}
		chatModel, err := u.modelkit.GetChatModel(ctx, modelkitModel)
		if err != nil {
			return "", fmt.Errorf("get chat model failed: %w", err)
		}

		template := prompt.FromMessages(schema.GoTemplate,
			schema.SystemMessage(domain.NodeFIMSystemPrompt),
			schema.UserMessage(domain.NodeFIMFormatter),
		)

		messages, err := template.Format(ctx, map[string]any{
			"Prefix": req.Prefix,
			"Suffix": req.Suffix,
		})
		if err != nil {
			return "", fmt.Errorf("failed to format message: %w", err)
		}

		// For FIM-style completion, we collect the response in a string instead of streaming
		var result strings.Builder
		onChunk := func(ctx context.Context, dataType, chunk string) error {
			result.WriteString(chunk)
			return nil
		}

		usage := &schema.TokenUsage{}
		err = u.llm.ChatWithAgent(ctx, chatModel, messages, usage, onChunk)
		if err != nil {
			return "", fmt.Errorf("chat with llm failed: %w", err)
		}

		completion := result.String()
		return completion, nil
	}
	return "", nil
}
