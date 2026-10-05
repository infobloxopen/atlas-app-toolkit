package gorm

import (
	"context"
	"fmt"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/v2/rpc/resource"
)

type echoORM struct {
	Id string
}

// echoProto's ToORM reports how the filter value was parsed into an Identifier.
type echoProto struct {
	Id *resource.Identifier
}

func (*echoProto) Reset()         {}
func (*echoProto) ProtoMessage()  {}
func (*echoProto) String() string { return "echo" }

func (p *echoProto) ToORM(context.Context) (echoORM, error) {
	return echoORM{Id: fmt.Sprintf("%q|%q|%q", p.Id.ApplicationName, p.Id.ResourceType, p.Id.ResourceId)}, nil
}

func TestProcessStringCondition_IdentifierFilterValue(t *testing.T) {
	tcases := []struct {
		name  string
		value string
		want  string
	}{
		{"full identifier", `app/res/id1`, `"app"|"res"|"id1"`},
		{"id only", `id1`, `""|""|"id1"`},
		{"backslash is literal", `a\b`, `""|""|"a\\b"`},
		{"double backslash is literal", `a\\b`, `""|""|"a\\\\b"`},
		{"escaped slash is literal", `a\/b`, `""|"a\\"|"b"`},
		{"quote is kept", `a"b`, `""|""|"a\"b"`},
		{"newline is kept", "a\nb", `""|""|"a\nb"`},
	}

	p := &DefaultFilteringConditionProcessor{&echoProto{}}
	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := p.ProcessStringCondition(context.Background(), []string{"id"}, tc.value)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
