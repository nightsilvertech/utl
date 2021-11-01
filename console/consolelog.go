package console

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/metadata"
	"os"
)

type LogType int
type contextConsoleKey string

const (
	callerDepth = 3
	logFilePerm = 0666
)

const (
	consoleContextRequestID       contextConsoleKey = `request-id`
	consoleContextServiceMetadata contextConsoleKey = `x-microservice-metadata`
)

const (
	LogInfo LogType = iota
	LogWarn
	LogErr
	LogData
)

func (lt LogType) String() string {
	return []string{"log_info", "log_warn", "log_err", "log_data"}[lt]
}

func CreateStdGoKitLog(serviceName string, debug bool) log.Logger {
	f, err := os.OpenFile("service.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, logFilePerm)
	if err != nil {
		panic(fmt.Sprintf("error opening file: %v", err))
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(f))
	logger = log.NewSyncLogger(logger)
	logger = log.With(
		logger,
		"service", serviceName,
		"time", log.DefaultTimestampUTC,
		"caller", log.Caller(callerDepth),
	)
	if debug {
		logger = level.NewFilter(logger, level.AllowDebug())
	}
	return logger
}

func RequestIDMetadataToContext() grpctransport.ServerRequestFunc {
	return func(ctx context.Context, md metadata.MD) context.Context {
		requestID, ok := md[string(consoleContextServiceMetadata)]
		if !ok {
			return ctx
		}
		if ok {
			ctx = context.WithValue(ctx, consoleContextRequestID, requestID[0])
		}
		return ctx
	}
}

func ContextToRequestIDMetadata() grpctransport.ClientRequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		requestID, ok := ctx.Value(consoleContextRequestID).(string)
		if ok {
			(*md)[string(consoleContextServiceMetadata)] = []string{requestID}
		}
		return ctx
	}
}

func createNewLogWithContextRequestID(parentCtx context.Context, l log.Logger, funcName string) (context.Context, log.Logger) {
	requestID := uuid.NewV4().String()
	l = log.With(l, "request_id", requestID, "func_name", funcName)
	return context.WithValue(parentCtx, consoleContextRequestID, requestID), l
}

func Log(ctx context.Context, l log.Logger, funcName string) (context.Context, log.Logger) {
	requestID, ok := ctx.Value(consoleContextRequestID).(string)
	if ok {
		if requestID != "" {
			l = log.With(l, "request_id", requestID, "func_name", funcName)
			return ctx, l
		} else {
			return createNewLogWithContextRequestID(ctx, l, funcName)
		}
	}
	return createNewLogWithContextRequestID(ctx, l, funcName)
}
