package server

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
	appmiddleware "waxp/echo/internal/middleware"
)

type Server struct {
	Echo   *echo.Echo
	Config *config.Config
	Pool   *pgxpool.Pool
}

func New(cfg *config.Config, pool *pgxpool.Pool) *Server {
	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit(10 * 1024 * 1024))

	if cfg.Env == "development" {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     []string{"http://localhost:5173"},
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
			AllowCredentials: true,
		}))
	}

	queries := db.New(pool)
	authHandler := handler.NewAuthHandler(queries, cfg.JWTSecret)
	siteHandler := handler.NewSiteHandler(queries, pool)
	localeHandler := handler.NewLocaleHandler(queries)

	e.GET("/health", handler.Health)

	api := e.Group("/api")

	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	protected := api.Group("", appmiddleware.JWTAuth(cfg.JWTSecret))
	protected.GET("/me", authHandler.Me)

	sites := protected.Group("/sites")
	sites.POST("", siteHandler.CreateWithDefaults)
	sites.GET("", siteHandler.List)
	sites.GET("/:id", siteHandler.GetByID)
	sites.PUT("/:id", siteHandler.Update)
	sites.DELETE("/:id", siteHandler.Delete)
	sites.POST("/:id/locales", localeHandler.Add)
	sites.DELETE("/:id/locales/:localeCode", localeHandler.Remove)

	pageHandler := handler.NewPageHandler(queries, pool)
	sectionHandler := handler.NewSectionHandler(queries)
	mediaHandler := handler.NewMediaHandler(queries, cfg.MediaDir)
	sites.POST("/:id/pages", pageHandler.Create)
	sites.GET("/:id/pages", pageHandler.List)
	sites.GET("/:id/pages/:pageId", pageHandler.GetByID)
	sites.PUT("/:id/pages/:pageId", pageHandler.Update)
	sites.DELETE("/:id/pages/:pageId", pageHandler.Delete)
	sites.GET("/:id/routes", pageHandler.Routes)
	sites.POST("/:id/sections/next-id", sectionHandler.GetNextSectionID)
	sites.POST("/:id/blocks/next-id", handler.NewBlockHandler(queries).GetNextBlockID)

	media := protected.Group("/media")
	media.POST("", mediaHandler.Upload)
	media.GET("", mediaHandler.List)
	media.GET("/:id", mediaHandler.GetByID)
	media.DELETE("/:id", mediaHandler.Delete)

	e.GET("/media/:name", handler.ServeMedia(cfg.MediaDir))

	return &Server{
		Echo:   e,
		Config: cfg,
		Pool:   pool,
	}
}

func (s *Server) Start(ctx context.Context) error {
	slog.Info("starting server", "port", s.Config.ServerPort, "env", s.Config.Env)

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
		slog.Info("shutting down server")
		shutdownCtx := context.Background()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
