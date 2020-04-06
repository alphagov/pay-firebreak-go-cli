package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/alphagov/pay-cli/pkg/config"
)

type RefundPaymentRequest struct {
	Amount int `json:"amount"`
}

func RefundPayment(id string, amount int, environment config.Environment) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("Invalid payment ID provided, unable to refund payment")
	}

	target := fmt.Sprintf("v1/payments/%s/refunds", id)
	url := fmt.Sprintf("https://publicapi.%s/%s", environment.BaseURL, target)
	request := RefundPaymentRequest{
		Amount: amount,
	}
	payload := strings.NewReader(request.format())
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", environment.APIKey))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 && res.StatusCode != 202 {
		return fmt.Errorf("Refund payment request returned non-success code %d", res.StatusCode)
	}

	var refund Refund
	refund.parse(res)
	refund.furnishToolboxURL(environment)
	refund.ChainOut()
	return nil
}

func (refundRequest *RefundPaymentRequest) format() string {
	result, _ := json.Marshal(refundRequest)
	return string(result)
}
