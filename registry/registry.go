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
	Next    string `json:"next"`
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

func (r *Registry) GetImageVersions(imageName string, stopByTag string) ([]string, error) {
	var imgVersions []string
	url := r.registry + "/v2/repositories/" + imageName + "/tags/?ordering=last_updated&page=1&page_size=100" // 100 is the max page_size
	const maxPages = 1000
	for i := 0; ; i++ {
		if i >= maxPages {
			return nil, fmt.Errorf("reached max pages limit (%d)", maxPages)
		}
		stopTagFound := false
		tagRes, err := r.retrieveTags(url)
		if err != nil {
			return nil, err
		}
		for _, t := range tagRes.Results {
			imgVersions = append(imgVersions, t.Name)
			if t.Name == stopByTag {
				stopTagFound = true
			}
		}
		if tagRes.Next != "" && !stopTagFound {
			url = tagRes.Next
		} else {
			break
		}
	}

	if len(imgVersions) < 1 {
		return nil, fmt.Errorf("could not find image versions for '%v'", imageName)
	}

	return imgVersions, nil
}

func (r *Registry) retrieveTags(url string) (*tagResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = r.httpHeader
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error '%v' for '%v'", resp.Status, url)
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

	return &tagRes, nil
}
