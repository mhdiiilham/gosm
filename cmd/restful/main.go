package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mhdiiilham/gosm/config"
	"github.com/mhdiiilham/gosm/database"
	"github.com/mhdiiilham/gosm/delivery"
	"github.com/mhdiiilham/gosm/logger"
	"github.com/mhdiiilham/gosm/pkg"
	"github.com/mhdiiilham/gosm/repository"
	"github.com/mhdiiilham/gosm/service"
	"github.com/mhdiiilham/gosm/thirdparty/kirimwa"
)

var version = "v0.0.1"

func main() {
	const ops = "main"
	var env string
	flag.StringVar(&env, "env", "local", "set the environment of the server")
	flag.Parse()

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer done()

	// when running inside docker container, we don't need to pas `-env` flag.
	if os.Getenv("APP_ENV") != "" {
		env = os.Getenv("APP_ENV")
	}

	logger.Infof(ctx, ops, "starting api (%s) env=%s", version, env)
	cfg, err := config.ReadConfiguration(ctx, env)
	if err != nil {
		panic(err)
	}

	logger.Infof(ctx, ops, "connecting to db")
	dbConn, err := database.ConnectPGSQL(cfg.Database.URL, cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, 60*time.Second)
	if err != nil {
		logger.Fatalf(ctx, ops, "connecting to db failed %v", err)
	}

	logger.Infof(ctx, ops, "starting echo...")
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.GET("/api", delivery.RootHandler(dbConn))

	// pkg here:
	passwordHasher := pkg.Hasher{}
	jwtToken := pkg.NewJwtGenerator(cfg.Name, cfg.JWTKey)
	kirimWaClient := kirimwa.NewKirimWAClient(cfg.Service.KirimWa.Key, cfg.Service.KirimWa.DeviceID)

	// Repositories here:
	userRepository := repository.NewUserRepository(dbConn)
	eventRepository := repository.NewEventRepository(dbConn)

	// Usecase here:
	authService := service.NewAuthorizationService(userRepository, passwordHasher, jwtToken)
	eventService := service.NewEventService(eventRepository, kirimWaClient, eventRepository.RunInTransactions)

	// register routes here:
	e.GET("/api/v1/public/guests", delivery.GetGuestByItShortID(eventService))

	middleware := delivery.NewMiddleware(jwtToken, userRepository)

	authHandler := delivery.NewAuthHandler(authService)
	authHandler.RegisterAuthRoutes(e.Group("api/v1/auth"))

	eventHandler := delivery.NewEventHandler(eventService)
	eventHandler.RegisterEventRoutes(e.Group("api/v1/events"), middleware)

	// Start server
	go func() {
		if err := e.Start(cfg.GetPort()); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	logger.Infof(ctx, ops, "closing db connection: %v", dbConn.Close())
	logger.Infof(ctx, ops, "server shutdown")
}
