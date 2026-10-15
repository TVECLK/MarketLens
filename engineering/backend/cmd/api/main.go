package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"marketlens-go-backend/internal/api"
	"marketlens-go-backend/internal/config"
	"marketlens-go-backend/internal/ingestion"
	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/service"

	"github.com/gin-gonic/gin"
)


func ginMode() string {
	if m := os.Getenv("GIN_MODE"); m == gin.DebugMode || m == gin.TestMode {
		return m
	}
	return gin.ReleaseMode
}

func main() {
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	gin.SetMode(ginMode())

	store := repositories.NewStore(db)
	services := service.NewServices(store)

	llmClient := ingestion.NewDeepSeekClient(http.DefaultClient)
	industryClassifier := ingestion.NewIndustryClassifier(
		services.IndustrySectors,
		services.IndustryDivisions,
		services.IndustryGroups,
		services.IndustryClasses,
		services.IndustrySubclasses,
		llmClient,
	)
	occupationClassifier := ingestion.NewOccupationClassifier(
		services.MajorGroups,
		services.SubMajorGroups,
		services.MinorGroups,
		services.UnitGroups,
		services.OccupationGroups,
		llmClient,
	)
	metadataBuilder := ingestion.NewMetadataBuilder(
		services.Formalities,
		services.Genders,
		services.VocationalEducations,
		services.EmploymentSectors,
		services.EducationLevels,
		services.Experiences,
		30*time.Minute,
	)
	ingestionSvc := ingestion.NewIngestionService(
		services.JobPosts,
		llmClient,
		industryClassifier,
		occupationClassifier,
		metadataBuilder,
		0.65,
	)

	h := api.NewHandler(store, services, ingestionSvc)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	api.RegisterRoutes(r, h)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("API server failed: %v", err)
	}
}
