package toolbox

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alphagov/pay-cli/pkg/config"
	// @TODO(sfount) move the libs for spinners out of card package
	"github.com/alphagov/pay-cli/pkg/card"
	"github.com/pkg/browser"
	"github.com/urfave/cli/v2"
)

func SearchForInput(input string, environment config.Environment, context *cli.Context, isInteractive bool) error {
	if strings.TrimSpace(input) == "" {
		return errors.New("Search term is required to open Toolbox, see `help` for valid search entities")
	}

	matchedFeature := bestMatchFlags(context)
	if matchedFeature == UNKNOWN {
		matchedFeature = bestMatchInput(input)
	}
	if matchedFeature == UNKNOWN {
		fmt.Printf("Unable to best match search term, specify using flags")
		return nil
	}

	if isInteractive {
		s := card.StartProgress("Got interactive input, delaying for Ledger")
		time.Sleep(1000 * time.Millisecond)
		s.Stop()
	}
	return openFeature(matchedFeature, input, environment)
}

func openFeature(matchedFeature Feature, input string, environment config.Environment) error {
	toolboxBaseURL := fmt.Sprintf("https://toolbox.%s", environment.BaseURL)
	switch matchedFeature {
	case TRANSACTION_ID:
		browser.OpenURL(fmt.Sprintf("%s/transactions/%s", toolboxBaseURL, input))
	case SERVICE_ID:
		browser.OpenURL(fmt.Sprintf("%s/services/%s", toolboxBaseURL, input))
	case GATEWAY_ACCOUNT_ID:
		browser.OpenURL(fmt.Sprintf("%s/gateway_accounts/%s", toolboxBaseURL, input))
	case TRANSACTION_REFERENCE:
		browser.OpenURL(fmt.Sprintf("%s/transactions?reference=%s", toolboxBaseURL, input))
	}
	return nil
}

type Feature int

const (
	UNKNOWN               Feature = 0
	TRANSACTION_ID        Feature = 1
	TRANSACTION_REFERENCE Feature = 2
	SERVICE_ID            Feature = 3
	GATEWAY_ACCOUNT_ID    Feature = 4
	// USER_ID               Feature = 4
)

func bestMatchInput(input string) Feature {
	if len(input) == 26 {
		return TRANSACTION_ID
	}
	if len(input) == 32 {
		return SERVICE_ID
	}
	if numberInput, err := strconv.Atoi(input); err == nil {
		if numberInput < 10000 {
			return GATEWAY_ACCOUNT_ID
		}
	}
	return TRANSACTION_REFERENCE
}

func bestMatchFlags(context *cli.Context) Feature {
	if context.IsSet("transaction") {
		return TRANSACTION_ID
	}
	if context.IsSet("reference") {
		return TRANSACTION_REFERENCE
	}
	if context.IsSet("service") {
		return SERVICE_ID
	}
	if context.IsSet("account") {
		return GATEWAY_ACCOUNT_ID
	}
	return UNKNOWN
}
