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

		md := make(metadata.MD)

		if len(body) == 0 {
			md[fieldPresenceMetaKey] = nil
			return md
		}

		var paths []string
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
				l := len(item.path)
				// if the item is an object, then enqueue all of its children
				for k, v := range m {
					newPath := make([]string, l+1)
					copy(newPath, item.path)
					newPath[l] = generator.CamelCase(k)
					queue = append(queue, pathItem{path: newPath, node: v})
				}
			} else if len(item.path) > 0 {
				// otherwise, it's a leaf node so print its path
				paths = append(paths, strings.Join(item.path, "."))
			}
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

		paths, found := HeaderN(ctx, fieldPresenceMetaKey, -1)
		if !found {
			return
		}
		fieldMask := &field_mask.FieldMask{Paths: paths}

		// If a field with type *FieldMask exists, set the paths in it
		t := reflect.ValueOf(req)
		if t.Kind() != reflect.Interface && t.Kind() != reflect.Ptr {
			return
		}
		t = t.Elem()
		if t.Kind() != reflect.Struct { // only Structs can have their fields enumerated
			return
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Type() == reflect.TypeOf(fieldMask) && f.IsNil() {
				f.Set(reflect.ValueOf(fieldMask))
			}
		}
		return
	}
}
