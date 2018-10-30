package query

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
			text: "field := 'AbC'",
			exp: &Filtering{
				&Filtering_StringCondition{
					&StringCondition{
						FieldPath:  []string{"field"},
						Value:      "AbC",
						Type:       StringCondition_IEQ,
						IsNegative: false,
					},
				},
			},
		},
		{
			text: "not field := 'AbC'",
			exp: &Filtering{
				&Filtering_StringCondition{
					&StringCondition{
						FieldPath:  []string{"field"},
						Value:      "AbC",
						Type:       StringCondition_IEQ,
						IsNegative: true,
					},
				},
			},
		},
		{
			text: "(field := 'AbC') and (field1 := 'BcD')",
			exp: &Filtering{
				&Filtering_Operator{
					Operator: &LogicalOperator{
						Left: &LogicalOperator_LeftStringCondition{
							&StringCondition{
								FieldPath:  []string{"field"},
								Value:      "AbC",
								Type:       StringCondition_IEQ,
								IsNegative: false,
							},
						},
						Right: &LogicalOperator_RightStringCondition{
							&StringCondition{
								FieldPath:  []string{"field1"},
								Value:      "BcD",
								Type:       StringCondition_IEQ,
								IsNegative: false,
							},
						},
					},
				},
			},
		},
		{
			text: "(field := 'AbC') and not(field1 := 'BcD')",
			exp: &Filtering{
				&Filtering_Operator{
					Operator: &LogicalOperator{
						Left: &LogicalOperator_LeftStringCondition{
							&StringCondition{
								FieldPath:  []string{"field"},
								Value:      "AbC",
								Type:       StringCondition_IEQ,
								IsNegative: false,
							},
						},
						Right: &LogicalOperator_RightStringCondition{
							&StringCondition{
								FieldPath:  []string{"field1"},
								Value:      "BcD",
								Type:       StringCondition_IEQ,
								IsNegative: true,
							},
						},
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
			text: "field in [1 , 9 ,21]",
			exp: &Filtering{
				&Filtering_NumberArrayCondition{
					&NumberArrayCondition{
						FieldPath:  []string{"field"},
						Values:     []float64{1, 9, 21},
						Type:       NumberArrayCondition_IN,
						IsNegative: false,
					},
				},
			},
		},
		{
			text: "not (field in [1 , 9 ,21])",
			exp: &Filtering{
				&Filtering_NumberArrayCondition{
					&NumberArrayCondition{
						FieldPath:  []string{"field"},
						Values:     []float64{1, 9, 21},
						Type:       NumberArrayCondition_IN,
						IsNegative: true,
					},
				},
			},
		},
		{
			text: "field in ['Hello' , 'World']",
			exp: &Filtering{
				&Filtering_StringArrayCondition{
					&StringArrayCondition{
						FieldPath:  []string{"field"},
						Values:     []string{"Hello", "World"},
						Type:       StringArrayCondition_IN,
						IsNegative: false,
					},
				},
			},
		},
		{
			text: "not (field in ['Hello' , 'World'])",
			exp: &Filtering{
				&Filtering_StringArrayCondition{
					&StringArrayCondition{
						FieldPath:  []string{"field"},
						Values:     []string{"Hello", "World"},
						Type:       StringArrayCondition_IN,
						IsNegative: true,
					},
				},
			},
		},
		{
			text: "(not (field in ['Hello' , 'World']) and (field := 'Mike'))",
			exp: &Filtering{
				&Filtering_Operator{
					&LogicalOperator{
						Left: &LogicalOperator_LeftStringArrayCondition{
							&StringArrayCondition{
								FieldPath:  []string{"field"},
								Values:     []string{"Hello", "World"},
								Type:       StringArrayCondition_IN,
								IsNegative: true,
							},
						},
						Right: &LogicalOperator_RightStringCondition{
							&StringCondition{
								FieldPath:  []string{"field"},
								Value:      "Mike",
								Type:       StringCondition_IEQ,
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
			text: "field ieq 'HeLLo'",
			exp: &Filtering{
				&Filtering_StringCondition{
					&StringCondition{
						FieldPath:  []string{"field"},
						Value:      "HeLLo",
						Type:       StringCondition_IEQ,
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
		"field1 =: 'AbC'",
		"field1 : = 'AbC'",
	}

	for _, test := range tests {
		token, err := p.Parse(test)
		assert.Nil(t, token)
		assert.IsType(t, &UnexpectedSymbolError{}, err)
	}
}
