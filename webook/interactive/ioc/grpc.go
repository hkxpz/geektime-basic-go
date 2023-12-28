package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"geektime-basic-go/webook/pkg/logger"

	intr "geektime-basic-go/webook/interactive/grpc"
	"geektime-basic-go/webook/pkg/grpcx"
)

func InitGRPCxServer(intr *intr.InteractiveServiceServer, l logger.Logger) *grpcx.Server {
	type Config struct {
		Port    int   `yaml:"port"`
		EtcdTTL int64 `yaml:"etcdTTL"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	intr.Register(server)
	return &grpcx.Server{
		Server:   server,
		Port:     cfg.Port,
		Name:     "interactive",
		L:        l,
		EtcdTTL:  cfg.EtcdTTL,
		EtcdAddr: viper.GetString("etcd.addr"),
	}
}
