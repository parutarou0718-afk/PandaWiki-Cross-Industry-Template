package usecase

import (
	"github.com/chaitin/panda-wiki/log"
)

type DingTalkBotUsecase struct {
	logger     *log.Logger
	appUsecase *AppUsecase
}

func NewDingTalkBotUsecase(logger *log.Logger, appUsecase *AppUsecase) *DingTalkBotUsecase {
	return &DingTalkBotUsecase{
		logger:     logger.WithModule("usecase.dingtalk_bot"),
		appUsecase: appUsecase,
	}
}
