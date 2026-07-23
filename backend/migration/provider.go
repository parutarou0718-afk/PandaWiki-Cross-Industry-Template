package migration

import (
	"github.com/google/wire"

	"github.com/chaitin/panda-wiki/migration/fns"
	"github.com/chaitin/panda-wiki/usecase"
)

var ProviderSet = wire.NewSet(
	// pg.ProviderSet,
	usecase.ProviderSet,
	fns.ProviderSet,

	wire.Struct(new(MigrationFuncs), "*"),

	NewManager,
)
