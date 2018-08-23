package errors

import (
	"reflect"
	"testing"

	"google.golang.org/grpc/codes"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/errdetails"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/errfields"
)

var UnexpectedValue = "Unexpected %q value: expected %v, got %v"

func checkContainer(t *testing.T, c *Container, e *Container) {
	if c.errCode != e.errCode {
		t.Errorf(UnexpectedValue, "errCode", e.errCode, c.errCode)
	}

	if c.errMessage != e.errMessage {
		t.Errorf(UnexpectedValue, "errMessage", e.errMessage, c.errMessage)
	}

	if c.errSet != e.errSet {
		t.Errorf(UnexpectedValue, "errSet", e.errSet, c.errSet)
	}

	if !reflect.DeepEqual(c.details, e.details) {
		t.Errorf(UnexpectedValue, "details", e.details, c.details)
	}

	if !reflect.DeepEqual(c.fields, e.fields) {
		t.Errorf(UnexpectedValue, "fields", e.details, c.details)
	}
}

func TestInitContainer(t *testing.T) {
	c := InitContainer()

	checkContainer(t, c, &Container{
		details:    nil,
		fields:     nil,
		errCode:    codes.Unknown,
		errMessage: "Unknown",
		errSet:     false,
	})
}

func TestError(t *testing.T) {
	c := InitContainer()

	if c.Error() != "Unknown" {
		t.Errorf("Unexpected error: expected %q, got %q", "Unknown", c.Error())
	}

	c.New(codes.InvalidArgument, "Custom error %v", "message")
	expected := "Custom error message"
	if c.Error() != expected {
		t.Errorf("Unexpected error: expected %q, got %q", expected, c.Error())
	}
}

func TestSet(t *testing.T) {
	c := InitContainer()

	// Check IfSet returning nil.

	if res := c.IfSet(codes.InvalidArgument, "<general error>"); res != nil {
		t.Errorf("Invalid IfSet result: expected nil, got %+v", res)
	}

	// Check that error details are appended and errCode and errMessage
	// are initialized accordingly.

	c = c.Set("test:object", codes.InvalidArgument, "<invalid arg error %d>", 1)

	checkContainer(t, c, &Container{
		details: []*errdetails.TargetInfo{
			errdetails.Newf(codes.InvalidArgument, "test:object", "<invalid arg error %d>", 1)},
		fields:     nil,
		errCode:    codes.InvalidArgument,
		errMessage: "<invalid arg error 1>",
		errSet:     true,
	})

	// Check IsSet returning errSet

	if c.IsSet() != true {
		t.Errorf("Invalid IsSet result: expected true, got false")
	}

	// Check that new errdetails are appended and errCode and errMessage
	// are overwritten successfully.

	c = c.Set("test:object2", codes.AlreadyExists, "<already exists error>")

	checkContainer(t, c, &Container{
		details: []*errdetails.TargetInfo{
			errdetails.Newf(codes.InvalidArgument, "test:object", "<invalid arg error %d>", 1),
			errdetails.Newf(codes.AlreadyExists, "test:object2", "<already exists error>")},
		fields:     nil,
		errCode:    codes.AlreadyExists,
		errMessage: "<already exists error>",
		errSet:     true,
	})

	// Check that ifset overwrites errCode and errMessage.

	checkContainer(t, c.IfSet(codes.InvalidArgument, "<general error>").(*Container), &Container{
		details: []*errdetails.TargetInfo{
			errdetails.Newf(codes.InvalidArgument, "test:object", "<invalid arg error %d>", 1),
			errdetails.Newf(codes.AlreadyExists, "test:object2", "<already exists error>")},
		fields:     nil,
		errCode:    codes.InvalidArgument,
		errMessage: "<general error>",
		errSet:     true,
	})

}

func TestNewContainer(t *testing.T) {
	c := NewContainer(codes.InvalidArgument, "<invalid arg error %d>", 1)

	checkContainer(t, c, &Container{
		details:    nil,
		fields:     nil,
		errCode:    codes.InvalidArgument,
		errMessage: "<invalid arg error 1>",
		errSet:     false,
	})
}

func TestNew(t *testing.T) {
	c := InitContainer()

	// Check IfSet returning nil.

	if res := c.IfSet(codes.InvalidArgument, "<general error>"); res != nil {
		t.Errorf("Invalid IfSet result: expected nil, got %+v", res)
	}

	// Check that errCode and errMessage are initialized accordingly.

	c = c.New(codes.InvalidArgument, "<invalid arg error %d>", 1)

	checkContainer(t, c, &Container{
		details:    nil,
		fields:     nil,
		errCode:    codes.InvalidArgument,
		errMessage: "<invalid arg error 1>",
		errSet:     false,
	})

	// Check IsSet returning errSet

	if c.IsSet() != false {
		t.Errorf("Invalid IsSet result: expected false, got true")
	}

	// Check that errCode and errMessage are overwritten accordingly.

	c = c.New(codes.InvalidArgument, "<invalid arg error %d>", 2)

	checkContainer(t, c, &Container{
		details:    nil,
		fields:     nil,
		errCode:    codes.InvalidArgument,
		errMessage: "<invalid arg error 2>",
		errSet:     false,
	})

	// Check that errCode and errMessage are not overwritten.

	if res := c.IfSet(codes.InvalidArgument, "<general error>"); res != nil {
		t.Errorf("Invalid IfSet result: expected nil, got %+v", res)
	}
}

func TestWithDetail(t *testing.T) {
	c := InitContainer()

	// Check WithDetail method.

	c = c.New(codes.InvalidArgument, "<general arg err %d>", 1).
		WithDetail(codes.InvalidArgument, "target", "<specific invalid arg err %d>", 1)

	checkContainer(t, c, &Container{
		details: []*errdetails.TargetInfo{
			errdetails.Newf(codes.InvalidArgument, "target", "<specific invalid arg err %d>", 1)},
		fields:     nil,
		errCode:    codes.InvalidArgument,
		errMessage: "<general arg err 1>",
		errSet:     true,
	})

	// Append one more item to details.

	c.WithDetail(codes.InvalidArgument, "target2", "<specific invalid arg err %d>", 2)

	checkContainer(t, c, &Container{
		details: []*errdetails.TargetInfo{
			errdetails.Newf(codes.InvalidArgument, "target", "<specific invalid arg err %d>", 1),
			errdetails.Newf(codes.InvalidArgument, "target2", "<specific invalid arg err %d>", 2)},
		fields:     nil,
		errCode:    codes.InvalidArgument,
		errMessage: "<general arg err 1>",
		errSet:     true,
	})

}

func TestWithDetails(t *testing.T) {
	c := InitContainer().New(codes.InvalidArgument, "<general arg err 1>")

	// Append empty details list.

	c.WithDetails()

	checkContainer(t, c, &Container{
		details:    nil,
		fields:     nil,
		errCode:    codes.InvalidArgument,
		errMessage: "<general arg err 1>",
		errSet:     false,
	})

	// Check IfSet returning nil.

	if res := c.IfSet(codes.InvalidArgument, "<general error>"); res != nil {
		t.Errorf("Invalid IfSet result: expected nil, got %+v", res)
	}

	// Append multiple details.

	c.WithDetails(
		errdetails.Newf(codes.AlreadyExists, "target3", "<already exists err>"),
		errdetails.Newf(codes.AlreadyExists, "target4", "<already exists err 2>"),
	)

	checkContainer(t, c, &Container{
		details: []*errdetails.TargetInfo{
			errdetails.Newf(codes.AlreadyExists, "target3", "<already exists err>"),
			errdetails.Newf(codes.AlreadyExists, "target4", "<already exists err 2>")},
		fields:     nil,
		errCode:    codes.InvalidArgument,
		errMessage: "<general arg err 1>",
		errSet:     true,
	})

	// Check that ifset overwrites errCode and errMessage.

	checkContainer(t, c.IfSet(codes.InvalidArgument, "<general error>").(*Container), &Container{
		details: []*errdetails.TargetInfo{
			errdetails.Newf(codes.AlreadyExists, "target3", "<already exists err>"),
			errdetails.Newf(codes.AlreadyExists, "target4", "<already exists err 2>")},
		fields:     nil,
		errCode:    codes.InvalidArgument,
		errMessage: "<general error>",
		errSet:     true,
	})

}

func TestWithField(t *testing.T) {
	c := InitContainer()

	// Check append field.

	c = c.WithField("field_x", "field error description")

	checkContainer(t, c, &Container{
		details: nil,
		fields: &errfields.FieldInfo{
			Fields: map[string]*errfields.StringListValue{
				"field_x": &errfields.StringListValue{
					Values: []string{"field error description"},
				},
			},
		},
		errCode:    codes.Unknown,
		errMessage: "Unknown",
		errSet:     true,
	})

	// Check IsSet returning errSet

	if c.IsSet() != true {
		t.Errorf("Invalid IsSet result: expected true, got false")
	}

	// Check append second description to field and another field.

	c = c.WithField("field_x", "field error description %d", 2).WithField(
		"field_y", "field error 2")

	checkContainer(t, c, &Container{
		details: nil,
		fields: &errfields.FieldInfo{
			Fields: map[string]*errfields.StringListValue{
				"field_x": &errfields.StringListValue{
					Values: []string{
						"field error description",
						"field error description 2",
					},
				},
				"field_y": &errfields.StringListValue{
					Values: []string{
						"field error 2",
					},
				},
			},
		},
		errCode:    codes.Unknown,
		errMessage: "Unknown",
		errSet:     true,
	})

	if c.IsSet() != true {
		t.Errorf("Invalid IsSet result: expected true, got false")
	}
}

func TestWithFields(t *testing.T) {
	c := InitContainer()

	// Check append empty fields.

	for _, v := range []map[string][]string{
		map[string][]string{},
		map[string][]string{"field_x": []string{}},
		map[string][]string{"": []string{"ho"}},
	} {
		c = c.WithFields(v)

		checkContainer(t, c, &Container{
			details:    nil,
			fields:     nil,
			errCode:    codes.Unknown,
			errMessage: "Unknown",
			errSet:     false,
		})

		// Check IsSet returning errSet

		if c.IsSet() != false {
			t.Errorf("Invalid IsSet result: expected false, got true")
		}

	}

	// Check WithFields append one field.

	c.WithFields(map[string][]string{
		"field_x": []string{"field error description"},
	})

	checkContainer(t, c, &Container{
		details: nil,
		fields: &errfields.FieldInfo{
			Fields: map[string]*errfields.StringListValue{
				"field_x": &errfields.StringListValue{
					Values: []string{"field error description"},
				},
			},
		},
		errCode:    codes.Unknown,
		errMessage: "Unknown",
		errSet:     true,
	})

	// Check WithFields append multiple fields.

	c.WithFields(map[string][]string{
		"field_x": []string{"field error description 2"},
		"field_y": []string{"field error 2"},
	})

	checkContainer(t, c, &Container{
		details: nil,
		fields: &errfields.FieldInfo{
			Fields: map[string]*errfields.StringListValue{
				"field_x": &errfields.StringListValue{
					Values: []string{
						"field error description",
						"field error description 2",
					},
				},
				"field_y": &errfields.StringListValue{
					Values: []string{
						"field error 2",
					},
				},
			},
		},
		errCode:    codes.Unknown,
		errMessage: "Unknown",
		errSet:     true,
	})

	if c.IsSet() != true {
		t.Errorf("Invalid IsSet result: expected true, got false")
	}
}

func TestErrorReturn(t *testing.T) {
	checkContainer(t, testErrFunc().(*Container), &Container{
		details: []*errdetails.TargetInfo{
			errdetails.Newf(codes.InvalidArgument, "target", "<specific invalid arg err 1>")},
		fields: &errfields.FieldInfo{
			Fields: map[string]*errfields.StringListValue{
				"foo": &errfields.StringListValue{
					Values: []string{"bar"},
				},
			},
		},
		errCode:    codes.InvalidArgument,
		errMessage: "<general error>",
		errSet:     true,
	})
}

func testErrFunc() error {
	return NewContainer(codes.InvalidArgument, "<general error>").
		WithDetail(codes.InvalidArgument, "target", "<specific invalid arg err %d>", 1).WithField("foo", "bar")
}

func TestGRPCStatus(t *testing.T) {
	// FIXME
}
