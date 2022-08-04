package gorm

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/gorm"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

type SortingCriteriaConverter interface {
	SortingCriteriaToGorm(ctx context.Context, cr *query.SortCriteria, obj interface{}) (string, string, error)
}

type FieldSelectionConverter interface {
	FieldSelectionToGorm(ctx context.Context, fs *query.FieldSelection, obj interface{}) ([]string, error)
}

type PaginationConverter interface {
	PaginationToGorm(ctx context.Context, p *query.Pagination) (offset, limit int32)
}

type CollectionOperatorsConverter interface {
	FilteringConditionConverter
	SortingCriteriaConverter
	FieldSelectionConverter
	PaginationConverter
}

func ApplyCollectionOperatorsEx(ctx context.Context, db *gorm.DB, obj interface{}, c CollectionOperatorsConverter, f *query.Filtering, s *query.Sorting, p *query.Pagination, fs *query.FieldSelection) (*gorm.DB, error) {
	db, fAssocToJoin, err := ApplyFilteringEx(ctx, db, f, obj, c)
	if err != nil {
		return nil, err
	}

	db, sAssocToJoin, err := ApplySortingEx(ctx, db, s, obj, c)
	if err != nil {
		return nil, err
	}

	if fAssocToJoin == nil && sAssocToJoin != nil {
		fAssocToJoin = make(map[string]struct{})
	}
	for k := range sAssocToJoin {
		fAssocToJoin[k] = struct{}{}
	}
	db, err = JoinAssociations(ctx, db, fAssocToJoin, obj)
	if err != nil {
		return nil, err
	}

	db = ApplyPaginationEx(ctx, db, p, c)

	db, err = ApplyFieldSelectionEx(ctx, db, fs, obj, c)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// ApplyFiltering applies filtering operator f to gorm instance db.
func ApplyFilteringEx(ctx context.Context, db *gorm.DB, f *query.Filtering, obj interface{}, c FilteringConditionConverter) (*gorm.DB, map[string]struct{}, error) {
	str, args, assocToJoin, err := FilteringToGormEx(ctx, f, obj, c)
	if err != nil {
		return nil, nil, err
	}
	if str != "" {
		return db.Where(str, args...), assocToJoin, nil
	}
	return db, nil, nil
}

// ApplySorting applies sorting operator s to gorm instance db.
func ApplySortingEx(ctx context.Context, db *gorm.DB, s *query.Sorting, obj interface{}, c SortingCriteriaConverter) (*gorm.DB, map[string]struct{}, error) {
	var crs []string
	var assocToJoin map[string]struct{}
	for _, cr := range s.GetCriterias() {
		dbCr, assoc, err := c.SortingCriteriaToGorm(ctx, cr, obj)
		if err != nil {
			return nil, nil, err
		}
		if assoc != "" {
			if assocToJoin == nil {
				assocToJoin = make(map[string]struct{})
			}
			assocToJoin[assoc] = struct{}{}
		}
		crs = append(crs, dbCr)
	}
	if len(crs) == 0 {
		return db, nil, nil
	}
	return db.Order(strings.Join(crs, ",")), assocToJoin, nil
}

// Deprecated: use ApplyCollectionOperatorsEx instead
// ApplyCollectionOperators applies collection operators to gorm instance db.
func ApplyCollectionOperators(ctx context.Context, db *gorm.DB, obj interface{}, pb proto.Message, f *query.Filtering, s *query.Sorting, p *query.Pagination, fs *query.FieldSelection) (*gorm.DB, error) {
	return ApplyCollectionOperatorsEx(ctx, db, obj, NewDefaultPbToOrmConverter(pb), f, s, p, fs)
}

// Deprecated: use ApplyFilteringEx instead
// ApplyFiltering applies filtering operator f to gorm instance db.
func ApplyFiltering(ctx context.Context, db *gorm.DB, f *query.Filtering, obj interface{}, pb proto.Message) (*gorm.DB, map[string]struct{}, error) {
	c := &DefaultFilteringConditionConverter{&DefaultFilteringConditionProcessor{pb}}
	return ApplyFilteringEx(ctx, db, f, obj, c)
}

// Deprecated: use ApplySortingEx instead
// ApplySorting applies sorting operator s to gorm instance db.
func ApplySorting(ctx context.Context, db *gorm.DB, s *query.Sorting, obj interface{}) (*gorm.DB, map[string]struct{}, error) {
	return ApplySortingEx(ctx, db, s, obj, &DefaultSortingCriteriaConverter{})
}

// JoinAssociations joins obj's associations from assoc to the current gorm query.
func JoinAssociations(ctx context.Context, db *gorm.DB, assoc map[string]struct{}, obj interface{}) (*gorm.DB, error) {
	for k := range assoc {
		tableName, sourceKeys, targetKeys, err := JoinInfo(ctx, obj, k)
		if err != nil {
			return nil, err
		}
		var keyPairs []string
		for i, k := range sourceKeys {
			keyPairs = append(keyPairs, k+" = "+targetKeys[i])
		}
		alias := gorm.ToDBName(k)
		join := fmt.Sprintf("LEFT JOIN %s %s ON %s", tableName, alias, strings.Join(keyPairs, " AND "))
		db = db.Joins(join)
	}
	return db, nil
}

// ApplyPaginationEx applies pagination operator p to gorm instance db.
func ApplyPaginationEx(ctx context.Context, db *gorm.DB, p *query.Pagination, c PaginationConverter) *gorm.DB {
	offset, limit := c.PaginationToGorm(ctx, p)

	if offset > 0 {
		db = db.Offset(offset)
	}

	if limit > 0 {
		db = db.Limit(limit)
	}

	return db
}

// ApplyPagination applies pagination operator p to gorm instance db.
func ApplyPagination(ctx context.Context, db *gorm.DB, p *query.Pagination) *gorm.DB {
	return ApplyPaginationEx(ctx, db, p, &DefaultPaginationConverter{})
}

// ApplyFieldSelectionEx applies field selection operator fs to gorm instance db.
func ApplyFieldSelectionEx(ctx context.Context, db *gorm.DB, fs *query.FieldSelection, obj interface{}, c FieldSelectionConverter) (*gorm.DB, error) {
	toPreload, err := c.FieldSelectionToGorm(ctx, fs, obj)
	if err != nil {
		return nil, err
	}
	for _, assoc := range toPreload {
		db, err = preload(db, obj, assoc)
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}

// Deprecated: use ApplyFieldSelectionEx instead
// ApplyFieldSelection applies field selection operator fs to gorm instance db.
func ApplyFieldSelection(ctx context.Context, db *gorm.DB, fs *query.FieldSelection, obj interface{}) (*gorm.DB, error) {
	return ApplyFieldSelectionEx(ctx, db, fs, obj, &DefaultFieldSelectionConverter{})
}
