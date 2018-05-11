package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilteringLexer(t *testing.T) {

	lexer := NewFilteringLexer(`()14 13.23 'abc'"bcd" field1 and or  not == eq ne != match ~ nomatch !~ gt > ge >= lt < le <= null`)
	tests := []Token{
		LparenToken{},
		RparenToken{},
		NumberToken{Value: 14},
		NumberToken{Value: 13.23},
		StringToken{Value: "abc"},
		StringToken{Value: "bcd"},
		FieldToken{Value: "field1"},
		AndToken{},
		OrToken{},
		NotToken{},
		EqToken{},
		EqToken{},
		NeToken{},
		NeToken{},
		MatchToken{},
		MatchToken{},
		NmatchToken{},
		NmatchToken{},
		GtToken{},
		GtToken{},
		GeToken{},
		GeToken{},
		LtToken{},
		LtToken{},
		LeToken{},
		LeToken{},
		NullToken{},
		EOFToken{},
	}

	for _, test := range tests {
		token, err := lexer.NextToken()
		assert.Equal(t, test, token)
		assert.Nil(t, err)
	}
}

func TestFilteringLexerNegative(t *testing.T) {
	tests := []string{
		"=!",
		"!!",
		"%",
		"'string",
	}

	for _, test := range tests {
		lexer := NewFilteringLexer(test)
		token, err := lexer.NextToken()
		assert.Nil(t, token)
		assert.IsType(t, &UnexpectedSymbolError{}, err)
	}

}
