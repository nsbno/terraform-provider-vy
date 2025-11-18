package version_handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nsbno/terraform-provider-vy/internal/aws_auth"
)

type Version struct {
	ApplicationName string `json:"application_name"`
	URI             string `json:"uri"`
	Store           string `json:"store"`
	Path            string `json:"path"`
	Version         string `json:"version"`
}

func (c Client) ReadVersion(application_name string, version *Version) error {
	request, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://%s/versions/%s", c.BaseUrl, application_name),
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

		return errors.New(fmt.Sprintf("could not read artifact version. %s", str))
	}

	err = json.NewDecoder(response.Body).Decode(version)
	if err != nil {
		return err
	}

	return nil
}
