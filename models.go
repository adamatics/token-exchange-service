package main

// RequestBody is the expected JSON structure for the incoming request
type RequestBody struct {
	AdalabToken string   `json:"adalab_token" example:"eyJ0eXAiOiJKV1QiLCJhbGci..." binding:"required"`
	Scopes      []string `json:"scopes" example:"https://graph.microsoft.com/.default" binding:"required"`
}

// RefreshTokenRequestBody is the expected JSON structure for the token refresh request
type RefreshTokenRequestBody struct {
	RefreshToken string   `json:"refresh_token" example:"0.ARwA6WgJJ9X2qk..." binding:"required"`
	Scopes       []string `json:"scopes,omitempty" example:"https://graph.microsoft.com/.default"`
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
