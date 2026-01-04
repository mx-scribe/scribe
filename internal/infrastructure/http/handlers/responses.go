package handlers

import (
	"encoding/json"
	"net/http"
)

// JSONResponse sends a JSON response with the given status code.
func JSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// ErrorResponse sends an error response with the given status code.
func ErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	JSONResponse(w, statusCode, map[string]string{
		"error": message,
	})
}

// SuccessResponse sends a 200 OK response with the given data.
func SuccessResponse(w http.ResponseWriter, data any) {
	JSONResponse(w, http.StatusOK, data)
}

// CreatedResponse sends a 201 Created response with the given data.
func CreatedResponse(w http.ResponseWriter, data any) {
	JSONResponse(w, http.StatusCreated, data)
}
