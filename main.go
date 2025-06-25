package main

import (
	"log"
	"net/http"
	"os"
)

// @title           Token Exchange Service API
// @version         1.0
// @description     Service for exchanging Azure AD tokens using OAuth 2.0 On-Behalf-Of flow

// Configuration variables (read from environment)
var (
	clientID     = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	tenantID     = os.Getenv("TENANT_ID")
	defaultScope = os.Getenv("DEFAULT_SCOPE")
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Serve the documentation based on the path
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "static/index.html")
		} else if r.URL.Path == "/swagger.json" {
			w.Header().Set("Content-Type", "application/json")
			http.ServeFile(w, r, "docs/swagger.json")
		} else if r.URL.Path == "/swagger.yaml" {
			w.Header().Set("Content-Type", "application/yaml")
			http.ServeFile(w, r, "docs/swagger.yaml")
		} else {
			http.NotFound(w, r)
		}
	case http.MethodPost:
		if r.URL.Path == "/" {
			handleExchangeTokenRequest(w, r)
		} else if r.URL.Path == "/refresh" {
			handleRefreshTokenRequest(w, r)
		} else {
			http.NotFound(w, r)
		}
	default:
		jsonError(w, "method_not_allowed", "Only GET and POST methods are allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	if clientID == "" || clientSecret == "" || tenantID == "" {
		log.Fatal("FATAL: Environment variables CLIENT_ID, CLIENT_SECRET, and TENANT_ID must be set.")
	}

	// Only show the first 10 characters of clientSecret, then ellipsis
	clientSecretDisplay := ""
	if len(clientSecret) > 10 {
		clientSecretDisplay = clientSecret[:10] + "..."
	} else {
		clientSecretDisplay = clientSecret
	}

	log.Printf("INFO: Loaded configuration: clientID=%s, clientSecret=%s, tenantID=%s\n", clientID, clientSecretDisplay, tenantID)
	if defaultScope != "" {
		log.Printf("INFO: Default scope is configured: defaultScope=%s\n", defaultScope)
	}

	// Main API endpoint
	http.HandleFunc("/", handleRequest)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000" // Default port
	}

	log.Printf("INFO: Token Exchange Service starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("FATAL: Server failed to start: %v", err)
	}
}
