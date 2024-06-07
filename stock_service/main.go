package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/aph138/shop/api/stock_grpc"
	"github.com/aph138/shop/pkg/db"
	"github.com/aph138/shop/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stockService struct {
	db     *mongo.Collection
	logger *slog.Logger
	pb.UnimplementedStockServer
}

const (
	TIMEOUT = time.Millisecond * 200
)

var (
	Address = ":9002"
	DBName  = os.Getenv("DB_NAME")
	DBUrl   = os.Getenv("DB_URL")
)

func main() {
	logger := slog.Default()

	opt := []grpc.ServerOption{}
	srv := grpc.NewServer(opt...)

	db, err := db.NewDB(DBUrl, logger, nil)
	if err != nil {
		panic("err when connecting to db: " + err.Error())
	}
	collection := db.Database(DBName).Collection("stock")
	if err = createUniqeIndex(collection); err != nil {
		panic("err when creating index: " + err.Error())
	}

	service := stockService{
		logger: logger,
		db:     collection,
	}

	pb.RegisterStockServer(srv, &service)

	go func() {
		port := os.Getenv("PORT")
		if port != "" {
			Address = fmt.Sprintf(":%s", port)
		}
		l, err := net.Listen("tcp", Address)
		if err != nil {
			panic("err when making listener: " + err.Error())
		}
		logger.Info("Stock Service in running at: " + l.Addr().String())
		if err = srv.Serve(l); err != nil {
			panic("err when starting server: " + err.Error())
		}

	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	<-sig
	logger.Info("srarting graceful shutdown")

	srv.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx.Done()

	logger.Info("stock service has been shutted down successfully")

}

func (s *stockService) AddItem(ctx context.Context, in *pb.Item) (*pb.AddItemResponse, error) {
	_ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	item := shared.Item{
		Name:        in.Name,
		Description: in.Description,
		Number:      in.Number,
		Price:       in.Price,
		Photos:      in.Photos,
	}
	r, err := s.db.InsertOne(_ctx, item)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, status.Error(codes.AlreadyExists, "duplicated name")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if r.InsertedID != nil {
		return &pb.AddItemResponse{Result: true}, nil
	}
	return &pb.AddItemResponse{Result: false}, nil
}
func (s *stockService) GetItem(ctx context.Context, in *pb.GetItemRequest) (*pb.Item, error) {
	_ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	id, err := primitive.ObjectIDFromHex(in.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	result := shared.Item{}
	err = s.db.FindOne(_ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Item{
		Id:          result.ID,
		Name:        result.Name,
		Description: result.Description,
		Price:       result.Price,
		Photos:      result.Photos,
		Number:      result.Number,
	}, nil
}

func (s *stockService) GetItemList(in *pb.GetItemListRequest, stream pb.Stock_GetItemListServer) error {
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()
	opt := options.Find().SetLimit(int64(in.Limit)).SetSkip(int64(in.Offset))
	cursor, err := s.db.Find(ctx, bson.D{{}}, opt)
	if err != nil {
		s.logger.Error(err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	defer cursor.Close(context.Background())
	var item shared.Item
	var result *pb.GetItemListResponse
	for cursor.Next(context.Background()) {
		err = cursor.Decode(&item)
		if err != nil {
			s.logger.Error(err.Error())
		}
		if len(item.Photos) > 0 {
			result = &pb.GetItemListResponse{
				Name:        item.Name,
				Description: item.Description,
				Price:       item.Price,
				Number:      item.Number,
				Photo:       item.Photos[0],
			}
		} else {
			result = &pb.GetItemListResponse{
				Name:        item.Name,
				Description: item.Description,
				Price:       item.Price,
				Number:      item.Number,
			}
		}
		stream.Send(result)
	}
	return nil
}
func createUniqeIndex(c *mongo.Collection) error {
	name := mongo.IndexModel{
		Keys:    bson.D{primitive.E{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := c.Indexes().CreateOne(context.Background(), name)
	return err
}
