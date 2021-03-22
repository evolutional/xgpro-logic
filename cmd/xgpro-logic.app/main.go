package main

import (
	"github.com/alecthomas/kong"

	"github.com/evolutional/xgpro-logic/internal/xgpro"
)

type Globals struct {
	Version string
}

type ViewCmd struct {
	Path string `arg required help:"Input path." type:"path"`
}

type CreateLgcCmd struct {
	Path       string `arg required help:"Input path." type:"path"`
	OutputPath string `arg required help:"Output path of created lgc file." type:"path"`
}

type CreateTomlCmd struct {
	Path       string `arg required help:"Input path." type:"path"`
	OutputPath string `arg required help:"Output path of created toml file." type:"path"`
}

func (cmd *ViewCmd) Run(globals *Globals) error {
	return xgpro.DumpLGCFile(cmd.Path)
}

func (cmd *CreateLgcCmd) Run(globals *Globals) error {
	lgc, err := xgpro.ParseTomlFile(cmd.Path)
	if err != nil {
		return err
	}
	return xgpro.WriteLgc(cmd.OutputPath, lgc)
}

func (cmd *CreateTomlCmd) Run(globals *Globals) error {
	lgc, err := xgpro.ParseLGCFile(cmd.Path)
	if err != nil {
		return err
	}
	return xgpro.WriteToml(cmd.OutputPath, lgc)
}

type CLI struct {
	Globals

	Describe ViewCmd       `cmd help:"Describes an .lgc file to stdout"`
	Lgc      CreateLgcCmd  `cmd help:"Create a lgc file from a toml file"`
	Toml     CreateTomlCmd `cmd help:"Create a toml file from a lgc file"`
}

func main() {

	cli := CLI{
		Globals: Globals{
			Version: "0.0.1",
		},
	}

	ctx := kong.Parse(&cli,
		kong.UsageOnError(),
	)
	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
