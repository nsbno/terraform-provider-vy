package version_handler_v2

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/nsbno/terraform-provider-vy/internal/aws_auth"
)

type LambdaArtifact struct {
	GitHubRepositoryName string `json:"github_repository_name"`
	WorkingDirectory     string `json:"working_directory"`
	GitSha               string `json:"git_sha"`
	Branch               string `json:"branch"`
	ServiceAccountID     string `json:"service_account_id"`
	ECRRepositoryName    string `json:"ecr_repository_name"`
	Region               string `json:"region"`
	BucketName           string `json:"bucket_name"`
}

func (c Client) ReadLambdaArtifact(githubRepositoryName string, ecrRepositoryName string, workingDirectory string,
	lambdaArtifact *LambdaArtifact) error {

	protocol := "https://"
	if c.HTTPClient != nil {
		protocol = "http://"
	}

	reqURL := fmt.Sprintf("%s%s/v2/versions/%s/lambda", protocol, c.BaseUrl, githubRepositoryName)
	var q []string
	if ecrRepositoryName != "" {
		q = append(q, "ecr_repository_name="+url.QueryEscape(ecrRepositoryName))
	}
	if workingDirectory != "" {
		q = append(q, "working_directory="+url.QueryEscape(workingDirectory))
	}
	if len(q) > 0 {
		reqURL = reqURL + "?" + strings.Join(q, "&")
	}

	request, err := http.NewRequest(
		http.MethodGet,
		reqURL,
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
		var apiErr apiErrorPayload
		if err := json.Unmarshal(str, &apiErr); err == nil && (apiErr.Message != "" || apiErr.ErrorType != "") {
			return fmt.Errorf(
				"%d: %s",
				response.StatusCode,
				apiErr.Message,
			)
		}

		return fmt.Errorf(
			"%d: %s",
			response.StatusCode,
			strings.TrimSpace(string(str)),
		)
	}

	err = json.NewDecoder(response.Body).Decode(lambdaArtifact)
	if err != nil {
		return err
	}

	return nil
}
