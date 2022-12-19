package composeparser

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

type parser struct {
	file        string
	yamlContent yaml.Node
	services    []*service
}

type service struct {
	name         string
	currentImage *image
	latestImage  *image
	imageNode    *yaml.Node
	options      *serviceOptions
}

func NewParser(file string) (*parser, error) {
	if file == "" {
		return nil, errors.New("file must be set")
	}

	p := &parser{
		file: file,
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	err = p.parse(f)

	return p, err
}

func (p *parser) UpdateVersions(reg registry) error {
	g := &errgroup.Group{}
	for i := range p.services {
		idx := i
		g.Go(func() (err error) {
			s := p.services[idx]
			if s.options.ignore {
				return
			}
			mode := updateMajor
			if s.options.onlyMinor {
				mode = updateMinor
			}
			if s.options.onlyPatch {
				mode = updatePatch
			}
			s.latestImage, err = s.currentImage.GetLatestVersion(reg, mode)
			if err != nil {
				return
			}
			s.imageNode.Value = s.latestImage.String()
			return
		})
	}
	return g.Wait()
}

func (p *parser) WriteToStdout() error {
	b, err := p.marshalYaml()
	if err != nil {
		return err
	}
	_, err = fmt.Println(string(b))
	return err
}

func (p *parser) WriteToOriginalFile() error {
	if p.file == "" {
		return errors.New("no original file given")
	}
	return p.WriteToFile(p.file)
}

func (p *parser) WriteToFile(file string) error {
	b, err := p.marshalYaml()
	if err != nil {
		return err
	}
	err = os.WriteFile(file, b, 0644)
	return err
}

func (p *parser) PrintSummary() {
	changed := []*service{}
	padChanged := 0
	warnings := []*service{}
	padWarnings := 0

	for _, s := range p.services {
		if !s.options.ignore && !s.currentImage.IsSameVersion(s.latestImage) {
			padLen := len(s.currentImage.String())
			if padChanged < padLen {
				padChanged = padLen
			}
			changed = append(changed, s)
		}
	}

	for _, s := range p.services {
		if s.options.ignore {
			continue
		}
		if s.options.warnAll && s.versionHasChanged() ||
			s.options.warnMajor && s.majorHasChanged() ||
			s.options.warnMinor && s.minorHasChanged() ||
			s.options.warnPatch && s.patchHasChanged() {
			padLen := len(s.currentImage.String())
			if padWarnings < padLen {
				padWarnings = padLen
			}
			warnings = append(warnings, s)
		}
	}

	if len(changed) > 0 {
		fmt.Println("Changed versions:")
		for _, s := range changed {
			fmt.Printf("  %-*s => %s\n", padChanged, s.currentImage, s.latestImage)
		}
		if len(warnings) > 0 {
			fmt.Println()
			fmt.Println("Warnings (requires attention):")
			for _, s := range warnings {
				fmt.Printf("  %-*s => %s\n", padWarnings, s.currentImage, s.latestImage)
			}
		}
	} else {
		fmt.Println("No version changes")
	}
}

func (p *parser) marshalYaml() (b []byte, err error) {
	b, err = yaml.Marshal(&p.yamlContent)
	return
}

func (s *service) versionHasChanged() bool {
	if s.currentImage == nil || s.latestImage == nil {
		return false
	}
	return !s.currentImage.IsSameVersion(s.latestImage)
}

func (s *service) majorHasChanged() bool {
	if s.currentImage == nil || s.latestImage == nil {
		return false
	}
	return !s.currentImage.IsSameMajor(s.latestImage)
}

func (s *service) minorHasChanged() bool {
	if s.currentImage == nil || s.latestImage == nil {
		return false
	}
	return !s.currentImage.IsSameMinor(s.latestImage)
}

func (s *service) patchHasChanged() bool {
	if s.currentImage == nil || s.latestImage == nil {
		return false
	}
	return !s.currentImage.IsSamePatch(s.latestImage)
}

func (p *parser) parse(reader io.Reader) error {
	yamlBytes, err := io.ReadAll(reader)
	normLineEndings := strings.Replace(string(yamlBytes), "\r\n", "\n", -1)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal([]byte(normLineEndings), &p.yamlContent)
	if err != nil {
		return err
	}

	if len(p.yamlContent.Content) < 1 {
		return errors.New("invalid YAML content")
	}

	_, servicesNode, err := getNodeByKey(p.yamlContent.Content[0], "services")
	if err != nil {
		return err
	}

	servicesNodeContent := servicesNode.Content
	servicesNodeContentLen := len(servicesNodeContent)
	for i := 0; i < servicesNodeContentLen; i = i + 2 {
		if servicesNodeContentLen <= i+1 {
			return errors.New("could not parese YAML: invalid services node content length")
		}
		imgNodeKey, imgNode, err := getNodeByKey(servicesNodeContent[i+1], "image")
		if err != nil {
			return err
		}
		img, err := newImageFromString(imgNode.Value)
		if err != nil {
			return err
		}
		service := &service{
			name:         servicesNodeContent[i].Value,
			currentImage: img,
			imageNode:    imgNode,
			options:      newServiceOptions(imgNodeKey.HeadComment, imgNode.LineComment),
		}
		p.services = append(p.services, service)
	}
	return nil
}

func getNodeByKey(node *yaml.Node, key string) (nodeKey *yaml.Node, nodeVal *yaml.Node, err error) {
	nodeContent := node.Content
	for i, n := range nodeContent {
		if n.Value == key {
			if len(nodeContent) <= i+1 {
				return nil, nil, errors.New("could not parese YAML: invalid node content length")
			}
			return nodeContent[i], nodeContent[i+1], nil
		}
	}
	return nil, nil, fmt.Errorf("no key '%v' found in YAML", key)
}
