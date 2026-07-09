package service

import (
	"github.com/andyle182810/gframework/valkey"
	"github.com/go_starter_kit/internal/config"
	"github.com/go_starter_kit/internal/repo"
	"resty.dev/v3"
)

type Service struct {
	cfg          *config.Config
	repo         *repo.PostgresRepo
	restyClient  *resty.Client
	valkeyClient *valkey.Valkey
}

func New(
	cfg *config.Config,
	repo *repo.PostgresRepo,
	restyClient *resty.Client,
	valkeyClient *valkey.Valkey,
) *Service {
	return &Service{
		cfg:          cfg,
		repo:         repo,
		restyClient:  restyClient,
		valkeyClient: valkeyClient,
	}
}
