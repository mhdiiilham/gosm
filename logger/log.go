package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

// RequestID represents a unique identifier for a request, used for tracking logs.
type RequestID string

// RequestIDKey is the context key used to store and retrieve the request ID.
var RequestIDKey RequestID = "request-id"

// Infof logs an informational message with contextual request ID and caller details.
// It accepts a context, caller name, formatted message, and optional values.
func Infof(ctx context.Context, caller, format string, values ...interface{}) {
	requestID := "-"
	ctxVal := ctx.Value(RequestIDKey)
	if ctxVal != nil {
		requestID = ctxVal.(string)
	}

	logrus.WithFields(logrus.Fields{
		"caller":     caller,
		"request_id": requestID,
	}).Infof(format, values...)
}

// Errorf logs an error message with contextual request ID and caller details.
// It accepts a context, caller name, formatted message, and optional values.
func Errorf(ctx context.Context, caller, format string, values ...interface{}) {
	requestID := "-"
	ctxVal := ctx.Value(RequestIDKey)
	if ctxVal != nil {
		requestID = ctxVal.(string)
	}

	logrus.WithFields(logrus.Fields{
		"caller":     caller,
		"request_id": requestID,
	}).Errorf(format, values...)
}

// Warn logs a warning message with contextual request ID and caller details.
// It accepts a context, caller name, formatted message, and optional values.
func Warn(ctx context.Context, caller, format string, values ...interface{}) {
	requestID := "-"
	ctxVal := ctx.Value(RequestIDKey)
	if ctxVal != nil {
		requestID = ctxVal.(string)
	}

	logrus.WithFields(logrus.Fields{
		"caller":     caller,
		"request_id": requestID,
	}).Warnf(format, values...)
}

// Fatalf logs a fatal error message with contextual information, including the caller and request ID.
// It retrieves the request ID from the provided context and logs the formatted message with a fatal level.
// This function will terminate the application after logging.
func Fatalf(ctx context.Context, caller, format string, values ...interface{}) {
	requestID := "-"
	ctxVal := ctx.Value(RequestIDKey)
	if ctxVal != nil {
		requestID = ctxVal.(string)
	}

	logrus.WithFields(logrus.Fields{
		"caller":     caller,
		"request_id": requestID,
	}).Fatalf(format, values...)
}
