package handler

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
)

func render(c *gin.Context, t templ.Component) {
	c.Writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)
	if err := t.Render(c.Request.Context(), c.Writer); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

type Handler interface {
	Close() error
}
