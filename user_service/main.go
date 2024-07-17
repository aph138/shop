package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aph138/shop/api/common"
	pb "github.com/aph138/shop/api/user_grpc"
	"github.com/aph138/shop/pkg/db"
	"github.com/aph138/shop/pkg/logger"
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

const TIMEOUT = time.Millisecond * 1500

var (
	errWrongPassword = status.Error(codes.Unauthenticated, "wrong username or password")
	DB_NAME          = os.Getenv("DB_NAME")
	Collection       = "user"
)

func main() {
	l := slog.Default()
	db, err := db.NewDB(os.Getenv("DB_URL"), l, nil)
	if err != nil {
		panic(err)
	}
	c := db.Database(DB_NAME).Collection(Collection)
	if err = createUniqeIndex(c); err != nil {
		panic(err.Error())
	}
	user := &UserService{
		logger:     l,
		collection: c,
	}
	grpcLogger := logger.NewLogger(l)
	opt := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpcLogger.UnaryServerLogger),
	}
	srv := grpc.NewServer(opt...)
	pb.RegisterUserServer(srv, user)

	go func() {
		listener, err := net.Listen("tcp", ":9001")
		if err != nil {
			panic(err)
		}
		defer listener.Close()
		l.Info("User Server is running on " + listener.Addr().String())
		if err := srv.Serve(listener); err != nil {
			panic(err)
		}
	}()

	//graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	<-sig
	l.Info("Starting User Service Graceful shutdown ...")
	srv.GracefulStop()
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*5000)
	defer cancel()
	if err = db.Disconnect(ctx); err != nil {
		l.Error(err.Error())
	}

	l.Info("User Service has been shutted down gracefully")

}

func (u *UserService) Signin(ctx context.Context, in *pb.SigninRequest) (*common.StringMessage, error) {
	_ctx, cancel := context.WithTimeout(ctx, TIMEOUT)
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
	return &common.StringMessage{Value: user.ID}, nil
}

func (u *UserService) Signup(ctx context.Context, in *pb.SignupRequest) (*common.StringMessage, error) {
	_ctx, cancel := context.WithTimeout(ctx, TIMEOUT)
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
	log.Println(user)
	result, err := u.collection.InsertOne(_ctx, user)
	if err != nil {
		//check if the username or email is duplicated or not
		if mongo.IsDuplicateKeyError(err) {
			if we, ok := err.(mongo.WriteException); ok && len(we.WriteErrors) > 0 {
				e := we.WriteErrors[0]
				if strings.Contains(e.Message, "email_1") {
					return nil, status.Error(codes.AlreadyExists, "duplicated email")
				} else if strings.Contains(e.Message, "username_1") {
					return nil, status.Error(codes.AlreadyExists, "duplicated username")
				} else {
					return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("duplicated %s", e.Message))
				}
			} else {
				return nil, status.Error(codes.AlreadyExists, "duplicated unkown field")
			}
		}
		u.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	id := result.InsertedID.(primitive.ObjectID).Hex()
	return &common.StringMessage{Value: id}, nil
}
func (u *UserService) UserList(in *common.Empty, stream pb.User_UserListServer) error {
	ctx := context.Background()
	//1 to include, 0 to exclude
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
func (u *UserService) GetUser(ctx context.Context, in *common.StringMessage) (*pb.GetUserResponse, error) {
	var result shared.User

	_ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	id, err := primitive.ObjectIDFromHex(in.Value)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	err = u.collection.FindOne(_ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "user not founded")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetUserResponse{Username: result.Username,
		Email:  result.Email,
		Role:   result.Role,
		Status: result.Status,
		Address: &pb.Address{
			Address: result.Address.Address,
			Phone:   result.Address.Phone},
	}, nil
}

func (u *UserService) DeleteUser(ctx context.Context, in *common.StringMessage) (*common.BoolMessage, error) {
	_ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	id, err := primitive.ObjectIDFromHex(in.Value)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	r, err := u.collection.DeleteOne(_ctx, bson.M{"_id": id})
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &common.BoolMessage{Value: r.DeletedCount >= 1}, nil

}
func (u *UserService) EditUser(ctx context.Context, in *pb.EditUserRequest) (*common.BoolMessage, error) {
	_ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	id, err := primitive.ObjectIDFromHex(in.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	user := shared.User{
		Email: in.Email,
		Address: shared.Address{
			Address: in.Address.Address,
			Phone:   in.Address.Phone,
		},
	}
	result, err := u.collection.UpdateByID(_ctx, id, bson.M{"$set": user})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, status.Error(codes.AlreadyExists, "duplicated email")
		}
		u.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &common.BoolMessage{Value: result.ModifiedCount == 1}, nil

}
func (u *UserService) ChangePassword(ctx context.Context, in *pb.ChangePasswordRequest) (*common.BoolMessage, error) {
	var result bool
	_ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	id, err := primitive.ObjectIDFromHex(in.Id)
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	var user shared.User
	opt := options.FindOne().SetProjection(bson.M{"password": 1})
	err = u.collection.FindOne(_ctx, bson.M{"_id": id}, opt).Decode(&user)
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.OldPassword)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "wrong password")
	}
	_ctx, cancel = context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	newPassword, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad new password")
	}
	r, err := u.collection.UpdateByID(_ctx, id, bson.M{"$set": bson.M{"password": string(newPassword)}})
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	if r.ModifiedCount >= 1 {
		result = true
	}
	return &common.BoolMessage{Value: result}, nil

}

func (u *UserService) AddToCart(ctx context.Context, in *pb.AddToCartRequest) (*common.BoolMessage, error) {
	//check if user id is valid
	id, err := primitive.ObjectIDFromHex(in.UserId)
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, "wrong user id")
	}
	//check if item id is valid
	item_id, err := primitive.ObjectIDFromHex(in.ItemId)
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, "wrong item id")
	}

	cart := shared.Cart{
		ItemID:   item_id,
		Quantity: in.Quntity,
	}

	// TODO check if the item is already in cart
	data := bson.M{
		"$push": bson.M{
			"cart": cart,
		},
	}
	//
	c, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	result, err := u.collection.UpdateByID(c, id, data)
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	if result.ModifiedCount <= 0 {
		return &common.BoolMessage{Value: false}, nil
	}
	return &common.BoolMessage{Value: true}, nil
}
func (u *UserService) DeleteFromCart(ctx context.Context, in *pb.DeleteFromCartRequest) (*common.BoolMessage, error) {
	user_id, err := primitive.ObjectIDFromHex(in.UserId)
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, "wrong user id")
	}
	item_id, err := primitive.ObjectIDFromHex(in.ItemId)
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, "wrong item id")
	}
	c, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	data := bson.M{
		"$pull": bson.M{
			"cart": bson.M{"item_id": item_id},
		},
	}
	result, err := u.collection.UpdateByID(c, user_id, data)
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	if result.ModifiedCount <= 0 {
		return &common.BoolMessage{Value: false}, nil
	}
	return &common.BoolMessage{Value: true}, nil
}
func (u *UserService) UpdateCart(ctx context.Context, in *pb.UpdateCartRequest) (*common.BoolMessage, error) {
	user_id, err := primitive.ObjectIDFromHex(in.UserId)
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, "wrong user id")
	}
	item_id, err := primitive.ObjectIDFromHex(in.ItemId)
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, "wrong item id")
	}
	c, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	filter := bson.M{
		"_id":          user_id,
		"cart.item_id": item_id,
	}
	//TODO: check quantity for overflow
	data := bson.M{
		"$set": bson.M{
			"cart.$.quantity": in.NewQuantity,
		},
	}

	result, err := u.collection.UpdateOne(c, filter, data)
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	if result.ModifiedCount <= 0 {
		return &common.BoolMessage{}, nil
	}
	return &common.BoolMessage{Value: false}, nil
}

// retrieve list of item (ID, Name, Link,Price,Poster)
func (u *UserService) Cart(ctx context.Context, in *common.StringMessage) (*pb.CartResponse, error) {
	user_id, err := primitive.ObjectIDFromHex(in.Value)
	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, "wrong user id")
	}

	c, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	// imitate relationship in mongo db
	match := bson.D{
		primitive.E{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: user_id}}},
	}
	lookup := bson.D{primitive.E{
		Key: "$lookup",
		Value: bson.D{
			primitive.E{Key: "from", Value: "stock"},
			primitive.E{Key: "localField", Value: "cart.item_id"}, // Note .item_id because we have list
			primitive.E{Key: "foreignField", Value: "_id"},
			primitive.E{Key: "as", Value: "cart"}, //field that will be used to display the result (if already exist will be overwritten)
		},
	}}
	pipeline := mongo.Pipeline{match, lookup}
	cursor, err := u.collection.Aggregate(c, pipeline)

	if err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer cursor.Close(context.Background())
	var user []shared.User

	if err = cursor.All(context.Background(), &user); err != nil {
		u.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	var result []*pb.CartItem
	u.logger.Info(fmt.Sprintf("%v", user))
	// TODO: check if user is not empty
	if len(user) == 0 {
		return nil, status.Error(codes.NotFound, "user not founded")
	}
	for _, i := range user[0].Cart {
		result = append(result, &pb.CartItem{
			Id:     i.ID,
			Name:   i.Name,
			Link:   i.Link,
			Price:  i.Price,
			Poster: i.Poster,
		})
	}
	return &pb.CartResponse{
		Item: result,
	}, nil
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
