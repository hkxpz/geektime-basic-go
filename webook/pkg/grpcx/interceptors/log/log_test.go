package log

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"geektime-basic-go/webook/internal/integration/startup"
	"geektime-basic-go/webook/pkg/grpcx/proto"
)

type LogTestSuite struct {
	suite.Suite
	interceptor grpc.UnaryServerInterceptor
}

func TestServerRegistration(t *testing.T) {
	suite.Run(t, new(LogTestSuite))
}

func (l *LogTestSuite) SetupSuite() {
	startup.InitViper()
	l.interceptor = NewInterceptorBuilder(startup.InitZapLogger()).defaultUnaryServerInterceptor()
}

func (l *LogTestSuite) TestServer() {
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(l.interceptor))
	lis, err := net.Listen("tcp", ":8090")
	require.NoError(l.T(), err)

	proto.RegisterUserServiceServer(server, &proto.User{})
	err = server.Serve(lis)
	assert.NoError(l.T(), err)
}

func (l *LogTestSuite) TestClient() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cc, err := grpc.DialContext(
		ctx, "localhost:8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(l.T(), err)
	defer cc.Close()
	uc := proto.NewUserServiceClient(cc)

	md := metadata.New(map[string]string{"app": "test-demo", "client-ip": GetOutboundIP()})
	resp, err := uc.GetUser(metadata.NewOutgoingContext(ctx, md), &proto.GetUserRequest{Class: "1", UserName: "小明", UserID: "1"})
	if err != nil {
		l.T().Log(err)
		return
	}
	l.T().Log(resp)
}

// GetOutboundIP 获得对外发送消息的 IP 地址
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
