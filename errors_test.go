package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

// TestNewApiError checks if the error is correctly created
func TestNewApiError(t *testing.T) {
	// Arrange: Define the error type and message
	errorType := NotFoundErrorType
	userMessage := "User not found"

	// Act: Create the ApiError
	apiError := NewApiError(errorType, userMessage)

	// Assert: Check the fields
	if apiError.ErrorType != errorType {
		t.Errorf("expected error type %s, got %s", errorType, apiError.ErrorType)
	}

	if apiError.Message != userMessage {
		t.Errorf("expected message %s, got %s", userMessage, apiError.Message)
	}

	if apiError.ErrorCode != http.StatusNotFound {
		t.Errorf("expected error code %d, got %d", http.StatusNotFound, apiError.ErrorCode)
	}
}

// TestWithInnerError checks if the internal error is correctly added
func TestWithInnerError(t *testing.T) {
	// Arrange: Create an internal error and user-facing error
	internalErr := fmt.Errorf("database connection failed")
	userMessage := "User not found"
	apiError := NewApiError(NotFoundErrorType, userMessage, WithInternalError(internalErr))

	// Act: Verify internal error is attached
	if apiError.InnerError == nil {
		t.Error("expected internal error, got nil")
	}

	if apiError.InnerError.Error() != internalErr.Error() {
		t.Errorf("expected internal error %s, got %s", internalErr.Error(), apiError.InnerError.Error())
	}
}

// TestMarshalJSON checks the JSON marshaling
func TestMarshalJSON(t *testing.T) {
	// Arrange: Create an ApiError
	apiError := NewApiError(NotFoundErrorType, "User not found")

	// Act: Marshal the ApiError to JSON
	jsonData, err := json.Marshal(apiError)
	if err != nil {
		t.Fatalf("failed to marshal ApiError: %v", err)
	}

	// Assert: Check if the JSON contains expected fields
	expectedJSON := `{"error_type":"NotFoundError","message":"User not found","error_code":404}`
	if string(jsonData) != expectedJSON {
		t.Errorf("expected %s, got %s", expectedJSON, string(jsonData))
	}
}
func TestMarshalJSONWithInnerError(t *testing.T) {
	// Arrange: Create an ApiError
	apiError := NewApiError(NotFoundErrorType, "User not found", WithInternalError(errors.New("test internal error")))

	// Act: Marshal the ApiError to JSON
	jsonData, err := apiError.MarshalJSON()
	if err != nil {
		t.Fatalf("failed to marshal ApiError: %v", err)
	}

	// Assert: Check if the JSON contains expected fields
	expectedJSON := `{"internal_error":"test internal error","error_type":"NotFoundError","message":"User not found","error_code":404}`
	if string(jsonData) != expectedJSON {
		t.Errorf("expected %s, got %s", expectedJSON, string(jsonData))
	}
}

// TestUnmarshalJSON checks the JSON unmarshaling
func TestUnmarshalJSON(t *testing.T) {
	// Arrange: Create a JSON string
	jsonStr := `{"error_type":"NotFoundError","message":"User not found","error_code":404}`

	// Act: Unmarshal the JSON string to ApiError
	var apiError ApiError
	err := json.Unmarshal([]byte(jsonStr), &apiError)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Assert: Check if fields are correctly populated
	if apiError.ErrorType != NotFoundErrorType {
		t.Errorf("expected error type %s, got %s", NotFoundErrorType, apiError.ErrorType)
	}

	if apiError.Message != "User not found" {
		t.Errorf("expected message %s, got %s", "User not found", apiError.Message)
	}

	if apiError.ErrorCode != http.StatusNotFound {
		t.Errorf("expected error code %d, got %d", http.StatusNotFound, apiError.ErrorCode)
	}
}

func TestUnmarshalJSONWithInnerError(t *testing.T) {
	// Arrange: Create a JSON string
	jsonStr := `{"internal_error":"test internal error","error_type":"NotFoundError","message":"User not found","error_code":404}`

	// Act: Unmarshal the JSON string to ApiError
	var apiError ApiError
	err := apiError.UnmarshalJSON([]byte(jsonStr))
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Assert: Check if fields are correctly populated
	if apiError.ErrorType != NotFoundErrorType {
		t.Errorf("expected error type %s, got %s", NotFoundErrorType, apiError.ErrorType)
	}

	if apiError.Message != "User not found" {
		t.Errorf("expected message %s, got %s", "User not found", apiError.Message)
	}

	if apiError.ErrorCode != http.StatusNotFound {
		t.Errorf("expected error code %d, got %d", http.StatusNotFound, apiError.ErrorCode)
	}

	if apiError.InnerError.Error() != "test internal error" {
		t.Errorf("expected Internal error%s, got %s", "test internal error", apiError.InnerError.Error())
	}
}

func TestUnmarshalJSONRaisedErrorWhenJsonIsNotWalid(t *testing.T) {
	// Arrange: Create a JSON string
	jsonStr := `{"internal_error":"test internal error","error_type":"NotFoundError","message":"User not found","error_code":404`

	// Act: Unmarshal the JSON string to ApiError
	var apiError ApiError
	err := apiError.UnmarshalJSON([]byte(jsonStr))
	if err == nil {
		t.Fatalf("unmarshal Invalid JSON: %v should raied error", err)
	}
}

func TestRegisterErrorType(t *testing.T) {
	newErrorType := "UserError"
	newErrorCode := http.StatusTeapot
	newErrorMessage := "I'm a teapot"

	// Call the RegisterErrorType function to add the new error
	RegisterErrorType(newErrorType, newErrorCode, newErrorMessage)

	// Retrieve the registered error type from the ErrorRegistry
	registeredError, exists := ErrorRegistry[newErrorType]

	// Check if the error type exists in the registry
	if !exists {
		t.Errorf("Expected error type %s to be registered but it wasn't", newErrorType)
	}

	// Verify the error code and message are correct
	if registeredError.ErrorCode != newErrorCode {
		t.Errorf("Expected error code %d, but got %d", newErrorCode, registeredError.ErrorCode)
	}

	if registeredError.Message != newErrorMessage {
		t.Errorf("Expected error message '%s', but got '%s'", newErrorMessage, registeredError.Message)
	}
}

func TestApiError_Error(t *testing.T) {
	apiError := NewApiError(NotFoundErrorType, "User not found")
	if apiError.Error() != "Error 404: User not found" {
		t.Errorf("Expected error message '%s', but got '%s'", "Error 404:User not found", apiError.Error())
	}
}

func TestApiError_HttPError(t *testing.T) {
	apiError := NewApiError(NotFoundErrorType, "User not found")
	code, message := apiError.HTTPError()
	if code != http.StatusNotFound {
		t.Errorf("Expected error code '%d', but got '%d'", http.StatusNotFound, code)
	}
	if message != "User not found" {
		t.Errorf("Expected error message '%s', but got '%s'", "User not found", message)
	}
}

// TestApiError_Unwrap ensures errors.Is and errors.As walk into the wrapped
// internal cause via the Unwrap method.
func TestApiError_Unwrap(t *testing.T) {
	sentinel := errors.New("db unreachable")
	apiError := NewApiError(InternalServerErrorType, "something broke", WithInternalError(sentinel))

	if !errors.Is(apiError, sentinel) {
		t.Errorf("errors.Is should match the wrapped sentinel via Unwrap")
	}

	var target *ApiError
	if !errors.As(apiError, &target) {
		t.Errorf("errors.As should unwrap into *ApiError")
	}

	if apiError.Unwrap() != sentinel {
		t.Errorf("Unwrap() should return the wrapped sentinel")
	}
}

func TestApiError_UnwrapReturnsNilWhenNoInnerError(t *testing.T) {
	apiError := NewApiError(NotFoundErrorType, "User not found")
	if apiError.Unwrap() != nil {
		t.Errorf("Unwrap() should return nil when no internal error is wrapped")
	}
}

// TestUnmarshalJSONSafeFromFormatVerbs guards against a previous
// fmt.Errorf(*aux.InternalError) footgun: a persisted error containing
// percent signs must round-trip literally, not be interpreted as format verbs.
func TestUnmarshalJSONSafeFromFormatVerbs(t *testing.T) {
	jsonStr := `{"internal_error":"query failed: %s AND %d","error_type":"InternalServerError","message":"boom","error_code":500}`

	var apiError ApiError
	if err := apiError.UnmarshalJSON([]byte(jsonStr)); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	got := apiError.InnerError.Error()
	want := "query failed: %s AND %d"
	if got != want {
		t.Errorf("expected internal error %q (literal percent signs), got %q", want, got)
	}
}

// TestApiError_InternalError confirms the method now returns the wrapped
// cause via the ApiErrors interface contract (finding #7).
func TestApiError_InternalError(t *testing.T) {
	sentinel := errors.New("network timeout")
	apiError := NewApiError(RequestTimeoutErrorType, "too slow", WithInternalError(sentinel))

	var asInterface ApiErrors = apiError
	if asInterface.InternalError() != sentinel {
		t.Errorf("InternalError() via ApiErrors interface should return the wrapped sentinel")
	}
}

func TestNewApiError_FallsBackToGenericOnUnknownType(t *testing.T) {
	apiError := NewApiError("NotARegisteredType", "something weird")
	if apiError.ErrorType != GenericErrorType {
		t.Errorf("expected fallback type %s, got %s", GenericErrorType, apiError.ErrorType)
	}
	if apiError.ErrorCode != http.StatusInternalServerError {
		t.Errorf("expected fallback code %d, got %d", http.StatusInternalServerError, apiError.ErrorCode)
	}
}
