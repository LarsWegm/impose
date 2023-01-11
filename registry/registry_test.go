package registry

import (
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type httpClientMock struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (c *httpClientMock) Do(req *http.Request) (*http.Response, error) {
	if c != nil && c.doFunc != nil {
		return c.doFunc(req)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(strings.NewReader(`{
	"count": 25,
	"next": "https://examplecom/v2/repositories/some/image/tags?page=2",
	"previous": null,
	"results": [
		{
			"creator": 1,
			"id": 12,
			"images": [
				{
					"architecture": "amd64",
					"features": "",
					"variant": null,
					"digest": "sha256:1f87eb2f537b91ccb906fa3d7700c39fca15442957d5fd21d2e5e45a13d9a6af",
					"os": "linux",
					"os_features": "",
					"os_version": null,
					"size": 123456789,
					"status": "active",
					"last_pulled": "2022-12-24T12:00:00.000000Z",
					"last_pushed": "2022-12-24T00:00:00.000000Z"
				}
			],
			"last_updated": "2022-12-24T00:00:00.000000Z",
			"last_updater": 123,
			"last_updater_username": "username",
			"name": "latest",
			"repository": 1,
			"full_size": 123456789,
			"v2": true,
			"tag_status": "active",
			"tag_last_pulled": "2022-12-24T12:00:00.000000Z",
			"tag_last_pushed": "2022-12-24T00:00:00.000000Z",
			"media_type": "application/vnd.docker.distribution.manifest.list.v2+json",
			"digest": "sha256:c0b15a3c334dc90fc4adc1cea31b42bf1f919d1d18870797c3ecdb4689d675a3"
		}
	]
}`)),
	}, nil
}

func TestGetImageVersions_latest(t *testing.T) {
	r := NewRegistry(&Config{})
	r.client = &httpClientMock{}
	actual, err := r.GetImageVersions("some/image")
	if err != nil {
		t.Fatal("expected no error")
	}
	expected := []string{
		"latest",
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestGetImageVersions_multipleVersions(t *testing.T) {
	r := NewRegistry(&Config{})
	r.client = &httpClientMock{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
	"results": [
		{
			"name": "1.0.0"
		},
		{
			"name": "1.1.0"
		},
		{
			"name": "2.0.0-rc"
		}
	]
}`)),
			}, nil
		},
	}
	actual, err := r.GetImageVersions("some/image")
	if err != nil {
		t.Fatal("expected no error")
	}
	expected := []string{
		"1.0.0",
		"1.1.0",
		"2.0.0-rc",
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestGetImageVersions_httpError(t *testing.T) {
	r := NewRegistry(&Config{})
	r.client = &httpClientMock{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader("{}")),
			}, nil
		},
	}
	_, err := r.GetImageVersions("some/image")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetImageVersions_invalidResponse(t *testing.T) {
	r := NewRegistry(&Config{})
	r.client = &httpClientMock{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("invalid")),
			}, nil
		},
	}
	_, err := r.GetImageVersions("some/image")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetImageVersions_httpClientError(t *testing.T) {
	r := NewRegistry(&Config{})
	r.client = &httpClientMock{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("some error")
		},
	}
	_, err := r.GetImageVersions("some/image")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetImageVersions_noVersionsFound(t *testing.T) {
	r := NewRegistry(&Config{})
	r.client = &httpClientMock{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
	"results": []
}`)),
			}, nil
		},
	}
	_, err := r.GetImageVersions("some/image")
	if err == nil {
		t.Error("expected error, got nil")
	}
}
