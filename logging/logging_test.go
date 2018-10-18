package logging

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

func TestAnnotator(t *testing.T) {
	for input, expect := range map[*map[string]string]metadata.MD{
		&map[string]string{}:                                                         metadata.MD{},
		&map[string]string{logLevelHeaderKey: "info"}:                                metadata.MD{logLevelMetaKey: []string{"info"}},
		&map[string]string{"unused": "info"}:                                         metadata.MD{},
		&map[string]string{logFlagHeaderKey: "unique-id"}:                            metadata.MD{logFlagMetaKey: []string{"unique-id"}},
		&map[string]string{logLevelHeaderKey: "info", logFlagHeaderKey: "unique-id"}: metadata.MD{logFlagMetaKey: []string{"unique-id"}, logLevelMetaKey: []string{"info"}},
	} {
		postReq := &http.Request{
			Method: "POST",
			Header: make(http.Header),
		}
		for k, v := range *input {
			postReq.Header.Add(k, v)
		}
		md := Annotator(context.Background(), postReq)
		if expect == nil && md != nil {
			t.Error("Did not produce expected nil metadata")
			continue
		}
		if !reflect.DeepEqual(md, expect) {
			t.Errorf("Did not produce expected metadata %+v, got %+v", expect, md)
		}

	}
}

func TestInterceptor(t *testing.T) {

	expectInHandler := func(addFields logrus.Fields, expect *logrus.Entry) func(ctx context.Context, req interface{}) (interface{}, error) {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			logger := ctxlogrus.Extract(ctx)
			if expect.Logger.Level != logger.Logger.Level {
				t.Errorf("Expected level %q != Observed level %q", expect.Logger.Level, logger.Logger.Level)
			}
			if !reflect.DeepEqual(expect.Data, logger.Data) {
				t.Errorf("Expected fields %+v != Observed fields %+v", expect.Data, logger.Data)
			}
			ctxlogrus.AddFields(ctx, addFields)
			return nil, nil
		}
	}
	for i, tc := range []struct {
		origLevel    logrus.Level
		startFields  logrus.Fields
		extraFields  logrus.Fields
		defaultLevel logrus.Level
		expectLevel  logrus.Level
		mdLevel      string
		mdTag        string
	}{
		{
			origLevel:    logrus.DebugLevel,
			startFields:  logrus.Fields{"source": "testing"},
			defaultLevel: logrus.InfoLevel,
			expectLevel:  logrus.WarnLevel,
			mdLevel:      "warning",
		},
		{
			origLevel:    logrus.DebugLevel,
			startFields:  logrus.Fields{"source": "testing"},
			defaultLevel: logrus.InfoLevel,
			expectLevel:  logrus.InfoLevel,
			mdLevel:      "invalid",
		},
		{
			origLevel:    logrus.WarnLevel,
			startFields:  logrus.Fields{"source": "testing"},
			extraFields:  logrus.Fields{"post-interceptor": "backpropagated?"},
			defaultLevel: logrus.WarnLevel,
			expectLevel:  logrus.InfoLevel,
			mdLevel:      "info",
		},
		{
			origLevel:    logrus.WarnLevel,
			startFields:  logrus.Fields{"source": "testing"},
			extraFields:  logrus.Fields{"post-interceptor": "backpropagated?"},
			defaultLevel: logrus.WarnLevel,
			expectLevel:  logrus.WarnLevel,
			mdTag:        "special value",
		},
		{
			origLevel:    logrus.WarnLevel,
			startFields:  logrus.Fields{"source": "testing"},
			extraFields:  logrus.Fields{"post-interceptor": "backpropagated?"},
			defaultLevel: logrus.WarnLevel,
			expectLevel:  logrus.InfoLevel,
			mdLevel:      "info",
			mdTag:        "special value",
		},
	} {
		t.Run(fmt.Sprintf("Per-request logging test - %d", i), func(t *testing.T) {

			ctx := context.Background()
			md := metadata.Pairs(logLevelMetaKey, tc.mdLevel)
			if tc.mdTag != "" {
				md = metadata.Join(md, metadata.Pairs(logFlagMetaKey, tc.mdTag))
			}
			newCtx := metadata.NewIncomingContext(ctx, md)
			newCtx = ctxlogrus.ToContext(newCtx, (&logrus.Logger{Level: tc.origLevel, Out: ioutil.Discard, Formatter: &logrus.JSONFormatter{}}).WithFields(tc.startFields))
			// if the mdTag exists we expect the interceptor to see that field, too
			if tc.mdTag != "" {
				tc.startFields[logFlagFieldName] = tc.mdTag
			}
			_, err := LogLevelInterceptor(tc.defaultLevel)(newCtx, nil, nil,
				expectInHandler(tc.extraFields, (&logrus.Logger{Level: tc.expectLevel}).WithFields(tc.startFields)))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			resLogger := ctxlogrus.Extract(newCtx)
			if resLogger.Logger.Level != tc.origLevel {
				t.Errorf("Expected level %q != Observed level %q", resLogger.Logger.Level, tc.origLevel)
			}

			// The context logger should retain fields added before and after the LogLevelInterceptor
			var expFields = make(logrus.Fields)
			for k, v := range tc.startFields {
				expFields[k] = v
			}
			for k, v := range tc.extraFields {
				expFields[k] = v
			}

			if !reflect.DeepEqual(resLogger.Data, expFields) {
				t.Errorf("Expected fields %+v != Observed fields %+v", expFields, resLogger.Data)
			}
		})
	}
}
