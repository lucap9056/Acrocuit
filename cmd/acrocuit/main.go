package main

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/handlers"
	"acrocuit/internal/logs"
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

	if cfg.Environment != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	err := runner.Run(func(lc *lifecycle.Coordinator) error {
		jwt := auth.NewJWTService(cfg.JWT)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		s, err := storage.New(ctx, cfg.Database.DSN)
		if err != nil {
			return err
		}
		lc.OnExit(s.Close)

		r := gin.New()
		r.Use(logs.RequestLogger(), response.Recovery())
		handlers.New(r, s, jwt)

		server := &http.Server{Addr: cfg.HTTP.Addr, Handler: r}
		go func() {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				lc.Exitf("server error: %v", err)
			}
		}()
		logs.Out.Info("server started", zap.String("addr", cfg.HTTP.Addr), zap.String("environment", cfg.Environment))

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
