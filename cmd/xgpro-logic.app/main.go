package main

import (
	"os"

	"github.com/alecthomas/kong"

	"github.com/evolutional/xgpro-logic/internal/xgpro"
)

type Globals struct {
}

type ViewCmd struct {
	Path       string `arg required help:"Input path." type:"path"`
	Toml       bool   `xor:"format" help:"Output as toml" type:"bool"`
	Json       bool   `xor:"format" help:"Output as json" type:"bool"`
	Xml        bool   `xor:"format" help:"Output as xml" type:"bool"`
	OutputFile string `short:"o" help:"Output file path." type:"path"`
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

	file := os.Stdout

	if cmd.OutputFile != "" {
		file, err = os.Create(cmd.OutputFile)
		if err != nil {
			return err
		}
	}

	if cmd.Toml {
		return xgpro.DescribeToml(lgc, file)
	}
	if cmd.Json {
		return xgpro.DescribeJson(lgc, file)
	}
	if cmd.Xml {
		return xgpro.DescribeXml(lgc, file)
	}

	return xgpro.DumpLGCFile(lgc, file)
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
		Globals: Globals{},
	}

	ctx := kong.Parse(&cli,
		kong.UsageOnError(),
	)
	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
