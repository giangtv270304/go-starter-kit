package repo

import (
	"github.com/andyle182810/gframework/postgres"
)

type PostgresRepo struct {
	pool postgres.DBPool
}

func New(
	pool postgres.DBPool,
) *PostgresRepo {
	return &PostgresRepo{
		pool: pool,
	}
}
