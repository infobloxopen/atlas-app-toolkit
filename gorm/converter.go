package gorm

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/infobloxopen/atlas-app-toolkit/query"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
	"github.com/infobloxopen/atlas-app-toolkit/util"
)

// DefaultFilteringConditionProcessor processes filter operator conversion
type DefaultFilteringConditionProcessor struct {
	pb proto.Message
}

// DefaultFilteringConditionConverter performs default convertion for Filter collection operator
type DefaultFilteringConditionConverter struct {
	Processor FilteringConditionProcessor
}

// DefaultSortingCriteriaConverter performs default convertion for Sorting collection operator
type DefaultSortingCriteriaConverter struct{}

// DefaultPaginationConverter performs default convertion for Paging collection operator
type DefaultPaginationConverter struct{}

// DefaultPbToOrmConverter performs default convertion for all collection operators
type DefaultPbToOrmConverter struct {
	DefaultFilteringConditionConverter
	DefaultSortingCriteriaConverter
	DefaultFieldSelectionConverter
	DefaultPaginationConverter
}

// NewDefaultPbToOrmConverter creates default converter for all collection operators
func NewDefaultPbToOrmConverter(pb proto.Message) CollectionOperatorsConverter {
	return &DefaultPbToOrmConverter{
		DefaultFilteringConditionConverter{&DefaultFilteringConditionProcessor{pb}},
		DefaultSortingCriteriaConverter{},
		DefaultFieldSelectionConverter{},
		DefaultPaginationConverter{},
	}
}

// LogicalOperatorToGorm returns GORM Plain SQL representation of the logical operator.
func (converter *DefaultFilteringConditionConverter) LogicalOperatorToGorm(ctx context.Context, lop *query.LogicalOperator, obj interface{}) (string, []interface{}, map[string]struct{}, error) {
	var lres string
	var largs []interface{}
	var lAssocToJoin map[string]struct{}
	var err error
	switch l := lop.Left.(type) {
	case *query.LogicalOperator_LeftOperator:
		lres, largs, lAssocToJoin, err = converter.LogicalOperatorToGorm(ctx, l.LeftOperator, obj)
	case *query.LogicalOperator_LeftStringCondition:
		lres, largs, lAssocToJoin, err = converter.StringConditionToGorm(ctx, l.LeftStringCondition, obj)
	case *query.LogicalOperator_LeftNumberCondition:
		lres, largs, lAssocToJoin, err = converter.NumberConditionToGorm(ctx, l.LeftNumberCondition, obj)
	case *query.LogicalOperator_LeftNullCondition:
		lres, largs, lAssocToJoin, err = converter.NullConditionToGorm(ctx, l.LeftNullCondition, obj)
	case *query.LogicalOperator_LeftNumberArrayCondition:
		lres, largs, lAssocToJoin, err = converter.NumberArrayConditionToGorm(ctx, l.LeftNumberArrayCondition, obj)
	case *query.LogicalOperator_LeftStringArrayCondition:
		lres, largs, lAssocToJoin, err = converter.StringArrayConditionToGorm(ctx, l.LeftStringArrayCondition, obj)
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
		rres, rargs, rAssocToJoin, err = converter.LogicalOperatorToGorm(ctx, r.RightOperator, obj)
	case *query.LogicalOperator_RightStringCondition:
		rres, rargs, rAssocToJoin, err = converter.StringConditionToGorm(ctx, r.RightStringCondition, obj)
	case *query.LogicalOperator_RightNumberCondition:
		rres, rargs, rAssocToJoin, err = converter.NumberConditionToGorm(ctx, r.RightNumberCondition, obj)
	case *query.LogicalOperator_RightNullCondition:
		rres, rargs, rAssocToJoin, err = converter.NullConditionToGorm(ctx, r.RightNullCondition, obj)
	case *query.LogicalOperator_RightNumberArrayCondition:
		rres, rargs, rAssocToJoin, err = converter.NumberArrayConditionToGorm(ctx, r.RightNumberArrayCondition, obj)
	case *query.LogicalOperator_RightStringArrayCondition:
		rres, rargs, rAssocToJoin, err = converter.StringArrayConditionToGorm(ctx, r.RightStringArrayCondition, obj)
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
func (converter *DefaultFilteringConditionConverter) StringConditionToGorm(ctx context.Context, c *query.StringCondition, obj interface{}) (string, []interface{}, map[string]struct{}, error) {
	var (
		assocToJoin   map[string]struct{}
		dbName, assoc string
		err           error
	)

	if IsJSONCondition(ctx, c.FieldPath, obj) {
		dbName, assoc, err = HandleJSONFieldPath(ctx, c.FieldPath, obj, c.Value)
	} else {
		dbName, assoc, err = HandleFieldPath(ctx, c.FieldPath, obj)
	}
	if err != nil {
		return "", nil, nil, err
	}

	if assoc != "" {
		assocToJoin = make(map[string]struct{})
		assocToJoin[assoc] = struct{}{}
	}
	var o string
	switch c.Type {
	case query.StringCondition_EQ, query.StringCondition_IEQ:
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
	if v, err := converter.Processor.ProcessStringCondition(ctx, c.FieldPath, c.Value); err != nil {
		value = c.Value
	} else {
		value = v
	}

	if c.Type == query.StringCondition_IEQ {
		return converter.insensitiveCaseStringConditionToGorm(neg, dbName, o), []interface{}{value}, assocToJoin, nil
	}

	return fmt.Sprintf("%s(%s %s ?)", neg, dbName, o), []interface{}{value}, assocToJoin, nil
}

func (converter *DefaultFilteringConditionConverter) insensitiveCaseStringConditionToGorm(neg, dbName, operator string) string {
	return fmt.Sprintf("%s(lower(%s) %s lower(?))", neg, dbName, operator)
}

// ProcessStringCondition processes a string condition to GORM Plain SQL representation
func (p *DefaultFilteringConditionProcessor) ProcessStringCondition(ctx context.Context, fieldPath []string, value string) (interface{}, error) {
	objType := indirectType(reflect.TypeOf(p.pb))
	pathLength := len(fieldPath)
	for i, part := range fieldPath {
		sf, ok := objType.FieldByName(util.Camel(part))
		if !ok {
			return nil, fmt.Errorf("Cannot find field %s in %s", part, objType)
		}
		if i < pathLength-1 {
			objType = indirectType(sf.Type)
			if !isProtoMessage(objType) {
				return nil, fmt.Errorf("%s: non-last field of %s field path should be a proto message", objType, fieldPath)
			}
		} else {
			if isIdentifier(indirectType(sf.Type)) {
				id := &resource.Identifier{}
				if err := jsonpb.UnmarshalString(fmt.Sprintf("\"%s\"", value), id); err != nil {
					return nil, err
				}
				newPb := reflect.New(objType)
				v := newPb.Elem().FieldByName(util.Camel(part))
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
				ormId := orm.FieldByName(util.Camel(part))
				if !ormId.IsValid() {
					return nil, fmt.Errorf("Cannot find field %s in %s", part, objType)
				}
				return reflect.Indirect(ormId).Interface(), nil

			}
		}
	}
	return value, nil
}

// NumberConditionToGorm returns GORM Plain SQL representation of the number condition.
func (converter *DefaultFilteringConditionConverter) NumberConditionToGorm(ctx context.Context, c *query.NumberCondition, obj interface{}) (string, []interface{}, map[string]struct{}, error) {
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
func (converter *DefaultFilteringConditionConverter) NullConditionToGorm(ctx context.Context, c *query.NullCondition, obj interface{}) (string, []interface{}, map[string]struct{}, error) {
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

func (converter *DefaultFilteringConditionConverter) NumberArrayConditionToGorm(ctx context.Context, c *query.NumberArrayCondition, obj interface{}) (string, []interface{}, map[string]struct{}, error) {
	var assocToJoin map[string]struct{}
	dbName, assoc, err := HandleFieldPath(ctx, c.FieldPath, obj)
	if err != nil {
		return "", nil, nil, err
	}

	if assoc != "" {
		assocToJoin = make(map[string]struct{})
		assocToJoin[assoc] = struct{}{}
	}
	o := "IN"
	var neg string
	if c.IsNegative {
		neg = "NOT"
	}

	placeholder := ""
	values := make([]interface{}, 0, len(c.Values))
	for _, val := range c.Values {
		placeholder += "?, "
		values = append(values, val)
	}

	return fmt.Sprintf("(%s %s %s (%s))", dbName, neg, o, strings.TrimSuffix(placeholder, ", ")), values, assocToJoin, nil
}

func (converter *DefaultFilteringConditionConverter) StringArrayConditionToGorm(ctx context.Context, c *query.StringArrayCondition, obj interface{}) (string, []interface{}, map[string]struct{}, error) {
	var (
		assocToJoin   map[string]struct{}
		dbName, assoc string
		err           error
	)
	if IsJSONCondition(ctx, c.FieldPath, obj) {
		dbName, assoc, err = HandleJSONFieldPath(ctx, c.FieldPath, obj, c.Values...)
	} else {
		dbName, assoc, err = HandleFieldPath(ctx, c.FieldPath, obj)
	}
	if err != nil {
		return "", nil, nil, err
	}

	if assoc != "" {
		assocToJoin = make(map[string]struct{})
		assocToJoin[assoc] = struct{}{}
	}
	o := "IN"
	var neg string
	if c.IsNegative {
		neg = "NOT"
	}

	values := make([]interface{}, 0, len(c.Values))
	placeholder := ""
	for _, str := range c.Values {
		placeholder += "?, "
		if val, err := converter.Processor.ProcessStringCondition(ctx, c.FieldPath, str); err == nil {
			values = append(values, val)
			continue
		}

		values = append(values, str)
	}

	return fmt.Sprintf("(%s %s %s (%s))", dbName, neg, o, strings.TrimSuffix(placeholder, ", ")), values, assocToJoin, nil
}

func (converter *DefaultSortingCriteriaConverter) SortingCriteriaToGorm(ctx context.Context, cr *query.SortCriteria, obj interface{}) (string, string, error) {
	dbCr, assoc, err := HandleFieldPath(ctx, strings.Split(cr.GetTag(), "."), obj)
	if cr.IsDesc() {
		dbCr += " desc"
	}
	return dbCr, assoc, err
}

func (converter *DefaultPaginationConverter) PaginationToGorm(ctx context.Context, p *query.Pagination) (offset, limit int32) {
	if p != nil {
		return p.GetOffset(), p.GetLimit()
	}
	return 0, 0
}
