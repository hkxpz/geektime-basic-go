package integration

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"geektime-basic-go/webook/internal/integration/startup"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
	"geektime-basic-go/webook/ioc"
	"geektime-basic-go/webook/pkg/ginx/handlefunc"
)

func TestArticleHandler_Like(t *testing.T) {
	startup.InitViper()
	ioc.InitOTEL()
	gin.SetMode(gin.ReleaseMode)
	handlefunc.InitCounter(prometheus.CounterOpts{
		Namespace:   "hkxpz",
		Subsystem:   "webook",
		Name:        "http_biz_code",
		Help:        "GIN 中 HTTP 请求",
		ConstLabels: map[string]string{"instanceID": "instance-1"},
	})

	server := gin.New()
	server.Use(otelgin.Middleware("webook"))
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", myjwt.UserClaims{
			ID: rand.Int63n(100000),
		})

		aid := rand.Int63n(100)
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(`{"id":%d,"like":true}`, aid))))
	})
	ah := startup.InitArticleHandlerWithKafka()
	ah.RegisterRoutes(server)
	go func() {
		if err := startup.InitInteractiveLikeEventConsumer().Start(); err != nil {
			require.NoError(t, err)
		}
	}()
	t.Error(server.Run(":8080"))
}
