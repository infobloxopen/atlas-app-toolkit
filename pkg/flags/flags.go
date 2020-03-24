package flags

import (
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	cliflag "k8s.io/component-base/cli/flag"
)

var programName = ""

func ProgramName() string {
	return programName
}

func init() {
	programName = os.Args[0]
}

// PrintFlags logs the flags in the flagset
func PrintFlags(flags *pflag.FlagSet, l logr.Logger) {
	flags.VisitAll(func(flag *pflag.Flag) {
		l.Info("FLAG", "name", flag.Name, "value", flag.Value)
	})
}

func SetCobraTemplateDefaults() {
	// used in cobra templates to display either `app-def-controller` or name of the current binary
	cobra.AddTemplateFunc("ProgramName", ProgramName)

	// used to enable replacement of `ProgramName` placeholder for cobra.Example, which has no template support
	cobra.AddTemplateFunc("prepare", func(s string) string { return strings.Replace(s, "{{ProgramName}}", programName, -1) })
}

func NormalizeFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.SetNormalizeFunc(cliflag.WarnWordSepNormalizeFunc) // Warn for "_" flags

	// Normalize all flags that are coming from other packages or pre-configurations
	// a.k.a. change all "_" to "-". e.g. glog package
	flags.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

	// From this point and forward we get warnings on flags that contain "_" separators
	cmd.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)

}
