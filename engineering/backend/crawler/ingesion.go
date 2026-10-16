package crawler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"marketlens-go-backend/models"
	"marketlens-go-backend/repositories"
)

// RawJobInput is one crawled job as it arrives in a batch request:
// {"job_id":, "employer":, "job_role":, "location":, "description":, "crawler_run_id":, "source":}
type RawJobInput struct {
	JobID        string `json:"job_id"`
	Employer     string `json:"employer"`
	JobRole      string `json:"job_role"`
	Location     string `json:"location"`
	Description  string `json:"description"`
	CrawlerRunID uint   `json:"crawler_run_id"`
	Source       string `json:"source"`
}

// FailedJob reports one job's save failure, keyed by the crawler-generated
// job_id. A job not present in a ProcessBatch response's failures either
// saved successfully or was a duplicate that got its crawler_run_id updated.
type FailedJob struct {
	JobID      string `json:"job_id"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
}

// IngestionService processes a batch of crawled jobs end to end: MinHash
// + duplicate check for every job, then a per-job save or duplicate
// update (each new job extracted + classified first).
type IngestionService struct {
	repo                 Repository
	llm                  *DeepSeekClient
	metadataBuilder      *MetadataBuilder
	industryClassifier   *IndustryClassifier
	occupationClassifier *OccupationClassifier

	numPerm          int
	numBands         int
	jaccardThreshold float64
}

// NewIngestionService wires dedup, classification, and metadata
// together with the repository.
func NewIngestionService(
	repo Repository,
	llm *DeepSeekClient,
	metadataBuilder *MetadataBuilder,
	jaccardThreshold float64,
) *IngestionService {
	const numPerm = 128
	const numBands = 8

	return &IngestionService{
		repo:                 repo,
		llm:                  llm,
		metadataBuilder:      metadataBuilder,
		industryClassifier:   NewIndustryClassifier(repo, llm),
		occupationClassifier: NewOccupationClassifier(repo, llm),
		numPerm:              numPerm,
		numBands:             numBands,
		jaccardThreshold:     jaccardThreshold,
	}
}

// ProcessBatch runs the full pipeline for one crawl batch. Every job is
// attempted, sequentially and in its own DB transaction, regardless of
// whether earlier jobs in the batch failed. Only failures are returned;
// a job not present in the result saved successfully or was a duplicate
// that got its crawler_run_id updated.
func (s *IngestionService) ProcessBatch(ctx context.Context, rawJobs []RawJobInput) []FailedJob {
	var failures []FailedJob

	for _, raw := range rawJobs {
		jobData := JobData{
			Employer:    raw.Employer,
			JobRole:     raw.JobRole,
			Location:    raw.Location,
			Description: raw.Description,
		}
		sig, lshIndexes := GenerateMinHashAndLSH(jobData, s.numPerm, s.numBands)

		isDuplicate, matchedJobID := CheckDuplicate(s.repo, lshIndexes, sig, raw.Location, s.jaccardThreshold)
		if isDuplicate {
			if err := s.repo.UpdateDuplicateJob(uint(matchedJobID), raw.CrawlerRunID); err != nil {
				failures = append(failures, FailedJob{
					JobID:      raw.JobID,
					StatusCode: http.StatusServiceUnavailable,
					Error:      fmt.Sprintf("updating duplicate job: %v", err),
				})
			}
			continue
		}

		jobPost, err := s.extractAndClassify(ctx, raw, sig)
		if err != nil {
			statusCode := http.StatusUnprocessableEntity
			if errors.Is(err, ErrTransientExtractionFailure) {
				statusCode = http.StatusServiceUnavailable
			}
			failures = append(failures, FailedJob{
				JobID:      raw.JobID,
				StatusCode: statusCode,
				Error:      fmt.Sprintf("job %q at %q: %v", raw.JobRole, raw.Employer, err),
			})
			continue
		}

		lshRecords := make([]models.LshIndex, 0, len(lshIndexes))
		for _, idx := range lshIndexes {
			lshRecords = append(lshRecords, models.LshIndex{
				BandNo:    idx.BandNo,
				BucketKey: idx.BucketKey,
			})
		}

		if err := s.repo.SaveOneJob(&jobPost, lshRecords); err != nil {
			statusCode := http.StatusServiceUnavailable
			if errors.Is(err, repositories.ErrPermanentSaveFailure) {
				statusCode = http.StatusUnprocessableEntity
			}
			failures = append(failures, FailedJob{
				JobID:      raw.JobID,
				StatusCode: statusCode,
				Error:      err.Error(),
			})
		}
	}

	return failures
}

// extractedJob is the shape DeepSeek is asked to return, matching Schema.
type extractedJob struct {
	JobType struct {
		Type string `json:"type"`
	} `json:"job_type"`
	WorkMode      string `json:"work_mode"`
	NoOfVacancies int    `json:"no_of_vacancies"`
	MetaData      struct {
		GeoData struct {
			Province string `json:"province"`
		} `json:"geo_data"`
		PostedAt              string  `json:"posted_at"`
		ConfidenceScore       float64 `json:"confidence_score"`
		FormalityID           uint    `json:"formality_id"`
		GenderID              uint    `json:"gender_id"`
		VocationalEducationID uint    `json:"vocational_education_id"`
		EmploymentSectorID    uint    `json:"employment_sector_id"`
		EducationLevelID      uint    `json:"education_level_id"`
		ExperienceID          uint    `json:"experience_id"`
	} `json:"meta_data"`
	Skills []struct {
		Skill string `json:"skill"`
	} `json:"skills"`
}

// extractAndClassify asks DeepSeek to fill in the fields not already
// known from the raw crawl, classifies industry + occupation, and
// assembles the final JobPost.
//
// employer/job_role/location/description/source come directly from the
// raw crawled input rather than being re-derived by the LLM.
func (s *IngestionService) extractAndClassify(ctx context.Context, raw RawJobInput, sig []uint32) (models.JobPost, error) {
	schema, instruction, err := s.metadataBuilder.Build()
	if err != nil {
		return models.JobPost{}, fmt.Errorf("building extraction schema: %w", err)
	}

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return models.JobPost{}, fmt.Errorf("marshaling extraction schema: %w", err)
	}

	jobText := fmt.Sprintf(
		"Employer: %s\nJob Role: %s\nLocation: %s\nSource: %s\n\nDescription:\n%s",
		raw.Employer, raw.JobRole, raw.Location, raw.Source, raw.Description,
	)

	prompt := instruction +
		"\n\nThe JSON you return MUST conform to this JSON Schema:\n" + string(schemaJSON) +
		"\n\nJob posting to extract from:\n" + jobText

	content, err := s.llm.Chat(ctx, prompt)
	if err != nil {
		return models.JobPost{}, fmt.Errorf("deepseek extraction call: %w", err)
	}

	var extracted extractedJob
	if err := json.Unmarshal([]byte(stripJSONFences(content)), &extracted); err != nil {
		return models.JobPost{}, fmt.Errorf("%w: parsing deepseek extraction response: %v", ErrTransientExtractionFailure, err)
	}

	industrySubclassID, err := s.industryClassifier.Classify(ctx, jobText)
	if err != nil {
		return models.JobPost{}, fmt.Errorf("industry classification: %w", err)
	}
	occupationGroupID, err := s.occupationClassifier.Classify(ctx, jobText)
	if err != nil {
		return models.JobPost{}, fmt.Errorf("occupation classification: %w", err)
	}

	var postedAt *time.Time
	if t, err := time.Parse(time.RFC3339, extracted.MetaData.PostedAt); err == nil {
		postedAt = &t
	}

	skills := make([]models.Skill, 0, len(extracted.Skills))
	for _, sk := range extracted.Skills {
		if sk.Skill != "" {
			skills = append(skills, models.Skill{Skill: sk.Skill})
		}
	}

	formalityID := extracted.MetaData.FormalityID
	genderID := extracted.MetaData.GenderID
	vocationalEducationID := extracted.MetaData.VocationalEducationID
	employmentSectorID := extracted.MetaData.EmploymentSectorID
	educationLevelID := extracted.MetaData.EducationLevelID
	experienceID := extracted.MetaData.ExperienceID
	crawlerRunID := raw.CrawlerRunID

	workMode := extracted.WorkMode
	if workMode == "" {
		workMode = string(models.WorkModeOnsite) 
	}

	return models.JobPost{
		Employer:       &models.Employer{Name: raw.Employer},
		JobType:        &models.JobType{Type: extracted.JobType.Type},
		JobRole:        raw.JobRole,
		WorkMode:       models.WorkMode(workMode),
		JobDescription: raw.Description,
		Location:       raw.Location,
		NoOfVacancies:  extracted.NoOfVacancies,
		Skills:         skills,
		MetaData: models.JobMetaData{
			CrawlerRunID:          &crawlerRunID,
			GeoData:               &models.GeoData{Province: extracted.MetaData.GeoData.Province},
			Source:                &models.Source{Name: raw.Source},
			AiVersion:             &models.AiVersion{Version: "deepseek-v1"},
			FormalityID:           nilIfZero(formalityID),
			GenderID:              nilIfZero(genderID),
			VocationalEducationID: nilIfZero(vocationalEducationID),
			EmploymentSectorID:    nilIfZero(employmentSectorID),
			EducationLevelID:      nilIfZero(educationLevelID),
			ExperienceID:          nilIfZero(experienceID),
			IndustrySubclassID:    nilIfZero(industrySubclassID),
			OccupationGroupID:     nilIfZero(occupationGroupID),
			PostedAt:              postedAt,
			ConfidenceScore:       extracted.MetaData.ConfidenceScore,
			MinhashSignature:      ToPQArray(sig),
		},
	}, nil
}

func nilIfZero(id uint) *uint {
	if id == 0 {
		return nil
	}
	return &id
}