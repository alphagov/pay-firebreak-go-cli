package cmd

import (
	"github.com/alphagov/pay-cli/pkg/card"
	"github.com/urfave/cli/v2"
)

func Card() *cli.Command {
	return &cli.Command{
		Name:      "card",
		Usage:     "Process a card payment, valid contexts are next_url and payment ID",
		Flags:     GlobalFlags,
		Action:    runCardCmd,
		ArgsUsage: "context",
		Before:    SetGlobalFlags,
	}
}

func runCardCmd(context *cli.Context) error {
	nextURL, err := GetArgOrStdin(context)
	if err != nil {
		return err
	}

	// @TODO(sfount) move to helper method
	Environment.Name = GetGlobalFlag("environment", context)
	apiKey, err := Environment.GetAPIKey()
	if err != nil {
		return err
	}
	baseURL, err := Environment.GetBaseURL()
	if err != nil {
		return err
	}
	Environment.APIKey = apiKey
	Environment.BaseURL = baseURL
	return card.MakeCardPayment(nextURL, Environment)
}
