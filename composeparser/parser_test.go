package composeparser

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNormalFile(t *testing.T) {
	parser, err := NewParser("fixtures/one-service/docker-compose.yml")
	if err != nil {
		t.Fatal(err)
	}

	const expected = 1
	actual := len(parser.services)
	if actual != expected {
		t.Errorf("expected %d services, got %d", expected, actual)
	}

}

func TestNoFileSet(t *testing.T) {
	_, err := NewParser("")

	if err == nil {
		t.Error("expected error when no file given")
	}
}

func TestInvalidFile(t *testing.T) {
	_, err := NewParser("fixtures/invalid-file/docker-compose.yml")
	if err == nil {
		t.Error("expected error when invalid file given")
	}
}

func TestStdout(t *testing.T) {
	parser, err := NewParser("fixtures/one-service/docker-compose.yml")
	if err != nil {
		t.Fatal(err)
	}

	actual := getStdout(t, func() {
		err = parser.WriteToStdout()
	})

	if err != nil {
		t.Fatal(err)
	}

	const expected = `version: '3'
services:
    my-service:
        image: alpine:3.15.5

`
	if actual != expected {
		t.Errorf("expected '%q', got '%q'", expected, actual)
	}
}

// There is a bug in yaml.Unmarshal: When files with CRLF line endings are used,
// multi line comments get an additional new line, e.g. this:
//
// -----yaml-----
// # multi line
// # comment
// --------------
//
// will get:
//
// -----yaml-----
// # multi line
//
// # comment
// --------------
//
// Therfore we normalize CRLF to LF before unmarshalling to get rid of the problem.
func TestNormalizeCRLF(t *testing.T) {
	parser, err := NewParser("fixtures/cr-lf-with-comments/docker-compose.yml")
	if err != nil {
		t.Fatal(err)
	}

	actual := getYamlStr(t, parser)

	const expected = `version: '3'
services:
    my-service:
        # This is a multi
        # line comment
        image: alpine:3.15.5
`
	if actual != expected {
		t.Errorf("expected '%q', got '%q'", expected, actual)
	}
}

func TestImageUpdate(t *testing.T) {
	parser, err := NewParser("fixtures/multiple-services/docker-compose.yml")
	if err != nil {
		t.Fatal(err)
	}

	reg := &registryMock{}
	parser.UpdateVersions(reg)

	actual := getYamlStr(t, parser)
	const expected = `version: '3'
services:
    my-service-1:
        image: alpine:1.0.0
    my-service-2:
        image: custom/image:1.0.0
    my-service-3:
        image: mysql:1.0.0
`
	if actual != expected {
		t.Errorf("expected '%q', got '%q'", expected, actual)
	}
}

func TestOptions(t *testing.T) {
	var tests = []struct {
		name     string
		file     string
		expected map[string]serviceOptions
	}{
		{
			"ignore in head comment",
			"fixtures/options/ignore/head.yml",
			map[string]serviceOptions{
				"my-service": {
					ignore: true,
				},
			},
		},
		{
			"ignore in inline comment",
			"fixtures/options/ignore/inline.yml",
			map[string]serviceOptions{
				"my-service": {
					ignore: true,
				},
			},
		},
		{
			"minor in head comment",
			"fixtures/options/minor/head.yml",
			map[string]serviceOptions{
				"my-service": {
					onlyMinor: true,
				},
			},
		},
		{
			"minor in inline comment",
			"fixtures/options/minor/inline.yml",
			map[string]serviceOptions{
				"my-service": {
					onlyMinor: true,
				},
			},
		},
		{
			"patch in head comment",
			"fixtures/options/patch/head.yml",
			map[string]serviceOptions{
				"my-service": {
					onlyPatch: true,
				},
			},
		},
		{
			"patch in inline comment",
			"fixtures/options/patch/inline.yml",
			map[string]serviceOptions{
				"my-service": {
					onlyPatch: true,
				},
			},
		},
		{
			"warnAll in head comment",
			"fixtures/options/warnAll/head.yml",
			map[string]serviceOptions{
				"my-service": {
					warnAll: true,
				},
			},
		},
		{
			"warnAll in inline comment",
			"fixtures/options/warnAll/inline.yml",
			map[string]serviceOptions{
				"my-service": {
					warnAll: true,
				},
			},
		},
		{
			"warnMajor in head comment",
			"fixtures/options/warnMajor/head.yml",
			map[string]serviceOptions{
				"my-service": {
					warnMajor: true,
				},
			},
		},
		{
			"warnMajor in inline comment",
			"fixtures/options/warnMajor/inline.yml",
			map[string]serviceOptions{
				"my-service": {
					warnMajor: true,
				},
			},
		},
		{
			"warnMinor in head comment",
			"fixtures/options/warnMinor/head.yml",
			map[string]serviceOptions{
				"my-service": {
					warnMinor: true,
				},
			},
		},
		{
			"warnMinor in inline comment",
			"fixtures/options/warnMinor/inline.yml",
			map[string]serviceOptions{
				"my-service": {
					warnMinor: true,
				},
			},
		},
		{
			"warnPatch in head comment",
			"fixtures/options/warnPatch/head.yml",
			map[string]serviceOptions{
				"my-service": {
					warnPatch: true,
				},
			},
		},
		{
			"warnPatch in inline comment",
			"fixtures/options/warnPatch/inline.yml",
			map[string]serviceOptions{
				"my-service": {
					warnPatch: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser(tt.file)
			if err != nil {
				t.Fatal(err)
			}
			actual := make(map[string]serviceOptions)
			for _, s := range parser.services {
				actual[s.name] = *s.options
			}
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Errorf("expected %+v, got %+v", tt.expected, actual)
			}
		})
	}
}

func getYamlStr(t *testing.T, p *parser) string {
	b, err := yaml.Marshal(&p.yamlContent)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func getStdout(t *testing.T, runFunc func()) string {
	origStdout := os.Stdout
	defer func() { os.Stdout = origStdout }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	runFunc()
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	bOut, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	err = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	return string(bOut)
}
