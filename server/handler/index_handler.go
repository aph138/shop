package handler

import (
	"log/slog"

	"github.com/aph138/shop/server/web"
	"github.com/aph138/shop/shared"
	"github.com/gin-gonic/gin"
)

type indexHandler struct {
	logger *slog.Logger
}

func NewIndexHandler(logger *slog.Logger) *indexHandler {
	return &indexHandler{logger: logger}
}
func (i *indexHandler) IndexGet(c *gin.Context) {
	user, ok := c.Request.Context().Value(ctxUserInfo).(shared.User)
	if ok {
		render(c, web.Index(user))
	} else {
		render(c, web.Index(shared.User{}))
	}

}
