package cmd

import (
	"os"

	"github.com/alphagov/pay-cli/pkg/toolbox"
	"github.com/urfave/cli/v2"
)

// Toolbox is a generic search link to toolbox endpoints
func Toolbox() *cli.Command {
	return &cli.Command{
		Name: "toolbox",
		Flags: append(
			[]cli.Flag{
				&cli.BoolFlag{
					Name:    "transaction",
					Aliases: []string{"t"},
					Usage:   "Specify transaction external id as the toolbox entity type",
				},
				&cli.BoolFlag{
					Name:    "reference",
					Aliases: []string{"r"},
					Usage:   "Specify transaction reference as the toolbox entity type",
				},
				&cli.BoolFlag{
					Name:    "service",
					Aliases: []string{"s"},
					Usage:   "Specify service external as the toolbox entity type",
				},
				&cli.BoolFlag{
					Name:    "account",
					Aliases: []string{"a"},
					Usage:   "Specify gateway account id as the toolbox entity type",
				},
			},
			GlobalFlags...,
		),
		Usage:  "Open entities in Toolbox relative to the current environment",
		Action: runToolboxCmd,
		Before: SetGlobalFlags,
	}
}

func runToolboxCmd(context *cli.Context) error {
	Environment.Name = GetGlobalFlag("environment", context)
	err := Environment.Init()
	if err != nil {
		return err
	}
	input, err := GetArgOrStdin(context)
	if err != nil {
		return err
	}
	fi, err := os.Stdin.Stat()
	if err != nil {
		return err
	}
	isInteractive := (fi.Mode() & os.ModeCharDevice) == 0

	return toolbox.SearchForInput(input, Environment, context, isInteractive)
}
