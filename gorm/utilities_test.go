package gorm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type Human struct {
	Name  string
	Age   uint32 `gorm:"column:years"`
	Child *Child
}

func (p Human) TableName() string {
	return "db_humans"
}

type Child struct {
	Name string
}

func TestHandleFieldPath(t *testing.T) {
	tests := []struct {
		fieldPath []string
		dbName    string
		err       bool
	}{
		{[]string{"name"}, "db_humans.name", false},
		{[]string{"age"}, "db_humans.years", false},
		{[]string{"child", "name"}, "children.name", false},
		{[]string{}, "", true},
	}
	for _, test := range tests {
		res, err := HandleFieldPath(test.fieldPath, &Human{})
		if test.err {
			assert.Equal(t, "", res)
			assert.NotNil(t, err)
		} else {
			assert.Equal(t, test.dbName, res)
			assert.Nil(t, err)
		}
	}
}
