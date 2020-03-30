package cmd

import (
	"github.com/alphagov/pay-cli/pkg/link"

	"github.com/urfave/cli/v2"
)

// Link is the entry point for the `link` command package
func Link() *cli.Command {
	return &cli.Command{
		Name:   "link",
		Flags:  GlobalFlags,
		Action: runLinkCmd,
		Usage:  "Configure CLI application with GOV.UK Pay API relative to the current environment",
		Before: SetGlobalFlags,
	}
}

func runLinkCmd(context *cli.Context) error {
	Environment.Name = GetGlobalFlag("environment", context)
	return link.ConfigureAPI(Environment)
}
