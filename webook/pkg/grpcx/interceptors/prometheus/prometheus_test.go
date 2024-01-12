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

type PrometheusTestSuite struct {
	suite.Suite
	interceptor grpc.UnaryServerInterceptor
}

func TestServerRegistration(t *testing.T) {
	suite.Run(t, new(PrometheusTestSuite))
}

func (p *PrometheusTestSuite) SetupSuite() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		fmt.Println(http.ListenAndServe(":8081", nil))
	}()
	p.interceptor = NewInterceptorBuilder("hkxpz", "webook").defaultUnaryServerInterceptor()
}

func (p *PrometheusTestSuite) TestServer() {
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(p.interceptor))
	lis, err := net.Listen("tcp", ":8090")
	require.NoError(p.T(), err)

	proto.RegisterUserServiceServer(server, &proto.User{})
	err = server.Serve(lis)
	assert.NoError(p.T(), err)
}

func (p *PrometheusTestSuite) TestClient() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cc, err := grpc.DialContext(
		ctx, "localhost:8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(p.T(), err)
	defer cc.Close()
	uc := proto.NewUserServiceClient(cc)

	for i := 0; i < 100; i++ {
		resp, e := uc.GetUser(ctx, &proto.GetUserRequest{Class: "1", UserName: "小明", UserID: "1"})
		if e != nil {
			p.T().Log(e)
			return
		}
		p.T().Log(resp)
	}
}
