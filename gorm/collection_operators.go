package gorm

import (
	"context"
	"strings"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/query"
	"github.com/jinzhu/gorm"
)

// ApplyCollectionOperators applies query operators taken from context ctx to gorm instance db.
func ApplyCollectionOperators(db *gorm.DB, ctx context.Context, obj interface{}) (*gorm.DB, error) {
	f, err := gateway.Filtering(ctx)
	if err != nil {
		return nil, err
	}
	db, err = ApplyFiltering(db, f, obj)
	if err != nil {
		return nil, err
	}

	s, err := gateway.Sorting(ctx)
	if err != nil {
		return nil, err
	}
	db, err = ApplySorting(db, s, obj)
	if err != nil {
		return nil, err
	}

	p, err := gateway.Pagination(ctx)
	if err != nil {
		return nil, err
	}
	db = ApplyPagination(db, p)

	fs := gateway.FieldSelection(ctx)
	if err != nil {
		return nil, err
	}
	db, err = ApplyFieldSelection(db, fs, obj)
	if err != nil {
		return nil, err
	}

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
