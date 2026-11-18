package repositories

import (
	"context"
	"marketlens-go-backend/internal/models"
	"time"

	"gorm.io/gorm"
)

func (r *Store) CreateCrawlerRun(ctx context.Context) (models.CrawlerRun, error) {
	now := time.Now()
	run := models.CrawlerRun{
		StartedAt: &now,
		Status:    "RUNNING",
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		err := tx.Model(&models.CrawlerRun{}).
			Where("status = ?", "RUNNING").
			Updates(map[string]interface{}{
				"status":      "FAILED",
				"finished_at": &now,
			}).Error
		if err != nil {
			return err
		}

		if err := tx.Create(&run).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return models.CrawlerRun{}, err
	}

	return run, nil
}

// GetCrawlerRunStatus returns the status of the crawler run with this id,
// or gorm.ErrRecordNotFound if no such run is recorded.
func (r *Store) GetCrawlerRunStatus(ctx context.Context, id uint) (string, error) {
	var run models.CrawlerRun
	if err := r.db.WithContext(ctx).Select("status").First(&run, id).Error; err != nil {
		return "", err
	}
	return run.Status, nil
}

func (r *Store) CompleteCrawlerRun(ctx context.Context, id uint, status string) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&models.CrawlerRun{}).Where("id = ?", id).Updates(map[string]interface{}{
		"finished_at": &now,
		"status":      status, // 'COMPLETED' or 'FAILED'
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
