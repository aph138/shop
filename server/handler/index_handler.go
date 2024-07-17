package handler

import (
	"log/slog"

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
	render(c, web.Index(getUserCtx(c)))

}
