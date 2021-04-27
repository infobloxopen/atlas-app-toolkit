package logging

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	testFields  = []string{testCustomJWTFieldKey}
	testHeaders = []string{testCustomHeaderKey}
)

func TestNew(t *testing.T) {
	expected := &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	}

	result := New("Info")

	assert.Equal(t, logrus.Level(4), result.Level)
	assert.Equal(t, expected, result.Formatter)
}

func TestInitOptions(t *testing.T) {
	expected := &options{
		fields:  testFields,
		headers: testHeaders,
	}

	opts := []Option{
		WithCustomFields(testFields),
		WithCustomHeaders(testHeaders),
	}

	result := initOptions(opts)
	assert.Equal(t, expected.fields, result.fields)
	assert.Equal(t, expected.headers, result.headers)
}

func TestWithCustomFields(t *testing.T) {
	opt := WithCustomFields(testFields)
	result := &options{}
	opt(result)
	assert.Equal(t, testFields, result.fields)
}

func TestWithCustomHeaders(t *testing.T) {
	opt := WithCustomHeaders(testHeaders)
	result := &options{}
	opt(result)
	assert.Equal(t, testHeaders, result.headers)
}
