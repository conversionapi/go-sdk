package enconvert

import (
	"errors"
	"fmt"
)

// APIError represents an HTTP-level failure returned by the Enconvert API.
// Its Error() string form is "[<status>] <message>", matching every other
// Enconvert SDK.
//
// Convenience predicates are provided for the status codes the API uses to
// signal specific conditions: 401/403 (authentication), 402 (quota/plan
// gate), and 429 (rate limit). Use errors.As for anything more specific.
type APIError struct {
	StatusCode int
	Message    string
}

// NewAPIError constructs an *APIError with the given status code and message.
func NewAPIError(statusCode int, message string) *APIError {
	return &APIError{StatusCode: statusCode, Message: message}
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
}

// newAuthenticationError mirrors the Node SDK's AuthenticationError, which
// always records status 401 regardless of whether the response was 401 or
// 403 — a quirk preserved here for parity.
func newAuthenticationError(message string) *APIError {
	if message == "" {
		message = "Invalid or missing API key"
	}
	return &APIError{StatusCode: 401, Message: message}
}

func newQuotaError(message string) *APIError {
	if message == "" {
		message = "Plan feature not enabled or quota exhausted"
	}
	return &APIError{StatusCode: 402, Message: message}
}

func newRateLimitError(message string) *APIError {
	if message == "" {
		message = "Rate limit exceeded"
	}
	return &APIError{StatusCode: 429, Message: message}
}

// IsAuthenticationError reports whether err is an *APIError representing an
// authentication failure (HTTP 401 or 403 from the API — always recorded as
// status 401, matching the other Enconvert SDKs).
func IsAuthenticationError(err error) bool {
	apiErr, ok := asAPIError(err)
	return ok && apiErr.StatusCode == 401
}

// IsQuotaError reports whether err is an *APIError representing a disabled
// plan feature or exhausted quota (HTTP 402).
func IsQuotaError(err error) bool {
	apiErr, ok := asAPIError(err)
	return ok && apiErr.StatusCode == 402
}

// IsRateLimitError reports whether err is an *APIError representing a rate
// limit (HTTP 429).
func IsRateLimitError(err error) bool {
	apiErr, ok := asAPIError(err)
	return ok && apiErr.StatusCode == 429
}

func asAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}
