package gorm

import (
	"fmt"
	"reflect"
	"strings"
)

// JoinInfo extracts the following information for assoc association of obj:
// - association table name
// - source join keys
// - target join keys
func JoinInfo(obj interface{}, assoc string) (string, []string, []string, error) {
	objType := indirectType(reflect.ValueOf(obj).Type())
	sf, ok := objType.FieldByName(assoc)
	if !ok {
		return "", nil, nil, fmt.Errorf("Cannot find field %s in %s", assoc, objType)
	}
	ok, assocKey := gormTag(&sf, "association_foreignkey")
	if !ok {
		return "", nil, nil, fmt.Errorf("association_foreignkey tag is absent in %s", objType)
	}
	assocKeys := strings.Split(assocKey, ",")
	ok, fKey := gormTag(&sf, "foreignkey")
	if !ok {
		return "", nil, nil, fmt.Errorf("foreignkey tag is absent in %s", objType)
	}
	fKeys := strings.Split(fKey, ",")

	if len(assocKeys) != len(fKeys) {
		return "", nil, nil, fmt.Errorf(`%s: the number of association keys is not equal to the number
of foreign keys in %s association`, objType, assoc)
	}

	assocType := indirectType(sf.Type)

	_, childTableName, dbAssocKeys, dbFKeys, err := parseParentChildAssoc(objType, assocType, assocKeys, fKeys)
	if err != nil {
		parentTableName, _, dbAssocKeys, dbFKeys, err := parseParentChildAssoc(assocType, objType, assocKeys, fKeys)
		if err != nil {
			return "", nil, nil, err
		}
		return parentTableName, dbFKeys, dbAssocKeys, nil
	}
	return childTableName, dbAssocKeys, dbFKeys, nil
}

func parseParentChildAssoc(parent reflect.Type, child reflect.Type, assocKeys []string, fKeys []string) (string, string, []string, []string, error) {
	parentTableName := tableName(parent)
	childTableName := tableName(child)
	var dbAssocKeys, dbFKeys []string
	for _, k := range assocKeys {
		sf, ok := parent.FieldByName(k)
		if !ok {
			return "", "", nil, nil, fmt.Errorf("Association key %s is not found in %s", k, parent)
		}
		dbAssocKeys = append(dbAssocKeys, parentTableName+"."+columnName(&sf))
	}
	for _, k := range fKeys {
		sf, ok := child.FieldByName(k)
		if !ok {
			return "", "", nil, nil, fmt.Errorf("Foreign key %s is not found in %s", k, child)
		}
		dbFKeys = append(dbFKeys, childTableName+"."+columnName(&sf))
	}
	return parentTableName, childTableName, dbAssocKeys, dbFKeys, nil
}
