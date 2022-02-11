package fcmsender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Notification struct {
	Body     string `json:"body"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Tag      string `json:"tag"`
}

type Data struct {
	JsonExtraData string `json:"jsonExtraData"`
}

type Payload struct {
	ServerKey    string       `json:"-"`
	To           string       `json:"to"`
	Notification Notification `json:"notification"`
	Data         Data         `json:"data"`
}

type Result struct {
	MessageId string `json:"message_id"`
}

type FcmResponse struct {
	MulticastId  int64   `json:"multicast_id"`
	Success      int     `json:"success"`
	Failure      int     `json:"failure"`
	CanonicalIds int      `json:"canonical_ids"`
	Results      []Result `json:"results"`
}

func Send(payload Payload) (res FcmResponse, err error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return res, err
	}
	req, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return res, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("key=%s", payload.ServerKey))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(bodyBytes, &res)
	if err != nil {
		return res, err
	}
	return res, nil
}
