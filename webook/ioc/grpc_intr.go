package ioc

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	intr "geektime-basic-go/webook/api/proto/gen/interactive"
	"geektime-basic-go/webook/interactive/service"
	intrRPC "geektime-basic-go/webook/internal/repository/rpc/interactive"
	"geektime-basic-go/webook/pkg/logger"
)

func InitEtcd() *clientv3.Client {
	var cfg clientv3.Config
	err := viper.UnmarshalKey("etcd", &cfg)
	if err != nil {
		panic(fmt.Sprintf("初始化 etcd client 反序列化配置失败: %s", err))
	}
	cli, err := clientv3.New(cfg)
	if err != nil {
		panic(fmt.Sprintf("初始化 etcd client 失败: %s", err))
	}
	return cli
}

func InitInteractiveGRPC(client *clientv3.Client) intr.InteractiveServiceClient {
	type Config struct {
		Secure bool   `json:"secure"`
		Name   string `json:"name"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("grpc.client.intr", &cfg); err != nil {
		panic(fmt.Sprintf("初始化 grpc client 失败, 反序列化配置失败: %s", err))
	}
	var opts []grpc.DialOption
	if !cfg.Secure {
		opts = append(opts)
	}

	bd, err := resolver.NewBuilder(client)
	if err != nil {
		panic(fmt.Sprintf("初始化 grpc resolver 失败: %s", err))
	}

	cc, err := grpc.Dial(
		"etcd:///service/"+cfg.Name,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithResolvers(bd),
	)
	if err != nil {
		panic(fmt.Sprintf("初始化 grpc client 失败, 连接失败: %s", err))
	}

	return intr.NewInteractiveServiceClient(cc)
}

func InitInteractiveRPC(svc service.InteractiveService, l logger.Logger) intr.InteractiveServiceClient {
	type Config struct {
		Addr      string
		Secure    bool
		Threshold int32
	}

	var cfg Config
	if err := viper.UnmarshalKey("grpc.client.intr", &cfg); err != nil {
		panic(fmt.Sprintf("初始化 grpc client 失败, 反序列化配置失败: %s", err))
	}
	var opts []grpc.DialOption
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		panic(fmt.Sprintf("初始化 grpc client 失败, 连接失败: %s", err))
	}

	remote := intrRPC.NewGRPCInteractiveRPC(intr.NewInteractiveServiceClient(cc))
	local := intrRPC.NewInteractiveServiceAdapter(svc)
	res := intrRPC.NewInteractiveClient(remote, local, cfg.Threshold)

	viper.OnConfigChange(func(in fsnotify.Event) {
		cfg = Config{}
		if e := viper.UnmarshalKey("grpc.client.intr", &cfg); e != nil {
			l.Error("重新加载 grpc.client.intr 的配置失败", logger.Error(e))
			return
		}
		res.UpdateThreshold(cfg.Threshold)
	})
	return res
}
