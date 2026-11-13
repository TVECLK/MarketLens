package ingestion

import (
	"net/http"
	"time"

	"marketlens-go-backend/internal/llm"
	"marketlens-go-backend/internal/service"
)

// llmHTTPTimeout bounds how long a single LLM call can take — generous
// enough for normal completion latency, finite enough that a hung
// upstream response can't hold a request open indefinitely.
const llmHTTPTimeout = 60 * time.Second

// NewServiceFromConfig builds the full IngestionService (LLM client,
// classifiers, metadata builder) from the shared services. It is the one
// place this wiring lives, so every binary that ingests jobs (cmd/api,
// cmd/mcp) gets an identical pipeline. The caller validates llmCfg first.
func NewServiceFromConfig(svcs *service.Services, llmCfg llm.Config, pipeline PipelineConfig) *IngestionService {
	llmClient := llm.NewClient(llmCfg, &http.Client{Timeout: llmHTTPTimeout})

	industryClassifier := NewIndustryClassifier(
		svcs.IndustrySectors,
		svcs.IndustryDivisions,
		svcs.IndustryGroups,
		svcs.IndustryClasses,
		svcs.IndustrySubclasses,
		llmClient,
	)
	occupationClassifier := NewOccupationClassifier(
		svcs.MajorGroups,
		svcs.SubMajorGroups,
		svcs.MinorGroups,
		svcs.UnitGroups,
		svcs.OccupationGroups,
		llmClient,
	)
	metadataBuilder := NewMetadataBuilder(
		svcs.Formalities,
		svcs.Genders,
		svcs.VocationalEducations,
		svcs.EmploymentSectors,
		svcs.EducationLevels,
		svcs.Experiences,
		pipeline.MetadataCacheTTL,
	)

	return NewIngestionService(
		svcs.JobPosts,
		svcs.Sources,
		llmClient,
		llmCfg.Model,
		industryClassifier,
		occupationClassifier,
		metadataBuilder,
		pipeline.DedupThreshold,
	)
}
