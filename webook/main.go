package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"geektime-basic-go/webook/ioc"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	ioc.InitViperWatch()
	initPrometheus()
	cancel := ioc.InitOTEL()
	defer cancel(context.Background())

	app := InitApp()
	for _, consumer := range app.consumers {
		if err := consumer.Start(); err != nil {
			panic(err)
		}
	}

	app.cron.Start()
	defer func() {
		<-app.cron.Stop().Done()
	}()

	server := app.web
	server.GET("/PING", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "PONG")
	})

	log.Fatalln(server.Run(":8080"))
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatalln(http.ListenAndServe(":8081", nil))
	}()
}
