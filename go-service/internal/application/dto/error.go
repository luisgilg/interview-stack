package dto

// ErrorResponse describes a generic error payload for HTTP responses.
type ErrorResponse struct {
	Error string `json:"error" example:"resource not found"`
}

// HealthResponse captures the data returned by the health endpoint.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}
