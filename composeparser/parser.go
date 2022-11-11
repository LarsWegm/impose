package composeparser

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"git.larswegmann.de/lars/impose/registry"
	"gopkg.in/yaml.v3"
)

type parser struct {
	file        string
	registry    *registry.Registry
	yamlContent yaml.Node
	services    []*Service
}

type Service struct {
	Name         string
	CurrentImage *registry.Image
	imageNode    *yaml.Node
	LatestImage  *registry.Image
}

func NewParser(file string, registry *registry.Registry) (*parser, error) {
	if file == "" {
		return nil, errors.New("file must be set")
	}

	p := &parser{
		registry: registry,
	}
	err := p.loadFromFile(file)

	return p, err
}

func (p *parser) loadFromFile(file string) error {
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

	servicesNode, err := getNodeByKey(p.yamlContent.Content[0], "services")
	if err != nil {
		return err
	}

	servicesNodeContent := servicesNode.Content
	servicesNodeContentLen := len(servicesNodeContent)
	for i := 0; i < servicesNodeContentLen; i = i + 2 {
		if servicesNodeContentLen <= i+1 {
			return errors.New("could not parese YAML: invalid services node content length")
		}
		imgNode, err := getNodeByKey(servicesNodeContent[i+1], "image")
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
		}
		p.services = append(p.services, service)
	}
	return nil
}

func (p *parser) UpdateVersions() (err error) {
	for _, s := range p.services {
		s.LatestImage, err = p.registry.GetLatestVersion(s.CurrentImage)
		if err != nil {
			return
		}
		s.imageNode.Value = s.LatestImage.String()
		fmt.Println(s.LatestImage.String())
	}
	return
}

func (p *parser) WriteToStdout() error {
	b, err := yaml.Marshal(&p.yamlContent)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func (p *parser) writeToFile() error {
	b, err := yaml.Marshal(&p.yamlContent)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	//os.WriteFile("out.yml", b, 0644)
	//os.WriteFile(p.file, b, 0644)
	return nil
}

func getNodeByKey(node *yaml.Node, key string) (*yaml.Node, error) {
	nodeContent := node.Content
	for i, n := range nodeContent {
		if n.Value == key {
			if len(nodeContent) <= i+1 {
				return nil, errors.New("could not parese YAML: invalid node content length")
			}
			return nodeContent[i+1], nil
		}
	}
	return nil, fmt.Errorf("no key '%v' found in YAML", key)
}
