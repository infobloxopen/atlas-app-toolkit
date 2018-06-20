package resource

import "testing"

func TestParseString(t *testing.T) {
	tcases := []struct {
		String                  string
		ExpectedApplicationName string
		ExpectedResourceType    string
		ExpectedResourceID      string
	}{
		{
			"/a/b/c/",
			"a",
			"b",
			"c",
		},
		{
			"/a/b/c/d",
			"a",
			"b",
			"c/d",
		},
		{
			"a/b/",
			"",
			"a",
			"b",
		},
		{
			"/c",
			"",
			"",
			"c",
		},
		{
			"",
			"",
			"",
			"",
		},
	}

	for _, tc := range tcases {
		aname, rtype, rid := ParseString(tc.String)
		if tc.ExpectedApplicationName != aname {
			t.Errorf("invalid application name %s, expected %s", aname, tc.ExpectedApplicationName)
		}
		if tc.ExpectedResourceType != rtype {
			t.Errorf("invalid resource type %s, expected %s", rtype, tc.ExpectedResourceType)
		}
		if tc.ExpectedResourceID != rid {
			t.Errorf("invalid resource id %s, expected %s", rid, tc.ExpectedResourceID)
		}
	}
}

func TestBuildString(t *testing.T) {
	tcases := []struct {
		ApplicationName string
		ResourceType    string
		ResourceID      string
		ExpectedID      string
	}{
		{
			"a",
			"b",
			"c",
			"a/b/c",
		},
		{
			"",
			"b",
			"c",
			"b/c",
		},
		{
			"",
			"",
			"c",
			"c",
		},
	}
	for _, tc := range tcases {
		id := BuildString(tc.ApplicationName, tc.ResourceType, tc.ResourceID)
		if id != tc.ExpectedID {
			t.Errorf("invalid resource identifier %s, expected %s", id, tc.ExpectedID)
		}
	}
}
