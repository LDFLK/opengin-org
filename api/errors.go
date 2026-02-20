package api

import "fmt"

// HTTPError is the base type for HTTP errors returned by the API.
type HTTPError struct {
	StatusCode int
	Detail     string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Detail)
}

// BadRequestError represents a 400 response.
type BadRequestError struct{ HTTPError }

// NotFoundError represents a 404 response.
type NotFoundError struct{ HTTPError }

// InternalServerError represents a 500 response.
type InternalServerError struct{ HTTPError }

// ServiceUnavailableError represents a 503 response.
type ServiceUnavailableError struct{ HTTPError }

// GatewayTimeoutError represents a 504 response.
type GatewayTimeoutError struct{ HTTPError }

// httpErrorFromStatus converts an HTTP status code and body into a typed error.
func httpErrorFromStatus(statusCode int, body string) error {
	switch statusCode {
	case 400:
		return &BadRequestError{HTTPError{StatusCode: statusCode, Detail: body}}
	case 404:
		return &NotFoundError{HTTPError{StatusCode: statusCode, Detail: body}}
	case 500:
		return &InternalServerError{HTTPError{StatusCode: statusCode, Detail: body}}
	case 503:
		return &ServiceUnavailableError{HTTPError{StatusCode: statusCode, Detail: body}}
	case 504:
		return &GatewayTimeoutError{HTTPError{StatusCode: statusCode, Detail: body}}
	default:
		return &HTTPError{StatusCode: statusCode, Detail: body}
	}
}
