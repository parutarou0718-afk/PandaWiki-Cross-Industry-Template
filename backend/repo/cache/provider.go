package cache

import (
	"github.com/google/wire"

	"github.com/chaitin/panda-wiki/store/cache"
)

var ProviderSet = wire.NewSet(
	cache.NewCache,
	NewKBRepo,
	NewGeoCache,
)
