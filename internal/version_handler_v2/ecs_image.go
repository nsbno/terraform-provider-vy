package version_handler_v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nsbno/terraform-provider-vy/internal/aws_auth"
)

type ECSVersion struct {
	GitHubRepositoryName string `json:"github_repository_name"`
	WorkingDirectory     string `json:"working_directory"`
	GitSha               string `json:"git_sha"`
	Branch               string `json:"branch"`
	ServiceAccountID     string `json:"service_account_id"`
	Region               string `json:"region"`
	ECRRepositoryName    string `json:"ecr_repository_name"`
	ECRRepositoryURI     string `json:"ecr_repository_uri"`
}

func (c Client) ReadECSImage(githubRepositoryName string, workingDirectory string, ecsVersion *ECSVersion) error {
	var url string
	if workingDirectory != "" {
		url = fmt.Sprintf("https://%s/v2/versions/%s/ecs/%s", c.BaseUrl, githubRepositoryName, workingDirectory)
		// If HTTPClient is set (for testing), construct URL without https:// prefix
		if c.HTTPClient != nil {
			url = fmt.Sprintf("http://%s/v2/versions/%s/ecs/%s", c.BaseUrl, githubRepositoryName, workingDirectory)
		}
	} else {
		url = fmt.Sprintf("https://%s/v2/versions/%s/ecs", c.BaseUrl, githubRepositoryName)
		// If HTTPClient is set (for testing), construct URL without https:// prefix
		if c.HTTPClient != nil {
			url = fmt.Sprintf("http://%s/v2/versions/%s/ecs", c.BaseUrl, githubRepositoryName)
		}
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

		return errors.New(fmt.Sprintf("could not find ECS Image. %s", str))
	}

	err = json.NewDecoder(response.Body).Decode(ecsVersion)
	if err != nil {
		return err
	}

	return nil
}
