package enroll_account

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Account struct {
	AccountId    string `json:"account_id"`
	SlackChannel string `json:"slack_channel"`
}

type CreateAccountRequest struct {
	SlackChannel string `json:"slack_channel"`
}

func (c Client) CreateAccount(slackChannel string) (*Account, error) {
	var data bytes.Buffer

	err := json.NewEncoder(&data).Encode(CreateAccountRequest{SlackChannel: slackChannel})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://%s/accounts", c.BaseUrl),
		&data,
	)
	if err != nil {
		return nil, err
	}

	response, err := signedRequest(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 201 {
		defer response.Body.Close()

		str, _ := io.ReadAll(response.Body)

		return nil, errors.New(fmt.Sprintf("Could not add account. %s", str))
	}

	var createdAccount *Account
	err = json.NewDecoder(response.Body).Decode(&createdAccount)
	if err != nil {
		return nil, err
	}

	return createdAccount, nil
}

func (c Client) ReadAccount(account *Account) error {
	request, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://%s/accounts", c.BaseUrl),
		nil,
	)
	if err != nil {
		return err
	}

	response, err := signedRequest(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		str, _ := io.ReadAll(response.Body)

		return errors.New(fmt.Sprintf("could not read deployment account. %s", str))
	}

	err = json.NewDecoder(response.Body).Decode(account)
	if err != nil {
		return err
	}

	return nil
}

func (c Client) DeleteAccount() error {
	request, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("https://%s/accounts", c.BaseUrl),
		nil,
	)
	if err != nil {
		return err
	}

	response, err := signedRequest(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		str, _ := io.ReadAll(response.Body)

		return errors.New(fmt.Sprintf("could not delete resource. %s", str))
	}

	return nil
}
