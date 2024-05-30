package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	pb "github.com/aph138/shop/api/user_grpc"
	"github.com/aph138/shop/pkg/auth"
	"github.com/aph138/shop/pkg/myredis"
	"github.com/aph138/shop/server/web/userview"
	"github.com/aph138/shop/shared"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// TODO: add address
const (
	WEEK_IN_SECOND = 60 * 60 * 24 * 7
	WEEK_IN_MINUTE = 60 * 24 * 7
	RetryMSG       = "Something went wrong. Please try again later"
	GRPC_TIMEOUT   = time.Millisecond * 500
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

func (u *userHandler) GetSignin(c *gin.Context) {
	render(c, userview.Signin())
}

func (u *userHandler) PostSignin(c *gin.Context) {
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

func (u *userHandler) GetSignup(c *gin.Context) {
	render(c, userview.Signup())
}

func (u *userHandler) PostSignup(c *gin.Context) {
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
			//TODO: change code to alreadytexists
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

func (u *userHandler) GetAllUser(c *gin.Context) {
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
	render(c, userview.UserList(userList))
}

type ctxKey string

var ctxUserInfo ctxKey = ctxKey("userInfo")

func (u *userHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//exclude public path from middleware
		if strings.Split(c.Request.URL.Path, "/")[1] == "public" {
			return
		}
		v, err := auth.NewValidator("jwt.ed.pub")
		if err != nil {
			u.logger.Error(err.Error())
			http.Error(c.Writer, RetryMSG, http.StatusInternalServerError)
			return
		}
		token, err := c.Request.Cookie("jwt")
		if err != nil {
			if err == http.ErrNoCookie {
				c.Next()
				return
			}
			u.logger.Error(err.Error())
			http.Error(c.Writer, RetryMSG, http.StatusInternalServerError)
			return
		}
		if token.Value == "" {
			c.Next()
			return
		}
		data, err := v.Validate(token.Value)
		if err != nil {
			u.logger.Error(err.Error())
			c.Next()
			return
		}
		/*
			Note: this approach is not optimized; instead we should save id and
			username in the context, and whenever the user requests for a restricted
			function, check for user's privilages in database. But in this app I want
			use reddis and here is a good opportunity.
		*/
		t := time.Now()
		id := data["id"]
		var user shared.User

		rdb, err := myredis.NewRedis(&redis.Options{
			Addr:     os.Getenv("REDIS_ADDRESS"),
			Password: "",
			DB:       0,
		})
		if err != nil {
			//TODO: do something? retry? get data directly from db?
			u.logger.Error(err.Error())
		}
		err = rdb.Get(id, &user)
		if err != nil {
			if err == redis.Nil {
				log.Println("Get from db")
				res, err := u.client.GetUser(context.TODO(), &pb.WithID{Id: id})
				if err != nil {
					u.logger.Error(err.Error())
					c.Next()
					return
				}
				user = shared.User{
					ID:       data["id"],
					Username: res.Username,
					Email:    res.Email,
					Role:     res.Role,
					Status:   res.Status,
					Address: shared.Address{
						Address: res.Address.Address,
						Phone:   res.Address.Phone,
					},
				}
				err = rdb.Set(id, user)
				if err != nil {
					u.logger.Error(err.Error())
				}
			} else {
				u.logger.Error(err.Error())
			}
		}
		s := time.Since(t)
		log.Printf("TOOK: %s", s.String())
		ctx := context.WithValue(c.Request.Context(), ctxUserInfo, user)
		r := c.Request.WithContext(ctx)
		c.Request = r
		c.Next()

	}
}
func (u *userHandler) DeleteUser(c *gin.Context) {
	u.logger.Info(c.Param("id"))
	ctx, cancel := context.WithTimeout(context.Background(), GRPC_TIMEOUT)
	defer cancel()
	res, err := u.client.DeleteUser(ctx, &pb.WithID{Id: c.Param("id")})
	if err != nil {
		u.logger.Error(err.Error())
		c.Writer.Write([]byte(RetryMSG))
		return
	}
	u.logger.Info(fmt.Sprintf("%t", res.Result))
	if res.Result {
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Flush()
	} else {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		c.Writer.Flush()
	}
}
func (u *userHandler) GetUserProfile(c *gin.Context) {
	render(c, userview.Profile(*getUserCtx(c)))
}
func (u *userHandler) PutUserProfile(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), GRPC_TIMEOUT)
	defer cancel()
	e := c.Request.FormValue("email")
	address := c.Request.FormValue("address")
	phone := c.Request.FormValue("phone")
	//TODO: validate phone number
	_, err := mail.ParseAddress(e)
	if err != nil {
		c.String(http.StatusBadRequest, "wrong email")
		return
	}
	add := &pb.Address{Address: address, Phone: phone}
	res, err := u.client.EditUser(ctx, &pb.EditUserRequest{Id: getUserCtx(c).ID, Email: e, Address: add})
	if err != nil {
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.AlreadyExists:
				{
					c.String(http.StatusBadRequest, "email already exists")
					return
				}
			case codes.InvalidArgument:
				{
					c.String(http.StatusBadRequest, "wrong id")
					return
				}
			default:
				{
					u.logger.Error(err.Error())
					c.JSON(http.StatusInternalServerError, RetryMSG)
					return
				}
			}
		}
		u.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, RetryMSG)
		return
	}
	if res.Result {
		c.String(http.StatusOK, "updated")
		return
	} else {
		c.String(http.StatusOK, "not updated.try again later")
		return
	}

}
func (u *userHandler) GetPassword(c *gin.Context) {
	render(c, userview.Password())
}
func (u *userHandler) PutPassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), GRPC_TIMEOUT)
	defer cancel()

	oldPass := c.Request.FormValue("oldPassword")
	newPass := c.Request.FormValue("newPassword")
	confirmPass := c.Request.FormValue("confirmPassword")
	if newPass != confirmPass {
		c.String(http.StatusBadRequest, "re-enter passwrd")
		return
	}
	req := &pb.ChangePasswordRequest{Id: getUserCtx(c).ID, OldPassword: oldPass, NewPassword: newPass}
	signinResult, err := u.client.ChangePassword(ctx, req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.Unauthenticated:
				{
					c.String(http.StatusBadRequest, "wrong password")
				}
			case codes.InvalidArgument:
				{
					c.String(http.StatusBadRequest, "bad new password")
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
	if signinResult.Result {
		c.String(http.StatusOK, "password has changed")
		return
	} else {
		c.String(http.StatusOK, "password hasn't changed. something went wrong.")
		return
	}

}

func getUserCtx(c *gin.Context) *shared.User {
	u, ok := c.Request.Context().Value(ctxUserInfo).(shared.User)
	if !ok {
		return &shared.User{}
	}
	return &u
}
func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.Println("ROOOLEEE", getUserCtx(ctx).Role)
		if getUserCtx(ctx).Role > 0 {
			ctx.Next()
		} else {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not allowed"})
			return
		}
	}
}
