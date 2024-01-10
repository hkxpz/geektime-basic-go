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
		md, _ := connInfo.Address.Metadata.(map[string]any)
		weightVal, _ := md["weight"]
		weight, _ := weightVal.(float64)
		if weight == 0 {
			weight = 10
		}
		conns = append(conns, &weightConn{
			cc:            conn,
			initWeigh:     int(weight),
			weight:        int(weight),
			currentWeight: int(weight),
			nodeType:      NodeTypeAvailable,
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
	return balancer.PickResult{SubConn: res.cc, Done: p.Done(res)}, nil
}

// Done 在这里执行 failover 有关的事情
func (p *WeightedPicker) Done(node *weightConn) func(info balancer.DoneInfo) {
	return func(info balancer.DoneInfo) {
		p.mutex.Lock()
		defer p.mutex.Unlock()
		if info.Err != nil {
			// 可以用和为准备状态, 等级加一, 权重减半
			if node.nodeType == NodeTypeAvailable || node.nodeType == NodeTypeNotReady {
				node.nodeType += 1
				node.weight /= 2
			}
			// 不可用状态, 当前权重降至最低
			if node.nodeType == NodeTypeAbnormal {
				node.currentWeight = node.minWeight
			}
			return
		}

		// 恢复一半权重
		if node.nodeType == NodeTypeNotReady {
			node.nodeType -= 1
			node.weight = node.initWeigh
			return
		}

		if node.nodeType == NodeTypeAbnormal {
			node.nodeType -= 1
			node.weight = node.initWeigh / 2
		}
	}
}

type nodeType int

const (
	NodeTypeUnknown = nodeType(iota)
	NodeTypeAvailable
	NodeTypeNotReady
	NodeTypeAbnormal
)

// weightConn 代表节点
type weightConn struct {
	// 初始权重
	initWeigh int
	// 权重
	weight int
	// 当前权重
	currentWeight int
	// 最小权重
	minWeight int

	nodeType nodeType

	// 真正的，grpc 里面的代表一个节点的表达
	cc balancer.SubConn
}
