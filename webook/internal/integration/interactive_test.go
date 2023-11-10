package integration

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"geektime-basic-go/webook/internal/integration/startup"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
)

func TestArticleHandler_Like(t *testing.T) {
	startup.InitViper()
	gin.SetMode(gin.ReleaseMode)
	server := gin.New()
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
