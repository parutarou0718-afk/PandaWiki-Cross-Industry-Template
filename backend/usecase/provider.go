package usecase

import (
	"github.com/google/wire"

	"github.com/chaitin/panda-wiki/repo/ipdb"
	mqRepo "github.com/chaitin/panda-wiki/repo/mq"
	"github.com/chaitin/panda-wiki/repo/pg"
	"github.com/chaitin/panda-wiki/store/rag"
	"github.com/chaitin/panda-wiki/store/s3"
)

var ProviderSet = wire.NewSet(
	pg.ProviderSet,
	mqRepo.ProviderSet,
	ipdb.ProviderSet,
	rag.ProviderSet,
	s3.ProviderSet,

	NewLLMUsecase,
	NewNodeUsecase,
	NewAppUsecase,
	NewConversationUsecase,
	NewUserUsecase,
	NewModelUsecase,
	NewKnowledgeBaseUsecase,
	NewChatUsecase,
	NewCrawlerUsecase,
	NewCreationUsecase,
	NewFileUsecase,
	NewSitemapUsecase,
	NewStatUseCase,
	NewCommentUsecase,
	NewWechatUsecase,
	NewWecomUsecase,
	NewWechatAppUsecase,
	NewAuthUsecase,
	NewNavUsecase,
	NewEditionUsecase,
	NewReportUsecase,
	NewReportProfileAdminUsecase,
	wire.Bind(new(reportProfileStore), new(*pg.ReportProfileRepo)),
	wire.Bind(new(reportStore), new(*pg.ReportRepo)),
	wire.Bind(new(reportAuthStore), new(*pg.AuthRepo)),
	wire.Bind(new(reportKnowledgeBaseStore), new(*pg.KnowledgeBaseRepository)),
	wire.Bind(new(reportLLM), new(*LLMUsecase)),
	wire.Bind(new(reportModelProvider), new(*ModelUsecase)),
	wire.Bind(new(reportEditionReader), new(*EditionUsecase)),
	wire.Bind(new(EditionStore), new(*pg.SystemSettingRepo)),
	wire.Bind(new(PromptStore), new(*pg.PromptRepo)),
	wire.Bind(new(EditionReader), new(*EditionUsecase)),
)
