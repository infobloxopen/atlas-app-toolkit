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

const (
	fieldPresenceMetaKey = "field-paths"
	pathsSeparator       = "$"
	bulkField            = "objects"
)

// NewPresenceAnnotator will parse the JSON input and then add the paths to the
// metadata to be pulled from context later
func NewPresenceAnnotator(methods ...string) func(context.Context, *http.Request) metadata.MD {
	return func(ctx context.Context, req *http.Request) metadata.MD {
		if req == nil {
			return nil
		}

		if !isValidMethod(req, methods...) {
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

		var root interface{}
		if err := json.Unmarshal(body, &root); err != nil {
			return nil
		}

		roots := getRoots(root)
		for _, r := range roots {
			queue := []pathItem{{node: r}}

			paths := []string{}
			for len(queue) > 0 {
				// dequeue an item
				item := queue[0]
				queue = queue[1:]

				if isLeaf(item) {
					paths = append(paths, strings.Join(item.path, "."))
				} else {
					if m, ok := item.node.(map[string]interface{}); ok {
						// if the item is an object, then enqueue all of its children
						for k, v := range m {
							newPath := extendPath(item.path, k, v)
							queue = append(queue, pathItem{path: newPath, node: v})
						}
					}
				}
			}

			entry := strings.Join(paths, pathsSeparator)
			if len(entry) == 0 {
				continue
			}

			md[fieldPresenceMetaKey] = append(md[fieldPresenceMetaKey], strings.Join(paths, pathsSeparator))
		}

		return md
	}
}

func extendPath(parrent []string, key string, value interface{}) []string {
	newPath := make([]string, len(parrent)+1)
	copy(newPath, parrent)
	newPath[len(newPath)-1] = generator.CamelCase(key)
	return newPath
}

func isLeaf(item pathItem) bool {
	if m, ok := item.node.(map[string]interface{}); ok {
		if len(m) == 0 && len(item.path) > 0 {
			return true
		}
	} else if len(item.path) > 0 {
		return true
	}

	return false
}

func getRoots(root interface{}) []interface{} {
	defaultRoot := []interface{}{root}
	m, ok := root.(map[string]interface{})
	if !ok {
		return defaultRoot
	}

	bulk, ok := m[bulkField]
	if !ok {
		return defaultRoot
	}

	slice, ok := bulk.([]interface{})
	if !ok {
		return defaultRoot
	}

	return slice
}

func isValidMethod(req *http.Request, methods ...string) bool {
	for _, m := range methods {
		if req.Method == m {
			return true
		}
	}

	return false
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

		fieldMask := fieldMaskFromPaths(paths)
		// If a field with type *FieldMask or []*FieldMask exists, set the paths in it
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

func fieldMaskFromPaths(paths []string) interface{} {
	if len(paths) == 0 {
		return &field_mask.FieldMask{}
	}

	if len(paths) > 1 {
		bulkFieldMasks := make([]*field_mask.FieldMask, len(paths))
		for i, p := range paths {
			bulkFieldMasks[i] = &field_mask.FieldMask{Paths: strings.Split(p, pathsSeparator)}
		}
		return bulkFieldMasks
	}

	return &field_mask.FieldMask{Paths: strings.Split(paths[0], pathsSeparator)}
}
