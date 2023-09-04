package failover

import (
	"context"
	"errors"
	"sync/atomic"

	"geektime-basic-go/webook/internal/service/sms"
)

type timeoutService struct {
	svcs []sms.Service
	idx  uint64

	// 连续超时次数
	cnt uint32

	// 连续超时次数阈值
	threshold uint32
}

func NewTimeoutService(svcs []sms.Service, threshold uint32) sms.Service {
	return &timeoutService{svcs: svcs, threshold: threshold}
}

func (t *timeoutService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	cnt := atomic.LoadUint32(&t.cnt)
	idx := atomic.LoadUint64(&t.idx)
	if cnt >= t.threshold {
		newIdx := (idx + 1) % uint64(len(t.svcs))
		// CAS 操作失败，说明有人切换了，所以你这里不需要检测返回值
		if atomic.CompareAndSwapUint64(&t.idx, idx, newIdx) {
			// 说明你切换了
			atomic.StoreUint32(&t.cnt, 0)
		}
		idx = newIdx
	}

	err := t.svcs[idx].Send(ctx, tplId, args, numbers...)
	switch {
	default:
	case err == nil:
		atomic.StoreUint32(&t.cnt, 0)
	case errors.Is(err, context.DeadlineExceeded):
		atomic.AddUint32(&t.cnt, 1)

	}
	return err
}
