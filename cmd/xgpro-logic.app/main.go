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
	Toml bool   `xor:"format" help:"Output as toml" type:"bool"`
	Json bool   `xor:"format" help:"Output as json" type:"bool"`
}

type CreateLgcCmd struct {
	Path        string `arg required help:"Input path." type:"path"`
	OutputPath  string `arg required help:"Output path of created lgc file." type:"path"`
	InputFormat string `short:"f" enum:"json,toml" default:"toml" help"Format of the input file"`
}

func (cmd *ViewCmd) Run(globals *Globals) error {
	lgc, err := xgpro.ParseLGCFile(cmd.Path)
	if err != nil {
		return err
	}

	if cmd.Toml {
		return xgpro.DescribeToml(lgc)
	}
	if cmd.Json {
		return xgpro.DescribeJson(lgc)
	}

	return xgpro.DumpLGCFile(lgc)
}

func (cmd *CreateLgcCmd) Run(globals *Globals) error {
	return xgpro.ConvertFile(cmd.Path, cmd.InputFormat, cmd.OutputPath)
}

type CLI struct {
	Globals

	Describe ViewCmd      `cmd help:"Describes an .lgc file to stdout"`
	Lgc      CreateLgcCmd `cmd help:"Create a lgc file from an input file"`
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
