package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Slack interface {
}

type slack struct {
	url    string
	client *http.Client
}

func NewSlack(url string) Slack {
	return &slack{
		url: url,
		client: &http.Client{
			Timeout: 2 * time.Second,
			Transport: http.RoundTripper(&http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				MaxIdleConns:          10,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}),
		},
	}
}

type message struct {
	Text    string  `json:"text"`
	Channel *string `json:"channel,omitempty"`
}

func (s *slack) SendMessage(channel, text string) error {
	msg := message{Text: text}
	if len(channel) > 0 {
		msg.Channel = &channel
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", s.url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return errors.New(string(b))
	}

	return nil
}
