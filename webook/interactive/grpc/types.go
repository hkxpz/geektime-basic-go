package grpc

import "google.golang.org/grpc"

type Service interface {
	Register(s *grpc.Server)
}
