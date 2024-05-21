package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/aph138/shop/pkg/auth"
	"github.com/aph138/shop/server/web"
	"github.com/gin-gonic/gin"
)

type indexHandler struct {
	logger *slog.Logger
}

func NewIndexHandler(logger *slog.Logger) *indexHandler {
	return &indexHandler{logger: logger}
}
func (i *indexHandler) IndexGet(c *gin.Context) {
	id := c.Request.Context().Value(ctxKey("id"))
	result := false
	if id != nil {
		result = true
	}
	render(c, web.Index(result))
}

type ctxKey string

func (i *indexHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, err := auth.NewValidator("jwt.ed.pub")
		if err != nil {
			i.logger.Error(err.Error())
			http.Error(c.Writer, RetryMSG, http.StatusInternalServerError)
			return
		}
		token, err := c.Request.Cookie("jwt")
		if err != nil {
			if err == http.ErrNoCookie {
				c.Next()
				return
			}
			i.logger.Error(err.Error())
			http.Error(c.Writer, RetryMSG, http.StatusInternalServerError)
			return
		}
		if token.Value == "" {
			c.Next()
			return
		}
		data, err := v.Validate(token.Value)
		if err != nil {
			i.logger.Error(err.Error())
			c.Next()
			return
		}
		ctx := context.WithValue(c.Request.Context(), ctxKey("id"), data["id"])
		r := c.Request.WithContext(ctx)
		c.Request = r
		c.Next()

	}
}
