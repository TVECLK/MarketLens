package main

import (
	"log"

	"marketlens-go-backend/internal/api"
	"marketlens-go-backend/internal/auth"
	"marketlens-go-backend/internal/config"
	"marketlens-go-backend/internal/ingestion"
	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// resolveGinMode falls back to release mode unless raw is explicitly one
// of gin's debug/test modes.
func resolveGinMode(raw string) string {
	if raw == gin.DebugMode || raw == gin.TestMode {
		return raw
	}
	return gin.ReleaseMode
}

func main() {
	cfg := config.LoadConfig()

	db, err := config.ConnectDatabase(cfg.DB)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	gin.SetMode(resolveGinMode(cfg.Server.GinModeEnv))

	store := repositories.NewStore(db)
	services := service.NewServices(store)

	if err := cfg.Auth.Validate(); err != nil {
		log.Fatalf("auth misconfigured: %v", err)
	}
	verifier := auth.NewVerifier(cfg.Auth)

	if err := cfg.LLM.Validate(); err != nil {
		log.Fatalf("llm client misconfigured: %v", err)
	}
	ingestionSvc := ingestion.NewServiceFromConfig(services, cfg.LLM, cfg.Ingestion)

	h := api.NewHandler(store, services, ingestionSvc)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	api.RegisterRoutes(r, h, verifier)

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("API server failed: %v", err)
	}
}
