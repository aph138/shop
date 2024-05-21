package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	pb "github.com/aph138/shop/api/user_grpc"
	"github.com/aph138/shop/pkg/auth"
	"github.com/aph138/shop/shared"
	"github.com/aph138/shop/web"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	WEEK_IN_SECOND = 60 * 60 * 24 * 7
	WEEK_IN_MINUTE = 60 * 24 * 7
	RetryMSG       = "Something went wrong. Please try again later"
)

type userHandler struct {
	client pb.UserClient
	logger *slog.Logger
	conn   *grpc.ClientConn
}

func NewUserHandler(url string, logger *slog.Logger) *userHandler {
	opt := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	userConn, err := grpc.Dial(url, opt...)
	if err != nil {
		logger.Error(err.Error())
	}
	return &userHandler{
		logger: logger,
		client: pb.NewUserClient(userConn),
		conn:   userConn,
	}
}
func (u *userHandler) Close() error {
	return u.conn.Close()
}
func (u *userHandler) SigninGet(c *gin.Context) {
	render(c, web.Signin())
}

func (u *userHandler) SigninPost(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*500)
	defer cancel()
	req := &pb.SigninRequest{
		Username: c.Request.FormValue("username"),
		Password: c.Request.FormValue("password")}

	res, err := u.client.Signin(ctx, req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.Unauthenticated:
				{
					http.Error(c.Writer, "wrong password", http.StatusInternalServerError)
					return
				}
			}
		}

		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	i, err := auth.NewIssuer("jwt.ed")
	if err != nil {
		u.logger.Error(err.Error())
		http.Error(c.Writer, RetryMSG, http.StatusInternalServerError)
		return
	}
	data := map[string]string{"id": res.Id}
	token, err := i.Token(data, WEEK_IN_MINUTE)
	if err != nil {
		u.logger.Error(err.Error())
		http.Error(c.Writer, RetryMSG, http.StatusInternalServerError)
		return
	}
	c.SetCookie("jwt", token, WEEK_IN_SECOND, "", "", true, true)
	c.Redirect(http.StatusFound, "/")

}

func (u *userHandler) SignupGet(c *gin.Context) {
	render(c, web.Signup())
}

func (u *userHandler) SignupPost(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*500)
	defer cancel()
	pass := c.Request.FormValue("password")
	if pass != c.Request.FormValue("confirmPassword") {
		return
	}
	req := pb.SignupRequest{
		Username: c.Request.FormValue("username"),
		Password: pass,
		Email:    c.Request.FormValue("email"),
	}
	res, err := u.client.Signup(ctx, &req)
	if err != nil {
		u.logger.Error(err.Error())
	}
	c.Writer.Write([]byte(res.Id))
}

func (u *userHandler) AllUserGet(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*500)
	defer cancel()
	list := &pb.Empty{}
	stream, err := u.client.UserList(ctx, list)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
	userList := []shared.User{}
	for {
		user, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			u.logger.Error(err.Error())
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		}
		u := shared.User{
			ID:       user.Id,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
			Status:   user.Status,
		}
		userList = append(userList, u)
	}
	c.JSON(200, userList)
}
