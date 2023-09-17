package main

import (
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func main() {
	InitViperRemote()
	server := InitWebServer()
	server.GET("/PING", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "PONG")
	})
	log.Fatalln(server.Run(":8081"))
}

func initViper() {
	cfile := pflag.String("config", "config/config.yaml", "配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func initViperWithWatchConfig() {
	initViper()
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println(in.Op)
	})
}

func InitViperRemote() {
	remote := pflag.String("remote", "127.0.0.1", "配置中心地址")
	pflag.Parse()
	viper.SetConfigType("yaml")
	if err := viper.AddRemoteProvider("etcd3", *remote, "/webook"); err != nil {
		panic(err)
	}
	if err := viper.ReadRemoteConfig(); err != nil {
		panic(err)
	}
}
