package gorm

import (
	"fmt"
	"strings"

	"github.com/infobloxopen/atlas-app-toolkit/query"
	"github.com/jinzhu/gorm"
)

// ApplyCollectionOperators applies collection operators to gorm instance db.
func ApplyCollectionOperators(db *gorm.DB, obj interface{}, f *query.Filtering, s *query.Sorting, p *query.Pagination, fs *query.FieldSelection) (*gorm.DB, error) {
	db, fAssocToJoin, err := ApplyFiltering(db, f, obj)
	if err != nil {
		return nil, err
	}

	db, sAssocToJoin, err := ApplySorting(db, s, obj)
	if err != nil {
		return nil, err
	}

	if fAssocToJoin == nil && sAssocToJoin != nil {
		fAssocToJoin = make(map[string]struct{})
	}
	for k := range sAssocToJoin {
		fAssocToJoin[k] = struct{}{}
	}
	db, err = JoinAssociations(db, fAssocToJoin, obj)
	if err != nil {
		return nil, err
	}

	db = ApplyPagination(db, p)

	db, err = ApplyFieldSelection(db, fs, obj)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// ApplyFiltering applies filtering operator f to gorm instance db.
func ApplyFiltering(db *gorm.DB, f *query.Filtering, obj interface{}) (*gorm.DB, map[string]struct{}, error) {
	str, args, assocToJoin, err := FilteringToGorm(f, obj)
	if err != nil {
		return nil, nil, err
	}
	if str != "" {
		return db.Where(str, args...), assocToJoin, nil
	}
	return db, nil, nil
}

// ApplySorting applies sorting operator s to gorm instance db.
func ApplySorting(db *gorm.DB, s *query.Sorting, obj interface{}) (*gorm.DB, map[string]struct{}, error) {
	var crs []string
	var assocToJoin map[string]struct{}
	for _, cr := range s.GetCriterias() {
		dbName, assoc, err := HandleFieldPath(strings.Split(cr.GetTag(), "."), obj)
		if err != nil {
			return nil, nil, err
		}
		if assoc != "" {
			if assocToJoin == nil {
				assocToJoin = make(map[string]struct{})
			}
			assocToJoin[assoc] = struct{}{}
		}
		if cr.IsDesc() {
			crs = append(crs, dbName+" desc")
		} else {
			crs = append(crs, dbName)
		}
	}
	if len(crs) == 0 {
		return db, nil, nil
	}
	return db.Order(strings.Join(crs, ",")), assocToJoin, nil
}

// JoinAssociations joins obj's associations from assoc to the current gorm query.
func JoinAssociations(db *gorm.DB, assoc map[string]struct{}, obj interface{}) (*gorm.DB, error) {
	for k := range assoc {
		tableName, sourceKeys, targetKeys, err := JoinInfo(obj, k)
		if err != nil {
			return nil, err
		}
		var keyPairs []string
		for i, k := range sourceKeys {
			keyPairs = append(keyPairs, k+" = "+targetKeys[i])
		}
		db = db.Joins(fmt.Sprintf("LEFT JOIN %s ON %s", tableName, strings.Join(keyPairs, " AND ")))
	}
	return db, nil
}

// ApplyPagination applies pagination operator p to gorm instance db.
func ApplyPagination(db *gorm.DB, p *query.Pagination) *gorm.DB {
	if p != nil {
		return db.Offset(p.GetOffset()).Limit(p.DefaultLimit())
	}
	return db
}

// ApplyFieldSelection applies field selection operator fs to gorm instance db.
func ApplyFieldSelection(db *gorm.DB, fs *query.FieldSelection, obj interface{}) (*gorm.DB, error) {
	toPreload, err := FieldSelectionToGorm(fs, obj)
	if err != nil {
		return nil, err
	}
	for _, assoc := range toPreload {
		db = db.Preload(assoc)
	}
	return db, nil
}
