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
	registry string
	user     string
	password string
}

type registry struct {
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

func NewRegistry(cfg Config) *registry {
	reg := &registry{
		registry:   "https://hub.docker.com",
		client:     &http.Client{},
		httpHeader: http.Header{},
		filterTags: map[string]bool{
			"latest": true,
		},
	}
	if cfg.registry != "" {
		reg.registry = cfg.registry
	}
	if cfg.user != "" && cfg.password != "" {
		basicAuth := base64.StdEncoding.EncodeToString([]byte(cfg.user + ":" + cfg.password))
		reg.httpHeader.Add("Authorization", "Basic "+basicAuth)
	}
	reg.httpHeader.Add("Accept", "application/json")
	return reg
}

func (r *registry) GetLatestVersion(image *Image) (string, error) {
	req, err := http.NewRequest("GET", r.registry+"/v2/repositories/"+image.GetNormalizedName()+"/tags/?ordering=last_updated&page=1&page_size=100", nil) // 100 is the max page_size
	if err != nil {
		return "", err
	}
	req.Header = r.httpHeader
	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registry http error: %v", resp.Status)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tagRes tagResponse
	err = json.Unmarshal(bodyBytes, &tagRes)
	if err != nil {
		return "", err
	}

	var imgVersions []*Image
	for _, t := range tagRes.Results {
		if !r.filterTags[t.Name] && image.MatchesScheme(t.Name) {
			img, err := NewImageFromComponents(image.Name, t.Name)
			if err != nil {
				return "", err
			}
			imgVersions = append(imgVersions, img)
		}
	}

	sort.Slice(imgVersions, func(i, j int) bool {
		return imgVersions[i].Less(imgVersions[j])
	})

	if len(imgVersions) < 1 {
		return "", fmt.Errorf("could not find a valid version for '%v'", image)
	}
	highestImgVer := imgVersions[len(imgVersions)-1]
	return highestImgVer.VersionStr, nil
}
