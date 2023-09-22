package failover

import (
	"context"
	"sync/atomic"
	"time"

	"geektime-basic-go/webook/internal/service/sms"
)

type respTimeService struct {
	svcs sms.Service

	// 响应记录次数
	maxCnt int64
	// 单位相应次数内平均响应时长的涨幅
	maxThreshold int64

	// 当前记录次数
	curcnt int64
	// 当前窗口响应时长
	curWindowsRespTime int64

	// 上一次窗口平均响应时长
	curWindowsAvgRespTime  int64
	lastWindowsAvgRespTime int64
}

func NewRespTimeService(svcs sms.Service, maxCnt, maxThreshold int64) sms.Service {
	return &respTimeService{svcs: svcs, maxCnt: maxCnt, maxThreshold: maxThreshold}
}

func (r *respTimeService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	curWindowsAvgRespTime := atomic.LoadInt64(&r.curWindowsAvgRespTime)
	if curWindowsAvgRespTime > r.lastWindowsAvgRespTime*(1+r.maxThreshold/100) {
		atomic.SwapInt64(&r.lastWindowsAvgRespTime, curWindowsAvgRespTime)
		return sms.ErrServiceProviderException
	}

	now := time.Now()
	err := r.svcs.Send(ctx, tplId, args, numbers...)
	curWindowsRespTime := atomic.LoadInt64(&r.curWindowsRespTime) + time.Since(now).Milliseconds()
	curcnt := atomic.LoadInt64(&r.curcnt)
	if curcnt+1 >= r.maxCnt {
		atomic.SwapInt64(&r.lastWindowsAvgRespTime, curWindowsRespTime/r.maxCnt)
		atomic.SwapInt64(&r.curcnt, 0)
		atomic.SwapInt64(&r.curWindowsRespTime, 0)
		return err
	}

	atomic.AddInt64(&r.curcnt, 1)
	atomic.SwapInt64(&r.curWindowsRespTime, curWindowsRespTime)
	atomic.SwapInt64(&r.curWindowsAvgRespTime, curWindowsRespTime/(curcnt+1))
	return err
}
