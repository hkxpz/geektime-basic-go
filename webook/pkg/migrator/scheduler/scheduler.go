package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"geektime-basic-go/webook/pkg/ginx/handlefunc"
	"geektime-basic-go/webook/pkg/gormx/connpool"
	"geektime-basic-go/webook/pkg/logger"
	"geektime-basic-go/webook/pkg/migrator"
	"geektime-basic-go/webook/pkg/migrator/events"
	"geektime-basic-go/webook/pkg/migrator/validator"
)

type Scheduler[T migrator.Entity] struct {
	lock       sync.Mutex
	src        *gorm.DB
	dst        *gorm.DB
	pool       *connpool.DoubleWritePool
	l          logger.Logger
	pattern    string
	cancelFull func()
	cancelIncr func()
	producer   events.Producer
}

func NewScheduler[T migrator.Entity](l logger.Logger, src *gorm.DB, dst *gorm.DB, pool *connpool.DoubleWritePool, producer events.Producer) *Scheduler[T] {
	return &Scheduler[T]{
		l:          l,
		src:        src,
		dst:        dst,
		pattern:    connpool.PatternSrcOnly,
		cancelFull: func() {},
		cancelIncr: func() {},
		pool:       pool,
		producer:   producer,
	}
}

func (s *Scheduler[T]) RegisterRoutes(server *gin.RouterGroup) {
	server.POST("/src_only", handlefunc.Wrap(s.SrcOnly))
	server.POST("/src_first", handlefunc.Wrap(s.SrcFirst))
	server.POST("/dst_only", handlefunc.Wrap(s.DstOnly))
	server.POST("/dst_first", handlefunc.Wrap(s.DstFirst))
	server.POST("/full/start", handlefunc.Wrap(s.StartFullValidation))
	server.POST("/full/stop", handlefunc.Wrap(s.StopFullValidation))
	server.POST("/incr/start", handlefunc.WrapReq[StartIncrRequest](s.StartIncrementValidation))
	server.POST("/incr/stop", handlefunc.Wrap(s.StopIncrementValidation))
}

func (s *Scheduler[T]) SrcOnly(c *gin.Context) (handlefunc.Response, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcOnly
	s.pool.ChangePattern(s.pattern)
	return handlefunc.Response{Msg: "OK"}, nil
}

func (s *Scheduler[T]) SrcFirst(c *gin.Context) (handlefunc.Response, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcFirst
	s.pool.ChangePattern(s.pattern)
	return handlefunc.Response{Msg: "OK"}, nil
}

func (s *Scheduler[T]) DstOnly(c *gin.Context) (handlefunc.Response, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstOnly
	s.pool.ChangePattern(s.pattern)
	return handlefunc.Response{Msg: "OK"}, nil
}

func (s *Scheduler[T]) DstFirst(c *gin.Context) (handlefunc.Response, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstFirst
	s.pool.ChangePattern(s.pattern)
	return handlefunc.Response{Msg: "OK"}, nil
}

func (s *Scheduler[T]) StartFullValidation(c *gin.Context) (handlefunc.Response, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	cancel := s.cancelFull
	v, err := s.newValidator()
	if err != nil {
		return handlefunc.InternalServerError(), err
	}
	var ctx context.Context
	ctx, s.cancelFull = context.WithCancel(context.Background())
	go func() {
		cancel()
		if err = v.Validate(ctx); err != nil {
			s.l.Error("异常退出全量校验", logger.Error(err))
		}
		s.l.Warn("退出全量校验")
	}()
	return handlefunc.Response{Msg: "OK"}, err
}

func (s *Scheduler[T]) StopFullValidation(c *gin.Context) (handlefunc.Response, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelFull()
	return handlefunc.Response{Msg: "OK"}, nil
}

func (s *Scheduler[T]) StartIncrementValidation(c *gin.Context, req StartIncrRequest) (handlefunc.Response, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	cancel := s.cancelFull
	v, err := s.newValidator()
	if err != nil {
		return handlefunc.InternalServerError(), err
	}
	v.SleepInterval(time.Duration(req.Interval) * time.Millisecond).UpdateAt(req.UpdateAt)
	var ctx context.Context
	ctx, s.cancelIncr = context.WithCancel(context.Background())
	go func() {
		// 先取消上次的校验
		cancel()
		if err = v.Validate(ctx); err != nil {
			s.l.Error("异常退出增量校验", logger.Error(err))
		}
		s.l.Warn("退出增量校验")
	}()
	return handlefunc.Response{Msg: "启动增量校验成功"}, nil
}

func (s *Scheduler[T]) StopIncrementValidation(c *gin.Context) (handlefunc.Response, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelIncr()
	return handlefunc.Response{Msg: "OK"}, nil
}

func (s *Scheduler[T]) newValidator() (*validator.GORMValidator[T], error) {
	switch s.pattern {
	case connpool.PatternSrcOnly, connpool.PatternSrcFirst:
		return validator.NewValidator[T](s.src, s.dst, "src", s.l, s.producer), nil
	case connpool.PatternDstOnly, connpool.PatternDstFirst:
		return validator.NewValidator[T](s.dst, s.src, "dst", s.l, s.producer), nil
	default:
		return nil, fmt.Errorf("未知的 pattern %s", s.pattern)
	}
}

type StartIncrRequest struct {
	UpdateAt int64 `json:"update_at"`
	Interval int64 `json:"interval"`
}
