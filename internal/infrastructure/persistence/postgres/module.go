package postgres

import "go.uber.org/fx"

// Module provides the PostgreSQL repository implementations to the fx dependency graph.
var Module = fx.Module("infrastructure.persistence.postgres",
	fx.Provide(
		NewUserRepository,
	),
)
