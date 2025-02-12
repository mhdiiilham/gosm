package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mhdiiilham/gosm/database"
	"github.com/mhdiiilham/gosm/delivery"
	"github.com/mhdiiilham/gosm/logger"
	"github.com/mhdiiilham/gosm/pkg"
	"github.com/mhdiiilham/gosm/repository"
	"github.com/mhdiiilham/gosm/server"
	"github.com/mhdiiilham/gosm/service"
)

func main() {
	// Server configuration here:
	const ops = "main"
	const port = "9091"
	dbURL := os.Getenv("DATABASE_URL")
	jwtSignature := os.Getenv("JWT_KEY")

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		done()
		if r := recover(); r != nil {
			panic(fmt.Sprintf("application panic: %v", r))
		}
	}()

	logger.Info(ctx, ops, "starting api...")

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

	srv, err := server.New(port)
	if err != nil {
		logger.Fatalf(ctx, ops, "failed to create new server: %v", err)
	}

	srv.ServeHTTPHandler(ctx, e)
	logger.Info(ctx, ops, "closing db connection: %v", dbConn.Close())
	logger.Info(ctx, ops, "server shutdown")
}
