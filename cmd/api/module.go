package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/fercho/school-tracking/services/auth/internal/core/auth"
	"github.com/fercho/school-tracking/services/auth/internal/infrastructure/api"
	"github.com/fercho/school-tracking/services/auth/internal/infrastructure/grpc"
	"github.com/fercho/school-tracking/services/auth/internal/infrastructure/persistence/postgres"
	"github.com/fercho/school-tracking/services/auth/pkg/env"
	"github.com/fercho/school-tracking/services/auth/pkg/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// AppModule assembles all fx modules and lifecycle hooks for the auth service.
func AppModule() fx.Option {
	return fx.Options(
		env.Module,
		logger.Module,
		auth.Module,
		api.Module,
		grpc.Module,
		postgres.Module,
		// Provide DB connection temporarily directly until a central DB pkg is made
		fx.Provide(
			func(cfg *env.Config, l *zap.Logger) (*gorm.DB, error) {
				preferSimple := false
				if strings.Contains(cfg.DatabaseURL, "-pooler") {
					preferSimple = true
				}
				db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{
					DSN:                  cfg.DatabaseURL,
					PreferSimpleProtocol: preferSimple,
				}), &gorm.Config{})
				if err != nil {
					l.Error("Failed to connect to database", zap.Error(err))
					return nil, err
				}
				l.Info("Connected to database successfully")
				return db, nil
			},
		),
		fx.Invoke(func(lc fx.Lifecycle, r *chi.Mux, cfg *env.Config, l *zap.Logger) {
			server := &http.Server{
				Addr:    ":" + cfg.HTTPPort,
				Handler: r,
			}

			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					l.Info("Starting HTTP server", zap.String("port", cfg.HTTPPort), zap.String("env", cfg.Environment))
					go func() {
						if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
							l.Error("HTTP server failed", zap.Error(err))
						}
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					l.Info("Stopping HTTP server")
					return server.Shutdown(ctx)
				},
			})
		}),
	)
}
