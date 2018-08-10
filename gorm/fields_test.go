package gorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Model struct {
	Property  string
	SubModel  SubModel
	SubModels []SubModel
}

type SubModel struct {
	SubProperty string
	SubSubModel SubSubModel
}

type SubSubModel struct {
	SubSubProperty string
}

func TestGormFieldSelection(t *testing.T) {
	tests := []struct {
		fs        string
		toPreload []string
		err       bool
	}{
		{
			"property",
			nil,
			false,
		},
		{
			"property,sub_model",
			[]string{"SubModel"},
			false,
		},
		{
			"property,sub_model.sub_property",
			[]string{"SubModel"},
			false,
		},
		{
			"sub_model,sub_models",
			[]string{"SubModel", "SubModels"},
			false,
		},
		{
			"sub_model,sub_models.sub_property",
			[]string{"SubModel", "SubModels"},
			false,
		},
		{
			"sub_model,sub_model.sub_sub_model.sub_sub_property",
			[]string{"SubModel.SubSubModel"},
			false,
		},
		{
			"unknown_property",
			nil,
			false,
		},
	}
	for _, test := range tests {
		toPreload, err := FieldSelectionStringToGorm(context.Background(), test.fs, &Model{})
		if test.err {
			assert.Nil(t, toPreload)
			assert.NotNil(t, err)
		} else {
			assert.Equal(t, test.toPreload, toPreload)
			assert.Nil(t, err)
		}
	}
}
