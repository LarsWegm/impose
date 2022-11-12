package composeparser

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"git.larswegmann.de/lars/impose/registry"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

type Parser struct {
	file        string
	yamlContent yaml.Node
	services    []*Service
}

type Service struct {
	Name         string
	CurrentImage *registry.Image
	LatestImage  *registry.Image
	imageNode    *yaml.Node
	options      *serviceOptions
}

func NewParser(file string) (*Parser, error) {
	if file == "" {
		return nil, errors.New("file must be set")
	}

	p := &Parser{}
	err := p.loadFromFile(file)

	return p, err
}

func (p *Parser) loadFromFile(file string) error {
	p.file = file
	yamlFile, err := os.ReadFile(p.file)
	normLineEndings := strings.Replace(string(yamlFile), "\r\n", "\n", -1)
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
		img, err := registry.NewImageFromString(imgNode.Value)
		if err != nil {
			return err
		}
		service := &Service{
			Name:         servicesNodeContent[i].Value,
			CurrentImage: img,
			imageNode:    imgNode,
			options:      newServiceOptions(imgNodeKey.HeadComment, imgNode.LineComment),
		}
		p.services = append(p.services, service)
	}
	return nil
}

func (p *Parser) UpdateVersions(reg *registry.Registry) error {
	g := &errgroup.Group{}
	for i := range p.services {
		idx := i
		g.Go(func() (err error) {
			s := p.services[idx]
			if s.options.ignore {
				return
			}
			mode := registry.UPDATE_MAJOR
			if s.options.onlyMinor {
				mode = registry.UPDATE_MINOR
			}
			if s.options.onlyPatch {
				mode = registry.UPDATE_PATCH
			}
			s.LatestImage, err = reg.GetLatestVersion(s.CurrentImage, mode)
			if err != nil {
				return
			}
			s.imageNode.Value = s.LatestImage.String()
			return
		})
	}
	return g.Wait()
}

func (p *Parser) WriteToStdout() error {
	b, err := yaml.Marshal(&p.yamlContent)
	if err != nil {
		return err
	}
	_, err = fmt.Println(string(b))
	return err
}

func (p *Parser) WriteToOriginalFile() error {
	return p.WriteToFile(p.file)
}

func (p *Parser) WriteToFile(file string) error {
	b, err := yaml.Marshal(&p.yamlContent)
	if err != nil {
		return err
	}
	err = os.WriteFile(file, b, 0644)
	return err
}

func (p *Parser) PrintSummary() {
	changed := []*Service{}
	padChanged := 0
	warnings := []*Service{}
	padWarnings := 0

	for _, s := range p.services {
		if !s.options.ignore && !s.CurrentImage.IsSameVersion(s.LatestImage) {
			padLen := len(s.CurrentImage.String())
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
		if s.options.warnAll && s.VersionHasChanged() ||
			s.options.warnMajor && s.MajorHasChanged() ||
			s.options.warnMinor && s.MinorHasChanged() ||
			s.options.warnPatch && s.PatchHasChanged() {
			padLen := len(s.CurrentImage.String())
			if padWarnings < padLen {
				padWarnings = padLen
			}
			warnings = append(warnings, s)
		}
	}

	if len(changed) > 0 {
		fmt.Println("Changed versions:")
		for _, s := range changed {
			fmt.Printf("  %-*s => %s\n", padChanged, s.CurrentImage, s.LatestImage)
		}
		if len(warnings) > 0 {
			fmt.Println()
			fmt.Println("Warnings (requires attention):")
			for _, s := range warnings {
				fmt.Printf("  %-*s => %s\n", padWarnings, s.CurrentImage, s.LatestImage)
			}
		}
	} else {
		fmt.Println("No version changes")
	}
}

func (s *Service) VersionHasChanged() bool {
	if s.CurrentImage == nil || s.LatestImage == nil {
		return false
	}
	return !s.CurrentImage.IsSameVersion(s.LatestImage)
}

func (s *Service) MajorHasChanged() bool {
	if s.CurrentImage == nil || s.LatestImage == nil {
		return false
	}
	return !s.CurrentImage.IsSameMajor(s.LatestImage)
}

func (s *Service) MinorHasChanged() bool {
	if s.CurrentImage == nil || s.LatestImage == nil {
		return false
	}
	return !s.CurrentImage.IsSameMinor(s.LatestImage)
}

func (s *Service) PatchHasChanged() bool {
	if s.CurrentImage == nil || s.LatestImage == nil {
		return false
	}
	return !s.CurrentImage.IsSamePatch(s.LatestImage)
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
