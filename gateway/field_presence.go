package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/generator"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const fieldPresenceMetaKey = "field-paths"

// PresenceAnnotator will parse the JSON input and then add the paths to the
// metadata to be pulled from context later
func NewPresenceAnnotator(methods ...string) func(context.Context, *http.Request) metadata.MD {
	return func(ctx context.Context, req *http.Request) metadata.MD {
		if req == nil {
			return nil
		}
		validMethod := false
		for _, m := range methods {
			if req.Method == m {
				validMethod = true
				break
			}
		}
		if !validMethod {
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
		if len(paths) == 0 {
			return md
		}
		md[fieldPresenceMetaKey] = paths
		return md
	}
}

// pathItem stores a in-progress deconstruction of a path for a fieldmask
type pathItem struct {
	// the list of prior fields leading up to node
	path []string

	// a generic decoded json object the current item to inspect for further path extraction
	node interface{}
}

type withFields interface {
	GetFields() *field_mask.FieldMask
}

// PresenceClientInterceptor gets the interceptor for populating a fieldmask in a
// proto message from the fields given in the metadata/context
func PresenceClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			err = invoker(ctx, method, req, reply, cc, opts...)
		}()
		if req == nil {
			return
		}

		// --- If a Fields field of type *FieldMask exists, set the paths in it
		if fields, ok := req.(withFields); ok && fields.GetFields() == nil {
			paths, found := HeaderN(ctx, fieldPresenceMetaKey, -1)
			if !found {
				return
			}
			reqObj := reflect.ValueOf(req)
			if reqObj.Kind() == reflect.Ptr && !reqObj.IsNil() {
				reqObj = reqObj.Elem()
			}
			field := reqObj.FieldByName("Fields")
			//instantiate object
			field.Set(reflect.New(field.Type().Elem()))
			//unchecked assertion should be safe because of if condition above
			field.Interface().(*field_mask.FieldMask).Paths = paths
		}
		return
	}
}
