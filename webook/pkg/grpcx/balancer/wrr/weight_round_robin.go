// Package wrr 传统版本的基于权重的负载均衡算法
package wrr

import (
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const WeightRoundRobin = "custom_weighted_round_robin"

func init() {
	balancer.Register(newBuilder())
}

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(WeightRoundRobin, &WeightedPickerBuilder{}, base.Config{})
}

type WeightedPickerBuilder struct{}

func (p *WeightedPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*weightConn, 0, len(info.ReadySCs))
	for conn, connInfo := range info.ReadySCs {
		weightVal, _ := connInfo.Address.Metadata.(map[string]any)["weight"]
		weight, _ := weightVal.(float64)
		if weight == 0 {
			weight = 10
		}
		conns = append(conns, &weightConn{
			cc:            conn,
			weight:        int(weight),
			currentWeight: int(weight),
		})
	}
	return &WeightedPicker{conns: conns}
}

type WeightedPicker struct {
	conns []*weightConn
	mutex sync.Mutex
}

// Pick 这个才是真的执行负载均衡的地方
func (p *WeightedPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(p.conns) < 1 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()
	var totalWeight int
	var res *weightConn
	for _, node := range p.conns {
		totalWeight += node.weight
		node.currentWeight += node.weight
		if res == nil || node.currentWeight > res.currentWeight {
			res = node
		}
	}
	res.currentWeight -= totalWeight
	return balancer.PickResult{SubConn: res.cc, Done: p.Done}, nil
}

// Done 在这里执行 failover 有关的事情
func (p *WeightedPicker) Done(info balancer.DoneInfo) { /* todo 动态修改权重 */ }

// weightConn 代表节点
type weightConn struct {
	// 权重
	weight        int
	currentWeight int

	//	真正的，grpc 里面的代表一个节点的表达
	cc balancer.SubConn
}
