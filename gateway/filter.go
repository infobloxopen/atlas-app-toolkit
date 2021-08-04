package gateway

import "github.com/grpc-ecosystem/grpc-gateway/v2/utilities"

// DefaultQueryFilter can be set to override the filter_{service}_{rpc}_{num}
// field in generated .pb.gw.go files to prevent parse errors in the gateway
// and potentially reduce log noise due to unrecognized fields
var DefaultQueryFilter = utilities.NewDoubleArray(defaultFilterFields)

var defaultFilterFields = [][]string{
	// collection ops and the expected names used for the collection ops objects in requests
	{"paging"}, {limitQueryKey}, {offsetQueryKey}, {pageTokenQueryKey},
	{"order_by"}, {sortQueryKey},
	{"fields"}, {fieldsQueryKey},
	{"filter"}, {filterQueryKey},
}

// QueryFilterWith will add extra fields to the standard fields in the default
// filter.
func QueryFilterWith(extraFields []string) *utilities.DoubleArray {
	qf := make([][]string, len(defaultFilterFields)+len(extraFields))
	copy(qf, defaultFilterFields)
	for _, f := range extraFields {
		qf = append(qf, []string{f})
	}
	return utilities.NewDoubleArray(qf)
}
