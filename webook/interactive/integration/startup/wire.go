//go:build wireinject

package startup

import (
	"github.com/google/wire"

	events "geektime-basic-go/webook/interactive/events/article"
	intrrepo "geektime-basic-go/webook/interactive/repository"
	intrcache "geektime-basic-go/webook/interactive/repository/cache"
	intrdao "geektime-basic-go/webook/interactive/repository/dao"
	intrscv "geektime-basic-go/webook/interactive/service"
)

var thirdProvider = wire.NewSet(
	InitDB,
	InitZapLogger,
	InitRedis,
	InitKafka,
	NewSyncProducer,
)

var interactiveSvcProvider = wire.NewSet(
	intrscv.NewInteractiveService,
	intrrepo.NewInteractiveRepository,
	intrdao.NewInteractiveDAO,
	intrcache.NewInteractiveCache,
)

func InitInteractiveService() intrscv.InteractiveService {
	wire.Build(
		thirdProvider,
		interactiveSvcProvider,
		events.NewChangeLikeSaramaSyncProducer,
	)
	return intrscv.NewInteractiveService(nil, nil, nil)
}
