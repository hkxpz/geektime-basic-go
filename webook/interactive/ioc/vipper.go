package ioc

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func InitViper() {
	cfile := pflag.String("config", "/etc/webook/config.yaml", "配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func InitViperWithWatchConfig() {
	InitViper()
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
