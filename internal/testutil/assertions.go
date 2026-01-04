package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// AssertStatusCode checks that the response has the expected status code.
func AssertStatusCode(t *testing.T, resp *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if resp.Code != expected {
		t.Errorf("Expected status code %d, got %d", expected, resp.Code)
	}
}

// AssertContentType checks that the response has the expected content type.
func AssertContentType(t *testing.T, resp *httptest.ResponseRecorder, expected string) {
	t.Helper()
	contentType := resp.Header().Get("Content-Type")
	if !strings.Contains(contentType, expected) {
		t.Errorf("Expected content type to contain %s, got %s", expected, contentType)
	}
}

// AssertJSONResponse checks that the response is valid JSON.
func AssertJSONResponse(t *testing.T, resp *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}
	return result
}

// AssertJSONArrayResponse checks that the response is a valid JSON array.
func AssertJSONArrayResponse(t *testing.T, resp *httptest.ResponseRecorder) []any {
	t.Helper()

	var result []any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode JSON array response: %v", err)
	}
	return result
}

// AssertJSONField checks that a JSON response has a field with expected value.
func AssertJSONField(t *testing.T, data map[string]any, field string, expected any) {
	t.Helper()

	value, exists := data[field]
	if !exists {
		t.Errorf("Expected field %s to exist in response", field)
		return
	}

	// Compare values (handle type differences)
	switch e := expected.(type) {
	case int:
		if v, ok := value.(float64); ok {
			if int(v) != e {
				t.Errorf("Expected %s to be %d, got %v", field, e, value)
			}
		} else {
			t.Errorf("Expected %s to be int, got %T", field, value)
		}
	case string:
		if v, ok := value.(string); ok {
			if v != e {
				t.Errorf("Expected %s to be %s, got %s", field, e, v)
			}
		} else {
			t.Errorf("Expected %s to be string, got %T", field, value)
		}
	default:
		if value != expected {
			t.Errorf("Expected %s to be %v, got %v", field, expected, value)
		}
	}
}

// AssertBodyContains checks that the response body contains a string.
func AssertBodyContains(t *testing.T, resp *httptest.ResponseRecorder, substring string) {
	t.Helper()
	body := resp.Body.String()
	if !strings.Contains(body, substring) {
		t.Errorf("Expected body to contain %q, got %q", substring, body)
	}
}

// AssertNoError checks that an error is nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// AssertError checks that an error is not nil.
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// MakeRequest creates an HTTP request for testing.
func MakeRequest(method, url string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// MakeJSONRequest creates an HTTP request with JSON body.
func MakeJSONRequest(t *testing.T, method, url string, data any) *http.Request {
	t.Helper()

	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(method, url, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	return req
}
