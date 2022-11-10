package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
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

type matcherFunc func(version string) bool

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

func (r *registry) GetLatestVersion(image string, refVersion string) (string, error) {
	if len(strings.Split(image, "/")) == 1 {
		image = "library/" + image
	}
	req, err := http.NewRequest("GET", r.registry+"/v2/repositories/"+image+"/tags/?ordering=last_updated&page=1&page_size=100", nil) // 100 is the max page_size
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

	var tags []string
	tagMatcher := getTagMatcher(refVersion)
	for _, t := range tagRes.Results {
		if !r.filterTags[t.Name] && tagMatcher(t.Name) {
			tags = append(tags, t.Name)
		}
	}

	sort.Slice(tags, func(i, j int) bool {
		return lessVersion(tags[i], tags[j])
	})

	if len(tags) < 1 {
		return "", fmt.Errorf("could not find a valid version for '%v:%v'", image, refVersion)
	}
	return tags[len(tags)-1], nil
}

var re3DigitsSuffix = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+.*$`)
var re2DigitsSuffix = regexp.MustCompile(`^[0-9]+\.[0-9]+.*$`)
var re1DigitSuffix = regexp.MustCompile(`^[0-9]+.*$`)
var reV3DigitsSuffix = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+.*$`)
var reV2DigitsSuffix = regexp.MustCompile(`^v[0-9]+\.[0-9]+.*$`)
var reV1DigitsSuffix = regexp.MustCompile(`^v[0-9]+.*$`)

func getTagMatcher(refVersion string) matcherFunc {
	_, suffix, _ := strings.Cut(refVersion, "-")

	if re3DigitsSuffix.MatchString(refVersion) {
		return func(version string) bool {
			// strings.HasSuffix is too inaccurate, we need to compare the exact suffix
			_, currentSuffix, _ := strings.Cut(version, "-")
			return re3DigitsSuffix.MatchString(version) && suffix == currentSuffix
		}
	}

	if re2DigitsSuffix.MatchString(refVersion) {
		return func(version string) bool {
			_, currentSuffix, _ := strings.Cut(version, "-")
			return re2DigitsSuffix.MatchString(version) && suffix == currentSuffix
		}
	}

	if re1DigitSuffix.MatchString(refVersion) {
		return func(version string) bool {
			_, currentSuffix, _ := strings.Cut(version, "-")
			return re1DigitSuffix.MatchString(version) && suffix == currentSuffix
		}
	}

	if reV3DigitsSuffix.MatchString(refVersion) {
		return func(version string) bool {
			_, currentSuffix, _ := strings.Cut(version, "-")
			return reV3DigitsSuffix.MatchString(version) && suffix == currentSuffix
		}
	}

	if reV2DigitsSuffix.MatchString(refVersion) {
		return func(version string) bool {
			_, currentSuffix, _ := strings.Cut(version, "-")
			return reV2DigitsSuffix.MatchString(version) && suffix == currentSuffix
		}
	}

	if reV1DigitsSuffix.MatchString(refVersion) {
		return func(version string) bool {
			_, currentSuffix, _ := strings.Cut(version, "-")
			return reV1DigitsSuffix.MatchString(version) && suffix == currentSuffix
		}
	}

	// Match all fallback
	return func(version string) bool {
		return true
	}
}

func lessVersion(v1 string, v2 string) bool {
	v1Major, v1Minor, v1Patch, v1Suffix := normalizeVersion(v1)
	v2Major, v2Minor, v2Patch, v2Suffix := normalizeVersion(v2)
	if v1Major < v2Major {
		return true
	}
	if v1Major == v2Major && v1Minor < v2Minor {
		return true
	}
	if v1Major == v2Major && v1Minor == v2Minor && v1Patch < v2Patch {
		return true
	}
	if v1Major == v2Major && v1Minor == v2Minor && v1Patch == v2Patch && v1Suffix < v2Suffix {
		return true
	}
	return false
}

func normalizeVersion(version string) (major int, minor int, patch int, suffix string) {
	version = strings.TrimPrefix(version, "v")
	version, suffix, _ = strings.Cut(version, "-")

	verSlice := strings.Split(version, ".")
	verSliceLen := len(verSlice)
	if verSliceLen > 0 {
		major, _ = strconv.Atoi(verSlice[0])
	}
	if verSliceLen > 1 {
		minor, _ = strconv.Atoi(verSlice[1])
	}
	if verSliceLen > 2 {
		patch, _ = strconv.Atoi(verSlice[2])
	}
	return
}
