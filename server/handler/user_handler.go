package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	pb "github.com/aph138/shop/api/user_grpc"
	"github.com/aph138/shop/pkg/auth"
	"github.com/aph138/shop/server/web"
	"github.com/aph138/shop/shared"
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

var (
	JWTFile = "jwt.ed"
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
					c.Writer.Write([]byte("Wrong password or username"))
					return
				}
			}
		}
		u.logger.Error(err.Error())
		c.Writer.Write([]byte(RetryMSG))
		return
	}
	i, err := auth.NewIssuer(JWTFile)
	if err != nil {
		u.logger.Error(err.Error())
		c.Writer.Write([]byte(RetryMSG))
		return
	}
	data := map[string]string{"id": res.Id}
	token, err := i.Token(data, WEEK_IN_MINUTE)
	if err != nil {
		u.logger.Error(err.Error())
		c.Writer.Write([]byte(RetryMSG))
		return
	}

	c.SetCookie("jwt", token, WEEK_IN_SECOND, "", "", true, true)
	c.Writer.Header().Add("HX-Redirect", "/")
	c.Writer.Flush()

}

func (u *userHandler) SignupGet(c *gin.Context) {
	render(c, web.Signup())
}

func (u *userHandler) SignupPost(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*500)
	defer cancel()
	pass := c.Request.FormValue("password")
	email := c.Request.FormValue("email")
	_, err := mail.ParseAddress(email)
	if err != nil {
		c.Writer.Write([]byte("Wrong email"))
		return
	}
	if pass != c.Request.FormValue("confirmPassword") {
		c.Writer.Write([]byte("Wrong repeated password"))
		return
	}
	req := pb.SignupRequest{
		Username: c.Request.FormValue("username"),
		Password: pass,
		Email:    email,
	}
	res, err := u.client.Signup(ctx, &req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			u.logger.Info(e.Message())
			switch e.Code() {
			case codes.InvalidArgument:
				{
					if strings.Contains(e.Message(), "email") {
						c.Writer.Write([]byte("repeated email"))
						return
					} else if strings.Contains(e.Message(), "username") {
						c.Writer.Write([]byte("repeated username"))
						return
					} else {
						c.Writer.Write([]byte("invalid input"))
						return
					}
				}
			default:
				{
					u.logger.Error(err.Error())
					c.Writer.Write([]byte(RetryMSG))
					return
				}
			}
		}

		u.logger.Error(err.Error())
		c.Writer.Write([]byte(RetryMSG))
		return
	}
	i, err := auth.NewIssuer(JWTFile)
	if err != nil {
		u.logger.Error(err.Error())
		c.Writer.Write([]byte(RetryMSG))
		return
	}
	data := map[string]string{"id": res.Id}
	token, err := i.Token(data, WEEK_IN_MINUTE)
	if err != nil {
		u.logger.Error(err.Error())
		c.Writer.Write([]byte(RetryMSG))
		return
	}

	c.SetCookie("jwt", token, WEEK_IN_SECOND, "", "", true, true)
	c.Writer.Header().Add("HX-Redirect", "/")
	c.Writer.Flush()
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
