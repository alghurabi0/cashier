package handler

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON success response.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := map[string]interface{}{"data": data}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"error":"failed to encode response","code":500}`, http.StatusInternalServerError)
	}
}

// Error writes a JSON error response.
func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := map[string]interface{}{
		"error": message,
		"code":  status,
	}
	json.NewEncoder(w).Encode(resp)
}

// ValidationError writes a JSON validation error response with field-level details.
func ValidationError(w http.ResponseWriter, errors map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)

	resp := map[string]interface{}{
		"error":   "validation failed",
		"code":    http.StatusUnprocessableEntity,
		"details": errors,
	}
	json.NewEncoder(w).Encode(resp)
}

// DecodeJSON decodes a JSON request body into the target struct.
// Returns false and writes an error response if decoding fails.
func DecodeJSON(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	if r.Body == nil {
		Error(w, http.StatusBadRequest, "request body is required")
		return false
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		Error(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return false
	}

	return true
}
