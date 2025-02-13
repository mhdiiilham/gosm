package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mhdiiilham/gosm/database"
	"github.com/mhdiiilham/gosm/delivery"
	"github.com/mhdiiilham/gosm/logger"
	"github.com/mhdiiilham/gosm/pkg"
	"github.com/mhdiiilham/gosm/repository"
	"github.com/mhdiiilham/gosm/service"
)

var version = "v0.0.1"
var env = "local"

func main() {
	// Server configuration here:
	const ops = "main"
	const port = ":9091"
	dbURL := os.Getenv("DATABASE_URL")
	jwtSignature := os.Getenv("JWT_KEY")

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer done()

	logger.Info(ctx, ops, "starting api (%s) env=%s", version, env)

	logger.Info(ctx, ops, "connecting to db")
	dbConn, err := database.ConnectPGSQL(dbURL)
	if err != nil {
		logger.Fatalf(ctx, ops, "connecting to db failed %v", err)
	}

	logger.Info(ctx, ops, "starting echo...")
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.GET("/api", delivery.RootHandler())

	// pkg here:
	passwordHasher := pkg.Hasher{}
	jwtToken := pkg.NewJwtGenerator("gosm", jwtSignature)

	// Repositories here:
	userRepository := repository.NewUserRepository(dbConn)

	// Usecase here:
	authService := service.NewAuthorizationService(userRepository, passwordHasher, jwtToken)

	authHandler := delivery.NewAuthHandler(authService)
	authHandler.RegisterAuthRoutes(e.Group("api/v1/auth"))

	// Start server
	go func() {
		if err := e.Start(port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	logger.Info(ctx, ops, "closing db connection: %v", dbConn.Close())
	logger.Info(ctx, ops, "server shutdown")
}
