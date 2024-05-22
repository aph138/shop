package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pb "github.com/aph138/shop/api/user_grpc"
	"github.com/aph138/shop/pkg/db"
	"github.com/aph138/shop/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	logger     *slog.Logger
	collection *mongo.Collection
	pb.UnimplementedUserServer
}

const TIMEOUT = 1500

var (
	//TODO define other errors
	errWrongPassword = status.Error(codes.Unauthenticated, "wrong username or password")
	DB_NAME          = os.Getenv("DB_NAME")
	Collection       = "user"
)

func main() {
	logger := slog.Default()
	db, err := db.NewDB(os.Getenv("mongoURL"), logger, nil)
	if err != nil {
		panic(err)
	}
	c := db.Database(DB_NAME).Collection(Collection)
	if err = createUniqeIndex(c); err != nil {
		panic(err.Error())
	}
	user := &UserService{
		logger:     logger,
		collection: c,
	}

	//TODO:interceptor
	opt := []grpc.ServerOption{}
	srv := grpc.NewServer(opt...)
	pb.RegisterUserServer(srv, user)

	go func() {
		listener, err := net.Listen("tcp", ":9001")
		if err != nil {
			panic(err)
		}
		defer listener.Close()
		logger.Info("User Server is running on " + listener.Addr().String())
		if err := srv.Serve(listener); err != nil {
			panic(err)
		}
	}()

	//graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	<-sig
	logger.Info("Starting User Service Graceful shutdown ...")
	srv.GracefulStop()
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*5000)
	defer cancel()
	if err = db.Disconnect(ctx); err != nil {
		logger.Error(err.Error())
	}

	logger.Info("User Service has been shutted down gracefully")

}

func (u *UserService) Signin(ctx context.Context, in *pb.SigninRequest) (*pb.SigninResponse, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Millisecond*TIMEOUT)
	defer cancel()
	query := bson.M{"username": in.Username}
	var user shared.User
	err := u.collection.FindOne(_ctx, query).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errWrongPassword
		}
		u.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return nil, errWrongPassword
	}
	return &pb.SigninResponse{Id: user.ID}, nil
}
func (u *UserService) Signup(ctx context.Context, in *pb.SignupRequest) (*pb.SigninResponse, error) {
	_ctx, cancel := context.WithTimeout(ctx, time.Millisecond*TIMEOUT)
	defer cancel()

	pass, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	user := shared.User{
		Username: in.Username,
		Email:    in.Email,
		Password: string(pass),
		Role:     in.Role,
		Status:   true,
	}

	data, err := bson.Marshal(user)
	if err != nil {

		u.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	result, err := u.collection.InsertOne(_ctx, data)
	if err != nil {
		//check if the username or email is duplicated or not
		if mongo.IsDuplicateKeyError(err) {
			if we, ok := err.(mongo.WriteException); ok && len(we.WriteErrors) > 0 {
				e := we.WriteErrors[0]
				if strings.Contains(e.Message, "email_1") {
					return nil, status.Error(codes.InvalidArgument, "duplicated email")
				} else if strings.Contains(e.Message, "username_1") {
					return nil, status.Error(codes.InvalidArgument, "duplicated username")
				} else {
					return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("duplicated %s", e.Message))
				}
			} else {
				return nil, status.Error(codes.InvalidArgument, "duplicated unkown field")
			}
		}
		u.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	id := result.InsertedID.(primitive.ObjectID).Hex()
	return &pb.SigninResponse{Id: id}, nil
}
func (u *UserService) UserList(in *pb.Empty, stream pb.User_UserListServer) error {
	//1 to include, 0 to exclude
	ctx := context.Background()
	option := options.Find().SetProjection(bson.M{"username": 1, "email": 1, "role": 1, "status": 1})
	cursor, err := u.collection.Find(ctx, bson.D{}, option)
	if err != nil {
		u.logger.Error(err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	defer cursor.Close(ctx)
	var user shared.User
	for cursor.Next(ctx) {
		err = cursor.Decode(&user)
		if err != nil {
			u.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}
		if err = stream.Send(&pb.UserListResponse{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
			Status:   user.Status,
		}); err != nil {
			return err
		}
	}
	// var result
	return nil
}
func createUniqeIndex(c *mongo.Collection) error {
	email := mongo.IndexModel{
		Keys:    bson.D{primitive.E{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true)}

	username := mongo.IndexModel{
		Keys:    bson.D{primitive.E{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true)}

	_, err := c.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{email, username})
	if err != nil {
		return err
	}
	return nil
}
