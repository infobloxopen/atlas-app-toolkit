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
	SubModelMix      SubModelMix
	SubModelMix2     SubModelMix2
	NonCamelModelMIX NonCamelModelMIX
}

type NonCamelModelMIX struct {
	NonCamelMixProperty string
	SubModel            SubModel
}

type SubModelMix struct {
	SubModelMixProperty string
	NonCamelMODEL       NonCamelMODEL
}

type SubModelMix2 struct {
	SubModelMixProperty string
	NonCamelSUBMODEL    NonCamelSUBMODEL
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
		fs                  string
		toPreload           []string
		toPreloadIgnoreCase []string
		err                 bool
	}{
		{
			"sub_model_mix2.non_camel_submodel",
			[]string{"SubModelMix2"},
			[]string{"SubModelMix2.NonCamelSUBMODEL", "SubModelMix2"},
			false,
		},
		{
			"non_camel_model_mix.sub_model.sub_property",
			nil,
			[]string{"NonCamelModelMIX.SubModel", "NonCamelModelMIX"},
			false,
		},
		{
			"non_camel_model_mix,sub_model,non_camel_2_model,cycle_model",
			[]string{"CycleModel", "SubModel"},
			[]string{"CycleModel", "NonCAMEL2Model", "NonCamelModelMIX", "SubModel"},
			false,
		},
		{
			"sub_model_mix.non_camel_model",
			[]string{"SubModelMix"},
			[]string{"SubModelMix.NonCamelMODEL", "SubModelMix"},
			false,
		},
		{
			"non_camel_model.non_camel_submodel.noncamelsubproperty",
			nil,
			[]string{"NonCamelMODEL.NonCamelSUBMODEL", "NonCamelMODEL"},
			false,
		},
		{
			"non_camel_model.non_camel_submodel.noncamelsubproperty",
			nil,
			[]string{"NonCamelMODEL.NonCamelSUBMODEL", "NonCamelMODEL"},
			false,
		},
		{
			"non_camel_model.non_camel_submodel",
			nil,
			[]string{"NonCamelMODEL.NonCamelSUBMODEL", "NonCamelMODEL"},
			false,
		},
		{
			"non_camel_model",
			nil,
			[]string{"NonCamelMODEL"},
			false,
		},
		{
			"non_camel_model.noncamelproperty",
			nil,
			[]string{"NonCamelMODEL"},
			false,
		},
		{
			"non_CAMEL_2_Model,Non_camel_2_model,non_camel2_model,non_camel_2model",
			nil,
			[]string{"NonCAMEL2Model", "NonCAMEL2Model", "NonCAMEL2Model", "NonCAMEL2Model"},
			false,
		},
		{
			"property",
			nil,
			nil,
			false,
		},
		{
			"property,sub_model",
			[]string{"SubModel"},
			[]string{"SubModel"},
			false,
		},
		{
			"property,sub_model.sub_property",
			[]string{"SubModel"},
			[]string{"SubModel"},
			false,
		},
		{
			"sub_model,sub_models",
			[]string{"SubModel", "SubModels"},
			[]string{"SubModel", "SubModels"},
			false,
		},
		{
			"sub_model,sub_models.sub_property",
			[]string{"SubModel", "SubModels"},
			[]string{"SubModel", "SubModels"},
			false,
		},
		{
			"sub_model",
			[]string{"SubModel"},
			[]string{"SubModel"},
			false,
		},
		{
			"sub_model.sub_sub_model",
			[]string{"SubModel.SubSubModel", "SubModel"},
			[]string{"SubModel.SubSubModel", "SubModel"},
			false,
		},
		{
			"sub_model.sub_sub_model.sub_sub_property",
			[]string{"SubModel.SubSubModel", "SubModel"},
			[]string{"SubModel.SubSubModel", "SubModel"},
			false,
		},
		{
			"unknown_property",
			nil,
			nil,
			false,
		},
		{
			"not_preload_obj,preload_obj",
			[]string{"NotPreloadObj", "PreloadObj"},
			[]string{"NotPreloadObj", "PreloadObj"},
			false,
		},
		{
			"",
			[]string{"SubModel.SubSubModel", "SubModel", "SubModels.SubSubModel", "SubModels",
				"CycleModel", "PreloadObj.SubSubModel", "PreloadObj",
				"NonCamelMODEL.NonCamelSUBMODEL", "NonCamelMODEL", "NonCAMEL2Model", "NonCamelSUBMODEL",
				"SubModelMix.NonCamelMODEL.NonCamelSUBMODEL", "SubModelMix.NonCamelMODEL", "SubModelMix",
				"SubModelMix2.NonCamelSUBMODEL", "SubModelMix2",
				"NonCamelModelMIX.SubModel.SubSubModel", "NonCamelModelMIX.SubModel", "NonCamelModelMIX"},
			[]string{"SubModel.SubSubModel", "SubModel", "SubModels.SubSubModel", "SubModels",
				"CycleModel", "PreloadObj.SubSubModel", "PreloadObj",
				"NonCamelMODEL.NonCamelSUBMODEL", "NonCamelMODEL", "NonCAMEL2Model", "NonCamelSUBMODEL",
				"SubModelMix.NonCamelMODEL.NonCamelSUBMODEL", "SubModelMix.NonCamelMODEL", "SubModelMix",
				"SubModelMix2.NonCamelSUBMODEL", "SubModelMix2",
				"NonCamelModelMIX.SubModel.SubSubModel", "NonCamelModelMIX.SubModel", "NonCamelModelMIX"},
			false,
		},
	}

	// as default, case-sensitive search is enabled
	for _, test := range tests {
		toPreload, err := FieldSelectionStringToGorm(context.Background(), test.fs, &Model{})
		if test.err {
			assert.Nil(t, toPreload)
			assert.NotNil(t, err)
		} else {
			fmt.Printf("expected=%v\nactual=%v\n\n", test.toPreload, toPreload)
			assert.Equal(t, test.toPreload, toPreload)
			assert.Nil(t, err)
		}
	}

	// disable case-insensitive search for field string
	EnableCaseSensitive(false)
	for _, test := range tests {
		toPreloadIgnoreCase, err := FieldSelectionStringToGorm(context.Background(), test.fs, &Model{})
		if test.err {
			assert.Nil(t, toPreloadIgnoreCase)
			assert.NotNil(t, err)
		} else {
			fmt.Printf("Case-sensitive search is DISABLED \nexpected=%v\nactual=%v\n\n", test.toPreloadIgnoreCase, toPreloadIgnoreCase)
			assert.Equal(t, test.toPreloadIgnoreCase, toPreloadIgnoreCase)
			assert.Nil(t, err)
		}
	}

	// enabled case-sensitive search for field string
	EnableCaseSensitive(true)
	for _, test := range tests {
		toPreload, err := FieldSelectionStringToGorm(context.Background(), test.fs, &Model{})
		if test.err {
			assert.Nil(t, toPreload)
			assert.NotNil(t, err)
		} else {
			fmt.Printf("Case-sensitive search is ENABLED\nexpected=%v\nactual=%v\n\n", test.toPreload, toPreload)
			assert.Equal(t, test.toPreload, toPreload)
			assert.Nil(t, err)
		}
	}
}
