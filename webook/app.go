package main

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type App struct {
	web  *gin.Engine
	cron *cron.Cron
}
