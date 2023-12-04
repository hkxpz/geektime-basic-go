package main

import (
	"geektime-basic-go/webook/interactive/events"
	"geektime-basic-go/webook/interactive/ioc"
	"geektime-basic-go/webook/pkg/grpcx"
)

func main() {
	ioc.InitViper()
	app := Init()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	panic(app.server.Serve())
}

type App struct {
	server    *grpcx.Server
	consumers []events.Consumer
}
