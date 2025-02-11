package exception

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/bufbuild/protovalidate-go"
	kit "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

type FuncLog func(ctx context.Context, msg string, err error, reasons map[string]string, details logrus.Fields) error
type Func func(msg string, reasons map[string]string) error

var (
	Internal        = NewLog(logrus.ErrorLevel, http.StatusInternalServerError, codes.Internal)
	NotFound        = NewLog(logrus.InfoLevel, http.StatusNotFound, codes.NotFound)
	Invalid         = NewLog(logrus.InfoLevel, http.StatusUnprocessableEntity, codes.InvalidArgument)
	Conflict        = NewLog(logrus.InfoLevel, http.StatusConflict, codes.AlreadyExists)
	Unauthenticated = NewLog(logrus.InfoLevel, http.StatusUnauthorized, codes.Unauthenticated)
	AccessDenied    = NewLog(logrus.InfoLevel, http.StatusForbidden, codes.PermissionDenied)
)

type exception struct {
	Message    string            `json:"message"`
	Reasons    map[string]string `json:"reasons,omitempty"`
	ErrorId    string            `json:"id"`
	httpStatus int
	grpcCode   codes.Code
}

type Exception interface {
	error
	kit.StatusCoder
	json.Marshaler
	GRPCStatus() *status.Status
}

type ValidationError struct {
	Violations []*protovalidate.Violation
}

// NewLog returns an error function containing the message and status codes. Errors are logged.
// The actual error and reasons can be passed later.
func NewLog(logLevel logrus.Level, httpStatus int, grpcCode codes.Code) FuncLog {
	return func(ctx context.Context, msg string, err error, reasons map[string]string, details logrus.Fields) error {
		var errId string
		if err != nil {
			errId = uuid.New().String()
		}

		log(ctx, errId, logLevel, msg, err, &httpStatus, &grpcCode, details)

		return exception{
			Message:    msg,
			Reasons:    reasons,
			ErrorId:    errId,
			httpStatus: httpStatus,
			grpcCode:   grpcCode,
		}
	}
}

func log(ctx context.Context, id string, logLevel logrus.Level, msg string, err error, httpStatus *int, grpcCode *codes.Code, details logrus.Fields) {
	if log := getLogger(ctx); log != nil {
		if err != nil {
			log = log.WithError(err).WithField("errorId", id)
		}
		if details != nil {
			log = log.WithFields(details)
		}
		if httpStatus != nil {
			log = log.WithField("statusCode", httpStatus)
		}
		if grpcCode != nil {
			log = log.WithField("grpcCode", grpcCode)
		}

		log.Log(logLevel, msg)
	}
}

// New returns an error function containing the message and status codes.
// The actual error and reasons can be passed later.
func New(httpStatus int, grpcCode codes.Code) Func {
	return func(msg string, reasons map[string]string) error {
		errId := uuid.New().String()

		return exception{
			Message:    msg,
			Reasons:    reasons,
			ErrorId:    errId,
			httpStatus: httpStatus,
			grpcCode:   grpcCode,
		}
	}
}

// ErrorLog returns an error containing the status codes and logs the error
func ErrorLog(ctx context.Context, logLevel logrus.Level, msg string, err error, reasons map[string]string, httpStatus int, grpcCode codes.Code, details logrus.Fields) error {
	return NewLog(logLevel, httpStatus, grpcCode)(ctx, msg, err, reasons, details)
}

// Error returns an error containing the status codes
func Error(msg string, reasons map[string]string, httpStatus int, grpcCode codes.Code) error {
	return New(httpStatus, grpcCode)(msg, reasons)
}

// Log logs the message
func Log(ctx context.Context, logLevel logrus.Level, msg string, err error, details logrus.Fields) {
	log(ctx, uuid.New().String(), logLevel, msg, err, nil, nil, details)
}

// LogTrace logs the message with the TraceLevel
func LogTrace(ctx context.Context, msg string, err error, details logrus.Fields) {
	Log(ctx, logrus.TraceLevel, msg, err, details)
}

// LogDebug logs the message with the DebugLevel
func LogDebug(ctx context.Context, msg string, err error, details logrus.Fields) {
	Log(ctx, logrus.DebugLevel, msg, err, details)
}

// LogInfo logs the message with the InfoLevel
func LogInfo(ctx context.Context, msg string, err error, details logrus.Fields) {
	Log(ctx, logrus.InfoLevel, msg, err, details)
}

// LogWarn logs the message with the WarnLevel
func LogWarn(ctx context.Context, msg string, err error, details logrus.Fields) {
	Log(ctx, logrus.WarnLevel, msg, err, details)
}

// LogError logs the error with the ErrorLevel
func LogError(ctx context.Context, msg string, err error, details logrus.Fields) {
	Log(ctx, logrus.ErrorLevel, msg, err, details)
}

// LogFatal logs the error with the FatalLevel. It does NOT exit the application.
func LogFatal(ctx context.Context, msg string, err error, details logrus.Fields) {
	Log(ctx, logrus.ErrorLevel, msg, err, details)
}

func ProtoValidationReasons(err error) map[string]string {
	reasons := make(map[string]string)

	var e *protovalidate.ValidationError
	if errors.As(err, &e) {
		for _, v := range e.Violations {
			reasons[string(v.FieldDescriptor.Name())] = v.Proto.GetMessage()
		}
	}

	return reasons
}

func (e exception) Error() string {
	return e.Message
}

// MarshalJSON marshals the error as JSON
func (e exception) MarshalJSON() ([]byte, error) {
	errJson := map[string]interface{}{
		"message": e.Message,
	}

	if e.ErrorId != "" {
		errJson["error_id"] = e.ErrorId
	}
	if e.Reasons != nil && len(e.Reasons) > 0 {
		errJson["reasons"] = e.Reasons
	}

	return json.Marshal(map[string]interface{}{
		"error": errJson,
	})
}

// StatusCode returns the related HTTP status code
func (e exception) StatusCode() int {
	return e.httpStatus
}

// GRPCStatus returns the related gRPC status code
func (e exception) GRPCStatus() *status.Status {
	return status.New(e.grpcCode, e.ErrorId)
}

func getLogger(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return nil
	}

	if l, ok := ctx.Value("log").(*logrus.Entry); ok {
		return l
	}

	return nil
}
