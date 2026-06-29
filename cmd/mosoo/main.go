package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/lathe-cli/lathe/pkg/config"
	"github.com/lathe-cli/lathe/pkg/lathe"
	"github.com/lathe-cli/lathe/pkg/runtime"
	kitup "github.com/samzong/kitup/go"
	kitupcobra "github.com/samzong/kitup/go-cobra"

	"github.com/langgenius/mosoo-cli-go/internal/agentmanifest"
	"github.com/langgenius/mosoo-cli-go/internal/buildinfo"
	"github.com/langgenius/mosoo-cli-go/internal/consolecommands"
	"github.com/langgenius/mosoo-cli-go/internal/doctor"
	"github.com/langgenius/mosoo-cli-go/internal/generated"
	"github.com/langgenius/mosoo-cli-go/internal/publicthreads"
	"github.com/langgenius/mosoo-cli-go/internal/target"
	publishskills "github.com/langgenius/mosoo-cli-go/publish/skills"
)

//go:embed cli.yaml
var manifestBytes []byte

func main() {
	buildinfo.Apply()

	m, err := config.Load(manifestBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load cli.yaml: %v\n", err)
		os.Exit(1)
	}
	config.Bind(m)
	root := lathe.NewApp(m)
	target.Install(root)
	root.AddCommand(agentmanifest.NewCommand())
	root.AddCommand(doctor.NewCommand())
	root.AddCommand(kitupcobra.NewSkillCommand(kitupcobra.Options{
		AppID:  "mosoo",
		Bundle: kitup.FSBundle(publishskills.Mosoo, "mosoo"),
	}))
	if err := generated.MountModules(root); err != nil {
		os.Exit(runtime.FormatError(err, "table", os.Stderr))
	}
	if err := consolecommands.Install(root); err != nil {
		os.Exit(runtime.FormatError(err, "table", os.Stderr))
	}
	if err := publicthreads.Install(root); err != nil {
		os.Exit(runtime.FormatError(err, "table", os.Stderr))
	}
	os.Exit(runtime.Execute(root))
}
