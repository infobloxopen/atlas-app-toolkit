package plugin

import (
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

// Imports that are added by default but unneeded
var unneededImports = []string{
	"import fmt \"fmt\"\n",
	"import math \"math\"\n",
	"import proto \"github.com/gogo/protobuf/proto\"\n",
	"import _ \"github.com/infobloxopen/protoc-gen-gorm/options\"\n",
	"import _ \"github.com/infobloxopen/protoc-gen-gorm/types\"\n",
	"var _ = proto.Marshal\n",
	"var _ = fmt.Errorf\n",
	"var _ = math.Inf\n",
	"import _ \"google/protobuf\"\n",
	"import _ \"google.golang.org/genproto/googleapis/api/annotations\"\n",
	"import _ \"github.com/lyft/protoc-gen-validate/validate\"\n",
	"import _ \"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options\"\n",
	"import _ \"github.com/infobloxopen/atlas-app-toolkit/query\"\n",
	"import _ \"github.com/infobloxopen/atlas-app-toolkit/rpc/resource\"\n",
	"import _ \"github.com/infobloxopen/atlas-contacts-app/pkg/valid\"\n",
	"import _ \"github.com/infobloxopen/atlas-app-toolkit/query/protoc-gen-perm/options/\"\n",
}

// CleanImports removes extraneous imports and lines from a proto response
func CleanImports(pFileText *string) *string {
	if pFileText == nil {
		return pFileText
	}
	fileText := *pFileText
	for _, dep := range unneededImports {
		fileText = strings.Replace(fileText, dep, "", -1)
	}
	return &fileText
}

// GenerateImports writes out required imports for the generated files
func (p *PermPlugin) GenerateImports(file *generator.FileDescriptor) {
	p.PrintImport("options", "github.com/infobloxopen/atlas-app-toolkit/query/protoc-gen-perm/options")
	p.PrintImport("query", "github.com/infobloxopen/atlas-app-toolkit/query")
}
