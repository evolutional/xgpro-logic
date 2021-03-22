# XGPRO-Logic

This is a little commandline utility that makes working with the XGecu Universal Programmer a little bit nicer.

The programmer has a useful "IC Test" function that allows one to run logic tests on a chip. However the user interface for working with that functionality is... not great.

It allows you to inspect `.lgc` files and convert them to/from [toml](https://github.com/toml-lang/toml) format.

With the file in a toml format you can add your own logic vector descriptions much faster and then convert back into the `.lgc` format to import into the Xgpro tool.

## Usage

The application has three commands:

### Describe
`describe <path>` which reads an `.lgc` file and prints some information about the contents:

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

### TOML
`toml <input> <output>` reads an `.lgc` file and outputs a `.toml` file that describes it

### LGC
`lgc <input> <output>` reads a `.toml` file and outputs an `.lgc` file that can be imported into the Xgpro tool

## Example TOML file

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

Symbol | Diretion | Description
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

## Building

Clone the repository and run `go build  -o xgpro-logic cmd/xgpro-logic.app/main.go` to create the binary file.


## Thanks

Special thanks to [@BreakIntoProg](https://twitter.com/breakintoprog) for the inspiration to write this tool