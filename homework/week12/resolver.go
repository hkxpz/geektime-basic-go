package week12

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
)

type resolverBuilder struct {
	client   *api.Client
	interval time.Duration
}

func (r *resolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	const serviceName = "service/user"
	res := &consulResolver{client: r.client, target: target, cc: cc, serviceName: serviceName, interval: time.Second}
	return res, res.subscribe(resolver.ResolveNowOptions{})
}

func (r *resolverBuilder) Scheme() string { return "consul" }

func (r *resolverBuilder) SetInterval(interval time.Duration) {
	r.interval = interval
}

type consulResolver struct {
	serviceName string
	client      *api.Client
	target      resolver.Target
	cc          resolver.ClientConn
	interval    time.Duration
	svcs        []*api.ServiceEntry
}

func (c *consulResolver) ResolveNow(options resolver.ResolveNowOptions) {
	if len(c.svcs) < 1 {
		c.cc.ReportError(errors.New("无可用候选节点"))
		return
	}
	addrs := make([]resolver.Address, 0, len(c.svcs))
	for _, service := range c.svcs {
		addrs = append(addrs, resolver.Address{Addr: fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port)})
	}

	if err := c.cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
		// 更新节点失败，一般也做不了什么，记录日志就可以
	}
}

func (c *consulResolver) Close() {}

func (c *consulResolver) subscribe(options resolver.ResolveNowOptions) error {
	go func() {
		var lastIndex uint64
		for {
			svcs, meta, err := c.client.Health().Service(c.serviceName, "", true, &api.QueryOptions{
				WaitIndex: lastIndex, // 设置长轮询
			})
			if err != nil {
				c.cc.ReportError(fmt.Errorf("error retrieving instances from Consul: %w", err))
				time.Sleep(c.interval)
				continue
			}

			// 检查索引是否有变化，有变化表示服务有更新
			if lastIndex != meta.LastIndex {
				// 处理服务变化事件...
				c.svcs = svcs
				c.ResolveNow(options)
			}

			time.Sleep(c.interval) // 短暂休息避免过度请求
		}
	}()
	return nil
}
