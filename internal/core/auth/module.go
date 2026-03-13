package auth

import "go.uber.org/fx"

// Module provides the core auth service and JWT manager to the fx dependency graph.
var Module = fx.Module("core.auth",
	fx.Provide(
		NewAuthService,
		NewJWTManager,
	),
)
