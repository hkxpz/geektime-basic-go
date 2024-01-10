package proto

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

// User is used to implement proto.user.UserServer.
type User struct {
	UnimplementedUserServiceServer
	Addr string
}

func (u *User) GetUser(ctx context.Context, r *GetUserRequest) (*GetUserResponse, error) {
	return &GetUserResponse{
		Class:    "普通用户",
		UserID:   r.UserID,
		UserName: r.UserName,
		Address:  u.Addr,
		Sex:      "女",
		Phone:    "10086",
	}, nil
}

// FailoverServer 必定失败
type FailoverServer struct {
	UnimplementedUserServiceServer
	Code  codes.Code
	Addr  string
	count int
}

func (f *FailoverServer) GetUser(ctx context.Context, r *GetUserRequest) (*GetUserResponse, error) {
	defer func() { f.count++ }()
	if f.count%3 != 0 {
		log.Println("命中了failover服务器")
		return &GetUserResponse{}, status.Error(f.Code, "模拟 failover")
	}

	return &GetUserResponse{
		Class:    "普通用户",
		UserID:   r.UserID,
		UserName: r.UserName,
		Address:  f.Addr,
		Sex:      "女",
		Phone:    "10086",
	}, nil
}
