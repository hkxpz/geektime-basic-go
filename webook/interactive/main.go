package main

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"geektime-basic-go/webook/interactive/events"
	"geektime-basic-go/webook/interactive/ioc"
	"geektime-basic-go/webook/pkg/ginx"
	"geektime-basic-go/webook/pkg/grpcx"
)

func main() {
	ioc.InitViper()
	initPrometheus()
	app := Init()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	go func() {
		panic(app.migratorServer.Start())
	}()

	panic(app.server.Serve())
}

type App struct {
	server         *grpcx.Server
	migratorServer *ginx.Server
	consumers      []events.Consumer
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		fmt.Println(http.ListenAndServe(":8081", nil))
	}()
}
