package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	intr "geektime-basic-go/webook/interactive/grpc"
	"geektime-basic-go/webook/pkg/grpcx"
)

func InitGRPCxServer(intr *intr.InteractiveServiceServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	intr.Register(server)
	return &grpcx.Server{Server: server, Addr: cfg.Addr}
}
