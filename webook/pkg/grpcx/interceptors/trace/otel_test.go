package trace

import (
	"context"
	"net"
	"testing"
	"time"

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
	"geektime-basic-go/webook/ioc"
	"geektime-basic-go/webook/pkg/grpcx/proto"
)

const target = "service/user"

type OTELTestSuite struct {
	suite.Suite
	client      *etcdv3.Client
	interceptor grpc.UnaryServerInterceptor
}

func (o *OTELTestSuite) SetupSuite() {
	startup.InitViper()
	ioc.InitOTEL()
	addr := viper.GetString("etcd.addr")
	require.NotEmpty(o.T(), addr, "没有找到配置 etcd.addr")

	var err error
	o.client, err = etcdv3.New(etcdv3.Config{Endpoints: []string{addr}})
	require.NoError(o.T(), err)

	o.interceptor = NewInterceptorBuilder(nil, nil, "").defaultUnaryServerInterceptor()
}

func TestServerRegistration(t *testing.T) {
	suite.Run(t, new(OTELTestSuite))
}

func (o *OTELTestSuite) TestServer() {
	addr := GetOutboundIP()
	go o.startServer(":8090", &proto.FailoverServer{Code: codes.Unavailable, Addr: addr + ":8090"})
	o.startServer(":8091", &proto.User{Addr: addr + ":8091"})
}

func (o *OTELTestSuite) startServer(port string, svr proto.UserServiceServer) {
	em, err := endpoints.NewManager(o.client, target)
	require.NoError(o.T(), err)

	addr := GetOutboundIP() + port
	key := target + "/" + addr

	err = em.AddEndpoint(context.Background(), key, endpoints.Endpoint{Addr: addr})
	require.NoError(o.T(), err)

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(o.interceptor))
	lis, err := net.Listen("tcp", port)
	require.NoError(o.T(), err)

	proto.RegisterUserServiceServer(server, svr)
	err = server.Serve(lis)
	assert.NoError(o.T(), err)

	err = em.DeleteEndpoint(context.Background(), key)
	assert.NoError(o.T(), err)
	server.GracefulStop()
}

// TestClientWithGRPCRoundRobin grpc 自带的轮询
func (o *OTELTestSuite) TestClient() {
	etcdResolver, err := resolver.NewBuilder(o.client)
	require.NoError(o.T(), err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	cc, err := grpc.DialContext(
		ctx, "etcd:///"+target,
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	cancel()
	require.NoError(o.T(), err)
	defer cc.Close()
	uc := proto.NewUserServiceClient(cc)

	for i := 0; i < 50; i++ {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		resp, err1 := uc.GetUser(ctx, &proto.GetUserRequest{Class: "1", UserName: "小明", UserID: "1"})
		cancel()
		if err1 != nil {
			o.T().Log(err1)
			continue
		}
		o.T().Log(resp.Address)
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
