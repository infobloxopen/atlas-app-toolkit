package gorm

import (
	"context"

	"github.com/infobloxopen/atlas-app-toolkit/collections"
	cgorm "github.com/infobloxopen/atlas-app-toolkit/collections/gorm"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/jinzhu/gorm"
)

// ApplyCollectionOperators applies collections operators taken from context ctx to gorm instance db.
func ApplyCollectionOperators(db *gorm.DB, ctx context.Context) (*gorm.DB, error) {
	f, err := gateway.Filtering(ctx)
	if err != nil {
		return nil, err
	}
	db, err = cgorm.ApplyFiltering(db, f)
	if err != nil {
		return nil, err
	}

	var s *collections.Sorting
	s, err = gateway.Sorting(ctx)
	if err != nil {
		return nil, err
	}
	db = cgorm.ApplySorting(db, s)

	var p *collections.Pagination
	p, err = gateway.Pagination(ctx)
	if err != nil {
		return nil, err
	}
	db = cgorm.ApplyPagination(db, p)

	fs := gateway.FieldSelection(ctx)
	if err != nil {
		return nil, err
	}
	db = cgorm.ApplyFieldSelection(db, fs)

	return db, nil
}
