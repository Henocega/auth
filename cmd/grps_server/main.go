package main

import (
	"context"
	"fmt"
	user "github.com/Henocega/auth/pkg/user_v1"
	"github.com/brianvoe/gofakeit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
)

const grpcPort = 50051

type server struct {
	user.UnimplementedUserV1Server
}

func (s *server) Get(ctx context.Context, req *user.GetRequest) (*user.GetResponse, error) {
	log.Printf("User id: %d", req.GetId())

	return &user.GetResponse{
		User: &user.User{
			Id: req.GetId(),
			Info: &user.UserInfo{
				Name:  gofakeit.BeerName(),
				Email: gofakeit.Name(),
				Role:  0,
			},
			CreatedAt: timestamppb.New(gofakeit.Date()),
			UpdatedAt: timestamppb.New(gofakeit.Date()),
		},
	}, nil
}

func (s *server) Create(ctx context.Context, req *user.CreateRequest) (*user.CreateResponse, error) {
	fmt.Printf("Create request: %v", req)

	return &user.CreateResponse{
		Id: gofakeit.Int64(),
	}, nil
}

func (s *server) Update(ctx context.Context, req *user.UpdateRequest) (*emptypb.Empty, error) {
	fmt.Printf("Update request: %v", req)

	return nil, nil
}

func (s *server) Delete(ctx context.Context, req *user.DeleteRequest) (*emptypb.Empty, error) {
	fmt.Printf("Delete request: %v", req)

	return nil, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	user.RegisterUserV1Server(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
