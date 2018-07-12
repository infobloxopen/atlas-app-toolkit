package gorm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type JoinsModel struct {
	Id         string
	ParentName string
	Child      JoinsChild  `gorm:"association_foreignkey:Id;foreignkey:ModelId"`
	Parent     JoinsParent `gorm:"association_foreignkey:Name;foreignkey:ParentName"`
}

type JoinsChild struct {
	ModelId string
}

type JoinsParent struct {
	Name string
}

func TestJoinInfo(t *testing.T) {
	tests := []struct {
		assoc      string
		tableName  string
		sourceKeys []string
		targetKeys []string
	}{
		{
			"Child",
			"joins_children",
			[]string{"joins_models.id"},
			[]string{"joins_children.model_id"},
		},
		{
			"Parent",
			"joins_parents",
			[]string{"joins_models.parent_name"},
			[]string{"joins_parents.name"},
		},
	}
	for _, test := range tests {
		tableName, sourceKeys, targetKeys, err := JoinInfo(&JoinsModel{}, test.assoc)
		assert.Equal(t, test.tableName, tableName)
		assert.Equal(t, test.sourceKeys, sourceKeys)
		assert.Equal(t, test.targetKeys, targetKeys)
		assert.Nil(t, err)
	}
}
