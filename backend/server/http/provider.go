package http

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewEcho,
	wire.Struct(new(HTTPServer), "*"),
)
