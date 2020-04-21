package logging

import (
	"context"
	"path"
	"time"

	"github.com/google/uuid"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/infobloxopen/atlas-app-toolkit/auth"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/requestid"
)

const (
	DefaultOphIDKey = "ophid"
	clientKind      = "client"
	serverKind      = "server"
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

func ClientInterceptor(logger *logrus.Logger, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()
		fields := newLoggerFields(method, startTime, clientKind)
		err := invoker(ctx, method, req, reply, cc, opts...)
		addSpecificFields(ctx, fields, logger, o)
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		durField, durVal := o.durationFunc(time.Since(startTime))
		fields["grpc.code"] = code.String()
		fields[durField] = durVal

		if err != nil {
			fields[logrus.ErrorKey] = err
		}

		levelLogf(
			logrus.NewEntry(logger).WithFields(fields),
			level,
			"finished unary call with code "+code.String())
		return err
	}
}

func ServerInterceptor(logger *logrus.Logger, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		fields := newLoggerFields(info.FullMethod, startTime, serverKind)
		newCtx := newLoggerForCall(ctx, logrus.NewEntry(logger), fields)

		resp, err := handler(newCtx, req)

		if !o.shouldLog(info.FullMethod, err) {
			return resp, err
		}

		addSpecificFields(ctx, fields, logger, o)
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		durField, durVal := o.durationFunc(time.Since(startTime))
		fields["grpc.code"] = code.String()
		fields[durField] = durVal

		if err != nil {
			fields[logrus.ErrorKey] = err
		}

		levelLogf(
			ctxlogrus.Extract(newCtx).WithFields(fields),
			level,
			"finished unary call with code "+code.String())

		return resp, err
	}
}

func addSpecificFields(ctx context.Context, fields logrus.Fields, logger *logrus.Logger, o *options) {
	reqID, exists := requestid.FromContext(ctx)
	if !exists || reqID == "" {
		reqID = uuid.New().String()
	}
	fields[requestid.DefaultRequestIDKey] = reqID
	ctx = metadata.AppendToOutgoingContext(ctx, requestid.DefaultRequestIDKey, reqID)

	if accountID, err := auth.GetAccountID(ctx, nil); err == nil {
		fields[auth.MultiTenancyField] = accountID
		ctx = metadata.AppendToOutgoingContext(ctx, auth.MultiTenancyField, accountID)
	} else {
		logger.Errorf("Unable to get %s from context", auth.MultiTenancyField)
	}

	if o.ophIDEnabled {
		if ophID, err := auth.GetJWTField(ctx, DefaultOphIDKey, nil); err == nil {
			fields[DefaultOphIDKey] = ophID
			ctx = metadata.AppendToOutgoingContext(ctx, DefaultOphIDKey, ophID)
		} else {
			logger.Errorf("Unable to get %s from context", DefaultOphIDKey)
		}
	}

	// TODO fields["stream"] stdout or stderr. Do we really need this?
}

func newLoggerFields(fullMethodString string, start time.Time, kind string) logrus.Fields {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	return logrus.Fields{
		grpc_logrus.SystemField: "grpc",
		grpc_logrus.KindField:   kind,
		"grpc.service":          service,
		"grpc.method":           method,
		"grpc.start_time":       start.Format(time.RFC3339Nano),
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
