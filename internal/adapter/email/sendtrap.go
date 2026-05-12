package email

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type sendtrap struct {
	apiKey string
	apiUrl string
}

type contact struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type payload struct {
	From     contact   `json:"from"`
	To       []contact `json:"to"`
	Subject  string    `json:"subject"`
	Text     string    `json:"text"`
	Html     *string   `json:"html,omitempty"`
	CC       *contact  `json:"cc,omitempty"`
	BCC      *contact  `json:"bcc,omitempty"`
	ReplayTo *contact  `json:"reply_to,omitempty"`
}

func NewSendtrap(apiKey, apiUrl string) *sendtrap {

	return &sendtrap{
		apiKey: apiKey,
		apiUrl: apiUrl,
	}
}

func (s *sendtrap) Send(name string, email string, subject string, html string) error {

	p := &payload{
		From: contact{
			Email: "contato@nohats.net.br",
			Name:  "Nohats",
		},
		To: []contact{
			{
				Email: email,
				Name:  name,
			},
		},
		Subject: subject,
		Text:    subject,
		Html:    &html,
	}

	pJson, _ := json.Marshal(p)

	client := &http.Client{}

	request, err := http.NewRequest("POST", s.apiUrl+"/send", strings.NewReader(string(pJson[:])))

	if err != nil {
		return err
	}

	request.Header.Add("Api-Token", s.apiKey)
	request.Header.Add("Content-Type", "application/json")

	response, err := client.Do(request)

	if err != nil {
		return fmt.Errorf("error on call external API: %w", err)
	}

	defer response.Body.Close()

	_, err = io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	return nil
}
