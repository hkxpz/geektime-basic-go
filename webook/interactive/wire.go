//go:build wireinject

package main

import (
	"github.com/google/wire"

	events "geektime-basic-go/webook/interactive/events/article"
	"geektime-basic-go/webook/interactive/grpc"
	"geektime-basic-go/webook/interactive/ioc"
	intrrepo "geektime-basic-go/webook/interactive/repository"
	intrcache "geektime-basic-go/webook/interactive/repository/cache"
	intrdao "geektime-basic-go/webook/interactive/repository/dao"
	intrscv "geektime-basic-go/webook/interactive/service"
)

var interactiveSvcProvider = wire.NewSet(
	intrscv.NewInteractiveService,
	intrrepo.NewInteractiveRepository,
	intrdao.NewInteractiveDAO,
	intrcache.NewInteractiveCache,
)

var eventsProvider = wire.NewSet(
	ioc.NewSyncProducer,
	events.NewChangeLikeSaramaSyncProducer,
	events.NewInteractiveReadEventConsumer,
	events.NewInteractiveLikeEventConsumer,
)

var thirdProvider = wire.NewSet(
	ioc.InitRedis,
	ioc.InitSRC,
	ioc.InitDST,
	ioc.InitDoubleWritePool,
	ioc.InitBizDB,
	ioc.InitZapLogger,
	ioc.InitKafka,
)

var migratorProvider = wire.NewSet(
	ioc.InitMigratorWeb,
	ioc.InitFixDataConsumer,
	ioc.InitMigratorProducer,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		eventsProvider,
		interactiveSvcProvider,
		migratorProvider,

		grpc.NewInteractiveServiceServer,
		ioc.InitGRPCxServer,
		ioc.NewConsumers,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
