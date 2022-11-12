package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
)

type Config struct {
	Registry string
	User     string
	Password string
}

type Registry struct {
	registry   string
	client     *http.Client
	httpHeader http.Header
	filterTags map[string]bool
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
		filterTags: map[string]bool{
			"latest": true,
		},
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

func (r *Registry) GetLatestVersion(image *Image, mode UpdateMode) (*Image, error) {
	req, err := http.NewRequest("GET", r.registry+"/v2/repositories/"+image.GetNormalizedName()+"/tags/?ordering=last_updated&page=1&page_size=100", nil) // 100 is the max page_size
	if err != nil {
		return nil, err
	}
	req.Header = r.httpHeader
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry http error: %v", resp.Status)
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

	var imgVersions []*Image
	image.SetVersionMatcher(mode)
	for _, t := range tagRes.Results {
		if !r.filterTags[t.Name] && image.MatchesScheme(t.Name) {
			img, err := NewImageFromComponents(image.Name, t.Name)
			if err != nil {
				return nil, err
			}
			imgVersions = append(imgVersions, img)
		}
	}

	sort.Slice(imgVersions, func(i, j int) bool {
		return imgVersions[i].Less(imgVersions[j])
	})

	if len(imgVersions) < 1 {
		return nil, fmt.Errorf("could not find a valid version for '%v'", image)
	}
	highestImgVer := imgVersions[len(imgVersions)-1]
	return highestImgVer, nil
}
