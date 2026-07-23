//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/chaitin/panda-wiki/config"
	"github.com/chaitin/panda-wiki/log"
	"github.com/chaitin/panda-wiki/migration"
)

func createApp() (*App, error) {
	wire.Build(
		wire.Struct(new(App), "*"),
		wire.NewSet(
			config.ProviderSet,
			log.ProviderSet,
			migration.ProviderSet,
		),
	)
	return &App{}, nil
}

type App struct {
	Config           *config.Config
	MigrationManager *migration.Manager
}
