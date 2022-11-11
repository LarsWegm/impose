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
	yamlContent yaml.Node
	services    []*Service
}

type Service struct {
	Name  string
	Image *Image
}

type Image struct {
	*registry.Image
	yamlNode *yaml.Node
}

func NewParser(file string) (*parser, error) {
	if file == "" {
		return nil, errors.New("file must be set")
	}

	p := &parser{}
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
		img, err := getImageFromNode(servicesNodeContent[i+1])
		if err != nil {
			return err
		}
		service := &Service{
			Name:  servicesNodeContent[i].Value,
			Image: img,
		}
		p.services = append(p.services, service)
	}
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

func getImageFromNode(node *yaml.Node) (*Image, error) {
	imgNode, err := getNodeByKey(node, "image")
	if err != nil {
		return nil, err
	}

	imgFromStr, err := registry.NewImageFromString(imgNode.Value)
	if err != nil {
		return nil, err
	}

	img := &Image{
		imgFromStr,
		imgNode,
	}

	return img, nil
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
