package handlers

// AppError extends fiber.Error with a machine-readable errorCode for frontend error handling.
// The global Fiber error handler (main.go) detects this type and emits the errorCode field.
type AppError struct {
	HTTPCode  int
	ErrorCode string
	Message   string
}

func (e *AppError) Error() string { return e.Message }

// NewAppError creates an AppError with an HTTP status code, machine-readable error code, and message.
func NewAppError(httpCode int, errorCode, message string) *AppError {
	return &AppError{
		HTTPCode:  httpCode,
		ErrorCode: errorCode,
		Message:   message,
	}
}
