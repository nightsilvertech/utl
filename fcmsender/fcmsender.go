package fcmsender

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	googleFcmApiBaseURL = `https://fcm.googleapis.com/fcm/send`
	headerSeparator     = `|`
	headerLength        = 2
	keyIndex            = 0
	valueIndex          = 1
)

type Notif struct {
	Body     string `json:"body"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

type Notification struct {
	To    string `json:"to"`
	Notif Notif  `json:"notification"`
}

func Send(appKey string, notification Notification) (res string, err error) {
	headers := []string{
		fmt.Sprintf("Authorizationa%skey=%s", headerSeparator, appKey),
	}
	jsonData, err := json.Marshal(notification)
	if err != nil {
		return res, err
	}
	req, err := http.NewRequest(http.MethodPost, googleFcmApiBaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return res, err
	}
	for _, header := range headers {
		keyValue := strings.Split(header, headerSeparator)
		if len(keyValue) == headerLength {
			req.Header.Set(keyValue[keyIndex], keyValue[valueIndex])
		}
	}
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
