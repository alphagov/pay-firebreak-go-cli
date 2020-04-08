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

type environment struct{
  name, baseUrl string
}

var ENVIRONMENTS = []environment {
  environment{
    name: "production",
    baseUrl: "payments.service.gov.uk",
  },
  environment{
    name: "staging",
    baseUrl: "staging.payments.service.gov.uk",
  },
  environment{
    name: "test",
    baseUrl: "pymnts.uk",
  },
  environment{
    name: "custom",
    baseUrl: "",
  },
}

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
  fmt.Printf("Choose the environment (enter a number between 0 and %d):\n", len(ENVIRONMENTS) - 1)
  for index, environment := range ENVIRONMENTS {
    fmt.Printf("%d %s\n", index, environment.name)
  }

  var userSelection int
	_, err := fmt.Scanf("%d", &userSelection)
  if err != nil {
    return "", err
  }

  if userSelection >= len(ENVIRONMENTS) {
    return "", errors.New("Invalid environment selection")
  }

  baseURL, err := parseUserBaseURLSelection(userSelection)
  if err != nil {
    return "", err
  }

	return baseURL, nil
}

func parseUserBaseURLSelection(option int) (string, error) {
  environment := ENVIRONMENTS[option]
  if environment.name == "custom" {
    return  getCustomBaseURL()
  }
  return environment.baseUrl, nil
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
