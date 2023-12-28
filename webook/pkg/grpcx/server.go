package grpcx

import (
	"context"
	"net"
	"strconv"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"

	"geektime-basic-go/webook/pkg/logger"
	"geektime-basic-go/webook/pkg/netx"
)

type Server struct {
	*grpc.Server
	Port        int
	etcdClient  *clientv3.Client
	etcdManager endpoints.Manager
	cancel      func()
	Name        string
	etcdKey     string
	EtcdTTL     int64
	EtcdAddr    string
	L           logger.Logger
}

func (s *Server) Serve() error {
	var ctx context.Context
	ctx, s.cancel = context.WithCancel(context.Background())
	port := strconv.Itoa(s.Port)
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	if err = s.register(ctx, port); err != nil {
		return err
	}
	return s.Server.Serve(l)
}

func (s *Server) register(ctx context.Context, port string) error {
	ec, err := clientv3.NewFromURL(s.EtcdAddr)
	if err != nil {
		return err
	}
	s.etcdClient = ec
	serviceName := "service/" + s.Name
	em, err := endpoints.NewManager(ec, serviceName)
	if err != nil {
		return err
	}
	s.etcdManager = em
	ip := netx.GetOutboundIP()
	s.etcdKey = serviceName + "/" + ip
	addr := ip + ":" + port
	leaseResp, err := ec.Grant(ctx, s.EtcdTTL)
	ch, err := ec.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return err
	}
	go func() {
		for chResp := range ch {
			s.L.Debug("续约：", logger.String("resp", chResp.String()))
		}
	}()
	return em.AddEndpoint(ctx, s.etcdKey, endpoints.Endpoint{Addr: addr}, clientv3.WithLease(leaseResp.ID))
}

func (s *Server) Close() error {
	s.cancel()
	if s.etcdManager != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.etcdManager.DeleteEndpoint(ctx, s.etcdKey); err != nil {
			return err
		}
	}
	if err := s.etcdClient.Close(); err != nil {
		return err
	}
	s.Server.GracefulStop()
	return nil
}
