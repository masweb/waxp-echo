package renderer

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"waxp/echo/internal/config"
	"waxp/echo/internal/db"
	"waxp/echo/internal/handler"
)

type Server struct {
	Echo   *echo.Echo
	Config *config.RenderConfig
	Pool   *pgxpool.Pool
}

func New(cfg *config.RenderConfig, pool *pgxpool.Pool) *Server {
	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	queries := db.New(pool)

	mediaBase := ""
	if cfg.Env == "development" {
		mediaBase = "http://localhost" + cfg.ServerPort
	}

	publicHandler := handler.NewPublicHandler(queries, pool, mediaBase)

	e.GET("/health", handler.Health)
	e.GET("/media/:name", handler.ServeMedia(cfg.MediaDir))
	e.GET("/*", publicHandler.ServePage)

	return &Server{
		Echo:   e,
		Config: cfg,
		Pool:   pool,
	}
}

func (s *Server) Start(ctx context.Context) error {
	slog.Info("starting render server", "port", s.Config.ServerPort, "env", s.Config.Env)

	server := &http.Server{
		Addr:    s.Config.ServerPort,
		Handler: s.Echo,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutting down render server")
		shutdownCtx := context.Background()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
