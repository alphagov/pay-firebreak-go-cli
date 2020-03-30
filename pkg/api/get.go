package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/alphagov/pay-cli/pkg/config"
)

func GetPayment(id string, environment config.Environment) (Payment, error) {
	var payment Payment

	if strings.TrimSpace(id) == "" {
		return payment, errors.New("Invalid payment ID provided, unable to get payment")
	}

	target := "/v1/payments"
	url := fmt.Sprintf("https://publicapi.%s/%s/%s", environment.BaseURL, target, id)
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", environment.APIKey))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return payment, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return payment, fmt.Errorf("Create payment request returned non-success code %d", res.StatusCode)
	}

	payment.parse(res)
	payment.furnishToolboxURL(environment)
	return payment, nil
}
