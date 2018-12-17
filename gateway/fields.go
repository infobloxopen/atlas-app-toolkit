package gateway

import (
	"context"
	"net/http"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

//retainFields function extracts the configuration for fields that
//need to be ratained either from gRPC response or from original testRequest
//(in case when gRPC side didn't set any preferences) and retains only
//this fields on outgoing response (dynmap).
func retainFields(ctx context.Context, req *http.Request, dynmap map[string]interface{}) {
	fieldsStr := ""
	if req != nil {
		//no fields in gprc response -> try to get from original testRequest
		vals := req.URL.Query()
		fieldsStr = vals.Get(fieldsQueryKey)
	}

	if fieldsStr == "" {
		return
	}

	fields := query.ParseFieldSelection(fieldsStr)
	if fields != nil {
		for k, result := range dynmap {
			if k != "page" {
				if results, ok := result.([]interface{}); ok {
					for _, r := range results {
						if m, ok := r.(map[string]interface{}); ok {
							doRetainFields(m, fields.Fields)
						}
					}
				} else if m, ok := result.(map[string]interface{}); ok {
					doRetainFields(m, fields.Fields)
				}
			}
		}
	}
}

func doRetainFields(obj map[string]interface{}, fields query.FieldSelectionMap) {
	if fields == nil || len(fields) == 0 {
		return
	}

	_, nnf := fields["_nnf"]
	_, nf := fields["_nf"]
	for key := range obj {
		// 如果一定排除，就直接delete
		if _, ok := fields["!"+key]; ok {
			delete(obj, key)
		}

		if _, ok := fields[key]; !ok {
			if _, isModel := obj[key].(map[string]interface{}); isModel {
				if nf { // 如果nf必须显示，此值又是model的情况下，就不删除
					continue
				}
			} else {
				if nnf { // 如果nnf必须显示，此值又不是model的情况下，就不删除
					continue
				}
			}
			delete(obj, key)
		} else {
			switch x := obj[key].(type) {
			case map[string]interface{}:
				fds := fields[key].Subs
				if fds != nil && len(fds) > 0 {
					doRetainFields(x, fds)
				}
			case []interface{}:
				for _, r := range obj[key].([]interface{}) {
					if m, ok := r.(map[string]interface{}); ok {
						fds := fields[key].Subs
						doRetainFields(m, fds)
					}
				}
			}
		}
	}
}
