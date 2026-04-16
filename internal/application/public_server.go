package application

import (
	"context"
	"errors"
	"fmt"
	"subscriptions-service/internal"
	"subscriptions-service/internal/lib/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

const apiV1 = "/api/v1"

type PublicServer struct {
	cfg    internal.AppConfig
	echo   *echo.Echo
	logger logger.Logger
}

func NewPublicServer(cfg internal.AppConfig, log logger.Logger) *PublicServer {
	return &PublicServer{cfg: cfg, logger: log}
}

func (s *PublicServer) Configure(container *Container) (*PublicServer, error) {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogMethod:   true,
		LogError:    true,
		LogLatency:  true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			s.logger.Info("http_request",
				"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency", v.Latency,
				"error", v.Error,
			)
			return nil
		},
	}))

	s.echo = e
	s.echo.GET("/swagger/*", echoSwagger.WrapHandler)
	s.v1(container)

	return s, nil
}

func (s *PublicServer) Start() error {
	if s.echo == nil {
		return errors.New("echo is not initialized")
	}
	return s.echo.Start(fmt.Sprintf(":%s", s.cfg.PublicServerConfig.Port))
}

func (s *PublicServer) ShutDown(ctx context.Context) error {
	if s.echo == nil {
		return errors.New("echo is not initialized")
	}
	return s.echo.Shutdown(ctx)
}

func (s *PublicServer) v1(container *Container) {
	v1 := s.echo.Group(apiV1)

	h := container.SubscriptionHandler

	v1.GET("/subscriptions/sum", h.Sum)
	v1.GET("/subscriptions", h.List)
	v1.POST("/subscriptions", h.Create)
	v1.GET("/subscriptions/:id", h.GetByID)
	v1.PUT("/subscriptions/:id", h.Update)
	v1.DELETE("/subscriptions/:id", h.Delete)
}
