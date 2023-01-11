package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Config struct {
	Registry string
	User     string
	Password string
}

type Registry struct {
	registry   string
	client     httpClient
	httpHeader http.Header
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type tagResponse struct {
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
}

func NewRegistry(cfg *Config) *Registry {
	reg := &Registry{
		registry:   "https://hub.docker.com",
		client:     &http.Client{},
		httpHeader: http.Header{},
	}
	if cfg.Registry != "" {
		reg.registry = cfg.Registry
	}
	if cfg.User != "" && cfg.Password != "" {
		basicAuth := base64.StdEncoding.EncodeToString([]byte(cfg.User + ":" + cfg.Password))
		reg.httpHeader.Add("Authorization", "Basic "+basicAuth)
	}
	reg.httpHeader.Add("Accept", "application/json")
	return reg
}

func (r *Registry) GetImageVersions(imageName string) ([]string, error) {
	req, err := http.NewRequest("GET", r.registry+"/v2/repositories/"+imageName+"/tags/?ordering=last_updated&page=1&page_size=100", nil) // 100 is the max page_size
	if err != nil {
		return nil, err
	}
	req.Header = r.httpHeader
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry http error for '%v': %v", imageName, resp.Status)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tagRes tagResponse
	err = json.Unmarshal(bodyBytes, &tagRes)
	if err != nil {
		return nil, err
	}

	var imgVersions []string
	for _, t := range tagRes.Results {
		imgVersions = append(imgVersions, t.Name)
	}

	if len(imgVersions) < 1 {
		return nil, fmt.Errorf("could not find image versions for '%v'", imageName)
	}

	return imgVersions, nil
}
