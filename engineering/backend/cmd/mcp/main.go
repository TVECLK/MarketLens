package main

import (
	"log"

	"marketlens-go-backend/internal/config"
	"marketlens-go-backend/internal/mcpserver"
	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/service"
)

// This is its own binary/process (separate from cmd/api), so an MCP
// startup failure no longer takes the customer-facing REST API down with
// it — it just exits this process.
func main() {
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	store := repositories.NewStore(db)
	services := service.NewServices(store)

	mcpServer := mcpserver.New(services)

	log.Println("MCP server listening on :9090/mcp")
	if err := mcpserver.StartHTTP(mcpServer, ":9090"); err != nil {
		log.Fatalf("MCP server failed: %v", err)
	}
}
