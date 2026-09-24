package main

import (
	"acrocuit/internal/auth"
	"acrocuit/internal/handlers"
	"acrocuit/internal/response"
	"acrocuit/internal/setup"
	"acrocuit/internal/storage"
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucap9056/go-lifecycle/v2/lifecycle"
	"github.com/lucap9056/go-lifecycle/v2/runner"
)

func main() {
	err := runner.Run(func(lc *lifecycle.Coordinator) error {
		cfg := setup.Load()

		jwt := auth.NewJWTService(cfg.JWT)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		s, err := storage.New(ctx, cfg.Database.DSN)
		if err != nil {
			return err
		}
		lc.OnExit(s.Close)

		r := gin.New()
		r.Use(response.Recovery())
		handlers.New(r, s, jwt)

		server := &http.Server{Addr: cfg.HTTP.Addr, Handler: r}
		go func() {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				lc.Exitf("server error: %v", err)
			}
		}()

		lc.OnExit(func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				log.Printf("server shutdown: %v", err)
			}
		})

		return nil
	})

	if err != nil {
		log.Fatalf("acrocuit: %v", err)
	}
}
