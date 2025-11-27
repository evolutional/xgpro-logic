# XGPRO-Logic

This is a little commandline utility that makes working with the XGecu Universal Programmer a little bit nicer.

The programmer has a useful "IC Test" function that allows one to run logic tests on a chip. However the user interface for working with that functionality is... not great.

It allows you to inspect `.lgc` files and convert them to/from [toml](https://github.com/toml-lang/toml) format.

With the file in a toml format you can add your own logic vector descriptions much faster and then convert back into the `.lgc` format to import into the Xgpro tool.

## Usage

The application has three commands.

For more information run `./xgpro-logic --help`

```
Usage: xgpro-logic <command>

Flags:
  -h, --help              Show context-sensitive help.
      --version=STRING

Commands:
  describe <path>
    Describes an .lgc file to stdout

  lgc <path> <output-path>
    Create a lgc file from a toml file

Run "xgpro-logic <command> --help" for more information on a command.
```

### Describe

```
Usage: xgpro-logic describe <path>

Describes an .lgc file to stdout

Arguments:
  <path>    Input path.

Flags:
  -h, --help              Show context-sensitive help.
      --version=STRING

      --toml              Output as toml
      --json              Output as json
      --xml               Output as xml
```

`describe <path>` which reads an `.lgc` file and dumps some information about the contents to stdout:

```
File contains 1 entries
Entry #0
        Name:   OliTest
        Pins:   8
        VCC:    5.0V
        Vectors: 2
                #0: 1 0 0 G 0 0 0 V 
                #1: 0 1 0 G 0 0 0 V 
```

You can supply the optional flags `--toml`, `--json` or `--xml` to pipe the data into another tool (such as `jq`).

### LGC
`lgc <input> <output>` reads an input file and outputs an `.lgc` file that can be imported into the Xgpro tool.

```
Usage: xgpro-logic lgc <path> <output-path>

Create a lgc file from an input file

Arguments:
  <path>           Input path.
  <output-path>    Output path of created lgc file.

Flags:
  -h, --help                   Show context-sensitive help.
      --version=STRING

  -f, --input-format="toml"
```

The input format can either be `toml` or `json`.

### Example TOML file

```toml
[[ics]]
# Example IC definition
name = "OLI's IC"
pins = 8
vcc = 5
vectors = [
    "100G000V",
    "010G000V",
]

[[ics]]
# Example IC definition
name = "Another IC"
pins = 10
vcc = 3.3
vectors = [
    "1000G0000V",
    "0100G0000V",
]
```

Each line in the `vectors` array is a string that contains a character indicating the logic level for the test. These match those used in the Xgpro tool itself, but they're listed here:

Symbol | Direction | Description
--- | --- | --
0 | Input | Logic LOW
1 | Input | Logic HIGH
L | Output | Logic LOW
H | Output | Logic HIGH
Z | Output | High Impedence
C | Input | Clock Pulse
X | - | Ignored
V | Input | VCC
G | Input | Ground

### Example Json file

The json format is structurally identical to the `toml` format.

```json
{
    "ics": [
        {
            "name": "OLI's IC",
            "pins": 8,
            "vcc": 5,
            "vectors": [
                "100G000V",
                "010G000V"
            ]
        },
        {
            "name": "Another IC",
            "pins": 10,
            "vcc": 3.3,
            "vectors": [
                "1000G0000V",
                "0100G0000V"
            ]
        }
    ]
}
```

### Example XML file

The xml format is made to be compatible with the linux/mac version of minipro.

```xml
<?xml version="1.0" encoding="utf-8"?>
<logicic>
  <database device="TL866II" type="LOGIC">
    <manufacturer name="Logic Ic">
      <ic name="OLI's IC" pins="8" voltage="5V" type="5">
        <vector id="00"> 1 0 0 G 0 0 0 V </vector>
        <vector id="01"> 0 1 0 G 0 0 0 V </vector>
      </ic>
      <ic name="Another IC" pins="10" voltage="3.3V" type="5">
        <vector id="00"> 1 0 0 0 G 0 0 0 0 V </vector>
        <vector id="01"> 0 1 0 0 G 0 0 0 0 V </vector>
      </ic>
    </manufacturer>
  </database>
</logicic>
```

WARNING: The voltage and type in minipro are always 5V and 5 respectively for all ICs, so
it might be that these are not used. In this xml export implementation, voltage will output vcc
and type is hardcoded to 5.
```
xgpro-logic describe examples/test_1j.lgc --xml > examples/test_1j.xml

minipro --logicic examples/test_1j.xml -p "OLI's IC" -T
```

## Building

Clone the repository and run: `go mod tidy` to fetch the dependencies.

Then run `go build  -o xgpro-logic ./cmd/xgpro-logic.app/main.go` to create the binary file.

Alternatively, if your platform supports `make`, then run `make build` and find the output in the `build` directory.

## Thanks

Special thanks to [@BreakIntoProg](https://twitter.com/breakintoprog) for the inspiration to write this tool