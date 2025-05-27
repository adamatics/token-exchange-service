package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// @title           Token Exchange Service API
// @version         1.0
// @description     Service for exchanging Azure AD tokens using OAuth 2.0 On-Behalf-Of flow

// RequestBody is the expected JSON structure for the incoming request
type RequestBody struct {
	UserAssertion string   `json:"user_assertion" example:"eyJ0eXAiOiJKV1QiLCJhbGci..." binding:"required"`
	Scopes        []string `json:"scopes" example:"https://graph.microsoft.com/.default" binding:"required"`
}

// TokenResponse represents the Azure AD token response
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJ0eXAiOiJKV1QiLCJhbGci..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int    `json:"expires_in" example:"3599"`
	RefreshToken string `json:"refresh_token,omitempty" example:"0.ARwA6WgJJ9X2qk..."`
	Scope        string `json:"scope" example:"https://graph.microsoft.com/.default"`
}

// ErrorResponse is the JSON structure for an error response
type ErrorResponse struct {
	Error            string `json:"error" example:"invalid_request"`
	ErrorDescription string `json:"error_description,omitempty" example:"The request is missing required parameters"`
}

// Configuration variables (read from environment)
var (
	clientID     = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	tenantID     = os.Getenv("TENANT_ID")
)

func exchangeToken(assertion, scope string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)

	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("assertion", assertion)
	data.Set("scope", scope)
	data.Set("requested_token_use", "on_behalf_of")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &tokenResp, nil
}

// @Summary         Exchange token
// @Description     Exchange a user access token for a downstream service access token
// @Tags            token
// @Accept          json
// @Produce         json
// @Param           request body RequestBody true "Token exchange request"
// @Success         200 {object} TokenResponse
// @Failure         400 {object} ErrorResponse
// @Failure         401 {object} ErrorResponse
// @Failure         405 {object} ErrorResponse
// @Router          / [post]
func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	var reqBody RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("ERROR: Failed to decode request body: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_request", ErrorDescription: "Failed to decode request body: " + err.Error()})
		return
	}
	defer r.Body.Close()

	if reqBody.UserAssertion == "" {
		log.Println("ERROR: user_assertion is missing from request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_request", ErrorDescription: "user_assertion is required"})
		return
	}
	if len(reqBody.Scopes) == 0 {
		log.Println("ERROR: scopes are missing from request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_request", ErrorDescription: "scopes are required"})
		return
	}

	// Join scopes with space for the token request
	scope := strings.Join(reqBody.Scopes, " ")
	log.Printf("INFO: Attempting token exchange for scopes: %v\n", scope)

	tokenResp, err := exchangeToken(reqBody.UserAssertion, scope)
	if err != nil {
		log.Printf("ERROR: Failed to exchange token: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "token_exchange_failed", ErrorDescription: err.Error()})
		return
	}

	log.Printf("INFO: Successfully acquired token for downstream API\n")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResp)
}

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
		handlePostRequest(w, r)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "method_not_allowed", ErrorDescription: "Only GET and POST methods are allowed"})
	}
}

func main() {
	if clientID == "" || clientSecret == "" || tenantID == "" {
		log.Fatal("FATAL: Environment variables CLIENT_ID, CLIENT_SECRET, and TENANT_ID must be set.")
	}

	log.Printf("INFO: Loaded configuration: clientID=%s, clientSecret=%s, tenantID=%s\n", clientID, clientSecret, tenantID)

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
