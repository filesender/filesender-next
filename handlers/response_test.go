package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

// Test sendJSON function
func TestSendJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	testData := map[string]string{"foo": "bar"}

	sendJSON(rr, 200, true, "", testData)

	// Check response headers
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type header = %v, want application/json", contentType)
	}

	// Decode response
	var resp Response
	err := json.NewDecoder(rr.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got %v", resp.Success)
	}
	if resp.Data == nil {
		t.Errorf("Expected data, got nil")
	}
}

// Test sendJSON with error case
func TestSendJSON_ErrorCase(t *testing.T) {
	rr := httptest.NewRecorder()

	sendJSON(rr, 400, false, "Something went wrong", nil)

	// Decode response
	var resp Response
	err := json.NewDecoder(rr.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Success {
		t.Errorf("Expected success=false, got %v", resp.Success)
	}
	if resp.Message != "Something went wrong" {
		t.Errorf("Expected message 'Something went wrong', got '%s'", resp.Message)
	}
}

// Test templating function
func TestSendTemplate(t *testing.T) {
	rr := httptest.NewRecorder()

	tmpl, err := loadTemplates()
	if err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

	// Execute the template manually
	// This is necessary since the current implementation uses embedded templating
	// Embedded templating can't go up in directory
	data := map[string]string{
		"Title": "Test Page",
	}
	err = tmpl.Execute(rr, data)
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	responseBody := rr.Body.String()

	// Check if "Test Page" is present
	if !strings.Contains(responseBody, "Test Page") {
		t.Errorf("Expected response to contain 'Test Page', but it didn't.\nResponse body: %s", responseBody)
	}

	// Check if "Hello, World!" is present
	if !strings.Contains(responseBody, "Hello, World!") {
		t.Errorf("Expected response to contain 'Hello, World!', but it didn't.\nResponse body: %s", responseBody)
	}
}

// Test sendError function
func TestSendError(t *testing.T) {
	rr := httptest.NewRecorder()

	sendError(rr, 404, "Not Found")

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %v", status)
	}

	expected := "Not Found\n"
	if rr.Body.String() != expected {
		t.Errorf("Expected response body %q, got %q", expected, rr.Body.String())
	}
}

// Helper function to read template files manually
func loadTemplates() (*template.Template, error) {
	basePath := filepath.Join("..", "cmd", "filesender", "templates", "base.html")
	testPath := filepath.Join("..", "cmd", "filesender", "templates", "test.html")

	tmpl, err := template.ParseFiles(basePath, testPath)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}
