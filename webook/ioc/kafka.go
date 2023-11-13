package ioc

import (
	"fmt"

	"geektime-basic-go/webook/internal/events"
	"geektime-basic-go/webook/internal/events/article"

	"github.com/IBM/sarama"
	"github.com/spf13/viper"
)

func InitKafka() sarama.Client {
	type config struct {
		Addrs []string `yaml:"addrs"`
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true

	var cfg config
	if err := viper.UnmarshalKey("kafka", &cfg); err != nil {
		panic(fmt.Sprintf("初始化 kafka 失败, 反序列化配置失败: %s", err))
	}
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
	if err != nil {
		panic(fmt.Sprintf("初始化 kafka 失败: %s", err))
	}
	return client
}

func NewSyncProducer(client sarama.Client) sarama.SyncProducer {
	res, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return res
}

// NewConsumers 面临的问题依旧是所有的 Consumer 在这里注册一下
func NewConsumers(c1 *article.InteractiveReadEventConsumer, c2 *article.ChangeLikeEventConsumer) []events.Consumer {
	return []events.Consumer{c1, c2}
}
