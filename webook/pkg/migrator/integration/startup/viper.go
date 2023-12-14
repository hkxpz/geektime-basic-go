package startup

import "github.com/spf13/viper"

func InitViper() {
	viper.SetConfigFile("/etc/webook/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
