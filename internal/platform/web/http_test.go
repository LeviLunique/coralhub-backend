package platformweb

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONBodyRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"Ana","extra":true}`))

	var payload struct {
		Name string `json:"name"`
	}

	err := DecodeJSONBody(req, &payload)
	if err != ErrInvalidJSONBody {
		t.Fatalf("DecodeJSONBody() error = %v, want %v", err, ErrInvalidJSONBody)
	}
}

func TestDecodeJSONBodyRejectsMultipleJSONDocuments(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"Ana"}{"other":"value"}`))

	var payload struct {
		Name string `json:"name"`
	}

	err := DecodeJSONBody(req, &payload)
	if err != ErrUnexpectedJSONData {
		t.Fatalf("DecodeJSONBody() error = %v, want %v", err, ErrUnexpectedJSONData)
	}
}
