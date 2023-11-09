package integration

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"testing"

	events "geektime-basic-go/webook/internal/events/article"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/internal/repository/cache/redis"
	"geektime-basic-go/webook/internal/repository/dao"
	"geektime-basic-go/webook/internal/repository/dao/article"
	"geektime-basic-go/webook/ioc"

	"github.com/stretchr/testify/require"

	myjwt "geektime-basic-go/webook/internal/web/jwt"

	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/internal/integration/startup"
)

func TestArticleHandler_Like(t *testing.T) {
	startup.InitViper()
	db := startup.InitDB()
	ic := redis.NewInteractiveCache(ioc.InitRedis())
	l := startup.InitLog()
	repo := repository.NewInteractiveRepository(ic, dao.NewInteractiveDAO(db), l)
	consumer := events.NewInteractiveLikeEventConsumer(ioc.InitKafka(), repo, l)
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", myjwt.UserClaims{
			ID: rand.Int63n(100000),
		})

		aid := rand.Int63n(50001)
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(`{"id":%d,"like":%t}`, aid, aid%2 == 0))))
	})
	ah := startup.InitArticleHandlerWithKafka(article.NewGormArticleDAO(db))
	ah.RegisterRoutes(server)
	go func() {
		if err := consumer.Start(); err != nil {
			require.NoError(t, err)
		}
	}()
	t.Error(server.Run(":8080"))
}
