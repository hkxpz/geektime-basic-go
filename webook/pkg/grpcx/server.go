package grpcx

import (
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	*grpc.Server
	Addr string
}

func (s *Server) Serve() error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.Server.Serve(l)
}
