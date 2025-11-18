package enroll_account

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nsbno/terraform-provider-vy/internal/aws_auth"
)

type EnvironmentAccount struct {
	AccountId      string `json:"account_id"`
	OwnerAccountId string `json:"owner_account_id"`
}

type EnvironmentAccountCreateRequest struct {
	OwnerAccountId string `json:"owner_account_id"`
}

func (c Client) RegisterEnvironmentAccount(ownerAccountId string) (*EnvironmentAccount, error) {
	var data bytes.Buffer

	err := json.NewEncoder(&data).Encode(EnvironmentAccountCreateRequest{OwnerAccountId: ownerAccountId})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://%s/environment_accounts", c.BaseUrl),
		&data,
	)
	if err != nil {
		return nil, err
	}

	response, err := aws_auth.SignedRequest(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 201 {
		defer response.Body.Close()

		str, _ := io.ReadAll(response.Body)

		return nil, errors.New(fmt.Sprintf("Could not add account. %s", str))
	}

	var createdAccount *EnvironmentAccount
	err = json.NewDecoder(response.Body).Decode(&createdAccount)
	if err != nil {
		return nil, err
	}

	return createdAccount, nil
}

func (c Client) ReadEnvironmentAccount(account *EnvironmentAccount) error {
	request, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://%s/environment_accounts", c.BaseUrl),
		nil,
	)
	if err != nil {
		return err
	}

	response, err := aws_auth.SignedRequest(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		str, _ := io.ReadAll(response.Body)

		return errors.New(fmt.Sprintf("could not read environment account. %s", str))
	}

	err = json.NewDecoder(response.Body).Decode(account)
	if err != nil {
		return err
	}

	return nil
}

func (c Client) DeleteEnvironmentAccount() error {
	request, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("https://%s/environment_accounts", c.BaseUrl),
		nil,
	)
	if err != nil {
		return err
	}

	response, err := aws_auth.SignedRequest(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		str, _ := io.ReadAll(response.Body)

		return errors.New(fmt.Sprintf("could not delete account. %s", str))
	}

	return nil
}
