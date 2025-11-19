package version_handler_v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nsbno/terraform-provider-vy/internal/aws_auth"
)

type ECRVersion struct {
	ECRRepositoryName string `json:"ecr_repository_name"`
	URI               string `json:"uri"`
	Store             string `json:"store"`
	Path              string `json:"path"`
	Version           string `json:"version"`
	GitSha            string `json:"git_sha"`
}

func (c Client) ReadECRImage(ecrRepositoryName string, ecrVersion *ECRVersion) error {
	url := fmt.Sprintf("https://%s/v2/ecr/versions/%s", c.BaseUrl, ecrRepositoryName)

	// If HTTPClient is set (for testing), construct URL without https:// prefix
	if c.HTTPClient != nil {
		url = fmt.Sprintf("http://%s/v2/ecr/versions/%s", c.BaseUrl, ecrRepositoryName)
	}

	request, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return err
	}

	var response *http.Response
	if c.HTTPClient != nil {
		// Use HTTP client for testing
		response, err = c.HTTPClient.Do(request)
	} else {
		// Use AWS signed request for production
		response, err = aws_auth.SignedRequest(request)
	}

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		str, _ := io.ReadAll(response.Body)

		return errors.New(fmt.Sprintf("could not find ECR Image. %s", str))
	}

	err = json.NewDecoder(response.Body).Decode(ecrVersion)
	if err != nil {
		return err
	}

	return nil
}
