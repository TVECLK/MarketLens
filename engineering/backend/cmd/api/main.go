package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"marketlens-go-backend/internal/api"
	"marketlens-go-backend/internal/auth"
	"marketlens-go-backend/internal/config"
	"marketlens-go-backend/internal/httpserver"
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

	// r.Run starts a bare http.Server: no timeout on slow clients (Slowloris),
	// and nothing handles SIGTERM, so a pod stop cuts requests off halfway.
	// Serve through the shared server instead. SIGTERM (a Kubernetes pod stop)
	// or Ctrl-C stops it accepting connections and gives in-flight requests
	// httpserver.ShutdownGracePeriod to finish before the process exits.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ln, err := net.Listen("tcp", ":"+cfg.Server.Port)
	if err != nil {
		log.Fatalf("API server failed to listen on :%s: %v", cfg.Server.Port, err)
	}
	log.Printf("API server listening on :%s", cfg.Server.Port)
	if err := httpserver.ServeUntilDone(ctx, httpserver.NewServer(r.Handler()), ln, httpserver.ShutdownGracePeriod, nil); err != nil {
		log.Fatalf("API server failed: %v", err)
	}
	log.Println("API server stopped")
}
