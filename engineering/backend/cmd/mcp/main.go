package main

import (
	"log"

	"marketlens-go-backend/internal/auth"
	"marketlens-go-backend/internal/config"
	"marketlens-go-backend/internal/mcpserver"
	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/service"
)

// This is its own binary/process (separate from cmd/api), so an MCP
// startup failure no longer takes the customer-facing REST API down with
// it — it just exits this process.
func main() {
	cfg := config.LoadConfig()

	db, err := config.ConnectDatabase(cfg.DB)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	store := repositories.NewStore(db)
	services := service.NewServices(store)

	mcpAuthCfg := auth.Config{
		BaseURL:     cfg.Auth.BaseURL,
		Issuer:      cfg.Auth.Issuer,
		Audience:    cfg.MCP.ResourceID,
		InsecureTLS: cfg.Auth.InsecureTLS,
	}
	if err := mcpAuthCfg.Validate(); err != nil {
		log.Fatalf("mcp auth misconfigured: %v", err)
	}
	mcpVerifier := auth.NewVerifier(mcpAuthCfg)

	mcpServer := mcpserver.New(services)

	meta := mcpserver.Metadata{
		ResourceID:    cfg.MCP.ResourceID,
		IssuerURL:     cfg.Auth.Issuer,
		PublicBaseURL: cfg.MCP.PublicBaseURL,
	}

	log.Println("MCP server listening on :9090/mcp")
	if err := mcpserver.StartHTTP(mcpServer, mcpVerifier, meta, ":9090"); err != nil {
		log.Fatalf("MCP server failed: %v", err)
	}
}
