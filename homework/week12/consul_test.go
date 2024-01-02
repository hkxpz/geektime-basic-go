package week12

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "geektime-basic-go/homework/week12/proto"
)

type ConsulSuite struct {
	suite.Suite
	client *api.Client
}

func (c *ConsulSuite) SetupSuite() {
	addr, ok := os.LookupEnv("CONSUL_ADDR")
	require.True(c.T(), ok, "没有找到环境变量 CONSUL_ADDR")

	var err error
	cfg := api.DefaultConfig()
	cfg.Address = addr
	c.client, err = api.NewClient(cfg)
	require.NoError(c.T(), err)
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(ConsulSuite))
}

// TestServerRegistration 测试 Consul 作为服务注册与发现中心
// 这是服务端注册的过程
func (c *ConsulSuite) TestServerRegistration() {
	const (
		target    = "service/user"
		serviceID = "node1"
	)
	t := c.T()

	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    target,
		Port:    8090,
		Tags:    []string{target},
		Address: GetOutboundIP(),
		Check: &api.AgentServiceCheck{
			TTL:     "5s",
			CheckID: serviceID,
		},
	}
	err := c.client.Agent().ServiceRegister(registration)
	require.NoError(t, err)

	cancelCtx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				err = c.client.Agent().UpdateTTL(serviceID, "Service up and running", api.HealthPassing)
				require.NoError(t, err)
			case <-cancelCtx.Done():
				ticker.Stop()
			}
		}
	}()

	server := grpc.NewServer()
	lis, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)

	pb.RegisterUserServer(server, &User{})
	err = server.Serve(lis)
	assert.NoError(t, err)

	// 优雅退出
	cancel()
	err = c.client.Agent().ServiceDeregister(serviceID)
	assert.NoError(t, err)
	server.GracefulStop()
}

// TestClientDiscovery 测试 Consul 作为服务注册与发现中心
// 这是客户端的发现过程
func (c *ConsulSuite) TestClientDiscovery() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cc, err := grpc.DialContext(
		ctx, "consul:///service/user",
		grpc.WithResolvers(&resolverBuilder{client: c.client}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(c.T(), err)
	defer cc.Close()
	uc := pb.NewUserClient(cc)
	resp, err := uc.GetUser(ctx, &pb.GetUserRequest{Class: "1", UserName: "小明", UserID: "1"})
	require.NoError(c.T(), err)
	c.T().Log(resp)
}

// GetOutboundIP 获得对外发送消息的 IP 地址
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "114.114.114.114:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
