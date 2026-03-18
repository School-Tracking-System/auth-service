package grpc

import (
	"context"
	"fmt"
	"net"

	authv1 "github.com/fercho/school-tracking/proto/gen/auth/v1"
	"github.com/fercho/school-tracking/services/auth/internal/infrastructure/grpc/handlers"
	"github.com/fercho/school-tracking/services/auth/pkg/env"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Module provides the gRPC server and handlers to the fx dependency graph.
var Module = fx.Module("infrastructure.grpc",
	fx.Provide(
		handlers.NewAuthHandler,
		NewGRPCServer,
	),
	fx.Invoke(registerHooks),
)

func NewGRPCServer(
	cfg *env.Config,
	logger *zap.Logger,
	authHandler *handlers.AuthHandler,
) *grpc.Server {
	server := grpc.NewServer()

	// Register services
	authv1.RegisterAuthServiceServer(server, authHandler)

	// Enable reflection for tools like grpcurl
	if cfg.Environment != "production" {
		reflection.Register(server)
	}

	return server
}

func registerHooks(lc fx.Lifecycle, server *grpc.Server, cfg *env.Config, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			port := cfg.GRPCPort
			if port == "" {
				port = "9090"
			}

			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
			if err != nil {
				return fmt.Errorf("failed to listen on gRPC port %s: %w", port, err)
			}

			logger.Info("Starting gRPC server", zap.String("port", port))

			go func() {
				if err := server.Serve(lis); err != nil {
					logger.Error("gRPC server failed", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping gRPC server gracefully")
			server.GracefulStop()
			return nil
		},
	})
}
