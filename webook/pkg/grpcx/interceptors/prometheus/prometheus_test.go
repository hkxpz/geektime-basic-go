package prometheus

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		fmt.Println(http.ListenAndServe(":8081", nil))
	}()
	l.interceptor = NewInterceptorBuilder("hkxpz", "webook").defaultUnaryServerInterceptor()
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

	for i := 0; i < 100; i++ {
		resp, e := uc.GetUser(ctx, &proto.GetUserRequest{Class: "1", UserName: "小明", UserID: "1"})
		if e != nil {
			l.T().Log(e)
			return
		}
		l.T().Log(resp)
	}
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
