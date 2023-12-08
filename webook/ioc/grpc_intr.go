package ioc

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	intr "geektime-basic-go/webook/api/proto/gen/interactive"
	"geektime-basic-go/webook/interactive/service"
	"geektime-basic-go/webook/internal/web/client"
	"geektime-basic-go/webook/pkg/logger"
)

func InitInteractiveGRPCClient(svc service.InteractiveService, l logger.Logger) intr.InteractiveServiceClient {
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

	remote := intr.NewInteractiveServiceClient(cc)
	local := client.NewInteractiveServiceAdapter(svc)
	res := client.NewInteractiveClient(remote, local, cfg.Threshold)

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
