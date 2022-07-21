package gorm

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Model struct {
	Property         string
	SubModel         SubModel
	SubModels        []SubModel
	CycleModel       *CycleModel
	NotPreloadObj    SubModel `gorm:"preload:false"`
	PreloadObj       SubModel `gorm:"preload:true"`
	NonCamelMODEL    NonCamelMODEL
	NonCAMEL2Model   NonCAMEL2Model
	NonCamelSUBMODEL NonCamelSUBMODEL
}

type NonCamelMODEL struct {
	NonCamelProperty string
	NonCamelSUBMODEL NonCamelSUBMODEL
}

type NonCamelSUBMODEL struct {
	NonCamelSubProperty string
}

type NonCAMEL2Model struct {
	NonCamelProperty string
	Model            *Model
}

type CycleModel struct {
	Property string
	Model    *Model
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
			"non_camel_model.non_camel_submodel.noncamelsubproperty",
			[]string{"NonCamelMODEL.NonCamelSUBMODEL", "NonCamelMODEL"},
			false,
		},
		{
			"non_camel_model.non_camel_submodel",
			[]string{"NonCamelMODEL.NonCamelSUBMODEL", "NonCamelMODEL"},
			false,
		},
		{
			"non_camel_model",
			[]string{"NonCamelMODEL"},
			false,
		},
		{
			"non_camel_model.noncamelproperty",
			[]string{"NonCamelMODEL"},
			false,
		},
		{
			"non_CAMEL_2_Model, Non_camel_2_model, non_camel2_model, non_camel_2model",
			[]string{"NonCAMEL2Model"},
			false,
		},
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
			[]string{"SubModel.SubSubModel", "SubModel"},
			false,
		},
		{
			"unknown_property",
			nil,
			false,
		},
		{
			"not_preload_obj,preload_obj",
			[]string{"NotPreloadObj", "PreloadObj"},
			false,
		},
		{
			"",
			[]string{"SubModel.SubSubModel", "SubModel", "SubModels.SubSubModel", "SubModels",
				"CycleModel", "PreloadObj.SubSubModel", "PreloadObj",
				"NonCamelMODEL.NonCamelSUBMODEL", "NonCamelMODEL", "NonCAMEL2Model", "NonCamelSUBMODEL"},
			false,
		},
	}
	for _, test := range tests {
		toPreload, err := FieldSelectionStringToGorm(context.Background(), test.fs, &Model{})
		if test.err {
			assert.Nil(t, toPreload)
			assert.NotNil(t, err)
		} else {
			fmt.Printf("expected=%v, actual=%v\n", test.toPreload, toPreload)
			assert.Equal(t, test.toPreload, toPreload)
			assert.Nil(t, err)
		}
	}
}
