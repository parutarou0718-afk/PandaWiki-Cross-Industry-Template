package mq

import (
	"github.com/google/wire"

	"github.com/chaitin/panda-wiki/repo/ipdb"
	"github.com/chaitin/panda-wiki/repo/mq"
	"github.com/chaitin/panda-wiki/repo/pg"
	"github.com/chaitin/panda-wiki/store/rag"
	"github.com/chaitin/panda-wiki/store/s3"
	"github.com/chaitin/panda-wiki/usecase"
)

type MQHandlers struct {
	RAGMQHandler        *RAGMQHandler
	RagDocUpdateHandler *RagDocUpdateHandler
	StatCronHandler     *CronHandler
}

var ProviderSet = wire.NewSet(
	pg.ProviderSet,
	rag.ProviderSet,
	mq.ProviderSet,
	ipdb.ProviderSet,
	s3.ProviderSet,

	usecase.NewLLMUsecase,
	usecase.NewStatUseCase,
	usecase.NewNodeUsecase,
	usecase.NewModelUsecase,
	usecase.NewEditionUsecase,
	wire.Bind(new(usecase.EditionStore), new(*pg.SystemSettingRepo)),
	wire.Bind(new(usecase.PromptStore), new(*pg.PromptRepo)),
	wire.Bind(new(usecase.EditionReader), new(*usecase.EditionUsecase)),

	NewRAGMQHandler,
	NewRagDocUpdateHandler,
	NewCronHandler,

	wire.Struct(new(MQHandlers), "*"),
)
