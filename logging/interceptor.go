package logging

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/infobloxopen/atlas-app-toolkit/auth"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/requestid"
)

const (
	DefaultAccountIDKey     = "account_id"
	DefaultRequestIDKey     = "request_id"
	DefaultSubjectKey       = "subject" // Might be used for different purposes
	DefaultDurationKey      = "grpc.time_ms"
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

func UnaryClientInterceptor(entry *logrus.Entry, opts ...Option) grpc.UnaryClientInterceptor {
	options := initOptions(opts)

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()
		fields := newLoggerFields(method, startTime, DefaultClientKindValue)

		setInterceptorFields(ctx, fields, entry.Logger, options, startTime)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			fields[logrus.ErrorKey] = err
		}

		code := status.Code(err)
		fields[DefaultGRPCCodeKey] = code.String()

		levelLogf(
			entry.WithFields(fields),
			options.codeToLevel(code),
			"finished unary call with code %s", code.String())

		return err
	}
}

func StreamClientInterceptor(entry *logrus.Entry, opts ...Option) grpc.StreamClientInterceptor {
	options := initOptions(opts)

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, option ...grpc.CallOption) (grpc.ClientStream, error) {
		startTime := time.Now()
		fields := newLoggerFields(method, startTime, DefaultClientKindValue)

		setInterceptorFields(ctx, fields, entry.Logger, options, startTime)

		clientStream, err := streamer(ctx, desc, cc, method, option...)
		if err != nil {
			fields[logrus.ErrorKey] = err
		}

		code := status.Code(err)
		fields[DefaultGRPCCodeKey] = code.String()

		levelLogf(
			entry.WithFields(fields),
			options.codeToLevel(code),
			"finished client streaming call with code %s", code.String())

		return clientStream, err
	}
}

func UnaryServerInterceptor(entry *logrus.Entry, opts ...Option) grpc.UnaryServerInterceptor {
	options := initOptions(opts)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		fields := newLoggerFields(info.FullMethod, startTime, DefaultServerKindValue)

		setInterceptorFields(ctx, fields, entry.Logger, options, startTime)

		newCtx := newLoggerForCall(ctx, entry, fields)

		resp, err := handler(newCtx, req)
		if err != nil {
			fields[logrus.ErrorKey] = err
		}

		code := status.Code(err)
		fields[DefaultGRPCCodeKey] = code.String()

		levelLogf(
			ctxlogrus.Extract(newCtx).WithFields(fields),
			options.codeToLevel(code),
			"finished unary call with code %s", code.String())

		return resp, err
	}
}

func StreamServerInterceptor(entry *logrus.Entry, opts ...Option) grpc.StreamServerInterceptor {
	options := initOptions(opts)

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()
		fields := newLoggerFields(info.FullMethod, startTime, DefaultServerKindValue)

		setInterceptorFields(stream.Context(), fields, entry.Logger, options, startTime)

		newCtx := newLoggerForCall(stream.Context(), entry, fields)

		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx

		err := handler(srv, wrapped)
		if err != nil {
			fields[logrus.ErrorKey] = err
		}

		code := status.Code(err)
		fields[DefaultGRPCCodeKey] = code.String()

		levelLogf(
			ctxlogrus.Extract(newCtx).WithFields(fields),
			options.codeToLevel(code),
			"finished server streaming call with code %s", code.String())

		return err
	}
}

func setInterceptorFields(ctx context.Context, fields logrus.Fields, logger *logrus.Logger, options *options, start time.Time) {
	// In latest versions of Go use
	// https://golang.org/src/time/time.go?s=25178:25216#L780
	duration := int64(time.Since(start) / 1e6)
	fields[DefaultDurationKey] = duration

	err := addRequestIDField(ctx, fields)
	if err != nil {
		logger.Warn(err)
	}

	err = addAccountIDField(ctx, fields)
	if err != nil {
		logger.Warn(err)
	}

	err = addCustomField(ctx, fields, DefaultSubjectKey)
	if err != nil {
		logger.Warn(err)
	}

	for _, v := range options.fields {
		err = addCustomField(ctx, fields, v)
		if err != nil {
			logger.Warn(err)
		}
	}

	for _, v := range options.headers {
		err = addHeaderField(ctx, fields, v)
		if err != nil {
			logger.Warn(err)
		}
	}
}

func addRequestIDField(ctx context.Context, fields logrus.Fields) error {
	reqID, exists := requestid.FromContext(ctx)
	if !exists || reqID == "" {
		return fmt.Errorf("Unable to get %q from context", DefaultRequestIDKey)
	}

	fields[DefaultRequestIDKey] = reqID

	return nil
}

func addAccountIDField(ctx context.Context, fields logrus.Fields) error {
	accountID, err := auth.GetAccountID(ctx, nil)
	if err != nil {
		return fmt.Errorf("Unable to get %q from context", DefaultAccountIDKey)
	}

	fields[DefaultAccountIDKey] = accountID

	return err
}

func addCustomField(ctx context.Context, fields logrus.Fields, customField string) error {
	field, err := auth.GetJWTField(ctx, customField, nil)
	if err != nil {
		return fmt.Errorf("Unable to get custom %q field from context", customField)
	}

	// In case of subject field is a map
	if customField == DefaultSubjectKey {

		replacer := strings.NewReplacer("map[", "", "]", "")
		field = replacer.Replace(field)
		inner := strings.Split(field, " ")

		m := map[string]interface{}{}

		for _, v := range inner {
			kv := strings.Split(v, ":")

			if len(kv) == 1 {
				fields[customField] = kv[0]

				return err
			}

			m[kv[0]] = kv[1]
		}

		fields[customField] = m

		return err
	}

	fields[customField] = field

	return err
}

func addHeaderField(ctx context.Context, fields logrus.Fields, header string) error {
	field, ok := gateway.Header(ctx, header)
	if !ok {
		return fmt.Errorf("Unable to get custom header %q from context", header)
	}

	fields[strings.ToLower(header)] = field

	return nil
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
