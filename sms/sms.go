package sms

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	vonageApiBaseURL = `https://rest.nexmo.com/sms/json`
)

func Send(message string, to string, from string, apiKey string, apiSecret string, pretend bool) (res string, err error) {
	if pretend {
		return "SENT", nil
	} else {
		params := url.Values{}
		params.Add("from", from)
		params.Add("text", message)
		params.Add("to", to)
		params.Add("api_key", apiKey)
		params.Add("api_secret", apiSecret)
		body := strings.NewReader(params.Encode())

		req, err := http.NewRequest(http.MethodPost, vonageApiBaseURL, body)
		if err != nil {
			return res, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return res, err
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		if resp.StatusCode > 201 {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			msg := fmt.Sprintf("http status code %d error response %s", resp.StatusCode, string(bodyBytes))
			return res, errors.New(msg)
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return res, err
		}
		return string(bodyBytes), nil
	}
}
