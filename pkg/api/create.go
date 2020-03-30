package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alphagov/pay-cli/pkg/config"
	"github.com/google/uuid"
)

type CreatePaymentRequest struct {
	Amount      int
	Reference   string
	Description string
	ReturnURL   string
}

func CreatePayment(environment config.Environment, amount int, shouldOutputNextURL bool) error {
	target := "v1/payments"
	url := fmt.Sprintf("https://publicapi.%s/%s", environment.BaseURL, target)
	paymentAmount := 2000
	if amount != 0 {
		paymentAmount = amount
	}
	request := CreatePaymentRequest{
		Amount:      paymentAmount,
		Reference:   uuid.New().String(),
		Description: fmt.Sprintf("Pay CLI generated payment %s", time.Now().Format(time.Stamp)),
		ReturnURL:   fmt.Sprintf("https://%s", environment.BaseURL),
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

	if res.StatusCode != 201 {
		return fmt.Errorf("Create payment request returned non-success code %d", res.StatusCode)
	}

	var payment Payment
	payment.parse(res)
	payment.furnishToolboxURL(environment)
	payment.ChainOut(shouldOutputNextURL)
	return nil
}

func (paymentRequest *CreatePaymentRequest) format() string {
	return fmt.Sprintf(
		"{\n\t\"amount\": %d,\n\t\"reference\": \"%s\",\n\t\"description\": \"%s\",\n\t\"return_url\": \"%s\"\n}",
		paymentRequest.Amount,
		paymentRequest.Reference,
		paymentRequest.Description,
		paymentRequest.ReturnURL,
	)
}
