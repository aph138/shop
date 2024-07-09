package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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

	userHandler := handler.NewUserHandler(os.Getenv("USER_SERVICE"), logger)
	stockHandler, err := handler.NewStockHandler(os.Getenv("STOCK_SERVICE"), logger)
	if err != nil {
		logger.Error("err when creating stock handler: " + err.Error())
	}
	indexHandler := handler.NewIndexHandler(logger)
	handlers := []handler.Handler{
		userHandler,
		stockHandler,
	}

	h := gin.Default()
	// set max size for uplaod
	h.MaxMultipartMemory = 8 >> 10 // 10MB

	h.Use(userHandler.AuthMiddleware())

	h.Static("/public", "./public")
	h.GET("/img/:folder/:file", serveImage)
	h.GET("/", indexHandler.IndexGet)
	//user handlers
	h.GET("/signin", userHandler.GetSignin)
	h.POST("/signin", userHandler.PostSignin)
	h.GET("/signup", userHandler.GetSignup)
	h.POST("/signup", userHandler.PostSignup)
	h.GET("/profile", userHandler.GetUserProfile)
	h.GET("/password", userHandler.GetPassword)
	h.PUT("/profile", userHandler.PutUserProfile)
	h.PUT("/password", userHandler.PutPassword)
	// stock handlers
	h.GET("/item/:name", stockHandler.GetItem)
	h.GET("/item", stockHandler.GetAll)
	h.POST("/cart", userHandler.PostCart)
	h.GET("/cart", userHandler.GetCart)
	h.DELETE("/cart/:id", userHandler.DeleteCart)
	h.PUT("/cart", userHandler.PutCart)

	//restricted functions
	adminGroupe := h.Group("/admin").Use(handler.AdminMiddleware())
	adminGroupe.DELETE("/delete/:id", userHandler.DeleteUser)
	adminGroupe.GET("/list", userHandler.GetAllUser)
	adminGroupe.POST("/item", stockHandler.PostAddItem)
	adminGroupe.GET("/item", stockHandler.GetAddItem)

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
	if err = srv.Shutdown(ctx); err != nil {
		logger.Error(err.Error())
	}
	for _, h := range handlers {
		if err = h.Close(); err != nil {
			logger.Error(err.Error())
		}
	}
	logger.Info("server shutted down")

}
func serveImage(c *gin.Context) {
	path := "uploads"
	folder := c.Param("folder")
	file := c.Param("file")
	safeFile := filepath.Base(file)
	finalPath := filepath.Join(path, folder, safeFile)
	// check if path is safe
	if !strings.HasPrefix(finalPath, path) {
		c.String(http.StatusBadRequest, "Bad Request")
		return
	}
	c.File(finalPath)

}
