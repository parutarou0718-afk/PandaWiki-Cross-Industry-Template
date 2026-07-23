package middleware

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewAuthMiddleware,
	NewShareAuthMiddleware,
	NewReadonlyMiddleware,
	NewSessionMiddleware,
)
