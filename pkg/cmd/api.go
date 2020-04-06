package cmd

import (
	"errors"

	"github.com/alphagov/pay-cli/pkg/api"
	"github.com/urfave/cli/v2"
)

// API is the top level command for the GOV.UK Pay api endpoints
func API() *cli.Command {
	return &cli.Command{
		Name:   "api",
		Usage:  "GOV.UK Pay API convenience methods",
		Flags:  GlobalFlags,
		Before: SetGlobalFlags,
		Subcommands: []*cli.Command{
			Get(),
			Create(),
			Refund(),
		},
	}
}

// Create is an accessor for the create payment endpoint
func Create() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create new payment",
		Flags: append(
			[]cli.Flag{
				&cli.BoolFlag{
					Name:    "output-next-url",
					Aliases: []string{"n"},
					Usage:   "Output the next_url on payment create, will default to the external ID",
				},
				&cli.IntFlag{
					Name:    "amount",
					Aliases: []string{"a"},
					Usage:   "Amount for payment in pence",
				},
				&cli.StringFlag{
					Name:    "language",
					Aliases: []string{"l"},
					Value:   "en",
					Usage:   "Language of the payment",
				},
			},
			GlobalFlags...,
		),
		Before: SetGlobalFlags,
		Action: runCreateCmd,
	}
}

func runCreateCmd(context *cli.Context) error {
	shouldOutputNextURL := context.Bool("output-next-url")
	Environment.Name = GetGlobalFlag("environment", context)
	err := Environment.Init()
	if err != nil {
		return err
	}
	amount := context.Int("amount")
	language := context.String("language")
	return api.CreatePayment(Environment, amount, language, shouldOutputNextURL)
}

func Get() *cli.Command {
	return &cli.Command{
		Name:   "get",
		Usage:  "Get one payment",
		Flags:  GlobalFlags,
		Before: SetGlobalFlags,
		Action: runGetCmd,
	}
}

func runGetCmd(context *cli.Context) error {
	Environment.Name = GetGlobalFlag("environment", context)
	err := Environment.Init()
	if err != nil {
		return err
	}
	ID, err := GetArgOrStdin(context)
	if err != nil {
		return err
	}
	payment, err := api.GetPayment(ID, Environment)
	if err != nil {
		return err
	}
	payment.ChainOut(false)
	return nil
}

func Refund() *cli.Command {
	return &cli.Command{
		Name:  "refund",
		Usage: "Refund payment",
		Flags: append(
			[]cli.Flag{
				&cli.IntFlag{
					Name:    "amount",
					Aliases: []string{"a"},
					Usage:   "Amount to be refunded in pence",
				},
			},
			GlobalFlags...,
		),
		Before: SetGlobalFlags,
		Action: runRefundCmd,
	}
}

func runRefundCmd(context *cli.Context) error {
	Environment.Name = GetGlobalFlag("environment", context)
	err := Environment.Init()
	if err != nil {
		return err
	}
	ID, err := GetArgOrStdin(context)
	if err != nil {
		return err
	}
	amount := context.Int("amount")
	if amount == 0 {
		return errors.New("Amount (--amount, -a) is required to refund a payment")
	}
	return api.RefundPayment(ID, amount, Environment)
}
