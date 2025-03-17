package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock JSON response writer
func TestRecvJSON(t *testing.T) {
	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name           string
		requestBody    string
		expectedResult bool
		statusCode     int
	}{
		{
			name:           "Valid JSON",
			requestBody:    `{"name": "John", "age": 30}`,
			expectedResult: true,
			statusCode:     http.StatusOK,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"name": "John", "age":}`,
			expectedResult: false,
			statusCode:     http.StatusBadRequest,
		},
		{
			name:           "Empty Body",
			requestBody:    ``,
			expectedResult: false,
			statusCode:     http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result testStruct
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(tt.requestBody)))
			request.Header.Set("Content-Type", "application/json")

			responseRecorder := httptest.NewRecorder()
			got := recvJSON(responseRecorder, request, &result)

			if got != tt.expectedResult {
				t.Errorf("recvJSON() got %v, expected %v", got, tt.expectedResult)
			}

			if responseRecorder.Code != tt.statusCode && !tt.expectedResult {
				t.Errorf("HTTP status code got %v, expected %v", responseRecorder.Code, tt.statusCode)
			}
		})
	}
}
