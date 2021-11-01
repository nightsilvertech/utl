package console

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/metadata"
	"os"
)

type LogType int
type contextKey int

const logFilePerm = 0666

const (
	contextRequestIDKey contextKey = iota
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
		"caller", log.Caller(3),
	)
	if debug {
		logger = level.NewFilter(logger, level.AllowDebug())
	}
	return logger
}

func Log(ctx context.Context, l log.Logger, funcName string) (context.Context, log.Logger) {
	ck, ok := ctx.Value("reqid").(string)
	if ok {
		if ck != "" {
			fmt.Println("request id exist")
			l = log.With(l, "request_id", ck, "func_name", funcName)
			return ctx, l
		} else {
			fmt.Println("request id not exist")
			requestID := uuid.NewV4().String()
			l = log.With(l, "request_id", requestID, "func_name", funcName)
			return metadata.NewIncomingContext(ctx, metadata.Pairs("requestid", requestID)), l
		}
	} else {
		fmt.Println("request id not exist not ok")
		requestID := uuid.NewV4().String()
		l = log.With(l, "request_id", requestID, "func_name", funcName)
		return metadata.NewIncomingContext(ctx, metadata.Pairs("requestid", requestID)), l
	}
}
