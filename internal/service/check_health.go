package service

import (
	"github.com/andyle182810/gframework/httpserver"
	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
)

type CheckHealthRequest struct{}

type CheckHealthResponse struct {
	Status string `json:"status"`
}

// CheckHealth godoc
//
//	@Summary		Check system health
//	@Description	Check the health status of the service
//	@Tags			api
//	@Accept			json
//	@Produce		json
//	@Param			X-Request-Id	header		string										false	"Request ID (UUID)"
//	@Success		200				{object}	httpserver.APIResponse[CheckHealthResponse]	"Return data"
//	@Failure		400				{object}	echo.HTTPError								"Error object"
//	@Router			/health [get]
func (s *Service) CheckHealth(ctx *echo.Context, req *CheckHealthRequest) (any, *echo.HTTPError) {
	delegator := func(
		log zerolog.Logger,
		ctx *echo.Context,
		req *CheckHealthRequest,
	) (*httpserver.HandlerResponse[CheckHealthResponse], *echo.HTTPError) {
		handle := NewCheckHealthHandle(log)

		return handle.Handle(ctx, req)
	}

	return httpserver.ExecuteStandardized(ctx, req, "CheckHealth", delegator)
}

type CheckHealthHandle struct {
	log zerolog.Logger
}

func NewCheckHealthHandle(
	log zerolog.Logger,
) *CheckHealthHandle {
	return &CheckHealthHandle{
		log: log,
	}
}

func (h *CheckHealthHandle) Handle(
	ctx *echo.Context,
	_ *CheckHealthRequest,
) (*httpserver.HandlerResponse[CheckHealthResponse], *echo.HTTPError) {
	return &httpserver.HandlerResponse[CheckHealthResponse]{
		Data: CheckHealthResponse{
			Status: "healthy",
		},
		Pagination: nil,
	}, nil
}
