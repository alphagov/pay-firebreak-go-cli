package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/alphagov/pay-cli/pkg/config"
	"github.com/tidwall/pretty"
)

type Link struct {
	Href   string `json:"href"`
	Method string `json:"method"`
}

type PaymentLinks struct {
	NextURL    Link `json:"next_url"`
	ToolboxURL Link `json:"toolbox_url"`
}

type Payment struct {
	ID              string       `json:"payment_id"`
	Amount          int          `json:"amount"`
	Reference       string       `json:"reference"`
	Description     string       `json:"description"`
	PaymentProvider string       `json:"payment_provider"`
	Links           PaymentLinks `json:"_links"`
}

type RefundLinks struct {
	Self       Link `json:"self"`
	Payment    Link `json:"payment"`
	ToolboxURL Link `json:"toolbox_url"`
}

type Refund struct {
	ID     string      `json:"refund_id"`
	Amount int         `json:"amount"`
	Status string      `json:"status"`
	Links  RefundLinks `json:"_links"`
}

// chainOut outputs the result of the response to stdout depending on the called context
func (payment *Payment) ChainOut(shouldOutputNextURL bool) error {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return err
	}

	// context is sending data to a pipe
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		var output string
		if shouldOutputNextURL {
			output = payment.Links.NextURL.Href
		} else {
			output = payment.ID
		}
		fmt.Print(output)
	} else {
		// context is directly back to terminal
		jsonBytes, err := json.Marshal(payment)
		if err != nil {
			return err
		}
		fmt.Printf("%s", pretty.Color(pretty.Pretty(jsonBytes), nil))
	}
	return nil
}

// chainOut outputs the result of the response to stdout depending on the called context
func (refund *Refund) ChainOut() error {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return err
	}

	// context is sending data to a pipe
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		fmt.Print(refund.ID)
	} else {
		// context is directly back to terminal
		jsonBytes, err := json.Marshal(refund)
		if err != nil {
			return err
		}
		fmt.Printf("%s", pretty.Color(pretty.Pretty(jsonBytes), nil))
	}
	return nil
}

func (payment *Payment) parse(res *http.Response) {
	json.NewDecoder(res.Body).Decode(payment)
}

func (refund *Refund) parse(res *http.Response) {
	json.NewDecoder(res.Body).Decode(refund)
}

func (payment *Payment) furnishToolboxURL(environment config.Environment) {
	payment.Links.ToolboxURL = Link{
		Href:   fmt.Sprintf("https://toolbox.%s/transactions/%s", environment.BaseURL, payment.ID),
		Method: "GET",
	}
}

func (refund *Refund) furnishToolboxURL(environment config.Environment) {
	refund.Links.ToolboxURL = Link{
		Href:   fmt.Sprintf("https://toolbox.%s/transactions/%s", environment.BaseURL, refund.ID),
		Method: "GET",
	}
}
