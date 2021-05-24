package resource_test

import (
	"encoding/json"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

func TestIdentifier_MarshalJSONPB(t *testing.T) {
	tests := []struct {
		Name               string
		Identifier         *resource.Identifier
		ExpectedJSONString string
	}{
		{
			Name: "valid identifier",
			Identifier: &resource.Identifier{
				ApplicationName: "app",
				ResourceType:    "resource",
				ResourceId:      "res1",
			},
			ExpectedJSONString: `"app/resource/res1"`,
		},
		{
			Name: "null identifier",
			Identifier: &resource.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "",
			},
			ExpectedJSONString: `"null"`,
		},
		{
			Name: "invalid identifier",
			Identifier: &resource.Identifier{
				ApplicationName: "app1",
				ResourceType:    "resource/A",
				ResourceId:      "1234/5678",
			},
			ExpectedJSONString: `"app1/resource/A/1234/5678"`,
		},
	}

	for _, tt := range tests {
		buff, err := json.Marshal(tt.Identifier)
		if err != nil {
			t.Errorf("failed to marshal identifier %s - %s", tt.Identifier, err)
		}
		actual := string(buff)
		if actual != tt.ExpectedJSONString {
			t.Errorf("invalid identifier %s, expected %s", actual, tt.ExpectedJSONString)
		}
	}
}

func TestIdentifier_UnmarhsalJSONPB(t *testing.T) {
	tests := []struct {
		name               string
		jsonData           string
		expectedIdentifier *resource.Identifier
		expectError        bool
		expectMismatch     bool
	}{
		{
			name:     "valid json identifier",
			jsonData: `"app/resource/res2"`,
			expectedIdentifier: &resource.Identifier{
				ApplicationName: "app",
				ResourceType:    "resource",
				ResourceId:      "res2",
			},
		},
		{
			name:     "null json identifier",
			jsonData: `"null"`,
			expectedIdentifier: &resource.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "",
			},
		},
		{
			name:     "invalid json identifier",
			jsonData: `"app1/resource/A/1234/5678"`,
			expectedIdentifier: &resource.Identifier{
				ApplicationName: "app1",
				ResourceType:    "resource/A",
				ResourceId:      "1234/5678",
			},
			expectMismatch: true,
		},
	}

	for _, tt := range tests {
		id := &resource.Identifier{}
		if err := json.Unmarshal([]byte(tt.jsonData), id); err != nil {
			if tt.expectError {
				return
			}
			t.Errorf("%s: failed to unmarshal identifier %s", tt.name, err)
		}
		if id.String() != tt.expectedIdentifier.String() {
			t.Errorf("%s: invalid identifier %s, expected %s", tt.name, id, tt.expectedIdentifier)
		}
		if !tt.expectMismatch && id.ApplicationName != tt.expectedIdentifier.ApplicationName {
			t.Errorf("%s: identifier doesn't match expected ApplicationName: actual %s, expected %s", tt.name, id.ApplicationName, tt.expectedIdentifier.ApplicationName)
		}
		if !tt.expectMismatch && id.ResourceType != tt.expectedIdentifier.ResourceType {
			t.Errorf("%s: identifier doesn't match expected ResourceType: actual %s, expected %s", tt.name, id.ResourceType, tt.expectedIdentifier.ResourceType)
		}
		if !tt.expectMismatch && id.ResourceId != tt.expectedIdentifier.ResourceId {
			t.Errorf("%s: identifier doesn't match expected ResourceId: actual %s, expected %s", tt.name, id.ResourceId, tt.expectedIdentifier.ResourceId)
		}
	}
}
