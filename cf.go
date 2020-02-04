package main

import (
	"fmt"
	"strings"

	"cf-tool/cmd"
	"cf-tool/config"
	"github.com/docopt/docopt-go"
	"github.com/fatih/color"
	"github.com/k0kubun/go-ansi"
)

const version = "v0.8.2"

func main() {
	usage := strings.Replace(Usage, `$%version%$`, version, 1)

	args, _ := docopt.Parse(usage, nil, true, fmt.Sprintf("Codeforces Tool (cf) %v", version), false)
	args[`{version}`] = version
	color.Output = ansi.NewAnsiStdout()
	config.Init()
	err := cmd.Eval(args)
	if err != nil {
		color.Red(err.Error())
	}
	color.Unset()
}
