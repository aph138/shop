package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aph138/shop/api/common"
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

const (
	WEEK_IN_SECOND = 60 * 60 * 24 * 7
	WEEK_IN_MINUTE = 60 * 24 * 7
	RetryMSG       = "Something went wrong. Please try again later"
	GRPC_TIMEOUT   = time.Millisecond * 500
)

var (
	JWTFile = "jwt.ed"
	Redis   = os.Getenv("REDIS_ADDRESS")
)

type userHandler struct {
	client pb.UserClient
	logger *slog.Logger
	conn   *grpc.ClientConn
	rdb    *myredis.MyRedis
}

func NewUserHandler(url string, logger *slog.Logger) *userHandler {
	opt := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	userConn, err := grpc.Dial(url, opt...)
	if err != nil {
		logger.Error(err.Error())
	}
	rdb, err := myredis.NewRedis(&redis.Options{
		Addr:     Redis,
		Password: "",
		DB:       0,
	})
	if err != nil {
		//TODO: do something? retry? get data directly from db?
		logger.Error(err.Error())
	}
	return &userHandler{
		logger: logger,
		client: pb.NewUserClient(userConn),
		conn:   userConn,
		rdb:    rdb,
	}
}

func (u *userHandler) Close() error {
	return u.conn.Close()
}

func (u *userHandler) GetSignin(c *gin.Context) {
	render(c, userview.Signin(&shared.User{}))
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
					c.String(http.StatusBadRequest, "Wrong password or username")
				}
			}
			return
		}
		u.logger.Error(err.Error())
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}
	i, err := auth.NewIssuer(JWTFile)
	if err != nil {
		u.logger.Error(err.Error())
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}
	data := map[string]string{"id": res.Value}
	token, err := i.Token(data, WEEK_IN_MINUTE)
	if err != nil {
		u.logger.Error(err.Error())
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}

	c.SetCookie("jwt", token, WEEK_IN_SECOND, "", "", true, true)
	c.Writer.Header().Add("HX-Redirect", "/")
	c.Writer.Flush()

}

func (u *userHandler) GetSignup(c *gin.Context) {
	render(c, userview.Signup(&shared.User{}))
}

func (u *userHandler) PostSignup(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*500)
	defer cancel()
	pass := c.Request.FormValue("password")
	email := c.Request.FormValue("email")
	_, err := mail.ParseAddress(email)
	if err != nil {
		c.String(http.StatusInternalServerError, "invalid email")
		return
	}
	if pass != c.Request.FormValue("confirmPassword") {
		c.String(http.StatusBadRequest, "passwords don't match")
		return
	}
	req := pb.SignupRequest{
		Username: c.Request.FormValue("username"),
		Role:     0,
		Password: pass,
		Email:    email,
	}
	res, err := u.client.Signup(ctx, &req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.AlreadyExists:
				{
					if strings.Contains(e.Message(), "email") {
						c.String(http.StatusBadRequest, "this email already exists")
					} else if strings.Contains(e.Message(), "username") {
						c.String(http.StatusInternalServerError, "this username already exists")
					} else {
						c.String(http.StatusBadRequest, "invalid input")
					}
				}
			default:
				{
					u.logger.Error(err.Error())
					c.String(http.StatusInternalServerError, RetryMSG)
				}
			}
			return
		}

		u.logger.Error(err.Error())
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}
	i, err := auth.NewIssuer(JWTFile)
	if err != nil {
		u.logger.Error(err.Error())
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}
	data := map[string]string{"id": res.Value}
	token, err := i.Token(data, WEEK_IN_MINUTE)
	if err != nil {
		u.logger.Error(err.Error())
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}

	c.SetCookie("jwt", token, WEEK_IN_SECOND, "", "", true, true)
	c.Writer.Header().Add("HX-Redirect", "/")
	c.Writer.Flush()
}

func (u *userHandler) GetAllUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*500)
	defer cancel()
	list := &common.Empty{}
	stream, err := u.client.UserList(ctx, list)
	if err != nil {
		u.logger.Error(err.Error())
		http.Error(c.Writer, RetryMSG, http.StatusInternalServerError)
		return
	}
	userList := []shared.User{}
	for {
		user, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			u.logger.Error(err.Error())
			http.Error(c.Writer, RetryMSG, http.StatusInternalServerError)
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
	render(c, userview.UserList(userList, getUserCtx(c)))
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
		id := data["id"]
		var user shared.User

		err = u.rdb.Get(id, &user)
		if err != nil {
			if err == redis.Nil {
				res, err := u.client.GetUser(context.TODO(), &common.StringMessage{Value: id})
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
				err = u.rdb.Set(id, user)
				if err != nil {
					u.logger.Error(err.Error())
				}
			} else {
				u.logger.Error(err.Error())
			}
		}
		ctx := context.WithValue(c.Request.Context(), ctxUserInfo, user)
		c.Request = c.Request.WithContext(ctx)
		c.Next()

	}
}
func (u *userHandler) DeleteUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), GRPC_TIMEOUT)
	defer cancel()
	res, err := u.client.DeleteUser(ctx, &common.StringMessage{Value: c.Param("id")})
	if err != nil {
		u.logger.Error(err.Error())
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}
	if res.Value {
		c.Writer.WriteHeader(http.StatusOK)
	} else {
		c.Writer.WriteHeader(http.StatusInternalServerError)
	}
	c.Writer.Flush()
}
func (u *userHandler) GetUserProfile(c *gin.Context) {
	render(c, userview.Profile(getUserCtx(c)))
}
func (u *userHandler) PutUserProfile(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), GRPC_TIMEOUT)
	defer cancel()
	email := c.Request.FormValue("email")
	address := c.Request.FormValue("address")
	phone := c.Request.FormValue("phone")
	//^(0|\+98)? -> optional start with 0 or +98
	//(9\d{9})$ -> required 10 digit string that starts with 9
	reg := regexp.MustCompile(`^(0|\+98)?(9\d{9})$`)
	if !reg.MatchString(phone) {
		c.String(http.StatusBadRequest, "invalid phone number")
		return
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid email")
		return
	}
	user := getUserCtx(c)
	add := &pb.Address{Address: address, Phone: phone}
	res, err := u.client.EditUser(ctx, &pb.EditUserRequest{Id: user.ID, Email: email, Address: add})
	if err != nil {
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.AlreadyExists:
				{
					c.String(http.StatusBadRequest, "email already exists")
				}
			case codes.InvalidArgument:
				{
					c.String(http.StatusBadRequest, "invalid id")
				}
			default:
				{
					u.logger.Error(err.Error())
					c.String(http.StatusInternalServerError, RetryMSG)
				}
			}
			return
		}
		u.logger.Error(err.Error())
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}
	if res.Value {
		user.Address.Address = address
		user.Address.Phone = phone
		user.Email = email
		u.rdb.Set(user.ID, user)
		c.String(http.StatusOK, "updated")
		return
	} else {
		c.String(http.StatusOK, "not updated")
		return
	}

}
func (u *userHandler) GetPassword(c *gin.Context) {
	render(c, userview.Password(getUserCtx(c)))
}
func (u *userHandler) PutPassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), GRPC_TIMEOUT)
	defer cancel()

	oldPass := c.Request.FormValue("oldPassword")
	newPass := c.Request.FormValue("newPassword")
	confirmPass := c.Request.FormValue("confirmPassword")
	if newPass != confirmPass {
		c.String(http.StatusBadRequest, "passwords don't match")
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
					c.String(http.StatusBadRequest, "invalid password")
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
	if signinResult.Value {
		c.String(http.StatusOK, "password has changed")
	} else {
		c.String(http.StatusOK, "password hasn't changed. something went wrong.")
	}

}

func (u *userHandler) GetCart(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), GRPC_TIMEOUT)
	defer cancel()
	user_id := getUserCtx(c).ID
	res, err := u.client.Cart(ctx, &common.StringMessage{Value: user_id})
	if err != nil {
		//TODO: check other errors
		c.String(http.StatusInternalServerError, RetryMSG)
	}
	//TODO: total price
	var result []shared.Item
	for _, i := range res.Item {
		result = append(result, shared.Item{
			ID:     i.Id,
			Link:   i.Link,
			Name:   i.Name,
			Poster: i.Poster,
			Price:  i.Price,
		})
	}
	render(c, userview.Cart(result, getUserCtx(c)))
}
func (u *userHandler) PostCart(c *gin.Context) {
	q := c.Request.FormValue("quantity")
	quantity, err := strconv.Atoi(q)
	if err != nil {
		u.logger.Error(err.Error())
		c.String(http.StatusBadRequest, "invalid quantity value")
		return
	}
	//check quantity
	if quantity <= 0 || quantity > 100 {
		c.String(http.StatusBadRequest, "quantity cannot be negative or greater than 100")
		return
	}
	itemID := c.Request.FormValue("id")
	if itemID == "" {
		c.String(http.StatusBadRequest, "empty id")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), GRPC_TIMEOUT)
	defer cancel()

	userID := getUserCtx(c).ID
	result, err := u.client.AddToCart(ctx, &pb.AddToCartRequest{UserId: userID, ItemId: itemID, Quntity: uint32(quantity)})
	if err != nil {
		//TODO: check other errors
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}
	if !result.Value {
		c.String(http.StatusNotModified, "did'nt add") //TODO: why couldn't add
		return
	}
	c.String(http.StatusCreated, "OK")

}
func (u *userHandler) DeleteCart(c *gin.Context) {
	userID := getUserCtx(c).ID
	itemID := c.Param("id")
	ctx, cancel := context.WithTimeout(context.Background(), GRPC_TIMEOUT)
	defer cancel()
	result, err := u.client.DeleteFromCart(ctx, &pb.DeleteFromCartRequest{UserId: userID, ItemId: itemID})
	if err != nil {
		//TODO: check other errors
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}
	if !result.Value {
		c.String(http.StatusNotModified, "didn't delete") // TODO: message
		return
	}
	c.Status(http.StatusOK)
}
func (u *userHandler) PutCart(c *gin.Context) {
	userID := getUserCtx(c).ID
	itemID := c.Request.PostFormValue("id")
	quantity := uint32(1)

	ctx, cancel := context.WithTimeout(context.Background(), GRPC_TIMEOUT)
	defer cancel()

	result, err := u.client.UpdateCart(ctx, &pb.UpdateCartRequest{UserId: userID, ItemId: itemID, NewQuantity: quantity})
	if err != nil {
		//TODO: check
		c.String(http.StatusInternalServerError, RetryMSG)
		return
	}
	if !result.Value {
		c.String(http.StatusNotModified, "") //TODO
		return
	}
	c.Status(http.StatusOK)
	c.Writer.Flush()
}

// TODO:
func getUserCtx(c *gin.Context) *shared.User {
	u, ok := c.Request.Context().Value(ctxUserInfo).(shared.User)
	if !ok {
		return &shared.User{}
	}
	return &u
}
func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if getUserCtx(ctx).Role > 0 {
			ctx.Next()
		} else {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not allowed"})
			return
		}
	}
}
