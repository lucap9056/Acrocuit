package main

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/handlers"
	"acrocuit/internal/logs"
	"acrocuit/internal/options"
	"acrocuit/internal/response"
	"acrocuit/internal/setup"
	"acrocuit/internal/storage"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucap9056/go-lifecycle/v2/lifecycle"
	"github.com/lucap9056/go-lifecycle/v2/runner"
	"go.uber.org/zap"
)

func main() {
	cfg := setup.Load()

	logs.InitLogger(cfg.Logging)
	defer logs.Sync()

	for _, warning := range cfg.Warnings {
		logs.Out.Warn(warning)
	}

	if !cfg.Options.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	err := runner.Run(func(lc *lifecycle.Coordinator) error {
		requireIdentity, err := auth.RequireIdentity(cfg.Options)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		s, err := storage.New(ctx, cfg.Options)
		if err != nil {
			return err
		}
		lc.OnExit(s.Close)

		r := gin.New()
		r.Use(logs.RequestLogger(), response.Recovery())
		handlers.New(r, s, requireIdentity, cfg.Options)

		server := &http.Server{Addr: cfg.HTTP.Addr, Handler: r}
		go func() {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				lc.Exitf("server error: %v", err)
			}
		}()
		startupFields := []zap.Field{
			zap.String("addr", cfg.HTTP.Addr),
			zap.String("environment", string(cfg.Options.Environment)),
			zap.String("auth", string(cfg.Options.Auth.Mode)),
			zap.String("database", string(cfg.Options.Database.Driver)),
		}
		if cfg.Options.Database.Driver == options.DriverSQLite {
			startupFields = append(startupFields, zap.String("database_path", cfg.Options.Database.Source))
		}
		logs.Out.Info("server started", startupFields...)

		lc.OnExit(func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				logs.Out.Error("server shutdown failed", zap.Error(err))
			}
		})

		return nil
	})

	if err != nil {
		logs.Out.Fatal("acrocuit exited with error", zap.Error(err))
	}
}
