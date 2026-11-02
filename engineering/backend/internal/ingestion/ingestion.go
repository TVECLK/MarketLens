package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"marketlens-go-backend/internal/llm"
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/service"
)

// RawJobInput is one crawled job to run through ProcessBatch.
type RawJobInput struct {
	JobID        string
	Employer     string
	JobRole      string
	Location     string
	Description  string
	CrawlerRunID uint
	Source       string
}

// hasNoUsableText reports whether raw has no content in any text
// field — such a job produces a degenerate MinHash signature (every
// element stays at its max-value default), making it look identical
// to every other such job.
func hasNoUsableText(raw RawJobInput) bool {
	return strings.TrimSpace(raw.Employer) == "" &&
		strings.TrimSpace(raw.JobRole) == "" &&
		strings.TrimSpace(raw.Location) == "" &&
		strings.TrimSpace(raw.Description) == ""
}

// FailedJob reports one job's save failure, keyed by the crawler-generated
// job_id. A job not present in a ProcessBatch response's failures either
// saved successfully or was a duplicate that got its crawler_run_id updated.
type FailedJob struct {
	JobID      string
	StatusCode int
	Error      string
}

// IngestionService processes a batch of crawled jobs end to end: MinHash
// + duplicate check for every job, then a per-job save or duplicate
// update (each new job extracted + classified first).
type IngestionService struct {
	jobPosts             *service.JobPostService
	llmClient            *llm.Client
	model                string
	metadataBuilder      *MetadataBuilder
	industryClassifier   *IndustryClassifier
	occupationClassifier *OccupationClassifier

	numPerm          int
	numBands         int
	jaccardThreshold float64
}

// NewIngestionService wires dedup, classification, and metadata
// together. industryClassifier, occupationClassifier, and
// metadataBuilder are built by the caller (each needs its own set of
// services)
func NewIngestionService(
	jobPosts *service.JobPostService,
	llmClient *llm.Client,
	model string,
	industryClassifier *IndustryClassifier,
	occupationClassifier *OccupationClassifier,
	metadataBuilder *MetadataBuilder,
	jaccardThreshold float64,
) *IngestionService {
	const numPerm = 128
	const numBands = 8

	return &IngestionService{
		jobPosts:             jobPosts,
		llmClient:            llmClient,
		model:                model,
		metadataBuilder:      metadataBuilder,
		industryClassifier:   industryClassifier,
		occupationClassifier: occupationClassifier,
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
		if hasNoUsableText(raw) {
			failures = append(failures, FailedJob{
				JobID:      raw.JobID,
				StatusCode: http.StatusBadRequest,
				Error:      "job has no usable text content",
			})
			continue
		}

		jobData := JobData{
			Employer:    raw.Employer,
			JobRole:     raw.JobRole,
			Location:    raw.Location,
			Description: raw.Description,
		}
		sig, lshIndexes := GenerateMinHashAndLSH(jobData, s.numPerm, s.numBands)

		isDuplicate, matchedJobID, err := CheckDuplicate(ctx, s.jobPosts, lshIndexes, sig, raw.Location, s.jaccardThreshold)
		if err != nil {
			failures = append(failures, FailedJob{
				JobID:      raw.JobID,
				StatusCode: http.StatusServiceUnavailable,
				Error:      fmt.Sprintf("checking duplicate: %v", err),
			})
			continue
		}
		if isDuplicate {
			if err := s.jobPosts.UpdateDuplicateJob(ctx, uint(matchedJobID), raw.CrawlerRunID); err != nil {
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

		if err := s.jobPosts.SaveOneJob(ctx, &jobPost, lshRecords); err != nil {
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

// extractedJob is the shape the LLM is asked to return, matching Schema.
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

// extractAndClassify asks the LLM to fill in the fields not already
// known from the raw crawl, classifies industry + occupation, and
// assembles the final JobPost.
//
// employer/job_role/location/description/source come directly from the
// raw crawled input rather than being re-derived by the LLM.
func (s *IngestionService) extractAndClassify(ctx context.Context, raw RawJobInput, sig []uint32) (models.JobPost, error) {
	schema, instruction, options, err := s.metadataBuilder.Build(ctx)
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

	prompt := buildExtractionPrompt(instruction, schemaJSON, jobText)

	content, err := s.llmClient.Chat(ctx, prompt)
	if err != nil {
		return models.JobPost{}, fmt.Errorf("llm extraction call: %w", wrapLLMErr(err))
	}

	var extracted extractedJob
	if err := json.Unmarshal([]byte(stripJSONFences(content)), &extracted); err != nil {
		return models.JobPost{}, fmt.Errorf("%w: parsing llm extraction response: %v", ErrTransientExtractionFailure, err)
	}

	industrySubclassID, err := s.industryClassifier.Classify(ctx, jobText)
	if err != nil {
		return models.JobPost{}, fmt.Errorf("industry classification: %w", err)
	}
	occupationGroupID, err := s.occupationClassifier.Classify(ctx, jobText)
	if err != nil {
		return models.JobPost{}, fmt.Errorf("occupation classification: %w", err)
	}

	postedAt, err := parsePostedAt(extracted.MetaData.PostedAt)
	if err != nil {
		return models.JobPost{}, fmt.Errorf("invalid posted_at date: %w", err)
	}

	skills := make([]models.Skill, 0, len(extracted.Skills))
	for _, sk := range extracted.Skills {
		if sk.Skill != "" {
			skills = append(skills, models.Skill{Skill: sk.Skill})
		}
	}

	formalityID := validateID("formality_id", extracted.MetaData.FormalityID, options.validFormalities)
	genderID := validateID("gender_id", extracted.MetaData.GenderID, options.validGenders)
	vocationalEducationID := validateID("vocational_education_id", extracted.MetaData.VocationalEducationID, options.validVocationalEducations)
	employmentSectorID := validateID("employment_sector_id", extracted.MetaData.EmploymentSectorID, options.validEmploymentSectors)
	educationLevelID := validateID("education_level_id", extracted.MetaData.EducationLevelID, options.validEducationLevels)
	experienceID := validateID("experience_id", extracted.MetaData.ExperienceID, options.validExperiences)
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
			AiVersion:             &models.AiVersion{Version: s.model},
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

func parsePostedAt(raw string) (time.Time, error) {
	if raw == "" {
		return time.Now().UTC(), nil
	}
	return time.Parse(time.RFC3339, raw)
}

func validateID(field string, id uint, valid map[uint]struct{}) uint {
	if id == 0 {
		return 0
	}
	if _, ok := valid[id]; !ok {
		log.Printf("llm returned %s=%d, which was not among the offered options — treating as not set", field, id)
		return 0
	}
	return id
}

func nilIfZero(id uint) *uint {
	if id == 0 {
		return nil
	}
	return &id
}