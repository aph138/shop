package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	pb "github.com/aph138/shop/api/stock_grpc"
	stockview "github.com/aph138/shop/server/web/stock"
	"github.com/aph138/shop/shared"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type stockHandler struct {
	conn   *grpc.ClientConn
	logger *slog.Logger
	client pb.StockClient
}

func NewStockHandler(grpc_url string, l *slog.Logger) (*stockHandler, error) {
	opt := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.Dial(grpc_url, opt...)
	if err != nil {
		return nil, err
	}
	return &stockHandler{
		logger: l,
		conn:   conn,
		client: pb.NewStockClient(conn),
	}, nil
}

func (s *stockHandler) Close() error {
	return s.conn.Close()
}
func (s *stockHandler) GetAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), GRPC_TIMEOUT)
	defer cancel()
	response, err := s.client.GetItemList(ctx, &pb.GetItemListRequest{Offset: 0, Limit: 10})
	if err != nil {
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}
	var result []shared.Item
	var item shared.Item
	for {
		r, err := response.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.String(http.StatusInternalServerError, RetryMSG)
			return
		}
		item.Name = r.Name
		item.Description = r.Description
		item.Price = r.Price
		item.Photos = []string{r.Photo}
		item.Number = r.Number
		result = append(result, item)
	}
	render(c, stockview.ItemList(result))

}
func (s *stockHandler) PostAddItem(c *gin.Context) {
	//TODO: add other fields
	name := c.Request.FormValue("name")
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()
	req := &pb.Item{
		Name: name,
	}
	res, err := s.client.AddItem(ctx, req)
	if err != nil {
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.AlreadyExists:
				{
					c.String(http.StatusBadRequest, "an item with this name alreay exists")
				}
			default:
				{
					c.String(http.StatusInternalServerError, RetryMSG)
				}
			}
			return
		}
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}
	if res.Result {
		c.String(http.StatusCreated, "item successfully has been added")
	} else {
		c.String(http.StatusNoContent, "item didn't add")
	}

}
func (s *stockHandler) GetAddItem(c *gin.Context) {
	render(c, stockview.AddItem())
}

func (s *stockHandler) GetItem(c *gin.Context) {
	item := shared.Item{}
	ctx, cancel := context.WithTimeout(context.Background(), GRPC_TIMEOUT)
	defer cancel()
	result, err := s.client.GetItem(ctx, &pb.GetItemRequest{})
	if err != nil {
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}
	item.Name = result.Name
	item.ID = result.Id
	item.Description = result.Description
	item.Photos = result.Photos
	item.Price = result.Price
	item.Number = result.Number

	render(c, stockview.Item(item))
}
func (s *stockHandler) GetItemList(c *gin.Context) {
	var items []shared.Item

	ctx, cancel := context.WithTimeout(context.Background(), GRPC_TIMEOUT)
	defer cancel()
	stream, err := s.client.GetItemList(ctx, &pb.GetItemListRequest{})
	if err != nil {

	}
	var item shared.Item
	for {
		i, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {

		}
		item = shared.Item{
			Name:        i.Name,
			Description: i.Description,
			Price:       i.Price,
			//TODO compelete
		}
		items = append(items, item)
	}

}
