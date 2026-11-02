package service

import (
	"context"
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
)

// CrawlerRunService fronts the Store's crawler-run tracking methods.
type CrawlerRunService struct {
	store *repositories.Store
}

func NewCrawlerRunService(store *repositories.Store) *CrawlerRunService {
	return &CrawlerRunService{store: store}
}

func (s *CrawlerRunService) CreateCrawlerRun(ctx context.Context) (models.CrawlerRun, error) {
	return s.store.CreateCrawlerRun(ctx)
}

// validCrawlerRunStatuses are the only values a caller may set via
// CompleteCrawlerRun — anything else (a typo, a stray value) is rejected
// instead of being written straight into the crawler_runs.status column,
// where other code relies on it only ever being RUNNING/COMPLETED/FAILED.
var validCrawlerRunStatuses = map[string]bool{
	"COMPLETED": true,
	"FAILED":    true,
}

func (s *CrawlerRunService) CompleteCrawlerRun(ctx context.Context, id uint, status string) error {
	if !validCrawlerRunStatuses[status] {
		return ValidationErrorf("status must be one of COMPLETED, FAILED")
	}
	return s.store.CompleteCrawlerRun(ctx, id, status)
}
