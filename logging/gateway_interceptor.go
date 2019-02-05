package logging

import (
	"context"
	"path"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/infobloxopen/atlas-app-toolkit/auth"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/requestid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type gwLogCfg struct {
	dynamicLogLvl bool
	noRequestID   bool
	acctIDKeyfunc jwt.Keyfunc
	withAcctID    bool
}

// GWLogOption is a type of function that alters a gwLogCfg in the instantiation
// of a GatewayLoggingInterceptor
type GWLogOption func(*gwLogCfg)

// DisableRequestID disables request-id inclusion (and generation if needed) in gw interceptor logs
func DisableRequestID(o *gwLogCfg) {
	o.noRequestID = true
}

// WithDynamicLogLevel enables or disables dynamic log levels like handled in
// the server interceptor
func WithDynamicLogLevel(enable bool) GWLogOption {
	return func(o *gwLogCfg) {
		o.dynamicLogLvl = enable
	}
}

// EnableDynamicLogLevel is a shorthand for WithDynamicLogLevel(true)
func EnableDynamicLogLevel(o *gwLogCfg) {
	o.dynamicLogLvl = true
}

// WithAccountID enables the account_id field in gw interceptor logs, like the
// server interceptor
func WithAccountID(keyfunc jwt.Keyfunc) GWLogOption {
	return func(o *gwLogCfg) {
		o.withAcctID = true
		o.acctIDKeyfunc = keyfunc
	}
}

// EnableAccountID is a shorthand for WithAccountID(nil)
func EnableAccountID(o *gwLogCfg) {
	o.withAcctID = true
	o.acctIDKeyfunc = nil
}

type sentinelKeyType int

const sentinelKey = sentinelKeyType(0)

// GatewayLoggingInterceptor handles the functions of the various toolkit interceptors
// offered for the grpc server, as well as the standard grpc_logrus server interceptor
// behavior (superset of grpc_logrus client interceptor behavior)
func GatewayLoggingInterceptor(logger *logrus.Logger, opts ...GWLogOption) grpc.UnaryClientInterceptor {
	cfg := &gwLogCfg{}
	for _, opt := range opts {
		opt(cfg)
	}
	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {

		service := path.Dir(method)[1:]
		grpcMethod := path.Base(method)
		startTime := time.Now()
		fields := logrus.Fields{
			grpc_logrus.SystemField: "grpc",
			grpc_logrus.KindField:   "gateway",
			"grpc.service":          service,
			"grpc.method":           grpcMethod,
			"grpc.start_time":       startTime.Format(time.RFC3339),
		}
		if d, ok := ctx.Deadline(); ok {
			fields["grpc.request.deadline"] = d.Format(time.RFC3339)
		}

		// Request ID -- defaults to on
		if !cfg.noRequestID {
			reqID, exists := requestid.FromContext(ctx)
			if !exists || reqID == "" {
				reqID = uuid.New().String()
			}
			fields[requestid.DefaultRequestIDKey] = reqID
			ctx = metadata.AppendToOutgoingContext(ctx, requestid.DefaultRequestIDKey, reqID)
		}

		// Custom log level
		lvl := logger.Level
		if cfg.dynamicLogLvl {
			if logFlag, ok := gateway.Header(ctx, logFlagMetaKey); ok {
				fields[logFlagFieldName] = logFlag[0]
			}
			if logLvl, ok := gateway.Header(ctx, logLevelMetaKey); ok {
				lvl, err = logrus.ParseLevel(logLvl)
				if err != nil {
					lvl = logger.Level
				}
			}
		}

		// Account ID retrieval -- ever so slightly hacky
		if cfg.withAcctID {
			md, _ := metadata.FromOutgoingContext(ctx)
			if accountID, err := auth.GetAccountID(metadata.NewIncomingContext(ctx, md), cfg.acctIDKeyfunc); err == nil {
				fields[auth.MultiTenancyField] = accountID
			} else {
				logger.Error(err)
			}
		}

		// inject logger into context (not done by normal grpc_logrus client interceptor)
		newLogger := CopyLoggerWithLevel(logger, lvl)
		newCtx := ctxlogrus.ToContext(ctx, newLogger.WithFields(fields))

		var sentinelValue bool
		err = invoker(context.WithValue(ctx, sentinelKey, &sentinelValue), method, req, reply, cc, opts...)

		// if the sentinel is set, no middlewares had errors, and it is assumed the
		// server will log the call instead of the gateway doing so
		if sentinelValue {
			return
		}

		// catch any changes made down the middleware chain by re-extracting
		resLogger := ctxlogrus.Extract(newCtx)

		durField, durVal := grpc_logrus.DurationToTimeMillisField(time.Now().Sub(startTime))
		fields = logrus.Fields{
			durField:    durVal,
			"grpc.code": grpc.Code(err).String(),
		}
		// set error message field
		if err != nil {
			fields[logrus.ErrorKey] = err
		}

		// print log message with all fields
		resLogger = resLogger.WithFields(fields)
		resLogger.Info("finished client unary call with code " + grpc.Code(err).String())

		return
	}
}

// GatewayLoggingSentinelInterceptor is meant to be the last interceptor in the
// client interceptor chain, it sets a value left in the context by the
// GatewayLoggingInterceptor so that it knows whether the call makes it to the
// server, and thus the server will log the call, and the gateway doesn't need to.
func GatewayLoggingSentinelInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		succeeded, ok := ctx.Value(sentinelKey).(*bool)
		if ok {
			*succeeded = true
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
