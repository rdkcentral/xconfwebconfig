package common

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestHttp400Error_Error(t *testing.T) {
	err := Http400Error{Message: "Bad request error"}

	expected := "Bad request error"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestHttp400Error_EmptyMessage(t *testing.T) {
	err := Http400Error{Message: ""}

	if err.Error() != "" {
		t.Errorf("Expected empty error message, got '%s'", err.Error())
	}
}

func TestHttp404Error_Error(t *testing.T) {
	err := Http404Error{Message: "Resource not found"}

	expected := "Resource not found"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestHttp404Error_LongMessage(t *testing.T) {
	longMsg := "This is a very long error message for a 404 error that contains detailed information about what was not found"
	err := Http404Error{Message: longMsg}

	if err.Error() != longMsg {
		t.Errorf("Expected error message '%s', got '%s'", longMsg, err.Error())
	}
}

func TestHttp500Error_Error(t *testing.T) {
	err := Http500Error{Message: "Internal server error"}

	expected := "Internal server error"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestHttp500Error_SpecialCharacters(t *testing.T) {
	specialMsg := "Error with special chars: !@#$%^&*()_+-={}[]|;':\",./<>?"
	err := Http500Error{Message: specialMsg}

	if err.Error() != specialMsg {
		t.Errorf("Expected error message '%s', got '%s'", specialMsg, err.Error())
	}
}

func TestRemoteHttpError_Error(t *testing.T) {
	err := RemoteHttpError{StatusCode: 404, Message: "Not Found"}

	expected := "Http404 Not Found"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestRemoteHttpError_DifferentStatusCodes(t *testing.T) {
	testCases := []struct {
		statusCode int
		message    string
		expected   string
	}{
		{200, "OK", "Http200 OK"},
		{400, "Bad Request", "Http400 Bad Request"},
		{401, "Unauthorized", "Http401 Unauthorized"},
		{403, "Forbidden", "Http403 Forbidden"},
		{500, "Internal Server Error", "Http500 Internal Server Error"},
		{502, "Bad Gateway", "Http502 Bad Gateway"},
	}

	for _, tc := range testCases {
		err := RemoteHttpError{StatusCode: tc.statusCode, Message: tc.message}
		if err.Error() != tc.expected {
			t.Errorf("For status %d and message '%s', expected '%s', got '%s'",
				tc.statusCode, tc.message, tc.expected, err.Error())
		}
	}
}

func TestNewRemoteError(t *testing.T) {
	status := 404
	message := "Resource not found"

	err := NewRemoteError(status, message)

	if err == nil {
		t.Fatal("NewRemoteError should return a non-nil error")
	}

	// Check if it's the right type
	remoteErr, ok := err.(RemoteHttpError)
	if !ok {
		t.Fatal("NewRemoteError should return a RemoteHttpError")
	}

	if remoteErr.StatusCode != status {
		t.Errorf("Expected status code %d, got %d", status, remoteErr.StatusCode)
	}

	if remoteErr.Message != message {
		t.Errorf("Expected message '%s', got '%s'", message, remoteErr.Message)
	}

	expectedErrorMsg := "Http404 Resource not found"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error string '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestRemoteHttpErrorAS_Error(t *testing.T) {
	err := RemoteHttpErrorAS{StatusCode: 500, Message: "Server error"}

	expected := "Server error"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestNewRemoteErrorAS(t *testing.T) {
	status := 403
	message := "Access denied"

	err := NewRemoteErrorAS(status, message)

	if err == nil {
		t.Fatal("NewRemoteErrorAS should return a non-nil error")
	}

	// Check if it's the right type
	remoteErr, ok := err.(RemoteHttpErrorAS)
	if !ok {
		t.Fatal("NewRemoteErrorAS should return a RemoteHttpErrorAS")
	}

	if remoteErr.StatusCode != status {
		t.Errorf("Expected status code %d, got %d", status, remoteErr.StatusCode)
	}

	if remoteErr.Message != message {
		t.Errorf("Expected message '%s', got '%s'", message, remoteErr.Message)
	}

	if err.Error() != message {
		t.Errorf("Expected error string '%s', got '%s'", message, err.Error())
	}
}

func TestGetXconfErrorStatusCode(t *testing.T) {
	// Test with nil error
	status := GetXconfErrorStatusCode(nil)
	if status != http.StatusOK {
		t.Errorf("Expected status %d for nil error, got %d", http.StatusOK, status)
	}

	// Test with RemoteHttpErrorAS
	remoteErr := NewRemoteErrorAS(404, "Not found")
	status = GetXconfErrorStatusCode(remoteErr)
	if status != 404 {
		t.Errorf("Expected status 404 for RemoteHttpErrorAS, got %d", status)
	}

	// Test with RemoteHttpErrorAS with different status code
	forbiddenErr := NewRemoteErrorAS(403, "Forbidden")
	status = GetXconfErrorStatusCode(forbiddenErr)
	if status != 403 {
		t.Errorf("Expected status 403 for RemoteHttpErrorAS, got %d", status)
	}

	// Test with non-RemoteHttpErrorAS error
	genericErr := errors.New("generic error")
	status = GetXconfErrorStatusCode(genericErr)
	if status != http.StatusInternalServerError {
		t.Errorf("Expected status %d for generic error, got %d", http.StatusInternalServerError, status)
	}

	// Test with Http400Error (should return 500 since it's not RemoteHttpErrorAS)
	http400Err := Http400Error{Message: "Bad request"}
	status = GetXconfErrorStatusCode(http400Err)
	if status != http.StatusInternalServerError {
		t.Errorf("Expected status %d for Http400Error, got %d", http.StatusInternalServerError, status)
	}
}

func TestGetXconfErrorStatusCode_VariousStatusCodes(t *testing.T) {
	testCases := []int{200, 201, 400, 401, 403, 404, 500, 502, 503}

	for _, expectedStatus := range testCases {
		err := NewRemoteErrorAS(expectedStatus, fmt.Sprintf("Error %d", expectedStatus))
		actualStatus := GetXconfErrorStatusCode(err)

		if actualStatus != expectedStatus {
			t.Errorf("Expected status %d, got %d", expectedStatus, actualStatus)
		}
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that error constants are properly defined
	if NotOK == nil {
		t.Error("NotOK should not be nil")
	}

	if NotFound == nil {
		t.Error("NotFound should not be nil")
	}

	if NotFirmwareConfig == nil {
		t.Error("NotFirmwareConfig should not be nil")
	}

	if NotFirmwareRule == nil {
		t.Error("NotFirmwareRule should not be nil")
	}

	if Forbidden == nil {
		t.Error("Forbidden should not be nil")
	}

	// Test error messages
	if NotOK.Error() != "!ok" {
		t.Errorf("Expected NotOK error message '!ok', got '%s'", NotOK.Error())
	}

	if NotFound.Error() != "Not found" {
		t.Errorf("Expected NotFound error message 'Not found', got '%s'", NotFound.Error())
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that error type variables are properly initialized
	if Http400ErrorType == nil {
		t.Error("Http400ErrorType should not be nil")
	}

	if Http404ErrorType == nil {
		t.Error("Http404ErrorType should not be nil")
	}

	if Http500ErrorType == nil {
		t.Error("Http500ErrorType should not be nil")
	}

	if RemoteHttpErrorType == nil {
		t.Error("RemoteHttpErrorType should not be nil")
	}

	if XconfErrorType == nil {
		t.Error("XconfErrorType should not be nil")
	}
}
