package presence

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const fieldPresenceMetaKey = "field-paths"

// Annotator will parse the JSON input and then add the paths to the
// metadata to be pulled from context later
func Annotator(ctx context.Context, req *http.Request) metadata.MD {
	if req == nil {
		return nil
	}
	if req.Method != "POST" && req.Method != "PUT" { // && req.Method != "PATCH"
		return nil
	}

	// Read body of request then reset it to be read in future
	body, err := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	if err != nil {
		return nil
	}

	paths := []string{}
	var root interface{}
	if err := json.Unmarshal(body, &root); err != nil {
		return nil
	}

	queue := []pathItem{{node: root}}
	for len(queue) > 0 {
		// dequeue an item
		item := queue[0]
		queue = queue[1:]
		if m, ok := item.node.(map[string]interface{}); ok {
			// if the item is an object, then enqueue all of its children
			for k, v := range m {
				queue = append(queue, pathItem{path: append(item.path, generator.CamelCase(k)), node: v})
			}
		} else if len(item.path) > 0 {
			// otherwise, it's a leaf node so print its path
			paths = append(paths, strings.Join(item.path, "."))
		}
	}

	md := make(metadata.MD)
	key := runtime.MetadataPrefix + fieldPresenceMetaKey
	for _, path := range paths {
		md[key] = append(md[key], path)
	}
	return md
}

// pathItem stores a in-progress deconstruction of a path for a fieldmask
type pathItem struct {
	// the list of prior fields leading up to node
	path []string

	// a generic decoded json object the current item to inspect for further path extraction
	node interface{}
}

// GetFieldsSlice pulls the paths from the context if put there via grpc metadata
func GetFieldsSlice(ctx context.Context) []string {
	paths, _ := gateway.HeaderN(ctx, fieldPresenceMetaKey, -1)
	return paths
}

// UnaryServerInterceptor gets the interceptor for populating a fieldmask in a
// proto message from the fields given in the metadata/context
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		// handle panic
		defer func() {
			if perr := recover(); perr != nil {
				err = status.Errorf(codes.Internal, "field presence interceptor: %s", perr)
				grpclog.Errorln(err)
				res = nil
			}
		}()

		if req == nil {
			grpclog.Warningf("field presence interceptor: empty request %+v", req)
			return handler(ctx, req)
		}

		// --- If a Fields field of type *FieldMask exists, set the paths in it
		var field reflect.Value
		// First: deref the request as far as possible
		for field = reflect.ValueOf(req); field.Kind() == reflect.Ptr && !field.IsNil(); field = field.Elem() {
		}
		// Second: check for the "Fields" field, only use if still Nil (no overwrite)
		if field = field.FieldByName("Fields"); field.IsValid() &&
			field.Type() == reflect.TypeOf(&field_mask.FieldMask{}) && field.IsNil() {

			//instantiate object
			field.Set(reflect.New(field.Type().Elem()))
			//unchecked assertion should be safe because of if condition above
			field.Interface().(*field_mask.FieldMask).Paths = GetFieldsSlice(ctx)
		}

		res, err = handler(ctx, req)
		if err != nil {
			return
		}
		return
	}
}
