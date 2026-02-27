package depsdev

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/JCzapla/dep-dashboard/internal/domain"
)


const baseURL = "https://api.deps.dev/v3"

type Client struct {
	http *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{http: httpClient}
}

type getPackageResponse struct {
	Versions []struct {
		VersionKey struct {
			Version string
		} `json:"versionKey"`
		IsDefault bool `json:"isDefault"`
	} `json:"versions"`
}


func (c *Client) FetchDefaultVersion(ctx context.Context, name string) (string, error) {
	apiURL := fmt.Sprintf("%s/systems/NPM/packages/%s",
		baseURL,
		url.PathEscape(name),
	)
	var result getPackageResponse
	if err := c.doRequest(ctx, http.MethodGet, apiURL, &result); err != nil {
		return "", err
	}
	
	for _, version := range result.Versions {
		if version.IsDefault {
			return version.VersionKey.Version, nil
		}
	}

	return result.Versions[0].VersionKey.Version, nil
}

type getVersionDependenciesResponse struct {
	Nodes []struct {
		VersionKey struct {
			System string
			Name string
			Version string
		} `json:"versionKey"`
		Relation string `json:"relation"`
	} `json:"nodes"`
}


func (c *Client) FetchDependencies(ctx context.Context, ref domain.PackageRef) ([]domain.DependencyNode, error) {
	apiURL := fmt.Sprintf("%s/systems/NPM/packages/%s/versions/%s:dependencies",
		baseURL,
		url.PathEscape(ref.Name),
		url.PathEscape(ref.Version),
	)

	var result getVersionDependenciesResponse
	if err := c.doRequest(ctx, http.MethodGet, apiURL, &result); err != nil {
		return nil, err
	}

	var nodes []domain.DependencyNode
	for _, n := range result.Nodes {
		nodes = append(nodes, domain.DependencyNode{
			Name: n.VersionKey.Name,
			Version: n.VersionKey.Version,
			Relation: n.Relation,
		})
	}

	return nodes, nil
}

type getVersionResponse struct {
	RelatedProjects []struct {
		ProjectKey struct {
			ID string `json:"id"`
		} `json:"projectKey"`
	} `json:"relatedProjects"`
}

func (c *Client) FetchProjectKey(ctx context.Context, ref domain.PackageRef) (string, error) {
	apiURL := fmt.Sprintf("%s/systems/NPM/packages/%s/versions/%s",
		baseURL,
		url.PathEscape(ref.Name),
		url.PathEscape(ref.Version),
	)
	var result getVersionResponse
	if err := c.doRequest(ctx, http.MethodGet, apiURL, &result); err != nil {
		return "", err
	}

	if len(result.RelatedProjects) == 0 {
		return "", nil
	}
	return result.RelatedProjects[0].ProjectKey.ID,nil
}

type getProjectResponse struct {
	Scorecard struct {
		OverallScore float64 `json:"overallScore"`
	} `json:"scorecard"`
}

func (c *Client) FetchScore(ctx context.Context, projectKey string) (float64, error) {
	apiURL := fmt.Sprintf("%s/projects/%s",
		baseURL,
		url.PathEscape(projectKey),
	)
	var result getProjectResponse
	if err := c.doRequest(ctx, http.MethodGet, apiURL, &result); err != nil {
		return 0.0, err
	}

	return result.Scorecard.OverallScore, nil
}

func (c *Client) doRequest(ctx context.Context, method string, url string, result any) error{
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("Error building request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("Error calling deps.dev API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Error from deps.dev: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("Error decoding response: %w", err)
	}

	return nil
}