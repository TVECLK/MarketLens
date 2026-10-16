package crawler

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"marketlens-go-backend/models"
)

// Schema is the static JSON schema for LLM extraction. It never depends
// on DB data, so it's a package-level value built once rather than
// rebuilt on every call.
var Schema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"employer": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
			"required": []string{"name"},
		},
		"job_role": map[string]any{"type": "string"},
		"job_type": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"type": map[string]any{"type": "string"},
			},
			"required": []string{"type"},
		},
		"job_description": map[string]any{"type": "string"},
		"location":        map[string]any{"type": "string"},
		"work_mode": map[string]any{
			"type": "string",
			"enum": []string{"remote", "onsite", "hybrid"},
		},
		"no_of_vacancies": map[string]any{"type": "integer"},
		"meta_data": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"geo_data": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"province": map[string]any{"type": "string"},
					},
					"required": []string{"province"},
				},
				"source": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"source": map[string]any{"type": "string"},
					},
					"required": []string{"source"},
				},
				"ai_version": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"version": map[string]any{"type": "string"},
					},
					"required": []string{"version"},
				},
				"posted_at": map[string]any{
					"type":    "string",
					"pattern": `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`,
				},
				"confidence_score":        map[string]any{"type": "number"},
				"formality_id":            map[string]any{"type": "integer"},
				"gender_id":               map[string]any{"type": "integer"},
				"vocational_education_id": map[string]any{"type": "integer"},
				"employment_sector_id":    map[string]any{"type": "integer"},
				"education_level_id":      map[string]any{"type": "integer"},
				"experience_id":           map[string]any{"type": "integer"},
			},
			"required": []string{
				"geo_data", "source", "ai_version", "posted_at", "confidence_score",
				"formality_id", "gender_id", "vocational_education_id",
				"employment_sector_id", "education_level_id", "experience_id",
			},
		},
		"skills": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"skill": map[string]any{"type": "string"},
				},
				"required": []string{"skill"},
			},
		},
	},
	"required": []string{
		"employer", "job_role", "job_type", "job_description",
		"location", "work_mode", "no_of_vacancies", "meta_data", "skills",
	},
}

const instructionTemplate = `Extract structural data from the raw job text into the specified JSON format.
CRITICAL: If a field is missing, use the strict default provided. Do not hallucinate.

Fields & Defaults:

1. 'employer': Object with company name.
Format: {"name": "Company Name"}
Default: {"name": ""}

2. 'job_role': Vacancy title string.
Default: ""

3. 'job_type': Object with contract type.
Format: {"type": "Full Time"} or {"type": "Part Time"} or {"type": "Contract"} or {"type": "Internship"}
Default: {"type": "Full Time"}

4. 'job_description': Combined summary of responsibilities, qualifications, and any offered benefits. Single string.
Default: ""

5. 'location': City/region string (e.g. "Colombo 03, Sri Lanka").
Default: "Sri Lanka"

6. 'no_of_vacancies': Number of open positions for this role, if stated.
Default: 1

7. 'work_mode': The job's work arrangement.
Must be one of: "remote", "onsite", "hybrid"
- "remote": explicit remote/work-from-home wording exists, with no on-site requirement
- "hybrid": explicit mention of a mix of remote and in-office work
- "onsite": no remote/hybrid wording, or the role explicitly requires in-office presence
Default: "onsite"

8. 'meta_data': Object containing all metadata fields:

- 'geo_data': Infer the Sri Lankan province from the location text.
    Format: {"province": "<province>"}
    Must match EXACTLY ONE from this list:
    ["Western", "Central", "Northern", "Eastern", "North Western",
        "North Central", "Uva", "Southern", "Sabaragamuwa"]
    Default: {"province": "Western"}

- 'source': The website or platform this job was crawled from.
    Format: {"source": "<source name>"}
    Example: {"source": "Ikman"} or {"source": "TopJobs"}
    Default: {"source": "Unknown"}

- 'ai_version': The AI model version used for extraction.
    Format: {"version": "deepseek-v1"}
    Always use: {"version": "deepseek-v1"}

- 'posted_at': ISO 8601 timestamp of when the job was posted, if detectable from the page.
    Format: "YYYY-MM-DDTHH:MM:SSZ"
    Default: current UTC timestamp.

- 'confidence_score': Your confidence in the extraction accuracy.
    Float between 0.00 and 1.00. Use 1.00 if all fields are clearly present.
    Default: 0.80

- 'formality_id': Must be one of these ids:
%s

- 'gender_id': Must be one of these ids:
%s

- 'vocational_education_id': Must be one of these ids:
%s

- 'employment_sector_id': Must be one of these ids:
%s

- 'education_level_id': Must be one of these ids:
%s

- 'experience_id': Must be one of these ids:
%s

9. 'skills': Array of skill objects extracted from the job description.
Format: [{"skill": "Skill Name"}, ...]
Extract only concrete technical or professional skills mentioned.
Default: []
`

// MetadataBuilder fetches the six reference lists (formality, gender,
// vocational education, employment sector, education level, experience)
// and assembles the extraction schema + instruction text.
//
// Reference lists are cached for ttl and only refetched once that
// expires, since this data changes rarely.
type MetadataBuilder struct {
	repo MetadataRepository
	ttl  time.Duration

	mu       sync.RWMutex
	cached   *optionSet
	cachedAt time.Time
}

type optionSet struct {
	formalities          string
	genders              string
	vocationalEducations string
	employmentSectors    string
	educationLevels      string
	experiences          string
}

// NewMetadataBuilder creates a schema builder that refreshes its cached
// reference data at most once per ttl.
func NewMetadataBuilder(repo MetadataRepository, ttl time.Duration) *MetadataBuilder {
	return &MetadataBuilder{repo: repo, ttl: ttl}
}

// Build returns the static extraction schema and the instruction text
// with the current reference option lists filled in.
func (b *MetadataBuilder) Build() (map[string]any, string, error) {
	options, err := b.getOptions()
	if err != nil {
		return nil, "", err
	}

	instruction := fmt.Sprintf(
		instructionTemplate,
		options.formalities,
		options.genders,
		options.vocationalEducations,
		options.employmentSectors,
		options.educationLevels,
		options.experiences,
	)

	return Schema, instruction, nil
}

func (b *MetadataBuilder) getOptions() (*optionSet, error) {
	b.mu.RLock()
	if b.cached != nil && time.Since(b.cachedAt) < b.ttl {
		cached := b.cached
		b.mu.RUnlock()
		return cached, nil
	}
	b.mu.RUnlock()

	return b.refresh()
}

func (b *MetadataBuilder) refresh() (*optionSet, error) {
	formalities, err := b.repo.GetAllFormalities()
	if err != nil {
		return nil, fmt.Errorf("%w: fetching formalities: %v", ErrTransientExtractionFailure, err)
	}
	genders, err := b.repo.GetAllGenders()
	if err != nil {
		return nil, fmt.Errorf("%w: fetching genders: %v", ErrTransientExtractionFailure, err)
	}
	vocationalEducations, err := b.repo.GetAllVocationalEducations()
	if err != nil {
		return nil, fmt.Errorf("%w: fetching vocational educations: %v", ErrTransientExtractionFailure, err)
	}
	employmentSectors, err := b.repo.GetAllEmploymentSectors()
	if err != nil {
		return nil, fmt.Errorf("%w: fetching employment sectors: %v", ErrTransientExtractionFailure, err)
	}
	educationLevels, err := b.repo.GetAllEducationLevels()
	if err != nil {
		return nil, fmt.Errorf("%w: fetching education levels: %v", ErrTransientExtractionFailure, err)
	}
	experiences, err := b.repo.GetAllExperiences()
	if err != nil {
		return nil, fmt.Errorf("%w: fetching experiences: %v", ErrTransientExtractionFailure, err)
	}

	options := &optionSet{
		formalities: formatOptions(formalities, func(f models.Formality) (uint, string) {
			return f.ID, f.FormalityType
		}),
		genders: formatOptions(genders, func(g models.Gender) (uint, string) {
			return g.ID, g.GenderType
		}),
		vocationalEducations: formatOptions(vocationalEducations, func(v models.VocationalEducation) (uint, string) {
			return v.ID, v.Level
		}),
		employmentSectors: formatOptions(employmentSectors, func(e models.EmploymentSector) (uint, string) {
			return e.ID, e.Sector
		}),
		educationLevels: formatOptions(educationLevels, func(e models.EducationLevel) (uint, string) {
			return e.ID, e.Level
		}),
		experiences: formatOptions(experiences, func(e models.Experience) (uint, string) {
			return e.ID, e.Name
		}),
	}

	b.mu.Lock()
	b.cached = options
	b.cachedAt = time.Now()
	b.mu.Unlock()

	return options, nil
}

// formatOptions converts a list of items into "id: label" lines, one per
// line, decoupled from any specific model's field name via extract.
func formatOptions[T any](items []T, extract func(T) (uint, string)) string {
	lines := make([]string, len(items))
	for i, item := range items {
		id, label := extract(item)
		lines[i] = fmt.Sprintf("%d: %s", id, label)
	}
	return strings.Join(lines, "\n")
}