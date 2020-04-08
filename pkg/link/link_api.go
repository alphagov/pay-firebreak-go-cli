package link

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"github.com/alphagov/pay-cli/pkg/config"
	"golang.org/x/crypto/ssh/terminal"
)

const PRODUCTION_BASE_URL = "payments.service.gov.uk"
const STAGING_BASE_URL = "staging.payments.service.gov.uk"
const TEST_BASE_URL = "pymnt.uk"
const PAY_API_KEY_LENGTH = 58

// ConfigureAPI links the CLI to the users GOV.UK Pay API configuration
func ConfigureAPI(environment config.Environment) error {
	apiKey, err := getConfigureAPIKey()
	if err != nil {
		return err
	}
	baseURL, err := getConfigureBaseURL()
	if err != nil {
		return err
	}

	environment.APIKey = apiKey
	environment.BaseURL = baseURL
	return environment.CreateEnvironment()
}

func getConfigureAPIKey() (string, error) {
	fmt.Print("Enter your API key: ")

	apiKeyBuffer, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", err
	}
	fmt.Print("\n")

	apiKey := strings.TrimSpace(string(apiKeyBuffer))

	if apiKey == "" {
		return "", errors.New("Empty API key, please provide valid API key")
	}

	if len(apiKey) != PAY_API_KEY_LENGTH {
		return "", fmt.Errorf("GOV.UK Pay API keys are %d characters long, please provide valid API key", PAY_API_KEY_LENGTH)
	}

	fmt.Printf("Your API key is: %s\n", redact(apiKey))

	return string(apiKey), nil
}

func getConfigureBaseURL() (string, error) {
	reader := bufio.NewReader(os.Stdin)
  fmt.Printf(
  `Which base url should be used? Enter a number:
  1. production
  2. staging
  3. test
  4. custom`)
  fmt.Printf("\n")

	userSelection, err := reader.ReadString('\n')

	if err != nil {
		return "", err
	}

  baseURL, err := parseUserBaseURLSelection(userSelection)
  if err != nil {
    return "", err
  }

	return baseURL, nil
}

func parseUserBaseURLSelection(option string) (string, error) {
  option = strings.TrimSpace(option)
  switch option {
  case "1":
    return PRODUCTION_BASE_URL, nil
  case "2":
    return STAGING_BASE_URL, nil
  case "3":
    return TEST_BASE_URL, nil
  case "4":
    return getCustomBaseURL()
  default:
    return getCustomBaseURL()
  }
}

func getCustomBaseURL() (string, error) {
	reader := bufio.NewReader(os.Stdin)
  fmt.Printf("Enter base url:\n")

	baseURL, err := reader.ReadString('\n')

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(baseURL), nil
}

func redact(target string) string {
	var builder strings.Builder
	builder.WriteString(target[0:6])
	builder.WriteString(strings.Repeat("*", len(target)-10))
	builder.WriteString(target[len(target)-4:])
	return builder.String()
}
