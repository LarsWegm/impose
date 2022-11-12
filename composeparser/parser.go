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
			s.LatestImage, err = reg.GetLatestVersion(s.CurrentImage)
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
