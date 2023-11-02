package main

import (
	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/internal/events"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
}
