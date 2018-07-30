package test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestProtocOnInvalidProtoFiles(t *testing.T) {
	data := map[string]string{
		"wrong_operation.proto": "error:Error for message 'User': 'SORT' is unknown permission operation for field 'first_name'",
		"wrong_type.proto":      "error:Error for message 'User': Field 'OnVacation' does not support permission operations, supported only by string and numeric types",
	}

	gosrc := "-I" + os.Getenv("GOPATH") + "/src"

	for fileName, expected := range data {
		out, _ := exec.Command("protoc", "-I.", gosrc,
			"--perm_out=\"./\"", fileName).CombinedOutput()
		output := string(out)
		if !strings.Contains(output, expected) {
			t.Errorf("Error, expected error '%s' got  '%s'", expected, output)
		}

	}
}
