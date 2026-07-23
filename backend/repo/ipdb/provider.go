package ipdb

import (
	"github.com/google/wire"

	ipdbStore "github.com/chaitin/panda-wiki/store/ipdb"
)

var ProviderSet = wire.NewSet(
	ipdbStore.NewIPDB,

	NewIPAddressRepo,
)
