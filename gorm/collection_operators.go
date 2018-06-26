package gorm

import (
	"context"
	"strings"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/query"
	"github.com/jinzhu/gorm"
)

func ApplyCollectionOperators(db *gorm.DB, ctx context.Context, obj interface{}) (*gorm.DB, error) {
	// ApplyCollectionOperators applies query operators taken from context ctx to gorm instance db.
	f, err := gateway.Filtering(ctx)
	if err != nil {
		return nil, err
	}
	db, err = ApplyFiltering(db, f, obj)
	if err != nil {
		return nil, err
	}

	var s *query.Sorting
	s, err = gateway.Sorting(ctx)
	if err != nil {
		return nil, err
	}
	db, err = ApplySorting(db, s, obj)
	if err != nil {
		return nil, err
	}

	var p *query.Pagination
	p, err = gateway.Pagination(ctx)
	if err != nil {
		return nil, err
	}
	db = ApplyPagination(db, p)

	fs := gateway.FieldSelection(ctx)
	if err != nil {
		return nil, err
	}
	db = ApplyFieldSelection(db, fs, obj)

	return db, nil
}

// ApplyFiltering applies filtering operator f to gorm instance db.
func ApplyFiltering(db *gorm.DB, f *query.Filtering, obj interface{}) (*gorm.DB, error) {
	str, args, err := FilteringToGorm(f, obj)
	if err != nil {
		return nil, err
	}
	if str != "" {
		return db.Where(str, args...), nil
	}
	return db, nil
}

// ApplySorting applies sorting operator s to gorm instance db.
func ApplySorting(db *gorm.DB, s *query.Sorting, obj interface{}) (*gorm.DB, error) {
	var crs []string
	for _, cr := range s.GetCriterias() {
		dbName, err := HandleFieldPath(strings.Split(cr.GetTag(), "."), obj)
		if err != nil {
			return nil, err
		}
		if cr.IsDesc() {
			crs = append(crs, dbName+" desc")
		} else {
			crs = append(crs, dbName)
		}
	}
	if len(crs) > 0 {
		return db.Order(strings.Join(crs, ",")), nil
	}
	return db, nil
}

// ApplyPagination applies pagination operator p to gorm instance db.
func ApplyPagination(db *gorm.DB, p *query.Pagination) *gorm.DB {
	return db.Offset(p.GetOffset()).Limit(p.DefaultLimit())
}

// ApplyFieldSelection applies field selection operator fs to gorm instance db.
func ApplyFieldSelection(db *gorm.DB, fs *query.FieldSelection, obj interface{}) *gorm.DB {
	var fields []string
	for _, f := range fs.GetFields() {
		fields = append(fields, f.GetName())
	}
	if len(fields) > 0 {
		return db.Select(fields)
	}
	return db
}
