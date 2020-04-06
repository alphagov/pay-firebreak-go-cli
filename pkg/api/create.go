package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/alphagov/pay-cli/pkg/config"
	"github.com/google/uuid"
)

type CreatePaymentRequest struct {
	Amount      int    `json:"amount"`
	Reference   string `json:"reference"`
	Description string `json:"description"`
	ReturnURL   string `json:"return_url"`
	Language    string `json:"language"`
}

func CreatePayment(environment config.Environment, request CreatePaymentRequest, shouldOutputNextURL bool) error {
	target := "v1/payments"
	url := fmt.Sprintf("https://publicapi.%s/%s", environment.BaseURL, target)
	defaultValues := CreatePaymentRequest{
		Amount:      2000,
		Reference:   uuid.New().String(),
		Description: fmt.Sprintf("Pay CLI generated payment %s", time.Now().Format(time.Stamp)),
		ReturnURL:   fmt.Sprintf("https://%s", environment.BaseURL),
		Language:    "en",
	}
	err := Replace(defaultValues, &request)
	if err != nil {
		return err
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
	result, _ := json.Marshal(paymentRequest)
	return string(result)
}

func IsZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

// Replace accepts struct, pointer - will replace values at pointer with struct values
// if they are defined
func Replace(a, b interface{}) error {
	vb := reflect.ValueOf(b).Elem()
	for i := 0; i < vb.NumField(); i++ {
		field := vb.Field(i)
		if field.CanInterface() && IsZeroOfUnderlyingType(field.Interface()) {
			name := vb.Type().Field(i).Name
			fa := reflect.ValueOf(a).FieldByName(name)
			if fa.IsValid() {
				if field.CanSet() {
					field.Set(fa)
				}
			}
		}
	}
	return nil
}
