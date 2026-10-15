package service

import (
	"context"
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
)

// JobPostService fronts the Store's job-post ingestion/dedup methods.
type JobPostService struct {
	store *repositories.Store
}

func NewJobPostService(store *repositories.Store) *JobPostService {
	return &JobPostService{store: store}
}

func (s *JobPostService) GetJobsByBucketKeys(ctx context.Context, bucketKeys []string) ([]models.JobPost, error) {
	return s.store.GetJobsByBucketKeys(ctx, bucketKeys)
}

func (s *JobPostService) SaveOneJob(ctx context.Context, job *models.JobPost, lshIndexRecords []models.LshIndex) error {
	return s.store.SaveOneJob(ctx, job, lshIndexRecords)
}

func (s *JobPostService) UpdateDuplicateJob(ctx context.Context, jobPostID, crawlerRunID uint) error {
	return s.store.UpdateDuplicateJob(ctx, jobPostID, crawlerRunID)
}

func (s *JobPostService) ReconcileStaleVacancies(ctx context.Context, currentRunID uint) (int64, error) {
	return s.store.ReconcileStaleVacancies(ctx, currentRunID)
}
