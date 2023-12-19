package ioc

import (
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"

	"geektime-basic-go/webook/interactive/repository/dao"
	"geektime-basic-go/webook/pkg/ginx"
	"geektime-basic-go/webook/pkg/ginx/handlefunc"
	"geektime-basic-go/webook/pkg/gormx/connpool"
	"geektime-basic-go/webook/pkg/logger"
	"geektime-basic-go/webook/pkg/migrator/events"
	"geektime-basic-go/webook/pkg/migrator/events/fixer"
	"geektime-basic-go/webook/pkg/migrator/scheduler"
)

const topic = "migrator_interactives"

func InitFixDataConsumer(l logger.Logger, src SrcDB, dst DstDB, client sarama.Client) *fixer.Consumer[dao.Interactive] {
	res, err := fixer.NewConsumer[dao.Interactive](client, l, src, dst, topic)
	if err != nil {
		panic(err)
	}
	return res
}

func InitMigratorProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer(p, topic)
}

func InitMigratorWeb(l logger.Logger, src SrcDB, dst DstDB, pool *connpool.DoubleWritePool, producer events.Producer) *ginx.Server {
	gin.SetMode(gin.ReleaseMode)
	web := gin.Default()
	handlefunc.InitCounter(prometheus.CounterOpts{
		Namespace:   "hkxpz",
		Subsystem:   "webook_intr",
		Name:        "http_biz_code",
		Help:        "GIN 中 HTTP 请求",
		ConstLabels: map[string]string{"instance_id": "my-instance-1"},
	})
	intrs := scheduler.NewScheduler[dao.Interactive](l, src, dst, pool, producer)
	intrs.RegisterRoutes(web.Group("/intr"))
	return &ginx.Server{
		Engine: web,
		Addr:   viper.GetString("migrator.http.addr"),
	}
}
