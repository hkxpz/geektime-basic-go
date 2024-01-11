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
	l.interceptor = NewLoggerInterceptorBuilder(startup.InitZapLogger()).defaultUnaryServerInterceptor()
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

	resp, err := uc.GetUser(context.Background(), &proto.GetUserRequest{Class: "1", UserName: "小明", UserID: "1"})
	l.T().Log(err)
	l.T().Log(resp)
}
