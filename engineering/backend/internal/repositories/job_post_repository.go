package repositories

import (
	"context"
	"errors"
	"fmt"
	"marketlens-go-backend/internal/models"
	"time"

	"gorm.io/gorm"
)

func (r *Store) GetJobsByBucketKeys(ctx context.Context, bucketKeys []string) ([]models.JobPost, error) {
	var jobs []models.JobPost

	err := r.db.WithContext(ctx).Distinct("job_post.*").
		Joins("JOIN lsh_index ON lsh_index.job_post_id = job_post.id").
		Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
		Where("lsh_index.bucket_key IN ? AND meta_data.end_date IS NULL", bucketKeys).
		Preload("MetaData").
		Find(&jobs).Error

	if err != nil {
		return nil, err
	}
	return jobs, nil
}

// SaveOneJob persists a single new job — employer/job_type/skills
// FirstOrCreate lookups, a geo lookup, source/ai_version FirstOrCreate,
// the job itself, then its LSH records keyed off the job's real DB id —
// in its own transaction, isolated from every other job in the batch.
func (r *Store) SaveOneJob(ctx context.Context, job *models.JobPost, lshIndexRecords []models.LshIndex) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Employer (FirstOrCreate)
		if job.Employer != nil && job.Employer.Name != "" {
			var employer models.Employer
			if err := tx.Where(models.Employer{Name: job.Employer.Name}).
				FirstOrCreate(&employer).Error; err != nil {
				return wrapSaveErr(err, "employer lookup failed")
			}
			job.EmployerID = &employer.ID
			job.Employer = nil
		}

		// JobType (FirstOrCreate)
		if job.JobType != nil && job.JobType.Type != "" {
			var jt models.JobType
			if err := tx.Where(models.JobType{Type: job.JobType.Type}).
				FirstOrCreate(&jt).Error; err != nil {
				return wrapSaveErr(err, "job_type lookup failed")
			}
			job.JobTypeID = &jt.ID
			job.JobType = nil
		}

		// Skills (FirstOrCreate per skill)
		var linkedSkills []models.Skill
		for _, s := range job.Skills {
			if s.Skill == "" {
				continue
			}
			var skill models.Skill
			if err := tx.Where(models.Skill{Skill: s.Skill}).
				FirstOrCreate(&skill).Error; err != nil {
				return wrapSaveErr(err, fmt.Sprintf("skill '%s' lookup failed", s.Skill))
			}
			linkedSkills = append(linkedSkills, skill)
		}
		job.Skills = linkedSkills

		// Geo (lookup only, no create)
		if job.MetaData.GeoData != nil && job.MetaData.GeoData.Province != "" {
			var geo models.GeoData
			err := tx.Where("province = ?", job.MetaData.GeoData.Province).
				First(&geo).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: province '%s' not registered in geo_data",
					ErrPermanentSaveFailure, job.MetaData.GeoData.Province)
			} else if err != nil {
				return wrapSaveErr(err, "geo lookup failed")
			}
			job.MetaData.GeoDataID = &geo.ID
			job.MetaData.GeoData = nil
		}

		// Source (FirstOrCreate)
		if job.MetaData.Source != nil && job.MetaData.Source.Name != "" {
			var source models.Source
			if err := tx.Where(models.Source{Name: job.MetaData.Source.Name}).
				FirstOrCreate(&source).Error; err != nil {
				return wrapSaveErr(err, "source lookup failed")
			}
			job.MetaData.SourceID = &source.ID
			job.MetaData.Source = nil
		}

		// AiVersion (FirstOrCreate)
		if job.MetaData.AiVersion != nil && job.MetaData.AiVersion.Version != "" {
			var av models.AiVersion
			if err := tx.Where(models.AiVersion{Version: job.MetaData.AiVersion.Version}).
				FirstOrCreate(&av).Error; err != nil {
				return wrapSaveErr(err, "ai_version lookup failed")
			}
			job.MetaData.AiVersionID = &av.ID
			job.MetaData.AiVersion = nil
		}

		// Save JobPost (cascades MetaData + join table)
		if err := tx.Create(job).Error; err != nil {
			return wrapSaveErr(err, "job insert failed")
		}

		for i := range lshIndexRecords {
			lshIndexRecords[i].JobPostID = job.ID
		}
		if len(lshIndexRecords) > 0 {
			if err := tx.Omit("JobPost").Create(&lshIndexRecords).Error; err != nil {
				return wrapSaveErr(err, "lsh_index insert failed")
			}
		}

		return nil
	})
}

// UpdateDuplicateJob records that an already-saved job was seen again in
// crawlerRunID, in its own transaction.
func (r *Store) UpdateDuplicateJob(ctx context.Context, jobPostID uint, crawlerRunID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Table("meta_data").
			Where("job_post_id = ?", jobPostID).
			Update("crawler_run_id", crawlerRunID).Error
	})
}

func (r *Store) ReconcileStaleVacancies(ctx context.Context, currentRunID uint) (int64, error) {
	var affectedRows int64
	now := time.Now()

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		var staleJobIDs []uint
		err := tx.Model(&models.JobMetaData{}).
			Where("crawler_run_id <> ? AND end_date IS NULL", currentRunID).
			Pluck("job_post_id", &staleJobIDs).Error
		if err != nil {
			return err
		}

		if len(staleJobIDs) == 0 {
			return nil
		}

		if err := tx.Where("job_post_id IN ?", staleJobIDs).Delete(&models.LshIndex{}).Error; err != nil {
			return err
		}

		result := tx.Model(&models.JobMetaData{}).
			Where("job_post_id IN ?", staleJobIDs).
			Updates(map[string]interface{}{
				"end_date": &now,
			})
		if result.Error != nil {
			return result.Error
		}

		affectedRows = result.RowsAffected
		return nil
	})

	if err != nil {
		return 0, err
	}

	return affectedRows, nil
}
