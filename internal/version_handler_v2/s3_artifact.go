package version_handler_v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nsbno/terraform-provider-vy/internal/aws_auth"
)

type S3Artifact struct {
	GitHubRepositoryName string `json:"github_repository_name"`
	WorkingDirectory     string `json:"working_directory"`
	URI                  string `json:"uri"`
	Store                string `json:"store"`
	Path                 string `json:"path"`
	Version              string `json:"version"`
	GitSha               string `json:"git_sha"`
}

func (c Client) ReadS3Artifact(githubRepositoryName string, workingDirectory string, s3artifact *S3Artifact) error {
	var url string
	if workingDirectory != "" {
		splitWorkingDirectory := strings.Split(workingDirectory, "/")
		lastPart := splitWorkingDirectory[len(splitWorkingDirectory)-1]

		url = fmt.Sprintf("https://%s/v2/s3/versions/%s/%s", c.BaseUrl, githubRepositoryName, lastPart)
		// If HTTPClient is set (for testing), construct URL without https:// prefix
		if c.HTTPClient != nil {
			url = fmt.Sprintf("http://%s/v2/s3/versions/%s/%s", c.BaseUrl, githubRepositoryName, lastPart)
		}
	} else {
		url = fmt.Sprintf("https://%s/v2/s3/versions/%s", c.BaseUrl, githubRepositoryName)
		// If HTTPClient is set (for testing), construct URL without https:// prefix
		if c.HTTPClient != nil {
			url = fmt.Sprintf("http://%s/v2/s3/versions/%s", c.BaseUrl, githubRepositoryName)
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

		return errors.New(fmt.Sprintf("could not find S3 Artifact. %s", str))
	}

	err = json.NewDecoder(response.Body).Decode(s3artifact)
	if err != nil {
		return err
	}

	return nil
}
