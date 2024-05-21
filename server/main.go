package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aph138/shop/server/handler"
	"github.com/gin-gonic/gin"
)

// //go:embed public
// var public embed.FS

var (
	logger *slog.Logger
)

func init() {
	logger = slog.Default()
}

func main() {
	h := gin.Default()

	userHandler := handler.NewUserHandler(os.Getenv("USER_SERVICE"), logger)
	indexHandler := handler.NewIndexHandler(logger)

	h.Use(indexHandler.AuthMiddleware())

	h.StaticFS("/public", http.Dir("./public"))
	h.GET("/", indexHandler.IndexGet)
	h.GET("/signin", userHandler.SigninGet)
	h.POST("/signin", userHandler.SigninPost)
	h.GET("/signup", userHandler.SignupGet)
	h.GET("/list", userHandler.AllUserGet)
	h.POST("/signup", userHandler.SignupPost)
	h.GET("/test", func(ctx *gin.Context) {
		ctx.Writer.Write([]byte(time.Now().String()))
	})
	srv := &http.Server{
		Addr:    os.Getenv("ADDRESS"),
		Handler: h.Handler(),
	}

	go func() {
		logger.Info("Server is running ...")
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	logger.Info("staring graceful shutdown")
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5000)
	defer cancel()
	var err error
	if err = srv.Shutdown(ctx); err != nil {
		logger.Error(err.Error())
	}
	if err = userHandler.Close(); err != nil {
		logger.Error(err.Error())
	}
	logger.Info("server shutted down")

}
