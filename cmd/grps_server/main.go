package main

import (
	"context"
	"flag"
	"github.com/Henocega/auth/internal/config"
	"github.com/Henocega/auth/internal/config/env"
	user "github.com/Henocega/auth/pkg/user_v1"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"time"
)

const grpcPort = 50051
const USER_TABLE = "\"user\""
const DSN = "host=localhost port=%s dbname=%s user=%s password=%s sslmode=disable"

type server struct {
	user.UnimplementedUserV1Server
	pool *pgxpool.Pool
}

func (s *server) Get(ctx context.Context, req *user.GetRequest) (*user.GetResponse, error) {

	builderSelect := sq.Select("id", "email", "name", "role", "created_at", "updated_at").
		From(USER_TABLE).
		Where(sq.Eq{"id": req.Id}).
		PlaceholderFormat(sq.Dollar).
		OrderBy("id ASC")

	query, args, err := builderSelect.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}

	rows := s.pool.QueryRow(ctx, query, args...)

	var id, role int64
	var email, name string
	var createdAt time.Time
	var updatedAt time.Time

	err = rows.Scan(&id, &email, &name, &role, &createdAt, &updatedAt)
	if err != nil {
		log.Fatalf("failed to scan user: %v", err)
	}

	return &user.GetResponse{
		User: &user.User{
			Id: id,
			Info: &user.UserInfo{
				Name:  name,
				Email: email,
				Role:  user.Role(role),
			},
			CreatedAt: timestamppb.New(createdAt),
			UpdatedAt: timestamppb.New(updatedAt),
		},
	}, nil
}

func (s *server) Create(ctx context.Context, req *user.CreateRequest) (*user.CreateResponse, error) {

	builderInsert := sq.Insert(USER_TABLE).
		PlaceholderFormat(sq.Dollar).
		Columns("email", "name", "role", "password", "updated_at", "created_at").
		Values(req.Info.Email, req.Info.Name, req.Info.Role, req.Password, time.Now(), time.Now()).
		Suffix("RETURNING id")
	log.Printf("Context: %v", ctx)

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}

	var userID int64
	err = s.pool.QueryRow(ctx, query, args...).Scan(&userID)
	if err != nil {
		log.Fatalf("failed to insert user: %v", err)
	}

	log.Printf("inserted user with id: %d", userID)

	return &user.CreateResponse{
		Id: userID,
	}, nil
}

func (s *server) Update(ctx context.Context, req *user.UpdateRequest) (*emptypb.Empty, error) {

	builderUpdate := sq.Update(USER_TABLE).
		PlaceholderFormat(sq.Dollar).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": req.Id})

	if req.Info.Email != nil {
		builderUpdate = builderUpdate.Set("email", req.Info.Email.Value)
	}

	if req.Info.Name != nil {
		builderUpdate = builderUpdate.Set("name", req.Info.Name.Value)
	}

	if req.Info.Role != 0 {
		builderUpdate = builderUpdate.Set("role", req.Info.Role)
	}

	builderUpdate.Where(sq.Eq{"id": req.Id})

	query, args, err := builderUpdate.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}

	_, err = s.pool.Exec(ctx, query, args...)

	if err != nil {
		log.Fatalf("failed to update user: %v", err)
	}

	return nil, nil
}

func (s *server) Delete(ctx context.Context, req *user.DeleteRequest) (*emptypb.Empty, error) {

	builderDelete := sq.Delete(USER_TABLE).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": req.Id})

	query, args, err := builderDelete.ToSql()
	if err != nil {
		log.Fatalf("failed to build query: %v", err)
	}

	_, err = s.pool.Exec(ctx, query, args...)

	if err != nil {
		log.Fatalf("failed to delete user: %v", err)
	}

	return nil, nil
}

func main() {
	flag.Parse()
	ctx := context.Background()

	err := config.Load(".env")

	if err != nil {
		log.Fatalf("Error to load config: %v", err)
	}

	grpcConfig, err := env.NewGRPCConfig()
	if err != nil {
		log.Fatalf("failed to get grpc config: %v", err)
	}

	pgConfig, err := env.NewPGConfig()
	if err != nil {
		log.Fatalf("failed to get pg config: %v", err)
	}

	lis, err := net.Listen("tcp", grpcConfig.Address())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	pool, err := pgxpool.Connect(ctx, pgConfig.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	s := grpc.NewServer()
	reflection.Register(s)
	user.RegisterUserV1Server(s, &server{pool: pool})

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
