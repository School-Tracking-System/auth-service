package api

import (
	"github.com/fercho/school-tracking/services/auth/internal/infrastructure/api/controllers"
	"go.uber.org/fx"
)

// Module provides the HTTP controllers and router to the fx dependency graph.
var Module = fx.Module("infrastructure.api",
	fx.Provide(
		controllers.NewAuthController,
		NewRouter,
	),
)
