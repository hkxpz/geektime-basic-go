package week12

import (
	"context"

	pb "geektime-basic-go/homework/week12/proto"
)

// User is used to implement proto.user.UserServer.
type User struct {
	pb.UnimplementedUserServer
}

func (u *User) GetUser(ctx context.Context, r *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	return &pb.GetUserResponse{
		Class:    "普通用户",
		UserID:   r.UserID,
		UserName: r.UserName,
		Address:  "192.168.0.251",
		Sex:      "女",
		Phone:    "10086",
	}, nil
}
