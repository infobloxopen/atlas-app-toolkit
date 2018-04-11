package op

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilteringParser(t *testing.T) {

	tests := []struct {
		text string
		exp  *Filtering
	}{
		{
			text: "not(not(not field1 == 'abc' or not field2 == 'bcd') and (field3 != 'cde'))",
			exp: &Filtering{
				&Filtering_Operator{
					&LogicalOperator{
						Left: &LogicalOperator_LeftOperator{
							&LogicalOperator{
								Left: &LogicalOperator_LeftStringCondition{
									&StringCondition{
										FieldPath:  []string{"field1"},
										Value:      "abc",
										Type:       StringCondition_EQ,
										IsNegative: true,
									},
								},
								Right: &LogicalOperator_RightStringCondition{
									&StringCondition{
										FieldPath:  []string{"field2"},
										Value:      "bcd",
										Type:       StringCondition_EQ,
										IsNegative: true,
									},
								},
								Type:       LogicalOperator_OR,
								IsNegative: true,
							},
						},
						Right: &LogicalOperator_RightStringCondition{
							&StringCondition{
								FieldPath:  []string{"field3"},
								Value:      "cde",
								Type:       StringCondition_EQ,
								IsNegative: true,
							},
						},
						Type:       LogicalOperator_AND,
						IsNegative: true,
					},
				},
			},
		},
		{
			text: "field1 == 'abc' or field2 == 'cde' and not field3 == 'cdf'",
			exp: &Filtering{
				&Filtering_Operator{
					&LogicalOperator{
						Left: &LogicalOperator_LeftStringCondition{
							&StringCondition{
								FieldPath:  []string{"field1"},
								Value:      "abc",
								Type:       StringCondition_EQ,
								IsNegative: false,
							},
						},
						Right: &LogicalOperator_RightOperator{
							&LogicalOperator{
								Left: &LogicalOperator_LeftStringCondition{
									&StringCondition{
										FieldPath:  []string{"field2"},
										Value:      "cde",
										Type:       StringCondition_EQ,
										IsNegative: false,
									},
								},
								Right: &LogicalOperator_RightStringCondition{
									&StringCondition{
										FieldPath:  []string{"field3"},
										Value:      "cdf",
										Type:       StringCondition_EQ,
										IsNegative: true,
									},
								},
								Type:       LogicalOperator_AND,
								IsNegative: false,
							},
						},
						Type:       LogicalOperator_OR,
						IsNegative: false,
					},
				},
			},
		},
		{
			text: "(field1 == 'abc' or field2 == 'cde') and (field3 == 'fbg' or field4 == 'zux')",
			exp: &Filtering{
				&Filtering_Operator{
					&LogicalOperator{
						Left: &LogicalOperator_LeftOperator{
							&LogicalOperator{
								Left: &LogicalOperator_LeftStringCondition{
									&StringCondition{
										FieldPath:  []string{"field1"},
										Value:      "abc",
										Type:       StringCondition_EQ,
										IsNegative: false,
									},
								},
								Right: &LogicalOperator_RightStringCondition{
									&StringCondition{
										FieldPath:  []string{"field2"},
										Value:      "cde",
										Type:       StringCondition_EQ,
										IsNegative: false,
									},
								},
								Type:       LogicalOperator_OR,
								IsNegative: false,
							},
						},
						Right: &LogicalOperator_RightOperator{
							&LogicalOperator{
								Left: &LogicalOperator_LeftStringCondition{
									&StringCondition{
										FieldPath:  []string{"field3"},
										Value:      "fbg",
										Type:       StringCondition_EQ,
										IsNegative: false,
									},
								},
								Right: &LogicalOperator_RightStringCondition{
									&StringCondition{
										FieldPath:  []string{"field4"},
										Value:      "zux",
										Type:       StringCondition_EQ,
										IsNegative: false,
									},
								},
								Type:       LogicalOperator_OR,
								IsNegative: false,
							},
						},
						Type:       LogicalOperator_AND,
						IsNegative: false,
					},
				},
			},
		},
		{
			text: "field == 'abc'",
			exp: &Filtering{
				&Filtering_StringCondition{
					&StringCondition{
						FieldPath:  []string{"field"},
						Value:      "abc",
						Type:       StringCondition_EQ,
						IsNegative: false,
					},
				},
			},
		},
		{
			text: "field != \"abc cde\"",
			exp: &Filtering{
				&Filtering_StringCondition{
					&StringCondition{
						FieldPath:  []string{"field"},
						Value:      "abc cde",
						Type:       StringCondition_EQ,
						IsNegative: true,
					},
				},
			},
		},
		{
			text: "field == 123",
			exp: &Filtering{
				&Filtering_NumberCondition{
					&NumberCondition{
						FieldPath:  []string{"field"},
						Value:      123,
						Type:       NumberCondition_EQ,
						IsNegative: false,
					},
				},
			},
		},
		{
			text: "field != 0.2343",
			exp: &Filtering{
				&Filtering_NumberCondition{
					&NumberCondition{
						FieldPath:  []string{"field"},
						Value:      0.2343,
						Type:       NumberCondition_EQ,
						IsNegative: true,
					},
				},
			},
		},
		{
			text: "field == null",
			exp: &Filtering{
				&Filtering_NullCondition{
					&NullCondition{
						FieldPath:  []string{"field"},
						IsNegative: false,
					},
				},
			},
		},
		{
			text: "field != null",
			exp: &Filtering{
				&Filtering_NullCondition{
					&NullCondition{
						FieldPath:  []string{"field"},
						IsNegative: true,
					},
				},
			},
		},
		{
			text: "not field != null",
			exp: &Filtering{
				&Filtering_NullCondition{
					&NullCondition{
						FieldPath:  []string{"field"},
						IsNegative: false,
					},
				},
			},
		},
		{
			text: "field ~ 'regex'",
			exp: &Filtering{
				&Filtering_StringCondition{
					&StringCondition{
						FieldPath:  []string{"field"},
						Value:      "regex",
						Type:       StringCondition_MATCH,
						IsNegative: false,
					},
				},
			},
		},
		{
			text: "field !~ 'regex'",
			exp: &Filtering{
				&Filtering_StringCondition{
					&StringCondition{
						FieldPath:  []string{"field"},
						Value:      "regex",
						Type:       StringCondition_MATCH,
						IsNegative: true,
					},
				},
			},
		},
		{
			text: "field < 123",
			exp: &Filtering{
				&Filtering_NumberCondition{
					&NumberCondition{
						FieldPath:  []string{"field"},
						Value:      123,
						Type:       NumberCondition_LT,
						IsNegative: false,
					},
				},
			},
		},
		{
			text: "not field <= 123",
			exp: &Filtering{
				&Filtering_NumberCondition{
					&NumberCondition{
						FieldPath:  []string{"field"},
						Value:      123,
						Type:       NumberCondition_LE,
						IsNegative: true,
					},
				},
			},
		},
		{
			text: "field > 123",
			exp: &Filtering{
				&Filtering_NumberCondition{
					&NumberCondition{
						FieldPath:  []string{"field"},
						Value:      123,
						Type:       NumberCondition_GT,
						IsNegative: false,
					},
				},
			},
		},
		{
			text: "field >= 123",
			exp: &Filtering{
				&Filtering_NumberCondition{
					&NumberCondition{
						FieldPath:  []string{"field"},
						Value:      123,
						Type:       NumberCondition_GE,
						IsNegative: false,
					},
				},
			},
		},
		{
			text: "",
			exp:  nil,
		},
	}

	p := NewFilteringParser()
	for _, test := range tests {
		result, err := p.Parse(test.text)
		assert.Equal(t, test.exp, result)
		assert.Nil(t, err)
	}
}

func TestFilteringParserNegative(t *testing.T) {
	p := NewFilteringParser()

	tests := []string{
		"(field1 == 'abc'",
		"((field1 == 'abc)'",
		"field1 == 'abc')",
		"null == field1",
		"field1 == field2",
		"field1 != field1",
		"field1 ~ 123",
		"field1 !~ 123",
		"field1 > 'abc'",
		"field1 >= 'bcd'",
		"field1 < or",
		"field1 <= null",
		"field1 or field2",
	}

	for _, test := range tests {
		token, err := p.Parse(test)
		assert.Nil(t, token)
		assert.IsType(t, &UnexpectedTokenError{}, err)
	}

	tests = []string{
		"field1 == 234.23.23",
		"field1 == 'abc",
		"field1 =! 'cdf'",
	}

	for _, test := range tests {
		token, err := p.Parse(test)
		assert.Nil(t, token)
		assert.IsType(t, &UnexpectedSymbolError{}, err)
	}
}
