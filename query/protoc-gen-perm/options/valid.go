package options

import (
	"fmt"
	"strings"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

type FilteringOptions struct {
	AllowMissingFields bool
	Options            map[string]FilteringOption
}

type FilteringOption struct {
	DisableSorting bool
	Deny           []string
}

func ValidateFilteringPermissions(f *query.Filtering, objName string, perms map[string]map[string]FilteringOption) error {
	var getOperator func(interface{}) error
	var permData FilteringOption

	validate := func(path []string, f interface{}) error {

		switch len(path) {
		case 2:
			objPerms, ok := perms[path[0]]
			if !ok {
				return nil
			}
			permData, ok = objPerms[path[1]]
			if !ok {
				return nil
			}
		case 1:
			objPerms, ok := perms[objName]
			if !ok {
				return nil
			}
			permData, ok = objPerms[path[0]]
			if !ok {
				return nil
			}
		default:
			return fmt.Errorf("Non suported")
		}

		tp := ""

		switch x := f.(type) {
		case *query.StringCondition:
			sc := &query.Filtering_StringCondition{x}
			tp = query.StringCondition_Type_name[int32(sc.StringCondition.Type)]
		case *query.NumberCondition:
			nc := &query.Filtering_NumberCondition{x}
			tp = query.NumberCondition_Type_name[int32(nc.NumberCondition.Type)]
		default:
			return nil
		}
		for _, val := range permData.Deny {
			if val == tp {
				fullPath := strings.Join(path, ".")
				return fmt.Errorf("Operation %s does not allowed for '%s'", tp, fullPath)
			}
		}
		return nil
	}

	var vres error

	getOperator = func(f interface{}) error {
		val := f.(*query.LogicalOperator)
		var vres error
		left := val.GetLeft()
		switch leftVal := left.(type) {
		case *query.LogicalOperator_LeftOperator:
			vres = getOperator(leftVal.LeftOperator)

		case *query.LogicalOperator_LeftStringCondition:
			vres = validate(leftVal.LeftStringCondition.GetFieldPath(), leftVal.LeftStringCondition)

		case *query.LogicalOperator_LeftNumberCondition:
			vres = validate(leftVal.LeftNumberCondition.GetFieldPath(), leftVal.LeftNumberCondition)

		case *query.LogicalOperator_LeftNullCondition:
			vres = validate(leftVal.LeftNullCondition.GetFieldPath(), leftVal.LeftNullCondition)
		}

		if vres != nil {
			return vres
		}

		right := val.GetRight()
		switch rightVal := right.(type) {
		case *query.LogicalOperator_RightOperator:
			getOperator(rightVal.RightOperator)

		case *query.LogicalOperator_RightStringCondition:
			vres = validate(rightVal.RightStringCondition.GetFieldPath(), rightVal.RightStringCondition)

		case *query.LogicalOperator_RightNumberCondition:
			vres = validate(rightVal.RightNumberCondition.GetFieldPath(), rightVal.RightNumberCondition)

		case *query.LogicalOperator_RightNullCondition:
			vres = validate(rightVal.RightNullCondition.GetFieldPath(), rightVal.RightNullCondition)
		}

		return vres
	}

	if f != nil {
		root := f.GetRoot()
		switch val := root.(type) {
		case *query.Filtering_Operator:
			vres = getOperator(val.Operator)

		case *query.Filtering_StringCondition:
			vres = validate(val.StringCondition.GetFieldPath(), val.StringCondition)

		case *query.Filtering_NumberCondition:
			vres = validate(val.NumberCondition.GetFieldPath(), val.NumberCondition)

		case *query.Filtering_NullCondition:
			vres = validate(val.NullCondition.GetFieldPath(), val.NullCondition)
		}
	}
	return vres
}

func ValidateSortingPermissions(p *query.Sorting, objName string, perms map[string]map[string]FilteringOption) error {
	var res error
	var permData FilteringOption

	if p != nil {
		for _, criteria := range p.GetCriterias() {
			tag := criteria.GetTag()
			path := strings.Split(tag, ".")

			switch len(path) {
			case 2:
				objPerms, ok := perms[path[0]]
				if !ok {
					return nil
				}
				permData, ok = objPerms[path[1]]
				if !ok {
					return nil
				}
			case 1:
				objPerms, ok := perms[objName]
				if !ok {
					return nil
				}
				permData, ok = objPerms[path[0]]
				if !ok {
					return nil
				}
			default:
				return fmt.Errorf("Non suported")
			}

			if permData.DisableSorting {
				res = fmt.Errorf("pagination doesn't allowd for '%s'", tag)
				break
			}
		}
	}
	return res
}
