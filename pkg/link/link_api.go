package link

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/alphagov/pay-cli/pkg/config"

	"github.com/logrusorgru/aurora"
	"golang.org/x/crypto/ssh/terminal"
)

const defaultBaseURL = "payments.service.gov.uk"
const payAPIKeyLength = 58

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

	if len(apiKey) != payAPIKeyLength {
		return "", fmt.Errorf("GOV.UK Pay API keys are %d characters long, please provide valid API key", payAPIKeyLength)
	}

	fmt.Printf("Your API key is: %s\n", redact(apiKey))

	return string(apiKey), nil
}

func getConfigureBaseURL() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter the base URL for this link [default: %s]: ", aurora.Bold(aurora.Cyan(defaultBaseURL)))
	baseURL, err := reader.ReadString('\n')

	if err != nil {
		return "", err
	}

	baseURL = strings.TrimSpace(baseURL)

	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return baseURL, nil
}

func redact(target string) string {
	var builder strings.Builder
	builder.WriteString(target[0:6])
	builder.WriteString(strings.Repeat("*", len(target)-10))
	builder.WriteString(target[len(target)-4:])
	return builder.String()
}
