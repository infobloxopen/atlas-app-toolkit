package gorm

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

type LogicalOperatorConverter interface {
	LogicalOperatorToGorm(ctx context.Context, lop *query.LogicalOperator, obj interface{}) (string, []interface{}, map[string]struct{}, error)
}

type NullConditionConverter interface {
	NullConditionToGorm(ctx context.Context, c *query.NullCondition, obj interface{}) (string, []interface{}, map[string]struct{}, error)
}

type StringConditionConverter interface {
	StringConditionToGorm(ctx context.Context, c *query.StringCondition, obj interface{}) (string, []interface{}, map[string]struct{}, error)
}

type StringArrayConditionConverter interface {
	StringArrayConditionToGorm(ctx context.Context, c *query.StringArrayCondition, obj interface{}) (string, []interface{}, map[string]struct{}, error)
}

type NumberConditionConverter interface {
	NumberConditionToGorm(ctx context.Context, c *query.NumberCondition, obj interface{}) (string, []interface{}, map[string]struct{}, error)
}

type NumberArrayConditionConverter interface {
	NumberArrayConditionToGorm(ctx context.Context, c *query.NumberArrayCondition, obj interface{}) (string, []interface{}, map[string]struct{}, error)
}

type FilteringConditionConverter interface {
	LogicalOperatorConverter
	NullConditionConverter
	StringConditionConverter
	StringArrayConditionConverter
	NumberConditionConverter
	NumberArrayConditionConverter
}

type FilteringConditionProcessor interface {
	ProcessStringCondition(ctx context.Context, fieldPath []string, value string) (interface{}, error)
}

// FilterStringToGorm is a shortcut to parse a filter string using default FilteringParser implementation
// and call FilteringToGorm on the returned filtering expression.
func FilterStringToGorm(ctx context.Context, filter string, obj interface{}, pb proto.Message) (string, []interface{}, map[string]struct{}, error) {
	f, err := query.ParseFiltering(filter)
	if err != nil {
		return "", nil, nil, err
	}
	c := &DefaultFilteringConditionConverter{&DefaultFilteringConditionProcessor{pb}}
	return FilteringToGormEx(ctx, f, obj, c)
}

//Deprecated: Use FilteringToGormEx instead
// FilteringToGorm returns GORM Plain SQL representation of the filtering expression.
func FilteringToGorm(ctx context.Context, m *query.Filtering, obj interface{}, pb proto.Message) (string, []interface{}, map[string]struct{}, error) {
	c := &DefaultFilteringConditionConverter{&DefaultFilteringConditionProcessor{pb}}
	return FilteringToGormEx(ctx, m, obj, c)
}

// FilteringToGorm returns GORM Plain SQL representation of the filtering expression.
func FilteringToGormEx(ctx context.Context, m *query.Filtering, obj interface{}, c FilteringConditionConverter) (string, []interface{}, map[string]struct{}, error) {
	if m == nil || m.Root == nil {
		return "", nil, nil, nil
	}
	switch r := m.Root.(type) {
	case *query.Filtering_Operator:
		return c.LogicalOperatorToGorm(ctx, r.Operator, obj)
	case *query.Filtering_StringCondition:
		return c.StringConditionToGorm(ctx, r.StringCondition, obj)
	case *query.Filtering_NumberCondition:
		return c.NumberConditionToGorm(ctx, r.NumberCondition, obj)
	case *query.Filtering_NullCondition:
		return c.NullConditionToGorm(ctx, r.NullCondition, obj)
	case *query.Filtering_NumberArrayCondition:
		return c.NumberArrayConditionToGorm(ctx, r.NumberArrayCondition, obj)
	case *query.Filtering_StringArrayCondition:
		return c.StringArrayConditionToGorm(ctx, r.StringArrayCondition, obj)
	default:
		return "", nil, nil, fmt.Errorf("%T type is not supported in Filtering", r)
	}
}
