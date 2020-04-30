package cmd

import (
	"github.com/alphagov/pay-cli/pkg/ci"
	"github.com/urfave/cli/v2"
)

// CI is the top level command for the GOV.UK Pay ci commands
func CI() *cli.Command {
	return &cli.Command{
		Name:   "ci",
		Usage:  "GOV.UK Pay CI convenience methods",
		Flags:  GlobalFlags,
		Before: SetGlobalFlags,
		Subcommands: []*cli.Command{
			CompareWithJenkins(),
		},
	}
}

func CompareWithJenkins() *cli.Command {
	return &cli.Command{
		Name:  "compare",
		Usage: "Compare the ci results between Concourse and Jenkins",
		Flags: append(
			[]cli.Flag{
				&cli.IntFlag{
					Name:    "number-of-prs",
					Aliases: []string{"n"},
					Value:   10,
					Usage:   "Number of prs to compare beginning with the latest",
				},
				&cli.StringFlag{
					Name:    "repo",
					Aliases: []string{"r"},
					Value:   "all",
					Usage:   "Repo to compare pr outcome, defaults to 'all'",
				},
			},
			GlobalFlags...,
		),
		Before: SetGlobalFlags,
		Action: runCompareWithJenkins,
	}
}

func runCompareWithJenkins(context *cli.Context) error {
	return ci.CompareWithJenkins(context.String("repo"), context.Int("number-of-prs"))
}
