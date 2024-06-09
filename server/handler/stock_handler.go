package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"path/filepath"
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
		item.Link = r.Link
		item.Name = r.Name
		item.Price = r.Price
		item.Poster = r.Poster
		item.Number = r.Number
		result = append(result, item)
	}
	render(c, stockview.ItemList(result))

}
func (s *stockHandler) PostAddItem(c *gin.Context) {

	//TODO: add other fields
	name := c.Request.FormValue("name")

	if len(name) < 1 {
		c.String(http.StatusBadRequest, "empty name field")
		return
	}

	link := c.Request.FormValue("link")
	if len(link) < 1 {
		//default value for link is product's name
		link = name
	}
	form, err := c.MultipartForm()
	if err != nil {
		log.Println(err.Error())
		c.String(http.StatusBadRequest, "multipart form error")
		return
	}
	// check if there is a poster
	posterFile := form.File["poster"]
	if len(posterFile) <= 0 {
		c.String(http.StatusBadRequest, "no poster")
		return
	}
	// save poster path
	folder := time.Now().Unix()
	poster := fmt.Sprintf("%d/%s", folder, filepath.Base(posterFile[0].Filename))
	dst := filepath.Join("./uploads/", fmt.Sprint(folder), poster)

	req := &pb.Item{
		Name:   name,
		Link:   link,
		Poster: poster,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	res, err := s.client.AddItem(ctx, req)
	if err != nil {

		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.AlreadyExists:
				{
					c.String(http.StatusBadRequest, "an item with this link alreay exists")
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

		//TODO: set max size uplaod
		//save images on disk if everything went good
		if err = c.SaveUploadedFile(posterFile[0], dst); err != nil {
			s.logger.Error(err.Error())
			c.String(http.StatusInternalServerError, RetryMSG)
			return
		}
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
	result, err := s.client.GetItem(ctx, &pb.GetItemRequest{Link: c.Param("name")})
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

	if len(item.ID) < 1 {
		c.String(http.StatusBadRequest, "no item has founded")
	}
	render(c, stockview.Item(item))
}
