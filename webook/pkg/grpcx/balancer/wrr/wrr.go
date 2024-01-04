// Package wrr 传统版本的基于权重的负载均衡算法
package wrr

import (
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const name = "custom_wrr"

func init() {
	balancer.Register(base.NewBalancerBuilder(name, &PickerBuilder{}, base.Config{}))
}

type PickerBuilder struct{}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, 0, len(info.ReadySCs))
	for sc, sci := range info.ReadySCs {
		cc := &conn{cc: sc}
		md, ok := sci.Address.Metadata.(map[string]any)
		if !ok {
			continue
		}
		weight, _ := md["weight"].(float64)
		cc.weight = int(weight)

		if cc.weight == 0 {
			cc.weight = 10
		}
		cc.currentWeight = cc.weight
		conns = append(conns, cc)
	}
	return &Picker{conns: conns}
}

type Picker struct {
	conns []*conn
	mutex sync.Mutex
}

// Pick 这个才是真的执行负载均衡的地方
func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if len(p.conns) < 1 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var total int
	var maxCC *conn
	for _, cc := range p.conns {
		total += cc.weight
		cc.currentWeight += cc.weight
		if maxCC == nil || cc.currentWeight > maxCC.currentWeight {
			maxCC = cc
		}
	}
	maxCC.currentWeight -= total
	return balancer.PickResult{SubConn: maxCC.cc, Done: func(info balancer.DoneInfo) {}}, nil
}

// conn 代表节点
type conn struct {
	// 权重
	weight        int
	currentWeight int

	//	真正的，grpc 里面的代表一个节点的表达
	cc balancer.SubConn
}
