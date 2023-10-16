package ioc

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB() *mongo.Database {
	cfg := struct {
		DSN          string `yaml:"dsn"`
		PrintCommand bool   `yaml:"print_command"`
	}{}
	if err := viper.UnmarshalKey("db.mongo", &cfg); err != nil {
		panic("初始化MongoDB获取配置失败:" + err.Error())
	}
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			if cfg.PrintCommand {
				fmt.Println(startedEvent.Command)
			}
		},
	}
	opts := options.Client().ApplyURI(cfg.DSN).SetMonitor(monitor)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic("初始化MongoDB连接失败:" + err.Error())
	}
	if err = client.Ping(ctx, nil); err != nil {
		panic("初始化MongoDB连接失败:" + err.Error())
	}
	return client.Database("webook")
}
