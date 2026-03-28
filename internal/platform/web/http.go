package platformweb

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

var (
	ErrEmptyJSONBody      = errors.New("empty json body")
	ErrInvalidJSONBody    = errors.New("invalid json body")
	ErrUnexpectedJSONData = errors.New("unexpected json data")
)

type ErrorResponse struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, code string, message string) {
	WriteJSON(w, statusCode, ErrorResponse{
		Error: APIError{
			Code:      strings.TrimSpace(code),
			Message:   strings.TrimSpace(message),
			RequestID: chimiddleware.GetReqID(r.Context()),
		},
	})
}

func DecodeJSONBody(r *http.Request, dst any) error {
	if r.Body == nil {
		return ErrEmptyJSONBody
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return ErrEmptyJSONBody
		}
		return ErrInvalidJSONBody
	}

	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return ErrInvalidJSONBody
	}

	return ErrUnexpectedJSONData
}
