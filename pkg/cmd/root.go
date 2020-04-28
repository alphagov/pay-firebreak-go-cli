package cmd

import (
	"bufio"
	"io"
	"log"
	"os"

	"github.com/alphagov/pay-cli/pkg/config"

	"github.com/urfave/cli/v2"
)

var Environment config.Environment

// Execute initialises the CLI application and configures all cmd packages
func Execute() {
	config.Init()
	pay := cli.NewApp()

	pay.Usage = "Common CLI utilities used by the GOV.UK Pay team"
	pay.Flags = GlobalFlags
	pay.Before = SetGlobalFlags

	pay.Commands = []*cli.Command{
		Link(),
		API(),
		Card(),
		Toolbox(),
    CI(),
	}

	if err := pay.Run(os.Args); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

// @TODO(sfount) move to utilities file

// workaround for https://github.com/urfave/cli/issues/585
// GlobalFlags list of global flags that should be available to all commands and subcommands
var GlobalFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "environment",
		Value:   "default",
		Aliases: []string{"e"},
		Usage:   "environment profile to use with commands",
	},
}

func SetGlobalFlags(context *cli.Context) error {
	if context.IsSet("environment") {
		context.App.Metadata["environment"] = context.String("environment")
	}
	return nil
}

func GetGlobalFlag(key string, context *cli.Context) string {
	if result, ok := context.App.Metadata[key].(string); ok {
		return result
	}
	return ""
}

// GetArgOrStdin allows a command to read from either the command arguments if called directly or the standard input if piped
func GetArgOrStdin(context *cli.Context) (string, error) {
	var response string
	fi, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		response, err = ReadStringEOFSafe()
		if err != nil {
			return "", err
		}
	} else {
		response = context.Args().Get(0)
	}
	return response, nil
}

func ReadStringEOFSafe() (string, error) {
	var output []byte
	reader := bufio.NewReader(os.Stdin)
	for {
		inputByte, err := reader.ReadByte()
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		output = append(output, inputByte)
	}
	return (string(output)), nil
}
