package query

import proto "github.com/golang/protobuf/proto"

func (nlop *NestedLogicalOperator) ToLogicalOperator() (*LogicalOperator, error) {
	bs, err := proto.Marshal(nlop)
	if err != nil {
		return &LogicalOperator{}, err
	}
	lop := LogicalOperator{}
	return &lop, proto.Unmarshal(bs, &lop)
}

func (lop *LogicalOperator) ToNestedLogicalOperator() (*NestedLogicalOperator, error) {
	bs, err := proto.Marshal(lop)
	if err != nil {
		return &NestedLogicalOperator{}, err
	}
	nlop := NestedLogicalOperator{}
	return &nlop, proto.Unmarshal(bs, &nlop)
}

var _ FilteringExpression = &NestedLogicalOperator{}
var _ FilteringExpression = &NestedLogicalOperator_LeftOperator{}
var _ FilteringExpression = &NestedLogicalOperator_RightOperator{}
var _ FilteringExpression = &NestedLogicalOperator_LeftStringCondition{}
var _ FilteringExpression = &NestedLogicalOperator_RightStringCondition{}
var _ FilteringExpression = &NestedLogicalOperator_LeftNumberCondition{}
var _ FilteringExpression = &NestedLogicalOperator_RightNumberCondition{}
var _ FilteringExpression = &NestedLogicalOperator_LeftNullCondition{}
var _ FilteringExpression = &NestedLogicalOperator_RightNullCondition{}
var _ FilteringExpression = &NestedLogicalOperator_LeftStringArrayCondition{}
var _ FilteringExpression = &NestedLogicalOperator_RightStringArrayCondition{}
var _ FilteringExpression = &NestedLogicalOperator_LeftNumberArrayCondition{}
var _ FilteringExpression = &NestedLogicalOperator_RightNumberArrayCondition{}

func (*LogicalOperator_LeftOperator) isNestedLogicalOperator_Left() {}

func (*LogicalOperator_RightOperator) isNestedLogicalOperator_Right() {}

func (*NestedLogicalOperator_LeftOperator) isLogicalOperator_Left() {}

func (*NestedLogicalOperator_LeftStringCondition) isLogicalOperator_Left() {}

func (*NestedLogicalOperator_LeftNumberCondition) isLogicalOperator_Left() {}

func (*NestedLogicalOperator_LeftNullCondition) isLogicalOperator_Left() {}

func (*NestedLogicalOperator_LeftStringArrayCondition) isLogicalOperator_Left() {}

func (*NestedLogicalOperator_LeftNumberArrayCondition) isLogicalOperator_Left() {}

func (*NestedLogicalOperator_RightOperator) isLogicalOperator_Right() {}

func (*NestedLogicalOperator_RightStringCondition) isLogicalOperator_Right() {}

func (*NestedLogicalOperator_RightNumberCondition) isLogicalOperator_Right() {}

func (*NestedLogicalOperator_RightNullCondition) isLogicalOperator_Right() {}

func (*NestedLogicalOperator_RightStringArrayCondition) isLogicalOperator_Right() {}

func (*NestedLogicalOperator_RightNumberArrayCondition) isLogicalOperator_Right() {}

func (m *NestedLogicalOperator_LeftOperator) Filter(obj interface{}) (bool, error) {
	return m.LeftOperator.Filter(obj)
}

func (m *NestedLogicalOperator_LeftStringCondition) Filter(obj interface{}) (bool, error) {
	return m.LeftStringCondition.Filter(obj)
}

func (m *NestedLogicalOperator_LeftNumberCondition) Filter(obj interface{}) (bool, error) {
	return m.LeftNumberCondition.Filter(obj)
}

func (m *NestedLogicalOperator_LeftNullCondition) Filter(obj interface{}) (bool, error) {
	return m.LeftNullCondition.Filter(obj)
}

func (m *NestedLogicalOperator_LeftStringArrayCondition) Filter(obj interface{}) (bool, error) {
	return m.LeftStringArrayCondition.Filter(obj)
}

func (m *NestedLogicalOperator_LeftNumberArrayCondition) Filter(obj interface{}) (bool, error) {
	return m.LeftNumberArrayCondition.Filter(obj)
}

func (m *NestedLogicalOperator_RightOperator) Filter(obj interface{}) (bool, error) {
	return m.RightOperator.Filter(obj)
}

func (m *NestedLogicalOperator_RightStringCondition) Filter(obj interface{}) (bool, error) {
	return m.RightStringCondition.Filter(obj)
}

func (m *NestedLogicalOperator_RightNumberCondition) Filter(obj interface{}) (bool, error) {
	return m.RightNumberCondition.Filter(obj)
}

func (m *NestedLogicalOperator_RightNullCondition) Filter(obj interface{}) (bool, error) {
	return m.RightNullCondition.Filter(obj)
}

func (m *NestedLogicalOperator_RightStringArrayCondition) Filter(obj interface{}) (bool, error) {
	return m.RightStringArrayCondition.Filter(obj)
}

func (m *NestedLogicalOperator_RightNumberArrayCondition) Filter(obj interface{}) (bool, error) {
	return m.RightNumberArrayCondition.Filter(obj)
}
