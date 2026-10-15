package api

import (
	"context"

	"marketlens-go-backend/internal/ingestion"
	"marketlens-go-backend/internal/service"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	pinger    Pinger
	ingestion *ingestion.IngestionService
	*service.Services
}

func NewHandler(pinger Pinger, services *service.Services, ingestionSvc *ingestion.IngestionService) *Handler {
	return &Handler{pinger: pinger, ingestion: ingestionSvc, Services: services}
}
