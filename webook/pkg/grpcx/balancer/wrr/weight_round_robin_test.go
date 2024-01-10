package wrr

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	"geektime-basic-go/webook/internal/integration/startup"
	pb "geektime-basic-go/webook/pkg/grpcx/balancer/wrr/proto"
)

const target = "service/user"

type EtcdTestSuite struct {
	suite.Suite
	client *etcdv3.Client
}

func (e *EtcdTestSuite) SetupSuite() {
	startup.InitViper()
	addr := viper.GetString("etcd.addr")
	require.NotEmpty(e.T(), addr, "没有找到配置 etcd.addr")

	var err error
	e.client, err = etcdv3.New(etcdv3.Config{Endpoints: []string{addr}})
	require.NoError(e.T(), err)
}

func TestServerRegistration(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}

func (e *EtcdTestSuite) TestServer() {
	addr := GetOutboundIP()
	go e.startServer(":8090", &pb.FailoverServer{Code: codes.Unavailable, Addr: addr + ":8090"}, 50)
	go e.startServer(":8092", &pb.User{Addr: addr + ":8092"}, 30)
	e.startServer(":8091", &pb.User{Addr: addr + ":8091"}, 20)
}

func (e *EtcdTestSuite) startServer(port string, svr pb.UserServiceServer, weight int) {
	em, err := endpoints.NewManager(e.client, target)
	require.NoError(e.T(), err)

	addr := GetOutboundIP() + port
	key := target + "/" + addr

	err = em.AddEndpoint(context.Background(), key, endpoints.Endpoint{Addr: addr, Metadata: map[string]any{"weight": weight}})
	require.NoError(e.T(), err)

	server := grpc.NewServer()
	lis, err := net.Listen("tcp", port)
	require.NoError(e.T(), err)

	pb.RegisterUserServiceServer(server, svr)
	err = server.Serve(lis)
	assert.NoError(e.T(), err)

	err = em.DeleteEndpoint(context.Background(), key)
	assert.NoError(e.T(), err)
	server.GracefulStop()
}

// TestClientWithGRPCRoundRobin grpc 自带的轮询
func (e *EtcdTestSuite) TestClientWithGRPCRoundRobin() {
	etcdResolver, err := resolver.NewBuilder(e.client)
	require.NoError(e.T(), err)

	cc, err := grpc.DialContext(
		context.Background(),
		"etcd:///"+target,
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]}`, WeightRoundRobin)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(e.T(), err)
	defer cc.Close()
	uc := pb.NewUserServiceClient(cc)

	for i := 0; i < 50; i++ {
		resp, err1 := uc.GetUser(context.Background(), &pb.GetUserRequest{Class: "1", UserName: "小明", UserID: "1"})
		if err1 != nil {
			e.T().Log(err1)
			continue
		}
		e.T().Log(resp.Address)
	}
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
