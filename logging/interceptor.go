package logging

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/google/uuid"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/infobloxopen/atlas-app-toolkit/auth"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/requestid"
)

const (
	DefaultAccountIDKey     = "account_id"
	DefaultRequestIDKey     = "request_id"
	DefaultGRPCCodeKey      = "grpc.code"
	DefaultGRPCMethodKey    = "grpc.method"
	DefaultGRPCServiceKey   = "grpc.service"
	DefaultGRPCStartTimeKey = "grpc.start_time"
	DefaultClientKindValue  = "client"
	DefaultServerKindValue  = "server"
)

// LogLevelInterceptor sets the level of the logger in the context to either
// the default or the value set in the context via grpc metadata.
// Also sets the custom log tag if present for pseudo-tracing purposes
func LogLevelInterceptor(defaultLevel logrus.Level) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		entry := ctxlogrus.Extract(ctx)
		lvl := defaultLevel
		logFlag, hasFlag := gateway.Header(ctx, logFlagMetaKey)
		if hasFlag {
			entry.Data[logFlagFieldName] = logFlag
		}
		if logLvl, ok := gateway.Header(ctx, logLevelMetaKey); ok {
			entry.Debugf("Using custom log-level of %q", logLvl)
			lvl, err = logrus.ParseLevel(logLvl)
			if err != nil {
				lvl = defaultLevel
			}
		}
		newLogger := CopyLoggerWithLevel(entry.Logger, lvl)
		newCtx := ctxlogrus.ToContext(ctx, newLogger.WithFields(entry.Data))
		res, err = handler(newCtx, req)

		// propagate any new or changed fields from later interceptors back up
		// the middleware chain
		resLogger := ctxlogrus.Extract(newCtx)
		ctxlogrus.AddFields(ctx, resLogger.Data)
		return
	}
}

func UnaryClientInterceptor(logger *logrus.Logger, opts ...Option) grpc.UnaryClientInterceptor {
	options := initOptions(opts)

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()
		fields := newLoggerFields(method, startTime, DefaultClientKindValue)

		ctx = fillInterceptor(ctx, fields, logger, options, startTime)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			fields[logrus.ErrorKey] = err
		}

		code := status.Code(err)
		fields[DefaultGRPCCodeKey] = code.String()

		levelLogf(
			logrus.NewEntry(logger).WithFields(fields),
			options.codeToLevel(code),
			"finished unary call with code "+code.String())

		return err
	}
}

func UnaryServerInterceptor(logger *logrus.Logger, opts ...Option) grpc.UnaryServerInterceptor {
	options := initOptions(opts)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		fields := newLoggerFields(info.FullMethod, startTime, DefaultServerKindValue)
		newCtx := newLoggerForCall(ctx, logrus.NewEntry(logger), fields)

		newCtx = fillInterceptor(newCtx, fields, logger, options, startTime)

		resp, err := handler(newCtx, req)
		if err != nil {
			fields[logrus.ErrorKey] = err
		}

		code := status.Code(err)
		fields[DefaultGRPCCodeKey] = code.String()

		levelLogf(
			ctxlogrus.Extract(newCtx).WithFields(fields),
			options.codeToLevel(code),
			"finished unary call with code "+code.String())

		return resp, err
	}
}

func fillInterceptor(ctx context.Context, fields logrus.Fields, logger *logrus.Logger, options *options, start time.Time) context.Context {
	durField, durVal := options.durationFunc(time.Since(start))
	fields[durField] = durVal

	ctx = addRequestIDField(ctx, fields, logger, options)
	ctx, err := addAccountIDField(ctx, fields, logger, options)
	if err != nil {
		logger.Warn(err)
	}

	for _, v := range options.fields {
		ctx, err = addCustomField(ctx, fields, logger, v)
		if err != nil {
			logger.Warn(err)
		}
	}

	return ctx
}

func addRequestIDField(ctx context.Context, fields logrus.Fields, logger *logrus.Logger, o *options) context.Context {
	reqID, exists := requestid.FromContext(ctx)
	if !exists || reqID == "" {
		reqID = uuid.New().String()
	}

	fields[DefaultRequestIDKey] = reqID

	return metadata.AppendToOutgoingContext(ctx, DefaultRequestIDKey, reqID)
}

func addAccountIDField(ctx context.Context, fields logrus.Fields, logger *logrus.Logger, o *options) (context.Context, error) {
	accountID, err := auth.GetAccountID(ctx, nil)
	if err != nil {
		return ctx, fmt.Errorf("Unable to get %q from context", DefaultAccountIDKey)
	}

	fields[DefaultAccountIDKey] = accountID

	return metadata.AppendToOutgoingContext(ctx, auth.MultiTenancyField, accountID), err
}

func addCustomField(ctx context.Context, fields logrus.Fields, logger *logrus.Logger, customField string) (context.Context, error) {
	field, err := auth.GetJWTField(ctx, customField, nil)
	if err != nil {
		return ctx, fmt.Errorf("Unable to get custom %q field from context", customField)
	}

	fields[customField] = field
	ctx = metadata.AppendToOutgoingContext(ctx, customField, field)

	return metadata.AppendToOutgoingContext(ctx, customField, field), err
}

func newLoggerFields(fullMethodString string, start time.Time, kind string) logrus.Fields {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)

	return logrus.Fields{
		grpc_logrus.SystemField: "grpc",
		grpc_logrus.KindField:   kind,
		DefaultGRPCServiceKey:   service,
		DefaultGRPCMethodKey:    method,
		DefaultGRPCStartTimeKey: start.Format(time.RFC3339Nano),
	}
}

func newLoggerForCall(ctx context.Context, entry *logrus.Entry, fields logrus.Fields) context.Context {
	callLog := entry.WithFields(fields)
	callLog = callLog.WithFields(ctxlogrus.Extract(ctx).Data)

	return ctxlogrus.ToContext(ctx, callLog)
}

// CopyLoggerWithLevel makes a copy of the given (logrus) logger at the logger
// level. If copying an entry, use CopyLoggerWithLevel(entry.Logger, level).WithFields(entry.Data)
// on the result (changes to these entries' fields will not affect each other).
func CopyLoggerWithLevel(logger *logrus.Logger, lvl logrus.Level) *logrus.Logger {
	newLogger := &logrus.Logger{
		Out:          logger.Out,
		Hooks:        make(logrus.LevelHooks),
		Level:        lvl,
		Formatter:    logger.Formatter,
		ReportCaller: logger.ReportCaller,
		ExitFunc:     logger.ExitFunc,
	}
	// Copy hooks, so that original Logger hooks are not altered
	for l, hook := range logger.Hooks {
		newLogger.Hooks[l] = hook
	}
	return newLogger
}
