package accesslog

import (
	"bytes"
	"context"
	"io"
	"time"

	"geektime-basic-go/webook/pkg/logger"

	"github.com/gin-gonic/gin"
)

const defaultBodyLength = 1024

type LogFunc func(ctx context.Context, log AccessLog)

type Builder struct {
	f                 LogFunc
	mxReqBodyLength   int
	allowReqBody      bool
	maxRespBodyLength int
	allowRespBody     bool
}

func NewBuilder(logFunc func(ctx context.Context, log AccessLog)) *Builder {
	return &Builder{f: logFunc, mxReqBodyLength: defaultBodyLength, maxRespBodyLength: defaultBodyLength}
}

func (b *Builder) AllowReqBody() *Builder {
	b.allowReqBody = true
	return b
}

func (b *Builder) AllowRespBody() *Builder {
	b.allowRespBody = true
	return b
}

func (b *Builder) SetMaxReqBodyLength(length int) *Builder {
	b.mxReqBodyLength = length
	return b
}

func (b *Builder) SetMaxRespBodyLength(length int) *Builder {
	b.maxRespBodyLength = length
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		log := AccessLog{
			Method: ctx.Request.Method,
			Path:   ctx.Request.URL.Path,
		}

		if b.allowReqBody && ctx.Request.Body != nil {
			rb, _ := ctx.GetRawData()
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(rb))

			if len(rb) > b.mxReqBodyLength {
				log.RespBody = rb[:b.mxReqBodyLength]
			}
			log.ReqBody = rb
		}

		if b.allowRespBody {
			ctx.Writer = respWriter{
				log:            &log,
				ResponseWriter: ctx.Writer,
			}
		}

		defer func() {
			log.Duration = time.Since(start).String()
			b.f(ctx, log)
		}()
		ctx.Next()
	}
}

type AccessLog struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	ReqBody    []byte `json:"req_body"`
	Duration   string `json:"duration"`
	StatusCode int    `json:"status_code"`
	RespBody   []byte `json:"resp_body"`
}

type respWriter struct {
	log *AccessLog
	gin.ResponseWriter
}

func (r respWriter) WriteHeader(statusCode int) {
	r.log.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r respWriter) Write(data []byte) (int, error) {
	r.log.RespBody = data
	return r.ResponseWriter.Write(data)
}

func (r respWriter) WriteString(data string) (int, error) {
	r.log.RespBody = []byte(data)
	return r.ResponseWriter.WriteString(data)
}

func DefaultLogFunc(l logger.Logger) LogFunc {
	return func(ctx context.Context, log AccessLog) {
		// 设置为 DEBUG 级别
		l.Debug("GIN 收到请求", logger.Field{
			Key:   "req",
			Value: log,
		})
	}
}
