package requestinfo

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/metadata"
	"reflect"
	"testing"
)

func TestRequestInfoFromContext(t *testing.T) {
	testCases := []struct {
		ctx    context.Context
		result RequestInfo
	}{
		{
			ctx:    metadata.NewOutgoingContext(context.Background(), metadata.New(generateRequestInfoMap(operationTypeToName[DeleteOperation], "resource-id"))),
			result: generateRequestInfo(DeleteOperation, "resource-id"),
		},
		{
			ctx:    metadata.NewOutgoingContext(context.Background(), metadata.New(generateRequestInfoMap(operationTypeToName[CreateOperation], "resource-id"))),
			result: generateRequestInfo(CreateOperation, "resource-id"),
		},
		{
			ctx:    metadata.NewOutgoingContext(context.Background(), metadata.New(generateRequestInfoMap(operationTypeToName[CreateOperation], ""))),
			result: generateRequestInfo(CreateOperation, ""),
		},
		{
			ctx:    metadata.NewOutgoingContext(context.Background(), metadata.New(generateRequestInfoMap(operationTypeToName[UpdateOperation], "resource-id"))),
			result: generateRequestInfo(UpdateOperation, "resource-id"),
		},
		{
			ctx:    metadata.NewOutgoingContext(context.Background(), metadata.New(generateRequestInfoMap(operationTypeToName[ReplaceOperation], "resource-id"))),
			result: generateRequestInfo(ReplaceOperation, "resource-id"),
		},
		{
			ctx:    metadata.NewOutgoingContext(context.Background(), metadata.New(generateRequestInfoMap(operationTypeToName[ListOperation], ""))),
			result: generateRequestInfo(ListOperation, ""),
		},
		{
			ctx:    metadata.NewOutgoingContext(context.Background(), metadata.New(generateRequestInfoMap(operationTypeToName[ReadOperation], "resource-id"))),
			result: generateRequestInfo(ReadOperation, "resource-id"),
		},
		{
			ctx:    metadata.NewOutgoingContext(context.Background(), metadata.New(generateRequestInfoMap("Some operation", "resource-id"))),
			result: generateRequestInfo(UnknownOperation, "resource-id"),
		},
	}

	for num, test := range testCases {
		res, err := FromContext(test.ctx)
		if err != nil {
			t.Errorf("Error must be nil in test %d but got %s", num+1, err.Error())
			t.Fail()
		}

		if !reflect.DeepEqual(res, test.result) {
			t.Errorf("Output error expected %+v get %+v", test.result, res)
			t.Fail()
		}
	}
}

func TestRequestInfoFromContextNegative(t *testing.T) {
	requestInfoMap := generateRequestInfoMap(operationTypeToName[DeleteOperation], "resource-id")
	delete(requestInfoMap, runtime.MetadataPrefix+appNameMetaKey)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(requestInfoMap))

	_, err := FromContext(ctx)
	if err == nil {
		t.Errorf("Errors not equal expected %s get nil", ErrAppNameIsMissing)
		t.Fail()
	} else if err != ErrAppNameIsMissing {
		t.Errorf("Errors not equal expected %s get nil", ErrAppNameIsMissing)
		t.Fail()
	}

	requestInfoMap = generateRequestInfoMap(operationTypeToName[DeleteOperation], "resource-id")
	delete(requestInfoMap, runtime.MetadataPrefix+resourceTypeMetaKey)
	ctx = metadata.NewOutgoingContext(context.Background(), metadata.New(requestInfoMap))

	_, err = FromContext(ctx)
	if err == nil {
		t.Errorf("Errors not equal expected %s get nil", ErrResourceTypeIsMissing)
		t.Fail()
	} else if err != ErrResourceTypeIsMissing {
		t.Errorf("Errors not equal expected %s get nil", ErrResourceTypeIsMissing)
		t.Fail()
	}

}
