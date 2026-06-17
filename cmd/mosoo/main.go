package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/lathe-cli/lathe/pkg/config"
	"github.com/lathe-cli/lathe/pkg/lathe"
	"github.com/lathe-cli/lathe/pkg/runtime"

	"github.com/langgenius/mosoo-cli-go/internal/generated"
)

//go:embed cli.yaml
var manifestBytes []byte

func main() {
	m, err := config.Load(manifestBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load cli.yaml: %v\n", err)
		os.Exit(1)
	}
	config.Bind(m)
	root := lathe.NewApp(m)
	if err := generated.MountModules(root); err != nil {
		os.Exit(runtime.FormatError(err, "table", os.Stderr))
	}
	os.Exit(runtime.Execute(root))
}
