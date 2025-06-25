package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// jsonError is a helper function to write a JSON-formatted error message
func jsonError(w http.ResponseWriter, error string, errorDescription string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: error, ErrorDescription: errorDescription})
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
func handleExchangeTokenRequest(w http.ResponseWriter, r *http.Request) {
	var reqBody RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("ERROR: Failed to decode request body: %v\n", err)
		jsonError(w, "invalid_request", "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if reqBody.AdalabToken == "" {
		log.Println("ERROR: adalab_token is missing from request")
		jsonError(w, "invalid_request", "adalab_token is required", http.StatusBadRequest)
		return
	}

	var scope string
	if len(reqBody.Scopes) > 0 {
		scope = strings.Join(reqBody.Scopes, " ")
	} else if defaultScope != "" {
		scope = defaultScope
		log.Printf("INFO: No scopes provided, using default scope: %v\n", defaultScope)
	} else {
		log.Println("ERROR: scopes are missing from request and no default_scope is set")
		jsonError(w, "invalid_request", "scopes are required in the request body when a default scope is not configured", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: Attempting token exchange for scopes: %v\n", scope)

	tokenResp, err := exchangeToken(reqBody.AdalabToken, scope)
	if err != nil {
		log.Printf("ERROR: Failed to exchange token: %v\n", err)
		jsonError(w, "token_exchange_failed", err.Error(), http.StatusUnauthorized)
		return
	}

	log.Printf("INFO: Successfully acquired token for downstream API\n")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResp)
}

// @Summary         Refresh token
// @Description     Exchange a refresh token for a new access token. If scopes are not provided, the service will use the configured default scope.
// @Tags            token
// @Accept          json
// @Produce         json
// @Param           request body RefreshTokenRequestBody true "Token refresh request"
// @Success         200 {object} TokenResponse
// @Failure         400 {object} ErrorResponse
// @Failure         401 {object} ErrorResponse
// @Router          /refresh [post]
func handleRefreshTokenRequest(w http.ResponseWriter, r *http.Request) {
	var reqBody RefreshTokenRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("ERROR: Failed to decode request body: %v\n", err)
		jsonError(w, "invalid_request", "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if reqBody.RefreshToken == "" {
		log.Println("ERROR: refresh_token is missing from request")
		jsonError(w, "invalid_request", "refresh_token is required", http.StatusBadRequest)
		return
	}

	var scope string
	if len(reqBody.Scopes) > 0 {
		scope = strings.Join(reqBody.Scopes, " ")
	} else if defaultScope != "" {
		scope = defaultScope
		log.Printf("INFO: No scopes provided, using default scope: %v\n", defaultScope)
	} else {
		log.Println("ERROR: scopes are missing from request and no default_scope is set")
		jsonError(w, "invalid_request", "scopes are required in the request body when a default scope is not configured", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: Attempting token refresh for scopes: %v\n", scope)

	tokenResp, err := refreshToken(reqBody.RefreshToken, scope)
	if err != nil {
		log.Printf("ERROR: Failed to refresh token: %v\n", err)
		jsonError(w, "token_refresh_failed", err.Error(), http.StatusUnauthorized)
		return
	}

	log.Printf("INFO: Successfully refreshed token for downstream API\n")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResp)
}
