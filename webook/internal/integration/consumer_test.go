package integration

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"

	"geektime-basic-go/webook/internal/integration/startup"
	"geektime-basic-go/webook/ioc"
)

func TestConsumer(t *testing.T) {
	syncProducer(t)
}

func syncProducer(t *testing.T) {
	startup.InitViper()
	producer := ioc.NewSyncProducer(ioc.InitKafka())
	for i := 0; i < 100; i++ {
		_, _, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: "article_read_event",
			Value: sarama.StringEncoder(`{"aid": 1, "uid": 123}`),
		})
		require.NoError(t, err)
	}
}
