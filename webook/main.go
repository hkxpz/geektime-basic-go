package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	server := InitWebServer()
	server.GET("/PING", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "PONG")
	})
	log.Fatalln(server.Run(":8081"))
}
