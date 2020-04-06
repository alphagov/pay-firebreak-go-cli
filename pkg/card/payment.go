package card

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/alphagov/pay-cli/pkg/api"
	"github.com/alphagov/pay-cli/pkg/config"
	"github.com/briandowns/spinner"
	"github.com/gorilla/schema"
	"github.com/logrusorgru/aurora"
	"golang.org/x/net/html"
)

type PostPaymentRequest struct {
	PaymentID       string `schema:"chargeId"`
	CardNumber      string `schema:"cardNo"`
	CardExpiryMonth string `schema:"expiryMonth"`
	CardExpiryYear  string `schema:"expiryYear"`
	CardCVC         string `schema:"cvc"`
	CardHolderName  string `schema:"cardholderName"`
	AddressLineOne  string `schema:"addressLine1"`
	AddressCity     string `schema:"addressCity"`
	AddressCountry  string `schema:"addressCountry"`
	AddressPostCode string `schema:"addressPostcode"`
	Email           string `schema:"email"`
	CSRF            string `schema:"csrfToken"`
}

type PostConfirmRequest struct {
	PaymentID string `schema:"chargeId"`
	CSRF      string `schema:"csrfToken"`
}

func MakeCardPayment(input string, environment config.Environment) error {
	if strings.TrimSpace(input) == "" {
		return errors.New("context is required to process a card payment, valid contexts are next_url and payment ID")
	}
	nextURL, err := getNextURLFromInput(input, environment)
	if err != nil {
		return err
	}
	return processCardPayment(nextURL, environment)
}

// @TODO(sfount) this might be considered a hack -- talk to someone to sense check this
// getNextURLFromInput parses a generic string input and returns a next url if it finds either a payment ID or a next url
func getNextURLFromInput(input string, environment config.Environment) (string, error) {
	// assume a payment ID has been provided directly
	if len(input) == 26 {
		// @TODO(sfount) separating progress from actual methods would enable them to become generic if needed
		s := StartProgress(fmt.Sprintf("Fetching next url for payment %s", aurora.Bold(aurora.Cyan(input))))
		payment, err := api.GetPayment(input, environment)
		ProgressSuccess(s)
		if err != nil {
			ProgressFail(s)
			return "", err
		}
		return payment.Links.NextURL.Href, nil
	} else if strings.Contains(input, "http") {
		// assume a next url has been provided directly
		return input, nil
	} else {
		return "", errors.New("Unrecognised input, unable to process card payment")
	}
}

type CardPaymentProcess struct {
	Environment  config.Environment
	NextURL      string
	CSRF         string
	PaymentID    string
	AuthAttempts int
}

func processCardPayment(nextURL string, environment config.Environment) error {
	willWrite, _ := ShouldWriteProgress()

	// cookies are required for frontend authenticating each request
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	client := http.Client{
		Jar: cookieJar,
	}
	process := CardPaymentProcess{
		NextURL:      nextURL,
		Environment:  environment,
		AuthAttempts: 0,
	}
	err = process.getCardDetailsPage(client)
	if err != nil {
		return err
	}
	err = process.postCardDetails(client)
	if err != nil {
		return err
	}
	err = process.getConfirmPage(client)
	if err != nil {
		return err
	}
	err = process.postConfirm(client)
	if err != nil {
		return err
	}
	if !willWrite {
		fmt.Print(process.PaymentID)
	} else {
		fmt.Printf("> Completed card payment %s", aurora.Bold(fmt.Sprintf("https://toolbox.%s/transactions/%s\n", environment.BaseURL, process.PaymentID)))
	}
	return nil
}

func (process *CardPaymentProcess) getCardDetailsPage(client http.Client) error {
	req, err := http.NewRequest("GET", process.NextURL, nil)
	if err != nil {
		return err
	}
	s := StartProgress("Loading card details page")
	res, err := client.Do(req)
	if err != nil {
		ProgressFail(s)
		return err
	}
	defer res.Body.Close()

	document, err := html.Parse(res.Body)
	if err != nil {
		return err
	}

	csrfNode := GetElementById(document, "csrf")
	csrfToken, csrfFound := GetAttribute(csrfNode, "value")

	if !csrfFound {
		return errors.New("Unable to parse CSRF token from card details page")
	}

	paymentID, err := ParsePaymentIDFromCardDetailsPage(document)
	if err != nil {
		ProgressFail(s)
		return err
	}
	ProgressSuccess(s)

	// @TODO(sfount) question side effects returning struct
	process.CSRF = csrfToken
	process.PaymentID = paymentID
	return nil
}

func (process *CardPaymentProcess) getConfirmPage(client http.Client) error {
	url := fmt.Sprintf("https://www.%s/card_details/%s/confirm", process.Environment.BaseURL, process.PaymentID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	s := StartProgress(fmt.Sprintf("Loading confirm page (%d)", process.AuthAttempts+1))
	process.AuthAttempts = process.AuthAttempts + 1
	res, err := client.Do(req)
	if err != nil {
		ProgressFail(s)
		return err
	}
	defer res.Body.Close()

	document, err := html.Parse(res.Body)
	if err != nil {
		return err
	}
	csrfNode := GetElementById(document, "csrf")
	csrfToken, csrfFound := GetAttribute(csrfNode, "value")

	if !csrfFound {
		ProgressFail(s)
		if process.AuthAttempts < 3 {
			time.Sleep(500 * time.Millisecond)
			return process.getConfirmPage(client)
		}
		return errors.New("Unable to parse CSRF token from confirmation page")
	}
	ProgressSuccess(s)
	// @TODO(sfount) question side effects returning struct
	process.CSRF = csrfToken
	return nil
}

// post card details doesn't work with Worldpay 3ds enabled accounts
func (process *CardPaymentProcess) postCardDetails(client http.Client) error {
	url := fmt.Sprintf("https://www.%s/card_details/%s", process.Environment.BaseURL, process.PaymentID)
	// @TODO(sfount) allow post params to be overriden by CLI flags
	postPaymentRequest := PostPaymentRequest{
		PaymentID:       process.PaymentID,
		CSRF:            process.CSRF,
		CardNumber:      "4242424242424242",
		CardExpiryMonth: "01",
		CardExpiryYear:  "2030",
		CardHolderName:  "Pay CLI User",
		CardCVC:         "123",
		AddressLineOne:  "10 Whitechapel High St",
		AddressCity:     "London",
		AddressCountry:  "GB",
		AddressPostCode: "E18QS",
		Email:           "pay@cli.gov.uk",
	}

	err, form := postPaymentRequest.format()
	if err != nil {
		return err
	}

	s := StartProgress("Submitting card details")
	res, err := client.PostForm(url, form)
	if err != nil {
		ProgressFail(s)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		ProgressFail(s)
		return fmt.Errorf("Post card details returned non-success status code %d", res.StatusCode)
	}
	ProgressSuccess(s)
	return nil
}

func (process *CardPaymentProcess) postConfirm(client http.Client) error {
	redirectClient := http.Client{
		Jar: client.Jar,

		// return the successful redirect in favour of following it - invalid return URLs shouldn't block this process
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	url := fmt.Sprintf("https://www.%s/card_details/%s/confirm", process.Environment.BaseURL, process.PaymentID)

	postConfirmRequest := PostConfirmRequest{
		PaymentID: process.PaymentID,
		CSRF:      process.CSRF,
	}
	err, form := postConfirmRequest.format()
	if err != nil {
		return err
	}

	s := StartProgress("Submitting confirm payment")
	res, err := redirectClient.PostForm(url, form)
	if err != nil {
		ProgressFail(s)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 303 {
		ProgressFail(s)
		return fmt.Errorf("Post confirm payment returned non-success status code %d", res.StatusCode)
	}
	ProgressSuccess(s)
	return nil
}

func ParsePaymentIDFromCardDetailsPage(document *html.Node) (string, error) {
	paymentForm := GetElementById(document, "card-details")
	actionPath, actionPathFound := GetAttribute(paymentForm, "action")
	if !actionPathFound {
		return "", errors.New("Unable to parse card details from POST action")
	}
	return strings.Split(actionPath, "/")[2], nil
}

func (postPaymentRequest *PostPaymentRequest) format() (error, url.Values) {
	encoder := schema.NewEncoder()
	form := url.Values{}
	err := encoder.Encode(postPaymentRequest, form)
	return err, form
}

func (postConfirmRequest *PostConfirmRequest) format() (error, url.Values) {
	encoder := schema.NewEncoder()
	form := url.Values{}
	err := encoder.Encode(postConfirmRequest, form)
	return err, form
}

// @TODO(sfount) move to utility

func StartProgress(message string) *spinner.Spinner {
	// @TODO(sfount) get from shared utility
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " " + message

	if willWrite, err := ShouldWriteProgress(); err == nil && willWrite {
		s.Start()
	}
	return s
}

func ProgressSuccess(s *spinner.Spinner) {
	s.FinalMSG = aurora.Bold(aurora.Green(">")).String() + s.Suffix + "\n"
	s.Stop()
}

func ProgressFail(s *spinner.Spinner) {
	s.FinalMSG = aurora.Bold(aurora.Red("X")).String() + s.Suffix + "\n"
	s.Stop()
}

func ShouldWriteProgress() (bool, error) {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false, err
	}

	return (fi.Mode() & os.ModeCharDevice) != 0, nil
}
