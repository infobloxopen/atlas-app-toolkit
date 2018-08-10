package gorm

import (
	"context"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/generator"

	"github.com/infobloxopen/atlas-app-toolkit/query"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

// FilterStringToGorm is a shortcut to parse a filter string using default FilteringParser implementation
// and call FilteringToGorm on the returned filtering expression.
func FilterStringToGorm(ctx context.Context, filter string, obj interface{}, pb proto.Message) (string, []interface{}, map[string]struct{}, error) {
	f, err := query.ParseFiltering(filter)
	if err != nil {
		return "", nil, nil, err
	}
	return FilteringToGorm(ctx, f, obj, pb)
}

// FilteringToGorm returns GORM Plain SQL representation of the filtering expression.
func FilteringToGorm(ctx context.Context, m *query.Filtering, obj interface{}, pb proto.Message) (string, []interface{}, map[string]struct{}, error) {
	if m == nil || m.Root == nil {
		return "", nil, nil, nil
	}
	switch r := m.Root.(type) {
	case *query.Filtering_Operator:
		return LogicalOperatorToGorm(ctx, r.Operator, obj, pb)
	case *query.Filtering_StringCondition:
		return StringConditionToGorm(ctx, r.StringCondition, obj, pb)
	case *query.Filtering_NumberCondition:
		return NumberConditionToGorm(ctx, r.NumberCondition, obj, pb)
	case *query.Filtering_NullCondition:
		return NullConditionToGorm(ctx, r.NullCondition, obj, pb)
	default:
		return "", nil, nil, fmt.Errorf("%T type is not supported in Filtering", r)
	}
}

// LogicalOperatorToGorm returns GORM Plain SQL representation of the logical operator.
func LogicalOperatorToGorm(ctx context.Context, lop *query.LogicalOperator, obj interface{}, pb proto.Message) (string, []interface{}, map[string]struct{}, error) {
	var lres string
	var largs []interface{}
	var lAssocToJoin map[string]struct{}
	var err error
	switch l := lop.Left.(type) {
	case *query.LogicalOperator_LeftOperator:
		lres, largs, lAssocToJoin, err = LogicalOperatorToGorm(ctx, l.LeftOperator, obj, pb)
	case *query.LogicalOperator_LeftStringCondition:
		lres, largs, lAssocToJoin, err = StringConditionToGorm(ctx, l.LeftStringCondition, obj, pb)
	case *query.LogicalOperator_LeftNumberCondition:
		lres, largs, lAssocToJoin, err = NumberConditionToGorm(ctx, l.LeftNumberCondition, obj, pb)
	case *query.LogicalOperator_LeftNullCondition:
		lres, largs, lAssocToJoin, err = NullConditionToGorm(ctx, l.LeftNullCondition, obj, pb)
	default:
		return "", nil, nil, fmt.Errorf("%T type is not supported in Filtering", l)
	}
	if err != nil {
		return "", nil, nil, err
	}

	var rres string
	var rargs []interface{}
	var rAssocToJoin map[string]struct{}
	switch r := lop.Right.(type) {
	case *query.LogicalOperator_RightOperator:
		rres, rargs, rAssocToJoin, err = LogicalOperatorToGorm(ctx, r.RightOperator, obj, pb)
	case *query.LogicalOperator_RightStringCondition:
		rres, rargs, rAssocToJoin, err = StringConditionToGorm(ctx, r.RightStringCondition, obj, pb)
	case *query.LogicalOperator_RightNumberCondition:
		rres, rargs, rAssocToJoin, err = NumberConditionToGorm(ctx, r.RightNumberCondition, obj, pb)
	case *query.LogicalOperator_RightNullCondition:
		rres, rargs, rAssocToJoin, err = NullConditionToGorm(ctx, r.RightNullCondition, obj, pb)
	default:
		return "", nil, nil, fmt.Errorf("%T type is not supported in Filtering", r)
	}
	if err != nil {
		return "", nil, nil, err
	}

	if lAssocToJoin == nil && rAssocToJoin != nil {
		lAssocToJoin = make(map[string]struct{})
	}
	for k := range rAssocToJoin {
		lAssocToJoin[k] = struct{}{}
	}

	var o string
	switch lop.Type {
	case query.LogicalOperator_AND:
		o = "AND"
	case query.LogicalOperator_OR:
		o = "OR"
	}
	var neg string
	if lop.IsNegative {
		neg = "NOT"
	}
	return fmt.Sprintf("%s(%s %s %s)", neg, lres, o, rres), append(largs, rargs...), lAssocToJoin, nil
}

// StringConditionToGorm returns GORM Plain SQL representation of the string condition.
func StringConditionToGorm(ctx context.Context, c *query.StringCondition, obj interface{}, pb proto.Message) (string, []interface{}, map[string]struct{}, error) {
	var assocToJoin map[string]struct{}
	dbName, assoc, err := HandleFieldPath(ctx, c.FieldPath, obj)
	if err != nil {
		return "", nil, nil, err
	}
	if assoc != "" {
		assocToJoin = make(map[string]struct{})
		assocToJoin[assoc] = struct{}{}
	}
	var o string
	switch c.Type {
	case query.StringCondition_EQ:
		o = "="
	case query.StringCondition_MATCH:
		o = "~"
	case query.StringCondition_GT:
		o = ">"
	case query.StringCondition_GE:
		o = ">="
	case query.StringCondition_LT:
		o = "<"
	case query.StringCondition_LE:
		o = "<="
	}
	var neg string
	if c.IsNegative {
		neg = "NOT"
	}

	var value interface{}
	if v, err := processStringCondition(ctx, c, pb); err != nil {
		value = c.Value
	} else {
		value = v
	}

	return fmt.Sprintf("%s(%s %s ?)", neg, dbName, o), []interface{}{value}, assocToJoin, nil
}

func processStringCondition(ctx context.Context, c *query.StringCondition, pb proto.Message) (interface{}, error) {
	objType := indirectType(reflect.ValueOf(pb).Type())
	pathLength := len(c.FieldPath)
	for i, part := range c.FieldPath {
		sf, ok := objType.FieldByName(generator.CamelCase(part))
		if !ok {
			return nil, fmt.Errorf("Cannot find field %s in %s", part, objType)
		}
		if i < pathLength-1 {
			objType = indirectType(sf.Type)
			if !isProtoMessage(objType) {
				return nil, fmt.Errorf("%s: non-last field of %s field path should be a proto message", objType, c.FieldPath)
			}
		} else {
			if isIdentifier(indirectType(sf.Type)) {
				id := &resource.Identifier{}
				if err := jsonpb.UnmarshalString(fmt.Sprintf("\"%s\"", c.Value), id); err != nil {
					return nil, err
				}
				newPb := reflect.New(objType)
				v := newPb.Elem().FieldByName(generator.CamelCase(part))
				v.Set(reflect.ValueOf(id))
				toOrm := newPb.MethodByName("ToORM")
				if !toOrm.IsValid() {
					return nil, fmt.Errorf("ToORM method cannot be found for %s", objType)
				}
				res := toOrm.Call([]reflect.Value{reflect.ValueOf(ctx)})
				if len(res) != 2 {
					return nil, fmt.Errorf("ToORM signature of %s is unknown", objType)
				}
				orm := res[0]
				err := res[1]
				if !err.IsNil() {
					if tErr, ok := err.Interface().(error); ok {
						return nil, tErr
					} else {
						return nil, fmt.Errorf("ToOrm second return value of %s is expected to be error", objType)
					}
				}
				ormId := orm.FieldByName(generator.CamelCase(part))
				if !ormId.IsValid() {
					return nil, fmt.Errorf("Cannot find field %s in %s", part, objType)
				}
				return reflect.Indirect(ormId).Interface(), nil

			}
		}
	}
	return c.Value, nil
}

// NumberConditionToGorm returns GORM Plain SQL representation of the number condition.
func NumberConditionToGorm(ctx context.Context, c *query.NumberCondition, obj interface{}, pb proto.Message) (string, []interface{}, map[string]struct{}, error) {
	var assocToJoin map[string]struct{}
	dbName, assoc, err := HandleFieldPath(ctx, c.FieldPath, obj)
	if err != nil {
		return "", nil, nil, err
	}
	if assoc != "" {
		assocToJoin = make(map[string]struct{})
		assocToJoin[assoc] = struct{}{}
	}
	var o string
	switch c.Type {
	case query.NumberCondition_EQ:
		o = "="
	case query.NumberCondition_GT:
		o = ">"
	case query.NumberCondition_GE:
		o = ">="
	case query.NumberCondition_LT:
		o = "<"
	case query.NumberCondition_LE:
		o = "<="
	}
	var neg string
	if c.IsNegative {
		neg = "NOT"
	}
	return fmt.Sprintf("%s(%s %s ?)", neg, dbName, o), []interface{}{c.Value}, assocToJoin, nil
}

// NullConditionToGorm returns GORM Plain SQL representation of the null condition.
func NullConditionToGorm(ctx context.Context, c *query.NullCondition, obj interface{}, pb proto.Message) (string, []interface{}, map[string]struct{}, error) {
	var assocToJoin map[string]struct{}
	dbName, assoc, err := HandleFieldPath(ctx, c.FieldPath, obj)
	if err != nil {
		return "", nil, nil, err
	}
	if assoc != "" {
		assocToJoin = make(map[string]struct{})
		assocToJoin[assoc] = struct{}{}
	}
	o := "IS NULL"
	var neg string
	if c.IsNegative {
		neg = "NOT"
	}
	return fmt.Sprintf("%s(%s %s)", neg, dbName, o), nil, assocToJoin, nil
}
